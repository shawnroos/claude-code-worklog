package automation

import (
	"context"
	"fmt"
	"time"

	"claude-work-tracker-ui/internal/hooks"
	"claude-work-tracker-ui/internal/models"
)

// TransitionRule defines a rule for automatic status transitions
type TransitionRule struct {
	Name        string
	Description string
	Condition   func(w *models.Work) bool
	Action      func(w *models.Work) *models.Work
	Priority    int // Higher priority rules are evaluated first
}

// TransitionEngine manages automatic transitions based on rules
type TransitionEngine struct {
	rules      []TransitionRule
	hookSystem *hooks.HookSystem
	config     *TransitionConfig
}

// TransitionConfig contains configuration for the transition engine
type TransitionConfig struct {
	Enabled                   bool
	RequireUserConfirmation   bool // For NOW transitions
	StaleThresholdDays        int
	InactivityThresholdHours  int
	AutoArchiveThresholdDays  int
}

// DefaultTransitionConfig returns default configuration
func DefaultTransitionConfig() *TransitionConfig {
	return &TransitionConfig{
		Enabled:                   true,
		RequireUserConfirmation:   true,
		StaleThresholdDays:        7,
		InactivityThresholdHours:  48,
		AutoArchiveThresholdDays:  90,
	}
}

// NewTransitionEngine creates a new transition engine
func NewTransitionEngine(hookSystem *hooks.HookSystem, config *TransitionConfig) *TransitionEngine {
	if config == nil {
		config = DefaultTransitionConfig()
	}

	engine := &TransitionEngine{
		hookSystem: hookSystem,
		config:     config,
	}

	// Initialize default rules
	engine.initializeDefaultRules()

	return engine
}

// initializeDefaultRules sets up the standard transition rules
func (te *TransitionEngine) initializeDefaultRules() {
	te.rules = []TransitionRule{
		// Draft to Active when progress starts
		{
			Name:        "draft_to_active",
			Description: "Move draft items to active when progress > 0",
			Priority:    100,
			Condition: func(w *models.Work) bool {
				return w.Metadata.Status == "draft" && w.Metadata.ProgressPercent > 0
			},
			Action: func(w *models.Work) *models.Work {
				w.Metadata.Status = "active"
				// Add automation metadata - would need to extend WorkMetadata struct
				// For now, just update the basic fields
				return w
			},
		},

		// Active to In Progress when significant progress
		{
			Name:        "active_to_in_progress",
			Description: "Move active items to in_progress when progress > 20%",
			Priority:    90,
			Condition: func(w *models.Work) bool {
				return w.Metadata.Status == "active" && w.Metadata.ProgressPercent > 20
			},
			Action: func(w *models.Work) *models.Work {
				w.Metadata.Status = "in_progress"
				// Auto-transition flag would be added to WorkMetadata struct
				// Transition reason would be added to WorkMetadata struct: "Significant progress made"
				return w
			},
		},

		// In Progress to Completed when progress = 100%
		{
			Name:        "in_progress_to_completed",
			Description: "Complete items when progress reaches 100%",
			Priority:    80,
			Condition: func(w *models.Work) bool {
				return w.Metadata.Status == "in_progress" && w.Metadata.ProgressPercent >= 100
			},
			Action: func(w *models.Work) *models.Work {
				w.Metadata.Status = "completed"
				w.Schedule = "closed"
				now := time.Now()
			w.CompletedAt = &now
				// Auto-transition flag would be added to WorkMetadata struct
				// Transition reason would be added to WorkMetadata struct: "Progress reached 100%"
				return w
			},
		},

		// Stale In Progress to Blocked
		{
			Name:        "stale_to_blocked",
			Description: "Mark stale in_progress items as blocked",
			Priority:    70,
			Condition: func(w *models.Work) bool {
				if w.Metadata.Status != "in_progress" || w.Schedule != "now" {
					return false
				}
				daysSinceActivity := time.Since(*w.Metadata.LastActivityAt).Hours() / 24
				return daysSinceActivity > float64(te.config.StaleThresholdDays)
			},
			Action: func(w *models.Work) *models.Work {
				w.Metadata.Status = "blocked"
				// Auto-transition flag would be added to WorkMetadata struct
				// Transition reason would be added to WorkMetadata struct: fmt.Sprintf("No activity for %d days", te.config.StaleThresholdDays)
				// Previous status would be added to WorkMetadata struct: "in_progress"
				return w
			},
		},

		// Auto-archive very old closed items
		{
			Name:        "auto_archive_old",
			Description: "Archive very old closed items",
			Priority:    60,
			Condition: func(w *models.Work) bool {
				if w.Schedule != "closed" || w.Metadata.Status == "archived" {
					return false
				}
				if w.CompletedAt == nil {
					return false
				}
				daysSinceCompleted := time.Since(*w.CompletedAt).Hours() / 24
				return daysSinceCompleted > float64(te.config.AutoArchiveThresholdDays)
			},
			Action: func(w *models.Work) *models.Work {
				w.Metadata.Status = "archived"
				// Auto-transition flag would be added to WorkMetadata struct
				// Transition reason would be added to WorkMetadata struct: fmt.Sprintf("Completed %d days ago", te.config.AutoArchiveThresholdDays)
				// Archived timestamp would be added to WorkMetadata struct: time.Now().Format(time.RFC3339)
				return w
			},
		},

		// Blocked items with resolved dependencies
		{
			Name:        "unblock_resolved",
			Description: "Unblock items when dependencies are resolved",
			Priority:    85,
			Condition: func(w *models.Work) bool {
				if w.Metadata.Status != "blocked" || len(w.Metadata.BlockedBy) == 0 {
					return false
				}
				// This would need access to other work items to check if blockers are resolved
				// For now, we'll skip this rule and implement it later with proper context
				return false
			},
			Action: func(w *models.Work) *models.Work {
				// Previous status would be stored in WorkMetadata field
				w.Metadata.Status = "active" // Simplified for now
				// Auto-transition flag would be added to WorkMetadata struct
				// Transition reason would be added to WorkMetadata struct: "Dependencies resolved"
				// Previous status field would be cleared in WorkMetadata
				return w
			},
		},
	}
}

