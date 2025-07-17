# TUI Testing Guide - Work Hierarchy

## Overview

This build includes the modified TUI that displays **Work items** instead of artifacts in the NOW/NEXT/LATER tabs. This represents the new hierarchical structure where:

- **Work items** are the primary schedulable containers
- **Artifacts** are supporting documents linked to Work items

## Running the Application

```bash
./claude-work-tracker
```

## Key Changes to Test

### 1. Work Item Display
- The NOW/NEXT/LATER tabs now show Work items
- Each item displays:
  - **Status badge** (IN_PROGRESS, BLOCKED, COMPLETED, or priority level)
  - **Title** (not summary)
  - **Description preview**
  - **Metadata**: progress %, artifact count, tags, git context

### 2. Navigation
- **Tab/Shift+Tab**: Switch between NOW/NEXT/LATER tabs
- **↑/↓**: Navigate items in the list
- **Enter**: View full Work item details
- **Esc**: Return to list view
- **←/→**: Navigate between items in full view
- **q**: Quit

### 3. Current Test Data

**NOW Tab (1 item)**:
- "Test TUI modification with Work items" - 50% progress

**NEXT Tab (1 item)**:
- "Add artifact browsing, association, and grouping interface" - 2 artifacts

**LATER Tab**: Empty

## Important Notes

1. **Local Directory Only**: This build reads from the local `.claude-work/` directory in this worktree, NOT from the main project directory.

2. **New Directory Structure**:
   ```
   .claude-work/
   ├── work/          # Work items (schedulable)
   │   ├── now/
   │   ├── next/
   │   └── later/
   ├── artifacts/     # Supporting documents
   │   ├── plans/
   │   ├── proposals/
   │   ├── analysis/
   │   ├── updates/
   │   └── decisions/
   └── groups/        # Artifact groupings
   ```

3. **Known Limitations**:
   - Artifact browsing interface not yet implemented
   - Association management UI pending
   - Group creation/management UI pending

## Testing Focus

Please verify:
1. Work items display correctly with all metadata
2. Navigation between tabs works smoothly
3. Full post view shows combined title, description, and content
4. Performance is acceptable with the new rendering
5. Status badges and progress indicators are clear

## Feedback Needed

- Are the status badges clear and meaningful?
- Is the progress percentage display helpful?
- Should artifact count be more prominent?
- Any UI elements that feel missing or confusing?