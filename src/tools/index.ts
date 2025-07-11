import { Tool } from '@modelcontextprotocol/sdk/types.js'
import { WorkStateManager } from '../services/WorkStateManager.js'
import { McpToolParams, McpToolResponse } from '../types/index.js'

export class WorkTrackingTools {
  private workStateManager = new WorkStateManager()

  public getTools(): Tool[] {
    return [
      {
        name: 'get_work_state',
        description: 'Get current work state including todos, findings, and session summary',
        inputSchema: {
          type: 'object',
          properties: {},
          required: []
        }
      },
      {
        name: 'save_plan',
        description: 'Save a plan with structured steps for future reference',
        inputSchema: {
          type: 'object',
          properties: {
            content: {
              type: 'string',
              description: 'The main plan description'
            },
            steps: {
              type: 'array',
              items: { type: 'string' },
              description: 'List of plan steps'
            }
          },
          required: ['content', 'steps']
        }
      },
      {
        name: 'save_proposal',
        description: 'Save a proposal or architectural decision with rationale',
        inputSchema: {
          type: 'object',
          properties: {
            content: {
              type: 'string',
              description: 'The proposal description'
            },
            rationale: {
              type: 'string',
              description: 'The reasoning behind the proposal'
            }
          },
          required: ['content', 'rationale']
        }
      },
      {
        name: 'search_work_items',
        description: 'Search through work items (todos, plans, proposals, findings)',
        inputSchema: {
          type: 'object',
          properties: {
            query: {
              type: 'string',
              description: 'Search query'
            },
            type: {
              type: 'string',
              enum: ['todo', 'plan', 'proposal', 'finding', 'report', 'summary'],
              description: 'Optional: filter by work item type'
            }
          },
          required: ['query']
        }
      },
      {
        name: 'get_session_summary',
        description: 'Get summary of current or specified session',
        inputSchema: {
          type: 'object',
          properties: {
            session_id: {
              type: 'string',
              description: 'Optional: specific session ID to get summary for'
            }
          },
          required: []
        }
      },
      {
        name: 'get_cross_worktree_status',
        description: 'Get work status across different git worktrees',
        inputSchema: {
          type: 'object',
          properties: {
            keyword: {
              type: 'string',
              description: 'Optional: keyword to filter related work'
            }
          },
          required: []
        }
      },
      {
        name: 'load_work_state',
        description: 'Load work state for specific branch or worktree',
        inputSchema: {
          type: 'object',
          properties: {
            branch: {
              type: 'string',
              description: 'Branch name to load work state from'
            }
          },
          required: []
        }
      },
      {
        name: 'save_work_state',
        description: 'Manually save current work state',
        inputSchema: {
          type: 'object',
          properties: {
            note: {
              type: 'string',
              description: 'Optional note about the save'
            }
          },
          required: []
        }
      }
    ]
  }

  public async handleToolCall(name: string, params: McpToolParams): Promise<McpToolResponse> {
    try {
      switch (name) {
        case 'get_work_state':
          return this.handleGetWorkState()
        
        case 'save_plan':
          return this.handleSavePlan(params)
        
        case 'save_proposal':
          return this.handleSaveProposal(params)
        
        case 'search_work_items':
          return this.handleSearchWorkItems(params)
        
        case 'get_session_summary':
          return this.handleGetSessionSummary(params)
        
        case 'get_cross_worktree_status':
          return this.handleGetCrossWorktreeStatus(params)
        
        case 'load_work_state':
          return this.handleLoadWorkState(params)
        
        case 'save_work_state':
          return this.handleSaveWorkState(params)
        
        default:
          return {
            success: false,
            error: `Unknown tool: ${name}`
          }
      }
    } catch (error) {
      return {
        success: false,
        error: `Error handling tool ${name}: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }

  private handleGetWorkState(): McpToolResponse {
    const workState = this.workStateManager.getCurrentWorkState()
    return {
      success: true,
      data: workState,
      message: `Found ${workState.active_todos.length} active todos and ${workState.recent_findings.length} recent findings`
    }
  }

  private handleSavePlan(params: McpToolParams): McpToolResponse {
    const { content, steps } = params
    if (!content || !steps || !Array.isArray(steps)) {
      return {
        success: false,
        error: 'Missing required parameters: content and steps'
      }
    }

    const planItem = this.workStateManager.savePlan(content, steps)
    return {
      success: true,
      data: planItem,
      message: `Plan saved with ${steps.length} steps`
    }
  }

  private handleSaveProposal(params: McpToolParams): McpToolResponse {
    const { content, rationale } = params
    if (!content || !rationale) {
      return {
        success: false,
        error: 'Missing required parameters: content and rationale'
      }
    }

    const proposalItem = this.workStateManager.saveProposal(content, rationale)
    return {
      success: true,
      data: proposalItem,
      message: 'Proposal saved successfully'
    }
  }

  private handleSearchWorkItems(params: McpToolParams): McpToolResponse {
    const { query, type } = params
    if (!query) {
      return {
        success: false,
        error: 'Missing required parameter: query'
      }
    }

    let items
    if (type) {
      items = this.workStateManager.getWorkItemsByType(type).filter(item => 
        item.content.toLowerCase().includes(query.toLowerCase())
      )
    } else {
      items = this.workStateManager.searchWorkItems(query)
    }

    return {
      success: true,
      data: items,
      message: `Found ${items.length} work items matching "${query}"`
    }
  }

  private handleGetSessionSummary(params: McpToolParams): McpToolResponse {
    const { session_id } = params
    
    if (session_id) {
      const summary = this.workStateManager.getSessionSummary(session_id)
      return {
        success: true,
        data: summary,
        message: summary ? 'Session summary found' : 'No summary found for this session'
      }
    } else {
      const workState = this.workStateManager.getCurrentWorkState()
      return {
        success: true,
        data: workState.session_summary,
        message: 'Current session summary'
      }
    }
  }

  private handleGetCrossWorktreeStatus(params: McpToolParams): McpToolResponse {
    const { keyword } = params
    // This would call the existing bash script
    const { execSync } = require('child_process')
    
    try {
      const command = keyword 
        ? `~/.claude/scripts/work-conflicts.sh "${keyword}"`
        : `~/.claude/scripts/work-status.sh`
      
      const output = execSync(command, { encoding: 'utf8' })
      
      return {
        success: true,
        data: { output },
        message: 'Cross-worktree status retrieved'
      }
    } catch (error) {
      return {
        success: false,
        error: `Error getting cross-worktree status: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }

  private handleLoadWorkState(params: McpToolParams): McpToolResponse {
    const { branch } = params
    
    try {
      const command = branch 
        ? `~/.claude/scripts/work.sh load "${branch}"`
        : `~/.claude/scripts/work.sh load`
      
      const { execSync } = require('child_process')
      const output = execSync(command, { encoding: 'utf8' })
      
      return {
        success: true,
        data: { output },
        message: `Work state loaded${branch ? ` for branch ${branch}` : ''}`
      }
    } catch (error) {
      return {
        success: false,
        error: `Error loading work state: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }

  private handleSaveWorkState(params: McpToolParams): McpToolResponse {
    const { note } = params
    
    try {
      const command = note 
        ? `~/.claude/scripts/work.sh save "${note}"`
        : `~/.claude/scripts/work.sh save`
      
      const { execSync } = require('child_process')
      const output = execSync(command, { encoding: 'utf8' })
      
      return {
        success: true,
        data: { output },
        message: 'Work state saved successfully'
      }
    } catch (error) {
      return {
        success: false,
        error: `Error saving work state: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }
}