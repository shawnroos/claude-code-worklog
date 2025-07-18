---
id: work-artifact-interface-1752484500
title: Add DOCS tab with artifact browsing interface
description: Create a DOCS tab positioned at the rightmost side of the terminal with artifact browsing, filtering, and association management
schedule: now
created_at: 2025-07-14T11:12:00Z
updated_at: 2025-07-17T17:15:00Z
overview_updated: 2025-07-17T17:15:00Z
updates_ref: updates/work-artifact-interface-1752484500.md
git_context:
  branch: main
  worktree: claude-work-tracker-ui
  working_directory: /Users/shawnroos/claude-work-tracker/worktrees/tasks/claude-work-tracker-ui
session_number: session-1752484500
technical_tags: [tui, artifacts, associations, groups]
artifact_refs: ["proposal-mcp-server-2025-07-12-mcp001", "proposal-work-command-2025-07-12-cmd001"]
metadata:
  status: active
  priority: medium
  estimated_effort: medium
  progress_percent: 0
  artifact_count: 2
  activity_score: 8.0
---

# Artifact Interface Implementation

*Last updated: 2025-07-17 17:15*

This work item implements a new DOCS tab in the TUI for browsing and managing artifacts. The DOCS tab will be positioned at the rightmost side of the terminal window, providing a clean interface for artifact discovery, filtering, and association management.

## Design Specifications

### Tab Layout
- **Position**: NOW | NEXT | LATER | DOCS (rightmost, always visible)
- **Navigation**: Tab/Shift+Tab includes DOCS in rotation
- **Styling**: Consistent with existing tabs

### DOCS Tab Structure
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

## Tasks

### Phase 1: DOCS Tab Implementation
- [ ] Add DOCS tab to TabbedWorkView (rightmost position)
- [ ] Implement tab styling and navigation
- [ ] Create ArtifactBrowserView component
- [ ] Wire keyboard shortcuts (Tab to reach DOCS)

### Phase 2: Layout Components
- [ ] Design horizontal filter bar with type checkboxes
- [ ] Add search input field to filter bar
- [ ] Create two-column master-detail layout
- [ ] Implement responsive column widths

### Phase 3: Artifact Data Layer
- [ ] Create Artifact model structure
- [ ] Implement artifact data client
- [ ] Add artifact loading from filesystem
- [ ] Support filtering and search queries

### Phase 4: Master-Detail Interface
- [ ] Left column: Scrollable artifact list
- [ ] Type indicators and metadata display
- [ ] Right column: Full artifact preview
- [ ] Keyboard navigation between columns (←/→)

### Phase 5: Association Features
- [ ] Display linked Work items in detail view
- [ ] Add/remove artifact associations
- [ ] Visual indicators for associated artifacts
- [ ] Quick association from Work item view

## Key Features

- **DOCS Tab**: Dedicated tab for artifact browsing, always rightmost
- **Filter Bar**: Horizontal layout with type filters and search
- **Master-Detail View**: Two-column layout for list and preview
- **Keyboard Navigation**: Full keyboard support with intuitive shortcuts
- **Association Management**: Link artifacts to Work items
- **Search & Filter**: Fast filtering by type, tags, or content

## Supporting Artifacts

### Proposals
![[proposal-mcp-server-2025-07-12-mcp001.md]]
![[proposal-work-command-2025-07-12-cmd001.md]]

## Updates

![[updates/work-artifact-interface-1752484500.md]]