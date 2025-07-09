#!/bin/bash

# Claude Code Session Completion Hook
# Captures completed work and preserves incomplete todos with git context

# Read input from Claude Code hook
INPUT=$(cat)
SESSION_ID=$(echo "$INPUT" | jq -r '.sessionId // empty')
WORKING_DIR=$(echo "$INPUT" | jq -r '.workingDirectory // "unknown"')

# Get git context if in a git repository
GIT_BRANCH=""
GIT_WORKTREE=""
GIT_REMOTE_URL=""

if [ "$WORKING_DIR" != "unknown" ] && [ -d "$WORKING_DIR" ]; then
    cd "$WORKING_DIR"
    if git rev-parse --git-dir > /dev/null 2>&1; then
        GIT_BRANCH=$(git branch --show-current 2>/dev/null || echo "detached")
        GIT_WORKTREE=$(git worktree list --porcelain | grep "^worktree $(pwd)$" -A1 | grep "^branch" | cut -d' ' -f2 2>/dev/null || echo "main")
        GIT_REMOTE_URL=$(git remote get-url origin 2>/dev/null || echo "local")
        
        # Determine if this is a worktree (not main worktree)
        MAIN_WORKTREE=$(git worktree list | head -1 | awk '{print $1}')
        if [ "$(pwd)" != "$MAIN_WORKTREE" ]; then
            WORKTREE_NAME=$(basename "$(pwd)")
        else
            WORKTREE_NAME="main"
        fi
    fi
fi

# Log for debugging
echo "[$(date)] Session complete: $SESSION_ID in $WORKING_DIR (branch: $GIT_BRANCH, worktree: $WORKTREE_NAME)" >> ~/.claude/hooks.log

# Skip if no session ID
if [ -z "$SESSION_ID" ] || [ "$SESSION_ID" = "null" ]; then
    exit 0
fi

# Find the current session's todo file
TODO_FILE="$HOME/.claude/todos/${SESSION_ID}-agent-${SESSION_ID}.json"

if [ ! -f "$TODO_FILE" ]; then
    echo "[$(date)] No todo file found: $TODO_FILE" >> ~/.claude/hooks.log
    exit 0
fi

# Parse todos
COMPLETED_TODOS=$(jq -r '.[] | select(.status == "completed") | "✅ " + .content' "$TODO_FILE" 2>/dev/null)
INCOMPLETE_TODOS=$(jq -r '.[] | select(.status != "completed") | {content: .content, status: .status, priority: .priority}' "$TODO_FILE" 2>/dev/null)

# Create work state directory in current working directory
if [ "$WORKING_DIR" != "unknown" ] && [ -d "$WORKING_DIR" ]; then
    WORK_STATE_DIR="$WORKING_DIR/.claude-work"
    mkdir -p "$WORK_STATE_DIR"
    
    # Update work history with completed todos
    if [ ! -z "$COMPLETED_TODOS" ]; then
        {
            echo ""
            echo "## Session $(date '+%Y-%m-%d %H:%M:%S')"
            if [ ! -z "$GIT_BRANCH" ]; then
                echo "**Branch:** \`$GIT_BRANCH\` | **Worktree:** \`$WORKTREE_NAME\`"
                echo ""
            fi
            echo "$COMPLETED_TODOS"
        } >> "$WORK_STATE_DIR/WORK_HISTORY.md"
    fi
    
    # Save incomplete todos for next session with git context
    if [ ! -z "$INCOMPLETE_TODOS" ]; then
        # Add git context to pending todos
        TODOS_WITH_CONTEXT=$(echo "$INCOMPLETE_TODOS" | jq -s --arg branch "$GIT_BRANCH" --arg worktree "$WORKTREE_NAME" '
            map(. + {
                "git_branch": $branch,
                "git_worktree": $worktree,
                "saved_at": (now | strftime("%Y-%m-%d %H:%M:%S"))
            })
        ')
        echo "$TODOS_WITH_CONTEXT" > "$WORK_STATE_DIR/PENDING_TODOS.json"
        
        {
            echo ""
            echo "### Incomplete Work"
            if [ ! -z "$GIT_BRANCH" ]; then
                echo "**Context:** \`$GIT_BRANCH\` branch in \`$WORKTREE_NAME\` worktree"
                echo ""
            fi
            echo "$INCOMPLETE_TODOS" | jq -r '"🔄 " + .content + " (" + .status + ")"'
        } >> "$WORK_STATE_DIR/WORK_HISTORY.md"
    else
        # Clear pending todos if all completed
        echo "[]" > "$WORK_STATE_DIR/PENDING_TODOS.json"
    fi
    
    echo "[$(date)] Work state saved to $WORK_STATE_DIR" >> ~/.claude/hooks.log
    
    # Count todos for presentation
    COMPLETED_COUNT=$(echo "$COMPLETED_TODOS" | wc -l | tr -d ' ')
    PENDING_COUNT=$(echo "$INCOMPLETE_TODOS" | jq -s 'length' 2>/dev/null || echo "0")
    
    # Fix counts (wc -l counts empty string as 1)
    if [ -z "$COMPLETED_TODOS" ]; then
        COMPLETED_COUNT=0
    fi
    
    # Show session summary to user
    ~/.claude/scripts/work-presentation.sh summary "$COMPLETED_COUNT" "$PENDING_COUNT" "$WORKTREE_NAME" "$GIT_BRANCH"
    
    # Show sync notification
    ~/.claude/scripts/work-presentation.sh sync "Updating global work state" "for $WORKTREE_NAME"
    
    # Update global work state aggregation
    if [ ! -z "$GIT_BRANCH" ] && [ ! -z "$WORKTREE_NAME" ]; then
        ~/.claude/scripts/update-global-state.sh "$WORKING_DIR" "$WORKTREE_NAME" "$GIT_BRANCH" &
    fi
    
    # Check for potential conflicts (async)
    (
        sleep 2  # Let global state update first
        if [ -f "$HOME/.claude/work-state/projects/$(basename "$WORKING_DIR")/worktrees" ]; then
            CONFLICT_COUNT=$(find "$HOME/.claude/work-state/projects/$(basename "$WORKING_DIR")/worktrees" -name "*.json" -not -name "$WORKTREE_NAME.json" | wc -l | tr -d ' ')
            ~/.claude/scripts/work-presentation.sh conflicts "$(basename "$WORKING_DIR")" "$CONFLICT_COUNT"
        fi
    ) &
fi

exit 0