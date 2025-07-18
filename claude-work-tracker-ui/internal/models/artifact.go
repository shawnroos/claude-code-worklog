package models

import (
	"time"
)

// Artifact represents a supporting document that informs Work but isn't directly scheduled
type Artifact struct {
	// Core identification
	ID            string    `yaml:"id" json:"id"`
	Type          string    `yaml:"type" json:"type"`                    // plan|proposal|analysis|update|decision
	Summary       string    `yaml:"summary" json:"summary"`              // Tweet-length summary
	TechnicalTags []string  `yaml:"technical_tags" json:"technical_tags"` // [design, frontend, api, etc]
	
	// Timestamps
	CreatedAt     time.Time `yaml:"created_at" json:"created_at"`
	UpdatedAt     time.Time `yaml:"updated_at" json:"updated_at"`
	
	// Context
	GitContext    GitContext `yaml:"git_context" json:"git_context"`
	SessionNumber string     `yaml:"session_number" json:"session_number"`
	
	// Associations - 3-tier system
	RelatedArtifacts []string `yaml:"related_artifacts,omitempty" json:"related_artifacts,omitempty"` // Strong references to other artifacts
	WorkRefs         []string `yaml:"work_refs,omitempty" json:"work_refs,omitempty"`               // Work items this artifact supports
	GroupID          string   `yaml:"group_id,omitempty" json:"group_id,omitempty"`                 // Explicit grouping
	
	// Artifact metadata
	Metadata      ArtifactMetadata `yaml:"metadata" json:"metadata"`
	
	// Content is the markdown body after frontmatter
	Content       string `yaml:"-" json:"content"`
	
	// Derived fields
	Filename      string `yaml:"-" json:"filename"`
	Filepath      string `yaml:"-" json:"filepath"`
}

// ArtifactMetadata contains artifact-specific metadata (evolved from MarkdownMetadata)
type ArtifactMetadata struct {
	// Common fields
	Status         string    `yaml:"status,omitempty" json:"status,omitempty"`                 // draft|active|archived|stale
	Confidence     string    `yaml:"confidence,omitempty" json:"confidence,omitempty"`        // low|medium|high
	
	// Association tracking
	WorkAssignments    []string  `yaml:"work_assignments,omitempty" json:"work_assignments,omitempty"`     // Work IDs this was assigned to
	LastAssignedAt     *time.Time `yaml:"last_assigned_at,omitempty" json:"last_assigned_at,omitempty"`     // When last assigned to work
	ReferenceCount     int       `yaml:"reference_count" json:"reference_count"`                           // How many things reference this
	
	// Lifecycle tracking  
	ActivityScore      float64   `yaml:"activity_score" json:"activity_score"`                             // Algorithm-calculated relevance
	LastActivityAt     *time.Time `yaml:"last_activity_at,omitempty" json:"last_activity_at,omitempty"`     // Last meaningful activity
	DecayWarning       bool      `yaml:"decay_warning" json:"decay_warning"`                               // Flag for stale artifacts
	OrphanedAt         *time.Time `yaml:"orphaned_at,omitempty" json:"orphaned_at,omitempty"`               // When it became orphaned
	
	// Type-specific fields (preserved from original MarkdownMetadata)
	
	// Plan-specific
	ImplementationStatus string   `yaml:"implementation_status,omitempty" json:"implementation_status,omitempty"` // not_started|in_progress|completed
	Phases              []string `yaml:"phases,omitempty" json:"phases,omitempty"`                             // List of plan phases
	EstimatedEffort     string   `yaml:"estimated_effort,omitempty" json:"estimated_effort,omitempty"`          // low|medium|high
	
	// Decision-specific  
	EnforcementActive      bool       `yaml:"enforcement_active,omitempty" json:"enforcement_active,omitempty"`       // Is this decision being enforced?
	Supersedes            []string   `yaml:"supersedes,omitempty" json:"supersedes,omitempty"`                     // IDs of decisions this replaces
	AlternativesConsidered []string   `yaml:"alternatives_considered,omitempty" json:"alternatives_considered,omitempty"`
	ReviewDate            *time.Time `yaml:"review_date,omitempty" json:"review_date,omitempty"`                   // When to review this decision
	
	// Analysis-specific
	AnalysisScope       []string `yaml:"analysis_scope,omitempty" json:"analysis_scope,omitempty"`           // What was analyzed
	ToolsUsed          []string `yaml:"tools_used,omitempty" json:"tools_used,omitempty"`                   // grep, read, etc
	ConfidenceLevel    string   `yaml:"confidence_level,omitempty" json:"confidence_level,omitempty"`        // low|medium|high
	
	// Update-specific
	UpdatesItem        string   `yaml:"updates_item,omitempty" json:"updates_item,omitempty"`               // ID of plan/proposal being updated
	ProgressPercentage int      `yaml:"progress_percentage,omitempty" json:"progress_percentage,omitempty"`  // 0-100
	BlockersIdentified []string `yaml:"blockers_identified,omitempty" json:"blockers_identified,omitempty"`
	
	// Proposal-specific
	ApprovalStatus     string   `yaml:"approval_status,omitempty" json:"approval_status,omitempty"`         // pending|approved|rejected|deferred
	EstimatedImpact    string   `yaml:"estimated_impact,omitempty" json:"estimated_impact,omitempty"`       // low|medium|high
	Dependencies       []string `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`               // What this depends on
}

