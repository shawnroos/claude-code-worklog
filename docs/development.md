# Developer Guide

Complete guide for contributing to the Claude Code Work Tracking System.

## 🛠️ Development Setup

### Prerequisites

- **Node.js** 18.0.0 or higher
- **npm** 8.0.0 or higher
- **jq** for JSON processing
- **git** for version control
- **bash** 4.0+ for shell scripts

### Quick Start

```bash
# Clone the repository
git clone https://github.com/shawnroos/claude-work-tracker.git
cd claude-work-tracker

# Install dependencies
npm install

# Build the project
npm run build

# Run tests
npm test

# Start development server
npm run dev
```

### Development Environment

```bash
# Set up development environment
export NODE_ENV=development
export CLAUDE_WORK_DEBUG=1

# Install development dependencies
npm install --dev

# Set up git hooks
npx husky install
```

## 📁 Project Structure

```
claude-work-tracker/
├── src/                        # TypeScript source code
│   ├── index.ts               # MCP server entry point
│   ├── services/              # Core business logic
│   │   └── WorkStateManager.ts
│   ├── tools/                 # MCP tool implementations
│   │   └── index.ts
│   └── types/                 # TypeScript type definitions
│       └── index.ts
├── scripts/                   # Bash automation scripts
│   ├── session-*.sh          # Session lifecycle hooks
│   ├── tool-*.sh             # Tool capture hooks
│   ├── work-*.sh             # Work management utilities
│   └── update-*.sh           # State aggregation scripts
├── docs/                      # Documentation
├── examples/                  # Configuration examples
├── dist/                      # Compiled JavaScript (generated)
├── node_modules/              # Dependencies (generated)
├── package.json              # Node.js configuration
├── tsconfig.json             # TypeScript configuration
└── README.md                 # Project overview
```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
npm test

# Run specific test suites
npm run test:unit
npm run test:integration
npm run test:hooks

# Run tests with coverage
npm run test:coverage

# Run tests in watch mode
npm run test:watch
```

### Test Structure

```bash
tests/
├── unit/                      # Unit tests
│   ├── services/
│   ├── tools/
│   └── types/
├── integration/               # Integration tests
│   ├── hooks/
│   ├── mcp-server/
│   └── cross-worktree/
├── fixtures/                  # Test data
└── helpers/                   # Test utilities
```

### Writing Tests

```typescript
// Example unit test
import { WorkStateManager } from '../src/services/WorkStateManager'

describe('WorkStateManager', () => {
  let manager: WorkStateManager
  
  beforeEach(() => {
    manager = new WorkStateManager()
  })
  
  test('should save work item', () => {
    const item = {
      id: 'test-id',
      type: 'todo',
      content: 'Test todo',
      status: 'pending',
      // ... other fields
    }
    
    manager.saveWorkItem(item)
    
    // Assertions
    expect(manager.getWorkItem('test-id')).toEqual(item)
  })
})
```

### Hook Testing

```bash
# Test hooks manually
echo '{"sessionId": "test", "workingDirectory": "/tmp"}' | ./scripts/session-complete.sh

# Test plan capture
echo '{"toolName": "exit_plan_mode", "toolInput": "{\"plan\": \"test\"}"}' | ./scripts/tool-complete-plan-capture.sh

# Test MCP server
echo '{"jsonrpc": "2.0", "method": "tools/list", "id": 1}' | node dist/index.js
```

## 🔧 Code Style and Standards

### TypeScript Guidelines

```typescript
// Use strict typing
interface WorkItem {
  id: string
  type: WorkItemType  // Use type unions, not enums
  content: string
  // ... other fields
}

// Prefer interfaces over types for objects
interface GitContext {
  branch: string
  worktree: string
}

// Use descriptive names
const getCurrentWorkState = (): WorkState => {
  // Implementation
}

// Error handling
try {
  const result = await riskyOperation()
  return { success: true, data: result }
} catch (error) {
  return { success: false, error: error.message }
}
```

### Shell Script Guidelines

```bash
#!/bin/bash
# Always use strict mode
set -euo pipefail

# Use descriptive variable names
WORK_STATE_DIR="$HOME/.claude/work-state"
SESSION_ID="$1"

# Check required parameters
if [ -z "$SESSION_ID" ]; then
    echo "Usage: $0 <session_id>"
    exit 1
fi

# Use functions for reusable code
get_git_context() {
    local branch=""
    if git rev-parse --git-dir > /dev/null 2>&1; then
        branch=$(git branch --show-current 2>/dev/null || echo "unknown")
    fi
    echo "$branch"
}

# Log for debugging
echo "[$(date)] Processing session: $SESSION_ID" >> ~/.claude/hooks.log
```

### Documentation Standards

```typescript
/**
 * Manages work state across sessions and worktrees
 */
class WorkStateManager {
  /**
   * Saves a work item to persistent storage
   * @param workItem - The work item to save
   * @throws Error if work item is invalid
   */
  public saveWorkItem(workItem: WorkItem): void {
    // Implementation
  }
}
```

## 🏗️ Architecture Patterns

### Service Layer Pattern

```typescript
// services/WorkStateManager.ts
export class WorkStateManager {
  private readonly baseDir: string
  
