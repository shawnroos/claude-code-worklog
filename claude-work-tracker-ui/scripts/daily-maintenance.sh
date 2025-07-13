#!/bin/bash

# Daily maintenance script for the work tracking system
# Can be run via cron for automated maintenance

WORK_DIR="${WORK_DIR:-.claude-work}"
LOG_FILE="${WORK_DIR}/maintenance.log"

echo "🔧 Starting daily maintenance - $(date)" | tee -a "$LOG_FILE"

# Ensure lifecycle tool is built
if [ ! -f "./lifecycle" ]; then
    echo "🔨 Building lifecycle tool..." | tee -a "$LOG_FILE"
    ./build-lifecycle.sh >> "$LOG_FILE" 2>&1
    if [ $? -ne 0 ]; then
        echo "❌ Failed to build lifecycle tool" | tee -a "$LOG_FILE"
        exit 1
    fi
fi

# Refresh activity scores
echo "🔄 Refreshing activity scores..." | tee -a "$LOG_FILE"
./lifecycle refresh >> "$LOG_FILE" 2>&1

# Run health check
echo "🏥 Checking system health..." | tee -a "$LOG_FILE"
HEALTH_OUTPUT=$(./lifecycle health 2>&1)
echo "$HEALTH_OUTPUT" | tee -a "$LOG_FILE"

# Extract health score (if available)
HEALTH_SCORE=$(echo "$HEALTH_OUTPUT" | grep -o "Overall Health: [0-9.]*%" | grep -o "[0-9.]*")

if [ ! -z "$HEALTH_SCORE" ]; then
    # Convert percentage to decimal for comparison
    HEALTH_DECIMAL=$(echo "$HEALTH_SCORE / 100" | bc -l 2>/dev/null || echo "0.8")
    
    # If health is below 70%, run auto-cleanup
    if (( $(echo "$HEALTH_DECIMAL < 0.7" | bc -l) )); then
        echo "⚠️  Health below 70% ($HEALTH_SCORE%), running auto-cleanup..." | tee -a "$LOG_FILE"
        ./lifecycle auto-cleanup >> "$LOG_FILE" 2>&1
    else
        echo "✅ Health is good ($HEALTH_SCORE%)" | tee -a "$LOG_FILE"
    fi
else
    echo "⚠️  Could not determine health score, skipping auto-cleanup" | tee -a "$LOG_FILE"
fi

# Generate summary report
echo "📊 Generating summary..." | tee -a "$LOG_FILE"
ANALYSIS_OUTPUT=$(./lifecycle analyze 2>&1)

# Count key metrics
ORPHANED_COUNT=$(echo "$ANALYSIS_OUTPUT" | grep -o "Orphaned Artifacts ([0-9]*)" | grep -o "[0-9]*" || echo "0")
STALE_WORK_COUNT=$(echo "$ANALYSIS_OUTPUT" | grep -o "Stale Work Items ([0-9]*)" | grep -o "[0-9]*" || echo "0")
RECOMMENDED_ACTIONS=$(echo "$ANALYSIS_OUTPUT" | grep -o "Recommended Actions ([0-9]*)" | grep -o "[0-9]*" || echo "0")

echo "📋 Daily Summary:" | tee -a "$LOG_FILE"
echo "   Health Score: ${HEALTH_SCORE:-Unknown}%" | tee -a "$LOG_FILE"
echo "   Orphaned Artifacts: $ORPHANED_COUNT" | tee -a "$LOG_FILE"
echo "   Stale Work Items: $STALE_WORK_COUNT" | tee -a "$LOG_FILE"
echo "   Recommended Actions: $RECOMMENDED_ACTIONS" | tee -a "$LOG_FILE"

# Alert if there are serious issues
if [ "$ORPHANED_COUNT" -gt 10 ] || [ "$STALE_WORK_COUNT" -gt 5 ] || [ "$RECOMMENDED_ACTIONS" -gt 15 ]; then
    echo "🚨 ALERT: System needs attention!" | tee -a "$LOG_FILE"
    echo "   Consider running: ./lifecycle cleanup" | tee -a "$LOG_FILE"
fi

echo "✅ Daily maintenance completed - $(date)" | tee -a "$LOG_FILE"
echo "────────────────────────────────────────" >> "$LOG_FILE"