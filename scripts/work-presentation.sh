#!/bin/bash

# Work Tracking Presentation Helper
# Controls how work tracking feedback is displayed to users

CONFIG_FILE="$HOME/.claude/work-tracking-config.json"
DEFAULT_MODE="summary"

# Load configuration
get_config() {
    local key="$1"
    local default="$2"
    
    if [ -f "$CONFIG_FILE" ]; then
        local value=$(jq -r "$key // \"$default\"" "$CONFIG_FILE" 2>/dev/null)
        if [ "$value" != "null" ] && [ ! -z "$value" ]; then
            echo "$value"
        else
            echo "$default"
        fi
    else
        echo "$default"
    fi
}

# Get current presentation mode
get_mode() {
    get_config ".presentation.mode" "$DEFAULT_MODE"
}

# Check if feature is enabled in current mode
is_enabled() {
    local feature="$1"
    local mode=$(get_mode)
    
    # Check mode-specific setting first
    local mode_setting=$(get_config ".modes.$mode.$feature" "")
    if [ ! -z "$mode_setting" ] && [ "$mode_setting" != "null" ]; then
        echo "$mode_setting"
        return
    fi
    
    # Fall back to global setting
    get_config ".presentation.$feature" "true"
}

# Get emoji for a type
get_emoji() {
    local type="$1"
    local style=$(get_config ".presentation.emoji_style" "minimal_colored")
    local emoji=$(get_config ".emoji_styles.$style.$type" "$type")
    # Use printf to properly interpret escape sequences
    printf "$emoji"
}

# Display session summary
show_session_summary() {
    local completed_count="$1"
    local pending_count="$2"
    local worktree="$3"
    local branch="$4"
    
    if [ "$(is_enabled "show_session_summary")" = "true" ]; then
        local completed_emoji=$(get_emoji "completed")
        local pending_emoji=$(get_emoji "pending")
        local success_emoji=$(get_emoji "success")
        
        echo ""
        echo "=== Work Session Summary ==="
        echo "$success_emoji **Session Complete** | Worktree: \`$worktree\` | Branch: \`$branch\`"
        
        if [ "$completed_count" -gt 0 ]; then
            echo "$completed_emoji **Completed:** $completed_count todos"
        fi
        
        if [ "$pending_count" -gt 0 ]; then
            echo "$pending_emoji **Pending:** $pending_count todos saved for next session"
        else
            echo "$completed_emoji **All todos completed!**"
        fi
        
        echo ""
    fi
}

# Display sync notification
show_sync_notification() {
    local action="$1"
    local details="$2"
    
    if [ "$(is_enabled "show_notifications")" = "true" ]; then
        local sync_emoji=$(get_emoji "sync")
        echo "$sync_emoji **Work Sync:** $action $details"
    fi
}

# Display conflicts alert
show_conflicts_alert() {
    local project="$1"
    local conflict_count="$2"
    
    if [ "$(is_enabled "show_conflicts_alert")" = "true" ] && [ "$conflict_count" -gt 0 ]; then
        local conflict_emoji=$(get_emoji "conflict")
        echo ""
        echo "$conflict_emoji **Potential Conflicts:** Found $conflict_count related todos in other worktrees"
        echo "💡 Run: \`work-conflicts\` to review"
        echo ""
    fi
}

# Set presentation mode
set_mode() {
    local new_mode="$1"
    
    if [ ! -f "$CONFIG_FILE" ]; then
        echo "Config file not found. Creating default..."
        return 1
    fi
    
    # Update mode in config
    jq ".presentation.mode = \"$new_mode\"" "$CONFIG_FILE" > "${CONFIG_FILE}.tmp" && mv "${CONFIG_FILE}.tmp" "$CONFIG_FILE"
    echo "$(get_emoji "success") Presentation mode set to: $new_mode"
}

# Main command dispatcher
case "$1" in
    "summary")
        show_session_summary "$2" "$3" "$4" "$5"
        ;;
    "sync")
        show_sync_notification "$2" "$3"
        ;;
    "conflicts")
        show_conflicts_alert "$2" "$3"
        ;;
    "mode")
        if [ -z "$2" ]; then
            echo "Current mode: $(get_mode)"
            echo "Available modes: quiet, summary, verbose"
        else
            set_mode "$2"
        fi
        ;;
    "test")
        # Test all presentation elements
        show_sync_notification "Testing" "presentation system"
        show_session_summary "3" "2" "feature-test" "test-branch"
        show_conflicts_alert "TestProject" "1"
        ;;
    *)
        echo "Usage: $0 {summary|sync|conflicts|mode|test}"
        echo "  summary <completed> <pending> <worktree> <branch>"
        echo "  sync <action> <details>"
        echo "  conflicts <project> <count>"
        echo "  mode [quiet|summary|verbose]"
        echo "  test"
        ;;
esac