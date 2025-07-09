#!/bin/bash

# Claude Work Tracker Installation Script
# Installs the automated work tracking system for Claude Code

set -e

INSTALL_DIR="$HOME/.claude"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🚀 Installing Claude Work Tracker..."

# Create directories
mkdir -p "$INSTALL_DIR"/{scripts,work-state,projects,todos}

# Backup existing files
if [ -f "$INSTALL_DIR/settings.local.json" ]; then
    cp "$INSTALL_DIR/settings.local.json" "$INSTALL_DIR/settings.local.json.backup.$(date +%Y%m%d_%H%M%S)"
    echo "✅ Backed up existing settings.local.json"
fi

if [ -f "$INSTALL_DIR/CLAUDE.md" ]; then
    cp "$INSTALL_DIR/CLAUDE.md" "$INSTALL_DIR/CLAUDE.md.backup.$(date +%Y%m%d_%H%M%S)"
    echo "✅ Backed up existing CLAUDE.md"
fi

# Copy scripts
cp -r "$SCRIPT_DIR/scripts/"* "$INSTALL_DIR/scripts/"
chmod +x "$INSTALL_DIR/scripts/"*.sh
echo "✅ Installed automation scripts"

# Copy configuration files
cp "$SCRIPT_DIR/examples/work-tracking-config.json" "$INSTALL_DIR/"
echo "✅ Installed configuration"

# Merge settings.local.json
if [ -f "$INSTALL_DIR/settings.local.json" ]; then
    # Merge with existing settings
    echo "🔧 Merging with existing settings..."
    jq -s '.[0] * .[1]' "$INSTALL_DIR/settings.local.json" "$SCRIPT_DIR/examples/settings.local.json" > "$INSTALL_DIR/settings.local.json.tmp"
    mv "$INSTALL_DIR/settings.local.json.tmp" "$INSTALL_DIR/settings.local.json"
else
    cp "$SCRIPT_DIR/examples/settings.local.json" "$INSTALL_DIR/"
fi
echo "✅ Configured Claude Code hooks"

# Update CLAUDE.md
if [ -f "$INSTALL_DIR/CLAUDE.md" ]; then
    echo "" >> "$INSTALL_DIR/CLAUDE.md"
    echo "# Work Tracking System" >> "$INSTALL_DIR/CLAUDE.md"
    echo "" >> "$INSTALL_DIR/CLAUDE.md"
    cat "$SCRIPT_DIR/examples/CLAUDE.md" | grep -A 1000 "## Session Initialization" >> "$INSTALL_DIR/CLAUDE.md"
else
    cp "$SCRIPT_DIR/examples/CLAUDE.md" "$INSTALL_DIR/CLAUDE.md"
fi
echo "✅ Updated CLAUDE.md configuration"

echo ""
echo "🎉 Claude Work Tracker installed successfully!"
echo ""
echo "Next steps:"
echo "  1. Run: ~/.claude/scripts/setup-wizard.sh (optional customization)"
echo "  2. Test: ~/.claude/scripts/work-presentation.sh test"
echo "  3. Start using Claude Code - the system works automatically!"
echo ""
echo "Commands:"
echo "  ~/.claude/scripts/save.sh              - Manual work save"
echo "  ~/.claude/scripts/work-status.sh       - Global work overview"
echo "  ~/.claude/scripts/work-conflicts.sh    - Find related work"
echo ""