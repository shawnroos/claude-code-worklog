# Installation Guide

Complete guide for installing and setting up Claude Work Tracker.

## Prerequisites

- **Claude Code** installed and configured
- **Git** for version control
- **Node.js 18+** (for MCP server)
- **jq** for JSON processing
- **bash** shell (zsh compatible)

## Quick Start

### One-Line Installation

```bash
curl -sSL https://raw.githubusercontent.com/username/claude-work-tracker/main/install.sh | bash
```

This automatically:
- ✅ Installs all components
- ✅ Configures Claude Code integration
- ✅ Sets up MCP server
- ✅ Preserves existing settings

## Step-by-Step Installation

### 1. Clone Repository

```bash
git clone https://github.com/username/claude-work-tracker.git ~/claude-work-tracker
cd ~/claude-work-tracker
```

### 2. Install Dependencies

```bash
npm install
```

### 3. Build MCP Server

```bash
npm run build
```

### 4. Configure Claude Code

Add to `~/.claude/claude_code.json`:

```json
{
  "mcpServers": {
    "work-tracker": {
      "command": "node",
      "args": ["~/claude-work-tracker/dist/index.js"]
    }
  }
}
```

### 5. Verify Installation

```bash
# Test MCP server
npm test

# Check work tracking
/work status
```

## What Gets Installed

### Directory Structure

```
~/.claude/
├── claude_code.json          # MCP configuration
├── scripts/                  # Shell integrations
│   ├── work.sh              # /work command handler
│   └── session-init.sh      # Session restoration
└── CLAUDE.md                # Global instructions

~/claude-work-tracker/
├── dist/                    # Compiled MCP server
├── src/                     # Source code
└── docs/                    # Documentation

project/.claude-work/        # Per-project storage
├── active/                  # Current work items
├── history/                 # Archived work
└── future/                  # Deferred work
```

### Components

1. **MCP Server** - API for work tracking
2. **Smart References** - AI-powered linking
3. **Shell Scripts** - Command integration
4. **Storage Layer** - Local data persistence

## Configuration

### Basic Settings

Create `~/.claude/work-tracker.json`:

```json
{
  "presentation": {
    "mode": "summary",        // quiet | summary | verbose
    "emoji_style": "modern"   // modern | minimal | classic
  },
  "features": {
    "auto_references": true,
    "smart_suggestions": true,
    "conflict_detection": true
  },
  "storage": {
    "history_days": 90,
    "compression": true
  }
}
```

### Advanced Configuration

#### Custom Storage Location

```bash
export CLAUDE_WORK_DIR="/custom/path"
```

#### Reference Sensitivity

```json
{
  "smart_references": {
    "similarity_threshold": 0.7,
    "confidence_threshold": 0.6,
    "max_suggestions": 10
  }
}
```

#### Git Integration

```json
{
  "git": {
    "track_branches": true,
    "associate_commits": false,
    "worktree_support": true
  }
}
```

## Platform-Specific Instructions

### macOS

```bash
# Install prerequisites
brew install node jq

# Install work tracker
curl -sSL https://raw.githubusercontent.com/username/claude-work-tracker/main/install.sh | bash
```

### Linux

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install nodejs npm jq

# Install work tracker
curl -sSL https://raw.githubusercontent.com/username/claude-work-tracker/main/install.sh | bash
```

### Windows (WSL)

```bash
# In WSL terminal
sudo apt-get update
sudo apt-get install nodejs npm jq

# Install work tracker
curl -sSL https://raw.githubusercontent.com/username/claude-work-tracker/main/install.sh | bash
```

## Troubleshooting

### Common Issues

#### "Command not found: /work"

```bash
# Add to ~/.zshrc or ~/.bashrc
source ~/.claude/scripts/work.sh
```

#### "MCP server not connecting"

```bash
# Check server status
ps aux | grep claude-work-tracker

# Restart Claude Code
# Then verify MCP configuration
cat ~/.claude/claude_code.json
```

#### "Permission denied"

```bash
# Fix script permissions
chmod +x ~/.claude/scripts/*.sh

# Fix storage permissions
chmod -R 755 ~/.claude-work/
```

### Verification Steps

```bash
# 1. Check installation
ls -la ~/claude-work-tracker/dist/

# 2. Test MCP server
node ~/claude-work-tracker/dist/index.js

# 3. Verify work command
/work help

# 4. Check storage
ls -la .claude-work/
```

## Updating

### Automatic Update

```bash
cd ~/claude-work-tracker
npm run update
```

### Manual Update

```bash
cd ~/claude-work-tracker
git pull origin main
npm install
npm run build
```

## Uninstalling

### Complete Removal

```bash
# Run uninstall script
~/claude-work-tracker/scripts/uninstall.sh

# Or manually
rm -rf ~/claude-work-tracker
rm -rf ~/.claude/scripts/work*
rm -rf .claude-work/  # In each project
```

### Preserve Data

```bash
# Backup before uninstalling
tar -czf claude-work-backup.tar.gz ~/.claude-work/

# Remove only executables
rm -rf ~/claude-work-tracker/dist/
rm -rf ~/claude-work-tracker/node_modules/
```

## Security Considerations

### Data Privacy
- All data stored locally
- No network transmission
- Git-ignored by default

### File Permissions
```bash
# Secure storage
chmod 700 ~/.claude-work/
chmod 600 ~/.claude-work/**/*.json
```

### Sensitive Data
Add to `.gitignore`:
```
.claude-work/
*.secret
*-private.json
```

## Next Steps

1. **Read the [User Guide](user-guide.md)** to learn usage
2. **Check [API Reference](api-reference.md)** for all tools
3. **Explore [Smart References](smart-references.md)** features
4. **Join the community** for tips and updates

## Getting Help

- **Documentation**: Read the guides in `/docs`
- **Issues**: GitHub issue tracker
- **Community**: Discord server (coming soon)
- **Email**: support@claudeworktracker.dev