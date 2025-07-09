#!/bin/bash

# Global Work State Aggregation Script
# Updates global project overview from local worktree states

PROJECT_DIR="$1"
WORKTREE_NAME="$2"
BRANCH_NAME="$3"

if [ -z "$PROJECT_DIR" ] || [ -z "$WORKTREE_NAME" ] || [ -z "$BRANCH_NAME" ]; then
    echo "Usage: $0 <project_dir> <worktree_name> <branch_name>"
    exit 1
fi

# Determine project name from directory
PROJECT_NAME=$(basename "$PROJECT_DIR")
GLOBAL_STATE_DIR="$HOME/.claude/work-state"
PROJECT_STATE_DIR="$GLOBAL_STATE_DIR/projects/$PROJECT_NAME"
WORKTREE_STATE_DIR="$PROJECT_STATE_DIR/worktrees"

# Create project directories
mkdir -p "$WORKTREE_STATE_DIR"

# Update individual worktree state
WORKTREE_STATE_FILE="$WORKTREE_STATE_DIR/$WORKTREE_NAME.json"
LOCAL_WORK_DIR="$PROJECT_DIR/.claude-work"

if [ -d "$LOCAL_WORK_DIR" ]; then
    # Create worktree state snapshot
    cat > "$WORKTREE_STATE_FILE" << EOF
{
  "worktree": "$WORKTREE_NAME",
  "branch": "$BRANCH_NAME",
  "project_dir": "$PROJECT_DIR",
  "last_updated": "$(date -Iseconds)",
  "pending_todos": $(cat "$LOCAL_WORK_DIR/PENDING_TODOS.json" 2>/dev/null || echo "[]"),
  "last_session_work": "$(tail -20 "$LOCAL_WORK_DIR/WORK_HISTORY.md" 2>/dev/null | base64 -w 0 || echo "")"
}
EOF
fi

# Aggregate all worktrees for this project
PROJECT_OVERVIEW="$PROJECT_STATE_DIR/ACTIVE_WORK.md"

cat > "$PROJECT_OVERVIEW" << EOF
# $PROJECT_NAME - Active Work Overview

*Last updated: $(date)*

## Worktree Status

EOF

# Add each worktree's status
for worktree_file in "$WORKTREE_STATE_DIR"/*.json; do
    if [ -f "$worktree_file" ]; then
        WORKTREE=$(jq -r '.worktree' "$worktree_file")
        BRANCH=$(jq -r '.branch' "$worktree_file")
        UPDATED=$(jq -r '.last_updated' "$worktree_file")
        PENDING_COUNT=$(jq -r '.pending_todos | length' "$worktree_file")
        
        echo "### $WORKTREE ($BRANCH)" >> "$PROJECT_OVERVIEW"
        echo "- **Last updated:** $UPDATED" >> "$PROJECT_OVERVIEW"
        echo "- **Pending todos:** $PENDING_COUNT" >> "$PROJECT_OVERVIEW"
        
        if [ "$PENDING_COUNT" -gt 0 ]; then
            echo "" >> "$PROJECT_OVERVIEW"
            jq -r '.pending_todos[] | "  - 🔄 " + .content + " (" + .status + ")"' "$worktree_file" >> "$PROJECT_OVERVIEW"
        fi
        echo "" >> "$PROJECT_OVERVIEW"
    fi
done

# Update global project overview
GLOBAL_OVERVIEW="$GLOBAL_STATE_DIR/PROJECT_OVERVIEW.md"

# Create or update global overview header
if [ ! -f "$GLOBAL_OVERVIEW" ]; then
    cat > "$GLOBAL_OVERVIEW" << EOF
# Global Work Overview

*This file aggregates active work across all projects and worktrees*

EOF
fi

# Update project section in global overview
PROJECT_SECTION_START="## $PROJECT_NAME"
PROJECT_SECTION_END="## "

# Remove existing project section and add updated one
if grep -q "^$PROJECT_SECTION_START" "$GLOBAL_OVERVIEW"; then
    # Remove old section
    sed -i '' "/^$PROJECT_SECTION_START/,/^$PROJECT_SECTION_END/d" "$GLOBAL_OVERVIEW"
fi

# Add updated project section
{
    echo "$PROJECT_SECTION_START"
    echo ""
    echo "**Total worktrees:** $(find "$WORKTREE_STATE_DIR" -name "*.json" | wc -l | tr -d ' ')"
    echo "**Total pending todos:** $(jq -s 'map(.pending_todos | length) | add' "$WORKTREE_STATE_DIR"/*.json 2>/dev/null || echo "0")"
    echo ""
    echo "[View detailed breakdown]($PROJECT_STATE_DIR/ACTIVE_WORK.md)"
    echo ""
} >> "$GLOBAL_OVERVIEW"

echo "[$(date)] Global state updated for $PROJECT_NAME/$WORKTREE_NAME" >> "$HOME/.claude/hooks.log"