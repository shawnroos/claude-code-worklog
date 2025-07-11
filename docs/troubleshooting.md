# Troubleshooting Guide

Common issues and solutions for the Claude Code Work Tracking System.

## 🔍 Quick Diagnostics

### System Health Check

```bash
# Check if work tracking is installed
ls ~/.claude/scripts/work.sh

# Test presentation system
~/.claude/scripts/work-presentation.sh test

# Check work status
~/.claude/scripts/work-status.sh

# Test MCP server (if installed)
npm test
```

### Log Files

```bash
# Check hook execution logs
tail -f ~/.claude/hooks.log

# Check npm/build logs
npm run build 2>&1 | tee build.log

# Check server startup logs
npm start 2>&1 | tee server.log
```

## 🚨 Common Issues

### Installation Problems

#### **Issue: Permission Denied**
```
bash: ./install.sh: Permission denied
```

**Solution:**
```bash
chmod +x install.sh
./install.sh
```

#### **Issue: Missing Dependencies**
```
Error: jq command not found
```

**Solution:**
```bash
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# CentOS/RHEL
sudo yum install jq
```

#### **Issue: Node.js Not Found**
```
Error: node command not found
```

**Solution:**
```bash
# macOS
brew install node

# Ubuntu/Debian
sudo apt-get install nodejs npm

# Or use Node Version Manager
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
nvm install node
```

### Hook System Issues

#### **Issue: Hooks Not Executing**
```
Hooks don't seem to be working
```

**Diagnosis:**
```bash
# Check hook configuration
cat ~/.claude/settings.local.json

# Test hook manually
echo '{"sessionId": "test", "workingDirectory": "$(pwd)"}' | ~/.claude/scripts/session-complete.sh
```

**Solution:**
```bash
# Fix hook permissions
chmod +x ~/.claude/scripts/*.sh

# Verify hook paths in settings
~/.claude/scripts/setup-wizard.sh
```

#### **Issue: Hook Errors in Log**
```
[2024-01-11] Error in session-complete.sh: jq: command not found
```

**Solution:**
```bash
# Install missing dependencies
brew install jq  # macOS
# or
sudo apt-get install jq  # Ubuntu/Debian

# Check PATH in hook environment
which jq
```

### MCP Server Issues

#### **Issue: Server Won't Start**
```
Error: Cannot find module '@modelcontextprotocol/sdk'
```

**Solution:**
```bash
# Install dependencies
npm install

# Rebuild if needed
npm run build
```

#### **Issue: Build Errors**
```
error TS2307: Cannot find module '@modelcontextprotocol/sdk/types.js'
```

**Solution:**
```bash
# Clean and reinstall
rm -rf node_modules package-lock.json
npm install
npm run build
```

#### **Issue: Server Connection Issues**
```
MCP server not responding
```

**Diagnosis:**
```bash
# Test server directly
echo '{"jsonrpc": "2.0", "method": "tools/list", "id": 1}' | node dist/index.js

# Check if server is running
ps aux | grep node
```

**Solution:**
```bash
# Restart server
npm start

# Check Claude Code MCP configuration
cat ~/.claude/settings.local.json
```

### Work State Issues

#### **Issue: Todos Not Persisting**
```
Todos disappear between sessions
```

**Diagnosis:**
```bash
# Check if todos are being saved
ls ~/.claude/todos/

# Check work state directory
ls .claude-work/

# Test todo hook
echo '{"sessionId": "test", "workingDirectory": "$(pwd)"}' | ~/.claude/scripts/tool-complete-enhanced.sh
```

**Solution:**
```bash
# Fix directory permissions
chmod 755 ~/.claude/todos
chmod 755 .claude-work

# Verify hook is installed
grep -r "tool_complete" ~/.claude/settings.local.json
```

#### **Issue: Work Intelligence Not Captured**
```
Plans and proposals not being saved
```

**Diagnosis:**
```bash
# Check work intelligence directory
ls ~/.claude/work-intelligence/

# Test plan capture hook
echo '{"toolName": "exit_plan_mode", "sessionId": "test", "toolInput": "{\"plan\": \"test plan\"}"}' | ~/.claude/scripts/tool-complete-plan-capture.sh
```

**Solution:**
```bash
# Install enhanced hook
cp scripts/tool-complete-plan-capture.sh ~/.claude/scripts/
chmod +x ~/.claude/scripts/tool-complete-plan-capture.sh

# Update hook configuration
# Edit ~/.claude/settings.local.json to use tool-complete-plan-capture.sh
```

### Git Integration Issues

#### **Issue: Git Context Not Detected**
```
Git branch shows as "unknown"
```

**Diagnosis:**
```bash
# Check if in git repository
git rev-parse --is-inside-work-tree

# Check git commands
git branch --show-current
git worktree list
```

**Solution:**
```bash
# Initialize git if needed
git init

# Or run from within git repository
cd /path/to/git/repo
~/.claude/scripts/work-status.sh
```

#### **Issue: Worktree Detection Problems**
```
Worktree shows as "main" when in feature branch
```

**Diagnosis:**
```bash
# Check worktree list
git worktree list --porcelain

# Check current directory
pwd
```

**Solution:**
```bash
# Ensure you're in the correct worktree
cd /path/to/feature/worktree

# Or fix worktree detection in scripts
# This is usually a path resolution issue
```

### Cross-Worktree Issues

#### **Issue: Conflicts Not Detected**
```
Cross-worktree conflicts not showing
```

