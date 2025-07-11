#!/bin/bash

# Claude Code Tool Complete Hook - Enhanced for Report Capture
# Tracks todo changes and captures findings/reports from various tools

# Read input from Claude Code hook
INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName // empty')
SESSION_ID=$(echo "$INPUT" | jq -r '.sessionId // empty')
WORKING_DIR=$(echo "$INPUT" | jq -r '.workingDirectory // "unknown"')
TRANSCRIPT_PATH=$(echo "$INPUT" | jq -r '.transcriptPath // empty')
TOOL_OUTPUT=$(echo "$INPUT" | jq -r '.toolOutput // empty')

# Log for debugging
echo "[$(date)] Tool complete: $TOOL_NAME for session $SESSION_ID" >> ~/.claude/hooks.log

# Skip if no session ID
if [ -z "$SESSION_ID" ] || [ "$SESSION_ID" = "null" ]; then
    exit 0
fi

# Create findings directory
FINDINGS_DIR="$HOME/.claude/findings"
mkdir -p "$FINDINGS_DIR"

# Function to extract and save findings
save_finding() {
    local finding_type="$1"
    local content="$2"
    local context="$3"
    
    if [ -z "$content" ] || [ "$content" = "null" ]; then
        return
    fi
    
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local finding_id="${SESSION_ID}_${timestamp}_${finding_type}"
    
    # Get git context if in a git repo
    local git_branch=""
    local git_worktree=""
    if [ -d "$WORKING_DIR/.git" ] || git -C "$WORKING_DIR" rev-parse --git-dir >/dev/null 2>&1; then
        git_branch=$(git -C "$WORKING_DIR" branch --show-current 2>/dev/null || echo "")
        git_worktree=$(git -C "$WORKING_DIR" worktree list --porcelain 2>/dev/null | grep "^worktree" | head -1 | cut -d' ' -f2 || echo "")
    fi
    
    # Create finding record
    jq -n \
        --arg id "$finding_id" \
        --arg type "$finding_type" \
        --arg content "$content" \
        --arg context "$context" \
        --arg tool "$TOOL_NAME" \
        --arg timestamp "$timestamp" \
        --arg session "$SESSION_ID" \
        --arg dir "$WORKING_DIR" \
        --arg branch "$git_branch" \
        --arg worktree "$git_worktree" \
        '{
            id: $id,
            type: $type,
            content: $content,
            context: $context,
            tool_name: $tool,
            timestamp: $timestamp,
            session_id: $session,
            working_directory: $dir,
            git_branch: $branch,
            git_worktree: $worktree
        }' > "$FINDINGS_DIR/${finding_id}.json"
    
    echo "[$(date)] Saved finding: $finding_type from $TOOL_NAME" >> ~/.claude/hooks.log
}

# Function to extract last Claude response from transcript
extract_last_response() {
    if [ -z "$TRANSCRIPT_PATH" ] || [ ! -f "$TRANSCRIPT_PATH" ]; then
        return
    fi
    
    # Get the last few lines and extract Claude's text responses
    tail -20 "$TRANSCRIPT_PATH" | \
        jq -r 'select(.type == "text" and .source == "assistant") | .text' | \
        tail -1
}

# Process different tools
case "$TOOL_NAME" in
    "Task")
        # Task tool often contains research findings
        if [ -n "$TOOL_OUTPUT" ]; then
            save_finding "research" "$TOOL_OUTPUT" "Task agent research results"
        fi
        ;;
        
    "Grep")
        # Capture search summaries
        local search_pattern=$(echo "$INPUT" | jq -r '.toolInput.pattern // empty')
        local match_count=$(echo "$TOOL_OUTPUT" | wc -l)
        if [ "$match_count" -gt 5 ]; then
            # Significant search results
            local summary="Found $match_count matches for pattern: $search_pattern"
            save_finding "search" "$summary" "$TOOL_OUTPUT"
        fi
        ;;
        
    "Read"|"NotebookRead")
        # Track file analysis
        local file_path=$(echo "$INPUT" | jq -r '.toolInput.file_path // .toolInput.notebook_path // empty')
        if [ -n "$file_path" ]; then
            local finding="Analyzed file: $file_path"
            save_finding "analysis" "$finding" "File reading for analysis"
        fi
        ;;
        
    "Bash")
        # Capture command outputs that might be reports
        local command=$(echo "$INPUT" | jq -r '.toolInput.command // empty')
        # Check for common reporting commands
        if [[ "$command" =~ (test|lint|build|npm run|pytest|cargo test) ]]; then
            if [ -n "$TOOL_OUTPUT" ]; then
                save_finding "test_results" "$TOOL_OUTPUT" "Command: $command"
            fi
        fi
        ;;
        
    "Edit"|"Write"|"MultiEdit")
        # Track implementation completion
        local file_path=$(echo "$INPUT" | jq -r '.toolInput.file_path // empty')
        if [ -n "$file_path" ]; then
            local finding="Modified file: $file_path"
            save_finding "implementation" "$finding" "Code modification"
        fi
        ;;
        
    "TodoWrite")
        # Original todo tracking functionality
        TODO_FILE="$HOME/.claude/todos/${SESSION_ID}-agent-${SESSION_ID}.json"
        
        if [ -f "$TODO_FILE" ] && [ -d "$WORKING_DIR" ]; then
            WORK_STATE_DIR="$WORKING_DIR/.claude-work"
            mkdir -p "$WORK_STATE_DIR"
            cp "$TODO_FILE" "$WORK_STATE_DIR/current_todos.json" 2>/dev/null
            echo "[$(date)] Todo state backed up to $WORK_STATE_DIR" >> ~/.claude/hooks.log
        fi
        ;;
esac

# Try to capture Claude's response that follows certain tool patterns
if [[ "$TOOL_NAME" =~ ^(Task|Grep|Read)$ ]]; then
    # Give Claude a moment to respond
    sleep 0.5
    local response=$(extract_last_response)
    if [ -n "$response" ] && [ ${#response} -gt 100 ]; then
        # Likely a substantial report
        save_finding "report" "$response" "Claude's analysis following $TOOL_NAME"
    fi
fi

exit 0