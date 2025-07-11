import { readFileSync, writeFileSync, existsSync, mkdirSync, readdirSync, statSync } from 'fs'
import { join, dirname } from 'path'
import { execSync } from 'child_process'
import { WorkItem, Finding, GitContext, WorkState, SessionSummary } from '../types'

export class WorkStateManager {
  private readonly baseDir = join(process.env.HOME || '', '.claude')
  private readonly todosDir = join(this.baseDir, 'todos')
  private readonly findingsDir = join(this.baseDir, 'findings')
  private readonly workStateDir = join(this.baseDir, 'work-state')
  private readonly projectsDir = join(this.workStateDir, 'projects')

  constructor() {
    this.ensureDirectories()
  }

  private ensureDirectories(): void {
    [this.baseDir, this.todosDir, this.findingsDir, this.workStateDir, this.projectsDir].forEach(dir => {
      if (!existsSync(dir)) {
        mkdirSync(dir, { recursive: true })
      }
    })
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

  private loadActiveTodos(): WorkItem[] {
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
    const filePath = join(this.todosDir, `${workItem.id}.json`)
    this.writeJsonFile(filePath, workItem)
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
}