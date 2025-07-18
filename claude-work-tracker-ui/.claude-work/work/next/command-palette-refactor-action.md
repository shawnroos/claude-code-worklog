---
id: work-command-palette-refactor-1737211200
title: Command Palette with Work Item Refactor Action
description: Implement a universal cmd+k command palette system with fuzzy search, starting with a refactor action that updates outdated work items based on codebase changes
schedule: next
created_at: 2025-01-18T12:00:00Z
updated_at: 2025-01-18T12:00:00Z
overview_updated: 2025-01-18T12:00:00Z
updates_ref: updates/work-command-palette-refactor-1737211200.md
git_context:
  branch: main
  worktree: main
  working_directory: /Users/shawnroos/claude-work-tracker
session_number: session-1737211200
technical_tags: [ui, command-palette, refactoring, bubble-tea, fuzzy-search]
artifact_refs: []
metadata:
  status: active
  priority: high
  estimated_effort: large
  progress_percent: 0
  artifact_count: 0
  activity_score: 30.0
  blocked_by: []
---

# Command Palette with Work Item Refactor Action

*Last updated: 2025-01-18 12:00*

## Overview

Implement a universal command palette system (cmd+k) that provides fuzzy search for all actions across the TUI. The first flagship action will be a work item refactor command that intelligently updates outdated work items based on current codebase state.

## Command Palette Design

### Core Features
- **Universal Access**: cmd+k available from any view
- **Fuzzy Search**: Real-time filtering as user types
- **Keyboard Navigation**: Arrow keys, enter to select, esc to cancel
- **Extensible**: Action registry pattern for easy addition of new commands
- **Context Aware**: Show relevant actions based on current view/selection

### UI Components
```go
// Command palette overlay structure
type CommandPalette struct {
    tea.Model
    searchInput  textinput.Model
    commandList  list.Model
    overlay      bool
    actions      []Action
    filtered     []Action
}

// Action interface for extensibility
type Action interface {
    ID() string
    Title() string
    Description() string
    Icon() string
    Execute(context ActionContext) tea.Cmd
    IsAvailable(context ActionContext) bool
}
```

### Visual Design
- Semi-transparent modal overlay
- Centered palette with rounded corners
- Search input at top with magnifying glass icon
- Filtered action list below with descriptions
- Keyboard shortcuts displayed inline

## Refactor Action Specification

### Purpose
Detect and update work items that have become outdated due to:
- API changes in the codebase
- Renamed functions/methods
- Moved files or restructured directories
- Deprecated patterns replaced with new ones
- Changed dependencies or configurations

### Detection Logic
1. **Scan References**: Extract code references from work item content
2. **Validate Existence**: Check if referenced files/functions still exist
3. **Detect Changes**: Identify renamed or moved elements
4. **Analyze Context**: Understand surrounding code changes

### Refactoring Process
1. **One-Shot Update**: AI analyzes changes and proposes updates
2. **Preserve Structure**: Maintain original document flow and intent
3. **Update References**: Fix file paths, function names, API calls
4. **Modernize Examples**: Update code snippets to current patterns
5. **Side-by-Side Diff**: Present changes for user review

### Diff Presentation
```
┌─────────────────────────────────────────────────┐
│  Refactor Work Item: "Add Authentication"       │
├─────────────────────────┬───────────────────────┤
│      Original           │      Refactored       │
├─────────────────────────┼───────────────────────┤
│ Call auth.Login()       │ Call auth.SignIn()    │
│ in server/auth.go:45    │ in pkg/auth/auth.go:62│
│                         │                       │
│ Use config.AuthSecret   │ Use env.AUTH_SECRET   │
│                         │                       │
└─────────────────────────┴───────────────────────┘
[A]ccept  [R]eject  [E]dit  [ESC]Cancel
```

## Implementation Phases

### Phase 1: Command Palette Infrastructure
- [ ] Create `internal/views/command_palette.go`
- [ ] Implement modal overlay rendering
- [ ] Add search input with bubble textinput
- [ ] Create filtered command list with bubble list
- [ ] Handle keyboard navigation and selection

### Phase 2: Action System
- [ ] Define Action interface
- [ ] Create ActionRegistry for command registration
- [ ] Implement ActionContext for passing state
- [ ] Add action availability checking
- [ ] Build command execution framework

