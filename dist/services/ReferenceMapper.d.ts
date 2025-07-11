import { WorkStateManager } from './WorkStateManager';
export interface ReferenceMap {
    nodes: ReferenceNode[];
    edges: ReferenceEdge[];
    clusters: ReferenceCluster[];
    summary: {
        total_items: number;
        total_references: number;
        cluster_count: number;
        orphaned_items: number;
    };
}
export interface ReferenceNode {
    id: string;
    label: string;
    type: 'active' | 'historical' | 'future';
    content_preview: string;
    priority?: string;
    status?: string;
    metadata: {
        feature_domain?: string;
        technical_domain?: string;
        strategic_theme?: string;
        reference_count: number;
    };
}
export interface ReferenceEdge {
    source: string;
    target: string;
    relationship_type: 'related' | 'continuation' | 'conflict' | 'dependency';
    strength: number;
    confidence: number;
    auto_generated: boolean;
}
export interface ReferenceCluster {
    id: string;
    name: string;
    nodes: string[];
    common_themes: string[];
    cluster_type: 'feature' | 'technical' | 'strategic' | 'temporal';
    centrality_score: number;
}
export declare class ReferenceMapper {
    private workStateManager;
    constructor(workStateManager: WorkStateManager);
    /**
     * Generate a complete reference map for current work context
     */
    generateReferenceMap(): ReferenceMap;
    /**
     * Generate a focused reference map for a specific work item
     */
    generateFocusedMap(itemId: string, depth?: number): ReferenceMap;
    /**
     * Get reference path between two work items
     */
    findReferencePath(sourceId: string, targetId: string): string[];
    /**
     * Generate ASCII visualization of reference relationships
     */
    generateASCIIVisualization(referenceMap: ReferenceMap): string;
    private createNodeFromItem;
    private buildMapRecursively;
    private findPathRecursively;
    private findItem;
    private generateClusters;
    private groupByProperty;
    private calculateCentrality;
    private getRelationshipIcon;
    private getStrengthBar;
}
//# sourceMappingURL=ReferenceMapper.d.ts.map