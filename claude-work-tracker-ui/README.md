# Claude Work Tracker TUI

A powerful terminal user interface for managing work items, tasks, and project artifacts with Claude Code.

## 🚀 Features

### Core Functionality
- **4-Tab Interface**: NOW, NEXT, LATER, and CLOSED tabs for work organization
- **Project-Wide Scanning**: Automatically discovers and aggregates work from all `.claude-work` directories
- **Multi-Worktree Support**: See work items from all git worktrees in one unified view
- **Real-Time Updates**: File watcher automatically refreshes UI when work items change
- **🤖 Intelligent Automation**: Automatic status transitions based on progress, activity, and Git workflow

### Work Management
- **Smart Scheduling**: Items organized by NOW (active), NEXT (queued), LATER (backlog), CLOSED (completed)
- **Intelligent Transitions**: Automatic status changes based on:
  - Progress milestones (draft → active → in_progress → completed)
  - Activity patterns and focus sessions
  - Git workflow events (commits, branches, PRs)
  - Inactivity detection and decay prevention
- **Quick Actions**: 
  - Press `c` to complete a work item
  - Press `x` to cancel a work item
  - Items automatically move to CLOSED with proper timestamps
- **Status Tracking**: Visual indicators for work status with automation flags

### Search & Navigation
- **Fuzzy Search**: Press `/` to search across title, description, tags, and content
- **Smart Sorting**: Items sorted by newest first (CompletedAt for CLOSED, UpdatedAt for others)
- **Keyboard Navigation**: Efficient keyboard shortcuts for all actions

### Visual Enhancements
- **Colored Status Badges**: 
  - ✅ Green for completed
  - ❌ Red for canceled
  - 📦 Gray for archived
- **Progress Indicators**: See completion percentage for each work item
- **Git Context**: Shows branch and worktree information
- **Automation Indicators**: Unicode symbols show automation status:
  - ◉ Auto-transitioned items
  - ◎ Pending transitions (require confirmation)
  - ⊘ Blocked items
  - ▶ Focus mode (high activity)
  - ⚠ Inactivity warnings
  - ▰▰▰ Activity levels
  - ⎇ Git-linked items
- **Responsive Design**: Adapts to terminal size changes

## 📦 Installation

### From Source
```bash
cd claude-work-tracker-ui
go build -o cw .
cp cw /usr/local/bin/  # Or add to PATH
```

### Using the Alias
Add to your shell configuration:
```bash
alias cw='/path/to/claude-work-tracker-ui/cw'
```

## 🎮 Usage

### Basic Commands
```bash
# Launch the TUI
cw

# The app will automatically scan for work items in:
# - Current directory's .claude-work/
# - All parent directories up to git root
# - All git worktrees in the project
```

### Keyboard Shortcuts

#### Navigation
- `Tab` / `Shift+Tab` - Switch between tabs
- `↑` / `↓` - Navigate items in list
- `Enter` - View full item details
- `Esc` - Back to list / Clear search
- `←` / `→` - Navigate items in detail view
- `q` - Quit application

#### Work Actions
- `c` - Complete current item (NOW tab only)
- `x` - Cancel current item (NOW tab only)
- `d` - Toggle detail view

#### Search
- `/` - Enter search mode
- Type to filter in real-time
- `Enter` - Confirm search
- `Esc` - Clear search

#### Automation & Configuration
- `ctrl+a` - Open automation configuration
- `ctrl+r` - Run automation rules manually
- `ctrl+h` - Toggle automation help/legend

## 📁 Directory Structure

Work items are organized in markdown files:
```
.claude-work/
├── work/
│   ├── now/       # Active work
│   ├── next/      # Queued work
│   ├── later/     # Future work
│   └── closed/    # Completed/canceled
├── artifacts/     # Supporting documents
└── updates/       # Work item history
```

## 🔧 Work Item Format

Work items are markdown files with YAML frontmatter:
```markdown
---
id: work-feature-auth-123456
title: Implement user authentication
schedule: now
status: in_progress
progress_percent: 60
tags: [auth, security, backend]
created_at: 2025-01-18T10:00:00Z
---

# Summary
Implement JWT-based authentication system

## Tasks
- [x] Set up auth middleware
- [x] Create login endpoint
- [ ] Add refresh token logic
- [ ] Write tests
```

