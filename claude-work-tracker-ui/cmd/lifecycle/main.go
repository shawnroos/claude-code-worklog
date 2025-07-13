package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"claude-work-tracker-ui/internal/data"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: lifecycle <command> [options]")
		fmt.Println("Commands:")
		fmt.Println("  analyze      - Analyze system decay and health")
		fmt.Println("  health       - Show system health metrics")
		fmt.Println("  cleanup      - Interactive cleanup mode")
		fmt.Println("  auto-cleanup - Execute all auto-safe cleanup actions")
		fmt.Println("  refresh      - Refresh all activity scores")
		os.Exit(1)
	}

	// Initialize managers
	client := data.NewEnhancedClient()
	associationMgr := client.GetAssociationManager()
	groupMgr := client.GetGroupManager()
	lifecycleMgr := data.NewLifecycleManager(client.GetMarkdownIO(), associationMgr, groupMgr)

	switch os.Args[1] {
	case "analyze":
		runAnalyze(lifecycleMgr)
	case "health":
		runHealth(lifecycleMgr)
	case "cleanup":
		runInteractiveCleanup(lifecycleMgr)
	case "auto-cleanup":
		runAutoCleanup(lifecycleMgr)
	case "refresh":
		runRefresh(lifecycleMgr)
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func runAnalyze(lm *data.LifecycleManager) {
	fmt.Println("🔍 Analyzing system decay and health...")
	
	analysis, err := lm.AnalyzeDecay()
	if err != nil {
		log.Fatalf("Failed to analyze decay: %v", err)
	}

	// Print summary
	fmt.Printf("\n📊 System Health Summary\n")
	fmt.Printf("═══════════════════════════\n")
	fmt.Printf("Overall Health Score: %.1f%% %s\n", 
		analysis.Summary.OverallHealthScore*100, 
		getHealthEmoji(analysis.Summary.OverallHealthScore))
	fmt.Printf("Total Items: %d\n", analysis.Summary.TotalItems)
	fmt.Printf("Healthy Items: %d\n", analysis.Summary.HealthyItems)
	fmt.Printf("Items Needing Review: %d\n", analysis.Summary.ItemsNeedingReview)
	fmt.Printf("Items Needing Action: %d\n", analysis.Summary.ItemsNeedingAction)

	// Print specific issues
	if len(analysis.OrphanedArtifacts) > 0 {
		fmt.Printf("\n🏝️  Orphaned Artifacts (%d)\n", len(analysis.OrphanedArtifacts))
		fmt.Printf("────────────────────────────\n")
		for _, artifact := range analysis.OrphanedArtifacts {
			fmt.Printf("• [%s] %s (%.0f days old)\n", 
				artifact.Type, 
				artifact.Summary, 
				daysSince(artifact.CreatedAt))
		}
	}

	if len(analysis.StaleWork) > 0 {
		fmt.Printf("\n💤 Stale Work Items (%d)\n", len(analysis.StaleWork))
		fmt.Printf("─────────────────────────\n")
		for _, work := range analysis.StaleWork {
			activityDays := "unknown"
			if work.Metadata.LastActivityAt != nil {
				activityDays = fmt.Sprintf("%.0f", daysSince(*work.Metadata.LastActivityAt))
			}
			fmt.Printf("• [%s] %s (last activity: %s days ago)\n", 
				work.Schedule, 
				work.Title, 
				activityDays)
		}
	}

	if len(analysis.UnsupportedWork) > 0 {
		fmt.Printf("\n🚫 Unsupported Work Items (%d)\n", len(analysis.UnsupportedWork))
		fmt.Printf("──────────────────────────────\n")
		for _, work := range analysis.UnsupportedWork {
			fmt.Printf("• [%s] %s (no artifacts)\n", work.Schedule, work.Title)
		}
	}

	if len(analysis.StaleGroups) > 0 {
		fmt.Printf("\n📦 Stale Groups (%d)\n", len(analysis.StaleGroups))
		fmt.Printf("─────────────────────\n")
		for _, group := range analysis.StaleGroups {
			fmt.Printf("• %s (%d artifacts, %.1f%% ready)\n", 
				group.Name, 
				group.Metadata.ArtifactCount,
				group.Metadata.ReadinessScore*100)
		}
	}

	// Print recommended actions
	if len(analysis.RecommendedActions) > 0 {
		fmt.Printf("\n💡 Recommended Actions (%d)\n", len(analysis.RecommendedActions))
		fmt.Printf("═══════════════════════════\n")
		
		highPriority := 0
		autoSafe := 0
		
		for i, action := range analysis.RecommendedActions {
			if i >= 10 { // Limit display to first 10
				fmt.Printf("... and %d more actions\n", len(analysis.RecommendedActions)-10)
				break
			}
			
			priority := getPriorityEmoji(action.Priority)
			autoIcon := ""
			if action.AutoSafe {
				autoIcon = "🤖"
				autoSafe++
			}
			if action.Priority == "high" {
				highPriority++
			}
			
			fmt.Printf("%d. %s %s [%s] %s %s\n", 
				i+1, priority, autoIcon, action.Type, action.Reason, action.ItemType)
			fmt.Printf("   %s\n", action.Details)
		}
		
		fmt.Printf("\n📈 Action Summary:\n")
		fmt.Printf("   High Priority: %d\n", highPriority)
		fmt.Printf("   Auto-Safe: %d\n", autoSafe)
	}

	if analysis.Summary.OverallHealthScore < 0.7 {
		fmt.Printf("\n⚠️  System health is below optimal. Consider running cleanup.\n")
	} else if len(analysis.RecommendedActions) == 0 {
		fmt.Printf("\n✅ System is healthy! No cleanup actions needed.\n")
	}

	fmt.Printf("\n💡 Next steps:\n")
	fmt.Printf("   • Use 'lifecycle cleanup' for interactive cleanup\n")
	
	// Count auto-safe actions
	autoSafeCount := 0
	for _, action := range analysis.RecommendedActions {
		if action.AutoSafe {
			autoSafeCount++
		}
	}
	
	if autoSafeCount > 0 {
		fmt.Printf("   • Use 'lifecycle auto-cleanup' to execute %d auto-safe actions\n", autoSafeCount)
	}
	fmt.Printf("   • Use 'lifecycle health' for ongoing monitoring\n")
}

func runHealth(lm *data.LifecycleManager) {
	fmt.Println("🏥 Checking system health...")
	
	metrics, err := lm.GetHealthMetrics()
	if err != nil {
		log.Fatalf("Failed to get health metrics: %v", err)
	}

	fmt.Printf("\n📊 System Health Metrics\n")
	fmt.Printf("═══════════════════════════\n")
	fmt.Printf("Overall Health: %.1f%% %s\n", 
		metrics.OverallHealth*100, 
		getHealthEmoji(metrics.OverallHealth))
	fmt.Printf("Total Items: %d\n", metrics.TotalItems)
	fmt.Printf("Healthy Items: %d\n", metrics.HealthyItems)
	fmt.Printf("Problematic Items: %d\n", metrics.ProblematicItems)
	fmt.Printf("Pending Actions: %d\n", metrics.PendingActions)
	fmt.Printf("High Priority Actions: %d\n", metrics.HighPriorityActions)
	fmt.Printf("Auto-Safe Actions: %d\n", metrics.AutoSafeActions)
	fmt.Printf("Health Trend: %s %s\n", metrics.HealthTrend, getTrendEmoji(metrics.HealthTrend))
	fmt.Printf("Last Analyzed: %s\n", metrics.LastAnalyzed.Format("2006-01-02 15:04:05"))

	// Health recommendations
	fmt.Printf("\n💡 Recommendations:\n")
	if metrics.OverallHealth > 0.9 {
		fmt.Printf("   ✅ Excellent health! Keep up the good work.\n")
	} else if metrics.OverallHealth > 0.7 {
		fmt.Printf("   👍 Good health. Monitor regularly.\n")
	} else if metrics.OverallHealth > 0.5 {
		fmt.Printf("   ⚠️  Fair health. Consider cleanup actions.\n")
	} else {
		fmt.Printf("   🚨 Poor health. Immediate cleanup recommended!\n")
	}

	if metrics.HighPriorityActions > 0 {
		fmt.Printf("   🔥 %d high priority actions need attention\n", metrics.HighPriorityActions)
	}

	if metrics.AutoSafeActions > 0 {
		fmt.Printf("   🤖 %d actions can be safely auto-executed\n", metrics.AutoSafeActions)
	}
}

func runInteractiveCleanup(lm *data.LifecycleManager) {
	fmt.Println("🧹 Interactive cleanup mode")
	fmt.Println("Finding cleanup opportunities...")
	
	analysis, err := lm.AnalyzeDecay()
	if err != nil {
		log.Fatalf("Failed to analyze decay: %v", err)
	}

	if len(analysis.RecommendedActions) == 0 {
		fmt.Println("✅ No cleanup actions needed. System is healthy!")
		return
	}

	fmt.Printf("\nFound %d recommended actions.\n\n", len(analysis.RecommendedActions))
	
	reader := bufio.NewReader(os.Stdin)
	executed := 0
	skipped := 0

	for i, action := range analysis.RecommendedActions {
		priority := getPriorityEmoji(action.Priority)
		autoIcon := ""
		if action.AutoSafe {
			autoIcon = "🤖 "
		}

		fmt.Printf("Action %d/%d: %s %s[%s] %s\n", 
			i+1, len(analysis.RecommendedActions), priority, autoIcon, action.Type, action.Reason)
		fmt.Printf("Item: %s (%s)\n", action.ItemID, action.ItemType)
		fmt.Printf("Details: %s\n", action.Details)
		
		if action.AutoSafe {
			fmt.Print("\nExecute this auto-safe action? [Y/n/s(kip all)/q(uit)]: ")
		} else {
			fmt.Print("\nExecute this action? [y/N/s(kip all)/q(uit)]: ")
		}
		
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		
		switch response {
		case "y", "yes":
			fmt.Printf("🔄 Executing action...\n")
			if err := lm.ExecuteCleanupAction(action); err != nil {
				fmt.Printf("❌ Failed: %v\n", err)
			} else {
				fmt.Printf("✅ Action completed successfully\n")
				executed++
			}
		case "s", "skip":
			fmt.Printf("⏭️  Skipping remaining %d actions\n", len(analysis.RecommendedActions)-i)
			skipped += len(analysis.RecommendedActions) - i
			break
		case "q", "quit":
			fmt.Println("👋 Cleanup cancelled")
			return
		case "":
			if action.AutoSafe {
				// Default to yes for auto-safe actions
				fmt.Printf("🔄 Executing auto-safe action...\n")
				if err := lm.ExecuteCleanupAction(action); err != nil {
					fmt.Printf("❌ Failed: %v\n", err)
				} else {
					fmt.Printf("✅ Action completed successfully\n")
					executed++
				}
			} else {
				fmt.Printf("⏭️  Skipped\n")
				skipped++
			}
		default:
			fmt.Printf("⏭️  Skipped\n")
			skipped++
		}
		fmt.Println()
	}

	fmt.Printf("🎉 Cleanup completed!\n")
	fmt.Printf("   Actions executed: %d\n", executed)
	fmt.Printf("   Actions skipped: %d\n", skipped)
}

func runAutoCleanup(lm *data.LifecycleManager) {
	fmt.Println("🤖 Running automatic cleanup...")
	
	result, err := lm.AutoCleanup()
	if err != nil {
		log.Fatalf("Failed to run auto cleanup: %v", err)
	}

	fmt.Printf("\n%s\n", result.Summary)
	
	if len(result.ActionsExecuted) > 0 {
		fmt.Printf("\n✅ Successfully executed %d actions:\n", len(result.ActionsExecuted))
		for _, action := range result.ActionsExecuted {
			fmt.Printf("   • [%s] %s - %s\n", action.Type, action.ItemID, action.Reason)
		}
	}
	
	if len(result.ActionsFailed) > 0 {
		fmt.Printf("\n❌ Failed to execute %d actions:\n", len(result.ActionsFailed))
		for _, failure := range result.ActionsFailed {
			fmt.Printf("   • [%s] %s - %s (Error: %s)\n", 
				failure.Action.Type, failure.Action.ItemID, failure.Action.Reason, failure.Error)
		}
	}

	if len(result.ActionsExecuted) == 0 && len(result.ActionsFailed) == 0 {
		fmt.Println("✅ No auto-safe actions found. System is clean!")
	}
}

func runRefresh(lm *data.LifecycleManager) {
	fmt.Println("🔄 Refreshing all activity scores...")
	
	if err := lm.RefreshAllActivityScores(); err != nil {
		log.Fatalf("Failed to refresh activity scores: %v", err)
	}
	
	fmt.Println("✅ Activity scores refreshed successfully")
	fmt.Println("💡 Run 'lifecycle analyze' to see updated decay analysis")
}

// Helper functions
func getHealthEmoji(score float64) string {
	if score > 0.9 {
		return "🟢"
	} else if score > 0.7 {
		return "🟡"
	} else if score > 0.5 {
		return "🟠"
	} else {
		return "🔴"
	}
}

func getPriorityEmoji(priority string) string {
	switch priority {
	case "high":
		return "🔥"
	case "medium":
		return "⚠️"
	case "low":
		return "💡"
	default:
		return "📋"
	}
}

func getTrendEmoji(trend string) string {
	switch trend {
	case "improving":
		return "📈"
	case "declining":
		return "📉"
	case "stable":
		return "📊"
	default:
		return "❓"
	}
}

func daysSince(t time.Time) float64 {
	return time.Since(t).Hours() / 24
}