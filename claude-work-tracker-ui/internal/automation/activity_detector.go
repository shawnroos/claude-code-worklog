package automation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"claude-work-tracker-ui/internal/hooks"
	"claude-work-tracker-ui/internal/models"
)

// ActivityDetector monitors and analyzes work item activity patterns
type ActivityDetector struct {
	mu              sync.RWMutex
	activityLog     map[string][]ActivityEvent
	focusSessions   map[string]*FocusSession
	config          *ActivityConfig
	hookSystem      *hooks.HookSystem
}

// ActivityEvent represents a single activity on a work item
type ActivityEvent struct {
	WorkID    string
	Timestamp time.Time
	EventType string // save, commit, progress_update, status_change
	Metadata  map[string]interface{}
}

// FocusSession represents a period of concentrated work
type FocusSession struct {
	WorkID      string
	StartTime   time.Time
	LastActivity time.Time
	EventCount  int
	Intensity   float64 // Events per minute
}

// ActivityConfig contains configuration for activity detection
type ActivityConfig struct {
	FocusThresholdMinutes   int     // Minutes between events to consider same session
	FocusMinEvents          int     // Minimum events to qualify as focus session
	HighIntensityThreshold  float64 // Events per minute for high intensity
	InactivityWarningHours  int     // Hours before warning about inactivity
	AutoPromoteOnFocus      bool    // Auto-promote to NOW when focus detected
}

// DefaultActivityConfig returns default configuration
func DefaultActivityConfig() *ActivityConfig {
	return &ActivityConfig{
		FocusThresholdMinutes:   10,
		FocusMinEvents:          3,
		HighIntensityThreshold:  0.5, // 1 event per 2 minutes
		InactivityWarningHours:  48,
		AutoPromoteOnFocus:      false, // Require confirmation
	}
}

// NewActivityDetector creates a new activity detector
func NewActivityDetector(hookSystem *hooks.HookSystem, config *ActivityConfig) *ActivityDetector {
	if config == nil {
		config = DefaultActivityConfig()
	}

	detector := &ActivityDetector{
		activityLog:   make(map[string][]ActivityEvent),
		focusSessions: make(map[string]*FocusSession),
		config:        config,
		hookSystem:    hookSystem,
	}

	// Register activity tracking hooks
	detector.registerHooks()

	return detector
}

// RecordActivity logs an activity event
func (ad *ActivityDetector) RecordActivity(workID string, eventType string, metadata map[string]interface{}) {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	event := ActivityEvent{
		WorkID:    workID,
		Timestamp: time.Now(),
		EventType: eventType,
		Metadata:  metadata,
	}

	// Add to activity log
	ad.activityLog[workID] = append(ad.activityLog[workID], event)

	// Update or create focus session
	ad.updateFocusSession(workID, event)
}

// updateFocusSession updates or creates a focus session based on activity
func (ad *ActivityDetector) updateFocusSession(workID string, event ActivityEvent) {
	session, exists := ad.focusSessions[workID]

	if !exists || time.Since(session.LastActivity).Minutes() > float64(ad.config.FocusThresholdMinutes) {
		// Start new session
		ad.focusSessions[workID] = &FocusSession{
			WorkID:       workID,
			StartTime:    event.Timestamp,
			LastActivity: event.Timestamp,
			EventCount:   1,
			Intensity:    0,
		}
	} else {
		// Update existing session
		session.LastActivity = event.Timestamp
		session.EventCount++
		duration := session.LastActivity.Sub(session.StartTime).Minutes()
		if duration > 0 {
			session.Intensity = float64(session.EventCount) / duration
		}
	}
}

// GetFocusSession returns the current focus session for a work item
func (ad *ActivityDetector) GetFocusSession(workID string) (*FocusSession, bool) {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	session, exists := ad.focusSessions[workID]
	if !exists {
		return nil, false
	}

	// Check if session is still active
	if time.Since(session.LastActivity).Minutes() > float64(ad.config.FocusThresholdMinutes) {
		return nil, false
	}

	return session, true
}

// AnalyzeActivity analyzes activity patterns for a work item
func (ad *ActivityDetector) AnalyzeActivity(workID string) map[string]interface{} {
	ad.mu.RLock()
	defer ad.mu.RUnlock()

	analysis := make(map[string]interface{})
	events := ad.activityLog[workID]

	if len(events) == 0 {
		analysis["has_activity"] = false
		return analysis
	}

	// Basic statistics
	analysis["has_activity"] = true
	analysis["total_events"] = len(events)
	analysis["first_activity"] = events[0].Timestamp
	analysis["last_activity"] = events[len(events)-1].Timestamp

	// Activity patterns
	var saveCount, commitCount, progressCount int
	for _, event := range events {
		switch event.EventType {
		case "save":
			saveCount++
		case "commit":
			commitCount++
		case "progress_update":
			progressCount++
		}
	}

	analysis["save_count"] = saveCount
	analysis["commit_count"] = commitCount
	analysis["progress_updates"] = progressCount

	// Focus session info
	if session, exists := ad.focusSessions[workID]; exists {
		analysis["has_focus_session"] = true
		analysis["focus_intensity"] = session.Intensity
		analysis["focus_duration"] = time.Since(session.StartTime).Minutes()
		analysis["is_high_intensity"] = session.Intensity >= ad.config.HighIntensityThreshold
	} else {
		analysis["has_focus_session"] = false
	}

	// Inactivity check
	daysSinceActivity := time.Since(events[len(events)-1].Timestamp).Hours() / 24
	analysis["days_since_activity"] = daysSinceActivity
	analysis["is_inactive"] = daysSinceActivity*24 > float64(ad.config.InactivityWarningHours)

	return analysis
}

