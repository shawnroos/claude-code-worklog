#!/bin/bash

# Migration script for hierarchical Work + Artifacts architecture
# Converts existing work items to Artifacts and creates Work containers

echo "🔄 Migrating to hierarchical Work + Artifacts architecture..."

# Function to determine artifact type directory
get_artifact_dir() {
    local type="$1"
    case "$type" in
        "plan") echo "plans" ;;
        "proposal") echo "proposals" ;;
        "analysis") echo "analysis" ;;
        "update") echo "updates" ;;
        "decision") echo "decisions" ;;
        *) echo "plans" ;; # default
    esac
}

# Function to extract type from filename
get_type_from_filename() {
    local filename="$1"
    if [[ "$filename" =~ ^plan- ]]; then
        echo "plan"
    elif [[ "$filename" =~ ^proposal- ]]; then
        echo "proposal"
    elif [[ "$filename" =~ ^analysis- ]]; then
        echo "analysis"
    elif [[ "$filename" =~ ^update- ]]; then
        echo "update"
    elif [[ "$filename" =~ ^decision- ]]; then
        echo "decision"
    else
        echo "plan" # default
    fi
}

# Migrate items from old structure to artifacts
for schedule_dir in now next later; do
    items_dir=".claude-work/items/$schedule_dir"
    if [ -d "$items_dir" ]; then
        echo "📂 Processing $schedule_dir items..."
        
        for item_file in "$items_dir"/*.md; do
            if [ -f "$item_file" ]; then
                filename=$(basename "$item_file")
                echo "  🔄 Migrating $filename"
                
                # Extract type from filename
                type=$(get_type_from_filename "$filename")
                artifact_dir=$(get_artifact_dir "$type")
                
                # Copy to artifacts directory
                cp "$item_file" ".claude-work/artifacts/$artifact_dir/"
                echo "    ✅ Moved to .claude-work/artifacts/$artifact_dir/"
            fi
        done
    fi
done

# Migrate existing decisions
if [ -d ".claude-work/decisions/active" ]; then
    echo "📂 Processing decisions..."
    for decision_file in .claude-work/decisions/active/*.md; do
        if [ -f "$decision_file" ]; then
            filename=$(basename "$decision_file")
            echo "  🔄 Migrating decision $filename"
            cp "$decision_file" ".claude-work/artifacts/decisions/"
            echo "    ✅ Moved to .claude-work/artifacts/decisions/"
        fi
    done
fi

# Show new structure
echo ""
echo "📁 New directory structure:"
tree .claude-work -I "*.json|README.md|WORK_HISTORY.md|test-*" 2>/dev/null || find .claude-work -type d | sort

echo ""
echo "✅ Migration completed!"
echo "📊 Summary:"
echo "   - Work items: $(find .claude-work/work -name "*.md" | wc -l)"
echo "   - Artifacts: $(find .claude-work/artifacts -name "*.md" | wc -l)"
echo "   - Groups: $(find .claude-work/groups -name "*.md" | wc -l)"