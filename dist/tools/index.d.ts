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
    private handleGetCrossWorktreeStatus;
    private handleLoadWorkState;
    private handleSaveWorkState;
}
//# sourceMappingURL=index.d.ts.map