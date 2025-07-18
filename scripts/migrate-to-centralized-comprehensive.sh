#!/bin/bash

# Comprehensive migration to centralized storage
# This script finds ALL work items and migrates them to ~/.claude/work-data

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔄 Comprehensive Work Item Migration${NC}"
echo "========================================"
echo ""

# Get the project ID from the centralized storage
PROJECT_ID="3f1749c3cf60da7a"  # claude-work-tracker project
CENTRAL_WORK_DIR="$HOME/.claude/work-data/work/$PROJECT_ID"

echo -e "Target storage: ${GREEN}$CENTRAL_WORK_DIR${NC}"
echo ""

# Create central directories if they don't exist
mkdir -p "$CENTRAL_WORK_DIR"/{now,next,later,closed}

# Create backup directory
BACKUP_DIR="/tmp/claude-work-migration-backup-$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"
echo -e "${YELLOW}Backup location: $BACKUP_DIR${NC}"
echo ""

# Find all work items
echo -e "${BLUE}Scanning for work items...${NC}"
WORK_ITEMS=$(find /Users/shawnroos/claude-work-tracker /Users/shawnroos/worktrees \
    -name "*.md" -path "*/.claude-work/work/*" -type f 2>/dev/null | sort)

TOTAL_ITEMS=$(echo "$WORK_ITEMS" | grep -c . || echo 0)
echo -e "Found ${GREEN}$TOTAL_ITEMS${NC} work items to migrate"
echo ""

# Track migration statistics
MIGRATED=0
DUPLICATES=0
ERRORS=0
# Use a file to track seen IDs for bash 3 compatibility
SEEN_IDS_FILE="/tmp/claude_migration_seen_$$"
> "$SEEN_IDS_FILE"

# Function to extract ID from markdown file
extract_work_id() {
    local file="$1"
    # Try to extract from frontmatter first
    local id=$(grep -A1 "^id:" "$file" 2>/dev/null | head -1 | sed 's/id: *//' | tr -d '"' | tr -d "'")
    
    # If no ID in frontmatter, use filename
    if [[ -z "$id" || "$id" == "id:" ]]; then
        id=$(basename "$file" .md)
    fi
    
    echo "$id"
}

