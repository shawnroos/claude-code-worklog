#!/bin/bash

# Enhanced Tool Complete Hook - Plan and Proposal Capture
# Captures plans from exit_plan_mode and other planning activities

# Read input from Claude Code hook
INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName // empty')
SESSION_ID=$(echo "$INPUT" | jq -r '.sessionId // empty')
WORKING_DIR=$(echo "$INPUT" | jq -r '.workingDirectory // "unknown"')
TRANSCRIPT_PATH=$(echo "$INPUT" | jq -r '.transcriptPath // empty')
TOOL_OUTPUT=$(echo "$INPUT" | jq -r '.toolOutput // empty')
TOOL_INPUT=$(echo "$INPUT" | jq -r '.toolInput // empty')

# Log for debugging
echo "[$(date)] Plan capture hook: $TOOL_NAME for session $SESSION_ID" >> ~/.claude/hooks.log

# Skip if no session ID
if [ -z "$SESSION_ID" ] || [ "$SESSION_ID" = "null" ]; then
    exit 0
fi

# Create work intelligence directory - use project-local storage
if [ -n "$WORKING_DIR" ] && [ "$WORKING_DIR" != "unknown" ]; then
    WORK_INTELLIGENCE_DIR="$WORKING_DIR/.claude-work/history"
    mkdir -p "$WORK_INTELLIGENCE_DIR"
else
    # Fallback to global storage
    WORK_INTELLIGENCE_DIR="$HOME/.claude/work-intelligence"
    mkdir -p "$WORK_INTELLIGENCE_DIR"
fi

# Function to extract and save work intelligence
save_work_intelligence() {
    local intelligence_type="$1"
    local content="$2"
    local context="$3"
    local metadata="$4"
    
    if [ -z "$content" ] || [ "$content" = "null" ]; then
        return
    fi
    
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local date_part=$(date +"%Y-%m-%d")
    local intelligence_id="${date_part}-${intelligence_type}-${SESSION_ID}"
    
    # Get git context if in a git repo
    local git_branch=""
    local git_worktree=""
    if [ -d "$WORKING_DIR/.git" ] || git -C "$WORKING_DIR" rev-parse --git-dir >/dev/null 2>&1; then
        git_branch=$(git -C "$WORKING_DIR" branch --show-current 2>/dev/null || echo "")
        git_worktree=$(git -C "$WORKING_DIR" worktree list --porcelain 2>/dev/null | grep "^worktree" | head -1 | cut -d' ' -f2 || echo "")
    fi
    
    # Create intelligence record
    jq -n \
        --arg id "$intelligence_id" \
        --arg type "$intelligence_type" \
        --arg content "$content" \
        --arg context "$context" \
        --arg metadata "$metadata" \
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
            metadata: $metadata,
            tool_name: $tool,
            timestamp: $timestamp,
            session_id: $session,
            working_directory: $dir,
            git_branch: $branch,
            git_worktree: $worktree
        }' > "$WORK_INTELLIGENCE_DIR/${intelligence_id}.json"
    
    echo "[$(date)] Saved work intelligence: $intelligence_type from $TOOL_NAME" >> ~/.claude/hooks.log
}

# Function to extract last Claude response from transcript
extract_last_response() {
    if [ -z "$TRANSCRIPT_PATH" ] || [ ! -f "$TRANSCRIPT_PATH" ]; then
        return
    fi
    
    # Get the last few lines and extract Claude's text responses
    tail -50 "$TRANSCRIPT_PATH" | \
        jq -r 'select(.type == "text" and .source == "assistant") | .text' | \
        tail -1
}

# Function to extract plan from exit_plan_mode
extract_plan_from_response() {
    local response="$1"
    
    # Look for plan patterns in the response
    if echo "$response" | grep -q -E "(plan|steps|implementation|approach)"; then
        # Extract numbered lists, bullet points, or structured content
        echo "$response" | grep -E "^[0-9]+\.|^-|^\*|^Step|^Phase" || echo "$response"
    fi
}

