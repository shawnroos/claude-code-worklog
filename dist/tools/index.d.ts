import { Tool } from '@modelcontextprotocol/sdk/types.js';
import { McpToolParams, McpToolResponse } from '../types/index.js';
export declare class WorkTrackingTools {
    private workStateManager;
    getTools(): Tool[];
    handleToolCall(name: string, params: McpToolParams): Promise<McpToolResponse>;
    private handleGetWorkState;
    private handleSavePlan;
    private handleSaveProposal;
    private handleSearchWorkItems;
    private handleGetSessionSummary;
    private handleLoadWorkState;
    private handleSaveWorkState;
    private handleQueryHistory;
    private handleGetHistoricalContext;
    private handleSummarizePeriod;
    private handlePromoteToActive;
    private handleArchiveActiveItem;
    private handleDeferToFuture;
    private handleListFutureGroups;
    private handleGroomFutureWork;
    private handleCreateWorkGroup;
    private handlePromoteWorkGroup;
    private handleGetContextualSuggestions;
    private handleGenerateSmartReferences;
    private handleCalculateSimilarity;
    private handleGetEnhancedWorkState;
    private handleGenerateReferenceMap;
    private handleGenerateFocusedReferenceMap;
    private handleFindReferencePath;
    private handleVisualizeReferences;
}
//# sourceMappingURL=index.d.ts.map