// EvaluateWork checks if any transition rules apply to a work item
func (te *TransitionEngine) EvaluateWork(ctx context.Context, work *models.Work) (*models.Work, bool, error) {
	if !te.config.Enabled {
		return work, false, nil
	}

	// Sort rules by priority (highest first)
	for _, rule := range te.rules {
		if rule.Condition(work) {
			// Create a copy to preserve original for hooks
			oldWork := *work

			// Apply the transition
			newWork := rule.Action(work)

			// Special handling for NOW transitions
			if oldWork.Schedule != "now" && newWork.Schedule == "now" && te.config.RequireUserConfirmation {
				// Mark for confirmation instead of auto-transitioning
				// Would need to add pending_transition field to WorkMetadata
				return newWork, false, nil
			}

			// Execute hooks
			hookCtx := &hooks.HookContext{
				WorkItem:    newWork,
				OldWorkItem: &oldWork,
				EventType:   hooks.AfterStatusChange,
				Timestamp:   time.Now(),
				Metadata: map[string]interface{}{
					"rule_name":        rule.Name,
					"rule_description": rule.Description,
				},
			}

			_, err := te.hookSystem.Execute(ctx, hookCtx)
			if err != nil {
				return work, false, fmt.Errorf("hook execution failed: %w", err)
			}

			return newWork, true, nil
		}
	}

	return work, false, nil
}

// AddRule adds a custom transition rule
func (te *TransitionEngine) AddRule(rule TransitionRule) {
	te.rules = append(te.rules, rule)
	// Re-sort by priority
	te.sortRulesByPriority()
}

// sortRulesByPriority sorts rules by priority (highest first)
func (te *TransitionEngine) sortRulesByPriority() {
	// Simple insertion sort for small number of rules
	for i := 1; i < len(te.rules); i++ {
		j := i
		for j > 0 && te.rules[j].Priority > te.rules[j-1].Priority {
			te.rules[j], te.rules[j-1] = te.rules[j-1], te.rules[j]
			j--
		}
	}
}

// GetPendingTransition checks if a work item has a pending transition
func (te *TransitionEngine) GetPendingTransition(work *models.Work) (string, bool) {
	// This would require adding a PendingTransition field to WorkMetadata
	// For now, return false to indicate no pending transitions
	return "", false
}

// ConfirmTransition applies a pending transition
func (te *TransitionEngine) ConfirmTransition(work *models.Work) bool {
	// This would require adding pending transition fields to WorkMetadata
	// For now, return false to indicate no pending transitions to confirm
	return false
}