# Project Work Intelligence Storage

This directory contains the local work intelligence for the claude-work-tracker project.

## Structure

### `/history/` - Complete Historical Archive
- **Purpose**: Complete chronological record of all work intelligence
- **Content**: Plans, proposals, findings, decisions, rationales
- **Access**: Not loaded into session context, searchable when needed
- **Lifecycle**: Permanent storage, organized by date and type

### `/active/` - Session-Critical Context
- **Purpose**: Curated subset of most relevant current work
- **Content**: Current priorities, recent decisions, historical references
- **Access**: Automatically loaded into Claude session context
- **Lifecycle**: Items age out after completion or time-based rules

## File Naming Conventions

### History Files
- `YYYY-MM-DD-descriptive-name.json` - Individual work items
- `YYYY-MM-DD-session-summary.json` - Session summaries
- `YYYY-MM-DD-weekly-digest.json` - Weekly digests

### Active Files
- `current-work-context.json` - Primary active context
- `priority-items.json` - High priority current items
- `recent-decisions.json` - Recent decisions with references

## Context Management Rules

### Automatic Archival
- Completed items move to history after 7 days
- Low priority items age out after 30 days
- Items accessed recently stay active longer

### Reference System
- Active context includes metadata pointing to historical items
- Format: `"reference": "history/filename.json"`
- Allows Claude to query specific historical items when needed

### Size Management
- Active context kept under 10kb to prevent context bloat
- Historical summaries created for long-running projects
- Automatic cleanup of duplicate or obsolete items

## Usage

### For Claude Sessions
- Active context automatically loaded
- Historical items accessible via MCP tools
- References provide breadcrumbs to detailed information

### For Users
- Browse history directory for complete project timeline
- Edit active context to prioritize specific items
- Query tools available via MCP server

## Tools Available

- `query_history(keyword, date_range)` - Search historical archive
- `get_historical_context(item_id)` - Retrieve specific historical item
- `summarize_period(start_date, end_date)` - Generate period summary
- `promote_to_active(history_item)` - Move historical item to active context
- `archive_active_item(item_id)` - Move active item to history