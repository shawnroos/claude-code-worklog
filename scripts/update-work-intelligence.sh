#!/bin/bash

# Work Intelligence Aggregation Script
# Updates global work intelligence from local work intelligence captures

PROJECT_DIR="$1"

if [ -z "$PROJECT_DIR" ]; then
    echo "Usage: $0 <project_dir>"
    exit 1
fi

# Determine project name from directory
PROJECT_NAME=$(basename "$PROJECT_DIR")
WORK_INTELLIGENCE_DIR="$HOME/.claude/work-intelligence"
GLOBAL_STATE_DIR="$HOME/.claude/work-state"
PROJECT_STATE_DIR="$GLOBAL_STATE_DIR/projects/$PROJECT_NAME"

# Create directories
mkdir -p "$PROJECT_STATE_DIR"

# Update work intelligence aggregation
INTELLIGENCE_OVERVIEW="$PROJECT_STATE_DIR/WORK_INTELLIGENCE.md"

cat > "$INTELLIGENCE_OVERVIEW" << EOF
# $PROJECT_NAME - Work Intelligence Overview

*Last updated: $(date)*

## Recent Plans and Proposals

EOF

# Process recent work intelligence files
if [ -d "$WORK_INTELLIGENCE_DIR" ]; then
    # Get files from last 7 days
    find "$WORK_INTELLIGENCE_DIR" -name "*.json" -type f -mtime -7 | \
    while read -r file; do
        if [ -f "$file" ]; then
            INTELLIGENCE_TYPE=$(jq -r '.type' "$file" 2>/dev/null || echo "unknown")
            CONTENT=$(jq -r '.content' "$file" 2>/dev/null || echo "")
            TIMESTAMP=$(jq -r '.timestamp' "$file" 2>/dev/null || echo "")
            CONTEXT=$(jq -r '.context' "$file" 2>/dev/null || echo "")
            GIT_BRANCH=$(jq -r '.git_branch' "$file" 2>/dev/null || echo "")
            
            # Filter by current project
            WORKING_DIR=$(jq -r '.working_directory' "$file" 2>/dev/null || echo "")
            if [[ "$WORKING_DIR" == *"$PROJECT_NAME"* ]]; then
                case "$INTELLIGENCE_TYPE" in
                    "plan")
                        echo "### 📋 Plan - $TIMESTAMP" >> "$INTELLIGENCE_OVERVIEW"
                        echo "**Branch:** $GIT_BRANCH | **Context:** $CONTEXT" >> "$INTELLIGENCE_OVERVIEW"
                        echo "" >> "$INTELLIGENCE_OVERVIEW"
                        echo "$CONTENT" >> "$INTELLIGENCE_OVERVIEW"
                        echo "" >> "$INTELLIGENCE_OVERVIEW"
                        ;;
                    "proposal")
                        echo "### 💡 Proposal - $TIMESTAMP" >> "$INTELLIGENCE_OVERVIEW"
                        echo "**Branch:** $GIT_BRANCH | **Context:** $CONTEXT" >> "$INTELLIGENCE_OVERVIEW"
                        echo "" >> "$INTELLIGENCE_OVERVIEW"
                        echo "$CONTENT" >> "$INTELLIGENCE_OVERVIEW"
                        echo "" >> "$INTELLIGENCE_OVERVIEW"
                        ;;
                    "strategic_insight")
                        echo "### 🎯 Strategic Insight - $TIMESTAMP" >> "$INTELLIGENCE_OVERVIEW"
                        echo "**Branch:** $GIT_BRANCH | **Context:** $CONTEXT" >> "$INTELLIGENCE_OVERVIEW"
                        echo "" >> "$INTELLIGENCE_OVERVIEW"
                        echo "$CONTENT" >> "$INTELLIGENCE_OVERVIEW"
                        echo "" >> "$INTELLIGENCE_OVERVIEW"
                        ;;
                    "decision_rationale")
                        echo "### ⚖️ Decision Rationale - $TIMESTAMP" >> "$INTELLIGENCE_OVERVIEW"
                        echo "**Branch:** $GIT_BRANCH | **Context:** $CONTEXT" >> "$INTELLIGENCE_OVERVIEW"
                        echo "" >> "$INTELLIGENCE_OVERVIEW"
                        echo "$CONTENT" >> "$INTELLIGENCE_OVERVIEW"
                        echo "" >> "$INTELLIGENCE_OVERVIEW"
                        ;;
                    "session_summary")
                        echo "### 📝 Session Summary - $TIMESTAMP" >> "$INTELLIGENCE_OVERVIEW"
                        echo "**Branch:** $GIT_BRANCH | **Context:** $CONTEXT" >> "$INTELLIGENCE_OVERVIEW"
                        echo "" >> "$INTELLIGENCE_OVERVIEW"
                        echo "$CONTENT" >> "$INTELLIGENCE_OVERVIEW"
                        echo "" >> "$INTELLIGENCE_OVERVIEW"
                        ;;
                esac
            fi
        fi
    done
fi

# Update global project overview with intelligence stats
GLOBAL_OVERVIEW="$GLOBAL_STATE_DIR/PROJECT_OVERVIEW.md"

if [ -f "$GLOBAL_OVERVIEW" ]; then
    # Count intelligence items for this project
    PLAN_COUNT=$(find "$WORK_INTELLIGENCE_DIR" -name "*.json" -type f -mtime -7 -exec jq -r 'select(.type == "plan") | .working_directory' {} \; | grep -c "$PROJECT_NAME" || echo "0")
    PROPOSAL_COUNT=$(find "$WORK_INTELLIGENCE_DIR" -name "*.json" -type f -mtime -7 -exec jq -r 'select(.type == "proposal") | .working_directory' {} \; | grep -c "$PROJECT_NAME" || echo "0")
    INSIGHT_COUNT=$(find "$WORK_INTELLIGENCE_DIR" -name "*.json" -type f -mtime -7 -exec jq -r 'select(.type == "strategic_insight") | .working_directory' {} \; | grep -c "$PROJECT_NAME" || echo "0")
    
    # Update project section with intelligence stats
    PROJECT_SECTION_START="## $PROJECT_NAME"
    
    if grep -q "^$PROJECT_SECTION_START" "$GLOBAL_OVERVIEW"; then
        # Add intelligence stats to existing project section
        sed -i '' "/^$PROJECT_SECTION_START/,/^## / {
            /^## /!{
                /^\*\*Total worktrees/a\\
**Recent plans:** $PLAN_COUNT | **Proposals:** $PROPOSAL_COUNT | **Insights:** $INSIGHT_COUNT
            }
        }" "$GLOBAL_OVERVIEW"
    fi
fi

echo "[$(date)] Work intelligence updated for $PROJECT_NAME" >> "$HOME/.claude/hooks.log"