  constructor(baseDir?: string) {
    this.baseDir = baseDir || join(process.env.HOME || '', '.claude')
  }
  
  // Public interface methods
  public getCurrentWorkState(): WorkState { /* ... */ }
  public saveWorkItem(item: WorkItem): void { /* ... */ }
  
  // Private implementation methods
  private ensureDirectories(): void { /* ... */ }
  private readJsonFile<T>(filePath: string): T | null { /* ... */ }
}
```

### Tool Registry Pattern

```typescript
// tools/index.ts
export class WorkTrackingTools {
  private tools: Map<string, Tool> = new Map()
  
  constructor() {
    this.registerDefaultTools()
  }
  
  public getTools(): Tool[] {
    return Array.from(this.tools.values())
  }
  
  public async handleToolCall(name: string, params: any): Promise<McpToolResponse> {
    const tool = this.tools.get(name)
    if (!tool) {
      throw new Error(`Unknown tool: ${name}`)
    }
    return tool.handler(params)
  }
}
```

### Hook Pattern

```bash
# scripts/session-complete.sh
#!/bin/bash

# Hook pattern: Input → Process → Output → Side Effects

# 1. Input validation
INPUT=$(cat)
SESSION_ID=$(echo "$INPUT" | jq -r '.sessionId // empty')

if [ -z "$SESSION_ID" ]; then
    exit 0
fi

# 2. Process data
process_session_data() {
    local session_id="$1"
    # Processing logic
}

# 3. Generate output
generate_summary() {
    local session_id="$1"
    # Output generation
}

# 4. Side effects (file writes, notifications)
update_global_state() {
    local session_id="$1"
    # State updates
}

# Execute hook
process_session_data "$SESSION_ID"
generate_summary "$SESSION_ID"
update_global_state "$SESSION_ID"
```

## 🔌 Adding New Features

### Adding a New MCP Tool

1. **Define the tool interface**:
```typescript
// src/tools/index.ts
{
  name: 'new_tool',
  description: 'Description of what the tool does',
  inputSchema: {
    type: 'object',
    properties: {
      param1: { type: 'string', description: 'Parameter description' }
    },
    required: ['param1']
  }
}
```

2. **Implement the handler**:
```typescript
private handleNewTool(params: McpToolParams): McpToolResponse {
  const { param1 } = params
  
  // Validation
  if (!param1) {
    return { success: false, error: 'Missing required parameter: param1' }
  }
  
  // Implementation
  try {
    const result = this.processNewTool(param1)
    return { success: true, data: result }
  } catch (error) {
    return { success: false, error: error.message }
  }
}
```

3. **Add to tool registry**:
```typescript
case 'new_tool':
  return this.handleNewTool(params)
```

4. **Add tests**:
```typescript
describe('new_tool', () => {
  test('should handle valid input', async () => {
    const result = await tools.handleToolCall('new_tool', { param1: 'test' })
    expect(result.success).toBe(true)
  })
})
```

### Adding a New Hook

1. **Create the hook script**:
```bash
#!/bin/bash
# scripts/new-hook.sh

# Hook implementation
INPUT=$(cat)
# Process input and generate output
```

2. **Register the hook**:
```json
// examples/settings.local.json
{
  "hooks": {
    "new_event": "~/.claude/scripts/new-hook.sh"
  }
}
```

3. **Test the hook**:
```bash
echo '{"testData": "value"}' | ./scripts/new-hook.sh
```

### Adding a New Intelligence Type

1. **Update type definitions**:
```typescript
// src/types/index.ts
type WorkItemType = 
  | 'todo'
  | 'plan'
  | 'proposal'
  | 'new_intelligence_type'  // Add new type
```

2. **Add classification logic**:
```typescript
// src/services/WorkStateManager.ts
private classifyWorkItem(content: string): WorkItemType {
  if (this.isNewIntelligenceType(content)) {
    return 'new_intelligence_type'
  }
  // ... existing logic
}

private isNewIntelligenceType(content: string): boolean {
  // Classification logic
  return content.includes('new_pattern')
}
```

3. **Update capture hooks**:
```bash
# scripts/tool-complete-plan-capture.sh
case "$intelligence_type" in
    "new_intelligence_type")
        # Handle new intelligence type
        ;;
esac
```

## 🐛 Debugging

### Debug Mode

```bash
# Enable debug logging
export CLAUDE_WORK_DEBUG=1

# Run with debug output
CLAUDE_WORK_DEBUG=1 npm start
CLAUDE_WORK_DEBUG=1 ~/.claude/scripts/work-status.sh
```

### Logging

```typescript
// Add debug logging
const debug = process.env.CLAUDE_WORK_DEBUG === '1'

