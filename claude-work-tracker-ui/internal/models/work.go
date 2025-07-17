package models

import (
	"time"
)

// Work represents a schedulable work container that can contain multiple artifacts
type Work struct {
	// Core identification
	ID            string    `yaml:"id" json:"id"`
	Title         string    `yaml:"title" json:"title"`                 // Primary work title
	Description   string    `yaml:"description" json:"description"`     // What needs to be accomplished
	Schedule      string    `yaml:"schedule" json:"schedule"`           // now|next|later
	
	// Timestamps
	CreatedAt     time.Time `yaml:"created_at" json:"created_at"`
	UpdatedAt     time.Time `yaml:"updated_at" json:"updated_at"`
	StartedAt     *time.Time `yaml:"started_at,omitempty" json:"started_at,omitempty"`
	CompletedAt   *time.Time `yaml:"completed_at,omitempty" json:"completed_at,omitempty"`
	
	// Context
	GitContext    GitContext `yaml:"git_context" json:"git_context"`
	SessionNumber string     `yaml:"session_number" json:"session_number"`
	
	// Associations - 3-tier system
	TechnicalTags  []string `yaml:"technical_tags" json:"technical_tags"`               // Incidental relationships
	ArtifactRefs   []string `yaml:"artifact_refs" json:"artifact_refs"`                 // Strong references to artifacts
	GroupID        string   `yaml:"group_id,omitempty" json:"group_id,omitempty"`       // Explicit grouping
	
	// Work-specific metadata
	Metadata      WorkMetadata `yaml:"metadata" json:"metadata"`
	
	// Enhanced structure fields
	OverviewUpdated *time.Time `yaml:"overview_updated,omitempty" json:"overview_updated,omitempty"`
	UpdatesRef      string     `yaml:"updates_ref,omitempty" json:"updates_ref,omitempty"`
	
	// Content is the markdown body after frontmatter
	Content       string `yaml:"-" json:"content"`
	
	// Derived fields  
	Filename      string `yaml:"-" json:"filename"`
	Filepath      string `yaml:"-" json:"filepath"`
	
	// Source tracking for multi-directory projects
	SourceDirectory string `yaml:"-" json:"source_directory,omitempty"` // Relative path from project root
	SourcePath      string `yaml:"-" json:"source_path,omitempty"`      // Full path to .claude-work directory
}

// WorkMetadata contains work-specific metadata and status
type WorkMetadata struct {
	// Status tracking
	Status            string   `yaml:"status" json:"status"`                                       // draft|active|in_progress|completed|archived|blocked
	Priority          string   `yaml:"priority,omitempty" json:"priority,omitempty"`              // low|medium|high|critical
	EstimatedEffort   string   `yaml:"estimated_effort,omitempty" json:"estimated_effort,omitempty"` // small|medium|large|epic
	
	// Progress tracking
	ProgressPercent   int      `yaml:"progress_percent" json:"progress_percent"`                   // 0-100
	Milestones        []string `yaml:"milestones,omitempty" json:"milestones,omitempty"`          // Key milestones
	CompletedTasks    []string `yaml:"completed_tasks,omitempty" json:"completed_tasks,omitempty"` // Completed sub-tasks
	PendingTasks      []string `yaml:"pending_tasks,omitempty" json:"pending_tasks,omitempty"`    // Remaining sub-tasks
	
	// Dependency management
	BlockedBy         []string `yaml:"blocked_by,omitempty" json:"blocked_by,omitempty"`          // Work IDs blocking this
	Blocks            []string `yaml:"blocks,omitempty" json:"blocks,omitempty"`                  // Work IDs this blocks
	Dependencies      []string `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`     // External dependencies
	
	// Context and outcomes
	SuccessCriteria   []string `yaml:"success_criteria,omitempty" json:"success_criteria,omitempty"` // How to know it's done
	AcceptanceCriteria []string `yaml:"acceptance_criteria,omitempty" json:"acceptance_criteria,omitempty"` // Acceptance tests
	DeliveryTargets   []string `yaml:"delivery_targets,omitempty" json:"delivery_targets,omitempty"` // What will be delivered
	
	// Quality and review
	ReviewRequired    bool     `yaml:"review_required" json:"review_required"`                    // Needs review before completion
	ReviewedBy        []string `yaml:"reviewed_by,omitempty" json:"reviewed_by,omitempty"`       // Who reviewed this
	QualityChecks     []string `yaml:"quality_checks,omitempty" json:"quality_checks,omitempty"` // Quality gates passed
	
	// Artifact management
	ArtifactCount     int      `yaml:"artifact_count" json:"artifact_count"`                     // Number of supporting artifacts
	LastArtifactAdded *time.Time `yaml:"last_artifact_added,omitempty" json:"last_artifact_added,omitempty"` // When last artifact was added
	
	// Lifecycle tracking
	ActivityScore     float64  `yaml:"activity_score" json:"activity_score"`                     // Algorithm-calculated activity
	LastActivityAt    *time.Time `yaml:"last_activity_at,omitempty" json:"last_activity_at,omitempty"` // Last meaningful activity
	DecayWarning      bool     `yaml:"decay_warning" json:"decay_warning"`                       // Flag for stale work
}

