#!/bin/bash

# Local Work Status Command
# Shows work status for current project only

CURRENT_DIR=$(pwd)
PROJECT_NAME=$(basename "$CURRENT_DIR")

echo "=== Local Work Status: $PROJECT_NAME ==="
echo ""

# Check if we're in a git repository
if git rev-parse --git-dir > /dev/null 2>&1; then
    BRANCH=$(git branch --show-current 2>/dev/null || echo "unknown")
    echo "📍 Current branch: $BRANCH"
    echo "📁 Working directory: $CURRENT_DIR"
    echo ""
fi

# Check for local work state
LOCAL_WORK_DIR="$CURRENT_DIR/.claude-work"
if [ -d "$LOCAL_WORK_DIR" ]; then
    echo "📋 Local Work State:"
    
    # Show pending todos
    if [ -f "$LOCAL_WORK_DIR/PENDING_TODOS.json" ]; then
        PENDING_COUNT=$(jq length "$LOCAL_WORK_DIR/PENDING_TODOS.json" 2>/dev/null || echo "0")
        echo "   Pending todos: $PENDING_COUNT"
    fi
    
    # Show recent work
    if [ -f "$LOCAL_WORK_DIR/WORK_HISTORY.md" ]; then
        echo "   Recent work:"
        tail -5 "$LOCAL_WORK_DIR/WORK_HISTORY.md" | grep -E "^(##|✅|🔄)" | head -3
    fi
else
    echo "📋 No local work state found"
    echo "   Work state will be created when you use the system"
fi

echo ""
echo "=== Available Commands ==="
echo "- work-status           : This local overview"
echo "- ~/.claude/scripts/work.sh : Manual work commands"
echo ""