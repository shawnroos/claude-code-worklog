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
    
    # Show active context
    if [ -f "$LOCAL_WORK_DIR/active/current-work-context.json" ]; then
        CURRENT_FOCUS=$(jq -r '.active_work.current_focus // "No current focus"' "$LOCAL_WORK_DIR/active/current-work-context.json" 2>/dev/null)
        ACTIVE_ITEMS=$(jq -r '.active_work.priority_items | length' "$LOCAL_WORK_DIR/active/current-work-context.json" 2>/dev/null || echo "0")
        echo "   Current focus: $CURRENT_FOCUS"
        echo "   Active items: $ACTIVE_ITEMS"
    fi
    
    # Show pending todos
    if [ -f "$LOCAL_WORK_DIR/PENDING_TODOS.json" ]; then
        PENDING_COUNT=$(jq length "$LOCAL_WORK_DIR/PENDING_TODOS.json" 2>/dev/null || echo "0")
        echo "   Pending todos: $PENDING_COUNT"
    fi
    
    # Show historical items count
    if [ -d "$LOCAL_WORK_DIR/history" ]; then
        HISTORY_COUNT=$(find "$LOCAL_WORK_DIR/history" -name "*.json" | wc -l | tr -d ' ')
        echo "   Historical items: $HISTORY_COUNT"
    fi
    
    # Show future work count
    if [ -d "$LOCAL_WORK_DIR/future" ]; then
        FUTURE_ITEMS=$(find "$LOCAL_WORK_DIR/future/items" -name "*.json" 2>/dev/null | wc -l | tr -d ' ')
        FUTURE_GROUPS=$(find "$LOCAL_WORK_DIR/future/groups" -name "*.json" 2>/dev/null | wc -l | tr -d ' ')
        echo "   Future work: $FUTURE_GROUPS groups, $FUTURE_ITEMS ungrouped items"
        if [ $((FUTURE_ITEMS + FUTURE_GROUPS)) -gt 0 ]; then
            echo "     Use list_future_groups() to view and groom_future_work() to organize"
        fi
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
echo "- ~/.claude/scripts/context-lifecycle.sh : Context management"
echo ""
echo "=== MCP Tools Available ==="
echo "Historical Tools:"
echo "- query_history         : Search historical work items"
echo "- get_historical_context: Get specific historical item"
echo "- summarize_period      : Generate period summary"
echo "- promote_to_active     : Bring historical item to active context"
echo ""
echo "Future Work Tools (Heuristic):"
echo "- defer_to_future       : Frictionless deferral during planning"
echo "- list_future_groups    : View groups and ungrouped items"
echo "- groom_future_work     : Analyze and reorganize with suggestions"
echo "- create_work_group     : Create logical groups"
echo "- promote_work_group    : Promote entire groups to active context"
echo ""