// Artifact status constants
const (
	ArtifactStatusDraft    = "draft"    // Initial creation
	ArtifactStatusActive   = "active"   // Ready for use/reference
	ArtifactStatusArchived = "archived" // No longer relevant
	ArtifactStatusStale    = "stale"    // Flagged for potential cleanup
)

// Helper methods for Artifact

// GetDisplayType returns a formatted type string
func (a *Artifact) GetDisplayType() string {
	switch a.Type {
	case TypePlan:
		return "Plan"
	case TypeProposal:
		return "Proposal"
	case TypeAnalysis:
		return "Analysis"
	case TypeUpdate:
		return "Update"
	case TypeDecision:
		return "Decision"
	default:
		return a.Type
	}
}

// IsOrphaned returns true if this artifact has no work assignments or references
func (a *Artifact) IsOrphaned() bool {
	return len(a.WorkRefs) == 0 && 
		   len(a.Metadata.WorkAssignments) == 0 && 
		   a.Metadata.ReferenceCount == 0 &&
		   a.GroupID == ""
}

// NeedsReview returns true if this artifact needs review (decisions past review date, etc)
func (a *Artifact) NeedsReview() bool {
	if a.Type == TypeDecision && a.Metadata.ReviewDate != nil {
		return time.Now().After(*a.Metadata.ReviewDate)
	}
	return false
}

// IsStale returns true if this artifact should be considered stale
func (a *Artifact) IsStale() bool {
	return a.Metadata.Status == ArtifactStatusStale || a.Metadata.DecayWarning
}

// GetTypeIcon returns an icon/emoji for the artifact type
func (a *Artifact) GetTypeIcon() string {
	switch a.Type {
	case TypePlan:
		return "📋"
	case TypeProposal:
		return "💡"
	case TypeAnalysis:
		return "🔍"
	case TypeUpdate:
		return "📝"
	case TypeDecision:
		return "⚖️"
	default:
		return "📄"
	}
}

// AssignToWork assigns this artifact to a work item
func (a *Artifact) AssignToWork(workID string) {
	// Add to work references if not already present
	for _, ref := range a.WorkRefs {
		if ref == workID {
			return // Already assigned
		}
	}
	
	a.WorkRefs = append(a.WorkRefs, workID)
	
	// Add to work assignments history
	found := false
	for _, assignment := range a.Metadata.WorkAssignments {
		if assignment == workID {
			found = true
			break
		}
	}
	if !found {
		a.Metadata.WorkAssignments = append(a.Metadata.WorkAssignments, workID)
	}
	
	now := time.Now()
	a.Metadata.LastAssignedAt = &now
	a.Metadata.LastActivityAt = &now
	a.UpdatedAt = now
	
	// Clear orphaned status
	a.Metadata.OrphanedAt = nil
	a.Metadata.DecayWarning = false
	a.Metadata.Status = ArtifactStatusActive
}

// UnassignFromWork removes this artifact from a work item
func (a *Artifact) UnassignFromWork(workID string) {
	var newRefs []string
	for _, ref := range a.WorkRefs {
		if ref != workID {
			newRefs = append(newRefs, ref)
		}
	}
	
	a.WorkRefs = newRefs
	a.UpdatedAt = time.Now()
	
	// Check if now orphaned
	if a.IsOrphaned() {
		now := time.Now()
		a.Metadata.OrphanedAt = &now
	}
}

// AddReference adds a reference to another artifact
func (a *Artifact) AddReference(artifactID string) {
	// Add to related artifacts if not already present
	for _, ref := range a.RelatedArtifacts {
		if ref == artifactID {
			return // Already exists
		}
	}
	
	a.RelatedArtifacts = append(a.RelatedArtifacts, artifactID)
	a.UpdatedAt = time.Now()
	
	now := time.Now()
	a.Metadata.LastActivityAt = &now
}

// RemoveReference removes a reference to another artifact
func (a *Artifact) RemoveReference(artifactID string) {
	var newRefs []string
	for _, ref := range a.RelatedArtifacts {
		if ref != artifactID {
			newRefs = append(newRefs, ref)
		}
	}
	
	a.RelatedArtifacts = newRefs
	a.UpdatedAt = time.Now()
}

// UpdateReferenceCount updates the count of external references to this artifact
func (a *Artifact) UpdateReferenceCount(count int) {
	a.Metadata.ReferenceCount = count
	
	// Reset decay warning if still being referenced
	if count > 0 {
		a.Metadata.DecayWarning = false
		if a.IsOrphaned() {
			a.Metadata.OrphanedAt = nil
		}
	}
}

