import { WorkItem, Finding, WorkState, SessionSummary } from '../types';
export declare class WorkStateManager {
    private readonly baseDir;
    private readonly todosDir;
    private readonly findingsDir;
    private readonly workStateDir;
    private readonly projectsDir;
    constructor();
    private ensureDirectories;
    private getCurrentGitContext;
    private readJsonFile;
    private writeJsonFile;
    private generateId;
    getCurrentWorkState(): WorkState;
    private loadActiveTodos;
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
}
//# sourceMappingURL=WorkStateManager.d.ts.map