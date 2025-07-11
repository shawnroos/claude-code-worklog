import { WorkItem, Finding, WorkState, SessionSummary } from '../types';
import { ContextualSuggestion } from './SmartReferenceEngine';
import { ReferenceMap } from './ReferenceMapper';
export declare class WorkStateManager {
    private readonly baseDir;
    private readonly todosDir;
    private readonly findingsDir;
    private readonly workStateDir;
    private readonly projectsDir;
    private readonly localWorkDir;
    private readonly historyDir;
    private readonly activeDir;
    private readonly futureDir;
    private readonly futureItemsDir;
    private readonly futureGroupsDir;
    private readonly futureSuggestionsFile;
    private smartReferenceEngine;
    private referenceMapper;
    constructor();
    private ensureDirectories;
    private getCurrentGitContext;
    private readJsonFile;
    private writeJsonFile;
    private generateId;
    getCurrentWorkState(): WorkState;
    loadActiveTodos(): WorkItem[];
    private loadRecentFindings;
    private generateSessionSummary;
    saveWorkItem(workItem: WorkItem): void;
    saveFinding(finding: Finding): void;
    getWorkItemsByType(type: WorkItem['type']): WorkItem[];
    searchWorkItems(query: string): WorkItem[];
    getSessionSummary(sessionId: string): SessionSummary | null;
    savePlan(content: string, steps: string[]): WorkItem;
    saveProposal(content: string, rationale: string): WorkItem;
    getCrossWorktreeConflicts(): string[];
    queryHistory(keyword: string, startDate?: string, endDate?: string, type?: string): WorkItem[];
    getHistoricalItem(itemId: string): WorkItem | null;
    summarizePeriod(startDate: string, endDate: string): any;
    promoteToActive(itemId: string): WorkItem | null;
    archiveActiveItem(itemId: string): WorkItem | null;
    deferToFuture(content: string, reason: string, originalType?: string): any;
    listFutureGroups(): any;
    promoteWorkGroup(groupName: string): WorkItem[];
    createWorkGroup(name: string, description: string, itemIds: string[]): any;
    groomFutureWork(): any;
    private extractSimilarityMetadata;
    private extractKeywords;
    private inferFeatureDomain;
    private inferTechnicalDomain;
    private inferCodeLocations;
    private inferStrategicTheme;
    private containsAny;
    private findFutureWorkItem;
    private getFutureWorkItemPath;
    private updateGroupingSuggestions;
    private findPotentialGroups;
    private calculateItemGroupSimilarity;
    private calculateSuggestionConfidence;
    private calculateGroupSimilarityScore;
    private generateGroupingSuggestions;
    private analyzeSimilarityPatterns;
    private generateRecommendations;
    private getUngroupedItems;
    private clusterItemsBySimlarity;
    private getAllFutureWorkItems;
    private groupItemsByFeatureDomain;
    private groupItemsByTechnicalDomain;
    private groupItemsByCodeLocation;
    private removeFutureWorkItem;
    /**
     * Get contextual suggestions for current active work
     */
    getContextualSuggestions(): ContextualSuggestion[];
    /**
     * Generate smart references for a specific work item
     */
    generateSmartReferences(itemId: string): any[];
    /**
     * Calculate similarity between two work items
     */
    calculateSimilarity(itemId1: string, itemId2: string): any;
    /**
     * Get enhanced work state with smart referencing context
     */
    getEnhancedWorkState(): any;
    private findWorkItem;
    private groupSuggestionsByType;
    /**
     * Generate complete reference map for current work context
     */
    generateReferenceMap(): ReferenceMap;
    /**
     * Generate focused reference map for a specific work item
     */
    generateFocusedReferenceMap(itemId: string, depth?: number): ReferenceMap;
    /**
     * Find reference path between two work items
     */
    findReferencePath(sourceId: string, targetId: string): string[];
    /**
     * Generate ASCII visualization of reference relationships
     */
    visualizeReferences(): string;
}
//# sourceMappingURL=WorkStateManager.d.ts.map