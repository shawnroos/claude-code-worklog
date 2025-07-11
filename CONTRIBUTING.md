# Contributing to Claude Code Work Tracker

Thank you for your interest in contributing to the Claude Code Work Tracking System! This guide will help you get started with contributing to the project.

## 🎯 Ways to Contribute

### 🐛 **Bug Reports**
- Report bugs using GitHub Issues
- Include detailed reproduction steps
- Provide system information and logs
- Use the bug report template

### 💡 **Feature Requests**
- Suggest new features or improvements
- Provide use cases and rationale
- Consider implementation complexity
- Use the feature request template

### 📝 **Documentation**
- Improve existing documentation
- Add missing documentation
- Fix typos and formatting
- Translate documentation

### 🔧 **Code Contributions**
- Fix bugs and implement features
- Improve performance and reliability
- Add tests and examples
- Enhance developer experience

## 🚀 Getting Started

### 1. **Fork and Clone**

```bash
# Fork the repository on GitHub
# Then clone your fork
git clone https://github.com/your-username/claude-work-tracker.git
cd claude-work-tracker

# Add upstream remote
git remote add upstream https://github.com/shawnroos/claude-work-tracker.git
```

### 2. **Set Up Development Environment**

```bash
# Install dependencies
npm install

# Build the project
npm run build

# Run tests
npm test

# Start development server
npm run dev
```

### 3. **Install Development Tools**

```bash
# Install development dependencies
npm install --dev

# Set up pre-commit hooks
npx husky install

# Install shell tools
# macOS
brew install jq shellcheck

# Ubuntu/Debian
sudo apt-get install jq shellcheck
```

## 📋 Development Workflow

### 1. **Create a Feature Branch**

```bash
# Create and switch to a new branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b bugfix/issue-description
```

### 2. **Make Changes**

```bash
# Make your changes
# Add tests for new functionality
# Update documentation as needed

# Check your changes
npm run lint
npm test
npm run build
```

### 3. **Commit Changes**

```bash
# Stage changes
git add .

# Commit with descriptive message
git commit -m "feat(scope): add new feature description"

# Push to your fork
git push origin feature/your-feature-name
```

### 4. **Create Pull Request**

- Go to GitHub and create a Pull Request
- Fill out the PR template completely
- Link related issues
- Request review from maintainers

## 📏 Code Standards

### **TypeScript Code Style**

```typescript
// Use strict typing
interface WorkItem {
  id: string
  type: WorkItemType
  content: string
  status: WorkItemStatus
}

// Use descriptive function names
const getCurrentWorkState = (): WorkState => {
  // Implementation
}

// Handle errors properly
try {
  const result = await riskyOperation()
  return { success: true, data: result }
} catch (error) {
  return { success: false, error: error.message }
}

// Use async/await for promises
const fetchData = async (): Promise<Data> => {
  const response = await fetch('/api/data')
  return response.json()
}
```

### **Shell Script Style**

```bash
#!/bin/bash
# Always use strict mode
set -euo pipefail

# Use descriptive variable names
readonly WORK_STATE_DIR="$HOME/.claude/work-state"
readonly SESSION_ID="$1"

# Check required parameters
if [ -z "$SESSION_ID" ]; then
    echo "Usage: $0 <session_id>" >&2
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

# Quote variables to prevent word splitting
echo "Processing session: $SESSION_ID"

# Log for debugging
echo "[$(date)] Processing session: $SESSION_ID" >> ~/.claude/hooks.log
```

### **Documentation Style**

```markdown
# Use Clear Headers

## Section Organization
- Use consistent header levels
- Include table of contents for long documents
- Add cross-references between sections

### Code Examples
```bash
# Always include comments in code examples
command --option value  # Explanation of what this does
```

### Links
- Use descriptive link text
- Verify all links work
- Use relative links for internal documents
```

## 🧪 Testing Guidelines

### **Unit Tests**

```typescript
// Test file: src/services/WorkStateManager.test.ts
import { WorkStateManager } from './WorkStateManager'

describe('WorkStateManager', () => {
  let manager: WorkStateManager
  
  beforeEach(() => {
    manager = new WorkStateManager()
  })
  
  describe('saveWorkItem', () => {
    it('should save work item successfully', () => {
      const item = createMockWorkItem()
      
      manager.saveWorkItem(item)
      
      expect(manager.getWorkItem(item.id)).toEqual(item)
    })
    
    it('should throw error for invalid work item', () => {
      expect(() => {
        manager.saveWorkItem(null as any)
      }).toThrow('Invalid work item')
    })
  })
})
```

### **Integration Tests**

```typescript
// Test file: tests/integration/hooks.test.ts
import { execSync } from 'child_process'
import { readFileSync } from 'fs'

describe('Hook Integration', () => {
  it('should capture session completion', () => {
    const input = {
      sessionId: 'test-session',
      workingDirectory: '/tmp/test'
    }
    
    const result = execSync(
      `echo '${JSON.stringify(input)}' | ./scripts/session-complete.sh`,
      { encoding: 'utf8' }
    )
    
    expect(result).toContain('Session complete')
  })
})
```

