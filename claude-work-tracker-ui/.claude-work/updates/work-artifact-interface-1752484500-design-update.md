---
id: update-design-1752528900
work_id: work-artifact-interface-1752484500
timestamp: 2025-07-17T17:35:00Z
title: DOCS Tab Design Specification
summary: Updated artifact interface design to use a DOCS tab positioned at the rightmost side of the terminal with horizontal filter bar and master-detail layout
author: Claude
session_id: session-1752528900
update_type: manual
tasks_completed: []
tasks_added:
  - "Add DOCS tab to TabbedWorkView (rightmost position)"
  - "Design horizontal filter bar with type checkboxes"
  - "Create two-column master-detail layout"
progress_delta: 0
final_progress: 0
key_changes:
  - "Changed from three-pane to two-column master-detail layout"
  - "Positioned DOCS as rightmost tab instead of separate view"
  - "Moved filters to horizontal bar below tab header"
  - "Specified 40/60 column width split"
---

# DOCS Tab Design Update

Based on user feedback, the artifact browser design has been updated to better integrate with the existing TUI:

## Key Design Changes

### Tab Integration
- DOCS tab added as the rightmost tab in the main navigation
- Always visible and accessible via Tab/Shift+Tab navigation
- Maintains consistent styling with NOW/NEXT/LATER tabs

### Layout Structure
```
┌─NOW─┬─NEXT─┬─LATER─┬────────────────DOCS─┐
│                                          │
│ [□ Plans] [□ Proposals] [□ Analysis]... │ ← Filter Bar
│ [Search: ________________]               │
├─────────────────┬────────────────────────┤
│ Artifact List   │ Artifact Details       │
│ (Master)        │ (Detail)               │
│ 40% width       │ 60% width              │
└─────────────────┴────────────────────────┘
```

### Improved UX
- Horizontal filter bar is more space-efficient
- Master-detail pattern is familiar and intuitive
- Rightmost positioning keeps work items primary focus
- Simpler navigation with fewer panes

## Implementation Priority
1. Add DOCS tab to existing tab structure
2. Create horizontal filter component
3. Implement master-detail view
4. Add artifact data layer
5. Enable association management

This design maintains the TUI's keyboard-driven efficiency while providing powerful artifact management capabilities.