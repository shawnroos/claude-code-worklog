#!/bin/bash

# Migrate scattered work items to centralized storage at project root
# This script consolidates all .claude-work directories into a single location

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔄 Claude Work Storage Migration Tool${NC}"
echo "======================================"
echo ""

# Function to find project root
find_project_root() {
    local current="$PWD"
    
    while [[ "$current" != "/" ]]; do
        # Check for .git directory (not file, to avoid submodules)
        if [[ -d "$current/.git" ]]; then
            echo "$current"
            return 0
        fi
        
        # Check for .git file (worktree)
        if [[ -f "$current/.git" ]]; then
            # Extract main repo from worktree
            local gitdir=$(grep "gitdir:" "$current/.git" | cut -d' ' -f2)
            if [[ -n "$gitdir" && "$gitdir" == *"/.git/worktrees/"* ]]; then
                # Extract main repo path
                echo "${gitdir%/.git/worktrees/*}"
                return 0
            fi
        fi
        
        # Check for other project indicators
        for indicator in "package.json" "go.mod" "Cargo.toml" "pyproject.toml"; do
            if [[ -f "$current/$indicator" ]]; then
                echo "$current"
                return 0
            fi
        done
        
        current=$(dirname "$current")
    done
    
    # Fallback to current directory
    echo "$PWD"
}

# Get project root
PROJECT_ROOT=$(find_project_root)
CENTRAL_WORK_DIR="$PROJECT_ROOT/.claude-work"

echo -e "Project root: ${GREEN}$PROJECT_ROOT${NC}"
echo -e "Central storage: ${GREEN}$CENTRAL_WORK_DIR${NC}"
echo ""

# Create backup directory with timestamp
BACKUP_DIR="$PROJECT_ROOT/.claude-work-backup-$(date +%Y%m%d_%H%M%S)"
echo -e "${YELLOW}Creating backup at: $BACKUP_DIR${NC}"
mkdir -p "$BACKUP_DIR"

# Find all .claude-work directories
echo -e "\n${BLUE}Scanning for distributed .claude-work directories...${NC}"
WORK_DIRS=$(find "$PROJECT_ROOT" -type d -name ".claude-work" 2>/dev/null | grep -v "$BACKUP_DIR" || true)

if [[ -z "$WORK_DIRS" ]]; then
    echo -e "${YELLOW}No .claude-work directories found.${NC}"
    exit 0
fi

# Count work items
TOTAL_ITEMS=0
DUPLICATES=0
# Use a simple approach for older bash versions
WORK_ITEMS_FILE="/tmp/claude_work_items_$$"
> "$WORK_ITEMS_FILE"

echo -e "\nFound .claude-work directories:"
while IFS= read -r dir; do
    if [[ "$dir" == "$CENTRAL_WORK_DIR" ]]; then
        echo -e "  ${GREEN}✓${NC} $dir (central - will preserve)"
    else
        echo -e "  ${YELLOW}⚠${NC}  $dir (distributed - will migrate)"
    fi
    
    # Count work items in this directory
    if [[ -d "$dir/work" ]]; then
        count=$(find "$dir/work" -name "*.md" -type f 2>/dev/null | wc -l || echo 0)
        echo -e "      Items: $count"
        TOTAL_ITEMS=$((TOTAL_ITEMS + count))
    fi
done <<< "$WORK_DIRS"

echo -e "\nTotal work items found: ${BLUE}$TOTAL_ITEMS${NC}"

# Step 1: Backup everything
echo -e "\n${YELLOW}Step 1: Creating full backup...${NC}"
while IFS= read -r dir; do
    relative_path="${dir#$PROJECT_ROOT/}"
    backup_path="$BACKUP_DIR/$relative_path"
    mkdir -p "$(dirname "$backup_path")"
    cp -r "$dir" "$backup_path"
    echo -e "  Backed up: $relative_path"
done <<< "$WORK_DIRS"

# Step 2: Analyze for duplicates
echo -e "\n${YELLOW}Step 2: Analyzing for duplicates...${NC}"
while IFS= read -r dir; do
    if [[ -d "$dir/work" ]]; then
        find "$dir/work" -name "*.md" -type f 2>/dev/null | while read -r file; do
            # Extract ID from filename or content
            basename=$(basename "$file")
            id="${basename%.md}"
            
            # Check if we've seen this ID before
            existing=$(grep "^$id|" "$WORK_ITEMS_FILE" | cut -d'|' -f2)
            if [[ -n "$existing" ]]; then
                echo -e "  ${YELLOW}Duplicate found:${NC} $id"
                echo -e "    Original: $existing"
                echo -e "    Duplicate: $file"
                DUPLICATES=$((DUPLICATES + 1))
                
                # Compare timestamps to keep the newer one
                orig_time=$(stat -f %m "$existing" 2>/dev/null || echo 0)
                dup_time=$(stat -f %m "$file" 2>/dev/null || echo 0)
                
                if [[ $dup_time -gt $orig_time ]]; then
                    echo -e "    ${GREEN}Keeping newer version from: $file${NC}"
                    # Update the entry
                    grep -v "^$id|" "$WORK_ITEMS_FILE" > "$WORK_ITEMS_FILE.tmp"
                    mv "$WORK_ITEMS_FILE.tmp" "$WORK_ITEMS_FILE"
                    echo "$id|$file" >> "$WORK_ITEMS_FILE"
                fi
            else
                echo "$id|$file" >> "$WORK_ITEMS_FILE"
            fi
        done
    fi