// Work status constants
const (
	WorkStatusDraft      = "draft"       // Initial creation
	WorkStatusActive     = "active"      // Ready to work on
	WorkStatusInProgress = "in_progress" // Currently being worked on
	WorkStatusCompleted  = "completed"   // Successfully finished
	WorkStatusArchived   = "archived"    // No longer relevant
	WorkStatusBlocked    = "blocked"     // Cannot proceed
	WorkStatusOnHold     = "on_hold"     // Temporarily paused
)

// Work priority constants
const (
	WorkPriorityLow      = "low"
	WorkPriorityMedium   = "medium"
	WorkPriorityHigh     = "high"
	WorkPriorityCritical = "critical"
)

// Work effort estimation constants
const (
	WorkEffortSmall  = "small"  // 1-2 days
	WorkEffortMedium = "medium" // 1-2 weeks
	WorkEffortLarge  = "large"  // 1-2 months
	WorkEffortEpic   = "epic"   // 2+ months, needs breakdown
)

// Helper methods for Work

// GetSchedulePriority returns a numeric priority based on schedule
func (w *Work) GetSchedulePriority() int {
	switch w.Schedule {
	case ScheduleNow:
		return 1
	case ScheduleNext:
		return 2
	case ScheduleLater:
		return 3
	default:
		return 4
	}
}

// GetDisplaySchedule returns a formatted schedule string
func (w *Work) GetDisplaySchedule() string {
	switch w.Schedule {
	case ScheduleNow:
		return "NOW"
	case ScheduleNext:
		return "NEXT"
	case ScheduleLater:
		return "LATER"
	default:
		return "Unscheduled"
	}
}

// IsActive returns true if this work should be actively worked on
func (w *Work) IsActive() bool {
	return w.Schedule == ScheduleNow && 
		   (w.Metadata.Status == WorkStatusActive || w.Metadata.Status == WorkStatusInProgress)
}

// IsBlocked returns true if this work is blocked
func (w *Work) IsBlocked() bool {
	return w.Metadata.Status == WorkStatusBlocked || len(w.Metadata.BlockedBy) > 0
}

// IsCompleted returns true if this work is finished
func (w *Work) IsCompleted() bool {
	return w.Metadata.Status == WorkStatusCompleted
}

// NeedsAttention returns true if this work needs immediate attention
func (w *Work) NeedsAttention() bool {
	return w.IsBlocked() || 
		   w.Metadata.DecayWarning ||
		   (w.IsActive() && w.Metadata.ProgressPercent == 0)
}

// GetEffortNumeric returns a numeric value for effort estimation
func (w *Work) GetEffortNumeric() int {
	switch w.Metadata.EstimatedEffort {
	case WorkEffortSmall:
		return 1
	case WorkEffortMedium:
		return 2
	case WorkEffortLarge:
		return 3
	case WorkEffortEpic:
		return 4
	default:
		return 2 // Default to medium
	}
}

// GetPriorityNumeric returns a numeric value for priority
func (w *Work) GetPriorityNumeric() int {
	switch w.Metadata.Priority {
	case WorkPriorityCritical:
		return 4
	case WorkPriorityHigh:
		return 3
	case WorkPriorityMedium:
		return 2
	case WorkPriorityLow:
		return 1
	default:
		return 2 // Default to medium
	}
}

// AddArtifact adds an artifact reference and updates metadata
func (w *Work) AddArtifact(artifactID string) {
	// Add to references if not already present
	for _, ref := range w.ArtifactRefs {
		if ref == artifactID {
			return // Already exists
		}
	}
	
	w.ArtifactRefs = append(w.ArtifactRefs, artifactID)
	w.Metadata.ArtifactCount = len(w.ArtifactRefs)
	
	now := time.Now()
	w.Metadata.LastArtifactAdded = &now
	w.Metadata.LastActivityAt = &now
	w.UpdatedAt = now
	
	// Reset decay warning when new artifacts are added
	w.Metadata.DecayWarning = false
}

