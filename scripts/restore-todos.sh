#!/bin/bash

# Restore Incomplete Todos Helper Script
# Call this when starting work in a project directory

WORKING_DIR=${1:-$(pwd)}
WORK_STATE_DIR="$WORKING_DIR/.claude-work"

# Get current git context
CURRENT_BRANCH=""
CURRENT_WORKTREE=""
if git rev-parse --git-dir > /dev/null 2>&1; then
    CURRENT_BRANCH=$(git branch --show-current 2>/dev/null || echo "detached")
    MAIN_WORKTREE=$(git worktree list | head -1 | awk '{print $1}')
    if [ "$(pwd)" != "$MAIN_WORKTREE" ]; then
        CURRENT_WORKTREE=$(basename "$(pwd)")
    else
        CURRENT_WORKTREE="main"
    fi
fi

if [ ! -d "$WORK_STATE_DIR" ]; then
    echo "No previous work state found in $WORKING_DIR"
    exit 0
fi

PENDING_FILE="$WORK_STATE_DIR/PENDING_TODOS.json"

if [ ! -f "$PENDING_FILE" ]; then
    echo "No pending todos found"
    exit 0
fi

PENDING_COUNT=$(jq length "$PENDING_FILE" 2>/dev/null)

if [ "$PENDING_COUNT" -eq 0 ]; then
    echo "No pending todos to restore"
    exit 0
fi

echo "Found $PENDING_COUNT pending todos from previous session:"
echo ""

# Check if todos have git context and if it matches current context
HAS_GIT_CONTEXT=$(jq -r 'if length > 0 then .[0] | has("git_branch") else false end' "$PENDING_FILE")

if [ "$HAS_GIT_CONTEXT" = "true" ] && [ ! -z "$CURRENT_BRANCH" ]; then
    # Show todos with git context awareness
    echo "=== Current Context ==="
    echo "Branch: $CURRENT_BRANCH | Worktree: $CURRENT_WORKTREE"
    echo ""
    
    # Show todos from same branch/worktree
    SAME_CONTEXT=$(jq --arg branch "$CURRENT_BRANCH" --arg worktree "$CURRENT_WORKTREE" '
        map(select(.git_branch == $branch and .git_worktree == $worktree))
    ' "$PENDING_FILE")
    
    SAME_COUNT=$(echo "$SAME_CONTEXT" | jq length)
    
    if [ "$SAME_COUNT" -gt 0 ]; then
        echo "=== Todos from current branch/worktree ==="
        echo "$SAME_CONTEXT" | jq -r '.[] | "🔄 " + .content + " (" + .status + ", " + .priority + " priority)"'
        echo ""
    fi
    
    # Show todos from other branches/worktrees
    OTHER_CONTEXT=$(jq --arg branch "$CURRENT_BRANCH" --arg worktree "$CURRENT_WORKTREE" '
        map(select(.git_branch != $branch or .git_worktree != $worktree))
    ' "$PENDING_FILE")
    
    OTHER_COUNT=$(echo "$OTHER_CONTEXT" | jq length)
    
    if [ "$OTHER_COUNT" -gt 0 ]; then
        echo "=== Todos from other branches/worktrees ==="
        echo "$OTHER_CONTEXT" | jq -r '.[] | "🔄 " + .content + " (" + .status + ", " + .priority + " priority) [" + .git_branch + "/" + .git_worktree + "]"'
        echo ""
    fi
else
    # Fallback for todos without git context
    jq -r '.[] | "🔄 " + .content + " (" + .status + ", " + .priority + " priority)"' "$PENDING_FILE"
    echo ""
fi

echo "Add these to your new todo list with:"
echo "TodoWrite tool using the pending todos above"