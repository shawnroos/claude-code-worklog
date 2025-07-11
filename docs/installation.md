# Installation Guide

This guide covers all installation methods and setup options for the Claude Code Work Tracking System.

## 🚀 Quick Installation

### One-Line Installation (Recommended)

```bash
curl -sSL https://raw.githubusercontent.com/shawnroos/claude-work-tracker/main/install.sh | bash
```

This will:
- ✅ Install all required scripts and configurations
- ✅ Backup existing settings safely
- ✅ Merge with your existing Claude Code setup
- ✅ Set up all necessary directories and permissions

### Local Installation

If you prefer to download and inspect before installing:

```bash
# Clone the repository
git clone https://github.com/shawnroos/claude-work-tracker.git
cd claude-work-tracker

# Run the installer
./install.sh
```

## 🔧 Post-Installation Setup

### 1. **Verify Installation**

```bash
# Test the presentation system
~/.claude/scripts/work-presentation.sh test

# Check work status
~/.claude/scripts/work-status.sh

# Try the /work command
/work
```

### 2. **Build MCP Server** (Optional)

```bash
# Install dependencies
npm install

# Build TypeScript
npm run build

# Test the server
npm start
```

### 3. **Configure MCP Integration** (Optional)

Add to your Claude Code MCP configuration:

```json
{
  "mcpServers": {
    "claude-work-tracker": {
      "command": "node",
      "args": ["/absolute/path/to/claude-work-tracker/dist/index.js"],
      "env": {}
    }
  }
}
```

### 4. **Run Setup Wizard** (Optional)

```bash
~/.claude/scripts/setup-wizard.sh
```

Customize:
- Presentation modes (quiet, summary, verbose)
- Emoji styles (minimal_colored, modern, classic, minimal)
- Notification preferences
- Auto-restore settings

## 📁 Installation Components

The installer creates and configures:

### Global Configuration (`~/.claude/`)
```
~/.claude/
├── CLAUDE.md                           # Global coding standards
├── settings.local.json                 # Claude Code hooks
├── work-tracking-config.json           # Presentation config
├── scripts/                            # All automation scripts
├── work-state/                         # Global work aggregation
├── work-intelligence/                  # Plans and proposals
├── projects/                           # Session logs
├── todos/                              # Per-session todos
└── findings/                           # Tool-captured findings
```

### Project-Level Structure (`.claude-work/`)
```
your-project/
└── .claude-work/
    ├── WORK_HISTORY.md                 # Work history log
    ├── PENDING_TODOS.json              # Incomplete todos
    └── current_todos.json              # Current session backup
```

### Installed Scripts
- `session-init.sh` - Session start hook
- `session-complete.sh` - Session end hook
- `tool-complete-plan-capture.sh` - Plan capture hook
- `work.sh` - Manual `/work` command
- `work-*.sh` - Cross-worktree utilities
- `update-work-intelligence.sh` - Intelligence aggregation

## ⚙️ Configuration Options

### Settings Files

**`~/.claude/settings.local.json`** - Hook configuration
```json
{
  "hooks": {
    "session_start": "~/.claude/scripts/session-init.sh",
    "session_complete": "~/.claude/scripts/session-complete.sh",
    "tool_complete": "~/.claude/scripts/tool-complete-plan-capture.sh"
  },
  "commands": {
    "work": "~/.claude/scripts/work.sh"
  }
}
```

**`~/.claude/work-tracking-config.json`** - Presentation settings
```json
{
  "presentation": {
    "mode": "summary",
    "show_notifications": true,
    "emoji_style": "minimal_colored"
  },
  "session": {
    "auto_restore": true,
    "auto_create_todos": false
  }
}
```

### Environment Variables

Optional environment variables for customization:

```bash
# Override default directories
export CLAUDE_WORK_DIR="$HOME/.claude"
export CLAUDE_WORK_STATE_DIR="$HOME/.claude/work-state"

# MCP server configuration
export CLAUDE_MCP_SERVER_PORT=3000
export CLAUDE_MCP_SERVER_HOST="localhost"
```

## 🔄 Updating

### Update from GitHub
```bash
curl -sSL https://raw.githubusercontent.com/shawnroos/claude-work-tracker/main/install.sh | bash
```

The installer will:
- Detect existing installation
- Backup current configuration
- Update scripts while preserving your settings
- Migrate any data format changes

### Manual Update
```bash
cd claude-work-tracker
git pull origin main
./install.sh
```

## 🧹 Uninstalling

### Complete Removal
```bash
~/.claude/uninstall.sh
```

This will:
- ✅ Remove all scripts and configurations
- ✅ Create backup of all data before removal
- ✅ Preserve work history and conversation logs
- ✅ Restore original Claude Code settings

### Selective Removal
```bash
# Remove only MCP server
rm -rf node_modules/ dist/ src/

# Remove only hooks (keep data)
rm ~/.claude/scripts/session-*.sh
rm ~/.claude/scripts/tool-complete-*.sh

# Remove only work intelligence (keep todos)
rm -rf ~/.claude/work-intelligence/
```

## 🔧 Troubleshooting Installation

### Common Issues

**Permission Errors:**
```bash
# Fix script permissions
chmod +x ~/.claude/scripts/*.sh

# Fix directory permissions
chmod 755 ~/.claude/
```

**Missing Dependencies:**
```bash
# Install jq (required for JSON processing)
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# Install Node.js (for MCP server)
# macOS
brew install node

# Ubuntu/Debian
sudo apt-get install nodejs npm
```

**Hook Not Working:**
```bash
# Check Claude Code settings
cat ~/.claude/settings.local.json

# Test hook manually
echo '{"sessionId": "test"}' | ~/.claude/scripts/session-complete.sh
```

### Verification Commands

```bash
# Check installation completeness
ls -la ~/.claude/scripts/

# Verify configuration
~/.claude/scripts/work-presentation.sh test

# Test work commands
/work status

# Test MCP server (if installed)
npm test
```

## 📚 Next Steps

After installation:

1. **Read the [Configuration Guide](configuration.md)** - Customize your setup
2. **Try the [API Reference](api-reference.md)** - Explore MCP server tools
3. **Check [Troubleshooting](troubleshooting.md)** - If you encounter issues
4. **Review [Architecture](architecture.md)** - Understand how it works

## 🤝 Getting Help

If you encounter issues:

1. Check the [Troubleshooting Guide](troubleshooting.md)
2. Review the installation logs: `~/.claude/hooks.log`
3. Open an issue: https://github.com/shawnroos/claude-work-tracker/issues
4. Include your system information and error messages