// CalculateActivityScore calculates activity based on recent changes, references, assignments
func (a *Artifact) CalculateActivityScore() float64 {
	score := 0.0
	now := time.Now()
	
	// Recent activity bonus
	if a.Metadata.LastActivityAt != nil {
		daysSince := now.Sub(*a.Metadata.LastActivityAt).Hours() / 24
		if daysSince < 1 {
			score += 10.0
		} else if daysSince < 7 {
			score += 5.0 - daysSince
		}
	}
	
	// Reference count bonus
	score += float64(a.Metadata.ReferenceCount) * 2.0
	
	// Work assignment bonus
	score += float64(len(a.WorkRefs)) * 3.0
	
	// Related artifacts bonus
	score += float64(len(a.RelatedArtifacts)) * 1.0
	
	// Group membership bonus
	if a.GroupID != "" {
		score += 2.0
	}
	
	// Type-specific bonuses
	switch a.Type {
	case TypeDecision:
		if a.Metadata.EnforcementActive {
			score += 5.0
		}
	case TypePlan:
		if a.Metadata.ImplementationStatus == "in_progress" {
			score += 3.0
		}
	case TypeUpdate:
		// Updates get bonus for being recent
		score += 2.0
	}
	
	a.Metadata.ActivityScore = score
	return score
}

// ShouldDecay returns true if this artifact should be flagged for decay
func (a *Artifact) ShouldDecay() bool {
	now := time.Now()
	
	// Don't decay active enforcement decisions
	if a.Type == TypeDecision && a.Metadata.EnforcementActive {
		return false
	}
	
	// Check for orphaned status
	if a.IsOrphaned() {
		if a.Metadata.OrphanedAt != nil {
			daysSinceOrphaned := now.Sub(*a.Metadata.OrphanedAt).Hours() / 24
			if daysSinceOrphaned > 30 {
				return true
			}
		} else {
			// Became orphaned but not tracked - check creation date
			daysSinceCreated := now.Sub(a.CreatedAt).Hours() / 24
			if daysSinceCreated > 14 {
				return true
			}
		}
	}
	
	// Check for stale activity
	if a.Metadata.LastActivityAt != nil {
		daysSince := now.Sub(*a.Metadata.LastActivityAt).Hours() / 24
		
		// Different decay rates based on type
		switch a.Type {
		case TypeUpdate:
			// Updates decay faster as they become historical
			if daysSince > 30 {
				return true
			}
		case TypeAnalysis:
			// Analysis can be relevant longer
			if daysSince > 90 {
				return true
			}
		case TypeDecision, TypePlan:
			// Decisions and plans have longer relevance
			if daysSince > 180 {
				return true
			}
		default:
			// Default decay period
			if daysSince > 60 {
				return true
			}
		}
	}
	
	return false
}

// MarkAsStale marks this artifact as stale and sets decay warning
func (a *Artifact) MarkAsStale() {
	a.Metadata.Status = ArtifactStatusStale
	a.Metadata.DecayWarning = true
	a.UpdatedAt = time.Now()
}

// Reactivate removes stale status and decay warning
func (a *Artifact) Reactivate() {
	a.Metadata.Status = ArtifactStatusActive
	a.Metadata.DecayWarning = false
	now := time.Now()
	a.Metadata.LastActivityAt = &now
	a.UpdatedAt = now
}

// ToMarkdownWorkItem converts artifact to legacy MarkdownWorkItem for compatibility
func (a *Artifact) ToMarkdownWorkItem() *MarkdownWorkItem {
	return &MarkdownWorkItem{
		ID:             a.ID,
		Type:           a.Type,
		Summary:        a.Summary,
		Schedule:       "unscheduled", // Artifacts don't have schedules
		TechnicalTags:  a.TechnicalTags,
		SessionNumber:  a.SessionNumber,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
		GitContext:     a.GitContext,
		Content:        a.Content,
		Filename:       a.Filename,
		Filepath:       a.Filepath,
		Metadata: MarkdownMetadata{
			Status:                 a.Metadata.Status,
			RelatedItems:          a.RelatedArtifacts,
			ImplementationStatus:   a.Metadata.ImplementationStatus,
			Phases:                a.Metadata.Phases,
			EstimatedEffort:       a.Metadata.EstimatedEffort,
			EnforcementActive:     a.Metadata.EnforcementActive,
			Supersedes:            a.Metadata.Supersedes,
			AlternativesConsidered: a.Metadata.AlternativesConsidered,
			ReviewDate:            a.Metadata.ReviewDate,
			AnalysisScope:         a.Metadata.AnalysisScope,
			ToolsUsed:             a.Metadata.ToolsUsed,
			ConfidenceLevel:       a.Metadata.ConfidenceLevel,
			UpdatesItem:           a.Metadata.UpdatesItem,
			ProgressPercentage:    a.Metadata.ProgressPercentage,
			BlockersIdentified:    a.Metadata.BlockersIdentified,
			ApprovalStatus:        a.Metadata.ApprovalStatus,
			EstimatedImpact:       a.Metadata.EstimatedImpact,
			Dependencies:          a.Metadata.Dependencies,
		},
	}
}