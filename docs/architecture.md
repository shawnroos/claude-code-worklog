# Architecture Documentation

Deep dive into the Claude Code Work Tracking System's architecture, design decisions, and data flow.

## 🏗️ System Overview

The Claude Code Work Tracking System is a **local work intelligence platform** that captures, preserves, and organizes development work within your current project:

```
Claude Code ↔ Hook System ↔ Work Intelligence Engine ↔ MCP Server
     ↓             ↓                    ↓                  ↓
  Sessions    Work Capture        Data Storage      Programmatic API
     ↓             ↓                    ↓                  ↓
  Git Context  Intelligence      Local Project      External Tools
              Classification        Storage
```

## 🎯 Core Components

### 1. **Hook System** - Event-Driven Capture
- **Session Hooks**: Capture session start/end events
- **Tool Hooks**: Capture tool usage and outputs
- **Plan Hooks**: Capture planning and decision-making
- **Intelligence Hooks**: Extract strategic insights

### 2. **Work Intelligence Engine** - Data Processing
- **Classification**: Categorize work items by type and intent
- **Context Enrichment**: Add git, temporal, and semantic context
- **Cross-Reference**: Link related work across sessions
- **Aggregation**: Summarize work patterns and insights

### 3. **Storage Layer** - Local Persistence
- **Local Session State**: Immediate work context
- **Project State**: Cross-session work aggregation
- **Work Intelligence**: Plans, proposals, insights
- **Branch Context**: Git branch-specific storage

### 4. **MCP Server** - Programmatic Interface
- **Tool Endpoints**: RESTful-style work operations
- **State Management**: Centralized work state access
- **Integration Layer**: Bridge to external tools

## 📊 Data Architecture

### Data Flow Diagram

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Claude Code   │    │   Hook System   │    │ Work Intelligence│
│                 │    │                 │    │    Engine       │
│ • Sessions      │───▶│ • session-*.sh  │───▶│ • Classification│
│ • Tool Usage    │    │ • tool-*.sh     │    │ • Context       │
│ • Planning      │    │ • plan-*.sh     │    │ • Aggregation   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │              ┌─────────────────┐              │
         │              │  Storage Layer  │              │
         └──────────────▶│                 │◀─────────────┘
                        │ • Local State   │
                        │ • Project State │
                        │ • Global State  │
                        │ • Intelligence  │
                        └─────────────────┘
                                 │
                        ┌─────────────────┐
                        │   MCP Server    │
                        │                 │
                        │ • API Endpoints │
                        │ • State Manager │
                        │ • Integration   │
                        └─────────────────┘
```

### Data Models

#### **WorkItem** - Core Work Unit
```typescript
interface WorkItem {
  id: string                    // Unique identifier
  type: WorkItemType           // Classification
  content: string              // Primary content
  status: WorkItemStatus       // Current state
  context: GitContext          // Git information
  session_id: string           // Session association
  timestamp: string            // Creation time
  metadata?: WorkItemMetadata  // Additional data
}
```

#### **Work Intelligence Taxonomy**
```typescript
type WorkItemType = 
  | 'todo'              // Action items
  | 'plan'              // Structured implementations  
  | 'proposal'          // Architectural decisions
  | 'finding'           // Research results
  | 'report'            // Analysis summaries
  | 'summary'           // Session outcomes
  | 'strategic_insight' // Key insights
  | 'decision_rationale'// Decision reasoning
```

#### **Context Enrichment**
```typescript
interface GitContext {
  branch: string              // Current branch
  worktree: string           // Worktree location
  remote_url?: string        // Repository URL
  working_directory: string  // Local path
}

