#!/bin/bash

# Claude Code Work Tracking Session Initialization
# Automatically runs at the start of each Claude session
# Restores pending todos and provides session context

# Silent mode by default (can be overridden)
SILENT_MODE=${1:-true}
WORKING_DIR=$(pwd)

# Configuration
CONFIG_FILE="$HOME/.claude/work-tracking-config.json"
WORK_STATE_DIR="$WORKING_DIR/.claude-work"

# Colors (only used if not silent)
if [ "$SILENT_MODE" != "true" ]; then
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    CYAN='\033[0;36m'
    NC='\033[0m'
else
    GREEN=''
    YELLOW=''
    BLUE=''
    CYAN=''
    NC=''
fi

# Function to get emoji
get_emoji() {
    local type="$1"
    if [ -f "$CONFIG_FILE" ]; then
        local style=$(jq -r '.presentation.emoji_style // "minimal_colored"' "$CONFIG_FILE" 2>/dev/null)
        local emoji=$(jq -r ".emoji_styles.$style.$type // \"$type\"" "$CONFIG_FILE" 2>/dev/null)
        printf "$emoji"
    else
        echo "$type"
    fi
}

# Check if work tracking is installed
if [ ! -f "$CONFIG_FILE" ]; then
    # System not installed, exit silently
    exit 0
fi

# Check if auto-restore is enabled
AUTO_RESTORE=$(jq -r '.session.auto_restore // true' "$CONFIG_FILE" 2>/dev/null)
if [ "$AUTO_RESTORE" = "false" ]; then
    exit 0
fi

# Get git context
GIT_BRANCH=""
GIT_WORKTREE=""
if git rev-parse --git-dir > /dev/null 2>&1; then
    GIT_BRANCH=$(git branch --show-current 2>/dev/null || echo "detached")
    MAIN_WORKTREE=$(git worktree list | head -1 | awk '{print $1}')
    if [ "$(pwd)" != "$MAIN_WORKTREE" ]; then
        GIT_WORKTREE=$(basename "$(pwd)")
    else
        GIT_WORKTREE="main"
    fi
fi

# Check for pending todos
PENDING_FILE="$WORK_STATE_DIR/PENDING_TODOS.json"
if [ ! -f "$PENDING_FILE" ]; then
    # No pending todos, exit silently
    exit 0
fi

PENDING_COUNT=$(jq length "$PENDING_FILE" 2>/dev/null || echo "0")
if [ "$PENDING_COUNT" -eq 0 ]; then
    # No pending todos, exit silently
    exit 0
fi

# Get current context for smart filtering
CURRENT_BRANCH=""
CURRENT_WORKTREE=""
if [ ! -z "$GIT_BRANCH" ]; then
    CURRENT_BRANCH="$GIT_BRANCH"
    CURRENT_WORKTREE="$GIT_WORKTREE"
fi

# Check if we should show restoration info
SHOW_NOTIFICATIONS=$(jq -r '.presentation.show_notifications // true' "$CONFIG_FILE" 2>/dev/null)

if [ "$SHOW_NOTIFICATIONS" = "true" ] && [ "$SILENT_MODE" != "true" ]; then
    echo -e "${CYAN}$(get_emoji "sync") Work Tracking: Session Started${NC}"
    
    if [ ! -z "$GIT_BRANCH" ]; then
        echo -e "${BLUE}📍 Context: $GIT_BRANCH branch in $GIT_WORKTREE worktree${NC}"
    fi
    
    # Check if todos have git context
    HAS_GIT_CONTEXT=$(jq -r 'if length > 0 then .[0] | has("git_branch") else false end' "$PENDING_FILE" 2>/dev/null)
    
    if [ "$HAS_GIT_CONTEXT" = "true" ] && [ ! -z "$CURRENT_BRANCH" ]; then
        # Show context-aware restoration
        SAME_CONTEXT=$(jq --arg branch "$CURRENT_BRANCH" --arg worktree "$CURRENT_WORKTREE" '
            map(select(.git_branch == $branch and .git_worktree == $worktree))
        ' "$PENDING_FILE" 2>/dev/null)
        
        SAME_COUNT=$(echo "$SAME_CONTEXT" | jq length 2>/dev/null || echo "0")
        
        if [ "$SAME_COUNT" -gt 0 ]; then
            echo -e "${GREEN}$(get_emoji "pending") Found $SAME_COUNT pending todos from this context${NC}"
            echo "$SAME_CONTEXT" | jq -r '.[] | "  • " + .content + " (" + .status + ")"' 2>/dev/null
        fi
        
        # Check for todos from other contexts
        OTHER_CONTEXT=$(jq --arg branch "$CURRENT_BRANCH" --arg worktree "$CURRENT_WORKTREE" '
            map(select(.git_branch != $branch or .git_worktree != $worktree))
        ' "$PENDING_FILE" 2>/dev/null)
        
        OTHER_COUNT=$(echo "$OTHER_CONTEXT" | jq length 2>/dev/null || echo "0")
        
        if [ "$OTHER_COUNT" -gt 0 ]; then
            echo -e "${YELLOW}$(get_emoji "conflict") $OTHER_COUNT todos from other contexts available${NC}"
            echo -e "${BLUE}💡 Run: restore-todos to see all pending work${NC}"
        fi
    else
        # Fallback for todos without git context
        echo -e "${GREEN}$(get_emoji "pending") Found $PENDING_COUNT pending todos${NC}"
        jq -r '.[] | "  • " + .content + " (" + .status + ")"' "$PENDING_FILE" 2>/dev/null
    fi
    
    echo -e "${BLUE}💡 Use TodoWrite to continue your work${NC}"
    echo ""
fi

# Optional: Auto-create todos based on context
AUTO_CREATE_TODOS=$(jq -r '.session.auto_create_todos // false' "$CONFIG_FILE" 2>/dev/null)

if [ "$AUTO_CREATE_TODOS" = "true" ] && [ ! -z "$GIT_BRANCH" ]; then
    # Could auto-create contextual todos here
    # For now, we'll just make the data available
    exit 0
fi

exit 0