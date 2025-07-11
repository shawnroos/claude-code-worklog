"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.ReferenceMapper = void 0;
class ReferenceMapper {
    workStateManager;
    constructor(workStateManager) {
        this.workStateManager = workStateManager;
    }
    /**
     * Generate a complete reference map for current work context
     */
    generateReferenceMap() {
        const activeItems = this.workStateManager.loadActiveTodos();
        const allNodes = [];
        const allEdges = [];
        // Create nodes for active items
        for (const item of activeItems) {
            const node = this.createNodeFromItem(item, 'active');
            allNodes.push(node);
            // Add edges from smart references
            if (item.metadata?.smart_references) {
                for (const ref of item.metadata.smart_references) {
                    const targetItem = this.workStateManager.getHistoricalItem(ref.target_id);
                    if (targetItem) {
                        // Add target node if not already present
                        if (!allNodes.find(n => n.id === ref.target_id)) {
                            const targetNode = this.createNodeFromItem(targetItem, 'historical');
                            allNodes.push(targetNode);
                        }
                        // Add edge
                        const edge = {
                            source: item.id,
                            target: ref.target_id,
                            relationship_type: ref.relationship_type,
                            strength: ref.similarity_score,
                            confidence: ref.confidence,
                            auto_generated: ref.auto_generated
                        };
                        allEdges.push(edge);
                    }
                }
            }
        }
        // Generate clusters
        const clusters = this.generateClusters(allNodes, allEdges);
        // Calculate summary
        const summary = {
            total_items: allNodes.length,
            total_references: allEdges.length,
            cluster_count: clusters.length,
            orphaned_items: allNodes.filter(n => !allEdges.some(e => e.source === n.id || e.target === n.id)).length
        };
        return {
            nodes: allNodes,
            edges: allEdges,
            clusters: clusters,
            summary: summary
        };
    }
    /**
     * Generate a focused reference map for a specific work item
     */
    generateFocusedMap(itemId, depth = 2) {
        const visited = new Set();
        const nodes = [];
        const edges = [];
        this.buildMapRecursively(itemId, depth, visited, nodes, edges);
        const clusters = this.generateClusters(nodes, edges);
        const summary = {
            total_items: nodes.length,
            total_references: edges.length,
            cluster_count: clusters.length,
            orphaned_items: 0
        };
        return {
            nodes: nodes,
            edges: edges,
            clusters: clusters,
            summary: summary
        };
    }
    /**
     * Get reference path between two work items
     */
    findReferencePath(sourceId, targetId) {
        const visited = new Set();
        const path = [];
        if (this.findPathRecursively(sourceId, targetId, visited, path)) {
            return path;
        }
        return []; // No path found
    }
    /**
     * Generate ASCII visualization of reference relationships
     */
    generateASCIIVisualization(referenceMap) {
        let visualization = '\n=== Work Item Reference Map ===\n\n';
        // Group nodes by type
        const activeNodes = referenceMap.nodes.filter(n => n.type === 'active');
        const historicalNodes = referenceMap.nodes.filter(n => n.type === 'historical');
        // Display active items with their references
        if (activeNodes.length > 0) {
            visualization += '📋 ACTIVE WORK:\n';
            for (const node of activeNodes) {
                visualization += `  ● ${node.label}\n`;
                visualization += `    ${node.content_preview}\n`;
                // Show outgoing references
                const outgoingEdges = referenceMap.edges.filter(e => e.source === node.id);
                if (outgoingEdges.length > 0) {
                    visualization += '    References:\n';
                    for (const edge of outgoingEdges) {
                        const targetNode = referenceMap.nodes.find(n => n.id === edge.target);
                        const relationshipIcon = this.getRelationshipIcon(edge.relationship_type);
                        const strengthBar = this.getStrengthBar(edge.strength);
                        visualization += `      ${relationshipIcon} ${targetNode?.label || edge.target} ${strengthBar}\n`;
                    }
                }
                visualization += '\n';
            }
        }
        // Display reference clusters
        if (referenceMap.clusters.length > 0) {
            visualization += '🔗 REFERENCE CLUSTERS:\n';
            for (const cluster of referenceMap.clusters) {
                visualization += `  📁 ${cluster.name} (${cluster.nodes.length} items)\n`;
                visualization += `     Themes: ${cluster.common_themes.join(', ')}\n`;
                visualization += `     Type: ${cluster.cluster_type}\n\n`;
            }
        }
        // Display summary
        visualization += '📊 SUMMARY:\n';
        visualization += `  Items: ${referenceMap.summary.total_items}\n`;
        visualization += `  References: ${referenceMap.summary.total_references}\n`;
        visualization += `  Clusters: ${referenceMap.summary.cluster_count}\n`;
        visualization += `  Orphaned: ${referenceMap.summary.orphaned_items}\n`;
        return visualization;
    }
    createNodeFromItem(item, type) {
        const metadata = item.metadata?.similarity_metadata || {
            keywords: [],
            feature_domain: '',
            technical_domain: '',
            code_locations: [],
            strategic_theme: ''
        };
        return {
            id: item.id,
            label: `${item.type}: ${item.content.slice(0, 40)}...`,
            type: type,
            content_preview: item.content.slice(0, 100),
            priority: item.metadata?.priority,
            status: item.status,
            metadata: {
                feature_domain: metadata.feature_domain,
                technical_domain: metadata.technical_domain,
                strategic_theme: metadata.strategic_theme,
                reference_count: item.metadata?.smart_references?.length || 0
            }
        };
    }
    buildMapRecursively(itemId, remainingDepth, visited, nodes, edges) {
        if (remainingDepth <= 0 || visited.has(itemId)) {
            return;
        }
        visited.add(itemId);
        // Find the item (active or historical)
        const activeItems = this.workStateManager.loadActiveTodos();
        let item = activeItems.find(i => i.id === itemId);
        let nodeType = 'active';
        if (!item) {
            const historicalItem = this.workStateManager.getHistoricalItem(itemId);
            item = historicalItem || undefined;
            nodeType = 'historical';
        }
        if (!item)
            return;
        // Add node
        const node = this.createNodeFromItem(item, nodeType);
        if (!nodes.find(n => n.id === itemId)) {
            nodes.push(node);
        }
        // Process references
        if (item.metadata?.smart_references) {
            for (const ref of item.metadata.smart_references) {
                // Add edge
                const edge = {
                    source: itemId,
                    target: ref.target_id,
                    relationship_type: ref.relationship_type,
                    strength: ref.similarity_score,
                    confidence: ref.confidence,
                    auto_generated: ref.auto_generated
                };
                if (!edges.find(e => e.source === edge.source && e.target === edge.target)) {
                    edges.push(edge);
                }
                // Recursively process target
                this.buildMapRecursively(ref.target_id, remainingDepth - 1, visited, nodes, edges);
            }
        }
    }
    findPathRecursively(sourceId, targetId, visited, currentPath) {
        if (sourceId === targetId) {
            currentPath.push(sourceId);
            return true;
        }
        if (visited.has(sourceId)) {
            return false;
        }
        visited.add(sourceId);
        currentPath.push(sourceId);
        // Check references from this item
        const item = this.findItem(sourceId);
        if (item?.metadata?.smart_references) {
            for (const ref of item.metadata.smart_references) {
                if (this.findPathRecursively(ref.target_id, targetId, visited, currentPath)) {
                    return true;
                }
            }
        }
        // Backtrack
        currentPath.pop();
        return false;
    }
    findItem(itemId) {
        const activeItems = this.workStateManager.loadActiveTodos();
        const activeItem = activeItems.find(i => i.id === itemId);
        if (activeItem)
            return activeItem;
        const historicalItem = this.workStateManager.getHistoricalItem(itemId);
        return historicalItem || null;
    }
    generateClusters(nodes, edges) {
        const clusters = [];
        // Group by feature domain
        const featureClusters = this.groupByProperty(nodes, 'feature_domain');
        for (const [domain, nodeIds] of featureClusters) {
            if (domain && nodeIds.length > 1) {
                clusters.push({
                    id: `feature-${domain}`,
                    name: `Feature: ${domain.replace('-', ' ')}`,
                    nodes: nodeIds,
                    common_themes: [domain],
                    cluster_type: 'feature',
                    centrality_score: this.calculateCentrality(nodeIds, edges)
                });
            }
        }
        // Group by technical domain
        const techClusters = this.groupByProperty(nodes, 'technical_domain');
        for (const [domain, nodeIds] of techClusters) {
            if (domain && nodeIds.length > 1) {
                clusters.push({
                    id: `tech-${domain}`,
                    name: `Technical: ${domain.replace('-', ' ')}`,
                    nodes: nodeIds,
                    common_themes: [domain],
                    cluster_type: 'technical',
                    centrality_score: this.calculateCentrality(nodeIds, edges)
                });
            }
        }
        return clusters;
    }
    groupByProperty(nodes, property) {
        const groups = new Map();
        for (const node of nodes) {
            const value = node.metadata[property] || '';
            if (value) {
                if (!groups.has(value)) {
                    groups.set(value, []);
                }
                groups.get(value).push(node.id);
            }
        }
        return groups;
    }
    calculateCentrality(nodeIds, edges) {
        // Simple centrality based on edge count
        const edgeCount = edges.filter(e => nodeIds.includes(e.source) || nodeIds.includes(e.target)).length;
        return edgeCount / Math.max(nodeIds.length, 1);
    }
    getRelationshipIcon(type) {
        switch (type) {
            case 'continuation': return '→';
            case 'conflict': return '⚠';
            case 'dependency': return '↗';
            default: return '~';
        }
    }
    getStrengthBar(strength) {
        const bars = Math.round(strength * 5);
        const filled = '█'.repeat(bars);
        const empty = '░'.repeat(5 - bars);
        return `[${filled}${empty}]`;
    }
}
exports.ReferenceMapper = ReferenceMapper;
//# sourceMappingURL=ReferenceMapper.js.map