// SuggestTransitions suggests transitions based on activity patterns
func (ad *ActivityDetector) SuggestTransitions(work *models.Work) []TransitionSuggestion {
	var suggestions []TransitionSuggestion

	analysis := ad.AnalyzeActivity(work.ID)

	// Check for focus session
	if hasFocus, _ := analysis["has_focus_session"].(bool); hasFocus {
		if work.Schedule != "now" {
			intensity, _ := analysis["focus_intensity"].(float64)
			if intensity >= ad.config.HighIntensityThreshold {
				suggestions = append(suggestions, TransitionSuggestion{
					Type:        "schedule_change",
					Target:      "now",
					Reason:      "High-intensity focus session detected",
					Confidence:  0.9,
					AutoApply:   ad.config.AutoPromoteOnFocus,
				})
			}
		}

		// Suggest status change based on activity
		if work.Metadata.Status == "draft" || work.Metadata.Status == "active" {
			suggestions = append(suggestions, TransitionSuggestion{
				Type:       "status_change",
				Target:     "in_progress",
				Reason:     "Active development detected",
				Confidence: 0.8,
				AutoApply:  true,
			})
		}
	}

	// Check for inactivity
	if isInactive, _ := analysis["is_inactive"].(bool); isInactive {
		if work.Schedule == "now" && work.Metadata.Status == "in_progress" {
			daysSince, _ := analysis["days_since_activity"].(float64)
			suggestions = append(suggestions, TransitionSuggestion{
				Type:       "schedule_change",
				Target:     "next",
				Reason:     fmt.Sprintf("No activity for %.1f days", daysSince),
				Confidence: 0.7,
				AutoApply:  false, // Always confirm demotion
			})
		}
	}

	return suggestions
}

// TransitionSuggestion represents a suggested transition
type TransitionSuggestion struct {
	Type       string  // schedule_change, status_change
	Target     string  // Target schedule or status
	Reason     string  // Human-readable reason
	Confidence float64 // 0-1 confidence score
	AutoApply  bool    // Whether to apply automatically
}

// registerHooks registers activity tracking hooks
func (ad *ActivityDetector) registerHooks() {
	// Track all status changes
	ad.hookSystem.Register(hooks.AfterStatusChange, "activity_tracker", func(ctx context.Context, hookCtx *hooks.HookContext) error {
		ad.RecordActivity(hookCtx.WorkItem.ID, "status_change", map[string]interface{}{
			"old_status": hookCtx.OldWorkItem.Metadata.Status,
			"new_status": hookCtx.WorkItem.Metadata.Status,
		})
		return nil
	})

	// Track progress updates
	ad.hookSystem.Register(hooks.ProgressUpdated, "progress_tracker", func(ctx context.Context, hookCtx *hooks.HookContext) error {
		ad.RecordActivity(hookCtx.WorkItem.ID, "progress_update", hookCtx.Metadata)
		return nil
	})

	// Track Git commits
	ad.hookSystem.Register(hooks.CommitDetected, "commit_tracker", func(ctx context.Context, hookCtx *hooks.HookContext) error {
		ad.RecordActivity(hookCtx.WorkItem.ID, "commit", hookCtx.Metadata)
		return nil
	})

	// Periodic inactivity check
	ad.hookSystem.Register(hooks.ActivityDetected, "focus_detector", func(ctx context.Context, hookCtx *hooks.HookContext) error {
		// Check if this creates or extends a focus session
		if session, exists := ad.GetFocusSession(hookCtx.WorkItem.ID); exists {
			if session.EventCount >= ad.config.FocusMinEvents && session.Intensity >= ad.config.HighIntensityThreshold {
				// Trigger focus mode hook
				focusCtx := &hooks.HookContext{
					WorkItem:  hookCtx.WorkItem,
					EventType: hooks.ActivityDetected,
					Timestamp: time.Now(),
					Metadata: map[string]interface{}{
						"focus_session": session,
						"intensity":     session.Intensity,
					},
				}
				ad.hookSystem.Execute(ctx, focusCtx)
			}
		}
		return nil
	})
}

// GetInactiveWorkItems returns work items that have been inactive
func (ad *ActivityDetector) GetInactiveWorkItems(workItems []models.Work) []models.Work {
	var inactive []models.Work

	for _, work := range workItems {
		analysis := ad.AnalyzeActivity(work.ID)
		if isInactive, _ := analysis["is_inactive"].(bool); isInactive {
			inactive = append(inactive, work)
		}
	}

	return inactive
}

// CleanupOldActivity removes activity logs older than specified days
func (ad *ActivityDetector) CleanupOldActivity(days int) {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -days)

	for workID, events := range ad.activityLog {
		var filtered []ActivityEvent
		for _, event := range events {
			if event.Timestamp.After(cutoff) {
				filtered = append(filtered, event)
			}
		}

		if len(filtered) > 0 {
			ad.activityLog[workID] = filtered
		} else {
			delete(ad.activityLog, workID)
		}
	}

	// Clean up old focus sessions
	for workID, session := range ad.focusSessions {
		if session.LastActivity.Before(cutoff) {
			delete(ad.focusSessions, workID)
		}
	}
}