# Function to detect proposals and decisions
detect_proposals() {
    local response="$1"
    
    if echo "$response" | grep -q -E "(I recommend|I suggest|proposal|approach|decision|architecture|design)"; then
        echo "$response"
    fi
}

# Process different tools for work intelligence
case "$TOOL_NAME" in
    "exit_plan_mode")
        # Capture plan from exit_plan_mode tool
        local plan_content=$(echo "$TOOL_INPUT" | jq -r '.plan // empty')
        if [ -n "$plan_content" ]; then
            save_work_intelligence "plan" "$plan_content" "Plan created via exit_plan_mode" "{\"structured\": true}"
        fi
        ;;
        
    "TodoWrite")
        # Enhanced todo capture - look for plans in todo content
        local todos=$(echo "$TOOL_INPUT" | jq -r '.todos[]? // empty')
        if [ -n "$todos" ]; then
            # Check if any todos contain plan-like content
            echo "$todos" | jq -r 'select(.content | test("plan|implement|design|architecture"; "i")) | @json' | while read -r todo; do
                if [ -n "$todo" ]; then
                    local todo_content=$(echo "$todo" | jq -r '.content')
                    save_work_intelligence "structured_todo" "$todo_content" "Todo with planning content" "{\"priority\": \"$(echo "$todo" | jq -r '.priority // "medium"')\"}"
                fi
            done
        fi
        ;;
        
    "Task")
        # Task tool results might contain strategic insights
        if [ -n "$TOOL_OUTPUT" ]; then
            # Check if output contains strategic content
            if echo "$TOOL_OUTPUT" | grep -q -E "(recommend|suggest|approach|strategy|architecture|design pattern)"; then
                save_work_intelligence "strategic_insight" "$TOOL_OUTPUT" "Strategic insight from Task agent" "{\"agent_research\": true}"
            fi
        fi
        ;;
        
    "Read"|"Grep"|"Bash")
        # These tools might trigger planning responses from Claude
        sleep 1  # Give Claude time to respond
        local response=$(extract_last_response)
        if [ -n "$response" ] && [ ${#response} -gt 200 ]; then
            # Check for plan content
            local plan_content=$(extract_plan_from_response "$response")
            if [ -n "$plan_content" ]; then
                save_work_intelligence "discovered_plan" "$plan_content" "Plan discovered during $TOOL_NAME analysis" "{\"trigger_tool\": \"$TOOL_NAME\"}"
            fi
            
            # Check for proposals
            local proposal_content=$(detect_proposals "$response")
            if [ -n "$proposal_content" ]; then
                save_work_intelligence "proposal" "$proposal_content" "Proposal made during $TOOL_NAME analysis" "{\"trigger_tool\": \"$TOOL_NAME\"}"
            fi
        fi
        ;;
esac

# Also try to capture any substantial Claude responses that might contain planning
if [[ "$TOOL_NAME" =~ ^(Read|Grep|Task|Bash)$ ]]; then
    # Give Claude a moment to respond
    sleep 2
    local response=$(extract_last_response)
    if [ -n "$response" ] && [ ${#response} -gt 300 ]; then
        # Look for session summary patterns
        if echo "$response" | grep -q -E "(summary|completed|accomplished|next steps|in conclusion)"; then
            save_work_intelligence "session_summary" "$response" "Session summary from Claude" "{\"auto_detected\": true}"
        fi
        
        # Look for architectural decisions
        if echo "$response" | grep -q -E "(because|rationale|reason|decided|chosen|approach)"; then
            save_work_intelligence "decision_rationale" "$response" "Decision rationale from Claude" "{\"auto_detected\": true}"
        fi
    fi
fi

# Update work intelligence aggregation
if [ -f "$HOME/.claude/scripts/update-work-intelligence.sh" ]; then
    "$HOME/.claude/scripts/update-work-intelligence.sh" "$WORKING_DIR" &
fi

exit 0