# API Reference

Complete reference for the Claude Work Tracker MCP server tools and endpoints.

## 🛠️ MCP Server Overview

The Claude Work Tracker MCP server provides programmatic access to all work tracking functionality through the Model Context Protocol. It exposes 8 tools for comprehensive work state management.

### Server Information
- **Name**: `claude-work-tracker`
- **Version**: `1.0.0`
- **Protocol**: Model Context Protocol (MCP)
- **Transport**: stdio

## 🔧 Tools

### `get_work_state`

Get current work state including active todos, recent findings, and session summary.

**Parameters:** None

**Returns:**
```typescript
{
  current_session: string
  active_todos: WorkItem[]
  recent_findings: Finding[]
  session_summary: SessionSummary
  cross_worktree_conflicts?: string[]
}
```

**Example Usage:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_work_state",
    "arguments": {}
  }
}
```

**Response:**
```json
{
  "content": [{
    "type": "text",
    "text": "{\n  \"current_session\": \"20240111_143022_12345\",\n  \"active_todos\": [...],\n  \"recent_findings\": [...],\n  \"session_summary\": {...}\n}"
  }]
}
```

---

### `save_plan`

Save a structured plan with implementation steps for future reference.

**Parameters:**
- `content` (string, required): The main plan description
- `steps` (array of strings, required): List of implementation steps

**Returns:**
```typescript
{
  id: string
  type: "plan"
  content: string
  status: "pending"
  context: GitContext
  session_id: string
  timestamp: string
  metadata: {
    plan_steps: string[]
    priority: "high"
  }
}
```

**Example Usage:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "save_plan",
    "arguments": {
      "content": "Implement user authentication system",
      "steps": [
        "Set up JWT token generation",
        "Create login/logout endpoints",
        "Add middleware for protected routes",
        "Implement user session management"
      ]
    }
  }
}
```

---

### `save_proposal`

Save a proposal or architectural decision with rationale.

**Parameters:**
- `content` (string, required): The proposal description
- `rationale` (string, required): The reasoning behind the proposal

**Returns:**
```typescript
{
  id: string
  type: "proposal"
  content: string
  status: "pending"
  context: GitContext
  session_id: string
  timestamp: string
  metadata: {
    decision_rationale: string
    priority: "high"
  }
}
```

**Example Usage:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "save_proposal",
    "arguments": {
      "content": "Use PostgreSQL for user data storage",
      "rationale": "PostgreSQL provides ACID compliance, excellent performance for relational data, and strong ecosystem support for our Node.js stack"
    }
  }
}
```

---

### `search_work_items`

Search through all work items including todos, plans, proposals, and findings.

**Parameters:**
- `query` (string, required): Search query
- `type` (string, optional): Filter by work item type
  - Valid values: `"todo"`, `"plan"`, `"proposal"`, `"finding"`, `"report"`, `"summary"`

**Returns:**
```typescript
WorkItem[]
```

**Example Usage:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "search_work_items",
    "arguments": {
      "query": "authentication",
      "type": "plan"
    }
  }
}
```

---

### `get_session_summary`

Get summary of current session or a specific session by ID.

**Parameters:**
- `session_id` (string, optional): Specific session ID to get summary for

**Returns:**
```typescript
{
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
```

**Example Usage:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_session_summary",
    "arguments": {
      "session_id": "20240111_143022_12345"
    }
  }
}
```

---

### `get_cross_worktree_status`

Get work status across different git worktrees with optional keyword filtering.

**Parameters:**
- `keyword` (string, optional): Keyword to filter related work

**Returns:**
```typescript
{
  output: string  // Formatted output from work-conflicts.sh or work-status.sh
}
```

**Example Usage:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_cross_worktree_status",
    "arguments": {
      "keyword": "authentication"
    }
  }
}
```

---

### `load_work_state`

Load work state for a specific branch or current branch.

**Parameters:**
- `branch` (string, optional): Branch name to load work state from

**Returns:**
```typescript
{
  output: string  // Output from work.sh load command
}
```

