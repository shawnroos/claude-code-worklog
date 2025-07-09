#!/bin/bash

# Work Conflicts Detection
# Identifies potentially related work across worktrees

GLOBAL_STATE_DIR="$HOME/.claude/work-state"
CURRENT_DIR=$(pwd)
SEARCH_TERM="$1"

# Try to determine current project context
PROJECT_NAME=""
if git rev-parse --git-dir > /dev/null 2>&1; then
    # Find the main worktree to determine project name
    MAIN_WORKTREE=$(git worktree list | head -1 | awk '{print $1}')
    PROJECT_NAME=$(basename "$MAIN_WORKTREE")
fi

if [ -z "$PROJECT_NAME" ]; then
    PROJECT_NAME=$(basename "$CURRENT_DIR")
fi

echo "=== Work Conflicts Analysis ==="
echo "Project: $PROJECT_NAME"
echo ""

PROJECT_STATE_DIR="$GLOBAL_STATE_DIR/projects/$PROJECT_NAME"

if [ ! -d "$PROJECT_STATE_DIR" ]; then
    echo "No work state found for project $PROJECT_NAME"
    exit 0
fi

# Get current worktree name
CURRENT_WORKTREE="main"
if git rev-parse --git-dir > /dev/null 2>&1; then
    MAIN_WORKTREE=$(git worktree list | head -1 | awk '{print $1}')
    if [ "$(pwd)" != "$MAIN_WORKTREE" ]; then
        CURRENT_WORKTREE=$(basename "$(pwd)")
    fi
fi

echo "Current worktree: $CURRENT_WORKTREE"
echo ""

# Analyze todo content for potential conflicts
WORKTREE_STATE_DIR="$PROJECT_STATE_DIR/worktrees"

if [ ! -z "$SEARCH_TERM" ]; then
    echo "=== Searching for: '$SEARCH_TERM' ==="
    echo ""
fi

for worktree_file in "$WORKTREE_STATE_DIR"/*.json; do
    if [ -f "$worktree_file" ]; then
        WORKTREE=$(jq -r '.worktree' "$worktree_file")
        BRANCH=$(jq -r '.branch' "$worktree_file")
        
        # Skip current worktree
        if [ "$WORKTREE" = "$CURRENT_WORKTREE" ]; then
            continue
        fi
        
        # Check for todos containing search term or common keywords
        KEYWORDS="auth authentication login user component ui test testing bug fix"
        if [ ! -z "$SEARCH_TERM" ]; then
            KEYWORDS="$SEARCH_TERM"
        fi
        
        MATCHES=""
        for keyword in $KEYWORDS; do
            MATCHING_TODOS=$(jq -r --arg keyword "$keyword" '.pending_todos[] | select(.content | ascii_downcase | contains($keyword | ascii_downcase)) | "  🔄 " + .content + " (" + .status + ")"' "$worktree_file" 2>/dev/null)
            if [ ! -z "$MATCHING_TODOS" ]; then
                MATCHES="$MATCHES$MATCHING_TODOS\n"
            fi
        done
        
        if [ ! -z "$MATCHES" ]; then
            echo "### $WORKTREE ($BRANCH)"
            echo -e "$MATCHES"
            echo ""
        fi
    fi
done

if [ -z "$SEARCH_TERM" ]; then
    echo "💡 Use: work-conflicts <keyword> to search for specific topics"
fi