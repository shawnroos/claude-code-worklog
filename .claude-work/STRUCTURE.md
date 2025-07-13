# Hierarchical Work + Artifacts Directory Structure

This directory implements the new two-tier architecture where **Work** items are schedulable containers and **Artifacts** are supporting documents.

## Directory Layout

```
.claude-work/
├── work/                   # Scheduled Work containers (actionable items)
│   ├── now/               # Currently active work
│   ├── next/              # Ready to start soon
│   └── later/             # Future work
├── artifacts/             # Supporting documents (not directly scheduled)
│   ├── plans/            # Strategic planning documents
│   ├── proposals/        # Feature suggestions and ideas
│   ├── analysis/         # Research and investigation reports
│   ├── updates/          # Progress updates and status reports
│   └── decisions/        # Decision records and architectural choices
├── groups/               # Explicit groupings of related artifacts
└── archive/              # Completed or stale items
    ├── work/            # Archived work containers
    ├── artifacts/       # Stale artifacts
    └── groups/          # Old groupings
```

## File Counts (Post-Migration)

- **Work Items**: 1 (scheduled containers)
- **Artifacts**: 7 (supporting documents)
  - Plans: 4
  - Proposals: 1  
  - Analysis: 1
  - Decisions: 1
  - Updates: 0
- **Groups**: 0 (explicit groupings)

## Association System

### 1. Tags (Incidental Relationships)
```yaml
technical_tags: [architecture, data-models, work-tracking]
```

### 2. References (Strong Relationships)
```yaml
# In Artifacts
work_refs: [work-id-1, work-id-2]
related_artifacts: [artifact-id-1, artifact-id-2]

# In Work
artifact_refs: [artifact-id-1, artifact-id-2]
```

### 3. Groups (Strongest Relationships)
```yaml
group_id: "auth-system-2025"
```

## Migration Status

✅ **Completed:**
- New directory structure created
- 7 artifacts migrated from old `items/` structure  
- 1 Work container created for current active work
- All items properly categorized by type

📋 **Current Distribution:**
- **NOW**: 1 Work item (hierarchical architecture implementation)
- **NEXT**: Empty (ready for new Work items)
- **LATER**: Empty (future Work items)

## Usage

### Creating Work Items
Work items go in `work/{schedule}/` and represent actionable containers:
```bash
# Example: work/now/work-implement-auth-2025-07-13-auth001.md
```

### Creating Artifacts  
Artifacts go in `artifacts/{type}/` and support Work items:
```bash
# Example: artifacts/plans/plan-auth-strategy-2025-07-13-auth002.md
```

### Creating Groups
Groups go in `groups/` and collect related artifacts:
```bash
# Example: groups/group-auth-system-2025-07-13-auth-group.md
```

This structure enables sophisticated relationship management while maintaining clear separation between actionable Work and supporting context.