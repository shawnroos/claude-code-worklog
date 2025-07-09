#!/bin/bash

# Claude Work Tracker Uninstall Script
# Safely removes the work tracking system while preserving data

set -e

INSTALL_DIR="$HOME/.claude"
BACKUP_DIR="$HOME/.claude-work-tracker-backup-$(date +%Y%m%d_%H%M%S)"

echo "🗑️  Uninstalling Claude Work Tracker..."

# Create backup
if [ -d "$INSTALL_DIR" ]; then
    echo "📦 Creating backup at: $BACKUP_DIR"
    cp -r "$INSTALL_DIR" "$BACKUP_DIR"
fi

# Remove scripts
if [ -d "$INSTALL_DIR/scripts" ]; then
    rm -rf "$INSTALL_DIR/scripts"
    echo "✅ Removed automation scripts"
fi

# Remove configuration
if [ -f "$INSTALL_DIR/work-tracking-config.json" ]; then
    rm "$INSTALL_DIR/work-tracking-config.json"
    echo "✅ Removed work tracking configuration"
fi

# Clean up hooks from settings.local.json
if [ -f "$INSTALL_DIR/settings.local.json" ]; then
    # Remove work tracking hooks
    jq 'del(.hooks.session_complete, .hooks.tool_complete)' "$INSTALL_DIR/settings.local.json" > "$INSTALL_DIR/settings.local.json.tmp"
    mv "$INSTALL_DIR/settings.local.json.tmp" "$INSTALL_DIR/settings.local.json"
    echo "✅ Cleaned hooks from settings.local.json"
fi

# Note: We preserve work-state, projects, and todos directories as they contain user data

echo ""
echo "✅ Claude Work Tracker uninstalled successfully!"
echo ""
echo "Preserved data:"
echo "  📁 $INSTALL_DIR/work-state/    - Work history"
echo "  📁 $INSTALL_DIR/projects/      - Session logs"
echo "  📁 $INSTALL_DIR/todos/         - Todo history"
echo ""
echo "Full backup created at:"
echo "  📦 $BACKUP_DIR"
echo ""
echo "To reinstall: run ./install.sh from the claude-work-tracker directory"
echo ""