#!/bin/bash

# Claude Work Tracking Save Command
# Manually saves current work state to the worklog system

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="$HOME/.claude/work-tracking-config.json"

# Load configuration helper
source "$SCRIPT_DIR/work-presentation.sh"

# Function to get current git context
get_git_context() {
    local branch=""
    local worktree=""
    
    if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
        branch=$(git branch --show-current 2>/dev/null || echo "main")
        worktree=$(basename "$(git rev-parse --show-toplevel)" 2>/dev/null || echo "unknown")
    fi
    
    echo "$branch|$worktree"
}

# Function to generate session ID
generate_session_id() {
    echo "$(date +%Y%m%d_%H%M%S)_$$"
}

# Function to log current work
log_work() {
    local session_id=$(generate_session_id)
    local git_context=$(get_git_context)
    local branch=$(echo "$git_context" | cut -d'|' -f1)
    local worktree=$(echo "$git_context" | cut -d'|' -f2)
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    # Create todo directory if it doesn't exist
    mkdir -p "$HOME/.claude/todos"
    
    # Check if there are any active todos to log
    # For now, we'll create a manual log entry
    local log_file="$HOME/.claude/todos/${session_id}.json"
    
    # Create log entry
    cat > "$log_file" << EOF
{
  "sessionId": "$session_id",
  "timestamp": "$timestamp",
  "branch": "$branch",
  "worktree": "$worktree",
  "workingDirectory": "$(pwd)",
  "logType": "manual",
  "todos": [],
  "notes": "Manual log entry created via /save command"
}
EOF
    
    # Update global state
    if [ -f "$SCRIPT_DIR/update-global-state.sh" ]; then
        "$SCRIPT_DIR/update-global-state.sh" >/dev/null 2>&1 || true
    fi
    
    # Provide feedback
    echo "📝 Work saved successfully!"
    echo "   Session: $session_id"
    echo "   Branch: $branch"
    echo "   Worktree: $worktree"
    
    # Show if there are any existing todos
    local todo_count=$(find "$HOME/.claude/todos" -name "*.json" 2>/dev/null | wc -l)
    echo "   Total sessions: $todo_count"
}

# Function to show recent logs
show_recent() {
    echo "📋 Recent work saves:"
    echo ""
    
    if [ -d "$HOME/.claude/todos" ]; then
        find "$HOME/.claude/todos" -name "*.json" -type f 2>/dev/null | \
        sort -r | head -5 | while read -r file; do
            if [ -f "$file" ]; then
                local session_id=$(jq -r '.sessionId // "unknown"' "$file" 2>/dev/null || echo "unknown")
                local timestamp=$(jq -r '.timestamp // "unknown"' "$file" 2>/dev/null || echo "unknown")
                local branch=$(jq -r '.branch // "unknown"' "$file" 2>/dev/null || echo "unknown")
                local worktree=$(jq -r '.worktree // "unknown"' "$file" 2>/dev/null || echo "unknown")
                
                echo "   $session_id | $branch | $worktree | $timestamp"
            fi
        done
    else
        echo "   No saves found"
    fi
}

# Function to show help
show_help() {
    echo "Claude Work Tracking Save Command"
    echo ""
    echo "Usage: ~/.claude/scripts/save.sh [command]"
    echo ""
    echo "Commands:"
    echo "  (default)   Save current work state"
    echo "  recent      Show recent save entries"
    echo ""
    echo "Options:"
    echo "  --help, -h  Show this help message"
    echo ""
    echo "Examples:"
    echo "  ~/.claude/scripts/save.sh"
    echo "  ~/.claude/scripts/save.sh recent"
}

# Main command processing
main() {
    case "$1" in
        "recent")
            show_recent
            ;;
        "--help"|"-h"|"help")
            show_help
            ;;
        "")
            log_work
            ;;
        *)
            echo "❌ Unknown command: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"