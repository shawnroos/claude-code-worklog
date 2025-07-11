import { readFileSync, writeFileSync, existsSync, mkdirSync, readdirSync, statSync } from 'fs'
import { join, dirname } from 'path'
import { execSync } from 'child_process'
import { WorkItem, Finding, GitContext, WorkState, SessionSummary, SimilarityMetadata } from '../types'
import { SmartReferenceEngine, ContextualSuggestion } from './SmartReferenceEngine'
import { ReferenceMapper, ReferenceMap } from './ReferenceMapper'

export class WorkStateManager {
  private readonly baseDir = join(process.env.HOME || '', '.claude')
  private readonly todosDir = join(this.baseDir, 'todos')
  private readonly findingsDir = join(this.baseDir, 'findings')
  private readonly workStateDir = join(this.baseDir, 'work-state')
  private readonly projectsDir = join(this.workStateDir, 'projects')
  private readonly localWorkDir = join(process.cwd(), '.claude-work')
  private readonly historyDir = join(this.localWorkDir, 'history')
  private readonly activeDir = join(this.localWorkDir, 'active')
  private readonly futureDir = join(this.localWorkDir, 'future')
  private readonly futureItemsDir = join(this.futureDir, 'items')
  private readonly futureGroupsDir = join(this.futureDir, 'groups')
  private readonly futureSuggestionsFile = join(this.futureDir, 'suggestions.json')
  private smartReferenceEngine: SmartReferenceEngine
  private referenceMapper: ReferenceMapper

  constructor() {
    this.ensureDirectories()
    this.smartReferenceEngine = new SmartReferenceEngine(this)
    this.referenceMapper = new ReferenceMapper(this)
  }

