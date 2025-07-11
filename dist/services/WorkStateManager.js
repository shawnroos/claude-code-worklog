"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.WorkStateManager = void 0;
const fs_1 = require("fs");
const path_1 = require("path");
const child_process_1 = require("child_process");
class WorkStateManager {
    baseDir = (0, path_1.join)(process.env.HOME || '', '.claude');
    todosDir = (0, path_1.join)(this.baseDir, 'todos');
    findingsDir = (0, path_1.join)(this.baseDir, 'findings');
    workStateDir = (0, path_1.join)(this.baseDir, 'work-state');
    projectsDir = (0, path_1.join)(this.workStateDir, 'projects');
    constructor() {
        this.ensureDirectories();
    }
    ensureDirectories() {
        [this.baseDir, this.todosDir, this.findingsDir, this.workStateDir, this.projectsDir].forEach(dir => {
            if (!(0, fs_1.existsSync)(dir)) {
                (0, fs_1.mkdirSync)(dir, { recursive: true });
            }
        });
    }
    getCurrentGitContext() {
        try {
            const cwd = process.cwd();
            const branch = (0, child_process_1.execSync)('git branch --show-current', { cwd, encoding: 'utf8' }).trim();
            const worktreeList = (0, child_process_1.execSync)('git worktree list --porcelain', { cwd, encoding: 'utf8' });
            const currentWorktree = worktreeList
                .split('\n')
                .find(line => line.startsWith(`worktree ${cwd}`));
            const worktree = currentWorktree ? 'feature' : 'main';
            const remoteUrl = (0, child_process_1.execSync)('git remote get-url origin', { cwd, encoding: 'utf8' }).trim();
            return {
                branch,
                worktree,
                remote_url: remoteUrl,
                working_directory: cwd
            };
        }
        catch (error) {
            return {
                branch: 'unknown',
                worktree: 'unknown',
                working_directory: process.cwd()
            };
        }
    }
    readJsonFile(filePath) {
        try {
            if (!(0, fs_1.existsSync)(filePath))
                return null;
            const content = (0, fs_1.readFileSync)(filePath, 'utf8');
            return JSON.parse(content);
        }
        catch (error) {
            console.error(`Error reading ${filePath}:`, error);
            return null;
        }
    }
    writeJsonFile(filePath, data) {
        try {
            const dir = (0, path_1.dirname)(filePath);
            if (!(0, fs_1.existsSync)(dir)) {
                (0, fs_1.mkdirSync)(dir, { recursive: true });
            }
            (0, fs_1.writeFileSync)(filePath, JSON.stringify(data, null, 2));
        }
        catch (error) {
            console.error(`Error writing ${filePath}:`, error);
        }
    }
    generateId() {
        return `${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    }
    getCurrentWorkState() {
        const gitContext = this.getCurrentGitContext();
        const sessionId = this.generateId();
        // Load active todos
        const activeTodos = this.loadActiveTodos();
        // Load recent findings
        const recentFindings = this.loadRecentFindings();
        // Generate session summary
        const sessionSummary = this.generateSessionSummary(sessionId, gitContext, activeTodos, recentFindings);
        return {
            current_session: sessionId,
            active_todos: activeTodos,
            recent_findings: recentFindings,
            session_summary: sessionSummary
        };
    }
    loadActiveTodos() {
        try {
            // Look for pending todos in current working directory
            const cwd = process.cwd();
            const pendingTodosPath = (0, path_1.join)(cwd, '.claude-work', 'PENDING_TODOS.json');
            if ((0, fs_1.existsSync)(pendingTodosPath)) {
                const pendingTodos = this.readJsonFile(pendingTodosPath);
                if (pendingTodos) {
                    return pendingTodos.map(todo => ({
                        id: todo.id || this.generateId(),
                        type: 'todo',
                        content: todo.content,
                        status: todo.status || 'pending',
                        context: this.getCurrentGitContext(),
                        session_id: todo.session_id || '',
                        timestamp: todo.saved_at || new Date().toISOString(),
                        metadata: {
                            priority: todo.priority || 'medium'
                        }
                    }));
                }
            }
            return [];
        }
        catch (error) {
            console.error('Error loading active todos:', error);
            return [];
        }
    }
    loadRecentFindings() {
        try {
            const findings = [];
            const files = (0, fs_1.readdirSync)(this.findingsDir)
                .filter(file => file.endsWith('.json'))
                .sort((a, b) => {
                const aTime = (0, fs_1.statSync)((0, path_1.join)(this.findingsDir, a)).mtime.getTime();
                const bTime = (0, fs_1.statSync)((0, path_1.join)(this.findingsDir, b)).mtime.getTime();
                return bTime - aTime;
            })
                .slice(0, 20); // Get 20 most recent
            for (const file of files) {
                const finding = this.readJsonFile((0, path_1.join)(this.findingsDir, file));
                if (finding) {
                    findings.push(finding);
                }
            }
            return findings;
        }
        catch (error) {
            console.error('Error loading recent findings:', error);
            return [];
        }
    }
    generateSessionSummary(sessionId, gitContext, todos, findings) {
        const completedTodos = todos.filter(t => t.status === 'completed').length;
        const pendingTodos = todos.filter(t => t.status !== 'completed').length;
        const plansCreated = todos.filter(t => t.type === 'plan').length;
        const proposalsMade = todos.filter(t => t.type === 'proposal').length;
        return {
            session_id: sessionId,
            timestamp: new Date().toISOString(),
            git_context: gitContext,
            completed_todos: completedTodos,
            pending_todos: pendingTodos,
            findings_count: findings.length,
            plans_created: plansCreated,
            proposals_made: proposalsMade,
            key_decisions: [],
            outcomes: []
        };
    }
    saveWorkItem(workItem) {
        const filePath = (0, path_1.join)(this.todosDir, `${workItem.id}.json`);
        this.writeJsonFile(filePath, workItem);
    }
    saveFinding(finding) {
        const filePath = (0, path_1.join)(this.findingsDir, `${finding.id}.json`);
        this.writeJsonFile(filePath, finding);
    }
    getWorkItemsByType(type) {
        try {
            const files = (0, fs_1.readdirSync)(this.todosDir)
                .filter(file => file.endsWith('.json'));
            const items = [];
            for (const file of files) {
                const item = this.readJsonFile((0, path_1.join)(this.todosDir, file));
                if (item && item.type === type) {
                    items.push(item);
                }
            }
            return items.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
        }
        catch (error) {
            console.error('Error loading work items by type:', error);
            return [];
        }
    }
    searchWorkItems(query) {
        try {
            const files = (0, fs_1.readdirSync)(this.todosDir)
                .filter(file => file.endsWith('.json'));
            const items = [];
            for (const file of files) {
                const item = this.readJsonFile((0, path_1.join)(this.todosDir, file));
                if (item && item.content.toLowerCase().includes(query.toLowerCase())) {
                    items.push(item);
                }
            }
            return items;
        }
        catch (error) {
            console.error('Error searching work items:', error);
            return [];
        }
    }
    getSessionSummary(sessionId) {
        // This would look for a session summary file
        const summaryPath = (0, path_1.join)(this.todosDir, `${sessionId}_summary.json`);
        return this.readJsonFile(summaryPath);
    }
    savePlan(content, steps) {
        const gitContext = this.getCurrentGitContext();
        const planItem = {
            id: this.generateId(),
            type: 'plan',
            content,
            status: 'pending',
            context: gitContext,
            session_id: this.generateId(),
            timestamp: new Date().toISOString(),
            metadata: {
                plan_steps: steps,
                priority: 'high'
            }
        };
        this.saveWorkItem(planItem);
        return planItem;
    }
    saveProposal(content, rationale) {
        const gitContext = this.getCurrentGitContext();
        const proposalItem = {
            id: this.generateId(),
            type: 'proposal',
            content,
            status: 'pending',
            context: gitContext,
            session_id: this.generateId(),
            timestamp: new Date().toISOString(),
            metadata: {
                decision_rationale: rationale,
                priority: 'high'
            }
        };
        this.saveWorkItem(proposalItem);
        return proposalItem;
    }
    getCrossWorktreeConflicts() {
        // Cross-worktree functionality disabled to reduce context overhead
        return [];
    }
}
exports.WorkStateManager = WorkStateManager;
//# sourceMappingURL=WorkStateManager.js.map