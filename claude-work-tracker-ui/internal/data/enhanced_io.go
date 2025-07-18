package data

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"claude-work-tracker-ui/internal/automation"
	"claude-work-tracker-ui/internal/git"
	"claude-work-tracker-ui/internal/hooks"
	"claude-work-tracker-ui/internal/models"
)

// EnhancedMarkdownIO extends MarkdownIO with automation capabilities
type EnhancedMarkdownIO struct {
	*MarkdownIO
	hookSystem      *hooks.HookSystem
	transitionEngine *automation.TransitionEngine
	gitManager      *git.ContextManager
	config          *EnhancedIOConfig
}

// EnhancedIOConfig contains configuration for enhanced IO
type EnhancedIOConfig struct {
	EnableAutomation bool
	EnableGitTracking bool
	GenerateUpdates  bool
}

// DefaultEnhancedIOConfig returns default configuration
func DefaultEnhancedIOConfig() *EnhancedIOConfig {
	return &EnhancedIOConfig{
		EnableAutomation: true,
		EnableGitTracking: true,
		GenerateUpdates: true,
	}
}

// NewEnhancedMarkdownIO creates a new enhanced markdown IO handler
func NewEnhancedMarkdownIO(baseDir string, config *EnhancedIOConfig) *EnhancedMarkdownIO {
	if config == nil {
		config = DefaultEnhancedIOConfig()
	}

	// Initialize components
	hookSystem := hooks.NewHookSystem(hooks.DefaultHookConfig())
	transitionEngine := automation.NewTransitionEngine(hookSystem, automation.DefaultTransitionConfig())
	gitManager := git.NewContextManager()

	// Create enhanced IO
	enhanced := &EnhancedMarkdownIO{
		MarkdownIO:       NewMarkdownIO(baseDir),
		hookSystem:       hookSystem,
		transitionEngine: transitionEngine,
		gitManager:       gitManager,
		config:           config,
	}

	// Register default hooks
	enhanced.registerDefaultHooks()

	return enhanced
}

// WriteWork writes a Work container with automation
func (e *EnhancedMarkdownIO) WriteWork(ctx context.Context, work *models.Work) error {
	oldWork := *work // Copy for comparison

	// Update Git context if enabled
	if e.config.EnableGitTracking {
		if err := e.gitManager.UpdateWorkItemContext(ctx, work); err != nil {
			// Log but don't fail
			fmt.Printf("Warning: failed to update git context: %v\n", err)
		}
	}

	// Execute before hooks
	hookCtx := &hooks.HookContext{
		WorkItem:    work,
		OldWorkItem: &oldWork,
		EventType:   hooks.BeforeStatusChange,
		Timestamp:   time.Now(),
	}

	if results, err := e.hookSystem.ExecuteSync(ctx, hookCtx); err != nil {
		return fmt.Errorf("before hooks failed: %w", err)
	} else {
		for _, result := range results {
			if !result.Success {
				fmt.Printf("Hook %s warning: %v\n", result.HookName, result.Error)
			}
		}
	}

	// Apply automatic transitions if enabled
	if e.config.EnableAutomation {
		transitioned, applied, err := e.transitionEngine.EvaluateWork(ctx, work)
		if err != nil {
			return fmt.Errorf("transition evaluation failed: %w", err)
		}
		if applied {
			work = transitioned
			// Generate update if status changed
			if e.config.GenerateUpdates && oldWork.Metadata.Status != work.Metadata.Status {
				e.generateStatusUpdate(work, oldWork.Metadata.Status, work.Metadata.Status)
			}
		}
	}

	// Handle schedule changes (file moves)
	if oldWork.Schedule != work.Schedule && oldWork.Filepath != "" {
		// Need to move file to new directory
		if err := e.moveWorkFile(work, oldWork.Schedule, work.Schedule); err != nil {
			return fmt.Errorf("failed to move work file: %w", err)
		}
	}

	// Write using base implementation
	if err := e.MarkdownIO.WriteWork(work); err != nil {
		return err
	}

	// Execute after hooks
	hookCtx.EventType = hooks.AfterStatusChange
	hookCtx.Timestamp = time.Now()

	if results, err := e.hookSystem.Execute(ctx, hookCtx); err != nil {
		// Log but don't fail after write
		fmt.Printf("After hooks error: %v\n", err)
	} else {
		for _, result := range results {
			if !result.Success {
				fmt.Printf("Hook %s warning: %v\n", result.HookName, result.Error)
			}
		}
	}

	return nil
}

// UpdateProgress updates work progress with automation
func (e *EnhancedMarkdownIO) UpdateProgress(ctx context.Context, work *models.Work, progress int) error {
	oldProgress := work.Metadata.ProgressPercent
	work.Metadata.ProgressPercent = progress
	work.UpdatedAt = time.Now()

	// Generate progress update if enabled
	if e.config.GenerateUpdates && oldProgress != progress {
		e.generateProgressUpdate(work, oldProgress, progress)
	}

	// Trigger progress hook
	hookCtx := &hooks.HookContext{
		WorkItem:  work,
		EventType: hooks.ProgressUpdated,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"old_progress": oldProgress,
			"new_progress": progress,
		},
	}

	e.hookSystem.Execute(ctx, hookCtx)

	// Write with potential transitions
	return e.WriteWork(ctx, work)
}

