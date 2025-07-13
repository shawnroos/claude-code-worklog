---
id: proposal-work-command-2025-07-12-cmd001
type: proposal
summary: Implement /work command for Claude conversations
schedule: next
technical_tags: [claude, command, integration, workflow]
session_number: session-work-cmd
created_at: 2025-07-12T23:06:00Z
updated_at: 2025-07-12T23:06:00Z
git_context:
  branch: main
  worktree: claude-work-tracker-ui
  working_directory: /Users/shawnroos/claude-work-tracker/claude-work-tracker-ui
metadata:
  status: active
  approval_status: pending
  estimated_impact: high
  dependencies: [proposal-mcp-server-2025-07-12-mcp001]
---

# Implement /work Command for Claude Conversations

## Proposal Summary

Create a comprehensive `/work` command system for Claude that enables seamless work item management directly within conversations, leveraging the MCP server for backend integration.

## Command Structure

### Core Commands
- `/work status` - Show current work overview across all schedules
- `/work list [schedule]` - List items in specific schedule (NOW/NEXT/LATER)
- `/work show <id>` - Display full work item details
- `/work create` - Interactive work item creation
- `/work move <id> <schedule>` - Move item between schedules

### Advanced Commands
- `/work search <query>` - Search work items by content/tags
- `/work summary` - Generate intelligent work summary
- `/work plan` - AI-assisted planning and prioritization
- `/work context` - Load current project context for discussion

## Usage Examples

### Quick Status Check
```
User: /work status
Claude: 📋 Work Overview:
• NOW (2): UI performance optimization, API integration
• NEXT (5): MCP server, mobile design, testing framework
• LATER (3): Documentation, monitoring, deployment

Most recent: "Fix critical auth bug" (completed 2h ago)
```

### Creating Work Items
```
User: /work create
Claude: I'll help you create a new work item. Let me ask a few questions:

What type of work is this?
1. 📋 Plan - Strategic planning or roadmap
2. 💡 Proposal - New feature or enhancement  
3. 🔍 Analysis - Research or investigation
4. 📝 Update - Progress update or status
5. ⚖️ Decision - Decision record or choice

User: 2
Claude: Great! What's the summary for this proposal?
User: Add dark mode to the application
Claude: ✅ Created: PROPOSAL - Add dark mode to the application
   📍 Schedule: NEXT
   🏷️ Tags: ui, themes, accessibility
   📝 Item ID: proposal-dark-mode-2025-07-12-dm001
```

### Intelligent Planning
```
User: /work plan
Claude: Based on your current work items, I recommend:

🔥 NOW priorities:
1. Complete "API integration" (blocking 3 NEXT items)
2. Finish "UI performance" (affects user experience)

📅 NEXT planning:
• "MCP server" → should move to NOW (enables other features)
• "Mobile design" → can wait until performance is complete
• Consider breaking "Testing framework" into smaller tasks

💡 Suggestions:
- The auth bug fix freed up capacity for new features
- Dark mode proposal pairs well with UI performance work
```

## Technical Implementation

### Command Parser
```go
type WorkCommand struct {
    Action    string            // status, list, create, etc.
    Target    string            // schedule, item ID, etc.
    Options   map[string]string // additional parameters
    Context   *ConversationContext
}

type CommandHandler interface {
    Execute(cmd *WorkCommand) (*CommandResponse, error)
    Validate(cmd *WorkCommand) error
    GetHelp() string
}
```

### MCP Integration
```go
type WorkCommandProcessor struct {
    mcpClient     *mcp.Client
    contextMgr    *ContextManager
    templateMgr   *TemplateManager
    aiAssistant   *AIAssistant
}

func (p *WorkCommandProcessor) ProcessCommand(input string) (*Response, error) {
    cmd := p.parseCommand(input)
    
    // Load current work context
    context := p.contextMgr.GetWorkContext()
    
    // Execute command via MCP
    result := p.mcpClient.ExecuteTool(cmd.Action, cmd.Parameters)
    
    // Format response with AI assistance
    return p.aiAssistant.FormatResponse(result, context)
}
```

### Response Templates
```go
var ResponseTemplates = map[string]string{
    "status": `📋 Work Overview:
{{range .Schedules}}
• {{.Name}} ({{.Count}}): {{.Preview}}
{{end}}

Recent activity: {{.RecentActivity}}`,
    
    "item_detail": `{{.TypeIcon}} {{.Type | upper}} - {{.Summary}}
📍 Schedule: {{.Schedule | upper}}
🏷️ Tags: {{.Tags | join ", "}}
📅 Created: {{.CreatedAt | timeago}}
{{if .GitContext.Branch}}🌿 Branch: {{.GitContext.Branch}}{{end}}

{{.Content | truncate 300}}`,
}
```

## Integration Features

### Context Awareness
- **Project Context**: Automatically load relevant project info
- **Session Memory**: Remember previous work discussions
- **Git Integration**: Show current branch and uncommitted work
- **Time Tracking**: Display work patterns and productivity insights

### AI-Powered Features
- **Smart Categorization**: Auto-suggest work item types and tags
- **Priority Recommendations**: AI-assisted prioritization
- **Content Generation**: Help draft work item descriptions
- **Progress Tracking**: Identify completion patterns and blockers

### Conversation Integration
- **Natural Language**: Parse work commands from natural conversation
- **Context Preservation**: Maintain work context across conversation turns
- **Multi-Modal**: Support text, voice, and visual work item creation
- **Collaborative**: Enable team-shared work conversations

## Implementation Phases

### Phase 1: Basic Commands (Week 1)
- [x] Command parser and routing
- [x] Core commands (status, list, show, create)
- [x] MCP client integration
- [x] Basic response formatting

### Phase 2: AI Enhancement (Week 2)
- [ ] Smart content generation
- [ ] Context-aware responses
- [ ] Natural language parsing
- [ ] Intelligent recommendations

### Phase 3: Advanced Features (Week 3)
- [ ] Real-time collaboration
- [ ] Voice command support
- [ ] Visual work item creation
- [ ] Advanced analytics and insights

## Success Criteria

- [ ] Commands execute in < 500ms
- [ ] Natural language processing accuracy > 90%
- [ ] Zero data loss during command execution
- [ ] Seamless integration with existing Claude workflows
- [ ] Positive user feedback on work item management efficiency

## User Experience Goals

1. **Effortless**: Work management feels like natural conversation
2. **Intelligent**: AI provides helpful suggestions and automation
3. **Fast**: Quick access to work information without context switching
4. **Reliable**: Consistent behavior and data integrity
5. **Collaborative**: Easy sharing and discussion of work items

This command system will make work tracking as natural as having a conversation with Claude about your projects and tasks.