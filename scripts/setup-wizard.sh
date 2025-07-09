#!/bin/bash

# Claude Code Work Tracking Setup Wizard
# Helps users configure the system after installation

CONFIG_FILE="$HOME/.claude/work-tracking-config.json"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}🎛️  Claude Code Work Tracking Setup Wizard${NC}"
echo -e "${BLUE}===========================================${NC}"
echo ""

if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${RED}❌ Configuration file not found. Please run the installer first.${NC}"
    exit 1
fi

echo -e "${GREEN}Welcome to the work tracking system!${NC}"
echo "This wizard will help you customize your experience."
echo ""

# Function to update config
update_config() {
    local key="$1"
    local value="$2"
    
    python3 -c "
import json
with open('$CONFIG_FILE', 'r') as f:
    config = json.load(f)

# Navigate to nested key
keys = '$key'.split('.')
current = config
for k in keys[:-1]:
    if k not in current:
        current[k] = {}
    current = current[k]
current[keys[-1]] = '$value'

with open('$CONFIG_FILE', 'w') as f:
    json.dump(config, f, indent=2)
"
}

# Presentation Mode Selection
echo -e "${CYAN}📺 Presentation Mode${NC}"
echo "How much feedback would you like during work sessions?"
echo ""
echo "1) Quiet     - Minimal output, no session summaries"
echo "2) Summary   - Balanced feedback with session summaries (recommended)"
echo "3) Verbose   - Detailed feedback with full context"
echo ""
while true; do
    read -p "Choose mode (1-3): " mode_choice
    case $mode_choice in
        1) update_config "presentation.mode" "quiet"; echo -e "${GREEN}✓ Set to quiet mode${NC}"; break;;
        2) update_config "presentation.mode" "summary"; echo -e "${GREEN}✓ Set to summary mode${NC}"; break;;
        3) update_config "presentation.mode" "verbose"; echo -e "${GREEN}✓ Set to verbose mode${NC}"; break;;
        *) echo "Please choose 1, 2, or 3.";;
    esac
done

echo ""

# Emoji Style Selection
echo -e "${CYAN}🎨 Visual Style${NC}"
echo "How would you like status indicators to appear?"
echo ""
echo "1) Minimal + Colors  - ✓ ○ ● ! with colors (recommended)"
echo "2) Modern Emojis     - ✅ 🔄 ⚡ ⚠️"
echo "3) Classic Brackets  - [✓] [○] [●] [!]"
echo "4) Minimal Plain     - ✓ ○ ● ! (no colors)"
echo ""
while true; do
    read -p "Choose style (1-4): " style_choice
    case $style_choice in
        1) update_config "presentation.emoji_style" "minimal_colored"; echo -e "${GREEN}✓ Set to minimal colored${NC}"; break;;
        2) update_config "presentation.emoji_style" "modern"; echo -e "${GREEN}✓ Set to modern emojis${NC}"; break;;
        3) update_config "presentation.emoji_style" "classic"; echo -e "${GREEN}✓ Set to classic brackets${NC}"; break;;
        4) update_config "presentation.emoji_style" "minimal"; echo -e "${GREEN}✓ Set to minimal plain${NC}"; break;;
        *) echo "Please choose 1, 2, 3, or 4.";;
    esac
done

echo ""

# Feature Toggles
echo -e "${CYAN}🔧 Feature Settings${NC}"
echo ""

# Session summaries
echo -n "Show session summaries when work completes? (Y/n): "
read -r summary_choice
if [[ $summary_choice =~ ^[Nn]$ ]]; then
    update_config "presentation.show_session_summary" "false"
    echo -e "${YELLOW}✓ Session summaries disabled${NC}"
else
    update_config "presentation.show_session_summary" "true"
    echo -e "${GREEN}✓ Session summaries enabled${NC}"
fi

# Conflict alerts
echo -n "Show alerts when related work exists in other worktrees? (Y/n): "
read -r conflict_choice
if [[ $conflict_choice =~ ^[Nn]$ ]]; then
    update_config "presentation.show_conflicts_alert" "false"
    echo -e "${YELLOW}✓ Conflict alerts disabled${NC}"
else
    update_config "presentation.show_conflicts_alert" "true"
    echo -e "${GREEN}✓ Conflict alerts enabled${NC}"
fi

# Sync notifications
echo -n "Show sync notifications during background updates? (Y/n): "
read -r sync_choice
if [[ $sync_choice =~ ^[Nn]$ ]]; then
    update_config "presentation.show_notifications" "false"
    echo -e "${YELLOW}✓ Sync notifications disabled${NC}"
else
    update_config "presentation.show_notifications" "true"
    echo -e "${GREEN}✓ Sync notifications enabled${NC}"
fi

echo ""

# Test the configuration
echo -e "${CYAN}🧪 Testing Configuration${NC}"
echo "Let's see how your settings look:"
echo ""

~/.claude/scripts/work-presentation.sh test

echo ""

# Configuration summary
echo -e "${BLUE}📋 Configuration Summary${NC}"
echo -e "${BLUE}========================${NC}"

current_mode=$(jq -r '.presentation.mode' "$CONFIG_FILE")
current_style=$(jq -r '.presentation.emoji_style' "$CONFIG_FILE")

echo "Mode: $current_mode"
echo "Style: $current_style"
echo "Config file: $CONFIG_FILE"

echo ""
echo -e "${GREEN}🎉 Setup Complete!${NC}"
echo ""
echo -e "${BLUE}Next Steps:${NC}"
echo "• Your next Claude session will automatically use these settings"
echo "• Change settings anytime: ~/.claude/scripts/work-presentation.sh mode [quiet|summary|verbose]"
echo "• Check work status: ~/.claude/scripts/work-status.sh"
echo "• View all commands: ~/.claude/scripts/work-status.sh"
echo ""
echo -e "${CYAN}💡 Pro Tips:${NC}"
echo "• Use 'quiet' mode during focused coding sessions"
echo "• Switch to 'verbose' when debugging or analyzing cross-worktree work"
echo "• The system works automatically - no need to remember commands!"
echo ""
echo -e "${GREEN}Happy tracking! 🎯${NC}"