### **Hook Testing**

```bash
# Test individual hooks
echo '{"sessionId": "test"}' | ./scripts/session-complete.sh

# Test with mock data
echo '{"toolName": "exit_plan_mode", "toolInput": "{\"plan\": \"test\"}"}' | ./scripts/tool-complete-plan-capture.sh

# Test MCP server
echo '{"jsonrpc": "2.0", "method": "tools/list", "id": 1}' | node dist/index.js
```

## 🔄 Pull Request Process

### **Before Submitting**

1. **Self-Review Checklist**
   - [ ] Code follows style guidelines
   - [ ] Tests pass (`npm test`)
   - [ ] Build succeeds (`npm run build`)
   - [ ] Documentation updated
   - [ ] No merge conflicts
   - [ ] Commit messages are descriptive

2. **Testing Checklist**
   - [ ] Unit tests added/updated
   - [ ] Integration tests pass
   - [ ] Manual testing performed
   - [ ] Edge cases considered
   - [ ] Performance impact assessed

3. **Documentation Checklist**
   - [ ] Code comments added
   - [ ] API documentation updated
   - [ ] User documentation updated
   - [ ] Examples provided
   - [ ] Changelog updated

### **PR Template**

```markdown
## Description
Brief description of the changes made

## Type of Change
- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed
- [ ] All tests pass

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No merge conflicts
- [ ] Appropriate labels added

## Related Issues
Closes #123
Related to #456

## Screenshots (if applicable)
Add screenshots to demonstrate the changes

## Additional Notes
Any additional information or context
```

### **Review Process**

1. **Automated Checks**
   - CI/CD pipeline runs automatically
   - Code quality checks performed
   - Security scanning executed

2. **Manual Review**
   - Maintainer review required
   - Feedback provided via GitHub comments
   - Changes requested if needed

3. **Merge Process**
   - Squash and merge preferred
   - Descriptive commit message required
   - Branch deleted after merge

## 🎯 Contribution Areas

### **High Priority**
- Bug fixes and stability improvements
- Performance optimizations
- Documentation improvements
- Test coverage expansion

### **Medium Priority**
- New MCP tools and features
- Enhanced work intelligence
- Integration improvements
- User experience enhancements

### **Low Priority**
- Code refactoring
- Additional configuration options
- Developer experience improvements
- Experimental features

### **Specific Needs**
- **Windows Support** - Testing and fixes for Windows compatibility
- **Shell Script Optimization** - Performance improvements for large datasets
- **MCP Tool Extensions** - Additional tools for specific use cases
- **Documentation Translation** - Non-English documentation
- **Integration Examples** - Real-world usage examples

## 🐛 Issue Guidelines

### **Bug Reports**

Use this template:

```markdown
## Bug Description
Clear and concise description of the bug

## Steps to Reproduce
1. Step 1
2. Step 2
3. Step 3

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- OS: [e.g., macOS 14.0, Ubuntu 22.04]
- Node.js: [e.g., v18.17.0]
- npm: [e.g., v9.6.7]
- Claude Code: [version if known]
- Work Tracker: [version]

## Logs
```
[Paste relevant logs here]
```

## Additional Context
Any other relevant information, screenshots, etc.
```

### **Feature Requests**

Use this template:

```markdown
## Feature Request Summary
Brief description of the proposed feature

## Problem Statement
What problem does this feature solve?

## Proposed Solution
How should this feature work?

## Use Cases
- Use case 1
- Use case 2
- Use case 3

## Alternatives Considered
Other approaches you've considered

## Implementation Notes
Technical considerations or constraints

## Priority
- [ ] Critical - Blocking current work
- [ ] High - Significant improvement
- [ ] Medium - Nice to have
- [ ] Low - Future consideration
```

## 📚 Resources

### **Documentation**
- [Installation Guide](docs/installation.md)
- [API Reference](docs/api-reference.md)
- [Architecture](docs/architecture.md)
- [Developer Guide](docs/development.md)
- [Troubleshooting](docs/troubleshooting.md)

### **External Resources**
- [Model Context Protocol](https://github.com/modelcontextprotocol/specification)
- [Claude Code Documentation](https://docs.anthropic.com/claude-code)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [Jest Testing Framework](https://jestjs.io/docs/getting-started)

### **Community**
- [GitHub Issues](https://github.com/shawnroos/claude-work-tracker/issues)
- [GitHub Discussions](https://github.com/shawnroos/claude-work-tracker/discussions)
- [Claude Code Community](https://claude.ai/code)

## 🎖️ Recognition

Contributors will be recognized in:
- README.md contributor section
- CHANGELOG.md acknowledgments
- GitHub contributor statistics
- Special recognition for significant contributions

## 📄 License

By contributing to this project, you agree that your contributions will be licensed under the MIT License.

## 🙋 Questions?

If you have questions about contributing:
- Check the [Developer Guide](docs/development.md)
- Open a [GitHub Discussion](https://github.com/shawnroos/claude-work-tracker/discussions)
- Create an issue with the "question" label

Thank you for contributing to the Claude Code Work Tracking System!