interface SessionContext {
  session_id: string         // Unique session
  start_time: string        // Session start
  duration?: number         // Session length
  tool_usage: ToolUsage[]   // Tools used
}
```

## 🔄 Hook System Architecture

### Hook Execution Flow

```
Claude Code Event → Hook Trigger → Data Capture → Intelligence Processing → Storage
```

### Hook Types and Responsibilities

#### **Session Hooks**
- **`session-init.sh`**: Session startup, context restoration
- **`session-complete.sh`**: Session teardown, state preservation

#### **Tool Hooks**
- **`tool-complete-enhanced.sh`**: General tool capture
- **`tool-complete-plan-capture.sh`**: Plan-specific capture

#### **Intelligence Hooks**
- **`update-work-intelligence.sh`**: Cross-session aggregation
- **`update-global-state.sh`**: Multi-project intelligence

### Hook Communication Protocol

```bash
# Input: JSON via stdin
{
  "sessionId": "20240111_143022_12345",
  "toolName": "exit_plan_mode",
  "toolInput": {"plan": "Implementation plan..."},
  "toolOutput": "Plan created successfully",
  "transcriptPath": "/path/to/session.jsonl",
  "workingDirectory": "/path/to/project"
}

# Output: File system updates + log entries
```

## 🗄️ Storage Architecture

### Storage Hierarchy

```
~/.claude/                          # Global configuration
├── work-state/                     # Project state storage
│   └── projects/                   # Per-project state
│       └── {project}/
│           └── ACTIVE_WORK.md      # Project overview
├── work-intelligence/              # Intelligence capture
│   └── {session}_{type}.json       # Individual intelligence items
├── todos/                          # Session todos
│   └── {session}-agent-{id}.json   # Todo snapshots
├── findings/                       # Tool findings
│   └── {session}_{tool}.json       # Tool outputs
└── projects/                       # Session logs
    └── {project}/                  # Conversation transcripts
```

### Local Project State

```
{project}/.claude-work/             # Local work state
├── WORK_HISTORY.md                 # Chronological work log
├── PENDING_TODOS.json              # Incomplete todos
└── current_todos.json              # Live session backup
```

### Data Persistence Strategy

1. **Immediate Persistence**: Critical data saved immediately
2. **Batch Updates**: Non-critical data aggregated periodically
3. **Incremental Backups**: State changes tracked incrementally
4. **Cross-Session Linking**: Related work linked across sessions

## 🧠 Work Intelligence Engine

### Intelligence Classification

```typescript
class WorkIntelligenceClassifier {
  classify(content: string, context: CaptureContext): WorkItemType {
    // Pattern matching for different intelligence types
    if (containsPlanPatterns(content)) return 'plan'
    if (containsProposalPatterns(content)) return 'proposal'
    if (containsInsightPatterns(content)) return 'strategic_insight'
    if (containsDecisionPatterns(content)) return 'decision_rationale'
    // ... more classification logic
  }
}
```

### Pattern Recognition

#### **Plan Detection**
- Numbered lists (1., 2., 3.)
- Step indicators ("Step 1", "Phase 1")
- Implementation language ("implement", "create", "build")
- Structured content from `exit_plan_mode`

#### **Proposal Detection**  
- Recommendation language ("I recommend", "I suggest")
- Architectural terms ("architecture", "design", "approach")
- Decision indicators ("decision", "choice", "option")
- Rationale patterns ("because", "rationale", "reason")

#### **Insight Detection**
- Analysis language ("analysis", "insight", "finding")
- Strategic terms ("strategy", "approach", "pattern")
- Research indicators ("research", "investigation", "study")

### Context Enrichment Process

```typescript
interface ContextEnrichment {
  temporal: {
    timestamp: string
    session_duration: number
    related_sessions: string[]
  }
  spatial: {
    git_context: GitContext
    file_relationships: string[]
    project_context: ProjectContext
  }
  semantic: {
    related_work: WorkItem[]
    keyword_tags: string[]
    topic_clusters: string[]
  }
}
```

## 🌐 MCP Server Architecture

### Server Components

```typescript
class WorkTrackingMCPServer {
  private server: MCPServer
  private workStateManager: WorkStateManager
  private toolRegistry: ToolRegistry
  
  // Core server lifecycle
  async initialize() { /* ... */ }
  async handleRequest() { /* ... */ }
  async shutdown() { /* ... */ }
}
```

### Tool Architecture

```typescript
interface MCPTool {
  name: string
  description: string
  inputSchema: JSONSchema
  handler: (params: any) => Promise<McpToolResponse>
}

class ToolRegistry {
  private tools: Map<string, MCPTool>
  
