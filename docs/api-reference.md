# API Reference

Complete reference for all MCP tools and endpoints available in Claude Work Tracker.

## Table of Contents

1. [Core Work Management](#core-work-management)
2. [Smart References](#smart-references)
3. [Historical Context](#historical-context)
4. [Future Work Management](#future-work-management)
5. [Visualization Tools](#visualization-tools)
6. [Data Types](#data-types)

## Core Work Management

### get_work_state

Retrieves the current work state including todos, findings, and session summary.

```typescript
get_work_state(): WorkState

// Returns
{
  current_session: string,
  active_todos: WorkItem[],
  recent_findings: Finding[],
  session_summary: SessionSummary,
  smart_suggestions?: ContextualSuggestion[]  // When enhanced
}

// Example
const state = await get_work_state()
console.log(`${state.active_todos.length} active todos`)
```

### save_plan

Saves a structured plan with steps for future reference.

```typescript
save_plan(
  content: string,      // Plan description
  steps: string[]       // Array of plan steps
): WorkItem

// Example
await save_plan(
  "Implement user authentication system",
  [
    "Create login component",
    "Add form validation", 
    "Integrate with backend API",
    "Add session management",
    "Implement logout functionality"
  ]
)
```

### save_proposal

Saves an architectural decision or proposal with rationale.

```typescript
save_proposal(
  content: string,      // Proposal description
  rationale: string     // Reasoning behind the proposal
): WorkItem

// Example
await save_proposal(
  "Use JWT tokens for authentication",
  "JWTs are stateless, scalable, and work well with our microservices architecture"
)
```

### search_work_items

Search through all work items with optional type filtering.

```typescript
search_work_items(
  query: string,        // Search query
  type?: WorkItemType   // Optional: 'todo' | 'plan' | 'proposal' | 'finding'
): WorkItem[]

// Example
const authWork = await search_work_items("authentication", "plan")
```

## Smart References

### get_contextual_suggestions

Get AI-powered suggestions based on current active work.

```typescript
get_contextual_suggestions(): ContextualSuggestion[]

// Returns array of
{
  type: 'promote_historical' | 'review_conflict' | 'reference_decision' | 'continue_work',
  priority: 'high' | 'medium' | 'low',
  message: string,
  target_item_id: string,
  confidence: number,
  action_hint?: string
}

// Example
const suggestions = await get_contextual_suggestions()
for (const s of suggestions.filter(s => s.priority === 'high')) {
  console.log(`⚡ ${s.message}`)
}
```

### generate_smart_references

Generate automatic references for a specific work item.

```typescript
generate_smart_references(
  item_id: string       // Work item ID
): SmartReference[]

// Returns array of
{
  target_item_id: string,
  similarity_score: number,
  relationship_type: 'related' | 'continuation' | 'conflict' | 'dependency',
  confidence: number,
  metadata: {
    common_keywords: string[],
    domain_overlap: string[],
    code_location_overlap: string[],
    strategic_alignment: string
  }
}

// Example
const refs = await generate_smart_references("todo-123")
console.log(`Found ${refs.length} related items`)
```

### calculate_similarity

Calculate semantic similarity between two work items.

```typescript
calculate_similarity(
  item_id_1: string,    // First work item
  item_id_2: string     // Second work item
): SimilarityScore

// Returns
{
  total_score: number,         // 0-1 overall similarity
  keyword_score: number,       // Keyword overlap
  domain_score: number,        // Domain alignment
  location_score: number,      // Code location similarity
  strategic_score: number,     // Strategic theme alignment
  content_score: number,       // Content similarity
  common_keywords: string[],
  domain_overlap: string[],
  code_location_overlap: string[],
  strategic_alignment: string
}

// Example
const similarity = await calculate_similarity("plan-auth", "todo-login")
if (similarity.total_score > 0.8) {
  console.log("Highly related items!")
}
```

### get_enhanced_work_state

Get work state enhanced with smart referencing context and suggestions.

```typescript
get_enhanced_work_state(): EnhancedWorkState

// Returns WorkState plus
{
  smart_suggestions: ContextualSuggestion[],
  reference_summary: {
    total_suggestions: number,
    high_priority: number,
    suggestion_types: { [type: string]: number }
  }
}

// Example
const enhanced = await get_enhanced_work_state()
console.log(`${enhanced.reference_summary.high_priority} high priority suggestions`)
```

## Historical Context

### query_history

Search through historical work items with advanced filtering.

```typescript
query_history(
  keyword: string,           // Search term
  start_date?: string,       // YYYY-MM-DD format
  end_date?: string,         // YYYY-MM-DD format
  type?: string             // Work item type
): HistoricalWorkItem[]

// Returns work items with contextual relevance
{
  ...WorkItem,
  contextual_relevance?: {
    confidence: number,
    relationship_type: string,
    priority: string,
    action_hint: string
  }
}

// Example
const pastAuth = await query_history(
  "authentication",
  "2025-07-01",
  "2025-07-11",
  "proposal"
)
```

### get_historical_context

Retrieve detailed information about a specific historical item.

```typescript
get_historical_context(
  item_id: string           // Historical item ID or filename
): WorkItem | null

// Example
const decision = await get_historical_context("decision-use-postgres-2025-07-08")
console.log(`Rationale: ${decision.metadata.decision_rationale}`)
```

### summarize_period

Generate a summary of work activity for a time period.

```typescript
summarize_period(
  start_date: string,       // YYYY-MM-DD format
  end_date: string          // YYYY-MM-DD format
): PeriodSummary

// Returns
{
  period: { start: string, end: string },
  total_items: number,
  by_type: { [type: string]: number },
  by_status: { [status: string]: number },
  key_items: Array<{
    id: string,
    type: string,
    content: string,
    timestamp: string
  }>
}

// Example
const weekSummary = await summarize_period("2025-07-01", "2025-07-07")
console.log(`Completed ${weekSummary.by_status.completed} items`)
```

### promote_to_active

Move a historical item back to active context.

```typescript
promote_to_active(
  item_id: string           // Historical item ID
): WorkItem

// Example
await promote_to_active("plan-caching-strategy-2025-06-30")
// Old plan now available in active context
```

### archive_active_item

Move an active item to historical archive.

```typescript
archive_active_item(
  item_id: string           // Active item ID
): WorkItem

// Example
await archive_active_item("todo-old-feature")
// Removes from active context, preserves in history
```

## Future Work Management

### defer_to_future

Intelligently defer work for future implementation.

```typescript
defer_to_future(
  content: string,          // Work description
  reason: string,           // Deferral reason
  type?: string            // 'plan' | 'proposal' | 'todo' | 'idea'
): FutureWorkItem

// Returns
{
  id: string,
  content: string,
  similarity_metadata: SimilarityMetadata,
  context: {
    deprioritized_reason: string,
    suggested_group: string | null
  },
  grouping_status: 'grouped' | 'ungrouped'
}

// Example
await defer_to_future(
  "Add social login providers",
  "Out of scope for MVP - focus on basic auth first",
  "todo"
)
```

### list_future_groups

View all future work groups and ungrouped items.

```typescript
list_future_groups(): FutureWorkOverview

// Returns
{
  groups: WorkGroup[],
  ungrouped_items: FutureWorkItem[],
  suggestions: GroupingSuggestion[],
  total_items: number
}

// Example
const future = await list_future_groups()
console.log(`${future.groups.length} work groups ready`)
```

### groom_future_work

Analyze and reorganize future work with intelligent suggestions.

```typescript
groom_future_work(): GroomingAnalysis

// Returns
{
  overview: FutureWorkOverview,
  suggestions: GroupingSuggestion[],
  similarity_analysis: {
    feature_clusters: { [domain: string]: number },
    technical_domains: { [domain: string]: number },
    code_locations: { [location: string]: number }
  },
  recommendations: string[]
}

// Example
const analysis = await groom_future_work()
for (const suggestion of analysis.suggestions) {
  console.log(`Consider grouping: ${suggestion.suggested_group_name}`)
}
```

### create_work_group

Create a logical group of related future work items.

```typescript
create_work_group(
  name: string,             // Group name
  description: string,      // What the group contains
  item_ids: string[]        // Future work item IDs
): WorkGroup

// Example
await create_work_group(
  "authentication-phase-2",
  "Advanced authentication features including OAuth and 2FA",
  ["item-123", "item-124", "item-125"]
)
```

### promote_work_group

Promote an entire group of related work to active context.

```typescript
promote_work_group(
  group_name: string        // Name of work group
): WorkItem[]

// Example
const items = await promote_work_group("authentication-phase-2")
console.log(`Promoted ${items.length} items to active work`)
```

## Visualization Tools

### generate_reference_map

Generate a complete reference map for current work context.

```typescript
generate_reference_map(): ReferenceMap

// Returns
{
  nodes: ReferenceNode[],
  edges: ReferenceEdge[],
  clusters: ReferenceCluster[],
  summary: {
    total_items: number,
    total_references: number,
    cluster_count: number,
    orphaned_items: number
  }
}

// Example
const map = await generate_reference_map()
console.log(`Work graph: ${map.nodes.length} nodes, ${map.edges.length} edges`)
```

### generate_focused_reference_map

Generate a reference map focused on a specific work item.

```typescript
generate_focused_reference_map(
  item_id: string,          // Work item to focus on
  depth?: number            // Traversal depth (1-5, default: 2)
): ReferenceMap

// Example
const focusedMap = await generate_focused_reference_map("current-todo", 3)
// Shows 3 levels of related work
```

### find_reference_path

Find the reference path between two work items.

```typescript
find_reference_path(
  source_id: string,        // Starting work item
  target_id: string         // Target work item
): string[]                 // Array of item IDs forming path

// Example
const path = await find_reference_path("todo-login", "plan-auth")
if (path.length > 0) {
  console.log(`Connected through ${path.length - 1} items`)
}
```

### visualize_references

Generate ASCII visualization of work item references.

```typescript
visualize_references(): { visualization: string }

// Example output
=== Work Item Reference Map ===

📋 ACTIVE WORK:
  ● todo: Implement user authentication
    References:
      → plan: Design auth architecture [█████]
      ⚠ proposal: Switch to OAuth-only [███░░]
      ~ todo: User profile implementation [████░]

🔗 REFERENCE CLUSTERS:
  📁 Authentication Features (5 items)
     Themes: user-management, security
     Type: feature

📊 SUMMARY:
  Items: 12
  References: 18
  Clusters: 3
  Orphaned: 1
```

## Data Types

### WorkItem

```typescript
interface WorkItem {
  id: string
  type: 'todo' | 'plan' | 'proposal' | 'finding' | 'report' | 'summary'
  content: string
  status: 'pending' | 'in_progress' | 'completed'
  context: GitContext
  session_id: string
  timestamp: string
  metadata?: {
    priority?: 'low' | 'medium' | 'high'
    plan_steps?: string[]
    decision_rationale?: string
    similarity_metadata?: SimilarityMetadata
    smart_references?: WorkItemReference[]
  }
}
```

### SimilarityMetadata

```typescript
interface SimilarityMetadata {
  keywords: string[]
  feature_domain: string
  technical_domain: string
  code_locations: string[]
  strategic_theme: string
}
```

### ContextualSuggestion

```typescript
interface ContextualSuggestion {
  type: 'promote_historical' | 'review_conflict' | 'reference_decision' | 'continue_work'
  priority: 'high' | 'medium' | 'low'
  message: string
  target_item_id: string
  confidence: number
  action_hint?: string
}
```

### WorkGroup

```typescript
interface WorkGroup {
  id: string
  name: string
  description: string
  items: string[]
  similarity_score: number
  strategic_value: 'low' | 'medium' | 'high'
  estimated_effort: 'small' | 'medium' | 'large'
  readiness_status: 'ready' | 'blocked' | 'waiting'
  created_date: string
  last_updated: string
}
```

## Error Handling

All tools return consistent error responses:

```typescript
{
  success: false,
  error: string  // Human-readable error message
}
```

Common error scenarios:
- Missing required parameters
- Work item not found
- Invalid date formats
- File system errors
- JSON parsing errors

## Rate Limits and Performance

- No rate limits on MCP tool calls
- Similarity calculations are optimized for <100ms response
- Historical queries use indexed search for fast retrieval
- Reference generation is cached for repeated calls
- Visualization limited to reasonable graph sizes (configurable depth)

## Best Practices

1. **Batch Operations**: Use multiple tool calls in parallel when needed
2. **Contextual Queries**: Always check suggestions before starting work
3. **Historical Awareness**: Query history before implementing new features
4. **Smart Deferral**: Use defer_to_future for scope management
5. **Reference Validation**: Verify high-confidence references before acting