## 🛠️ Migration Tools

### Organize Existing Work
```bash
# Show current distribution
./scripts/migrate-work-items.sh status

# Move completed items to CLOSED
./scripts/migrate-work-items.sh completed

# Move 0% progress items from NOW to NEXT
./scripts/migrate-work-items.sh inactive

# Run all migrations
./scripts/migrate-work-items.sh all
```

## 🎨 Customization

### Tab Configuration
The interface shows 4 tabs with item counts:
- `● NOW (X)` - Active work in progress
- `○ NEXT (Y)` - Ready to start
- `⊖ LATER (Z)` - Future/backlog items
- `✓ CLOSED (N)` - Completed/canceled work

### Status Indicators
Work items show status badges:
- `IN_PROGRESS` - Currently being worked on
- `BLOCKED` - Waiting on dependencies
- `COMPLETED` - Successfully finished
- `CANCELED` - Stopped work
- Priority levels (HIGH/MEDIUM/LOW)

## 🤖 Intelligent Automation

### Automatic Status Transitions

The work tracker intelligently manages work item lifecycles:

#### Progress-Based Transitions
- **Draft → Active**: When progress > 0%
- **Active → In Progress**: When progress > 20%
- **In Progress → Completed**: When progress reaches 100%

#### Activity-Based Transitions
- **Focus Detection**: High activity triggers priority updates
- **Inactivity Warnings**: Items stale for >48 hours show warnings
- **Decay Prevention**: Stale items suggest schedule changes

#### Git-Driven Automation
- **Branch Tracking**: Items automatically link to Git branches
- **Commit Integration**: Code changes update activity scores
- **Context Synchronization**: Git metadata auto-updates

### Automation Configuration

Access automation settings via `ctrl+a`:
- **Transition Thresholds**: Customize progress and time limits
- **Activity Detection**: Configure focus session parameters
- **Git Integration**: Enable/disable Git workflow tracking
- **Confirmation Rules**: Set which transitions require approval

### User Control

- **NOW Transitions**: Always require explicit user confirmation
- **Manual Overrides**: Action menu provides manual control
- **Disable Options**: Turn off automation per work item
- **Audit Trail**: All transitions logged with timestamps

## 🔄 Auto-Migration

When completing or canceling items:
1. Status updates automatically
2. File moves from current directory to `closed/`
3. CompletedAt timestamp is set
4. Old file is deleted after successful move

This keeps your work directories clean and organized.

## 🐛 Troubleshooting

### Work items not showing
- Ensure `.claude-work` directory exists
- Check file permissions
- Verify markdown files have correct frontmatter

### Search not working
- Search is case-insensitive
- Fuzzy matching allows partial matches
- Try simpler search terms

### Binary not updating
- Check `which cw` to find active binary
- Ensure PATH includes correct directory
- Rebuild with `go build -o cw .`

### Automation not working
- Check automation is enabled in settings (`ctrl+a`)
- Verify work item has required metadata fields
- Look for automation indicators (◉◎⊘) in the UI
- Check Git context is properly detected

## 🚀 Advanced Features

### Project-Wide Scanning
The TUI automatically discovers all `.claude-work` directories:
- Scans from current directory up to git root
- Includes all git worktrees
- Shows source location for each item

### Real-Time Sync
File changes are detected automatically:
- New items appear immediately
- Edits refresh the display
- Deletions remove items from view

### Smart Filtering
The CLOSED tab intelligently filters:
- Scans all directories (now/next/later)
- Shows only completed/canceled/archived items
- Maintains original directory context

### Automation Engine
Advanced automation capabilities:
- **Hook System**: Event-driven architecture for extensibility
- **Rule Engine**: Priority-based transition rules
- **Activity Analysis**: Focus session detection and intensity tracking
- **Git Integration**: Branch, commit, and PR workflow awareness
- **Future Vision**: Machine learning, team collaboration, and enterprise features

## 📝 License

MIT License - See LICENSE file for details