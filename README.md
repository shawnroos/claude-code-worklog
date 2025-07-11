# Claude Work Tracker

An intelligent work tracking system that seamlessly integrates with Claude Code to provide persistent memory, automatic context management, and smart work relationships across coding sessions.

## 🎯 What It Does

Claude Work Tracker solves the fundamental problem of context loss between AI coding sessions. It automatically:

- **Remembers Everything**: Captures todos, decisions, plans, and insights across sessions
- **Connects Related Work**: Uses AI to automatically link related tasks, decisions, and code changes
- **Provides Smart Context**: Surfaces relevant historical work when you need it
- **Manages Scope**: Intelligently defers and groups future work to maintain focus

## 🚀 Quick Start

```bash
# One-line installation
curl -sSL https://raw.githubusercontent.com/username/claude-work-tracker/main/install.sh | bash

# Build and configure
cd ~/claude-work-tracker
npm install
npm run build

# Add to Claude Code settings
# ~/.claude/claude_code.json
{
  "mcpServers": {
    "work-tracker": {
      "command": "node",
      "args": ["~/claude-work-tracker/dist/index.js"]
    }
  }
}
```

## 📖 How It Works

### 1. Automatic Work Capture

The system automatically captures work items as you code with Claude:

```typescript
// When you create a todo in Claude, it's automatically saved
const todo = {
  content: "Implement user authentication",
  status: "pending",
  context: {
    branch: "feature-auth",
    file: "src/auth/login.ts",
    session: "2025-07-11-auth-implementation"
  }
}
```

### 2. Smart Reference System

The AI-powered reference engine automatically connects related work:

```
📋 Current Work: "Implement user authentication"
   ↗ Continues: "Design auth architecture" (2 days ago)
   ⚠ Conflicts with: "Switch to OAuth-only" (last week)
   ~ Related to: "User profile implementation" (similarity: 0.85)
```

### 3. Contextual Intelligence

When you start work, the system provides relevant context:

```bash
$ /work status

🧠 Smart Suggestions:
  ⚡ HIGH: Review "Auth security audit findings" before continuing
  → MEDIUM: Consider "Session management proposal" for this implementation
  ℹ INFO: Similar work in "feature-user-profiles" branch
```

## 🛠️ Core Features

### Work Tracking
- **Persistent Todos**: Never lose track of tasks between sessions
- **Automatic Capture**: Plans, proposals, and decisions are saved automatically
- **Git Integration**: Work is associated with branches and commits
- **Session Continuity**: Pick up exactly where you left off

### Smart References
- **Automatic Linking**: AI connects related work items without manual effort
- **Similarity Analysis**: Multi-dimensional analysis of content, context, and code
- **Conflict Detection**: Warns about contradictory decisions or approaches
- **Dependency Tracking**: Understands work relationships and prerequisites

### Context Management
- **Two-Tier Storage**: Active work for immediate context, historical for reference
- **Intelligent Promotion**: Brings relevant historical items back when needed
- **Automatic Archival**: Completed work moves to searchable history
- **Efficient Queries**: Fast, context-aware search across all work

### Future Work Management
- **Smart Deferral**: "Let's do A and B now, defer C to future"
- **Automatic Grouping**: Similar future items cluster together
- **Batch Operations**: Promote entire feature groups when ready
- **Scope Control**: Maintain focus without losing good ideas

## 💡 Usage Examples

### Basic Commands

```bash
# Check your current work context
/work status

# Save work state with a note
/work save "Completed auth implementation, ready for review"

# Load work for a specific branch
/work load feature-payments

# Search historical work
/work search "authentication"
```

### MCP Tools (via Claude)

```typescript
// Get smart suggestions for current work
await get_contextual_suggestions()

// Find how two pieces of work relate
await calculate_similarity(
  item_id_1: "impl-auth-2025-07-11",
  item_id_2: "design-auth-2025-07-09"
)

// Visualize work relationships
await visualize_references()
```

### Advanced Workflows

```bash
# Defer work with intelligent grouping
await defer_to_future(
  content: "Add social login providers",
  reason: "Out of scope for MVP"
)

# Promote a group of related work
await promote_work_group("authentication-enhancements")

# Query historical context
await query_history(
  keyword: "performance",
  type: "proposal"
)
```

## 📊 Architecture Overview

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Claude Code   │────▶│   MCP Server     │────▶│ Work State Mgr  │
│                 │     │                  │     │                 │
│  /work commands │     │  - Tools API     │     │ - Active Items  │
│  Auto-capture   │     │  - Handlers      │     │ - History       │
│  Smart suggest  │     │  - Validation    │     │ - Future Work   │
└─────────────────┘     └──────────────────┘     └─────────────────┘
                                 │
                                 ▼
                    ┌───────────────────────┐
                    │  Smart Reference Eng  │
                    │                       │
                    │ - Similarity Analysis │
                    │ - Auto Linking       │
                    │ - Suggestions        │
                    └───────────────────────┘
```

## 🔧 Configuration

### Presentation Modes

```bash
# Set feedback level
~/.claude/scripts/work-presentation.sh mode quiet    # Minimal output
~/.claude/scripts/work-presentation.sh mode summary  # Balanced (default)
~/.claude/scripts/work-presentation.sh mode verbose  # Detailed feedback
```

### Storage Locations

- **Active Work**: `.claude-work/active/` in each project
- **History**: `.claude-work/history/` in each project
- **Future Work**: `.claude-work/future/` in each project
- **Global Scripts**: `~/.claude/scripts/`

## 📚 Documentation

- [**Installation Guide**](docs/installation.md) - Detailed setup instructions
- [**User Guide**](docs/user-guide.md) - Complete usage documentation
- [**API Reference**](docs/api-reference.md) - MCP tools and endpoints
- [**Architecture**](docs/architecture.md) - System design details
- [**Smart References**](docs/smart-references.md) - How the AI linking works

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

## 📄 License

MIT License - See [LICENSE](LICENSE) for details.