#!/bin/bash

# Context Lifecycle Management Script
# Manages aging, promotion, and summarization of work intelligence

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
WORK_DIR="$PROJECT_DIR/.claude-work"
HISTORY_DIR="$WORK_DIR/history"
ACTIVE_DIR="$WORK_DIR/active"

# Configuration
MAX_ACTIVE_SIZE_KB=10
COMPLETED_ITEM_ARCHIVE_DAYS=7
LOW_PRIORITY_ARCHIVE_DAYS=30
ACCESSED_RECENTLY_DAYS=14

# Function to get file size in KB
get_file_size_kb() {
    local file="$1"
    if [ -f "$file" ]; then
        local size_bytes=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null)
        echo $((size_bytes / 1024))
    else
        echo 0
    fi
}

# Function to check if item should be archived
should_archive_item() {
    local item_file="$1"
    local current_date=$(date +%s)
    
    if [ ! -f "$item_file" ]; then
        return 1
    fi
    
    # Get item metadata
    local status=$(jq -r '.status // "pending"' "$item_file" 2>/dev/null)
    local priority=$(jq -r '.priority // "medium"' "$item_file" 2>/dev/null)
    local timestamp=$(jq -r '.timestamp // ""' "$item_file" 2>/dev/null)
    local last_accessed=$(jq -r '.last_accessed // ""' "$item_file" 2>/dev/null)
    
    # Parse timestamp
    local item_date
    if [ -n "$timestamp" ]; then
        item_date=$(date -d "$timestamp" +%s 2>/dev/null || date -j -f "%Y-%m-%dT%H:%M:%SZ" "$timestamp" +%s 2>/dev/null)
    else
        item_date=$current_date
    fi
    
    # Check archival conditions
    local age_days=$(( (current_date - item_date) / 86400 ))
    
    # Archive completed items after configured days
    if [ "$status" = "completed" ] && [ $age_days -gt $COMPLETED_ITEM_ARCHIVE_DAYS ]; then
        return 0
    fi
    
    # Archive low priority items after configured days
    if [ "$priority" = "low" ] && [ $age_days -gt $LOW_PRIORITY_ARCHIVE_DAYS ]; then
        return 0
    fi
    
    # Don't archive recently accessed items
    if [ -n "$last_accessed" ]; then
        local last_access_date=$(date -d "$last_accessed" +%s 2>/dev/null || date -j -f "%Y-%m-%dT%H:%M:%SZ" "$last_accessed" +%s 2>/dev/null)
        local access_age_days=$(( (current_date - last_access_date) / 86400 ))
        if [ $access_age_days -lt $ACCESSED_RECENTLY_DAYS ]; then
            return 1
        fi
    fi
    
    return 1
}

# Function to archive active item
archive_active_item() {
    local item_file="$1"
    local item_basename=$(basename "$item_file")
    local archive_date=$(date +%Y-%m-%d)
    local archive_file="$HISTORY_DIR/${archive_date}-${item_basename}"
    
    echo "🗄️  Archiving: $item_basename"
    
    # Add archival metadata
    jq '. + {
        "archived_date": "'"$(date -u +%Y-%m-%dT%H:%M:%SZ)"'",
        "archived_from": "active",
        "archive_reason": "lifecycle_management"
    }' "$item_file" > "$archive_file"
    
    # Remove from active
    rm "$item_file"
    
    echo "   → Moved to: history/$(basename "$archive_file")"
}

