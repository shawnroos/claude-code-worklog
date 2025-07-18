#!/bin/bash

# Work Item Migration Tool
# Helps organize work items based on their status

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Find project root
PROJECT_ROOT=$(git rev-parse --show-toplevel 2>/dev/null || pwd)

echo -e "${GREEN}Work Item Migration Tool${NC}"
echo "========================"
echo ""

# Function to migrate completed items to closed
migrate_completed_to_closed() {
    echo -e "${YELLOW}Migrating completed items to CLOSED...${NC}"
    
    # Find all .claude-work directories
    find "$PROJECT_ROOT" -type d -name ".claude-work" | while read -r claude_dir; do
        work_dir="$claude_dir/work"
        
        if [ ! -d "$work_dir" ]; then
            continue
        fi
        
        # Check each schedule directory
        for schedule in now next later; do
            schedule_dir="$work_dir/$schedule"
            if [ ! -d "$schedule_dir" ]; then
                continue
            fi
            
            # Create closed directory if needed
            closed_dir="$work_dir/closed"
            mkdir -p "$closed_dir"
            
            # Find completed/canceled items
            find "$schedule_dir" -name "*.md" -type f | while read -r file; do
                # Check if file contains completed/canceled status
                if grep -q "status: completed\|status: canceled\|status: archived" "$file"; then
                    basename=$(basename "$file")
                    echo "  Moving $basename from $schedule to closed"
                    mv "$file" "$closed_dir/"
                fi
            done
        done
    done
    
    echo -e "${GREEN}✓ Completed items migrated${NC}"
}

# Function to move inactive items from NOW to NEXT
migrate_inactive_to_next() {
    echo -e "${YELLOW}Moving inactive items from NOW to NEXT...${NC}"
    
    find "$PROJECT_ROOT" -type d -name ".claude-work" | while read -r claude_dir; do
        now_dir="$claude_dir/work/now"
        next_dir="$claude_dir/work/next"
        
        if [ ! -d "$now_dir" ]; then
            continue
        fi
        
        mkdir -p "$next_dir"
        
        # Find items with 0% progress or draft status
        find "$now_dir" -name "*.md" -type f | while read -r file; do
            # Check for 0% progress or draft/pending status
            if grep -q "progress_percent: 0\|status: draft\|status: pending" "$file"; then
                # Also check that it's not in_progress
                if ! grep -q "status: in_progress" "$file"; then
                    basename=$(basename "$file")
                    echo "  Moving $basename from now to next (0% progress)"
                    mv "$file" "$next_dir/"
                fi
            fi
        done
    done
    
    echo -e "${GREEN}✓ Inactive items moved to NEXT${NC}"
}

# Function to archive old closed items
archive_old_closed() {
    echo -e "${YELLOW}Archiving old closed items (>30 days)...${NC}"
    
    find "$PROJECT_ROOT" -type d -name ".claude-work" | while read -r claude_dir; do
        closed_dir="$claude_dir/work/closed"
        archive_dir="$claude_dir/archive"
        
        if [ ! -d "$closed_dir" ]; then
            continue
        fi
        
        mkdir -p "$archive_dir"
        
        # Find files older than 30 days
        find "$closed_dir" -name "*.md" -type f -mtime +30 | while read -r file; do
            basename=$(basename "$file")
            echo "  Archiving $basename (>30 days old)"
            mv "$file" "$archive_dir/"
        done
    done
    
    echo -e "${GREEN}✓ Old items archived${NC}"
}

# Function to show current distribution
show_distribution() {
    echo -e "${YELLOW}Current work item distribution:${NC}"
    echo ""
    
    find "$PROJECT_ROOT" -type d -name ".claude-work" | while read -r claude_dir; do
        work_dir="$claude_dir/work"
        # macOS compatible relative path
        rel_path="${claude_dir#$PROJECT_ROOT/}"
        
        if [ ! -d "$work_dir" ]; then
            continue
        fi
        
        echo "  $rel_path:"
        
        for schedule in now next later closed; do
            count=$(find "$work_dir/$schedule" -name "*.md" -type f 2>/dev/null | wc -l)
            printf "    %-8s: %d items\n" "$schedule" "$count"
        done
        echo ""
    done
}

# Main menu
case "${1:-menu}" in
    completed)
        migrate_completed_to_closed
        ;;
    inactive)
        migrate_inactive_to_next
        ;;
    archive)
        archive_old_closed
        ;;
    all)
        migrate_completed_to_closed
        echo ""
        migrate_inactive_to_next
        echo ""
        show_distribution
        ;;
    status)
        show_distribution
        ;;
    *)
        echo "Usage: $0 {completed|inactive|archive|all|status}"
        echo ""
        echo "Commands:"
        echo "  completed  - Move completed/canceled items to CLOSED"
        echo "  inactive   - Move 0% progress items from NOW to NEXT"
        echo "  archive    - Archive closed items older than 30 days"
        echo "  all        - Run all migrations"
        echo "  status     - Show current distribution"
        echo ""
        echo "Example:"
        echo "  $0 all      # Run all migrations"
        echo "  $0 status   # Just show current state"
        ;;
esac