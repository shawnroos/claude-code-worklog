---
id: work-hierarchical-architecture-2025-07-13-tasks001
title: Implement Hierarchical Work + Artifacts Architecture
description: Transform the current flat work item system into a sophisticated two-tier architecture where Work items become primary schedulable containers and current work items become supporting Artifacts
schedule: now
created_at: 2025-07-13T01:00:00Z
updated_at: 2025-07-13T01:15:00Z
started_at: 2025-07-13T01:15:00Z
git_context:
  branch: feature/tasks
  worktree: tasks
  working_directory: /Users/shawnroos/claude-work-tracker/worktrees/tasks
session_number: session-tasks-architecture
technical_tags: [architecture, data-models, work-tracking, hierarchy]
artifact_refs: [plan-hierarchical-work-architecture-2025-07-13-tasks001]
group_id: ""
metadata:
  status: in_progress
  priority: high
  estimated_effort: large
  progress_percent: 25
  milestones: 
    - "Phase 1: Core Data Models & Structure"
    - "Phase 2: Enhanced Storage & Data Layer"
    - "Phase 3: Intelligent Consolidation System"
    - "Phase 4: Updated UI & Workflow"
  completed_tasks:
    - "Create Work model as primary schedulable container"
    - "Create Artifact model for supporting documents"
    - "Create Group model for explicit groupings"
    - "Create new directory structure"
  pending_tasks:
    - "Modify MarkdownIO to handle Work vs Artifact distinction"
    - "Add association tracking and reference resolution"
    - "Update EnhancedClient with new hierarchy support"
    - "Modify TUI to show Work in NOW/NEXT/LATER"
  blocked_by: []
  blocks: []
  dependencies: []
  success_criteria:
    - "Clear separation between schedulable Work and supporting Artifacts"
    - "Intuitive association system with multiple relationship types"
    - "Intelligent consolidation suggestions"
    - "Lifecycle management with decay logic"
    - "Migration path from current flat structure"
  acceptance_criteria:
    - "Work items can be created from multiple artifacts"
    - "TUI shows Work in scheduling tabs, not artifacts"
    - "Association system supports tags, references, and groups"
    - "Consolidation script can create Work from grouped artifacts"
  delivery_targets:
    - "New data models (Work, Artifact, Group)"
    - "Updated storage layer with new directory structure"
    - "Enhanced consolidation tooling"
    - "Updated TUI interface"
  review_required: true
  reviewed_by: []
  quality_checks: 
    - "Model compilation and type safety"
    - "Directory structure validation"
  artifact_count: 1
  last_artifact_added: 2025-07-13T01:15:00Z
  activity_score: 8.5
  last_activity_at: 2025-07-13T01:15:00Z
  decay_warning: false
---

# Implement Hierarchical Work + Artifacts Architecture

## Current Progress

This Work item represents the implementation of a new hierarchical architecture for work tracking. The transformation moves from a flat system where work items are directly scheduled to a sophisticated two-tier system.

### Phase 1: Core Data Models & Structure ✅ 25% Complete

**Completed:**
- [x] Created Work model as primary schedulable container
- [x] Created Artifact model for supporting documents  
- [x] Created Group model for explicit groupings
- [x] Created new directory structure (work/, artifacts/, groups/)

**In Progress:**
- [ ] Modify MarkdownIO to handle Work vs Artifact distinction
- [ ] Add association tracking and reference resolution

### Phase 2: Enhanced Storage & Data Layer (Upcoming)

- [ ] Implement group management and lifecycle tracking
- [ ] Update EnhancedClient with new hierarchy support

### Phase 3: Intelligent Consolidation System (Planned)

- [ ] Enhance existing consolidation script for Work creation from artifacts
- [ ] Add intelligent grouping suggestions based on content similarity
- [ ] Implement decay logic for orphaned artifacts and unsupported Work items
- [ ] Create artifact clustering and relationship detection

### Phase 4: Updated UI & Workflow (Planned)

- [ ] Modify TUI to show Work in NOW/NEXT/LATER instead of artifacts
- [ ] Add artifact browsing, association, and grouping interface
- [ ] Implement Work creation workflow from multiple artifacts
- [ ] Add stale item management and cleanup suggestions

## Architecture Overview

### The New Hierarchy

```
Work (scheduled in NOW/NEXT/LATER)
├── Supporting Artifacts (unscheduled items)
│   ├── 📋 Plans
│   ├── 💡 Proposals  
│   ├── 🔍 Analysis
│   ├── 📝 Updates
│   └── ⚖️ Decisions
└── Consolidated Tasks/Information (bubbled up from artifacts)
```

### Directory Structure

```
.claude-work/
├── work/           # Scheduled Work containers
│   ├── now/        # Active work (this item is here)
│   ├── next/       # Ready to start soon
│   └── later/      # Future work
├── artifacts/      # Supporting work items
│   ├── plans/
│   ├── proposals/
│   ├── analysis/
│   ├── updates/
│   └── decisions/
├── groups/         # Explicit groupings
└── archive/        # Completed/stale items
    ├── work/
    ├── artifacts/
    └── groups/
```

### Association System (3-Tier)

1. **Tags** (incidental relationships) - Lightweight connections via shared topics
2. **References** (stronger relationships) - Explicit links between related items  
3. **Explicit Grouping** (strongest relationships) - Formal groupings that spawn Work

## Supporting Artifacts

This Work item is supported by:
- **plan-hierarchical-work-architecture-2025-07-13-tasks001**: Original planning document with detailed implementation phases

## Success Metrics

- [x] Core data models created and compile successfully
- [x] New directory structure established
- [ ] Storage layer updated to handle new hierarchy
- [ ] TUI modified to show Work instead of artifacts
- [ ] Migration tooling for existing work items
- [ ] Intelligent consolidation working

## Next Actions

1. **Update MarkdownIO** to handle both Work and Artifact storage/retrieval
2. **Add association tracking** to manage references between Work and Artifacts
3. **Update EnhancedClient** to support the new hierarchy
4. **Begin TUI modifications** to display Work items in scheduling tabs

This represents a fundamental architectural improvement that will make work tracking more sophisticated while maintaining the narrative depth that makes the system valuable.