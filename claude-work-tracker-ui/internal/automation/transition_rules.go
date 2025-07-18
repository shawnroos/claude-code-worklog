package automation

import (
	"context"
	"fmt"
	"time"

	"github.com/shawnroos/claude-work-tracker-ui/internal/hooks"
	"github.com/shawnroos/claude-work-tracker-ui/internal/model"
)

// TransitionRule defines a rule for automatic status transitions
type TransitionRule struct {
	Name        string
	Description string
	Condition   func(w *model.Work) bool
	Action      func(w *model.Work) *model.Work
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
			Condition: func(w *model.Work) bool {
				return w.Status == "draft" && w.Progress > 0
			},
			Action: func(w *model.Work) *model.Work {
				w.Status = "active"
				w.Metadata["auto_transitioned"] = true
				w.Metadata["transition_reason"] = "Progress started"
				return w
			},
		},

		// Active to In Progress when significant progress
		{
			Name:        "active_to_in_progress",
			Description: "Move active items to in_progress when progress > 20%",
			Priority:    90,
			Condition: func(w *model.Work) bool {
				return w.Status == "active" && w.Progress > 20
			},
			Action: func(w *model.Work) *model.Work {
				w.Status = "in_progress"
				w.Metadata["auto_transitioned"] = true
				w.Metadata["transition_reason"] = "Significant progress made"
				return w
			},
		},

		// In Progress to Completed when progress = 100%
		{
			Name:        "in_progress_to_completed",
			Description: "Complete items when progress reaches 100%",
			Priority:    80,
			Condition: func(w *model.Work) bool {
				return w.Status == "in_progress" && w.Progress >= 100
			},
			Action: func(w *model.Work) *model.Work {
				w.Status = "completed"
				w.Schedule = "closed"
				w.CompletedAt = time.Now()
				w.Metadata["auto_transitioned"] = true
				w.Metadata["transition_reason"] = "Progress reached 100%"
				return w
			},
		},

		// Stale In Progress to Blocked
		{
			Name:        "stale_to_blocked",
			Description: "Mark stale in_progress items as blocked",
			Priority:    70,
			Condition: func(w *model.Work) bool {
				if w.Status != "in_progress" || w.Schedule != "now" {
					return false
				}
				daysSinceActivity := time.Since(w.LastActivityAt).Hours() / 24
				return daysSinceActivity > float64(te.config.StaleThresholdDays)
			},
			Action: func(w *model.Work) *model.Work {
				w.Status = "blocked"
				w.Metadata["auto_transitioned"] = true
				w.Metadata["transition_reason"] = fmt.Sprintf("No activity for %d days", te.config.StaleThresholdDays)
				w.Metadata["previous_status"] = "in_progress"
				return w
			},
		},

		// Auto-archive very old closed items
		{
			Name:        "auto_archive_old",
			Description: "Archive very old closed items",
			Priority:    60,
			Condition: func(w *model.Work) bool {
				if w.Schedule != "closed" || w.Status == "archived" {
					return false
				}
				if w.CompletedAt.IsZero() {
					return false
				}
				daysSinceCompleted := time.Since(w.CompletedAt).Hours() / 24
				return daysSinceCompleted > float64(te.config.AutoArchiveThresholdDays)
			},
			Action: func(w *model.Work) *model.Work {
				w.Status = "archived"
				w.Metadata["auto_transitioned"] = true
				w.Metadata["transition_reason"] = fmt.Sprintf("Completed %d days ago", te.config.AutoArchiveThresholdDays)
				w.Metadata["archived_at"] = time.Now().Format(time.RFC3339)
				return w
			},
		},

		// Blocked items with resolved dependencies
		{
			Name:        "unblock_resolved",
			Description: "Unblock items when dependencies are resolved",
			Priority:    85,
			Condition: func(w *model.Work) bool {
				if w.Status != "blocked" || len(w.BlockedBy) == 0 {
					return false
				}
				// This would need access to other work items to check if blockers are resolved
				// For now, we'll skip this rule and implement it later with proper context
				return false
			},
			Action: func(w *model.Work) *model.Work {
				if prevStatus, ok := w.Metadata["previous_status"].(string); ok {
					w.Status = prevStatus
				} else {
					w.Status = "active"
				}
				w.Metadata["auto_transitioned"] = true
				w.Metadata["transition_reason"] = "Dependencies resolved"
				delete(w.Metadata, "previous_status")
				return w
			},
		},
	}
}

// EvaluateWork checks if any transition rules apply to a work item
func (te *TransitionEngine) EvaluateWork(ctx context.Context, work *model.Work) (*model.Work, bool, error) {
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
				newWork.Metadata["pending_transition"] = "now"
				newWork.Metadata["transition_rule"] = rule.Name
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
func (te *TransitionEngine) GetPendingTransition(work *model.Work) (string, bool) {
	if pending, ok := work.Metadata["pending_transition"].(string); ok {
		return pending, true
	}
	return "", false
}

// ConfirmTransition applies a pending transition
func (te *TransitionEngine) ConfirmTransition(work *model.Work) bool {
	if pending, ok := work.Metadata["pending_transition"].(string); ok {
		if pending == "now" {
			work.Schedule = "now"
		}
		delete(work.Metadata, "pending_transition")
		delete(work.Metadata, "transition_rule")
		return true
	}
	return false
}