// registerDefaultHooks sets up standard automation hooks
func (e *EnhancedMarkdownIO) registerDefaultHooks() {
	// Activity detection hook
	e.hookSystem.Register(hooks.AfterStatusChange, "activity_detector", func(ctx context.Context, hookCtx *hooks.HookContext) error {
		work := hookCtx.WorkItem
		work.Metadata.LastActivityAt = time.Now()
		work.Metadata.ActivityScore++
		return nil
	})

	// Git commit tracking hook
	e.hookSystem.Register(hooks.CommitDetected, "commit_tracker", func(ctx context.Context, hookCtx *hooks.HookContext) error {
		work := hookCtx.WorkItem
		if commitInfo, ok := hookCtx.Metadata["commit_info"].(map[string]string); ok {
			// Add commit to completed tasks
			commitTask := fmt.Sprintf("Commit: %s - %s", commitInfo["hash"][:8], commitInfo["message"])
			work.Metadata.CompletedTasks = append(work.Metadata.CompletedTasks, commitTask)
		}
		return nil
	})

	// Decay warning hook
	e.hookSystem.Register(hooks.InactivityWarning, "decay_warner", func(ctx context.Context, hookCtx *hooks.HookContext) error {
		work := hookCtx.WorkItem
		if work.Schedule == "now" && work.Metadata.Status == "in_progress" {
			// Add warning to metadata
			work.Metadata.Warnings = append(work.Metadata.Warnings, fmt.Sprintf("Inactive for %d days - consider moving to NEXT", 7))
		}
		return nil
	})
}

// Helper methods

func (e *EnhancedMarkdownIO) moveWorkFile(work *models.Work, oldSchedule, newSchedule string) error {
	oldDir := e.getWorkDirectory(oldSchedule)
	newDir := e.getWorkDirectory(newSchedule)

	if oldDir == newDir {
		return nil // No move needed
	}

	oldPath := filepath.Join(oldDir, work.Filename)
	newPath := filepath.Join(newDir, work.Filename)

	// TODO: Implement file move logic
	// This would involve reading the file, deleting from old location, and writing to new location

	return nil
}

func (e *EnhancedMarkdownIO) generateStatusUpdate(work *models.Work, oldStatus, newStatus string) {
	update := fmt.Sprintf("Status changed from %s to %s", oldStatus, newStatus)
	if reason, ok := work.Metadata.Metadata["transition_reason"].(string); ok {
		update += fmt.Sprintf(" - %s", reason)
	}

	// Add to update list
	if work.Metadata.UpdateEntries == nil {
		work.Metadata.UpdateEntries = []models.UpdateEntry{}
	}

	work.Metadata.UpdateEntries = append(work.Metadata.UpdateEntries, models.UpdateEntry{
		Timestamp: time.Now(),
		Type:      "status_change",
		Content:   update,
		Metadata: map[string]interface{}{
			"old_status": oldStatus,
			"new_status": newStatus,
		},
	})
}

func (e *EnhancedMarkdownIO) generateProgressUpdate(work *models.Work, oldProgress, newProgress int) {
	update := fmt.Sprintf("Progress updated from %d%% to %d%%", oldProgress, newProgress)

	// Determine milestone if applicable
	if newProgress == 25 || newProgress == 50 || newProgress == 75 || newProgress == 100 {
		update += fmt.Sprintf(" - Reached %d%% milestone", newProgress)
	}

	// Add to update list
	if work.Metadata.UpdateEntries == nil {
		work.Metadata.UpdateEntries = []models.UpdateEntry{}
	}

	work.Metadata.UpdateEntries = append(work.Metadata.UpdateEntries, models.UpdateEntry{
		Timestamp: time.Now(),
		Type:      "progress_update",
		Content:   update,
		Metadata: map[string]interface{}{
			"old_progress": oldProgress,
			"new_progress": newProgress,
		},
	})
}

// Public methods for external use

// GetHookSystem returns the hook system for external registration
func (e *EnhancedMarkdownIO) GetHookSystem() *hooks.HookSystem {
	return e.hookSystem
}

// GetTransitionEngine returns the transition engine for custom rules
func (e *EnhancedMarkdownIO) GetTransitionEngine() *automation.TransitionEngine {
	return e.transitionEngine
}

// CheckPendingTransitions checks all work items for pending transitions
func (e *EnhancedMarkdownIO) CheckPendingTransitions(ctx context.Context) ([]models.Work, error) {
	var pendingWorks []models.Work

	// Check all schedules
	for _, schedule := range []string{"now", "next", "later"} {
		works, err := e.ReadWorksBySchedule(schedule)
		if err != nil {
			continue
		}

		for _, work := range works {
			if pending, hasPending := e.transitionEngine.GetPendingTransition(&work); hasPending {
				work.Metadata.Metadata["pending_transition_target"] = pending
				pendingWorks = append(pendingWorks, work)
			}
		}
	}

	return pendingWorks, nil
}