### Phase 3: Integration
- [ ] Modify app.go to handle cmd+k globally
- [ ] Add command palette instance to app state
- [ ] Implement overlay rendering in main view
- [ ] Preserve underlying view state
- [ ] Handle focus management

### Phase 4: Refactor Action
- [ ] Create RefactorWorkItemAction
- [ ] Implement outdated detection logic
- [ ] Build codebase change analyzer
- [ ] Generate refactored content
- [ ] Create diff view component

### Phase 5: Enhanced Features
- [ ] Add recent commands section
- [ ] Implement command history
- [ ] Add keyboard shortcut hints
- [ ] Create command categories
- [ ] Add command search highlighting

## Technical Implementation

### Key Bindings Integration
```go
// In app.go Update method
case "ctrl+k", "cmd+k":
    if !a.showCommandPalette {
        a.showCommandPalette = true
        a.commandPalette.Reset()
        return a, a.commandPalette.Init()
    }
```

### Action Registration
```go
// Register actions during app initialization
registry := NewActionRegistry()
registry.Register(NewRefactorWorkItemAction(dataClient))
registry.Register(NewQuickSwitchTabAction())
registry.Register(NewSearchWorkItemsAction())
registry.Register(NewCreateWorkItemAction())
```

### Refactor Implementation
```go
type RefactorWorkItemAction struct {
    dataClient *data.EnhancedClient
    analyzer   *CodebaseAnalyzer
}

func (r *RefactorWorkItemAction) Execute(ctx ActionContext) tea.Cmd {
    workItem := ctx.SelectedWork
    
    // Detect outdated references
    outdated := r.analyzer.FindOutdatedReferences(workItem)
    
    // Generate refactored version
    refactored := r.RefactorContent(workItem, outdated)
    
    // Show diff view
    return ShowDiffView(workItem, refactored)
}
```

## User Experience

### Command Flow
1. User presses cmd+k from any view
2. Command palette overlays current view
3. User types to filter commands (e.g., "refac")
4. "Refactor outdated work item" appears at top
5. User presses enter to execute
6. If on work item: immediate refactor
7. If not: prompt to select work item
8. Diff view shows proposed changes
9. User accepts/rejects/edits changes

### Performance Considerations
- Lazy load command descriptions
- Cache filtered results during typing
- Debounce search input
- Preload common commands
- Background index for faster search

## Command Palette Actions

Based on analysis of the codebase and existing features, here's the comprehensive list of commands organized by implementation status:

### ✅ Already Implemented (Keyboard Shortcuts)
These features exist but would benefit from command palette access:

1. **Search in Current Tab** - `/` key - Filter work items in current schedule
2. **Complete Work Item** - `c` key - Mark as completed and move to CLOSED (NOW tab only)
3. **Cancel Work Item** - `x` key - Mark as canceled and move to CLOSED (NOW tab only)
4. **Toggle Detail View** - `d` key - Show/hide item details in list
5. **View Full Post** - `enter` key - Open full markdown view
6. **Switch Tabs** - `tab`/`shift+tab` - Navigate between NOW/NEXT/LATER/CLOSED
7. **Navigate Items** - `←/→` arrows - Previous/next item in full view
8. **Quit** - `q` key - Exit application

### 🔨 Core Features (Data Layer Exists)
These have backend support but need command palette UI:

9. **Load Work by Schedule** - GetWorkBySchedule() method exists
10. **Search All Work** - SearchWork() method in EnhancedClient
11. **Create Work Item** - WriteWork() method in MarkdownIO
12. **Update Work Item** - WriteWork() handles updates too
13. **Delete Work Item** - File system operations available
14. **Move Between Schedules** - Schedule field can be updated

### 💡 Recommended New Features

#### Work Item Management
15. **Archive Work Item** - Move old completed items to archive directory
16. **Update Work Progress** - Set progress percentage (0-100)
17. **Change Priority** - Update priority (critical/high/medium/low)
18. **Edit Work Item** - Open full editor interface
19. **Duplicate Work Item** - Create copy with new ID
20. **Split Work Item** - Break into multiple smaller items
21. **Merge Work Items** - Combine related items

