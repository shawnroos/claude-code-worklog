# Claude Code Work Tracking System

A comprehensive work tracking system for Claude Code that provides persistent todo management and cross-worktree awareness.

## 🚀 Quick Install

**One-line installation:**
```bash
curl -sSL https://raw.githubusercontent.com/shawnroos/claude-code-worklog/main/install.sh | bash
```

**After installation:**
```bash
# Optional: Run setup wizard to customize your experience
~/.claude/scripts/setup-wizard.sh

# Test the system
~/.claude/scripts/work-presentation.sh test
```

**That's it!** The system starts working automatically in your next Claude session.

---

## 📖 Table of Contents

- [🚀 Quick Install](#-quick-install)
- [✨ Features](#features)
- [🎮 Usage](#usage)
- [⚙️ Configuration](#configuration)
- [🔧 Manual Installation](#manual-installation)
- [❓ FAQ](#faq)
- [🗑️ Uninstall](#-uninstall)

---

## Features

### 🎯 Core Functionality
- **Persistent Todo Tracking**: Todos survive across Claude sessions
- **Git Context Awareness**: Associates work with specific branches and worktrees
- **Cross-Worktree Intelligence**: Detects related work across different feature branches
- **Hybrid Architecture**: Local efficiency + global visibility when needed

### 🎛️ Presentation Control
- **Three modes**: quiet, summary (default), verbose
- **Customizable styling**: minimal_colored (default), modern, classic, minimal
- **Smart notifications**: Session summaries, conflict alerts, sync status

### 🔄 Automatic Workflows
- **Session hooks**: Automatically capture completed work and save incomplete todos
- **Background sync**: Updates global state without blocking your workflow
- **Conflict detection**: Alerts when related work exists in other worktrees

## Manual Installation

If you prefer to install manually or want to understand the system better:

### Required Files
- `CLAUDE.md` - Global coding standards and preferences
- `settings.local.json` - Hook configuration and permissions
- `work-tracking-config.json` - Presentation and behavior settings
- `scripts/` - All automation scripts

### Directory Structure
```
~/.claude/
├── CLAUDE.md                     # Global preferences
├── settings.local.json           # Hooks & permissions
├── work-tracking-config.json     # Presentation config
├── scripts/                      # Automation scripts
│   ├── session-complete.sh       # Session end hook
│   ├── tool-complete.sh          # Todo update hook
│   ├── update-global-state.sh    # Global state aggregation
│   ├── restore-todos.sh          # Todo restoration
│   ├── work-*.sh                 # Cross-worktree commands
│   └── work-presentation.sh      # Display control
├── work-state/                   # Global work aggregation
├── projects/                     # Session conversation logs
└── todos/                        # Per-session todo files
```

## Usage

### Basic Commands
```bash
# Check for incomplete todos from previous sessions
~/.claude/scripts/restore-todos.sh

# Manually save current work state
~/.claude/scripts/save.sh

# Global overview of active work
~/.claude/scripts/work-status.sh

# Find related work in other worktrees
~/.claude/scripts/work-conflicts.sh auth

# Load full project context (higher token usage)
~/.claude/scripts/work-context.sh ProjectName
```

### Presentation Control
```bash
# Set presentation mode
~/.claude/scripts/work-presentation.sh mode quiet|summary|verbose

# Test current settings
~/.claude/scripts/work-presentation.sh test
```

### Git Worktree Integration
The system automatically detects and tracks:
- Current branch name
- Worktree location (main vs feature worktrees)
- Cross-worktree todo relationships
- Project-level work aggregation

## Configuration

### Presentation Modes
- **quiet**: Minimal feedback, no session summaries
- **summary**: Balanced feedback with session summaries (default)
- **verbose**: Detailed feedback with todo diffs and worktree details

### Emoji Styles
- **minimal_colored**: ✓ ○ ● ! with colors (default)
- **modern**: ✅ 🔄 ⚡ ⚠️ 
- **classic**: [✓] [○] [●] [!]
- **minimal**: ✓ ○ ● ! (no colors)

## How It Works

### Session Lifecycle
1. **Start session**: Optionally restore incomplete todos with `restore-todos.sh`
2. **During work**: TodoWrite hook backs up current state
3. **Session end**: Stop hook captures completed work and preserves incomplete todos
4. **Background sync**: Global state updated with worktree context

### Cross-Worktree Intelligence
- Maintains global project overview across all worktrees
- Detects potential conflicts when similar work exists elsewhere
- Preserves context about which branch/worktree todos originated from
- Enables smart restoration based on current git context

## Example Output

```bash
=== Work Session Summary ===
★ **Session Complete** | Worktree: `feature-auth` | Branch: `feature-auth`
✓ **Completed:** 3 todos  
○ **Pending:** 2 todos saved for next session

↻ **Work Sync:** Updating global work state for feature-auth

! **Potential Conflicts:** Found 1 related todos in other worktrees
💡 Run: `work-conflicts` to review
```

## Benefits

### For Individual Workflows
- Never lose track of incomplete work between sessions
- Context-aware todo restoration based on current branch
- Visual feedback about work progress and conflicts

### For Multi-Worktree Development
- Coordinate work across feature branches
- Prevent duplicate effort on similar tasks
- Maintain project-level visibility while preserving worktree isolation

### For Team Collaboration
- Standardized work tracking across team members
- Git-aware context preservation
- Configurable presentation to match team preferences

## ❓ FAQ

### How does it work?
The system uses Claude Code's hooks feature to automatically capture your work when sessions end and restore context when sessions begin.

### Will it interfere with my existing Claude setup?
No! The installer safely merges with your existing configuration and creates backups. You can uninstall anytime.

### Does it work with git worktrees?
Yes! The system is specifically designed for git worktree workflows and provides cross-worktree intelligence.

### Can I customize the appearance?
Absolutely! Run `~/.claude/scripts/setup-wizard.sh` to customize presentation modes and visual styles.

### What if I don't like it?
Easy! Run `~/.claude/uninstall.sh` to remove everything safely while preserving your data.

### Is my data safe?
Yes! The system never sends your data anywhere - everything stays local. Plus, backups are created during install/uninstall.

---

## 🗑️ Uninstall

To remove the work tracking system:

```bash
~/.claude/uninstall.sh
```

This will:
- ✅ Remove all work tracking scripts and configurations  
- ✅ Create a backup of everything before removal
- ✅ Preserve your work history and conversation data
- ✅ Clean up hooks and permissions automatically

To reinstall later, just use the one-line installer again!

---

## 🤝 Contributing

Found a bug or have an idea? Open an issue at: https://github.com/shawnroos/claude-code-worklog/issues

---

## License

This work tracking system is part of the Claude Code ecosystem and follows the same usage guidelines.