  private ensureDirectories(): void {
    [this.baseDir, this.todosDir, this.findingsDir, this.workStateDir, this.projectsDir, this.localWorkDir, this.historyDir, this.activeDir, this.futureDir, this.futureItemsDir, this.futureGroupsDir].forEach(dir => {
      if (!existsSync(dir)) {
        mkdirSync(dir, { recursive: true })
      }
    })
    
    // Ensure suggestions file exists
    if (!existsSync(this.futureSuggestionsFile)) {
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
      })
    }
  }

  private getCurrentGitContext(): GitContext {
    try {
      const cwd = process.cwd()
      const branch = execSync('git branch --show-current', { cwd, encoding: 'utf8' }).trim()
      const worktreeList = execSync('git worktree list --porcelain', { cwd, encoding: 'utf8' })
      const currentWorktree = worktreeList
        .split('\n')
        .find(line => line.startsWith(`worktree ${cwd}`))
      const worktree = currentWorktree ? 'feature' : 'main'
      const remoteUrl = execSync('git remote get-url origin', { cwd, encoding: 'utf8' }).trim()
      
      return {
        branch,
        worktree,
        remote_url: remoteUrl,
        working_directory: cwd
      }
    } catch (error) {
      return {
        branch: 'unknown',
        worktree: 'unknown',
        working_directory: process.cwd()
      }
    }
  }

  private readJsonFile<T>(filePath: string): T | null {
    try {
      if (!existsSync(filePath)) return null
      const content = readFileSync(filePath, 'utf8')
      return JSON.parse(content)
    } catch (error) {
      console.error(`Error reading ${filePath}:`, error)
      return null
    }
  }

  private writeJsonFile<T>(filePath: string, data: T): void {
    try {
      const dir = dirname(filePath)
      if (!existsSync(dir)) {
        mkdirSync(dir, { recursive: true })
      }
      writeFileSync(filePath, JSON.stringify(data, null, 2))
    } catch (error) {
      console.error(`Error writing ${filePath}:`, error)
    }
  }

  private generateId(): string {
    return `${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  public getCurrentWorkState(): WorkState {
    const gitContext = this.getCurrentGitContext()
    const sessionId = this.generateId()
    
    // Load active todos
    const activeTodos = this.loadActiveTodos()
    
    // Load recent findings
    const recentFindings = this.loadRecentFindings()
    
    // Generate session summary
    const sessionSummary = this.generateSessionSummary(sessionId, gitContext, activeTodos, recentFindings)
    
    return {
      current_session: sessionId,
      active_todos: activeTodos,
      recent_findings: recentFindings,
      session_summary: sessionSummary
    }
  }

  public loadActiveTodos(): WorkItem[] {
    try {
      // Look for pending todos in current working directory
      const cwd = process.cwd()
      const pendingTodosPath = join(cwd, '.claude-work', 'PENDING_TODOS.json')
      
      if (existsSync(pendingTodosPath)) {
        const pendingTodos = this.readJsonFile<any[]>(pendingTodosPath)
        if (pendingTodos) {
          return pendingTodos.map(todo => ({
            id: todo.id || this.generateId(),
            type: 'todo' as const,
            content: todo.content,
            status: todo.status || 'pending',
            context: this.getCurrentGitContext(),
            session_id: todo.session_id || '',
            timestamp: todo.saved_at || new Date().toISOString(),
            metadata: {
              priority: todo.priority || 'medium'
            }
          }))
        }
      }
      
      return []
    } catch (error) {
      console.error('Error loading active todos:', error)
      return []
    }
  }

  private loadRecentFindings(): Finding[] {
    try {
      const findings: Finding[] = []
      const files = readdirSync(this.findingsDir)
        .filter(file => file.endsWith('.json'))
        .sort((a, b) => {
          const aTime = statSync(join(this.findingsDir, a)).mtime.getTime()
          const bTime = statSync(join(this.findingsDir, b)).mtime.getTime()
          return bTime - aTime
        })
        .slice(0, 20) // Get 20 most recent
      
      for (const file of files) {
        const finding = this.readJsonFile<Finding>(join(this.findingsDir, file))
        if (finding) {
          findings.push(finding)
        }
      }
      
      return findings
    } catch (error) {
      console.error('Error loading recent findings:', error)
      return []
    }
  }

  private generateSessionSummary(sessionId: string, gitContext: GitContext, todos: WorkItem[], findings: Finding[]): SessionSummary {
    const completedTodos = todos.filter(t => t.status === 'completed').length
    const pendingTodos = todos.filter(t => t.status !== 'completed').length
    const plansCreated = todos.filter(t => t.type === 'plan').length
    const proposalsMade = todos.filter(t => t.type === 'proposal').length
    
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
    }
  }

  public saveWorkItem(workItem: WorkItem): void {
    // Ensure similarity metadata is extracted
    if (!workItem.metadata?.similarity_metadata) {
      if (!workItem.metadata) {
        workItem.metadata = {}
      }
      workItem.metadata.similarity_metadata = this.extractSimilarityMetadata(workItem.content)
    }
    
    // Generate smart references
    const smartReferences = this.smartReferenceEngine.generateAutomaticReferences(workItem)
    if (smartReferences.length > 0) {
      workItem.metadata.smart_references = smartReferences.map(ref => ({
        target_id: ref.target_item_id,
        similarity_score: ref.similarity_score,
        relationship_type: ref.relationship_type,
        confidence: ref.confidence,
        auto_generated: ref.auto_generated
      }))
    }
    
    const filePath = join(this.todosDir, `${workItem.id}.json`)
    this.writeJsonFile(filePath, workItem)
    
    // Update references in other items if needed
    this.smartReferenceEngine.updateReferencesOnChange(workItem.id)
  }

  public saveFinding(finding: Finding): void {
    const filePath = join(this.findingsDir, `${finding.id}.json`)
    this.writeJsonFile(filePath, finding)
  }

  public getWorkItemsByType(type: WorkItem['type']): WorkItem[] {
    try {
      const files = readdirSync(this.todosDir)
        .filter(file => file.endsWith('.json'))
      
      const items: WorkItem[] = []
      for (const file of files) {
        const item = this.readJsonFile<WorkItem>(join(this.todosDir, file))
        if (item && item.type === type) {
          items.push(item)
        }
      }
      
      return items.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
    } catch (error) {
      console.error('Error loading work items by type:', error)
      return []
    }
  }

  public searchWorkItems(query: string): WorkItem[] {
    try {
      const files = readdirSync(this.todosDir)
        .filter(file => file.endsWith('.json'))
      
      const items: WorkItem[] = []
      for (const file of files) {
        const item = this.readJsonFile<WorkItem>(join(this.todosDir, file))
        if (item && item.content.toLowerCase().includes(query.toLowerCase())) {
          items.push(item)
        }
      }
      
      return items
    } catch (error) {
      console.error('Error searching work items:', error)
      return []
    }
  }

  public getSessionSummary(sessionId: string): SessionSummary | null {
    // This would look for a session summary file
    const summaryPath = join(this.todosDir, `${sessionId}_summary.json`)
    return this.readJsonFile<SessionSummary>(summaryPath)
  }

  public savePlan(content: string, steps: string[]): WorkItem {
    const gitContext = this.getCurrentGitContext()
    const planItem: WorkItem = {
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
    }
    
    this.saveWorkItem(planItem)
    return planItem
  }

  public saveProposal(content: string, rationale: string): WorkItem {
    const gitContext = this.getCurrentGitContext()
    const proposalItem: WorkItem = {
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
    }
    
    this.saveWorkItem(proposalItem)
    return proposalItem
  }

  public getCrossWorktreeConflicts(): string[] {
    // Cross-worktree functionality disabled to reduce context overhead
    return []
  }

  public queryHistory(keyword: string, startDate?: string, endDate?: string, type?: string): WorkItem[] {
    try {
      const files = readdirSync(this.historyDir)
        .filter(file => file.endsWith('.json'))
        .filter(file => {
          if (type) {
            return file.includes(type)
          }
          return true
        })
        .filter(file => {
          if (startDate || endDate) {
            const dateMatch = file.match(/(\d{4}-\d{2}-\d{2})/)
            if (dateMatch) {
              const fileDate = dateMatch[1]
              if (startDate && fileDate < startDate) return false
              if (endDate && fileDate > endDate) return false
            }
          }
          return true
        })
      
      const items: WorkItem[] = []
      for (const file of files) {
        const item = this.readJsonFile<WorkItem>(join(this.historyDir, file))
        if (item && item.content.toLowerCase().includes(keyword.toLowerCase())) {
          items.push(item)
        }
      }
      
      return items.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
    } catch (error) {
      console.error('Error querying history:', error)
      return []
    }
  }

  public getHistoricalItem(itemId: string): WorkItem | null {
    try {
      // Try to find by item ID first
      const files = readdirSync(this.historyDir)
        .filter(file => file.endsWith('.json'))
      
      for (const file of files) {
        const item = this.readJsonFile<WorkItem>(join(this.historyDir, file))
        if (item && (item.id === itemId || file.includes(itemId))) {
          return item
        }
      }
      
      // Try to find by filename
      const directPath = join(this.historyDir, itemId.endsWith('.json') ? itemId : `${itemId}.json`)
      if (existsSync(directPath)) {
        return this.readJsonFile<WorkItem>(directPath)
      }
      
      return null
    } catch (error) {
      console.error('Error getting historical item:', error)
      return null
    }
  }

  public summarizePeriod(startDate: string, endDate: string): any {
    try {
      const files = readdirSync(this.historyDir)
        .filter(file => file.endsWith('.json'))
        .filter(file => {
          const dateMatch = file.match(/(\d{4}-\d{2}-\d{2})/)
          if (dateMatch) {
            const fileDate = dateMatch[1]
            return fileDate >= startDate && fileDate <= endDate
          }
          return false
        })
      
      const items: WorkItem[] = []
      for (const file of files) {
        const item = this.readJsonFile<WorkItem>(join(this.historyDir, file))
        if (item) {
          items.push(item)
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
      }
      
      return summary
    } catch (error) {
      console.error('Error summarizing period:', error)
      return null
    }
  }

  public promoteToActive(itemId: string): WorkItem | null {
    try {
      const historicalItem = this.getHistoricalItem(itemId)
      if (!historicalItem) {
        return null
      }
      
      // Create active version
      const activeItem = {
        ...historicalItem,
        status: 'pending' as const,
        promoted_from_history: true,
        promoted_at: new Date().toISOString()
      }
      
      // Save to active directory
      const activeFilePath = join(this.activeDir, `${activeItem.id}.json`)
      this.writeJsonFile(activeFilePath, activeItem)
      
      return activeItem
    } catch (error) {
      console.error('Error promoting to active:', error)
      return null
    }
  }

  public archiveActiveItem(itemId: string): WorkItem | null {
    try {
      const activeFilePath = join(this.activeDir, `${itemId}.json`)
      const activeItem = this.readJsonFile<WorkItem>(activeFilePath)
      
      if (!activeItem) {
        return null
      }
      
      // Add archival metadata
      const archivedItem = {
        ...activeItem,
        archived_at: new Date().toISOString(),
        archived_from: 'active'
      }
      
      // Save to history
      const historyFilePath = join(this.historyDir, `${new Date().toISOString().split('T')[0]}-${itemId}.json`)
      this.writeJsonFile(historyFilePath, archivedItem)
      
      // Remove from active
      const fs = require('fs')
      if (existsSync(activeFilePath)) {
        fs.unlinkSync(activeFilePath)
      }
      
      return archivedItem
    } catch (error) {
      console.error('Error archiving active item:', error)
      return null
    }
  }

  public deferToFuture(content: string, reason: string, originalType: string = 'idea'): any {
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
      }
      
      // Save to items directory
      const itemFilePath = join(this.futureItemsDir, `item-${futureWorkItem.id}.json`)
      this.writeJsonFile(itemFilePath, futureWorkItem)
      
      // Update suggestions with new item
      this.updateGroupingSuggestions(futureWorkItem)
      
      return futureWorkItem
    } catch (error) {
      console.error('Error deferring to future work:', error)
      throw error
    }
  }

  public listFutureGroups(): any {
    try {
      const groups: any[] = []
      const ungroupedItems: any[] = []
      
      // Load all groups
      if (existsSync(this.futureGroupsDir)) {
        const groupFiles = readdirSync(this.futureGroupsDir).filter(file => file.endsWith('.json'))
        for (const file of groupFiles) {
          const group = this.readJsonFile<any>(join(this.futureGroupsDir, file))
          if (group) {
            groups.push(group)
          }
        }
      }
      
      // Load ungrouped items
      if (existsSync(this.futureItemsDir)) {
        const itemFiles = readdirSync(this.futureItemsDir).filter(file => file.endsWith('.json'))
        for (const file of itemFiles) {
          const item = this.readJsonFile<any>(join(this.futureItemsDir, file))
          if (item && item.grouping_status === 'ungrouped') {
            ungroupedItems.push(item)
          }
        }
      }
      
      // Load suggestions
      const suggestions = this.readJsonFile<any>(this.futureSuggestionsFile) || { grouping_suggestions: [] }
      
      return {
        groups: groups.sort((a, b) => new Date(b.last_updated).getTime() - new Date(a.last_updated).getTime()),
        ungrouped_items: ungroupedItems.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()),
        suggestions: suggestions.grouping_suggestions,
        total_items: groups.reduce((sum, group) => sum + (group.items?.length || 0), 0) + ungroupedItems.length
      }
    } catch (error) {
      console.error('Error listing future groups:', error)
      return { groups: [], ungrouped_items: [], suggestions: [], total_items: 0 }
    }
  }

  public promoteWorkGroup(groupName: string): WorkItem[] {
    try {
      const groupFilePath = join(this.futureGroupsDir, `${groupName}.json`)
      const group = this.readJsonFile<any>(groupFilePath)
      
      if (!group) {
        return []
      }
      
      const promotedItems: WorkItem[] = []
      
      // Promote each item in the group
      for (const itemId of group.items || []) {
        const item = this.findFutureWorkItem(itemId)
        if (item) {
          const activeItem: WorkItem = {
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
          }
          
          // Save to active directory
          const activeFilePath = join(this.activeDir, `${activeItem.id}.json`)
          this.writeJsonFile(activeFilePath, activeItem)
          
          promotedItems.push(activeItem)
        }
      }
      
      // Remove promoted items and group
      for (const itemId of group.items || []) {
        this.removeFutureWorkItem(itemId)
      }
      
      // Remove the group file
      const fs = require('fs')
      if (existsSync(groupFilePath)) {
        fs.unlinkSync(groupFilePath)
      }
      
      return promotedItems
    } catch (error) {
      console.error('Error promoting work group:', error)
      return []
    }
  }

  public createWorkGroup(name: string, description: string, itemIds: string[]): any {
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
      }
      
      // Save group
      const groupFilePath = join(this.futureGroupsDir, `${name.toLowerCase().replace(/\s+/g, '-')}.json`)
      this.writeJsonFile(groupFilePath, group)
      
      // Update items to mark them as grouped
      for (const itemId of itemIds) {
        const itemPath = this.getFutureWorkItemPath(itemId)
        if (itemPath) {
          const item = this.readJsonFile<any>(itemPath)
          if (item) {
            item.grouping_status = 'grouped'
            item.context.suggested_group = name
            this.writeJsonFile(itemPath, item)
          }
        }
      }
      
      return group
    } catch (error) {
      console.error('Error creating work group:', error)
      throw error
    }
  }

  public groomFutureWork(): any {
    try {
      const analysis = {
        overview: this.listFutureGroups(),
        suggestions: this.generateGroupingSuggestions(),
        similarity_analysis: this.analyzeSimilarityPatterns(),
        recommendations: this.generateRecommendations()
      }
      
      // Update suggestions file
      const suggestions = this.readJsonFile<any>(this.futureSuggestionsFile) || {}
      suggestions.last_updated = new Date().toISOString()
      suggestions.grouping_suggestions = analysis.suggestions
      suggestions.similarity_analysis = analysis.similarity_analysis
      this.writeJsonFile(this.futureSuggestionsFile, suggestions)
      
      return analysis
    } catch (error) {
      console.error('Error grooming future work:', error)
      throw error
    }
  }

  private extractSimilarityMetadata(content: string): SimilarityMetadata {
    // Enhanced similarity metadata extraction based on content analysis
    const keywords = this.extractKeywords(content)
    const featureDomain = this.inferFeatureDomain(content, keywords)
    const technicalDomain = this.inferTechnicalDomain(content, keywords)
    const codeLocations = this.inferCodeLocations(content, keywords)
    const strategicTheme = this.inferStrategicTheme(content, keywords)
    
    return {
      keywords: keywords,
      feature_domain: featureDomain,
      technical_domain: technicalDomain,
      code_locations: codeLocations,
      strategic_theme: strategicTheme
    }
  }

  private extractKeywords(content: string): string[] {
    // Simple keyword extraction
    const words = content.toLowerCase()
      .replace(/[^\w\s]/g, ' ')
      .split(/\s+/)
      .filter(word => word.length > 3)
    
    // Remove common words
    const stopWords = ['this', 'that', 'with', 'from', 'they', 'been', 'have', 'their', 'said', 'each', 'which']
    return words.filter(word => !stopWords.includes(word)).slice(0, 10)
  }

  private inferFeatureDomain(content: string, keywords: string[]): string {
    // Simple feature domain inference
    if (this.containsAny(content, ['auth', 'login', 'user', 'password', 'signup'])) return 'user-management'
    if (this.containsAny(content, ['search', 'filter', 'find', 'query', 'sort'])) return 'search-and-filtering'
    if (this.containsAny(content, ['profile', 'settings', 'preferences', 'account'])) return 'user-profile'
    if (this.containsAny(content, ['payment', 'billing', 'subscription', 'checkout'])) return 'payments'
    if (this.containsAny(content, ['report', 'analytics', 'dashboard', 'metrics'])) return 'reporting'
    if (this.containsAny(content, ['performance', 'optimization', 'speed', 'cache'])) return 'performance'
    return 'general'
  }

  private inferTechnicalDomain(content: string, keywords: string[]): string {
    if (this.containsAny(content, ['frontend', 'ui', 'component', 'react', 'vue', 'angular'])) return 'frontend'
    if (this.containsAny(content, ['backend', 'api', 'server', 'endpoint', 'database'])) return 'backend-api'
    if (this.containsAny(content, ['database', 'sql', 'migration', 'schema', 'query'])) return 'database'
    if (this.containsAny(content, ['testing', 'test', 'unit', 'integration', 'e2e'])) return 'testing'
    if (this.containsAny(content, ['deployment', 'deploy', 'infrastructure', 'docker', 'ci/cd'])) return 'infrastructure'
    return 'general'
  }

  private inferCodeLocations(content: string, keywords: string[]): string[] {
    const locations: string[] = []
    if (this.containsAny(content, ['auth', 'login', 'user'])) locations.push('src/auth/')
    if (this.containsAny(content, ['api', 'endpoint', 'route'])) locations.push('src/api/')
    if (this.containsAny(content, ['component', 'ui', 'frontend'])) locations.push('src/components/')
    if (this.containsAny(content, ['database', 'model', 'schema'])) locations.push('src/models/')
    return locations
  }

  private inferStrategicTheme(content: string, keywords: string[]): string {
    if (this.containsAny(content, ['user', 'experience', 'usability', 'interface'])) return 'user-experience'
    if (this.containsAny(content, ['performance', 'speed', 'optimization', 'efficiency'])) return 'performance'
    if (this.containsAny(content, ['security', 'auth', 'secure', 'protection'])) return 'security'
    if (this.containsAny(content, ['analytics', 'reporting', 'metrics', 'insights'])) return 'business-intelligence'
    if (this.containsAny(content, ['developer', 'tooling', 'testing', 'documentation'])) return 'developer-experience'
    return 'general'
  }

  private containsAny(text: string, words: string[]): boolean {
    const lowerText = text.toLowerCase()
    return words.some(word => lowerText.includes(word))
  }

  private findFutureWorkItem(itemId: string): any | null {
    // Search in items directory
    if (existsSync(this.futureItemsDir)) {
      const files = readdirSync(this.futureItemsDir).filter(file => file.endsWith('.json'))
      
      for (const file of files) {
        const item = this.readJsonFile<any>(join(this.futureItemsDir, file))
        if (item && item.id === itemId) {
          return item
        }
      }
    }
    
    return null
  }

  private getFutureWorkItemPath(itemId: string): string | null {
    // Search in items directory
    if (existsSync(this.futureItemsDir)) {
      const files = readdirSync(this.futureItemsDir).filter(file => file.endsWith('.json'))
      
      for (const file of files) {
        const item = this.readJsonFile<any>(join(this.futureItemsDir, file))
        if (item && item.id === itemId) {
          return join(this.futureItemsDir, file)
        }
      }
    }
    
    return null
  }

  private updateGroupingSuggestions(newItem: any): void {
    try {
      const suggestions = this.readJsonFile<any>(this.futureSuggestionsFile) || { grouping_suggestions: [] }
      
      // Find potential groups for this item
      const potentialGroups = this.findPotentialGroups(newItem)
      
      if (potentialGroups.length > 0) {
        const suggestion = {
          item_id: newItem.id,
          suggested_groups: potentialGroups,
          confidence: this.calculateSuggestionConfidence(newItem, potentialGroups),
          created_at: new Date().toISOString()
        }
        
        suggestions.grouping_suggestions.push(suggestion)
        suggestions.last_updated = new Date().toISOString()
        
        this.writeJsonFile(this.futureSuggestionsFile, suggestions)
      }
    } catch (error) {
      console.error('Error updating grouping suggestions:', error)
    }
  }

  private findPotentialGroups(item: any): string[] {
    const potentialGroups: string[] = []
    
    // Check existing groups for similarity
    if (existsSync(this.futureGroupsDir)) {
      const groupFiles = readdirSync(this.futureGroupsDir).filter(file => file.endsWith('.json'))
      
      for (const file of groupFiles) {
        const group = this.readJsonFile<any>(join(this.futureGroupsDir, file))
        if (group && this.calculateItemGroupSimilarity(item, group) > 0.6) {
          potentialGroups.push(group.name)
        }
      }
    }
    
    return potentialGroups
  }

  private calculateItemGroupSimilarity(item: any, group: any): number {
    // Simple similarity calculation based on metadata overlap
    let similarity = 0
    const itemMeta = item.similarity_metadata || {}
    
    // Check if any group items share similar metadata
    // This is a simplified version - in practice would be more sophisticated
    if (itemMeta.feature_domain && group.name.toLowerCase().includes(itemMeta.feature_domain.replace('-', ' '))) {
      similarity += 0.5
    }
    
    if (itemMeta.technical_domain && group.description?.toLowerCase().includes(itemMeta.technical_domain)) {
      similarity += 0.3
    }
    
    return Math.min(similarity, 1.0)
  }

  private calculateSuggestionConfidence(item: any, groups: string[]): number {
    // Simple confidence calculation
    return groups.length > 0 ? 0.8 : 0.3
  }

  private calculateGroupSimilarityScore(itemIds: string[]): number {
    // Simple group similarity calculation
    return itemIds.length > 1 ? 0.8 : 1.0
  }

  private generateGroupingSuggestions(): any[] {
    // Generate intelligent grouping suggestions
    const suggestions: any[] = []
    
    // Find ungrouped items that could be grouped together
    const ungroupedItems = this.getUngroupedItems()
    const clusters = this.clusterItemsBySimlarity(ungroupedItems)
    
    for (const cluster of clusters) {
      if (cluster.items.length > 1) {
        suggestions.push({
          suggested_group_name: cluster.theme,
          items: cluster.items.map((item: any) => item.id),
          similarity_score: cluster.similarity_score,
          rationale: cluster.rationale
        })
      }
    }
    
    return suggestions
  }

  private analyzeSimilarityPatterns(): any {
    // Analyze patterns in future work items
    const allItems = this.getAllFutureWorkItems()
    
    const featureClusters = this.groupItemsByFeatureDomain(allItems)
    const technicalDomains = this.groupItemsByTechnicalDomain(allItems)
    const codeLocations = this.groupItemsByCodeLocation(allItems)
    
    return {
      feature_clusters: featureClusters,
      technical_domains: technicalDomains,
      code_locations: codeLocations
    }
  }

  private generateRecommendations(): string[] {
    // Generate strategic recommendations
    const recommendations: string[] = []
    const overview = this.listFutureGroups()
    
    if (overview.ungrouped_items.length > 3) {
      recommendations.push(`Consider grouping ${overview.ungrouped_items.length} ungrouped items`)
    }
    
    if (overview.groups.length === 0) {
      recommendations.push('Create your first work group to organize future items')
    }
    
    return recommendations
  }

  // Helper methods for analysis
  private getUngroupedItems(): any[] {
    const items: any[] = []
    if (existsSync(this.futureItemsDir)) {
      const files = readdirSync(this.futureItemsDir).filter(file => file.endsWith('.json'))
      for (const file of files) {
        const item = this.readJsonFile<any>(join(this.futureItemsDir, file))
        if (item && item.grouping_status === 'ungrouped') {
          items.push(item)
        }
      }
    }
    return items
  }

  private clusterItemsBySimlarity(items: any[]): any[] {
    // Simple clustering - group by feature domain
    const clusters: { [key: string]: any[] } = {}
    
    for (const item of items) {
      const domain = item.similarity_metadata?.feature_domain || 'general'
      if (!clusters[domain]) {
        clusters[domain] = []
      }
      clusters[domain].push(item)
    }
    
    return Object.entries(clusters).map(([domain, items]) => ({
      theme: domain.replace('-', ' ').replace(/\b\w/g, (l: string) => l.toUpperCase()),
      items: items,
      similarity_score: items.length > 1 ? 0.8 : 0.3,
      rationale: `Items related to ${domain.replace('-', ' ')}`
    }))
  }

  private getAllFutureWorkItems(): any[] {
    const items: any[] = []
    if (existsSync(this.futureItemsDir)) {
      const files = readdirSync(this.futureItemsDir).filter(file => file.endsWith('.json'))
      for (const file of files) {
        const item = this.readJsonFile<any>(join(this.futureItemsDir, file))
        if (item) {
          items.push(item)
        }
      }
    }
    return items
  }

  private groupItemsByFeatureDomain(items: any[]): { [key: string]: number } {
    const domains: { [key: string]: number } = {}
    for (const item of items) {
      const domain = item.similarity_metadata?.feature_domain || 'general'
      domains[domain] = (domains[domain] || 0) + 1
    }
    return domains
  }

  private groupItemsByTechnicalDomain(items: any[]): { [key: string]: number } {
    const domains: { [key: string]: number } = {}
    for (const item of items) {
      const domain = item.similarity_metadata?.technical_domain || 'general'
      domains[domain] = (domains[domain] || 0) + 1
    }
    return domains
  }

  private groupItemsByCodeLocation(items: any[]): { [key: string]: number } {
    const locations: { [key: string]: number } = {}
    for (const item of items) {
      const itemLocations = item.similarity_metadata?.code_locations || []
      for (const location of itemLocations) {
        locations[location] = (locations[location] || 0) + 1
      }
    }
    return locations
  }

  private removeFutureWorkItem(itemId: string): void {
    const filePath = this.getFutureWorkItemPath(itemId)
    if (filePath && existsSync(filePath)) {
      const fs = require('fs')
      fs.unlinkSync(filePath)
    }
  }

  // Smart Referencing Methods

  /**
   * Get contextual suggestions for current active work
   */
  public getContextualSuggestions(): ContextualSuggestion[] {
    const activeItems = this.loadActiveTodos()
    return this.smartReferenceEngine.getContextualSuggestions(activeItems)
  }

  /**
   * Generate smart references for a specific work item
   */
  public generateSmartReferences(itemId: string): any[] {
    const activeItems = this.loadActiveTodos()
    const item = activeItems.find(i => i.id === itemId)
    
    if (!item) {
      // Try to find in historical items
      const historicalItem = this.getHistoricalItem(itemId)
      if (historicalItem) {
        return this.smartReferenceEngine.generateAutomaticReferences(historicalItem)
      }
      return []
    }
    
    return this.smartReferenceEngine.generateAutomaticReferences(item)
  }

  /**
   * Calculate similarity between two work items
   */
  public calculateSimilarity(itemId1: string, itemId2: string): any {
    const item1 = this.findWorkItem(itemId1)
    const item2 = this.findWorkItem(itemId2)
    
    if (!item1 || !item2) {
      return null
    }
    
    return this.smartReferenceEngine.calculateSemanticSimilarity(item1, item2)
  }

  /**
   * Get enhanced work state with smart referencing context
   */
  public getEnhancedWorkState(): any {
    const baseWorkState = this.getCurrentWorkState()
    const suggestions = this.getContextualSuggestions()
    
    return {
      ...baseWorkState,
      smart_suggestions: suggestions,
      reference_summary: {
        total_suggestions: suggestions.length,
        high_priority: suggestions.filter(s => s.priority === 'high').length,
        suggestion_types: this.groupSuggestionsByType(suggestions)
      }
    }
  }

  private findWorkItem(itemId: string): WorkItem | null {
    // First try active items
    const activeItems = this.loadActiveTodos()
    const activeItem = activeItems.find(i => i.id === itemId)
    if (activeItem) return activeItem
    
    // Then try historical items
    return this.getHistoricalItem(itemId)
  }

  private groupSuggestionsByType(suggestions: ContextualSuggestion[]): any {
    const grouped: { [key: string]: number } = {}
    
    for (const suggestion of suggestions) {
      grouped[suggestion.type] = (grouped[suggestion.type] || 0) + 1
    }
    
    return grouped
  }

  // Reference Mapping Methods

  /**
   * Generate complete reference map for current work context
   */
  public generateReferenceMap(): ReferenceMap {
    return this.referenceMapper.generateReferenceMap()
  }

  /**
   * Generate focused reference map for a specific work item
   */
  public generateFocusedReferenceMap(itemId: string, depth: number = 2): ReferenceMap {
    return this.referenceMapper.generateFocusedMap(itemId, depth)
  }

  /**
   * Find reference path between two work items
   */
  public findReferencePath(sourceId: string, targetId: string): string[] {
    return this.referenceMapper.findReferencePath(sourceId, targetId)
  }

  /**
   * Generate ASCII visualization of reference relationships
   */
  public visualizeReferences(): string {
    const referenceMap = this.generateReferenceMap()
    return this.referenceMapper.generateASCIIVisualization(referenceMap)
  }
}