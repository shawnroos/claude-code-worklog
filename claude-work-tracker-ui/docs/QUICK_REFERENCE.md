# Claude Work Tracker - Quick Reference

## 🚀 Launch
```bash
cw
```

## ⌨️ Essential Shortcuts

### Navigation
- `Tab` → Next tab
- `↑`/`↓` → Navigate items  
- `Enter` → View details
- `Esc` → Go back
- `q` → Quit

### Actions (NOW tab)
- `c` → Complete item ✅
- `x` → Cancel item ❌
- `/` → Search 🔍

### In Detail View
- `←`/`→` → Previous/Next item
- `Space` → Page down
- `Esc` → Back to list

## 📂 Directory Structure
```
.claude-work/
├── work/
│   ├── now/     # Active work
│   ├── next/    # Queued items  
│   ├── later/   # Backlog
│   └── closed/  # Done/Canceled
```

## 🔍 Search Tips
- Press `/` to search
- Fuzzy matching: "wtf" → "work_tracker_file"
- Search in: title, description, tags, content
- `Enter` to confirm, `Esc` to clear

## 📊 Tab Icons
- `●` NOW - Active work
- `○` NEXT - Ready to start
- `⊖` LATER - Future work  
- `✓` CLOSED - Completed

## 🎯 Status Badges
- `IN_PROGRESS` - Being worked on
- `BLOCKED` - Waiting
- `✅ COMPLETED` - Done (green)
- `❌ CANCELED` - Stopped (red)
- `📦 ARCHIVED` - Old (gray)

## 🛠️ Migration Script
```bash
# Check distribution
./scripts/migrate-work-items.sh status

# Clean up everything  
./scripts/migrate-work-items.sh all
```

## 💡 Pro Tips
1. Keep NOW under 5 items
2. Press `c` when done → auto-moves to CLOSED
3. Search with `/` to filter long lists
4. Items sorted newest first
5. Multi-worktree aware

## 📝 Work Item Format
```yaml
---
id: work-feature-123
title: Build awesome feature
schedule: now
status: in_progress
progress_percent: 60
tags: [frontend, react]
---

# Description here

## Tasks
- [x] Part 1
- [ ] Part 2
```