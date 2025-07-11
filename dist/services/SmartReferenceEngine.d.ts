import { WorkItem, SimilarityScore } from '../types';
import { WorkStateManager } from './WorkStateManager';
export interface SmartReference {
    id: string;
    target_item_id: string;
    similarity_score: number;
    confidence: number;
    relationship_type: 'related' | 'continuation' | 'conflict' | 'dependency';
    auto_generated: boolean;
    created_at: string;
    metadata: {
        common_keywords: string[];
        domain_overlap: string[];
        code_location_overlap: string[];
        strategic_alignment: string;
    };
}
export interface ContextualSuggestion {
    type: 'promote_historical' | 'review_conflict' | 'reference_decision' | 'continue_work';
    priority: 'high' | 'medium' | 'low';
    message: string;
    target_item_id: string;
    confidence: number;
    action_hint?: string;
}
export declare class SmartReferenceEngine {
    private workStateManager;
    private readonly similarityThreshold;
    private readonly confidenceThreshold;
    constructor(workStateManager: WorkStateManager);
    /**
     * Generate automatic references for a given work item
     */
    generateAutomaticReferences(activeItem: WorkItem): SmartReference[];
    /**
     * Calculate semantic similarity between two work items
     */
    calculateSemanticSimilarity(item1: WorkItem, item2: WorkItem): SimilarityScore;
    /**
     * Get contextual suggestions based on current active work
     */
    getContextualSuggestions(activeItems: WorkItem[]): ContextualSuggestion[];
    /**
     * Update references when a work item changes
     */
    updateReferencesOnChange(itemId: string): void;
    private getRelevantHistoricalItems;
    private calculateContentSimilarity;
    private extractWords;
    private calculateConfidence;
    private determineRelationshipType;
    private createSuggestion;
    private updateItemReferences;
    private addCrossReference;
    private generateId;
}
//# sourceMappingURL=SmartReferenceEngine.d.ts.map