done <<< "$WORK_DIRS"

echo -e "Found ${YELLOW}$DUPLICATES${NC} duplicate items"

# Step 3: Create central structure
echo -e "\n${YELLOW}Step 3: Creating central storage structure...${NC}"
mkdir -p "$CENTRAL_WORK_DIR"/{work/{now,next,later,closed},artifacts,updates,groups}
echo -e "  ${GREEN}✓${NC} Created central directory structure"

# Step 4: Migrate work items
echo -e "\n${YELLOW}Step 4: Migrating work items to central storage...${NC}"
MIGRATED=0

while IFS='|' read -r id source_file; do
    
    # Determine schedule from path
    schedule="later" # default
    if [[ "$source_file" == */now/* ]]; then
        schedule="now"
    elif [[ "$source_file" == */next/* ]]; then
        schedule="next"
    elif [[ "$source_file" == */later/* ]]; then
        schedule="later"
    elif [[ "$source_file" == */closed/* ]]; then
        schedule="closed"
    fi
    
    # Target path
    target_file="$CENTRAL_WORK_DIR/work/$schedule/$(basename "$source_file")"
    
    # Copy to central location
    if [[ "$source_file" != "$target_file" ]]; then
        cp "$source_file" "$target_file"
        echo -e "  ${GREEN}✓${NC} Migrated: $id → $schedule"
        MIGRATED=$((MIGRATED + 1))
        
        # Update git context in the file
        if grep -q "working_directory:" "$target_file"; then
            # Update working_directory to project root
            sed -i.bak "s|working_directory:.*|working_directory: $PROJECT_ROOT|" "$target_file"
            rm -f "$target_file.bak"
        fi
    fi
done < "$WORK_ITEMS_FILE"

echo -e "\nMigrated ${GREEN}$MIGRATED${NC} work items"

# Clean up temp file
rm -f "$WORK_ITEMS_FILE"

# Step 5: Clean up distributed directories
echo -e "\n${YELLOW}Step 5: Cleaning up distributed directories...${NC}"
echo -e "${RED}This will remove distributed .claude-work directories!${NC}"
echo -n "Proceed with cleanup? (y/N): "
read -r response

if [[ "$response" =~ ^[Yy]$ ]]; then
    while IFS= read -r dir; do
        if [[ "$dir" != "$CENTRAL_WORK_DIR" ]]; then
            echo -e "  Removing: $dir"
            rm -rf "$dir"
        fi
    done <<< "$WORK_DIRS"
    echo -e "${GREEN}✓ Cleanup completed${NC}"
else
    echo -e "${YELLOW}Skipped cleanup. Distributed directories remain.${NC}"
    echo -e "To manually remove later, delete these directories:"
    while IFS= read -r dir; do
        if [[ "$dir" != "$CENTRAL_WORK_DIR" ]]; then
            echo "  rm -rf $dir"
        fi
    done <<< "$WORK_DIRS"
fi

# Summary
echo -e "\n${GREEN}═══════════════════════════════════════${NC}"
echo -e "${GREEN}Migration Summary:${NC}"
echo -e "${GREEN}═══════════════════════════════════════${NC}"
echo -e "  Total items found: $TOTAL_ITEMS"
echo -e "  Duplicates resolved: $DUPLICATES"
echo -e "  Items migrated: $MIGRATED"
echo -e "  Backup location: $BACKUP_DIR"
echo -e "  Central storage: $CENTRAL_WORK_DIR"
echo -e "\n${GREEN}✅ Migration completed successfully!${NC}"

# Create a migration report
REPORT_FILE="$CENTRAL_WORK_DIR/migration-report-$(date +%Y%m%d_%H%M%S).txt"
cat > "$REPORT_FILE" << EOF
Claude Work Storage Migration Report
Generated: $(date)

Project Root: $PROJECT_ROOT
Central Storage: $CENTRAL_WORK_DIR
Backup Location: $BACKUP_DIR

Summary:
- Total items found: $TOTAL_ITEMS
- Duplicates resolved: $DUPLICATES  
- Items migrated: $MIGRATED

Directories processed:
$(echo "$WORK_DIRS")

Migration completed successfully.
EOF

echo -e "\nMigration report saved to: ${BLUE}$REPORT_FILE${NC}"