#### Artifact Operations
22. **Create Artifact** - New plan/update/analysis/decision
23. **Link Artifact** - Associate with current work item
24. **View Artifacts** - Show all linked artifacts
25. **Export Artifacts** - Export in markdown/PDF/HTML
26. **Update References** - Refresh artifact reference counts

#### Navigation & Search
27. **Quick Tab Switch** - Jump directly to specific tab
28. **Go to Work Item** - Navigate by title or ID
29. **Recent Items** - Show last 10 accessed items
30. **Related Work** - Find items with similar tags
31. **Search Across All Tabs** - Global search (enhancement of existing)

#### Bulk Operations
32. **Select Multiple** - Enter multi-select mode
33. **Bulk Complete** - Complete all selected items
34. **Bulk Move** - Move selected to different schedule
35. **Bulk Tag** - Add/remove tags from selection
36. **Bulk Export** - Export selected as bundle
37. **Bulk Archive** - Archive old completed items

#### View Controls
38. **Toggle Preview** - Switch raw/rendered markdown
39. **Sort by Date** - Newest/oldest first (partial support exists)
40. **Sort by Priority** - Critical → Low
41. **Sort by Progress** - Most/least complete
42. **Filter by Status** - Show specific statuses only
43. **Filter by Tags** - Show items with specific tags
44. **Group by Tag** - Organize by technical tags

#### Updates & Activity
45. **Add Update** - Quick update to current item
46. **View Updates** - Show update history
47. **Session Summary** - Generate current session report
48. **Activity Timeline** - Recent changes across project
49. **Clear Activity** - Reset activity decay warnings

#### Git Integration
50. **Show Git Context** - Display branch/worktree info (data exists)
51. **Copy Branch Name** - Copy associated branch to clipboard
52. **Switch Branch** - Checkout work item's branch
53. **Create Branch** - New branch for work item

#### Refactoring (Flagship Feature)
54. **Refactor Outdated** - Update based on codebase changes
55. **Validate References** - Check all code references
56. **Update Paths** - Fix moved/renamed files
57. **Modernize Code** - Update to current patterns
58. **Batch Refactor** - Refactor multiple items

#### Import/Export
59. **Import Markdown** - Import work from .md files
60. **Import CSV** - Bulk import from spreadsheet
61. **Export Project** - Full backup as archive
62. **Export Report** - Generate status report
63. **Export Templates** - Save item as template

#### Help & Settings
64. **Show Shortcuts** - Display keyboard reference
65. **Command Help** - Explain selected command
66. **Open Docs** - Link to documentation
67. **Toggle Theme** - Switch color themes
68. **Configure Keys** - Customize keybindings

#### System Commands  
69. **Refresh All** - Force reload all data (loadWorkItems exists)
70. **Clear Cache** - Reset render/embed caches (cache maps exist)
71. **Rebuild Index** - Recreate search index
72. **Run Diagnostics** - Check data integrity
73. **Show Debug Info** - Display system state

## Implementation Priority

### Phase 1 (MVP)
- Command palette infrastructure
- Search command (already exists as '/')
- Quick tab switch
- Create/Complete/Cancel work items
- Refactor outdated action

### Phase 2 (Enhanced)
- Bulk operations
- View controls and filtering
- Git integration commands
- Import/export basics

### Phase 3 (Advanced)
- Update tracking
- Artifact management
- Advanced refactoring
- Template system

## Future Extensions

### Advanced Refactoring
- **Semantic Understanding**: Use LSP for precise code analysis
- **Multi-File Refactor**: Update across related work items
- **Auto-Update Mode**: Continuous background checking
- **Refactor Profiles**: Different update strategies
- **History Tracking**: See all refactors applied

### Integration Points
- Git hooks to trigger refactor suggestions
- VS Code extension for seamless editing
- CLI commands for scriptable access
- Web UI with same command palette
- API for external tool integration

## Success Metrics

1. **Performance**: Commands execute in <100ms
2. **Accuracy**: 95%+ relevant commands in top 3 results  
3. **Adoption**: 80%+ users utilize command palette daily
4. **Extensibility**: New commands added in <30 min
5. **Satisfaction**: Reduced time to complete common tasks

This command palette system will transform the work tracking experience from menu-driven to command-driven, making power users more efficient while maintaining discoverability for new users.