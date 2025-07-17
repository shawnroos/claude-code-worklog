---
id: work-project-wide-scanning-1752529800
title: Implement comprehensive project-wide work and artifact scanning
description: Enable cw to recursively scan all .claude-work directories in a project and aggregate both work items and artifacts for a complete project view
schedule: now
created_at: 2025-07-17T17:50:00Z
updated_at: 2025-07-17T17:50:00Z
overview_updated: 2025-07-17T17:50:00Z
updates_ref: updates/work-project-wide-scanning-1752529800.md
git_context:
  branch: feature/tasks
  worktree: claude-work-tracker-ui
  working_directory: /Users/shawnroos/claude-work-tracker/worktrees/tasks/claude-work-tracker-ui
session_number: session-1752529800
technical_tags: [scanning, aggregation, project-wide, filesystem]
artifact_refs: []
metadata:
  status: in_progress
  priority: high
  estimated_effort: medium
  progress_percent: 40
  artifact_count: 0
  activity_score: 20.0
---

# Project-Wide Work and Artifact Scanning

*Last updated: 2025-07-17 17:50*

Implement comprehensive recursive scanning to make the `cw` alias truly project-wide. When run from any directory, it should find all `.claude-work` directories in the project and aggregate both work items and artifacts for a complete view.

## Current Limitations

The existing system has the infrastructure but doesn't fully utilize it:
- ✅ `ProjectScanner.GetAllWorkDirectories()` finds all `.claude-work` directories
- ✅ Single-directory work and artifact loading works
- ❌ **Missing:** Multi-directory aggregation in client methods
- ❌ **Missing:** `GetWorkBySchedule()` method that TabbedWorkView expects

## Design Specifications

### Project-Wide Aggregation
```
Project Root
├── .claude-work/              ← Main project work
│   ├── work/now/
│   ├── work/next/
│   ├── work/later/
│   └── artifacts/
├── feature-a/
│   └── .claude-work/          ← Feature-specific work
├── worktrees/tasks/
│   └── .claude-work/          ← Worktree-specific work
└── docs/
    └── .claude-work/          ← Documentation work
```

### Aggregated View
- **Work Items**: ALL work from ALL directories by schedule
- **Artifacts**: ALL artifacts from ALL directories by type
- **Context Preservation**: Show which directory each item comes from

## Tasks

### Phase 1: Core Infrastructure
- [ ] Implement `GetWorkBySchedule(schedule string) ([]*models.Work, error)`
- [ ] Implement `GetAllArtifactsProjectWide() ([]*models.Artifact, error)`
- [ ] Update `GetArtifactsByType()` to scan all directories
- [ ] Add directory source tracking to results

### Phase 2: Enhanced Client Methods
- [ ] Create `GetWorkByScheduleProjectWide()` wrapper method
- [ ] Update existing artifact methods to use multi-directory scanning
- [ ] Add conflict detection for duplicate IDs across directories
- [ ] Implement priority/precedence rules for conflicts

### Phase 3: Directory Context
- [ ] Add source directory metadata to Work and Artifact models
- [ ] Create directory-aware display formatting
- [ ] Add filtering/grouping by source directory
- [ ] Implement directory-relative path display

### Phase 4: Performance Optimization
- [ ] Add caching layer for directory scanning results
- [ ] Implement lazy loading for large projects
- [ ] Add file watching for real-time updates
- [ ] Optimize filesystem access patterns

### Phase 5: Integration Testing
- [ ] Test with multiple worktrees active
- [ ] Verify performance with large projects
- [ ] Test conflict resolution scenarios
- [ ] Validate real-time sync across directories

## Implementation Strategy

### Multi-Directory Work Loading
```go
func (c *EnhancedClient) GetWorkBySchedule(schedule string) ([]*models.Work, error) {
    workDirs := c.scanner.GetAllWorkDirectories()
    var allWork []*models.Work
    
    for _, workDirInfo := range workDirs {
        workDir := workDirInfo.Path
        scheduleDir := filepath.Join(workDir, "work", schedule)
        
        if items := loadWorkFromDirectory(scheduleDir); len(items) > 0 {
            // Add source context
            for _, item := range items {
                item.SourceDirectory = workDirInfo.RelativePath
                item.SourcePath = workDir
            }
            allWork = append(allWork, items...)
        }
    }
    
    return sortWorkByPriority(allWork), nil
}
```

### Multi-Directory Artifact Loading
```go
func (c *EnhancedClient) GetAllArtifactsProjectWide() ([]*models.Artifact, error) {
    workDirs := c.scanner.GetAllWorkDirectories()
    var allArtifacts []*models.Artifact
    
    for _, workDirInfo := range workDirs {
        artifactsDir := filepath.Join(workDirInfo.Path, "artifacts")
        
        if artifacts := loadArtifactsFromDirectory(artifactsDir); len(artifacts) > 0 {
            // Add source context
            for _, artifact := range artifacts {
                artifact.SourceDirectory = workDirInfo.RelativePath
                artifact.SourcePath = workDirInfo.Path
            }
            allArtifacts = append(allArtifacts, artifacts...)
        }
    }
    
    return sortArtifactsByRelevance(allArtifacts), nil
}
```

## Key Benefits

- **True Project-Wide View**: See all work and artifacts regardless of current directory
- **Worktree Awareness**: Includes work from all git worktrees
- **Context Preservation**: Know which directory/feature each item belongs to
- **Unified Interface**: Single command shows complete project state
- **Performance**: Leverages existing ProjectScanner infrastructure

## Success Criteria

1. `cw` shows work items from all directories when run from any location
2. DOCS tab displays artifacts from entire project
3. Performance remains acceptable with multiple directories
4. Source directory context is clearly displayed
5. No duplicate items or ID conflicts

This implementation will make `cw` the definitive project-wide work tracking interface, providing complete visibility into all work and artifacts across the entire project structure.