# Claude Code Work Tracking System

A comprehensive work tracking system for Claude Code that provides persistent todo management, cross-worktree awareness, work intelligence capture, and **MCP server integration**.

## 🚀 Quick Start

**One-line installation:**
```bash
curl -sSL https://raw.githubusercontent.com/shawnroos/claude-work-tracker/main/install.sh | bash
```

**After installation:**
```bash
# Test the system
~/.claude/scripts/work-presentation.sh test

# Build and start the MCP server
npm run build
npm start

# Try the /work command
/work
```

## ✨ Features

### 🎯 **Core Functionality**
- **Persistent Todo Tracking** - Todos survive across Claude sessions
- **Work Intelligence Capture** - Automatically captures plans, proposals, and strategic insights
- **Git Context Awareness** - Associates work with specific branches
- **Local Project Focus** - Efficient tracking within current project
- **MCP Server Integration** - Programmatic access via Model Context Protocol

### 🧠 **Work Intelligence System**
- **Plan Capture** - Automatically saves plans from `exit_plan_mode` and planning discussions
- **Proposal Tracking** - Captures architectural decisions and strategic recommendations
- **Strategic Insights** - Extracts key insights from research and analysis
- **Decision Rationale** - Preserves reasoning behind important decisions
- **Session Summaries** - Comprehensive session-end summaries
- **Cross-Session Continuity** - Links plans → implementations → outcomes

### ⚡ **Manual Commands**
- **`/work load`** - Restore work state for current or specific branch
- **`/work save`** - Save current work state with optional notes
- **`/work view`** - View work overview with optional filtering
- **`/work status`** - Current session status and recent activity

## 📚 Documentation

- **[Installation Guide](docs/installation.md)** - Detailed installation and setup
- **[API Reference](docs/api-reference.md)** - MCP server tools and endpoints
- **[Architecture](docs/architecture.md)** - System design and data flow
- **[Configuration](docs/configuration.md)** - Customization and settings
- **[Troubleshooting](docs/troubleshooting.md)** - Common issues and solutions
- **[Developer Guide](docs/development.md)** - Contributing and building

## 🔧 Quick Commands

### Basic Usage
```bash
/work                    # Show help and available commands
/work load               # Load work state for current branch
/work save              # Save current work state
/work view              # Global work overview
/work status            # Current session status
```

### Advanced Usage
```bash
/work load feature-auth  # Load work state from specific branch
/work save "checkpoint"  # Save with descriptive note
/work view auth         # View auth-related work
```

## 🛠️ MCP Server

The system includes an MCP server for programmatic access:

```bash
# Build and start
npm run build
npm start

# Configure in Claude Code
{
  "mcpServers": {
    "claude-work-tracker": {
      "command": "node",
      "args": ["/path/to/claude-work-tracker/dist/index.js"]
    }
  }
}
```

**Available Tools:** `get_work_state`, `save_plan`, `save_proposal`, `search_work_items`, `get_session_summary`, `load_work_state`, `save_work_state`

## 🎯 Work Intelligence Types

The system captures and categorizes:
- **Plans** - Structured implementation plans with steps
- **Proposals** - Architectural decisions and recommendations  
- **Strategic Insights** - Key insights from research and analysis
- **Decision Rationale** - Reasoning behind important decisions
- **Session Summaries** - Comprehensive session outcomes
- **Discovered Plans** - Plans identified during code analysis

## 📈 Benefits

### For Individual Workflows
- Never lose track of plans, proposals, or incomplete work
- Context-aware restoration based on current branch
- Comprehensive work intelligence across sessions

### For Branch-Based Development
- Track work separately for each git branch
- Maintain clean work history per feature
- Easy context switching between branches

### For AI-Assisted Development
- Preserve Claude's planning and strategic insights
- Maintain continuity of architectural decisions
- Enable sophisticated work intelligence queries

## 🗑️ Uninstall

```bash
~/.claude/uninstall.sh
```

Creates backups and safely removes all components while preserving your work history.

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and contribution guidelines.

## 📋 License

MIT License - Part of the Claude Code ecosystem.