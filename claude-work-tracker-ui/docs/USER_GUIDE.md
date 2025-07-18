# Claude Work Tracker TUI - User Guide

## Table of Contents
1. [Getting Started](#getting-started)
2. [Understanding the Interface](#understanding-the-interface)
3. [Managing Work Items](#managing-work-items)
4. [Search and Filter](#search-and-filter)
5. [Keyboard Shortcuts Reference](#keyboard-shortcuts-reference)
6. [Work Item Lifecycle](#work-item-lifecycle)
7. [Best Practices](#best-practices)
8. [Tips and Tricks](#tips-and-tricks)

## Getting Started

### First Launch
When you run `cw` for the first time, the TUI will:
1. Scan your current directory for `.claude-work` folders
2. Discover all work items across your project
3. Display them organized by schedule (NOW/NEXT/LATER/CLOSED)

### Understanding Work Organization
Work items follow a natural progression:
- **NOW** → What you're actively working on
- **NEXT** → What's ready to start next
- **LATER** → Future work and ideas
- **CLOSED** → Completed or canceled items

## Understanding the Interface

### The Tab Bar
```
● NOW (3) | ○ NEXT (5) | ⊖ LATER (2) | ✓ CLOSED (12)
```
- Icons indicate the tab type
- Numbers show item count
- Active tab is highlighted

### Work Item Display
Each item shows:
```
IN_PROGRESS  Implement user authentication              
Backend work for secure login system
progress:60% • tags: auth, security • branch:main • updated: 2h ago
```

### Search Bar
When active (press `/`):
```
🔍 Search: auth│
```
Or showing results:
```
🔍 "auth" (3/8 results) - Press / to search again
```

## Managing Work Items

### Creating Work Items
Work items are markdown files. Create them manually or through Claude:
```bash
# Create in appropriate directory
vim .claude-work/work/now/implement-feature.md
```

### Moving Between Schedules
Currently, move files manually:
```bash
# Move from NOW to NEXT
mv .claude-work/work/now/item.md .claude-work/work/next/
```

### Completing Work
1. Navigate to item in NOW tab
2. Press `c` to complete
3. Item automatically moves to CLOSED with:
   - Status: completed
   - CompletedAt timestamp
   - File moved to `closed/` directory

### Canceling Work
1. Navigate to item in NOW tab
2. Press `x` to cancel
3. Item moves to CLOSED with canceled status

## Search and Filter

### Basic Search
1. Press `/` from any tab
2. Start typing your search query
3. Results filter in real-time
4. Press `Enter` to lock search
5. Press `Esc` to clear

### Search Capabilities
- **Fuzzy matching**: "impl auth" matches "implement authentication"
- **Multi-field search**: Searches title, description, tags, and content
- **Case-insensitive**: "AUTH" matches "auth"

### Search Examples
- `auth` - Find all authentication-related items
- `bug fix` - Find bug fixes
- `todo` - Find items with TODO markers
- Tag names like `frontend`, `api`, etc.

## Keyboard Shortcuts Reference

### Navigation
| Key | Action |
|-----|--------|
| `Tab` | Next tab |
| `Shift+Tab` | Previous tab |
| `↑`/`k` | Previous item |
| `↓`/`j` | Next item |
| `Enter` | View details |
| `Esc` | Back/Cancel |
| `q` | Quit |

### Actions
| Key | Action | Context |
|-----|--------|---------|
| `c` | Complete item | NOW tab only |
| `x` | Cancel item | NOW tab only |
| `d` | Toggle detail view | List view |
| `/` | Search | Any tab |

### In Detail View
| Key | Action |
|-----|--------|
| `←` | Previous item |
| `→` | Next item |
| `Space`/`PgDn` | Scroll down |
| `PgUp` | Scroll up |
| `Esc` | Back to list |

## Work Item Lifecycle

### 1. Creation
- Work items start in NOW, NEXT, or LATER
- Default status: `active` or `draft`
- Progress: 0%

### 2. Active Development
- Move to NOW when starting work
- Update status to `in_progress`
- Track progress percentage
- Add task checkboxes

### 3. Completion
- Press `c` to complete
- Automatically:
  - Sets status to `completed`
  - Adds CompletedAt timestamp
  - Moves file to `closed/`
  
### 4. Archival
- Old closed items can be archived
- Use migration script for bulk operations
- Keeps CLOSED tab manageable

## Best Practices

### 1. Keep NOW Focused
- Only items actively being worked on
- Aim for 3-5 items max
- Use progress percentages

### 2. Regular Reviews
- Move stale items from NOW to NEXT
- Review NEXT queue priority
- Archive old CLOSED items monthly

### 3. Effective Tagging
- Use consistent tag names
- Include technology tags: `react`, `api`, `database`
- Add context tags: `bug`, `feature`, `refactor`

### 4. Progress Tracking
- Update progress_percent regularly
- Use task lists in content
- Add updates when significant progress made

### 5. Search Strategies
- Use tags for categorization
- Include keywords in descriptions
- Add technical terms for better discovery

## Tips and Tricks

### 1. Quick Status Check
The tab counts give instant project overview:
- High NOW count? Time to complete or defer items
- Empty NEXT? Plan upcoming work
- Large CLOSED? Run archive migration

### 2. Bulk Operations
Use the migration script for cleanup:
```bash
# See everything at once
./scripts/migrate-work-items.sh status

# Clean up in bulk
./scripts/migrate-work-items.sh all
```

### 3. Multi-Worktree Workflow
- Work items stay with their branches
- See all work across worktrees
- Git context shows item source

### 4. Search Shortcuts
- Single letters often work: `a` for "auth"
- Combine terms: `bug auth` for auth bugs
- Use partial words: `impl` for "implement"

### 5. Terminal Setup
- Use a terminal with good Unicode support
- Recommended minimum: 80x24 characters
- Larger terminals show more content

### 6. Performance Tips
- Archive old items to keep lists manageable
- Use search to filter large lists
- Close unnecessary worktrees

## Common Workflows

### Daily Standup
1. Open `cw`
2. Review NOW tab for current work
3. Check NEXT for upcoming items
4. Complete finished items with `c`

### Sprint Planning
1. Review NEXT and LATER tabs
2. Move priority items to NEXT
3. Estimate and tag appropriately
4. Check CLOSED for recent completions

### End of Day
1. Update progress on NOW items
2. Complete any finished work
3. Move blocked items to NEXT
4. Quick search for tomorrow's work

### Code Review Prep
1. Search for feature name
2. Review related work items
3. Check completion status
4. Note any follow-up tasks

## Troubleshooting

### Items Not Appearing
- Check `.claude-work` directory exists
- Verify markdown has valid frontmatter
- Ensure correct file permissions

### Search Not Finding Items
- Try simpler search terms
- Check item actually contains term
- Verify fuzzy match expectations

### Completed Items Not Moving
- Check write permissions on directories
- Ensure `closed/` directory exists
- Look for error messages

### Performance Issues
- Archive old closed items
- Reduce number of worktrees
- Check for filesystem issues