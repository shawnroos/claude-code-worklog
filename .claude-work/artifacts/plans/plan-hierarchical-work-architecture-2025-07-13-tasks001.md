---
id: plan-hierarchical-work-architecture-2025-07-13-tasks001
type: plan
summary: Implement hierarchical Work + Artifacts architecture for improved work tracking organization
schedule: now
technical_tags: [architecture, data-models, work-tracking, hierarchy]
session_number: session-tasks-architecture
created_at: 2025-07-13T01:00:00Z
updated_at: 2025-07-13T01:00:00Z
git_context:
  branch: feature/tasks
  worktree: tasks
  working_directory: /Users/shawnroos/claude-work-tracker/worktrees/tasks
metadata:
  status: active
  implementation_status: planning
  phases: [models, storage, consolidation, ui]
  estimated_effort: high
---

# Implement Hierarchical Work + Artifacts Architecture

## Overview

Transform the current flat work item system into a sophisticated two-tier architecture where:
- **Work** items become the primary schedulable containers in NOW/NEXT/LATER
- **Artifacts** (current work items) become supporting documents that inform Work
- Multiple artifacts can feed into a single Work item for better context aggregation

## Current vs. Proposed Architecture

### Current State
- Work Items (Plan, Proposal, Analysis, Update, Decision) live directly in NOW/NEXT/LATER
- Each item is treated as standalone work to be scheduled and executed

### Proposed State
- **Work** becomes the primary container that gets scheduled in NOW/NEXT/LATER
- **Artifacts** become supporting documents that exist in a separate space
- Artifacts feed into and inform Work objects but don't directly schedule themselves

## The New Hierarchy

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

## Directory Structure

```
.claude-work/
├── work/           # Scheduled Work containers
│   ├── now/
│   ├── next/
│   └── later/
├── artifacts/      # Supporting work items
│   ├── plans/
│   ├── proposals/
│   ├── analysis/
│   ├── updates/
│   └── decisions/
└── groups/         # Explicit groupings
```

## Association System (3-Tier)

1. **Tags** (incidental relationships)
   - `technical_tags: [auth, security, backend]`
   - Lightweight connections based on shared topics

2. **References** (stronger relationships)
   - `related_items: [artifact-id-1, artifact-id-2]`
   - Explicit links between related artifacts

3. **Explicit Grouping** (strongest relationships)
   - `group_id: auth-system-2025`
   - Formal groupings that can spawn Work items

## Implementation Phases

### Phase 1: Core Data Models & Structure
- [x] Plan captured and todos created
- [ ] Create `Work` model as primary schedulable container
- [ ] Restructure existing work items as artifacts with association metadata
- [ ] Implement 3-tier association system (tags, references, explicit groups)
- [ ] Update directory structure: `work/` for scheduled items, `artifacts/` for supporting docs

### Phase 2: Enhanced Storage & Data Layer
- [ ] Modify `MarkdownIO` to handle Work vs Artifact distinction
- [ ] Add association tracking and reference resolution
- [ ] Implement group management and lifecycle tracking
- [ ] Update `EnhancedClient` with new hierarchy support

### Phase 3: Intelligent Consolidation System
- [ ] Enhance existing consolidation script for Work creation from artifacts
- [ ] Add intelligent grouping suggestions based on content similarity and references
- [ ] Implement decay logic for orphaned artifacts and unsupported Work items
- [ ] Create artifact clustering and relationship detection

### Phase 4: Updated UI & Workflow
- [ ] Modify TUI to show Work in NOW/NEXT/LATER (not artifacts)
- [ ] Add artifact browsing, association, and grouping interface
- [ ] Implement Work creation workflow from multiple artifacts
- [ ] Add stale item management and cleanup suggestions

## Workflow Benefits

1. **Context Aggregation**: Multiple related artifacts can inform a single Work object
2. **Loose Coupling**: Artifacts can exist without immediate commitment to execution
3. **Natural Grouping**: Similar items can be discovered and clustered before becoming Work
4. **Flexible Association**: A proposal might spawn analysis, which leads to plans, all feeding one Work
5. **Cleaner Scheduling**: Only actionable Work gets scheduled, not every supporting document

## Creation Workflows

### Manual
- Direct creation of Work from selected artifacts
- User explicitly chooses which artifacts to consolidate

### Script-Based (Middle Ground)
- Consolidation tool suggests groupings and creates Work
- Semi-automated with human review and approval

### Intelligent
- Auto-suggest related artifacts when creating Work
- Machine learning-based clustering and relationship detection

## Lifecycle Management with Decay Logic

### Orphaned Artifacts
- No references/groups → mark stale after X days
- Decay based on activity and relevance

### Unsupported Work
- No backing artifacts → flag for review
- Work items that never get supporting documentation

### Activity Tracking
- Last modified timestamps
- Reference count and usage metrics
- Group membership and relationship strength

## Success Criteria

- [ ] Clear separation between schedulable Work and supporting Artifacts
- [ ] Intuitive association system that supports multiple relationship types
- [ ] Intelligent consolidation suggestions that reduce manual effort
- [ ] Lifecycle management that keeps the system clean and relevant
- [ ] UI that makes the hierarchy clear and actionable
- [ ] Migration path from current flat structure to hierarchical system

This architecture will transform the work tracking system from a flat collection of scheduled items into a sophisticated knowledge management system that captures context while maintaining actionable scheduling.