  register(tool: MCPTool) { /* ... */ }
  execute(name: string, params: any) { /* ... */ }
}
```

### State Management

```typescript
class WorkStateManager {
  // Data access layer
  getCurrentWorkState(): WorkState
  saveWorkItem(item: WorkItem): void
  searchWorkItems(query: string): WorkItem[]
  
  // Intelligence operations
  savePlan(content: string, steps: string[]): WorkItem
  saveProposal(content: string, rationale: string): WorkItem
  
  // Cross-worktree operations
  getCrossWorktreeConflicts(): string[]
  aggregateGlobalState(): void
}
```

## 🔀 Branch-Based Intelligence

### Branch Context Management

```typescript
interface BranchContext {
  current_branch: string        // Active git branch
  branch_type: 'main' | 'feature' | 'hotfix' | 'bugfix'
  base_branch: string          // Parent branch
  work_items: WorkItem[]       // Branch-specific work
}
```

### Branch Switching

```typescript
class BranchManager {
  switchContext(newBranch: string): void {
    // 1. Save current branch work state
    this.saveCurrentBranchState()
    
    // 2. Load new branch work state
    const branchState = this.loadBranchState(newBranch)
    
    // 3. Restore work context
    this.restoreWorkContext(branchState)
  }
}
```

### Local State Organization

```typescript
interface LocalState {
  project: ProjectState
  current_branch: BranchContext
  work_history: WorkItem[]
  intelligence: WorkIntelligence[]
}
```

## 🔧 Integration Architecture

### Claude Code Integration

```typescript
interface ClaudeCodeIntegration {
  hooks: {
    session_start: HookHandler
    session_complete: HookHandler
    tool_complete: HookHandler
  }
  commands: {
    work: CommandHandler
  }
  configuration: {
    settings: ClaudeSettings
    permissions: PermissionSet
  }
}
```

### External Tool Integration

```typescript
interface ExternalIntegration {
  mcp_server: MCPServerEndpoint
  bash_scripts: BashScriptSet
  file_system: FileSystemAdapter
  git_integration: GitAdapter
}
```

## 📈 Performance Considerations

### Scalability Design

1. **Incremental Processing**: Only process changes, not full state
2. **Lazy Loading**: Load data on-demand
3. **Caching Strategy**: Cache frequently accessed data
4. **Background Processing**: Non-critical operations run async

### Memory Management

```typescript
class MemoryManager {
  private cache: LRUCache<string, WorkItem>
  private maxCacheSize: number = 1000
  
  // Efficient data access patterns
  getWorkItem(id: string): WorkItem | null
  evictOldItems(): void
  optimizeMemoryUsage(): void
}
```

### File System Optimization

- **Structured Directories**: Logical organization for fast access
- **JSON Streaming**: Large datasets processed incrementally
- **Compression**: Historical data compressed for space efficiency
- **Indexing**: Quick lookup indices for common queries

## 🛡️ Security Architecture

### Data Privacy

1. **Local Storage**: All data stays on user's machine
2. **No Network Transmission**: No data sent to external services
3. **Access Control**: File permissions restrict access
4. **Audit Trails**: All operations logged for transparency

### Secure Defaults

```typescript
interface SecuritySettings {
  file_permissions: '644' | '755'
  directory_permissions: '755'
  sensitive_data_handling: 'encrypt' | 'exclude'
  log_retention: number // days
}
```

## 🔮 Future Architecture Considerations

### Planned Enhancements

1. **Distributed Intelligence**: Multi-machine synchronization
2. **Advanced Analytics**: Machine learning for pattern recognition
3. **Integration Ecosystem**: Plugin architecture for extensions
4. **Real-time Collaboration**: Team-based work intelligence

### Extensibility Design

```typescript
interface ExtensionAPI {
  registerHook(event: string, handler: HookHandler): void
  registerTool(tool: MCPTool): void
  registerClassifier(classifier: IntelligenceClassifier): void
  registerStorage(adapter: StorageAdapter): void
}
```

This architecture provides a robust, scalable foundation for comprehensive work intelligence capture and management while maintaining performance and security.