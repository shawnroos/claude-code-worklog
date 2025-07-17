---
id: work-closed-tab-1752533400
title: Implement CLOSED tab for completed and canceled work items
description: Add a CLOSED tab to the TUI for better work state management and clearer NOW/NEXT/LATER semantics
schedule: now
created_at: 2025-07-17T18:50:00Z
updated_at: 2025-07-17T18:50:00Z
overview_updated: 2025-07-17T18:50:00Z
updates_ref: updates/work-closed-tab-1752533400.md
git_context:
  branch: feature/tasks
  worktree: worktrees/tasks
  working_directory: /Users/shawnroos/claude-work-tracker/worktrees/tasks/claude-work-tracker-ui
session_number: session-1752533400
technical_tags: [ui, workflow, state-management, tabs]
artifact_refs: []
metadata:
  status: in_progress
  priority: high
  estimated_effort: medium
  progress_percent: 60
  artifact_count: 0
  activity_score: 25.0
---

# CLOSED Tab Implementation for Better Work State Management

*Last updated: 2025-07-17 18:50*

Implement a CLOSED tab in the TUI to properly separate completed/canceled work from active work. This clarifies the semantic meaning of each tab and prevents NOW from becoming cluttered with non-active items.

## Problem Statement

Currently:
- **NOW tab** contains both in-progress AND completed items
- No clear place for completed or canceled work
- Difficult to see what's actually being worked on
- NOW loses its meaning as "currently in progress"

## Solution Design

### New Tab Structure
```
┌─NOW─┬─NEXT─┬─LATER─┬─CLOSED─┬────────DOCS─┐
│                                           │
│ Only items with status:                   │
│ - IN_PROGRESS                            │
│ - ACTIVE + progress > 0                  │
└───────────────────────────────────────────┘
```

### Tab Semantics
- **NOW**: Actively being worked on (in_progress, active with progress)
- **NEXT**: Ready to start, prioritized queue
- **LATER**: Future work, ideas, backlog
- **CLOSED**: Completed, archived, canceled
- **DOCS**: Artifact browser (future)

### Work Item States → Tab Mapping
```yaml
status: in_progress     → NOW
status: active          → NOW (if progress > 0) else NEXT
status: pending         → NEXT
status: planned         → LATER
status: completed       → CLOSED
status: archived        → CLOSED
status: canceled        → CLOSED
```

## Implementation Tasks

### Phase 1: Add CLOSED Tab
- [ ] Add CLOSED to tab enum and navigation
- [ ] Update keyboard shortcuts (Tab cycles through 4 tabs)
- [ ] Style CLOSED tab consistently
- [ ] Update help text

### Phase 2: Update Schedule Logic
- [ ] Modify GetWorkBySchedule to include "closed" schedule
- [ ] Update work directory structure to include `work/closed/`
- [ ] Implement status-to-schedule mapping logic
- [ ] Add automatic schedule assignment based on status

### Phase 3: Migration Tools
- [ ] Create script to move completed items to closed/
- [ ] Update existing work items with proper schedules
- [ ] Implement auto-migration on status change
- [ ] Add validation to prevent invalid combinations

### Phase 4: Enhanced Filtering
- [ ] Filter NOW to only show in-progress items
- [ ] Add status badges in CLOSED tab (✅ completed, ❌ canceled)
- [ ] Sort CLOSED by completion date (newest first)
- [ ] Add item count to each tab header

### Phase 5: Workflow Improvements
- [ ] Quick action to move items between tabs
- [ ] Keyboard shortcut to complete/cancel current item
- [ ] Bulk operations for cleanup
- [ ] Auto-archive old closed items

## Directory Structure
```
.claude-work/
├── work/
│   ├── now/       # Only in-progress work
│   ├── next/      # Ready to start
│   ├── later/     # Future/backlog
│   └── closed/    # Completed/canceled
```

## UI Mockup
```
┌─NOW (2)─┬─NEXT (5)─┬─LATER (3)─┬─CLOSED (12)─┐
│                                                │
│ IN_PROGRESS  Implement CLOSED tab              │
│ IN_PROGRESS  Fix authentication bug            │
│                                                │
│ [Only active work items shown]                 │
└────────────────────────────────────────────────┘

┌─NOW─┬─NEXT─┬─LATER─┬─CLOSED (12)──────────────┐
│                                                │
│ ✅ COMPLETED  Project-wide scanning            │
│ ✅ COMPLETED  TUI refactoring                  │
│ ❌ CANCELED   Old feature proposal             │
│ ✅ COMPLETED  Bug fix for issue #123           │
│                                                │
│ [Completed and canceled items]                 │
└────────────────────────────────────────────────┘
```

## Migration Strategy

### Automatic Migration Rules
1. Scan all NOW items
2. If status = completed/archived/canceled → move to closed/
3. If status = active AND progress = 0 → move to next/
4. If status = planned/idea → move to later/

### Manual Cleanup Commands
```bash
# Move completed items to closed
claude-work-migrate --completed-to-closed

# Move inactive items from NOW to NEXT
claude-work-migrate --inactive-to-next

# Archive old closed items (>30 days)
claude-work-migrate --archive-old-closed
```

## Benefits

1. **Clear Focus**: NOW shows only what's being worked on
2. **Better Organization**: Completed work has its own space
3. **Improved Workflow**: Natural progression through states
4. **Historical View**: Easy to see what's been accomplished
5. **Cleaner Interface**: Each tab has clear purpose

## Success Criteria

1. NOW tab shows only in-progress items
2. CLOSED tab properly displays completed/canceled work
3. Easy navigation between all 4 tabs
4. Automatic migration of existing items
5. Clear visual distinction between states

This implementation will significantly improve work tracking clarity and make the NOW tab truly represent current work in progress.