**Diagnosis:**
```bash
# Check global state
ls ~/.claude/work-state/projects/

# Test conflict detection
~/.claude/scripts/work-conflicts.sh authentication
```

**Solution:**
```bash
# Update global state manually
~/.claude/scripts/update-global-state.sh $(pwd) $(basename $(pwd)) $(git branch --show-current)

# Check project naming consistency
basename $(pwd)
```

## 🔧 Advanced Troubleshooting

### Debug Mode

Enable debug mode for detailed logging:

```bash
# Enable debug in hooks
export CLAUDE_WORK_DEBUG=1

# Run commands with debug output
CLAUDE_WORK_DEBUG=1 ~/.claude/scripts/work-status.sh
```

### Manual State Inspection

```bash
# Check todo file format
cat ~/.claude/todos/latest-session.json | jq .

# Check work intelligence format
cat ~/.claude/work-intelligence/latest-intelligence.json | jq .

# Check global state
cat ~/.claude/work-state/PROJECT_OVERVIEW.md
```

### Reset Components

```bash
# Reset all work state (preserves history)
rm -rf ~/.claude/work-state
mkdir -p ~/.claude/work-state

# Reset MCP server
rm -rf node_modules dist
npm install
npm run build

# Reset hooks (keep data)
rm ~/.claude/scripts/session-*.sh
rm ~/.claude/scripts/tool-complete-*.sh
./install.sh
```

## 📊 Performance Issues

### Large Data Sets

#### **Issue: Slow Work Status Commands**
```
work-status.sh takes too long to run
```

**Solution:**
```bash
# Limit search scope
find ~/.claude/work-intelligence -name "*.json" -mtime -7  # Last 7 days only

# Archive old data
mkdir ~/.claude/archive
mv ~/.claude/work-intelligence/2023* ~/.claude/archive/
```

#### **Issue: Large Log Files**
```
hooks.log growing too large
```

**Solution:**
```bash
# Rotate log files
mv ~/.claude/hooks.log ~/.claude/hooks.log.old
touch ~/.claude/hooks.log

# Or set up logrotate
cat > ~/.claude/logrotate.conf << 'EOF'
~/.claude/hooks.log {
    daily
    rotate 7
    compress
    missingok
    notifempty
}
EOF
```

### Memory Issues

#### **Issue: MCP Server Memory Usage**
```
High memory usage from Node.js process
```

**Solution:**
```bash
# Limit Node.js heap size
node --max-old-space-size=512 dist/index.js

# Or add to package.json
"start": "node --max-old-space-size=512 dist/index.js"
```

## 🔍 Diagnostic Commands

### System Information

```bash
# Check system details
uname -a
node --version
npm --version
jq --version

# Check directory sizes
du -sh ~/.claude/
du -sh ~/.claude/work-intelligence/
du -sh ~/.claude/todos/
```

### Configuration Validation

```bash
# Validate JSON configuration
jq . ~/.claude/settings.local.json
jq . ~/.claude/work-tracking-config.json

# Check hook permissions
ls -la ~/.claude/scripts/

# Test all hooks
for hook in ~/.claude/scripts/session-*.sh; do
    echo "Testing $hook"
    echo '{"sessionId": "test"}' | "$hook"
done
```

## 🆘 Getting Help

### Before Asking for Help

1. **Check this troubleshooting guide**
2. **Review the logs**: `~/.claude/hooks.log`
3. **Test basic functionality**: `~/.claude/scripts/work-presentation.sh test`
4. **Try a clean install**: Backup data, uninstall, reinstall

### Information to Include

When reporting issues, include:

```bash
# System information
uname -a
node --version
npm --version
jq --version

# Configuration
cat ~/.claude/settings.local.json
cat ~/.claude/work-tracking-config.json

# Recent logs
tail -50 ~/.claude/hooks.log

# Directory structure
ls -la ~/.claude/
ls -la ~/.claude/scripts/
```

### Where to Get Help

1. **GitHub Issues**: https://github.com/shawnroos/claude-work-tracker/issues
2. **Documentation**: Check other docs in this repository
3. **Community**: Claude Code community forums

### Creating a Good Bug Report

```markdown
## Bug Description
Brief description of the issue

## Steps to Reproduce
1. Step one
2. Step two
3. Step three

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- OS: [e.g., macOS 14.0]
- Node.js: [e.g., v18.17.0]
- Claude Code: [version if known]

## Logs
```
[Include relevant log excerpts]
```

## Additional Context
Any other relevant information
```

## 🔄 Recovery Procedures

### Complete System Reset

```bash
# Backup important data
cp -r ~/.claude/work-state ~/work-state-backup
cp -r ~/.claude/todos ~/todos-backup

# Uninstall completely
~/.claude/uninstall.sh

# Clean reinstall
curl -sSL https://raw.githubusercontent.com/shawnroos/claude-work-tracker/main/install.sh | bash

# Restore data
cp -r ~/work-state-backup ~/.claude/work-state
cp -r ~/todos-backup ~/.claude/todos
```

### Partial Recovery

```bash
# Restore just the hooks
./install.sh --hooks-only

# Restore just the MCP server
npm install
npm run build

# Restore just the configuration
cp examples/settings.local.json ~/.claude/
cp examples/work-tracking-config.json ~/.claude/
```

This comprehensive troubleshooting guide should help resolve most common issues with the Claude Code Work Tracking System.