# User Guide

This guide covers everything you need to know to effectively use Claude Work Tracker in your daily development workflow.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Core Concepts](#core-concepts)
3. [Daily Workflow](#daily-workflow)
4. [Commands Reference](#commands-reference)
5. [Smart Features](#smart-features)
6. [Best Practices](#best-practices)
7. [Troubleshooting](#troubleshooting)

## Getting Started

### First Session

When you start working with Claude Code after installation:

```bash
# Claude automatically checks for pending work
$ claude
Checking for pending work...
No pending todos found.

# Start working - todos are captured automatically
You: Let's implement user authentication
Claude: I'll help you implement user authentication. Let me create a todo list...
[Todo automatically saved]
```

### Understanding the System

The work tracker operates on three levels:

1. **Active Work** - Current todos, plans, and decisions
2. **Historical Context** - Completed work searchable for reference
3. **Future Work** - Deferred items intelligently grouped

## Core Concepts

### Work Items

Everything you do is captured as a work item:

- **Todos** - Tasks to complete
- **Plans** - Structured approaches with steps
- **Proposals** - Architectural decisions
- **Findings** - Research results and insights
- **Decisions** - Choices made with rationale

### Smart References

The system automatically creates references between related work:

```
Current: "Implement login form"
├── Continues → "Design auth flow" (yesterday)
├── Related → "User profile form" (80% similar)
└── Conflicts → "Remove password auth" (warning!)
```

### Context Awareness

Work is always associated with:
- Git branch
- Working directory
- Session timestamp
- Related files
- Similarity metadata

## Daily Workflow

### Starting Your Day

```bash
# Check what's pending
/work status

# The system shows:
# - Active todos from last session
# - Smart suggestions for what to work on
# - Relevant historical context
# - Any conflicts to be aware of
```

### During Development

The system works automatically in the background:

```typescript
// When Claude creates a plan
Claude: "Here's my plan for implementing auth:
1. Create login component
2. Add validation
3. Connect to backend"
// → Automatically saved as a plan

// When you make decisions
You: "Let's use JWT tokens instead of sessions"
Claude: "Good choice because..."
// → Decision and rationale captured

// When you defer work
You: "Let's implement social login later"
// → Automatically grouped in future work
```

### Ending Your Session

```bash
# Save with a summary
/work save "Completed basic auth, JWT implementation next"

# Or just let auto-save handle it
# Work is automatically preserved
```

## Commands Reference

### Slash Commands

| Command | Description | Example |
|---------|-------------|---------|
| `/work` | Show help and status | `/work` |
| `/work status` | Detailed current state | `/work status` |
| `/work load` | Load work for branch | `/work load feature-auth` |
| `/work save` | Save with note | `/work save "Ready for review"` |
| `/work search` | Search all work | `/work search authentication` |

### MCP Tools

These are available to Claude automatically:

#### Work State Management
- `get_work_state()` - Get current work context
- `save_plan(content, steps)` - Save structured plan
- `save_proposal(content, rationale)` - Save architectural decision

#### Smart References
- `get_contextual_suggestions()` - Get AI-powered suggestions
- `generate_smart_references(item_id)` - Create references for item
- `calculate_similarity(item1, item2)` - Compare work items
- `visualize_references()` - See work relationships

#### Historical Context
- `query_history(keyword, type)` - Search past work
- `get_historical_context(item_id)` - Get specific item details
- `promote_to_active(item_id)` - Bring historical item back

#### Future Work
- `defer_to_future(content, reason)` - Defer with smart grouping
- `list_future_groups()` - See deferred work groups
- `promote_work_group(name)` - Activate a group of work

## Smart Features

### Contextual Suggestions

The system analyzes your current work and suggests:

```
🧠 Contextual Suggestions:
  
⚡ HIGH PRIORITY:
  "Review security findings from auth audit"
  → Found conflict with current implementation
  → Action: Use get_historical_context('auth-audit-2025-07-09')

📌 MEDIUM PRIORITY:  
  "Consider session management proposal"
  → Continuation of current work
  → Action: Could provide useful patterns

ℹ️ RELATED:
  "Similar implementation in user-profiles branch"
  → 85% similarity score
  → Action: Check for reusable code
```

### Automatic Grouping

Future work is intelligently organized:

```
📁 Future Work Groups:

🔐 Authentication Enhancements (3 items)
  - Add OAuth providers
  - Implement 2FA
  - Add password reset flow
  Theme: User Management
  Readiness: Waiting on base auth

⚡ Performance Optimizations (2 items)
  - Add caching layer
  - Optimize database queries
  Theme: System Performance
  Readiness: After MVP
```

### Conflict Detection

The system warns about contradictions:

```
⚠️ CONFLICT DETECTED:
Current: "Use MongoDB for user data"
Conflicts with: "Decision: Use PostgreSQL for all data" (3 days ago)

Suggestion: Review previous decision rationale before proceeding
Use: get_historical_context('decision-postgres-2025-07-08')
```

## Best Practices

### 1. Let the System Work Automatically

Don't manually track things - let the system capture:
- Todos from Claude's task lists
- Plans from planning discussions  
- Decisions from architectural choices
- Insights from research

### 2. Use Natural Language

The system understands context:
```
Good: "Implement user login with email/password"
Good: "Research best practices for JWT refresh tokens"
Good: "Defer social login - not needed for MVP"
```

### 3. Review Suggestions

Start each session by checking suggestions:
```bash
/work status
# Review smart suggestions before starting
# They often prevent rework or mistakes
```

### 4. Leverage Historical Context

Before implementing something new:
```typescript
// Check if it's been tried before
await query_history("cache implementation")

// Understand previous decisions
await get_historical_context("decision-no-redis")
```

### 5. Keep Focus with Future Work

Don't lose good ideas:
```typescript
// During discussion
You: "We should add email templates later"
Claude: "I'll defer that to future work"

// Later when ready
await promote_work_group("email-features")
```

## Troubleshooting

### Common Issues

**No suggestions appearing**
- Ensure you have active work items
- Check that similarity metadata is being extracted
- Verify MCP server is running

**Work not being saved**
- Check file permissions in `.claude-work/`
- Ensure proper JSON formatting
- Verify git repository is initialized

**Incorrect branch association**
- Run `git status` to verify branch
- Check worktree configuration
- Ensure git context is available

### Debug Commands

```bash
# Check system health
~/.claude/scripts/work-presentation.sh test

# View raw work state
cat .claude-work/active/current-work-context.json

# Check MCP server logs
npm run dev  # Shows detailed logs
```

### Getting Help

1. Check the [Troubleshooting Guide](troubleshooting.md)
2. Review [API Reference](api-reference.md) for tool details
3. Submit issues on GitHub with:
   - Error messages
   - Steps to reproduce
   - System configuration

## Advanced Usage

### Custom Workflows

Create aliases for common patterns:

```bash
# In ~/.zshrc or ~/.bashrc
alias work-auth="/work search auth && /work load feature-auth"
alias work-review="/work status && /work visualize"
```

### Integration with Git Hooks

```bash
# In .git/hooks/post-checkout
#!/bin/bash
~/.claude/scripts/work.sh load
```

### Batch Operations

```typescript
// Process multiple related items
const authItems = await query_history("auth", type: "todo")
for (const item of authItems) {
  await generate_smart_references(item.id)
}
```

Remember: The system is designed to be invisible when you don't need it and invaluable when you do. Let it work in the background and enjoy the benefits of never losing context again.