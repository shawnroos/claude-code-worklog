---
id: update-created-1752531600
work_id: work-centralized-storage-1752531600
timestamp: 2025-07-17T18:20:00Z
title: Centralized Storage Architecture Designed
summary: Created comprehensive plan for centralizing all work items at project root with git context as metadata, solving conflict and persistence issues
author: Claude
session_id: session-1752531600
update_type: manual
tasks_completed: []
tasks_added:
  - "Create migration tool to consolidate distributed work items"
  - "Implement automatic git context capture"
  - "Modify EnhancedClient to always use project root"
  - "Build conflict detection and resolution"
  - "Update TUI to show context information"
progress_delta: 0
final_progress: 0
key_changes:
  - "Designed centralized storage architecture"
  - "Specified git context metadata structure"
  - "Created 6-phase implementation plan"
  - "Defined migration strategy for existing work"
---

# Centralized Storage Architecture Plan

## Problem Solved

The current distributed storage model creates conflicts when:
- Work items exist in both main and worktree directories
- Updates happen from different locations
- Worktrees are deleted but work should persist

## Solution Architecture

### Single Source of Truth
All work items will live in `project-root/.claude-work/` regardless of where you're working. Git context (branch, worktree, location) becomes metadata within each work item.

### Rich Context Tracking
```yaml
git_context:
  current_branch: feature/x
  current_worktree: worktrees/feature-x
  created_in: main
  worked_in: [main, worktrees/feature-x, worktrees/tasks]
  last_updated_from: worktrees/tasks
  update_history: [...]
```

## Implementation Strategy

### Phase 1: Storage Migration
- Consolidate all distributed work items
- Detect and resolve conflicts
- Create rollback mechanism

### Phase 2: Context Management  
- Automatic git context capture
- Stale context detection
- History tracking

### Phase 3: Data Access Layer
- EnhancedClient always uses project root
- Context-aware filtering
- Branch-based grouping

### Phase 4-6: Smart Resolution, UI, Performance

## Key Benefits

1. **No Conflicts**: Single location for all updates
2. **Persistence**: Survives worktree deletion
3. **Complete Visibility**: True project-wide view
4. **Git Integration**: Automatic context tracking
5. **Historical Trail**: Full update history

This architecture provides robust conflict-free work tracking across complex multi-worktree projects.