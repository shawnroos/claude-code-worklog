#!/bin/bash

# Full Project Context Loader
# WARNING: Higher token usage - loads all worktree contexts

GLOBAL_STATE_DIR="$HOME/.claude/work-state"
PROJECT_NAME="${1:-$(basename $(pwd))}"

PROJECT_STATE_DIR="$GLOBAL_STATE_DIR/projects/$PROJECT_NAME"

echo "=== Full Project Context: $PROJECT_NAME ==="
echo "⚠️  WARNING: This loads full context from all worktrees (higher token usage)"
echo ""

if [ ! -d "$PROJECT_STATE_DIR" ]; then
    echo "No work state found for project $PROJECT_NAME"
    exit 0
fi

# Show detailed project overview
PROJECT_OVERVIEW="$PROJECT_STATE_DIR/ACTIVE_WORK.md"
if [ -f "$PROJECT_OVERVIEW" ]; then
    cat "$PROJECT_OVERVIEW"
fi

echo ""
echo "=== Detailed Worktree States ==="
echo ""

WORKTREE_STATE_DIR="$PROJECT_STATE_DIR/worktrees"

for worktree_file in "$WORKTREE_STATE_DIR"/*.json; do
    if [ -f "$worktree_file" ]; then
        WORKTREE=$(jq -r '.worktree' "$worktree_file")
        BRANCH=$(jq -r '.branch' "$worktree_file")
        PROJECT_DIR=$(jq -r '.project_dir' "$worktree_file")
        UPDATED=$(jq -r '.last_updated' "$worktree_file")
        
        echo "### $WORKTREE ($BRANCH)"
        echo "**Directory:** $PROJECT_DIR"
        echo "**Last Updated:** $UPDATED"
        echo ""
        
        # Show pending todos
        PENDING_COUNT=$(jq -r '.pending_todos | length' "$worktree_file")
        if [ "$PENDING_COUNT" -gt 0 ]; then
            echo "**Pending Todos:**"
            jq -r '.pending_todos[] | "- 🔄 " + .content + " (" + .status + ", " + .priority + " priority)"' "$worktree_file"
        else
            echo "**Status:** ✅ No pending todos"
        fi
        
        echo ""
        
        # Show recent work history (if available)
        WORK_HISTORY=$(jq -r '.last_session_work' "$worktree_file")
        if [ ! -z "$WORK_HISTORY" ] && [ "$WORK_HISTORY" != "null" ] && [ "$WORK_HISTORY" != "" ]; then
            echo "**Recent Work:**"
            echo "$WORK_HISTORY" | base64 -d | tail -10
            echo ""
        fi
        
        echo "---"
        echo ""
    fi
done

echo ""
echo "💡 This full context is now available in your conversation for cross-worktree analysis"