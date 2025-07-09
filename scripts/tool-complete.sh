#!/bin/bash

# Claude Code Todo Tool Hook
# Tracks todo changes in real-time

# Read input from Claude Code hook
INPUT=$(cat)
TOOL_NAME=$(echo "$INPUT" | jq -r '.toolName // empty')
SESSION_ID=$(echo "$INPUT" | jq -r '.sessionId // empty')
WORKING_DIR=$(echo "$INPUT" | jq -r '.workingDirectory // "unknown"')

# Log for debugging
echo "[$(date)] Tool complete: $TOOL_NAME for session $SESSION_ID in $WORKING_DIR" >> ~/.claude/hooks.log

# Only process TodoWrite tool calls
if [ "$TOOL_NAME" != "TodoWrite" ]; then
    exit 0
fi

# Skip if no session ID or working directory
if [ -z "$SESSION_ID" ] || [ "$SESSION_ID" = "null" ] || [ "$WORKING_DIR" = "unknown" ]; then
    exit 0
fi

# Find the current session's todo file
TODO_FILE="$HOME/.claude/todos/${SESSION_ID}-agent-${SESSION_ID}.json"

if [ ! -f "$TODO_FILE" ]; then
    echo "[$(date)] No todo file found: $TODO_FILE" >> ~/.claude/hooks.log
    exit 0
fi

# Create work state directory in current working directory
if [ -d "$WORKING_DIR" ]; then
    WORK_STATE_DIR="$WORKING_DIR/.claude-work"
    mkdir -p "$WORK_STATE_DIR"
    
    # Copy current todos to work state (backup current state)
    cp "$TODO_FILE" "$WORK_STATE_DIR/current_todos.json" 2>/dev/null
    
    echo "[$(date)] Todo state backed up to $WORK_STATE_DIR" >> ~/.claude/hooks.log
fi

exit 0