# Function to determine schedule from path or content
determine_schedule() {
    local file="$1"
    local path="$file"
    
    # Check path first
    if [[ "$path" == */now/* ]]; then
        echo "now"
    elif [[ "$path" == */next/* ]]; then
        echo "next"
    elif [[ "$path" == */later/* ]]; then
        echo "later"
    elif [[ "$path" == */closed/* ]]; then
        echo "closed"
    elif [[ "$path" == */completed/* ]]; then
        echo "closed"
    else
        # Check frontmatter
        local schedule=$(grep "^schedule:" "$file" 2>/dev/null | head -1 | awk '{print $2}' | tr -d '"' | tr -d "'")
        if [[ -n "$schedule" ]]; then
            echo "$schedule"
        else
            echo "later" # Default
        fi
    fi
}

# Process each work item
echo -e "${BLUE}Processing work items...${NC}"
echo ""

while IFS= read -r file; do
    if [[ -z "$file" ]]; then
        continue
    fi
    
    # Extract work ID
    work_id=$(extract_work_id "$file")
    
    # Determine schedule
    schedule=$(determine_schedule "$file")
    
    # Display info
    echo -e "${CYAN}Processing:${NC} $(basename "$file")"
    echo -e "  ID: $work_id"
    echo -e "  Schedule: $schedule"
    echo -e "  Source: ${file#/Users/shawnroos/}"
    
    # Check for duplicates
    existing_file=$(grep "^$work_id|" "$SEEN_IDS_FILE" | cut -d'|' -f2)
    if [[ -n "$existing_file" ]]; then
        echo -e "  ${YELLOW}⚠ Duplicate ID found${NC}"
        
        # Compare timestamps
        existing_time=$(stat -f %m "$existing_file" 2>/dev/null || echo 0)
        new_time=$(stat -f %m "$file" 2>/dev/null || echo 0)
        
        if [[ $new_time -gt $existing_time ]]; then
            echo -e "  ${GREEN}→ Using newer version${NC}"
            # Backup the old one
            cp "$existing_file" "$BACKUP_DIR/$(basename "$existing_file").old"
            # Update the entry
            grep -v "^$work_id|" "$SEEN_IDS_FILE" > "$SEEN_IDS_FILE.tmp"
            mv "$SEEN_IDS_FILE.tmp" "$SEEN_IDS_FILE"
        else
            echo -e "  ${YELLOW}→ Keeping existing version${NC}"
            # Backup this one
            cp "$file" "$BACKUP_DIR/$(basename "$file").duplicate"
            DUPLICATES=$((DUPLICATES + 1))
            echo ""
            continue
        fi
    fi
    
    # Target path
    target_dir="$CENTRAL_WORK_DIR/$schedule"
    target_file="$target_dir/$(basename "$file")"
    
    # Backup original
    cp "$file" "$BACKUP_DIR/"
    
    # Copy to central location
    if cp "$file" "$target_file" 2>/dev/null; then
        echo -e "  ${GREEN}✓ Migrated to $schedule${NC}"
        echo "$work_id|$target_file" >> "$SEEN_IDS_FILE"
        MIGRATED=$((MIGRATED + 1))
        
        # Update project_id in the file if missing
        if ! grep -q "project_id:" "$target_file"; then
            # Add project_id to git_context
            sed -i.bak '/git_context:/,/^[^ ]/ {
                /working_directory:/ a\
  project_id: '"$PROJECT_ID"'
            }' "$target_file" 2>/dev/null || true
            rm -f "$target_file.bak"
        fi
    else
        echo -e "  ${RED}✗ Migration failed${NC}"
        ERRORS=$((ERRORS + 1))
    fi
    
    echo ""
done <<< "$WORK_ITEMS"

# Find and report any work items in the backup directories
echo -e "${BLUE}Checking backup directories...${NC}"
BACKUP_ITEMS=$(find /Users/shawnroos/claude-work-tracker/.claude-work-backup-* \
    -name "*.md" -path "*/work/*" -type f 2>/dev/null | wc -l || echo 0)

if [[ $BACKUP_ITEMS -gt 0 ]]; then
    echo -e "${YELLOW}Note: Found $BACKUP_ITEMS work items in backup directories${NC}"
    echo -e "These were already processed in the main migration."
fi

echo ""

# Summary
echo -e "${GREEN}════════════════════════════════════════${NC}"
echo -e "${GREEN}Migration Summary${NC}"
echo -e "${GREEN}════════════════════════════════════════${NC}"
echo -e "Total items found: $TOTAL_ITEMS"
echo -e "Successfully migrated: ${GREEN}$MIGRATED${NC}"
echo -e "Duplicates resolved: ${YELLOW}$DUPLICATES${NC}"
echo -e "Errors: ${RED}$ERRORS${NC}"
echo -e ""
echo -e "Central storage: $CENTRAL_WORK_DIR"
echo -e "Backup location: $BACKUP_DIR"

# Verify final count
FINAL_COUNT=$(find "$CENTRAL_WORK_DIR" -name "*.md" -type f | wc -l)
echo -e ""
echo -e "Work items in central storage: ${GREEN}$FINAL_COUNT${NC}"

# Show distribution
echo -e ""
echo -e "${BLUE}Distribution by schedule:${NC}"
for schedule in now next later closed; do
    count=$(find "$CENTRAL_WORK_DIR/$schedule" -name "*.md" -type f 2>/dev/null | wc -l || echo 0)
    printf "  %-8s: %d\n" "$schedule" "$count"
done

echo -e ""
echo -e "${GREEN}✅ Migration completed!${NC}"

# Create migration report
REPORT_FILE="$CENTRAL_WORK_DIR/../migration-report-$(date +%Y%m%d_%H%M%S).txt"
cat > "$REPORT_FILE" << EOF
Claude Work Migration Report
Generated: $(date)

Project ID: $PROJECT_ID
Central Storage: $CENTRAL_WORK_DIR

Migration Statistics:
- Total items found: $TOTAL_ITEMS
- Successfully migrated: $MIGRATED
- Duplicates resolved: $DUPLICATES
- Errors: $ERRORS

Final work item count: $FINAL_COUNT

Distribution:
$(for schedule in now next later closed; do
    count=$(find "$CENTRAL_WORK_DIR/$schedule" -name "*.md" -type f 2>/dev/null | wc -l || echo 0)
    printf "  %-8s: %d\n" "$schedule" "$count"
done)

Backup location: $BACKUP_DIR
EOF

echo -e ""
echo -e "Migration report saved to: ${BLUE}$REPORT_FILE${NC}"

# Offer to clean up old directories
echo -e ""
echo -e "${YELLOW}Clean up old .claude-work directories?${NC}"
echo -e "This will remove:"
find /Users/shawnroos/claude-work-tracker /Users/shawnroos/worktrees \
    -name ".claude-work" -type d 2>/dev/null | grep -v backup | head -10

echo -e ""
echo -n "Proceed with cleanup? (y/N): "
read -r response

if [[ "$response" =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Cleaning up old directories...${NC}"
    
    # Remove .claude-work directories (excluding backups)
    find /Users/shawnroos/claude-work-tracker /Users/shawnroos/worktrees \
        -name ".claude-work" -type d 2>/dev/null | \
        grep -v backup | \
        while read -r dir; do
            echo -e "  Removing: $dir"
            rm -rf "$dir"
        done
    
    echo -e "${GREEN}✓ Cleanup completed${NC}"
else
    echo -e "${YELLOW}Skipped cleanup${NC}"
fi

# Clean up temp file
rm -f "$SEEN_IDS_FILE"