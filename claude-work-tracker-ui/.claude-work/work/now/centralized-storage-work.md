---
id: work-centralized-storage-1752531600
title: Implement centralized work storage with git context metadata
description: Redesign work storage to use project-root centralization with worktree/branch context as metadata, solving conflicts and enabling true project-wide visibility
schedule: now
created_at: 2025-07-17T18:20:00Z
updated_at: 2025-07-17T18:20:00Z
overview_updated: 2025-07-17T18:20:00Z
updates_ref: updates/work-centralized-storage-1752531600.md
git_context:
  branch: feature/tasks
  worktree: worktrees/tasks
  working_directory: /Users/shawnroos/claude-work-tracker/worktrees/tasks/claude-work-tracker-ui
session_number: session-1752531600
technical_tags: [storage, architecture, git-integration, conflict-resolution]
artifact_refs: []
metadata:
  status: active
  priority: critical
  estimated_effort: large
  progress_percent: 0
  artifact_count: 0
  activity_score: 30.0
  blocked_by: []
---

# Centralized Work Storage with Git Context Metadata

*Last updated: 2025-07-17 18:20*

Implement a centralized storage architecture where all work items live at the project root `.claude-work` directory, with git worktree and branch information stored as metadata. This solves conflict issues, enables true project-wide visibility, and survives worktree lifecycle changes.

## Problem Statement

Current distributed storage creates conflicts when:
- Work items exist in multiple `.claude-work` directories
- Worktrees are created for features with existing work items
- Updates happen from different locations
- Worktrees are deleted but work should persist

## Solution Design

### Centralized Architecture
```
project-root/
├── .claude-work/                    ← ALL work items live here
│   ├── work/
│   │   ├── now/
│   │   ├── next/
│   │   └── later/
│   └── artifacts/
├── worktrees/
│   ├── feature-x/                   ← No .claude-work here
│   └── tasks/                       ← No .claude-work here
```

### Work Item Structure
```yaml
---
id: work-feature-x-123
title: Implement feature X
git_context:
  current_branch: feature/x
  current_worktree: worktrees/feature-x
  created_in: main
  worked_in: [main, worktrees/feature-x, worktrees/tasks]
  last_updated_from: worktrees/tasks
  update_history:
    - timestamp: 2025-07-17T18:00:00Z
      location: worktrees/feature-x
      branch: feature/x
    - timestamp: 2025-07-17T17:00:00Z
      location: main
      branch: main
metadata:
  worktree_active: true
  branch_active: true
---
```

## Implementation Tasks

### Phase 1: Storage Migration
- [ ] Create migration tool to consolidate distributed work items
- [ ] Implement conflict detection for duplicate work items
- [ ] Design merge strategy for conflicting updates
- [ ] Build rollback mechanism for safety

### Phase 2: Context Management
- [ ] Implement automatic git context capture
- [ ] Create context validation on work item load
- [ ] Add stale context detection and cleanup
- [ ] Build context history tracking

### Phase 3: Data Access Layer
- [ ] Modify EnhancedClient to always use project root
- [ ] Implement context-aware filtering methods
- [ ] Add worktree lifecycle hooks
- [ ] Create branch-based work grouping

### Phase 4: Smart Resolution
- [ ] Implement "closest context" resolution
- [ ] Add explicit context override options
- [ ] Create context inheritance rules
- [ ] Build conflict notification system

### Phase 5: UI Integration
- [ ] Update TUI to show context information
- [ ] Add filtering by branch/worktree
- [ ] Implement context switching shortcuts
- [ ] Create visual context indicators

### Phase 6: Performance & Optimization
- [ ] Build context index for fast lookups
- [ ] Implement lazy loading for large projects
- [ ] Add context caching layer
- [ ] Optimize git operations

## Key Benefits

1. **Single Source of Truth**: All work items in one location
2. **No Conflicts**: Updates always go to same file
3. **Survives Deletion**: Work persists when worktrees removed
4. **Complete Visibility**: True project-wide view
5. **Git-Aware**: Automatic branch/worktree tracking
6. **Historical Context**: Full update trail

## Migration Strategy

### Step 1: Discovery
```go
func discoverAllWorkItems() map[string][]*Work {
    // Scan all .claude-work directories
    // Group by ID for conflict detection
    // Return map of ID → []locations
}
```

### Step 2: Conflict Resolution
```go
func resolveConflicts(items map[string][]*Work) *Work {
    // Use latest updated_at
    // Merge git_context.worked_in
    // Combine update histories
    // Return consolidated item
}
```

### Step 3: Centralization
```go
func centralizeWorkItems() error {
    // Write to project root
    // Update git context
    // Remove distributed copies
    // Create backup
}
```

## Context-Aware Operations

### Creating Work
```go
func CreateWork(title, description string) (*Work, error) {
    work := &Work{
        ID: generateID(),
        GitContext: captureCurrentContext(),
    }
    // Always save to project root
    return saveToProjectRoot(work)
}
```

### Updating Work
```go
func UpdateWork(id string, updates map[string]interface{}) error {
    work := loadFromProjectRoot(id)
    work.GitContext.LastUpdatedFrom = getCurrentLocation()
    work.GitContext.UpdateHistory = append(history, newEntry)
    return saveToProjectRoot(work)
}
```

### Filtering Work
```go
func GetWorkByContext(branch, worktree string) ([]*Work, error) {
    allWork := loadAllFromProjectRoot()
    return filterByContext(allWork, branch, worktree)
}
```

## Success Criteria

1. **All work items** stored at project root only
2. **No duplicate** work items across directories
3. **Context preserved** through git metadata
4. **Performance maintained** with large projects
5. **Backward compatible** with existing work items
6. **Seamless UX** - users unaware of centralization

## Future Enhancements

1. **Distributed Sync**: Optional sync between team members
2. **Branch Templates**: Auto-create work when branching
3. **Worktree Templates**: Standard work for worktree types
4. **Context Intelligence**: Smart work suggestions by context
5. **Git Hook Integration**: Auto-update context on git operations

This centralized approach with rich context metadata provides the robustness and flexibility needed for complex multi-worktree development workflows while maintaining simplicity and preventing conflicts.