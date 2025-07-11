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
    localWorkDir = (0, path_1.join)(process.cwd(), '.claude-work');
    historyDir = (0, path_1.join)(this.localWorkDir, 'history');
    activeDir = (0, path_1.join)(this.localWorkDir, 'active');
    futureDir = (0, path_1.join)(this.localWorkDir, 'future');
    futureItemsDir = (0, path_1.join)(this.futureDir, 'items');
    futureGroupsDir = (0, path_1.join)(this.futureDir, 'groups');
    futureSuggestionsFile = (0, path_1.join)(this.futureDir, 'suggestions.json');
    constructor() {
        this.ensureDirectories();
    }
    ensureDirectories() {
        [this.baseDir, this.todosDir, this.findingsDir, this.workStateDir, this.projectsDir, this.localWorkDir, this.historyDir, this.activeDir, this.futureDir, this.futureItemsDir, this.futureGroupsDir].forEach(dir => {
            if (!(0, fs_1.existsSync)(dir)) {
                (0, fs_1.mkdirSync)(dir, { recursive: true });
            }
        });
        // Ensure suggestions file exists
        if (!(0, fs_1.existsSync)(this.futureSuggestionsFile)) {
            this.writeJsonFile(this.futureSuggestionsFile, {
                last_updated: new Date().toISOString(),
                grouping_suggestions: [],
                similarity_analysis: {
                    feature_clusters: [],
                    technical_domains: [],
                    code_locations: []
                },
                auto_grouping_enabled: true,
                similarity_threshold: 0.7
            });
        }
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
    queryHistory(keyword, startDate, endDate, type) {
        try {
            const files = (0, fs_1.readdirSync)(this.historyDir)
                .filter(file => file.endsWith('.json'))
                .filter(file => {
                if (type) {
                    return file.includes(type);
                }
                return true;
            })
                .filter(file => {
                if (startDate || endDate) {
                    const dateMatch = file.match(/(\d{4}-\d{2}-\d{2})/);
                    if (dateMatch) {
                        const fileDate = dateMatch[1];
                        if (startDate && fileDate < startDate)
                            return false;
                        if (endDate && fileDate > endDate)
                            return false;
                    }
                }
                return true;
            });
            const items = [];
            for (const file of files) {
                const item = this.readJsonFile((0, path_1.join)(this.historyDir, file));
                if (item && item.content.toLowerCase().includes(keyword.toLowerCase())) {
                    items.push(item);
                }
            }
            return items.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
        }
        catch (error) {
            console.error('Error querying history:', error);
            return [];
        }
    }
    getHistoricalItem(itemId) {
        try {
            // Try to find by item ID first
            const files = (0, fs_1.readdirSync)(this.historyDir)
                .filter(file => file.endsWith('.json'));
            for (const file of files) {
                const item = this.readJsonFile((0, path_1.join)(this.historyDir, file));
                if (item && (item.id === itemId || file.includes(itemId))) {
                    return item;
                }
            }
            // Try to find by filename
            const directPath = (0, path_1.join)(this.historyDir, itemId.endsWith('.json') ? itemId : `${itemId}.json`);
            if ((0, fs_1.existsSync)(directPath)) {
                return this.readJsonFile(directPath);
            }
            return null;
        }
        catch (error) {
            console.error('Error getting historical item:', error);
            return null;
        }
    }
    summarizePeriod(startDate, endDate) {
        try {
            const files = (0, fs_1.readdirSync)(this.historyDir)
                .filter(file => file.endsWith('.json'))
                .filter(file => {
                const dateMatch = file.match(/(\d{4}-\d{2}-\d{2})/);
                if (dateMatch) {
                    const fileDate = dateMatch[1];
                    return fileDate >= startDate && fileDate <= endDate;
                }
                return false;
            });
            const items = [];
            for (const file of files) {
                const item = this.readJsonFile((0, path_1.join)(this.historyDir, file));
                if (item) {
                    items.push(item);
                }
            }
            const summary = {
                period: { start: startDate, end: endDate },
                total_items: items.length,
                by_type: {
                    plans: items.filter(i => i.type === 'plan').length,
                    proposals: items.filter(i => i.type === 'proposal').length,
                    findings: items.filter(i => i.type === 'finding').length,
                    todos: items.filter(i => i.type === 'todo').length
                },
                by_status: {
                    completed: items.filter(i => i.status === 'completed').length,
                    pending: items.filter(i => i.status === 'pending').length,
                    in_progress: items.filter(i => i.status === 'in_progress').length
                },
                key_items: items.slice(0, 5).map(i => ({
                    id: i.id,
                    type: i.type,
                    content: i.content.slice(0, 100) + '...',
                    timestamp: i.timestamp
                }))
            };
            return summary;
        }
        catch (error) {
            console.error('Error summarizing period:', error);
            return null;
        }
    }
    promoteToActive(itemId) {
        try {
            const historicalItem = this.getHistoricalItem(itemId);
            if (!historicalItem) {
                return null;
            }
            // Create active version
            const activeItem = {
                ...historicalItem,
                status: 'pending',
                promoted_from_history: true,
                promoted_at: new Date().toISOString()
            };
            // Save to active directory
            const activeFilePath = (0, path_1.join)(this.activeDir, `${activeItem.id}.json`);
            this.writeJsonFile(activeFilePath, activeItem);
            return activeItem;
        }
        catch (error) {
            console.error('Error promoting to active:', error);
            return null;
        }
    }
    archiveActiveItem(itemId) {
        try {
            const activeFilePath = (0, path_1.join)(this.activeDir, `${itemId}.json`);
            const activeItem = this.readJsonFile(activeFilePath);
            if (!activeItem) {
                return null;
            }
            // Add archival metadata
            const archivedItem = {
                ...activeItem,
                archived_at: new Date().toISOString(),
                archived_from: 'active'
            };
            // Save to history
            const historyFilePath = (0, path_1.join)(this.historyDir, `${new Date().toISOString().split('T')[0]}-${itemId}.json`);
            this.writeJsonFile(historyFilePath, archivedItem);
            // Remove from active
            const fs = require('fs');
            if ((0, fs_1.existsSync)(activeFilePath)) {
                fs.unlinkSync(activeFilePath);
            }
            return archivedItem;
        }
        catch (error) {
            console.error('Error archiving active item:', error);
            return null;
        }
    }
    deferToFuture(content, reason, originalType = 'idea') {
        try {
            const futureWorkItem = {
                id: this.generateId(),
                type: 'future_work',
                original_type: originalType,
                content: content,
                similarity_metadata: this.extractSimilarityMetadata(content),
                context: {
                    deprioritized_from: 'active',
                    deprioritized_date: new Date().toISOString(),
                    deprioritized_reason: reason,
                    suggested_group: null
                },
                grouping_status: 'ungrouped',
                priority_when_promoted: 'medium',
                created_at: new Date().toISOString()
            };
            // Save to items directory
            const itemFilePath = (0, path_1.join)(this.futureItemsDir, `item-${futureWorkItem.id}.json`);
            this.writeJsonFile(itemFilePath, futureWorkItem);
            // Update suggestions with new item
            this.updateGroupingSuggestions(futureWorkItem);
            return futureWorkItem;
        }
        catch (error) {
            console.error('Error deferring to future work:', error);
            throw error;
        }
    }
    listFutureGroups() {
        try {
            const groups = [];
            const ungroupedItems = [];
            // Load all groups
            if ((0, fs_1.existsSync)(this.futureGroupsDir)) {
                const groupFiles = (0, fs_1.readdirSync)(this.futureGroupsDir).filter(file => file.endsWith('.json'));
                for (const file of groupFiles) {
                    const group = this.readJsonFile((0, path_1.join)(this.futureGroupsDir, file));
                    if (group) {
                        groups.push(group);
                    }
                }
            }
            // Load ungrouped items
            if ((0, fs_1.existsSync)(this.futureItemsDir)) {
                const itemFiles = (0, fs_1.readdirSync)(this.futureItemsDir).filter(file => file.endsWith('.json'));
                for (const file of itemFiles) {
                    const item = this.readJsonFile((0, path_1.join)(this.futureItemsDir, file));
                    if (item && item.grouping_status === 'ungrouped') {
                        ungroupedItems.push(item);
                    }
                }
            }
            // Load suggestions
            const suggestions = this.readJsonFile(this.futureSuggestionsFile) || { grouping_suggestions: [] };
            return {
                groups: groups.sort((a, b) => new Date(b.last_updated).getTime() - new Date(a.last_updated).getTime()),
                ungrouped_items: ungroupedItems.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()),
                suggestions: suggestions.grouping_suggestions,
                total_items: groups.reduce((sum, group) => sum + (group.items?.length || 0), 0) + ungroupedItems.length
            };
        }
        catch (error) {
            console.error('Error listing future groups:', error);
            return { groups: [], ungrouped_items: [], suggestions: [], total_items: 0 };
        }
    }
    promoteWorkGroup(groupName) {
        try {
            const groupFilePath = (0, path_1.join)(this.futureGroupsDir, `${groupName}.json`);
            const group = this.readJsonFile(groupFilePath);
            if (!group) {
                return [];
            }
            const promotedItems = [];
            // Promote each item in the group
            for (const itemId of group.items || []) {
                const item = this.findFutureWorkItem(itemId);
                if (item) {
                    const activeItem = {
                        id: item.id,
                        type: item.original_type,
                        content: item.content,
                        status: 'pending',
                        context: this.getCurrentGitContext(),
                        session_id: this.generateId(),
                        timestamp: new Date().toISOString(),
                        metadata: {
                            promoted_from_future: true,
                            promoted_from_group: groupName,
                            priority: item.priority_when_promoted || 'medium'
                        }
                    };
                    // Save to active directory
                    const activeFilePath = (0, path_1.join)(this.activeDir, `${activeItem.id}.json`);
                    this.writeJsonFile(activeFilePath, activeItem);
                    promotedItems.push(activeItem);
                }
            }
            // Remove promoted items and group
            for (const itemId of group.items || []) {
                this.removeFutureWorkItem(itemId);
            }
            // Remove the group file
            const fs = require('fs');
            if ((0, fs_1.existsSync)(groupFilePath)) {
                fs.unlinkSync(groupFilePath);
            }
            return promotedItems;
        }
        catch (error) {
            console.error('Error promoting work group:', error);
            return [];
        }
    }
    createWorkGroup(name, description, itemIds) {
        try {
            const group = {
                id: this.generateId(),
                name: name,
                description: description,
                items: itemIds,
                similarity_score: this.calculateGroupSimilarityScore(itemIds),
                strategic_value: 'medium',
                estimated_effort: 'medium',
                readiness_status: 'ready',
                created_date: new Date().toISOString(),
                last_updated: new Date().toISOString()
            };
            // Save group
            const groupFilePath = (0, path_1.join)(this.futureGroupsDir, `${name.toLowerCase().replace(/\s+/g, '-')}.json`);
            this.writeJsonFile(groupFilePath, group);
            // Update items to mark them as grouped
            for (const itemId of itemIds) {
                const itemPath = this.getFutureWorkItemPath(itemId);
                if (itemPath) {
                    const item = this.readJsonFile(itemPath);
                    if (item) {
                        item.grouping_status = 'grouped';
                        item.context.suggested_group = name;
                        this.writeJsonFile(itemPath, item);
                    }
                }
            }
            return group;
        }
        catch (error) {
            console.error('Error creating work group:', error);
            throw error;
        }
    }
    groomFutureWork() {
        try {
            const analysis = {
                overview: this.listFutureGroups(),
                suggestions: this.generateGroupingSuggestions(),
                similarity_analysis: this.analyzeSimilarityPatterns(),
                recommendations: this.generateRecommendations()
            };
            // Update suggestions file
            const suggestions = this.readJsonFile(this.futureSuggestionsFile) || {};
            suggestions.last_updated = new Date().toISOString();
            suggestions.grouping_suggestions = analysis.suggestions;
            suggestions.similarity_analysis = analysis.similarity_analysis;
            this.writeJsonFile(this.futureSuggestionsFile, suggestions);
            return analysis;
        }
        catch (error) {
            console.error('Error grooming future work:', error);
            throw error;
        }
    }
    extractSimilarityMetadata(content) {
        // Simple similarity metadata extraction based on content analysis
        const keywords = this.extractKeywords(content);
        const featureDomain = this.inferFeatureDomain(content, keywords);
        const technicalDomain = this.inferTechnicalDomain(content, keywords);
        const codeLocations = this.inferCodeLocations(content, keywords);
        const strategicTheme = this.inferStrategicTheme(content, keywords);
        return {
            keywords: keywords,
            feature_domain: featureDomain,
            technical_domain: technicalDomain,
            code_locations: codeLocations,
            strategic_theme: strategicTheme
        };
    }
    extractKeywords(content) {
        // Simple keyword extraction
        const words = content.toLowerCase()
            .replace(/[^\w\s]/g, ' ')
            .split(/\s+/)
            .filter(word => word.length > 3);
        // Remove common words
        const stopWords = ['this', 'that', 'with', 'from', 'they', 'been', 'have', 'their', 'said', 'each', 'which'];
        return words.filter(word => !stopWords.includes(word)).slice(0, 10);
    }
    inferFeatureDomain(content, keywords) {
        // Simple feature domain inference
        if (this.containsAny(content, ['auth', 'login', 'user', 'password', 'signup']))
            return 'user-management';
        if (this.containsAny(content, ['search', 'filter', 'find', 'query', 'sort']))
            return 'search-and-filtering';
        if (this.containsAny(content, ['profile', 'settings', 'preferences', 'account']))
            return 'user-profile';
        if (this.containsAny(content, ['payment', 'billing', 'subscription', 'checkout']))
            return 'payments';
        if (this.containsAny(content, ['report', 'analytics', 'dashboard', 'metrics']))
            return 'reporting';
        if (this.containsAny(content, ['performance', 'optimization', 'speed', 'cache']))
            return 'performance';
        return 'general';
    }
    inferTechnicalDomain(content, keywords) {
        if (this.containsAny(content, ['frontend', 'ui', 'component', 'react', 'vue', 'angular']))
            return 'frontend';
        if (this.containsAny(content, ['backend', 'api', 'server', 'endpoint', 'database']))
            return 'backend-api';
        if (this.containsAny(content, ['database', 'sql', 'migration', 'schema', 'query']))
            return 'database';
        if (this.containsAny(content, ['testing', 'test', 'unit', 'integration', 'e2e']))
            return 'testing';
        if (this.containsAny(content, ['deployment', 'deploy', 'infrastructure', 'docker', 'ci/cd']))
            return 'infrastructure';
        return 'general';
    }
    inferCodeLocations(content, keywords) {
        const locations = [];
        if (this.containsAny(content, ['auth', 'login', 'user']))
            locations.push('src/auth/');
        if (this.containsAny(content, ['api', 'endpoint', 'route']))
            locations.push('src/api/');
        if (this.containsAny(content, ['component', 'ui', 'frontend']))
            locations.push('src/components/');
        if (this.containsAny(content, ['database', 'model', 'schema']))
            locations.push('src/models/');
        return locations;
    }
    inferStrategicTheme(content, keywords) {
        if (this.containsAny(content, ['user', 'experience', 'usability', 'interface']))
            return 'user-experience';
        if (this.containsAny(content, ['performance', 'speed', 'optimization', 'efficiency']))
            return 'performance';
        if (this.containsAny(content, ['security', 'auth', 'secure', 'protection']))
            return 'security';
        if (this.containsAny(content, ['analytics', 'reporting', 'metrics', 'insights']))
            return 'business-intelligence';
        if (this.containsAny(content, ['developer', 'tooling', 'testing', 'documentation']))
            return 'developer-experience';
        return 'general';
    }
    containsAny(text, words) {
        const lowerText = text.toLowerCase();
        return words.some(word => lowerText.includes(word));
    }
    findFutureWorkItem(itemId) {
        // Search in items directory
        if ((0, fs_1.existsSync)(this.futureItemsDir)) {
            const files = (0, fs_1.readdirSync)(this.futureItemsDir).filter(file => file.endsWith('.json'));
            for (const file of files) {
                const item = this.readJsonFile((0, path_1.join)(this.futureItemsDir, file));
                if (item && item.id === itemId) {
                    return item;
                }
            }
        }
        return null;
    }
    getFutureWorkItemPath(itemId) {
        // Search in items directory
        if ((0, fs_1.existsSync)(this.futureItemsDir)) {
            const files = (0, fs_1.readdirSync)(this.futureItemsDir).filter(file => file.endsWith('.json'));
            for (const file of files) {
                const item = this.readJsonFile((0, path_1.join)(this.futureItemsDir, file));
                if (item && item.id === itemId) {
                    return (0, path_1.join)(this.futureItemsDir, file);
                }
            }
        }
        return null;
    }
    updateGroupingSuggestions(newItem) {
        try {
            const suggestions = this.readJsonFile(this.futureSuggestionsFile) || { grouping_suggestions: [] };
            // Find potential groups for this item
            const potentialGroups = this.findPotentialGroups(newItem);
            if (potentialGroups.length > 0) {
                const suggestion = {
                    item_id: newItem.id,
                    suggested_groups: potentialGroups,
                    confidence: this.calculateSuggestionConfidence(newItem, potentialGroups),
                    created_at: new Date().toISOString()
                };
                suggestions.grouping_suggestions.push(suggestion);
                suggestions.last_updated = new Date().toISOString();
                this.writeJsonFile(this.futureSuggestionsFile, suggestions);
            }
        }
        catch (error) {
            console.error('Error updating grouping suggestions:', error);
        }
    }
    findPotentialGroups(item) {
        const potentialGroups = [];
        // Check existing groups for similarity
        if ((0, fs_1.existsSync)(this.futureGroupsDir)) {
            const groupFiles = (0, fs_1.readdirSync)(this.futureGroupsDir).filter(file => file.endsWith('.json'));
            for (const file of groupFiles) {
                const group = this.readJsonFile((0, path_1.join)(this.futureGroupsDir, file));
                if (group && this.calculateItemGroupSimilarity(item, group) > 0.6) {
                    potentialGroups.push(group.name);
                }
            }
        }
        return potentialGroups;
    }
    calculateItemGroupSimilarity(item, group) {
        // Simple similarity calculation based on metadata overlap
        let similarity = 0;
        const itemMeta = item.similarity_metadata || {};
        // Check if any group items share similar metadata
        // This is a simplified version - in practice would be more sophisticated
        if (itemMeta.feature_domain && group.name.toLowerCase().includes(itemMeta.feature_domain.replace('-', ' '))) {
            similarity += 0.5;
        }
        if (itemMeta.technical_domain && group.description?.toLowerCase().includes(itemMeta.technical_domain)) {
            similarity += 0.3;
        }
        return Math.min(similarity, 1.0);
    }
    calculateSuggestionConfidence(item, groups) {
        // Simple confidence calculation
        return groups.length > 0 ? 0.8 : 0.3;
    }
    calculateGroupSimilarityScore(itemIds) {
        // Simple group similarity calculation
        return itemIds.length > 1 ? 0.8 : 1.0;
    }
    generateGroupingSuggestions() {
        // Generate intelligent grouping suggestions
        const suggestions = [];
        // Find ungrouped items that could be grouped together
        const ungroupedItems = this.getUngroupedItems();
        const clusters = this.clusterItemsBySimlarity(ungroupedItems);
        for (const cluster of clusters) {
            if (cluster.items.length > 1) {
                suggestions.push({
                    suggested_group_name: cluster.theme,
                    items: cluster.items.map((item) => item.id),
                    similarity_score: cluster.similarity_score,
                    rationale: cluster.rationale
                });
            }
        }
        return suggestions;
    }
    analyzeSimilarityPatterns() {
        // Analyze patterns in future work items
        const allItems = this.getAllFutureWorkItems();
        const featureClusters = this.groupItemsByFeatureDomain(allItems);
        const technicalDomains = this.groupItemsByTechnicalDomain(allItems);
        const codeLocations = this.groupItemsByCodeLocation(allItems);
        return {
            feature_clusters: featureClusters,
            technical_domains: technicalDomains,
            code_locations: codeLocations
        };
    }
    generateRecommendations() {
        // Generate strategic recommendations
        const recommendations = [];
        const overview = this.listFutureGroups();
        if (overview.ungrouped_items.length > 3) {
            recommendations.push(`Consider grouping ${overview.ungrouped_items.length} ungrouped items`);
        }
        if (overview.groups.length === 0) {
            recommendations.push('Create your first work group to organize future items');
        }
        return recommendations;
    }
    // Helper methods for analysis
    getUngroupedItems() {
        const items = [];
        if ((0, fs_1.existsSync)(this.futureItemsDir)) {
            const files = (0, fs_1.readdirSync)(this.futureItemsDir).filter(file => file.endsWith('.json'));
            for (const file of files) {
                const item = this.readJsonFile((0, path_1.join)(this.futureItemsDir, file));
                if (item && item.grouping_status === 'ungrouped') {
                    items.push(item);
                }
            }
        }
        return items;
    }
    clusterItemsBySimlarity(items) {
        // Simple clustering - group by feature domain
        const clusters = {};
        for (const item of items) {
            const domain = item.similarity_metadata?.feature_domain || 'general';
            if (!clusters[domain]) {
                clusters[domain] = [];
            }
            clusters[domain].push(item);
        }
        return Object.entries(clusters).map(([domain, items]) => ({
            theme: domain.replace('-', ' ').replace(/\b\w/g, (l) => l.toUpperCase()),
            items: items,
            similarity_score: items.length > 1 ? 0.8 : 0.3,
            rationale: `Items related to ${domain.replace('-', ' ')}`
        }));
    }
    getAllFutureWorkItems() {
        const items = [];
        if ((0, fs_1.existsSync)(this.futureItemsDir)) {
            const files = (0, fs_1.readdirSync)(this.futureItemsDir).filter(file => file.endsWith('.json'));
            for (const file of files) {
                const item = this.readJsonFile((0, path_1.join)(this.futureItemsDir, file));
                if (item) {
                    items.push(item);
                }
            }
        }
        return items;
    }
    groupItemsByFeatureDomain(items) {
        const domains = {};
        for (const item of items) {
            const domain = item.similarity_metadata?.feature_domain || 'general';
            domains[domain] = (domains[domain] || 0) + 1;
        }
        return domains;
    }
    groupItemsByTechnicalDomain(items) {
        const domains = {};
        for (const item of items) {
            const domain = item.similarity_metadata?.technical_domain || 'general';
            domains[domain] = (domains[domain] || 0) + 1;
        }
        return domains;
    }
    groupItemsByCodeLocation(items) {
        const locations = {};
        for (const item of items) {
            const itemLocations = item.similarity_metadata?.code_locations || [];
            for (const location of itemLocations) {
                locations[location] = (locations[location] || 0) + 1;
            }
        }
        return locations;
    }
    removeFutureWorkItem(itemId) {
        const filePath = this.getFutureWorkItemPath(itemId);
        if (filePath && (0, fs_1.existsSync)(filePath)) {
            const fs = require('fs');
            fs.unlinkSync(filePath);
        }
    }
}
exports.WorkStateManager = WorkStateManager;
//# sourceMappingURL=WorkStateManager.js.map