# Function to manage active context size
manage_active_context_size() {
    local context_file="$ACTIVE_DIR/current-work-context.json"
    
    if [ ! -f "$context_file" ]; then
        return 0
    fi
    
    local current_size=$(get_file_size_kb "$context_file")
    
    if [ $current_size -gt $MAX_ACTIVE_SIZE_KB ]; then
        echo "⚠️  Active context size ($current_size KB) exceeds limit ($MAX_ACTIVE_SIZE_KB KB)"
        
        # Create condensed version
        jq '
        {
            "active_work": {
                "current_focus": .active_work.current_focus,
                "session_id": .active_work.session_id,
                "timestamp": .active_work.timestamp,
                "priority_items": (.active_work.priority_items // [] | map(select(.priority == "high" or .priority == "critical"))),
                "recent_decisions": (.active_work.recent_decisions // [] | .[0:3]),
                "historical_references": (.active_work.historical_references // [] | .[0:5])
            },
            "context_metadata": (.context_metadata // {} | . + {
                "last_condensed": "'"$(date -u +%Y-%m-%dT%H:%M:%SZ)"'",
                "condensed_reason": "size_limit_exceeded"
            })
        }' "$context_file" > "$context_file.tmp"
        
        mv "$context_file.tmp" "$context_file"
        
        local new_size=$(get_file_size_kb "$context_file")
        echo "   → Condensed to $new_size KB"
    fi
}

# Function to check future work and suggest grooming
check_future_work_grooming() {
    local future_dir="$WORK_DIR/future"
    
    if [ ! -d "$future_dir" ]; then
        return 0
    fi
    
    echo "🔍 Checking future work organization..."
    
    # Count ungrouped items
    local ungrouped_count=0
    if [ -d "$future_dir/items" ]; then
        ungrouped_count=$(find "$future_dir/items" -name "*.json" 2>/dev/null | wc -l | tr -d ' ')
    fi
    
    # Count existing groups
    local groups_count=0
    if [ -d "$future_dir/groups" ]; then
        groups_count=$(find "$future_dir/groups" -name "*.json" 2>/dev/null | wc -l | tr -d ' ')
    fi
    
    # Suggest grooming if there are many ungrouped items
    if [ $ungrouped_count -gt 5 ]; then
        echo "📋 Found $ungrouped_count ungrouped future work items"
        echo "   Consider running groom_future_work() to organize them"
        if [ $groups_count -eq 0 ]; then
            echo "   No groups exist yet - good opportunity to create initial organization"
        fi
    elif [ $ungrouped_count -gt 0 ]; then
        echo "📝 $ungrouped_count ungrouped items, $groups_count existing groups"
    fi
    
    # Check if suggestions file exists and has recommendations
    local suggestions_file="$future_dir/suggestions.json"
    if [ -f "$suggestions_file" ]; then
        local suggestions_count=$(jq -r '.grouping_suggestions | length' "$suggestions_file" 2>/dev/null || echo "0")
        if [ "$suggestions_count" != "0" ] && [ $suggestions_count -gt 0 ]; then
            echo "💡 $suggestions_count intelligent grouping suggestions available"
            echo "   Use list_future_groups() to review suggestions"
        fi
    fi
}

# Function to generate weekly digest
generate_weekly_digest() {
    local digest_date=$(date +%Y-%m-%d)
    local digest_file="$HISTORY_DIR/${digest_date}-weekly-digest.json"
    
    # Skip if digest already exists for this week
    if [ -f "$digest_file" ]; then
        return 0
    fi
    
    echo "📊 Generating weekly digest..."
    
    # Get work items from the last 7 days
    local week_ago=$(date -d "7 days ago" +%Y-%m-%d 2>/dev/null || date -j -v-7d +%Y-%m-%d 2>/dev/null)
    
    # Create digest
    cat > "$digest_file" << EOF
{
    "type": "weekly_digest",
    "digest_date": "$digest_date",
    "period": {
        "start": "${week_ago}T00:00:00Z",
        "end": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    },
    "summary": {
        "total_items": $(find "$HISTORY_DIR" -name "*.json" -newer "$digest_file" 2>/dev/null | wc -l),
        "completed_plans": $(find "$HISTORY_DIR" -name "*plan*.json" -newer "$digest_file" 2>/dev/null | xargs grep -l '"status": "completed"' 2>/dev/null | wc -l),
        "accepted_proposals": $(find "$HISTORY_DIR" -name "*proposal*.json" -newer "$digest_file" 2>/dev/null | xargs grep -l '"status": "accepted"' 2>/dev/null | wc -l),
        "key_decisions": []
    },
    "metadata": {
        "generated_by": "context-lifecycle.sh",
        "auto_generated": true
    }
}
EOF
    
    echo "   → Created: history/$(basename "$digest_file")"
}

# Function to update item access timestamp
update_access_timestamp() {
    local item_file="$1"
    
    if [ -f "$item_file" ]; then
        jq '. + {"last_accessed": "'"$(date -u +%Y-%m-%dT%H:%M:%SZ)"'"}' "$item_file" > "$item_file.tmp"
        mv "$item_file.tmp" "$item_file"
    fi
}

# Main lifecycle management function
run_lifecycle_management() {
    echo "🔄 Running context lifecycle management..."
    
    # Ensure directories exist
    mkdir -p "$HISTORY_DIR" "$ACTIVE_DIR"
    
    # Archive eligible active items
    for item_file in "$ACTIVE_DIR"/*.json; do
        if [ -f "$item_file" ] && [ "$(basename "$item_file")" != "current-work-context.json" ]; then
            if should_archive_item "$item_file"; then
                archive_active_item "$item_file"
            fi
        fi
    done
    
    # Manage active context size
    manage_active_context_size
    
    # Check future work grooming opportunities
    check_future_work_grooming
    
    # Generate weekly digest if needed
    generate_weekly_digest
    
    echo "✅ Context lifecycle management complete"
}

# Command line interface
case "${1:-run}" in
    "run")
        run_lifecycle_management
        ;;
    "archive")
        if [ -n "$2" ]; then
            archive_active_item "$ACTIVE_DIR/$2"
        else
            echo "Usage: $0 archive <item_file>"
            exit 1
        fi
        ;;
    "access")
        if [ -n "$2" ]; then
            update_access_timestamp "$ACTIVE_DIR/$2"
        else
            echo "Usage: $0 access <item_file>"
            exit 1
        fi
        ;;
    "digest")
        generate_weekly_digest
        ;;
    *)
        echo "Usage: $0 {run|archive|access|digest}"
        echo "  run     - Run full lifecycle management"
        echo "  archive - Archive specific item"
        echo "  access  - Update access timestamp"
        echo "  digest  - Generate weekly digest"
        exit 1
        ;;
esac