---
id: update-created-1752529800
work_id: work-project-wide-scanning-1752529800
timestamp: 2025-07-17T17:50:00Z
title: Project-Wide Scanning Work Item Created
summary: Created comprehensive work item for implementing project-wide work and artifact scanning to make cw truly recursive across all .claude-work directories
author: Claude
session_id: session-1752529800
update_type: manual
tasks_completed: []
tasks_added:
  - "Implement GetWorkBySchedule(schedule string) method"
  - "Implement GetAllArtifactsProjectWide() method"
  - "Update GetArtifactsByType() for multi-directory scanning"
  - "Add directory source tracking to results"
progress_delta: 0
final_progress: 0
key_changes:
  - "Identified missing GetWorkBySchedule() method that TabbedWorkView expects"
  - "Designed multi-directory aggregation strategy"
  - "Planned source directory context preservation"
  - "Created 5-phase implementation plan"
---

# Project-Wide Scanning Implementation Plan

## Problem Identified

The current cw system has excellent infrastructure via `ProjectScanner.GetAllWorkDirectories()` but doesn't fully utilize it:

- ✅ **Infrastructure exists**: ProjectScanner finds all `.claude-work` directories
- ❌ **Missing integration**: Client methods only scan single directory
- ❌ **Missing method**: `GetWorkBySchedule()` that TabbedWorkView calls doesn't exist

## Solution Design

### Multi-Directory Aggregation
The system will scan ALL `.claude-work` directories found by ProjectScanner and aggregate:

1. **Work Items**: From `work/now/`, `work/next/`, `work/later/` across all directories
2. **Artifacts**: From `artifacts/*/` across all directories  
3. **Source Context**: Track which directory each item comes from

### Key Methods to Implement
- `GetWorkBySchedule(schedule string) ([]*models.Work, error)`
- `GetAllArtifactsProjectWide() ([]*models.Artifact, error)`
- Enhanced `GetArtifactsByType()` with multi-directory support

## Implementation Phases

1. **Core Infrastructure**: Basic multi-directory scanning
2. **Enhanced Client Methods**: Update existing methods  
3. **Directory Context**: Add source tracking
4. **Performance Optimization**: Caching and lazy loading
5. **Integration Testing**: Validate across worktrees

This will make `cw` show complete project state regardless of current directory, providing true project-wide visibility into all work and artifacts.