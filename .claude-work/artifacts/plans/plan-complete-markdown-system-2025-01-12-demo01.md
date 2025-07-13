---
id: plan-complete-markdown-system-demo
type: plan
summary: Complete implementation of markdown-based work tracking system with 5 types and NOW/NEXT/LATER scheduling
schedule: now
technical_tags: [data-models, ui, markdown, yaml]
session_number: session-demo-123
created_at: 2025-01-12T01:00:00Z
updated_at: 2025-01-12T01:30:00Z
git_context:
  branch: visualizer
  worktree: claude-work-tracker-ui
  working_directory: /Users/shawnroos/claude-work-tracker/claude-work-tracker-ui
metadata:
  status: active
  implementation_status: in_progress
  phases: [models, io-functions, enhanced-client, ui-integration, testing]
  estimated_effort: high
---

# Complete Markdown Work Tracking System Implementation

## Overview

We are implementing a comprehensive markdown-based work tracking system that preserves the full narrative depth of project thinking while providing efficient organization through the 5-type system (Plan, Proposal, Analysis, Update, Decision) and NOW/NEXT/LATER scheduling.

## Phase 1: Core Models ✅

### Completed
- Created `MarkdownWorkItem` model with YAML frontmatter support
- Implemented 5-type system constants and validation
- Added schedule property with NOW/NEXT/LATER values
- Created type-specific metadata structures for each work item type

### Key Features
- Rich content preservation with summary + full markdown content
- Technical tags for categorization
- Git context tracking for worktree awareness
- Extensible metadata system based on work item type

## Phase 2: IO Functions ✅

### Completed
- Implemented `MarkdownIO` for reading/writing markdown files
- Created automatic filename generation following pattern: `{type}-{description}-{date}-{id}.md`
- Added directory structure management based on schedule
- Implemented search and filtering capabilities

### Directory Structure
```
.claude-work/
  items/
    now/        # Currently active work
    next/       # Ready to start soon  
    later/      # Future enhancements
    completed/  # Archived completed work
  decisions/
    active/     # Enforced decisions
    superseded/ # Old decisions
```

## Phase 3: Enhanced Data Client ✅

### Completed
- Created `EnhancedClient` that wraps existing client
- Backward compatibility with legacy JSON format
- New methods for markdown-based operations
- Schedule and type-based filtering

### Key Capabilities
- Create work items with proper frontmatter
- Update scheduling (move between NOW/NEXT/LATER)
- Complete items and move to archive
- Search across all content including full text

## Phase 4: Enhanced UI 🔄 IN PROGRESS

### Completed
- Created `EnhancedWorkItemsModel` with dual view modes
- Toggle between Schedule view (NOW/NEXT/LATER) and Type view (Plan/Proposal/etc)
- Rich detail view showing full content and metadata
- Integrated with main app architecture

### Current Status
- Basic UI working with enhanced models
- Schedule-based filtering operational
- Type-based filtering operational
- Detail views show full markdown content

### Next Steps
1. Test with real markdown files
2. Improve content rendering for long-form text
3. Add markdown highlighting in detail view
4. Implement interactive schedule updates

## Phase 5: Shell Script Integration

### Planned
- Update `work.sh` to understand markdown format
- Add `/work schedule <item> <timing>` command
- Enhance `work-status.sh` for 5-type system
- Create consolidation tool for duplicate detection

## Technical Decisions Made

### File Format Choice: Markdown + YAML
**Decision**: Use markdown files with YAML frontmatter instead of JSON
**Reasoning**: 
- Human readable and editable
- Preserves formatting in rich content
- Better git diffs and version control
- Standard frontmatter parsing available
- Natural for narrative content

### 5-Type System
**Decision**: Simplify from 7 types to 5 focused types
**Types**: Plan, Proposal, Analysis, Update, Decision
**Reasoning**: Each type maps to specific Claude workflows and has distinct purposes

### NOW/NEXT/LATER Scheduling
**Decision**: Use 3-tier scheduling instead of complex priority systems
**Reasoning**: 
- Simple mental model
- Avoids deadline commitments
- Flexible and adaptable
- Industry-proven approach

## Success Metrics

- ✅ Rich content preservation (full narratives saved)
- ✅ Efficient organization (5 types + scheduling)
- ✅ Backward compatibility (legacy JSON still works)
- 🔄 Enhanced UI experience (in progress)
- ⏳ Cross-session continuity (shell integration needed)

## Example Usage

Once complete, agents will be able to:

```markdown
# Create a new plan
item := CreateWorkItem("plan", "Implement caching layer", fullPlanContent, "next", ["backend", "performance"])

# Move to active work
UpdateWorkItemSchedule(item.ID, "now")

# Complete when done
CompleteWorkItem(item.ID)
```

This preserves the complete narrative while providing structured organization for effective project management.