**Example Usage:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "load_work_state",
    "arguments": {
      "branch": "feature-authentication"
    }
  }
}
```

---

### `save_work_state`

Manually save current work state with optional note.

**Parameters:**
- `note` (string, optional): Optional note about the save

**Returns:**
```typescript
{
  output: string  // Output from work.sh save command
}
```

**Example Usage:**
```json
{
  "method": "tools/call",
  "params": {
    "name": "save_work_state",
    "arguments": {
      "note": "Checkpoint after implementing authentication"
    }
  }
}
```

## 📊 Data Types

### `WorkItem`
```typescript
interface WorkItem {
  id: string
  type: 'todo' | 'plan' | 'proposal' | 'finding' | 'report' | 'summary'
  content: string
  status: 'pending' | 'in_progress' | 'completed'
  context: GitContext
  session_id: string
  timestamp: string
  metadata?: WorkItemMetadata
}
```

### `Finding`
```typescript
interface Finding {
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
```

### `GitContext`
```typescript
interface GitContext {
  branch: string
  worktree: string
  remote_url?: string
  working_directory: string
}
```

### `SessionSummary`
```typescript
interface SessionSummary {
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
```

## 🔧 Server Configuration

### Starting the Server

```bash
# Build the server
npm run build

# Start the server
npm start
```

### Claude Code Integration

Add to your Claude Code MCP configuration:

```json
{
  "mcpServers": {
    "claude-work-tracker": {
      "command": "node",
      "args": ["/absolute/path/to/claude-work-tracker/dist/index.js"],
      "env": {}
    }
  }
}
```

### Environment Variables

```bash
# Optional: Override default directories
export CLAUDE_WORK_DIR="$HOME/.claude"
export CLAUDE_WORK_STATE_DIR="$HOME/.claude/work-state"
```

## 🛠️ Error Handling

### Common Error Responses

**Tool Not Found:**
```json
{
  "content": [{
    "type": "text",
    "text": "Error: Unknown tool: invalid_tool_name"
  }],
  "isError": true
}
```

**Missing Required Parameters:**
```json
{
  "content": [{
    "type": "text",
    "text": "Error: Missing required parameters: content and steps"
  }],
  "isError": true
}
```

**File System Errors:**
```json
{
  "content": [{
    "type": "text",
    "text": "Error: Unable to read work state directory"
  }],
  "isError": true
}
```

## 📝 Usage Examples

### Complete Workflow Example

```typescript
// 1. Get current work state
const workState = await mcpClient.call('get_work_state', {})

// 2. Save a new plan
const plan = await mcpClient.call('save_plan', {
  content: "Implement user registration flow",
  steps: [
    "Create user registration form",
    "Add email validation",
    "Implement password hashing",
    "Set up email verification"
  ]
})

// 3. Save a related proposal
const proposal = await mcpClient.call('save_proposal', {
  content: "Use bcrypt for password hashing",
  rationale: "bcrypt is well-tested, has good performance characteristics, and includes built-in salt generation"
})

// 4. Search for related work
const relatedWork = await mcpClient.call('search_work_items', {
  query: "user registration",
  type: "todo"
})

// 5. Save work state with checkpoint
const saveResult = await mcpClient.call('save_work_state', {
  note: "Completed user registration planning phase"
})
```

### Integration with Existing Scripts

The MCP server integrates seamlessly with the existing bash scripts:

```bash
# These commands work alongside MCP server
/work load feature-auth
/work save "checkpoint"
/work view authentication
/work conflicts user
```

## 🔍 Debugging

### Testing Individual Tools

```bash
# Test tools list
echo '{"jsonrpc": "2.0", "method": "tools/list", "id": 1}' | node dist/index.js

# Test get_work_state
echo '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "get_work_state", "arguments": {}}, "id": 1}' | node dist/index.js
```

### Server Logs

The server logs errors to stderr:
```bash
# View server logs
npm start 2> server.log
```

### Data Inspection

```bash
# View saved work items
ls ~/.claude/todos/

# View work intelligence
ls ~/.claude/work-intelligence/

# View findings
ls ~/.claude/findings/
```

## 📚 Related Documentation

- [Installation Guide](installation.md) - Setting up the server
- [Configuration](configuration.md) - Customizing server behavior
- [Architecture](architecture.md) - Understanding the system design
- [Troubleshooting](troubleshooting.md) - Common issues and solutions