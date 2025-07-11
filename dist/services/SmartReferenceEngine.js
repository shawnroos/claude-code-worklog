"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.SmartReferenceEngine = void 0;
class SmartReferenceEngine {
    workStateManager;
    similarityThreshold = 0.7;
    confidenceThreshold = 0.6;
    constructor(workStateManager) {
        this.workStateManager = workStateManager;
    }
    /**
     * Generate automatic references for a given work item
     */
    generateAutomaticReferences(activeItem) {
        const references = [];
        // Query historical items that might be related
        const historicalItems = this.getRelevantHistoricalItems(activeItem);
        for (const historicalItem of historicalItems) {
            const similarity = this.calculateSemanticSimilarity(activeItem, historicalItem);
            if (similarity.total_score >= this.similarityThreshold) {
                const reference = {
                    id: this.generateId(),
                    target_item_id: historicalItem.id,
                    similarity_score: similarity.total_score,
                    confidence: this.calculateConfidence(similarity),
                    relationship_type: this.determineRelationshipType(activeItem, historicalItem, similarity),
                    auto_generated: true,
                    created_at: new Date().toISOString(),
                    metadata: {
                        common_keywords: similarity.common_keywords,
                        domain_overlap: similarity.domain_overlap,
                        code_location_overlap: similarity.code_location_overlap,
                        strategic_alignment: similarity.strategic_alignment
                    }
                };
                references.push(reference);
            }
        }
        return references.sort((a, b) => b.similarity_score - a.similarity_score);
    }
    /**
     * Calculate semantic similarity between two work items
     */
    calculateSemanticSimilarity(item1, item2) {
        const metadata1 = item1.metadata?.similarity_metadata || {
            keywords: [],
            feature_domain: '',
            technical_domain: '',
            code_locations: [],
            strategic_theme: ''
        };
        const metadata2 = item2.metadata?.similarity_metadata || {
            keywords: [],
            feature_domain: '',
            technical_domain: '',
            code_locations: [],
            strategic_theme: ''
        };
        // Keyword similarity
        const keywords1 = metadata1.keywords || [];
        const keywords2 = metadata2.keywords || [];
        const commonKeywords = keywords1.filter((k) => keywords2.includes(k));
        const keywordScore = commonKeywords.length / Math.max(keywords1.length, keywords2.length, 1);
        // Domain similarity
        const domainOverlap = [];
        if (metadata1.feature_domain === metadata2.feature_domain && metadata1.feature_domain) {
            domainOverlap.push(metadata1.feature_domain);
        }
        if (metadata1.technical_domain === metadata2.technical_domain && metadata1.technical_domain) {
            domainOverlap.push(metadata1.technical_domain);
        }
        const domainScore = domainOverlap.length > 0 ? 0.8 : 0;
        // Code location similarity
        const locations1 = metadata1.code_locations || [];
        const locations2 = metadata2.code_locations || [];
        const locationOverlap = locations1.filter((l) => locations2.includes(l));
        const locationScore = locationOverlap.length / Math.max(locations1.length, locations2.length, 1);
        // Strategic theme alignment
        const strategicAlignment = metadata1.strategic_theme === metadata2.strategic_theme ? metadata1.strategic_theme : '';
        const strategicScore = strategicAlignment ? 0.6 : 0;
        // Content similarity (basic text comparison)
        const contentScore = this.calculateContentSimilarity(item1.content, item2.content);
        // Weighted total score
        const totalScore = (keywordScore * 0.3 +
            domainScore * 0.25 +
            locationScore * 0.2 +
            strategicScore * 0.15 +
            contentScore * 0.1);
        return {
            total_score: totalScore,
            keyword_score: keywordScore,
            domain_score: domainScore,
            location_score: locationScore,
            strategic_score: strategicScore,
            content_score: contentScore,
            common_keywords: commonKeywords,
            domain_overlap: domainOverlap,
            code_location_overlap: locationOverlap,
            strategic_alignment: strategicAlignment
        };
    }
    /**
     * Get contextual suggestions based on current active work
     */
    getContextualSuggestions(activeItems) {
        const suggestions = [];
        for (const activeItem of activeItems) {
            const references = this.generateAutomaticReferences(activeItem);
            for (const reference of references) {
                if (reference.confidence >= this.confidenceThreshold) {
                    const suggestion = this.createSuggestion(activeItem, reference);
                    if (suggestion) {
                        suggestions.push(suggestion);
                    }
                }
            }
        }
        return suggestions.sort((a, b) => {
            // Sort by priority first, then confidence
            const priorityOrder = { high: 3, medium: 2, low: 1 };
            if (priorityOrder[a.priority] !== priorityOrder[b.priority]) {
                return priorityOrder[b.priority] - priorityOrder[a.priority];
            }
            return b.confidence - a.confidence;
        });
    }
    /**
     * Update references when a work item changes
     */
    updateReferencesOnChange(itemId) {
        // Find active items that might need reference updates
        const activeItems = this.workStateManager.loadActiveTodos();
        const changedItem = activeItems.find(item => item.id === itemId);
        if (changedItem) {
            // Regenerate references for the changed item
            const newReferences = this.generateAutomaticReferences(changedItem);
            // Update the item with new references
            this.updateItemReferences(changedItem, newReferences);
            // Also check if other active items need reference updates due to this change
            for (const otherItem of activeItems) {
                if (otherItem.id !== itemId) {
                    const similarity = this.calculateSemanticSimilarity(otherItem, changedItem);
                    if (similarity.total_score >= this.similarityThreshold) {
                        // Create cross-reference
                        this.addCrossReference(otherItem, changedItem, similarity);
                    }
                }
            }
        }
    }
    getRelevantHistoricalItems(activeItem) {
        // Get keywords from active item for search
        const metadata = activeItem.metadata?.similarity_metadata || {
            keywords: [],
            feature_domain: '',
            technical_domain: '',
            code_locations: [],
            strategic_theme: ''
        };
        const keywords = metadata.keywords || [];
        const featureDomain = metadata.feature_domain;
        const technicalDomain = metadata.technical_domain;
        let historicalItems = [];
        // Search by keywords
        for (const keyword of keywords.slice(0, 3)) { // Limit to top 3 keywords
            const results = this.workStateManager.queryHistory(keyword);
            historicalItems.push(...results);
        }
        // Search by domain if available
        if (featureDomain) {
            const domainResults = this.workStateManager.queryHistory(featureDomain);
            historicalItems.push(...domainResults);
        }
        if (technicalDomain) {
            const techResults = this.workStateManager.queryHistory(technicalDomain);
            historicalItems.push(...techResults);
        }
        // Remove duplicates and limit results
        const uniqueItems = historicalItems.filter((item, index, self) => index === self.findIndex(i => i.id === item.id));
        return uniqueItems.slice(0, 20); // Limit to 20 most relevant items
    }
    calculateContentSimilarity(content1, content2) {
        // Simple content similarity using word overlap
        const words1 = this.extractWords(content1);
        const words2 = this.extractWords(content2);
        const commonWords = words1.filter(word => words2.includes(word));
        return commonWords.length / Math.max(words1.length, words2.length, 1);
    }
    extractWords(content) {
        return content.toLowerCase()
            .replace(/[^\w\s]/g, ' ')
            .split(/\s+/)
            .filter(word => word.length > 3)
            .slice(0, 50); // Limit words for performance
    }
    calculateConfidence(similarity) {
        // Confidence based on multiple factors
        let confidence = similarity.total_score;
        // Boost confidence if multiple dimensions align
        const alignedDimensions = [
            similarity.keyword_score > 0.3,
            similarity.domain_score > 0,
            similarity.location_score > 0.3,
            similarity.strategic_score > 0
        ].filter(Boolean).length;
        confidence += alignedDimensions * 0.1;
        return Math.min(confidence, 1.0);
    }
    determineRelationshipType(item1, item2, similarity) {
        // Determine relationship based on content analysis and metadata
        const content1 = item1.content.toLowerCase();
        const content2 = item2.content.toLowerCase();
        // Check for continuation patterns
        if (content1.includes('continue') || content1.includes('follow up') ||
            content2.includes('continue') || content2.includes('follow up')) {
            return 'continuation';
        }
        // Check for conflict patterns
        if (content1.includes('instead') || content1.includes('alternative') ||
            content2.includes('instead') || content2.includes('alternative')) {
            return 'conflict';
        }
        // Check for dependency patterns
        if (content1.includes('depends on') || content1.includes('requires') ||
            content2.includes('depends on') || content2.includes('requires')) {
            return 'dependency';
        }
        // Default to related
        return 'related';
    }
    createSuggestion(activeItem, reference) {
        const historicalItem = this.workStateManager.getHistoricalItem(reference.target_item_id);
        if (!historicalItem)
            return null;
        switch (reference.relationship_type) {
            case 'continuation':
                return {
                    type: 'continue_work',
                    priority: 'high',
                    message: `Consider reviewing previous work: "${historicalItem.content.slice(0, 80)}..." before continuing`,
                    target_item_id: reference.target_item_id,
                    confidence: reference.confidence,
                    action_hint: 'Use promote_to_active() to bring this context into current work'
                };
            case 'conflict':
                return {
                    type: 'review_conflict',
                    priority: 'high',
                    message: `Potential conflict with previous decision: "${historicalItem.content.slice(0, 80)}..."`,
                    target_item_id: reference.target_item_id,
                    confidence: reference.confidence,
                    action_hint: 'Review the historical item to ensure consistency'
                };
            case 'dependency':
                return {
                    type: 'promote_historical',
                    priority: 'medium',
                    message: `Current work may depend on: "${historicalItem.content.slice(0, 80)}..."`,
                    target_item_id: reference.target_item_id,
                    confidence: reference.confidence,
                    action_hint: 'Consider if this dependency affects current implementation'
                };
            default: // 'related'
                return {
                    type: 'reference_decision',
                    priority: 'low',
                    message: `Related previous work: "${historicalItem.content.slice(0, 80)}..."`,
                    target_item_id: reference.target_item_id,
                    confidence: reference.confidence,
                    action_hint: 'May provide useful context or lessons learned'
                };
        }
    }
    updateItemReferences(item, references) {
        // Update the item's metadata with new references
        if (!item.metadata) {
            item.metadata = {};
        }
        item.metadata.smart_references = references.map(ref => ({
            target_id: ref.target_item_id,
            similarity_score: ref.similarity_score,
            relationship_type: ref.relationship_type,
            confidence: ref.confidence,
            auto_generated: ref.auto_generated
        }));
        // Save the updated item
        this.workStateManager.saveWorkItem(item);
    }
    addCrossReference(sourceItem, targetItem, similarity) {
        const reference = {
            id: this.generateId(),
            target_item_id: targetItem.id,
            similarity_score: similarity.total_score,
            confidence: this.calculateConfidence(similarity),
            relationship_type: this.determineRelationshipType(sourceItem, targetItem, similarity),
            auto_generated: true,
            created_at: new Date().toISOString(),
            metadata: {
                common_keywords: similarity.common_keywords,
                domain_overlap: similarity.domain_overlap,
                code_location_overlap: similarity.code_location_overlap,
                strategic_alignment: similarity.strategic_alignment
            }
        };
        this.updateItemReferences(sourceItem, [reference]);
    }
    generateId() {
        return `${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    }
}
exports.SmartReferenceEngine = SmartReferenceEngine;
//# sourceMappingURL=SmartReferenceEngine.js.map