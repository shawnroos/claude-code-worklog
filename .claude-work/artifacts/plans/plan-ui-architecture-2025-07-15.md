---
id: plan-ui-architecture-2025-07-15
type: plan
summary: UI Architecture Plan for Hierarchical Work System
schedule: now
technical_tags: [ui, architecture, hierarchical, design]
session_number: session-ui-arch
created_at: 2025-07-15T21:36:00Z
updated_at: 2025-07-15T21:36:00Z
git_context:
  branch: main
  worktree: claude-work-tracker
  working_directory: /Users/shawnroos/claude-work-tracker
metadata:
  status: active
  implementation_status: in_progress
  estimated_impact: high
---

# UI Architecture Plan for Hierarchical Work System

## Overview

This plan outlines the architecture for the new hierarchical TUI interface that treats Work items as primary schedulable containers with supporting Artifacts.

## Architecture Components

### 1. Work Items (Primary Containers)
- **Purpose**: Schedulable work containers in NOW/NEXT/LATER
- **Display**: Title, status, priority, progress, artifact count
- **Content**: Description + main content + embedded artifacts
- **Navigation**: Tab-based interface for schedules

### 2. Artifacts (Supporting Documents)
- **Types**: Plans, proposals, analysis, updates, decisions
- **Purpose**: Supporting documentation linked to Work items
- **Display**: Auto-embedded in Work item detail view
- **Content**: Full markdown content with metadata

### 3. Associations
- **Linking**: Work items reference artifacts via `artifact_refs`
- **Resolution**: `GetWorkArtifacts()` fetches associated content
- **Display**: Automatic embedding under "Associated Artifacts" section

## Implementation Status

### Phase 1: Core TUI ✅
- [x] Work model integration
- [x] Automatic artifact rendering
- [x] Status badges and progress indicators
- [x] Tab navigation between schedules

### Phase 2: Enhanced Features 🔄
- [ ] Artifact browsing interface
- [ ] Association management UI
- [ ] Group creation and management
- [ ] Search and filtering capabilities

## Technical Details

### Data Flow
1. Load Work items by schedule (NOW/NEXT/LATER)
2. Display Work items in tabbed interface
3. On detail view, auto-fetch associated artifacts
4. Render combined content with markdown formatting

### Key Methods
- `GetWorkBySchedule()`: Load work items by schedule
- `GetWorkArtifacts()`: Fetch artifacts for work item
- `renderItemDetail()`: Display work + embedded artifacts
- `glamourRender.Render()`: Markdown formatting

## Testing Approach

- Test with various work item types and statuses
- Verify artifact embedding works correctly
- Test navigation and interaction
- Validate markdown rendering quality