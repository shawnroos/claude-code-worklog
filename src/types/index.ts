export interface GitContext {
  branch: string
  worktree: string
  remote_url?: string
  working_directory: string
}

export interface WorkItemMetadata {
  plan_steps?: string[]
  decision_rationale?: string
  implementation_notes?: string
  priority?: 'low' | 'medium' | 'high'
  tags?: string[]
  promoted_from_future?: boolean
  promoted_from_history?: boolean
  promoted_from_group?: string
  original_schedule?: any
  archived_from?: string
  archived_at?: string
}

export interface WorkItem {
  id: string
  type: 'todo' | 'plan' | 'proposal' | 'finding' | 'report' | 'summary'
  content: string
  status: 'pending' | 'in_progress' | 'completed'
  context: GitContext
  session_id: string
  timestamp: string
  metadata?: WorkItemMetadata
}

export interface Finding {
  id: string
  type: 'research' | 'search' | 'analysis' | 'test_results' | 'implementation' | 'report'
  content: string
  context: string
  tool_name: string
  timestamp: string
  session_id: string
  working_directory: string
  git_branch: string
  git_worktree: string
}

export interface SessionSummary {
  session_id: string
  timestamp: string
  git_context: GitContext
  completed_todos: number
  pending_todos: number
  findings_count: number
  plans_created: number
  proposals_made: number
  key_decisions: string[]
  outcomes: string[]
}

export interface WorkState {
  current_session: string
  active_todos: WorkItem[]
  recent_findings: Finding[]
  session_summary?: SessionSummary
  cross_worktree_conflicts?: string[]
}

export interface McpToolParams {
  [key: string]: any
}

export interface McpToolResponse {
  success: boolean
  data?: any
  error?: string
  message?: string
}