if (debug) {
  console.error(`[DEBUG] Processing work item: ${workItem.id}`)
}
```

```bash
# In bash scripts
if [ "$CLAUDE_WORK_DEBUG" = "1" ]; then
    echo "[DEBUG] Processing session: $SESSION_ID" >&2
fi
```

### Common Debug Commands

```bash
# Check file system state
ls -la ~/.claude/
ls -la ~/.claude/work-intelligence/
ls -la .claude-work/

# Validate JSON files
jq . ~/.claude/settings.local.json
jq . ~/.claude/work-tracking-config.json

# Test hooks individually
echo '{"sessionId": "debug"}' | ~/.claude/scripts/session-complete.sh

# Check logs
tail -f ~/.claude/hooks.log
```

## 📦 Building and Distribution

### Build Process

```bash
# Clean build
npm run clean
npm run build

# Development build with watch
npm run dev

# Production build
npm run build:prod
```

### Distribution

```bash
# Create distribution package
npm pack

# Publish to npm (if configured)
npm publish

# Create GitHub release
git tag v1.0.0
git push origin v1.0.0
```

## 🔄 Git Workflow

### Branch Strategy

```bash
# Main branches
main          # Production-ready code
develop       # Development integration

# Feature branches
feature/      # New features
bugfix/       # Bug fixes
hotfix/       # Emergency fixes
docs/         # Documentation updates
```

### Commit Convention

```bash
# Commit format
type(scope): description

# Examples
feat(mcp): add new work intelligence tool
fix(hooks): resolve session capture issue
docs(api): update API reference documentation
refactor(storage): improve file system performance
test(integration): add cross-worktree tests
```

### Pull Request Process

1. **Create feature branch**:
```bash
git checkout -b feature/new-feature
```

2. **Implement changes**:
```bash
# Make changes
git add .
git commit -m "feat(scope): description"
```

3. **Test thoroughly**:
```bash
npm test
npm run lint
npm run build
```

4. **Create pull request**:
- Clear description of changes
- Link to related issues
- Include test results
- Update documentation

## 📊 Performance Considerations

### Optimization Guidelines

1. **Lazy Loading**: Load data on-demand
2. **Caching**: Cache frequently accessed data
3. **Batching**: Batch file operations
4. **Async Processing**: Use async for I/O operations

### Memory Management

```typescript
// Use weak references for temporary data
const cache = new WeakMap<object, WorkItem>()

// Clean up resources
class ResourceManager {
  private cleanup(): void {
    // Release resources
  }
}
```

### File System Optimization

```bash
# Batch file operations
{
  echo "data1"
  echo "data2"
  echo "data3"
} > output.txt

# Use efficient search
find ~/.claude/work-intelligence -name "*.json" -newer reference.txt
```

## 🔒 Security Considerations

### Data Privacy

- **Local Storage**: Keep all data on user's machine
- **No Network**: Avoid network requests with user data
- **File Permissions**: Set appropriate file permissions
- **Sanitization**: Sanitize user input

### Secure Coding

```typescript
// Input validation
const validateWorkItem = (item: unknown): item is WorkItem => {
  return typeof item === 'object' &&
         item !== null &&
         typeof (item as any).id === 'string' &&
         typeof (item as any).content === 'string'
}

// Path traversal protection
const sanitizePath = (path: string): string => {
  return path.replace(/\.\./g, '').replace(/\/+/g, '/')
}
```

## 🤝 Contributing Guidelines

### Code Review Checklist

- [ ] Code follows style guidelines
- [ ] Tests pass and provide adequate coverage
- [ ] Documentation updated
- [ ] No sensitive data exposed
- [ ] Performance impact considered
- [ ] Backward compatibility maintained

### Issue Templates

```markdown
## Bug Report

**Description**
Clear description of the bug

**Steps to Reproduce**
1. Step 1
2. Step 2
3. Step 3

**Expected Behavior**
What should happen

**Actual Behavior**
What actually happens

**Environment**
- OS: [e.g. macOS 14.0]
- Node.js: [e.g. v18.17.0]
- Version: [e.g. v1.0.0]
```

### Feature Request Template

```markdown
## Feature Request

**Problem**
What problem does this solve?

**Proposed Solution**
How should this be implemented?

**Alternatives**
Other approaches considered

**Additional Context**
Any other relevant information
```

## 📚 Additional Resources

### Documentation

- [Installation Guide](installation.md)
- [API Reference](api-reference.md)
- [Architecture](architecture.md)
- [Configuration](configuration.md)
- [Troubleshooting](troubleshooting.md)

### External Resources

- [Model Context Protocol](https://github.com/modelcontextprotocol/specification)
- [Claude Code Documentation](https://docs.anthropic.com/claude-code)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [Bash Scripting Guide](https://www.gnu.org/software/bash/manual/)

### Community

- [GitHub Issues](https://github.com/shawnroos/claude-work-tracker/issues)
- [GitHub Discussions](https://github.com/shawnroos/claude-work-tracker/discussions)
- [Claude Code Community](https://claude.ai/code)

This guide should help you get started with contributing to the Claude Code Work Tracking System. Welcome to the project!