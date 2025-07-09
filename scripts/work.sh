#!/bin/bash

# Claude Work Tracking Manual Command
# Provides manual equivalents of automated features: /work load, /work save, /work view

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="$HOME/.claude/work-tracking-config.json"

# Load configuration helper
source "$SCRIPT_DIR/work-presentation.sh" 2>/dev/null || true

# Function to show help
show_help() {
    echo "Claude Work Tracking Manual Commands"
    echo ""
    echo "Usage: ~/.claude/scripts/work.sh <command> [options]"
    echo ""
    echo "Commands:"
    echo "  load [branch]   Load and restore work state (optionally from specific branch)"
    echo "  save [note]     Save current work state manually with optional note"
    echo "  view [filter]   View global work overview (optionally filtered by keyword)"
    echo "  status          Show current session status and active todos"
    echo "  conflicts <key> Find related work across worktrees"
    echo ""
    echo "Options:"
    echo "  --help, -h      Show this help message"
    echo "  --quiet, -q     Minimal output"
    echo "  --verbose, -v   Detailed output"
    echo ""
    echo "Examples:"
    echo "  ~/.claude/scripts/work.sh load              # Load todos for current branch"
    echo "  ~/.claude/scripts/work.sh load feature-auth # Load todos from feature-auth branch"
    echo "  ~/.claude/scripts/work.sh save 'checkpoint' # Save with note"
    echo "  ~/.claude/scripts/work.sh view auth         # View auth-related work"
    echo "  ~/.claude/scripts/work.sh conflicts api     # Find API-related work conflicts"
}

# Function to load work state
work_load() {
    local target_branch="$1"
    local current_branch=""
    
    if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
        current_branch=$(git branch --show-current 2>/dev/null || echo "main")
    fi
    
    # Use provided branch or current branch
    local load_branch="${target_branch:-$current_branch}"
    
    echo "🔄 Loading work state..."
    echo "   Target branch: $load_branch"
    
    # Call the existing restore script with branch context
    if [ -f "$SCRIPT_DIR/restore-todos.sh" ]; then
        RESTORE_BRANCH="$load_branch" "$SCRIPT_DIR/restore-todos.sh"
    else
        echo "❌ Restore script not found. Run installation first."
        exit 1
    fi
}

# Function to save work state
work_save() {
    local note="$1"
    local session_id=$(date +%Y%m%d_%H%M%S)_$$
    local git_context=$(get_git_context)
    local branch=$(echo "$git_context" | cut -d'|' -f1)
    local worktree=$(echo "$git_context" | cut -d'|' -f2)
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    echo "💾 Saving work state..."
    echo "   Session: $session_id"
    echo "   Branch: $branch"
    echo "   Worktree: $worktree"
    
    # Create todo directory if it doesn't exist
    mkdir -p "$HOME/.claude/todos"
    
    # Create manual save entry
    local log_file="$HOME/.claude/todos/${session_id}.json"
    
    cat > "$log_file" << EOF
{
  "sessionId": "$session_id",
  "timestamp": "$timestamp",
  "branch": "$branch",
  "worktree": "$worktree",
  "workingDirectory": "$(pwd)",
  "logType": "manual_work_command",
  "todos": [],
  "notes": "${note:-Manual save via /work save command}"
}
EOF
    
    # Update global state
    if [ -f "$SCRIPT_DIR/update-global-state.sh" ]; then
        "$SCRIPT_DIR/update-global-state.sh" >/dev/null 2>&1 || true
    fi
    
    echo "✅ Work state saved successfully!"
    
    # Show summary
    local todo_count=$(find "$HOME/.claude/todos" -name "*.json" 2>/dev/null | wc -l)
    echo "   Total sessions: $todo_count"
}

# Function to view global work overview
work_view() {
    local filter="$1"
    
    echo "📊 Global Work Overview"
    echo ""
    
    # Call existing work status script
    if [ -f "$SCRIPT_DIR/work-status.sh" ]; then
        if [ -n "$filter" ]; then
            echo "🔍 Filtering by: $filter"
            echo ""
            "$SCRIPT_DIR/work-status.sh" | grep -i "$filter" || echo "   No matching work found for '$filter'"
        else
            "$SCRIPT_DIR/work-status.sh"
        fi
    else
        echo "❌ Work status script not found. Run installation first."
        exit 1
    fi
}

# Function to show current status
work_status() {
    echo "📋 Current Session Status"
    echo ""
    
    # Show current git context
    local git_context=$(get_git_context)
    local branch=$(echo "$git_context" | cut -d'|' -f1)
    local worktree=$(echo "$git_context" | cut -d'|' -f2)
    
    echo "   Current branch: $branch"
    echo "   Current worktree: $worktree"
    echo "   Working directory: $(pwd)"
    echo ""
    
    # Show recent sessions
    echo "📝 Recent Work Sessions:"
    if [ -d "$HOME/.claude/todos" ]; then
        find "$HOME/.claude/todos" -name "*.json" -type f 2>/dev/null | \
        sort -r | head -3 | while read -r file; do
            if [ -f "$file" ]; then
                local session_id=$(jq -r '.sessionId // "unknown"' "$file" 2>/dev/null || echo "unknown")
                local file_branch=$(jq -r '.branch // "unknown"' "$file" 2>/dev/null || echo "unknown")
                local file_worktree=$(jq -r '.worktree // "unknown"' "$file" 2>/dev/null || echo "unknown")
                local timestamp=$(jq -r '.timestamp // "unknown"' "$file" 2>/dev/null || echo "unknown")
                
                echo "   • $session_id | $file_branch | $file_worktree"
            fi
        done
    else
        echo "   No sessions found"
    fi
}

# Function to find work conflicts
work_conflicts() {
    local keyword="$1"
    
    if [ -z "$keyword" ]; then
        echo "❌ Please provide a keyword to search for conflicts"
        echo "   Usage: /work conflicts <keyword>"
        exit 1
    fi
    
    echo "🔍 Finding work conflicts for: $keyword"
    echo ""
    
    # Call existing conflicts script
    if [ -f "$SCRIPT_DIR/work-conflicts.sh" ]; then
        "$SCRIPT_DIR/work-conflicts.sh" "$keyword"
    else
        echo "❌ Work conflicts script not found. Run installation first."
        exit 1
    fi
}

# Helper function to get git context
get_git_context() {
    local branch=""
    local worktree=""
    
    if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
        branch=$(git branch --show-current 2>/dev/null || echo "main")
        worktree=$(basename "$(git rev-parse --show-toplevel)" 2>/dev/null || echo "unknown")
    else
        branch="no-git"
        worktree="no-git"
    fi
    
    echo "$branch|$worktree"
}

# Main command processing
main() {
    case "$1" in
        "load")
            work_load "$2"
            ;;
        "save")
            work_save "$2"
            ;;
        "view")
            work_view "$2"
            ;;
        "status")
            work_status
            ;;
        "conflicts")
            work_conflicts "$2"
            ;;
        "--help"|"-h"|"help"|"")
            show_help
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