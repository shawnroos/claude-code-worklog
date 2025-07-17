---
id: update-created-1752533400
work_id: work-closed-tab-1752533400
timestamp: 2025-07-17T18:50:00Z
title: CLOSED Tab Design for Work State Management
summary: Designed CLOSED tab implementation to separate completed/canceled work from active items, clarifying NOW semantics as "in progress only"
author: Claude
session_id: session-1752533400
update_type: manual
tasks_completed: []
tasks_added:
  - "Add CLOSED to tab enum and navigation"
  - "Update work directory structure to include work/closed/"
  - "Create migration script for existing items"
  - "Implement status-to-schedule mapping logic"
progress_delta: 0
final_progress: 0
key_changes:
  - "Defined clear tab semantics"
  - "Created status to tab mapping rules"
  - "Designed migration strategy"
  - "Specified UI improvements"
---

# CLOSED Tab Design Created

## Problem Identified

User observation: "A lot of work is ending up in NOW that isn't actually being worked on"

This is a critical UX issue - NOW should mean "actively in progress" not just "important or active".

## Solution Design

### New 4-Tab Structure
- **NOW**: Only items being actively worked on (in_progress status)
- **NEXT**: Ready to start, prioritized queue
- **LATER**: Future work, ideas, backlog  
- **CLOSED**: Completed, archived, canceled work

### Automatic Tab Assignment
```yaml
status: in_progress → NOW
status: completed   → CLOSED
status: canceled    → CLOSED
status: active      → NEXT (unless progress > 0)
```

### Benefits
1. NOW becomes focused on current work only
2. Clear progression: LATER → NEXT → NOW → CLOSED
3. Historical record in CLOSED tab
4. No more clutter in active tabs

## Implementation Plan

1. Add CLOSED tab to navigation
2. Create `work/closed/` directory structure
3. Implement migration tools
4. Update filtering logic
5. Add keyboard shortcuts

This will make work tracking much clearer and more intuitive!