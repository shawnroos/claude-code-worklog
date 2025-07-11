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
      },
      {
        name: 'query_history',
        description: 'Search through historical work items with optional date range filtering',
        inputSchema: {
          type: 'object',
          properties: {
            keyword: {
              type: 'string',
              description: 'Search keyword or phrase'
            },
            start_date: {
              type: 'string',
              description: 'Start date in YYYY-MM-DD format (optional)'
            },
            end_date: {
              type: 'string',
              description: 'End date in YYYY-MM-DD format (optional)'
            },
            type: {
              type: 'string',
              enum: ['plan', 'proposal', 'finding', 'report', 'summary', 'digest'],
              description: 'Optional: filter by work item type'
            }
          },
          required: ['keyword']
        }
      },
      {
        name: 'get_historical_context',
        description: 'Retrieve detailed information from a specific historical item',
        inputSchema: {
          type: 'object',
          properties: {
            item_id: {
              type: 'string',
              description: 'Historical item identifier or filename'
            }
          },
          required: ['item_id']
        }
      },
      {
        name: 'summarize_period',
        description: 'Generate summary of work activity for a specific time period',
        inputSchema: {
          type: 'object',
          properties: {
            start_date: {
              type: 'string',
              description: 'Start date in YYYY-MM-DD format'
            },
            end_date: {
              type: 'string',
              description: 'End date in YYYY-MM-DD format'
            }
          },
          required: ['start_date', 'end_date']
        }
      },
      {
        name: 'promote_to_active',
        description: 'Move a historical item to active context',
        inputSchema: {
          type: 'object',
          properties: {
            item_id: {
              type: 'string',
              description: 'Historical item identifier to promote'
            }
          },
          required: ['item_id']
        }
      },
      {
        name: 'archive_active_item',
        description: 'Move an active item to historical archive',
        inputSchema: {
          type: 'object',
          properties: {
            item_id: {
              type: 'string',
              description: 'Active item identifier to archive'
            }
          },
          required: ['item_id']
        }
      },
      {
        name: 'defer_to_future',
        description: 'Frictionless deferral of work for future implementation',
        inputSchema: {
          type: 'object',
          properties: {
            content: {
              type: 'string',
              description: 'Work item description'
            },
            reason: {
              type: 'string',
              description: 'Reason for deprioritization (e.g., "Out of scope for current sprint")'
            },
            type: {
              type: 'string',
              enum: ['plan', 'proposal', 'todo', 'idea'],
              description: 'Type of work item (optional, defaults to "idea")'
            }
          },
          required: ['content', 'reason']
        }
      },
      {
        name: 'list_future_groups',
        description: 'View current future work groups and ungrouped items',
        inputSchema: {
          type: 'object',
          properties: {},
          required: []
        }
      },
      {
        name: 'groom_future_work',
        description: 'Analyze and reorganize future work with intelligent suggestions',
        inputSchema: {
          type: 'object',
          properties: {},
          required: []
        }
      },
      {
        name: 'create_work_group',
        description: 'Create a logical group of related future work items',
        inputSchema: {
          type: 'object',
          properties: {
            name: {
              type: 'string',
              description: 'Group name (e.g., "Authentication Features")'
            },
            description: {
              type: 'string',
              description: 'Description of what the group contains'
            },
            item_ids: {
              type: 'array',
              items: { type: 'string' },
              description: 'List of future work item IDs to include in group'
            }
          },
          required: ['name', 'description', 'item_ids']
        }
      },
      {
        name: 'promote_work_group',
        description: 'Promote an entire group of related work back to active context',
        inputSchema: {
          type: 'object',
          properties: {
            group_name: {
              type: 'string',
              description: 'Name of the work group to promote'
            }
          },
          required: ['group_name']
        }
      },
      {
        name: 'get_contextual_suggestions',
        description: 'Get smart suggestions for current active work based on historical context',
        inputSchema: {
          type: 'object',
          properties: {},
          required: []
        }
      },
      {
        name: 'generate_smart_references',
        description: 'Generate automatic references for a specific work item',
        inputSchema: {
          type: 'object',
          properties: {
            item_id: {
              type: 'string',
              description: 'Work item ID to generate references for'
            }
          },
          required: ['item_id']
        }
      },
      {
        name: 'calculate_similarity',
        description: 'Calculate similarity score between two work items',
        inputSchema: {
          type: 'object',
          properties: {
            item_id_1: {
              type: 'string',
              description: 'First work item ID'
            },
            item_id_2: {
              type: 'string',
              description: 'Second work item ID'
            }
          },
          required: ['item_id_1', 'item_id_2']
        }
      },
      {
        name: 'get_enhanced_work_state',
        description: 'Get work state enhanced with smart referencing context and suggestions',
        inputSchema: {
          type: 'object',
          properties: {},
          required: []
        }
      },
      {
        name: 'generate_reference_map',
        description: 'Generate visual reference map showing relationships between work items',
        inputSchema: {
          type: 'object',
          properties: {},
          required: []
        }
      },
      {
        name: 'generate_focused_reference_map',
        description: 'Generate focused reference map for a specific work item',
        inputSchema: {
          type: 'object',
          properties: {
            item_id: {
              type: 'string',
              description: 'Work item ID to focus on'
            },
            depth: {
              type: 'number',
              description: 'Depth of reference traversal (default: 2)',
              minimum: 1,
              maximum: 5
            }
          },
          required: ['item_id']
        }
      },
      {
        name: 'find_reference_path',
        description: 'Find reference path between two work items',
        inputSchema: {
          type: 'object',
          properties: {
            source_id: {
              type: 'string',
              description: 'Source work item ID'
            },
            target_id: {
              type: 'string',
              description: 'Target work item ID'
            }
          },
          required: ['source_id', 'target_id']
        }
      },
      {
        name: 'visualize_references',
        description: 'Generate ASCII visualization of work item references',
        inputSchema: {
          type: 'object',
          properties: {},
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
        
        
        case 'load_work_state':
          return this.handleLoadWorkState(params)
        
        case 'save_work_state':
          return this.handleSaveWorkState(params)
        
        case 'query_history':
          return this.handleQueryHistory(params)
        
        case 'get_historical_context':
          return this.handleGetHistoricalContext(params)
        
        case 'summarize_period':
          return this.handleSummarizePeriod(params)
        
        case 'promote_to_active':
          return this.handlePromoteToActive(params)
        
        case 'archive_active_item':
          return this.handleArchiveActiveItem(params)
        
        case 'defer_to_future':
          return this.handleDeferToFuture(params)
        
        case 'list_future_groups':
          return this.handleListFutureGroups(params)
        
        case 'groom_future_work':
          return this.handleGroomFutureWork(params)
        
        case 'create_work_group':
          return this.handleCreateWorkGroup(params)
        
        case 'promote_work_group':
          return this.handlePromoteWorkGroup(params)
        
        case 'get_contextual_suggestions':
          return this.handleGetContextualSuggestions(params)
        
        case 'generate_smart_references':
          return this.handleGenerateSmartReferences(params)
        
        case 'calculate_similarity':
          return this.handleCalculateSimilarity(params)
        
        case 'get_enhanced_work_state':
          return this.handleGetEnhancedWorkState(params)
        
        case 'generate_reference_map':
          return this.handleGenerateReferenceMap(params)
        
        case 'generate_focused_reference_map':
          return this.handleGenerateFocusedReferenceMap(params)
        
        case 'find_reference_path':
          return this.handleFindReferencePath(params)
        
        case 'visualize_references':
          return this.handleVisualizeReferences(params)
        
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

  private handleQueryHistory(params: McpToolParams): McpToolResponse {
    const { keyword, start_date, end_date, type } = params
    
    try {
      const historyResults = this.workStateManager.queryHistory(keyword, start_date, end_date, type)
      
      // Enhance results with contextual relevance if there are active items
      const enhancedResults = historyResults.map(item => {
        const contextualSuggestions = this.workStateManager.getContextualSuggestions()
        const relevantSuggestion = contextualSuggestions.find(s => s.target_item_id === item.id)
        
        return {
          ...item,
          contextual_relevance: relevantSuggestion ? {
            confidence: relevantSuggestion.confidence,
            relationship_type: relevantSuggestion.type,
            priority: relevantSuggestion.priority,
            action_hint: relevantSuggestion.action_hint
          } : null
        }
      })
      
      return {
        success: true,
        data: enhancedResults,
        message: `Found ${historyResults.length} historical items matching "${keyword}"${enhancedResults.some(r => r.contextual_relevance) ? ' (enhanced with contextual relevance)' : ''}`
      }
    } catch (error) {
      return {
        success: false,
        error: `Error querying history: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }

  private handleGetHistoricalContext(params: McpToolParams): McpToolResponse {
    const { item_id } = params
    
    if (!item_id) {
      return {
        success: false,
        error: 'Missing required parameter: item_id'
      }
    }
    
    try {
      const historicalItem = this.workStateManager.getHistoricalItem(item_id)
      
      if (!historicalItem) {
        return {
          success: false,
          error: `Historical item not found: ${item_id}`
        }
      }
      
      return {
        success: true,
        data: historicalItem,
        message: `Retrieved historical context for ${item_id}`
      }
    } catch (error) {
      return {
        success: false,
        error: `Error retrieving historical context: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }

  private handleSummarizePeriod(params: McpToolParams): McpToolResponse {
    const { start_date, end_date } = params
    
    if (!start_date || !end_date) {
      return {
        success: false,
        error: 'Missing required parameters: start_date and end_date'
      }
    }
    
    try {
      const summary = this.workStateManager.summarizePeriod(start_date, end_date)
      
      return {
        success: true,
        data: summary,
        message: `Generated summary for period ${start_date} to ${end_date}`
      }
    } catch (error) {
      return {
        success: false,
        error: `Error generating period summary: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }

  private handlePromoteToActive(params: McpToolParams): McpToolResponse {
    const { item_id } = params
    
    if (!item_id) {
      return {
        success: false,
        error: 'Missing required parameter: item_id'
      }
    }
    
    try {
      const promoted = this.workStateManager.promoteToActive(item_id)
      
      return {
        success: true,
        data: promoted,
        message: `Promoted ${item_id} to active context`
      }
    } catch (error) {
      return {
        success: false,
        error: `Error promoting to active: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }

  private handleArchiveActiveItem(params: McpToolParams): McpToolResponse {
    const { item_id } = params
    
    if (!item_id) {
      return {
        success: false,
        error: 'Missing required parameter: item_id'
      }
    }
    
    try {
      const archived = this.workStateManager.archiveActiveItem(item_id)
      
      return {
        success: true,
        data: archived,
        message: `Archived ${item_id} to historical storage`
      }
    } catch (error) {
      return {
        success: false,
        error: `Error archiving active item: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }

  private handleDeferToFuture(params: McpToolParams): McpToolResponse {
    const { content, reason, type } = params
    
    if (!content || !reason) {
      return {
        success: false,
        error: 'Missing required parameters: content and reason'
      }
    }
    
    try {
      const futureWorkItem = this.workStateManager.deferToFuture(
        content, 
        reason,
        type || 'idea'
      )
      
      return {
        success: true,
        data: futureWorkItem,
        message: `Deferred work to future: ${content.slice(0, 50)}...`
      }
    } catch (error) {
      return {
        success: false,
        error: `Error deferring to future: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }

  private handleListFutureGroups(params: McpToolParams): McpToolResponse {
    try {
      const futureGroups = this.workStateManager.listFutureGroups()
      
      return {
        success: true,
        data: futureGroups,
        message: `Found ${futureGroups.groups.length} groups and ${futureGroups.ungrouped_items.length} ungrouped items`
      }
    } catch (error) {
      return {
        success: false,
        error: `Error listing future groups: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }

  private handleGroomFutureWork(params: McpToolParams): McpToolResponse {
    try {
      const analysis = this.workStateManager.groomFutureWork()
      
      return {
        success: true,
        data: analysis,
        message: `Analyzed future work: ${analysis.overview.total_items} total items, ${analysis.suggestions.length} grouping suggestions`
      }
    } catch (error) {
      return {
        success: false,
        error: `Error grooming future work: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }

  private handleCreateWorkGroup(params: McpToolParams): McpToolResponse {
    const { name, description, item_ids } = params
    
    if (!name || !description || !item_ids || !Array.isArray(item_ids)) {
      return {
        success: false,
        error: 'Missing required parameters: name, description, and item_ids'
      }
    }
    
    try {
      const workGroup = this.workStateManager.createWorkGroup(name, description, item_ids)
      
      return {
        success: true,
        data: workGroup,
        message: `Created work group "${name}" with ${item_ids.length} items`
      }
    } catch (error) {
      return {
        success: false,
        error: `Error creating work group: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }

  private handlePromoteWorkGroup(params: McpToolParams): McpToolResponse {
    const { group_name } = params
    
    if (!group_name) {
      return {
        success: false,
        error: 'Missing required parameter: group_name'
      }
    }
    
    try {
      const promotedItems = this.workStateManager.promoteWorkGroup(group_name)
      
      return {
        success: true,
        data: promotedItems,
        message: `Promoted work group "${group_name}" with ${promotedItems.length} items to active context`
      }
    } catch (error) {
      return {
        success: false,
        error: `Error promoting work group: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }

  private handleGetContextualSuggestions(params: McpToolParams): McpToolResponse {
    try {
      const suggestions = this.workStateManager.getContextualSuggestions()
      
      return {
        success: true,
        data: suggestions,
        message: `Found ${suggestions.length} contextual suggestions based on current active work`
      }
    } catch (error) {
      return {
        success: false,
        error: `Error getting contextual suggestions: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }

  private handleGenerateSmartReferences(params: McpToolParams): McpToolResponse {
    const { item_id } = params
    
    if (!item_id) {
      return {
        success: false,
        error: 'Missing required parameter: item_id'
      }
    }
    
    try {
      const references = this.workStateManager.generateSmartReferences(item_id)
      
      return {
        success: true,
        data: references,
        message: `Generated ${references.length} smart references for item ${item_id}`
      }
    } catch (error) {
      return {
        success: false,
        error: `Error generating smart references: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }

  private handleCalculateSimilarity(params: McpToolParams): McpToolResponse {
    const { item_id_1, item_id_2 } = params
    
    if (!item_id_1 || !item_id_2) {
      return {
        success: false,
        error: 'Missing required parameters: item_id_1 and item_id_2'
      }
    }
    
    try {
      const similarity = this.workStateManager.calculateSimilarity(item_id_1, item_id_2)
      
      if (!similarity) {
        return {
          success: false,
          error: 'One or both work items not found'
        }
      }
      
      return {
        success: true,
        data: similarity,
        message: `Calculated similarity score: ${similarity.total_score.toFixed(3)}`
      }
    } catch (error) {
      return {
        success: false,
        error: `Error calculating similarity: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }

  private handleGetEnhancedWorkState(params: McpToolParams): McpToolResponse {
    try {
      const enhancedState = this.workStateManager.getEnhancedWorkState()
      
      return {
        success: true,
        data: enhancedState,
        message: `Enhanced work state with ${enhancedState.smart_suggestions?.length || 0} smart suggestions`
      }
    } catch (error) {
      return {
        success: false,
        error: `Error getting enhanced work state: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }

  private handleGenerateReferenceMap(params: McpToolParams): McpToolResponse {
    try {
      const referenceMap = this.workStateManager.generateReferenceMap()
      
      return {
        success: true,
        data: referenceMap,
        message: `Generated reference map with ${referenceMap.summary.total_items} items and ${referenceMap.summary.total_references} references`
      }
    } catch (error) {
      return {
        success: false,
        error: `Error generating reference map: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }

  private handleGenerateFocusedReferenceMap(params: McpToolParams): McpToolResponse {
    const { item_id, depth = 2 } = params
    
    if (!item_id) {
      return {
        success: false,
        error: 'Missing required parameter: item_id'
      }
    }
    
    try {
      const referenceMap = this.workStateManager.generateFocusedReferenceMap(item_id, depth)
      
      return {
        success: true,
        data: referenceMap,
        message: `Generated focused reference map for ${item_id} with depth ${depth}`
      }
    } catch (error) {
      return {
        success: false,
        error: `Error generating focused reference map: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }

  private handleFindReferencePath(params: McpToolParams): McpToolResponse {
    const { source_id, target_id } = params
    
    if (!source_id || !target_id) {
      return {
        success: false,
        error: 'Missing required parameters: source_id and target_id'
      }
    }
    
    try {
      const path = this.workStateManager.findReferencePath(source_id, target_id)
      
      return {
        success: true,
        data: { path: path },
        message: path.length > 0 ? `Found reference path with ${path.length} steps` : 'No reference path found'
      }
    } catch (error) {
      return {
        success: false,
        error: `Error finding reference path: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }

  private handleVisualizeReferences(params: McpToolParams): McpToolResponse {
    try {
      const visualization = this.workStateManager.visualizeReferences()
      
      return {
        success: true,
        data: { visualization: visualization },
        message: 'Generated ASCII visualization of work item references'
      }
    } catch (error) {
      return {
        success: false,
        error: `Error generating visualization: ${error instanceof Error ? error.message : String(error)}`
      }
    }
  }
}