// RemoveArtifact removes an artifact reference and updates metadata
func (w *Work) RemoveArtifact(artifactID string) {
	var newRefs []string
	for _, ref := range w.ArtifactRefs {
		if ref != artifactID {
			newRefs = append(newRefs, ref)
		}
	}
	
	w.ArtifactRefs = newRefs
	w.Metadata.ArtifactCount = len(w.ArtifactRefs)
	w.UpdatedAt = time.Now()
}

// UpdateProgress updates the progress percentage and related metadata
func (w *Work) UpdateProgress(percent int) {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	
	w.Metadata.ProgressPercent = percent
	now := time.Now()
	w.Metadata.LastActivityAt = &now
	w.UpdatedAt = now
	
	// Auto-complete if 100%
	if percent == 100 && w.Metadata.Status != WorkStatusCompleted {
		w.Metadata.Status = WorkStatusCompleted
		w.CompletedAt = &now
	}
	
	// Reset decay warning on progress updates
	w.Metadata.DecayWarning = false
}

// MarkAsCompleted marks the work as completed with timestamp
func (w *Work) MarkAsCompleted() {
	now := time.Now()
	w.Metadata.Status = WorkStatusCompleted
	w.Metadata.ProgressPercent = 100
	w.CompletedAt = &now
	w.UpdatedAt = now
}

// MarkAsBlocked marks the work as blocked with optional blockers
func (w *Work) MarkAsBlocked(blockedBy ...string) {
	w.Metadata.Status = WorkStatusBlocked
	if len(blockedBy) > 0 {
		w.Metadata.BlockedBy = append(w.Metadata.BlockedBy, blockedBy...)
	}
	w.UpdatedAt = time.Now()
}

// CalculateActivityScore calculates activity based on recent changes, artifacts, progress
func (w *Work) CalculateActivityScore() float64 {
	score := 0.0
	now := time.Now()
	
	// Recent activity bonus
	if w.Metadata.LastActivityAt != nil {
		daysSince := now.Sub(*w.Metadata.LastActivityAt).Hours() / 24
		if daysSince < 1 {
			score += 10.0
		} else if daysSince < 7 {
			score += 5.0 - daysSince
		}
	}
	
	// Artifact count bonus
	score += float64(w.Metadata.ArtifactCount) * 2.0
	
	// Progress bonus
	score += float64(w.Metadata.ProgressPercent) * 0.1
	
	// Schedule bonus
	switch w.Schedule {
	case ScheduleNow:
		score += 5.0
	case ScheduleNext:
		score += 2.0
	}
	
	// Status bonus
	if w.Metadata.Status == WorkStatusInProgress {
		score += 3.0
	} else if w.Metadata.Status == WorkStatusActive {
		score += 1.0
	}
	
	w.Metadata.ActivityScore = score
	return score
}

// GetLastUpdateTime returns the most recent update time including artifact updates
func (w *Work) GetLastUpdateTime() time.Time {
	lastUpdate := w.UpdatedAt
	
	// Check if LastActivityAt is more recent
	if w.Metadata.LastActivityAt != nil && w.Metadata.LastActivityAt.After(lastUpdate) {
		lastUpdate = *w.Metadata.LastActivityAt
	}
	
	// Check if LastArtifactAdded is more recent
	if w.Metadata.LastArtifactAdded != nil && w.Metadata.LastArtifactAdded.After(lastUpdate) {
		lastUpdate = *w.Metadata.LastArtifactAdded
	}
	
	return lastUpdate
}

// ShouldDecay returns true if this work should be flagged for decay
func (w *Work) ShouldDecay() bool {
	now := time.Now()
	
	// Don't decay completed work
	if w.IsCompleted() {
		return false
	}
	
	// Check for stale activity
	if w.Metadata.LastActivityAt != nil {
		daysSince := now.Sub(*w.Metadata.LastActivityAt).Hours() / 24
		
		// NOW items decay faster
		if w.Schedule == ScheduleNow && daysSince > 7 {
			return true
		}
		
		// NEXT items have longer grace period
		if w.Schedule == ScheduleNext && daysSince > 30 {
			return true
		}
		
		// LATER items have very long grace period
		if w.Schedule == ScheduleLater && daysSince > 90 {
			return true
		}
	}
	
	// No artifacts for extended period
	if w.Metadata.ArtifactCount == 0 {
		daysSinceCreated := now.Sub(w.CreatedAt).Hours() / 24
		if daysSinceCreated > 14 {
			return true
		}
	}
	
	return false
}