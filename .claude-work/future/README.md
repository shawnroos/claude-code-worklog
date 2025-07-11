# Future Work Management

This directory contains work items that have been deprioritized for future implementation using heuristic similarity-based grouping. This helps maintain focus on current scope while preserving valuable ideas and intelligently organizing them for future work.

## Structure

### `/items/` - Individual Deferred Items
- Individual work items awaiting grouping or recently added
- Named by unique ID: `item-abc123.json`
- Contains similarity metadata for intelligent grouping

### `/groups/` - Intelligent Similarity-Based Groups
- Coherent groups of related work items
- Named by feature/domain: `search-and-filtering.json`, `user-profile-features.json`
- Grouped by similarity rather than technical scheduling

### `suggestions.json` - Grouping Intelligence
- Automatic grouping suggestions based on content similarity
- Similarity analysis and clustering recommendations
- Configuration for intelligent grouping algorithms

## Future Work Item Schema

```json
{
  "id": "future-item-id",
  "type": "future_work",
  "original_type": "plan|proposal|todo|idea",
  "content": "Description of the future work",
  "similarity_metadata": {
    "keywords": ["authentication", "security", "user"],
    "feature_domain": "user-management",
    "technical_domain": "backend-api",
    "code_locations": ["src/auth/", "src/api/users/"],
    "strategic_theme": "user-experience"
  },
  "context": {
    "deprioritized_from": "active|history", 
    "deprioritized_date": "2025-07-11T18:30:00Z",
    "deprioritized_reason": "Out of scope for current sprint",
    "suggested_group": "authentication-features"
  },
  "grouping_status": "ungrouped|suggested|grouped",
  "priority_when_promoted": "high|medium|low"
}
```

## Work Group Schema

```json
{
  "id": "group-id",
  "name": "Authentication Features",
  "description": "User authentication and security-related features",
  "items": ["item-123", "item-456", "item-789"],
  "similarity_score": 0.85,
  "strategic_value": "high",
  "estimated_effort": "medium",
  "readiness_status": "ready|planning|blocked",
  "created_date": "2025-07-11T18:30:00Z",
  "last_updated": "2025-07-11T18:30:00Z"
}
```

## Management Commands

### MCP Tools (Simplified)
- `defer_to_future(content, reason)` - Frictionless deferral during planning
- `groom_future_work()` - Combined analyze + reorganize functionality  
- `list_future_groups()` - View current groupings and suggestions
- `create_work_group(name, items, description)` - Manual grouping
- `promote_work_group(group_name)` - Bring back coherent group

### Intelligent Grouping
- Content similarity analysis when items are added
- Automatic suggestions based on feature/technical/strategic similarity
- Manual grouping with intelligent recommendations
- Batch promotion of coherent work groups

## Benefits

1. **Frictionless Scope Management**: Easily defer work during planning without categorization overhead
2. **Intelligent Organization**: Automatic similarity-based grouping of related work
3. **Strategic Overview**: Understand the landscape of deferred work through natural clusters
4. **Batch Promotion**: Bring back coherent groups of related work together
5. **Natural Workflow**: Fits how developers actually think about and organize work

## Usage Patterns

### During Planning/Proposal Evaluation
- Quick scope decisions: "Let's do A and B, put C in future work"
- Zero friction deferral without breaking flow
- Natural language reasoning preserved

### During Grooming/Review Sessions  
- Periodic review of accumulated future work
- Intelligent suggestions for logical groupings
- Strategic planning based on work clusters

### During Feature Planning
- Promote entire coherent groups when ready
- Understand related work that belongs together
- Plan iterations around natural work boundaries

## Heuristic Grouping Examples

### Feature Similarity
- "Authentication Features": login, signup, password reset, 2FA
- "Search & Filtering": advanced search, filters, sorting, pagination  
- "User Profile": profile editing, preferences, settings, avatars

### Technical Domain Similarity
- "Performance Optimization": caching, database optimization, lazy loading
- "Frontend Components": reusable UI components, design system updates
- "API Development": new endpoints, data validation, error handling

### Strategic Similarity  
- "User Experience": onboarding flow, help system, accessibility
- "Developer Experience": tooling, documentation, testing infrastructure
- "Business Intelligence": analytics, reporting, metrics dashboard