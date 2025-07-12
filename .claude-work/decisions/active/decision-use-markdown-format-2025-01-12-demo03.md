---
id: decision-use-markdown-format-demo
type: decision
summary: Use markdown files with YAML frontmatter instead of JSON for work item storage
schedule: now
technical_tags: [architecture, file-format, data-storage]
session_number: session-demo-125
created_at: 2025-01-12T01:00:00Z
updated_at: 2025-01-12T01:00:00Z
git_context:
  branch: visualizer
  worktree: claude-work-tracker-ui
  working_directory: /Users/shawnroos/claude-work-tracker/claude-work-tracker-ui
metadata:
  status: active
  enforcement_active: true
  alternatives_considered: [json-files, yaml-files, database, plain-text]
  review_date: 2025-04-12T00:00:00Z
---

# Decision: Use Markdown + YAML for Work Item Storage

## Decision Made

**We will use markdown files with YAML frontmatter for storing work items instead of JSON files.**

This decision applies to all new work item storage and the system will maintain backward compatibility with existing JSON files during the transition period.

## Context

The original work tracking system used JSON files for storing work items, which worked for basic metadata but created problems when trying to preserve rich, narrative content that included:

- Multi-paragraph descriptions
- Formatted text with headers, lists, code blocks
- Complex planning documents with multiple sections
- Detailed reasoning and decision rationale

## Decision Factors

### Primary Driver: Rich Content Preservation
The core goal of this system is to preserve the full narrative depth of project thinking across context window resets. JSON format created several problems:

```json
{
  "content": "# Plan\n\n## Phase 1\nImplement auth...\n\n### Technical Details\nWe need to..."
}
```

This made content unreadable and difficult to edit directly.

### Human Readability
Markdown files are:
- Directly readable in any text editor
- Easy to edit without special tools
- Naturally formatted for documentation
- Compatible with existing developer workflows

### Version Control Benefits
- Clean git diffs showing actual content changes
- Easy to review in pull requests
- Natural merging of content changes
- Better blame/history tracking

### Tool Ecosystem
- Many tools already parse YAML frontmatter (Jekyll, Hugo, Obsidian)
- Standard parsing libraries available
- Can leverage existing markdown processors
- Future integration with documentation systems

## Alternatives Considered

### Pure JSON Files ❌
**Pros**: Easy programmatic access, structured data
**Cons**: Poor readability for long content, escaped formatting, hard to edit

### Pure YAML Files ❌  
**Pros**: Human readable, good for metadata
**Cons**: Poor support for long-form content, formatting limitations

### Database Storage ❌
**Pros**: Query capabilities, ACID transactions
**Cons**: Deployment complexity, not portable, overkill for local tool

### Plain Text Files ❌
**Pros**: Maximum simplicity, universally readable
**Cons**: No structured metadata, hard to parse programmatically

## Implementation Requirements

### File Structure
```markdown
---
id: work-item-id
type: plan
summary: Brief description
# ... other metadata
---

# Full Content

Complete narrative with rich formatting...
```

### Backward Compatibility
- Enhanced client reads both markdown and JSON formats
- Existing JSON files continue to work
- Gradual migration path available
- No disruption to current workflows

### Performance Considerations
- YAML frontmatter parsing is fast
- File system access patterns remain the same
- Caching strategies still applicable
- Search performance maintained

## Enforcement Guidelines

### New Work Items
- **MUST** be created in markdown format
- **MUST** include proper YAML frontmatter
- **MUST** follow filename conventions

### Existing Work Items  
- **MAY** continue to use JSON format
- **SHOULD** be migrated opportunistically
- **MUST NOT** be broken by new system

### Tool Development
- **MUST** support both formats during transition
- **SHOULD** prefer markdown for new features
- **MUST** provide migration utilities

## Review Schedule

This decision will be reviewed on **April 12, 2025** to:
- Assess migration progress
- Evaluate developer experience
- Consider any discovered limitations
- Decide on timeline for JSON deprecation

## Success Metrics

After 3 months of implementation:
- ✅ 90% of new work items use markdown format
- ✅ Developer feedback on readability is positive
- ✅ No significant performance degradation
- ✅ Git workflow integration is smooth
- ✅ Tool ecosystem benefits are realized

## Related Decisions

This decision enables:
- Enhanced narrative preservation capabilities
- Better integration with documentation workflows  
- Improved developer experience for direct file editing
- Future integration with static site generators

## Rollback Plan

If significant issues are discovered:
1. Pause creation of new markdown files
2. Create migration tool back to JSON
3. Maintain dual format support indefinitely
4. Evaluate alternative approaches

The markdown format provides a strong foundation for preserving the rich context that makes this work tracking system valuable while maintaining the structured data access needed for tooling.