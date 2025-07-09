#!/bin/bash

# Cross-Worktree Work Status Command
# Shows active work across all worktrees

GLOBAL_STATE_DIR="$HOME/.claude/work-state"
CURRENT_DIR=$(pwd)

echo "=== Global Work Status ==="
echo ""

if [ ! -d "$GLOBAL_STATE_DIR" ]; then
    echo "No global work state found. Work in a project to initialize."
    exit 0
fi

# Show global overview if it exists
GLOBAL_OVERVIEW="$GLOBAL_STATE_DIR/PROJECT_OVERVIEW.md"
if [ -f "$GLOBAL_OVERVIEW" ]; then
    cat "$GLOBAL_OVERVIEW"
else
    echo "No projects tracked yet."
fi

echo ""
echo "=== Commands ==="
echo "- work-status           : This overview"
echo "- work-conflicts        : Check for related work in other worktrees"
echo "- work-context <project>: Load full project context"
echo ""