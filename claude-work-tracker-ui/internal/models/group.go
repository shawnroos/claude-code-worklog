package models

import (
	"fmt"
	"time"
)

// Group represents an explicit grouping of artifacts that can be consolidated into Work
type Group struct {
	// Core identification
	ID          string    `yaml:"id" json:"id"`
	Name        string    `yaml:"name" json:"name"`                         // Human-readable group name
	Description string    `yaml:"description" json:"description"`           // What this group represents
	Theme       string    `yaml:"theme,omitempty" json:"theme,omitempty"`   // Common theme/topic
	
	// Timestamps
	CreatedAt   time.Time `yaml:"created_at" json:"created_at"`
	UpdatedAt   time.Time `yaml:"updated_at" json:"updated_at"`
	
	// Context
	GitContext  GitContext `yaml:"git_context" json:"git_context"`
	SessionNumber string   `yaml:"session_number" json:"session_number"`
	
	// Group composition
	ArtifactIDs []string   `yaml:"artifact_ids" json:"artifact_ids"`        // Artifacts in this group
	WorkRefs    []string   `yaml:"work_refs,omitempty" json:"work_refs,omitempty"` // Work items created from this group
	TechnicalTags []string `yaml:"technical_tags" json:"technical_tags"`    // Common tags across artifacts
	
	// Group metadata
	Metadata    GroupMetadata `yaml:"metadata" json:"metadata"`
	
	// Derived fields
	Filename    string    `yaml:"-" json:"filename"`
	Filepath    string    `yaml:"-" json:"filepath"`
}

// GroupMetadata contains group-specific metadata and tracking
type GroupMetadata struct {
	// Status tracking
	Status         string    `yaml:"status" json:"status"`                               // active|consolidated|archived
	ConsolidatedAt *time.Time `yaml:"consolidated_at,omitempty" json:"consolidated_at,omitempty"` // When converted to Work
	ConsolidatedBy string    `yaml:"consolidated_by,omitempty" json:"consolidated_by,omitempty"`  // How it was consolidated
	
	// Composition analysis
	ArtifactCount    int       `yaml:"artifact_count" json:"artifact_count"`                     // Number of artifacts
	TypeDistribution map[string]int `yaml:"type_distribution" json:"type_distribution"`           // Count by artifact type
	ConfidenceScore  float64   `yaml:"confidence_score" json:"confidence_score"`                 // How well grouped (0-1)
	SimilarityScore  float64   `yaml:"similarity_score" json:"similarity_score"`                 // Content similarity (0-1)
	
	// Lifecycle tracking
	LastModified     *time.Time `yaml:"last_modified,omitempty" json:"last_modified,omitempty"`   // Last artifact addition/removal
	ActivityScore    float64    `yaml:"activity_score" json:"activity_score"`                     // Algorithm-calculated activity
	RecommendedForWork bool     `yaml:"recommended_for_work" json:"recommended_for_work"`         // AI suggests consolidation
	
	// Quality metrics
	CompletionScore  float64   `yaml:"completion_score" json:"completion_score"`                 // How complete the group feels (0-1)
	ReadinessScore   float64   `yaml:"readiness_score" json:"readiness_score"`                   // Ready for work conversion (0-1)
	CohesionScore    float64   `yaml:"cohesion_score" json:"cohesion_score"`                     // How well artifacts relate (0-1)
	
	// External relationships
	RelatedGroups    []string  `yaml:"related_groups,omitempty" json:"related_groups,omitempty"` // Other groups with similar themes
	MergeCandidate   string    `yaml:"merge_candidate,omitempty" json:"merge_candidate,omitempty"` // Group ID this could merge with
	SplitSuggested   bool      `yaml:"split_suggested" json:"split_suggested"`                   // Should be split into multiple groups
	
	// Consolidation hints
	SuggestedWorkTitle string   `yaml:"suggested_work_title,omitempty" json:"suggested_work_title,omitempty"` // AI-suggested work title
	SuggestedSchedule  string   `yaml:"suggested_schedule,omitempty" json:"suggested_schedule,omitempty"`     // now|next|later
	SuggestedPriority  string   `yaml:"suggested_priority,omitempty" json:"suggested_priority,omitempty"`     // low|medium|high|critical
	ConsolidationNotes []string `yaml:"consolidation_notes,omitempty" json:"consolidation_notes,omitempty"`   // AI notes for consolidation
}

// Group status constants
const (
	GroupStatusActive       = "active"       // Actively collecting artifacts
	GroupStatusConsolidated = "consolidated" // Already converted to Work
	GroupStatusArchived     = "archived"     // No longer relevant
	GroupStatusCandidate    = "candidate"    // Suggested grouping, needs validation
)

// Helper methods for Group

// AddArtifact adds an artifact to this group
func (g *Group) AddArtifact(artifactID string) {
	// Check if already in group
	for _, id := range g.ArtifactIDs {
		if id == artifactID {
			return // Already present
		}
	}
	
	g.ArtifactIDs = append(g.ArtifactIDs, artifactID)
	g.Metadata.ArtifactCount = len(g.ArtifactIDs)
	
	now := time.Now()
	g.Metadata.LastModified = &now
	g.UpdatedAt = now
	
	// Recalculate scores when composition changes
	g.CalculateScores()
}

// RemoveArtifact removes an artifact from this group
func (g *Group) RemoveArtifact(artifactID string) {
	var newIDs []string
	for _, id := range g.ArtifactIDs {
		if id != artifactID {
			newIDs = append(newIDs, id)
		}
	}
	
	g.ArtifactIDs = newIDs
	g.Metadata.ArtifactCount = len(g.ArtifactIDs)
	
	now := time.Now()
	g.Metadata.LastModified = &now
	g.UpdatedAt = now
	
	// Recalculate scores when composition changes
	g.CalculateScores()
}

// IsEmpty returns true if the group has no artifacts
func (g *Group) IsEmpty() bool {
	return len(g.ArtifactIDs) == 0
}

// IsReadyForWork returns true if this group is ready for consolidation into Work
func (g *Group) IsReadyForWork() bool {
	return g.Metadata.Status == GroupStatusActive &&
		   g.Metadata.ArtifactCount >= 2 &&
		   g.Metadata.ReadinessScore >= 0.7 &&
		   g.Metadata.CohesionScore >= 0.6
}

// IsConsolidated returns true if this group has been converted to Work
func (g *Group) IsConsolidated() bool {
	return g.Metadata.Status == GroupStatusConsolidated
}

// MarkAsConsolidated marks this group as consolidated into Work
func (g *Group) MarkAsConsolidated(workID string, method string) {
	g.Metadata.Status = GroupStatusConsolidated
	now := time.Now()
	g.Metadata.ConsolidatedAt = &now
	g.Metadata.ConsolidatedBy = method
	
	// Add work reference
	g.WorkRefs = append(g.WorkRefs, workID)
	g.UpdatedAt = now
}

// UpdateTypeDistribution updates the count of artifact types in this group
func (g *Group) UpdateTypeDistribution(artifacts []*Artifact) {
	g.Metadata.TypeDistribution = make(map[string]int)
	
	for _, artifact := range artifacts {
		// Only count artifacts that are in this group
		found := false
		for _, id := range g.ArtifactIDs {
			if id == artifact.ID {
				found = true
				break
			}
		}
		
		if found {
			g.Metadata.TypeDistribution[artifact.Type]++
		}
	}
}

// GetDominantType returns the artifact type that appears most in this group
func (g *Group) GetDominantType() string {
	maxCount := 0
	dominantType := ""
	
	for artifactType, count := range g.Metadata.TypeDistribution {
		if count > maxCount {
			maxCount = count
			dominantType = artifactType
		}
	}
	
	return dominantType
}

// HasMixedTypes returns true if the group contains multiple artifact types
func (g *Group) HasMixedTypes() bool {
	return len(g.Metadata.TypeDistribution) > 1
}

// CalculateScores calculates various quality and readiness scores for the group
func (g *Group) CalculateScores() {
	// Basic scores based on size and age
	sizeScore := float64(g.Metadata.ArtifactCount) / 10.0 // Max at 10 artifacts
	if sizeScore > 1.0 {
		sizeScore = 1.0
	}
	
	// Age scoring - newer groups score higher initially
	daysSinceCreated := time.Since(g.CreatedAt).Hours() / 24
	ageScore := 1.0
	if daysSinceCreated > 7 {
		ageScore = 1.0 - (daysSinceCreated-7)/30.0 // Decay over month
		if ageScore < 0.3 {
			ageScore = 0.3
		}
	}
	
	// Cohesion scoring (placeholder - would use content analysis in full implementation)
	cohesionScore := 0.5 // Base cohesion
	if len(g.TechnicalTags) > 0 {
		cohesionScore += 0.2 // Bonus for shared tags
	}
	if !g.HasMixedTypes() {
		cohesionScore += 0.2 // Bonus for type consistency
	}
	if cohesionScore > 1.0 {
		cohesionScore = 1.0
	}
	
	// Completion scoring
	completionScore := sizeScore * 0.7 + cohesionScore * 0.3
	
	// Readiness scoring
	readinessScore := (sizeScore + cohesionScore + ageScore) / 3.0
	
	// Activity scoring
	activityScore := 0.0
	if g.Metadata.LastModified != nil {
		daysSinceModified := time.Since(*g.Metadata.LastModified).Hours() / 24
		if daysSinceModified < 1 {
			activityScore = 10.0
		} else if daysSinceModified < 7 {
			activityScore = 5.0 - daysSinceModified
		}
	}
	activityScore += float64(g.Metadata.ArtifactCount) * 0.5
	
	// Update scores
	g.Metadata.CohesionScore = cohesionScore
	g.Metadata.CompletionScore = completionScore
	g.Metadata.ReadinessScore = readinessScore
	g.Metadata.ActivityScore = activityScore
	
	// Update recommendation
	g.Metadata.RecommendedForWork = g.IsReadyForWork()
}

// GenerateWorkSuggestion generates suggested Work properties from this group
func (g *Group) GenerateWorkSuggestion() (title, description, schedule, priority string) {
	// Title generation (simplified)
	if g.Metadata.SuggestedWorkTitle != "" {
		title = g.Metadata.SuggestedWorkTitle
	} else {
		dominantType := g.GetDominantType()
		switch dominantType {
		case TypePlan:
			title = "Implement " + g.Theme
		case TypeProposal:
			title = "Evaluate and Implement " + g.Theme
		case TypeAnalysis:
			title = "Act on Analysis of " + g.Theme
		default:
			title = "Work on " + g.Theme
		}
		if title == "Work on " {
			title = g.Name
		}
	}
	
	// Description
	description = g.Description
	if description == "" {
		description = fmt.Sprintf("Work item consolidated from %d artifacts related to %s", 
			g.Metadata.ArtifactCount, g.Theme)
	}
	
	// Schedule suggestion
	schedule = g.Metadata.SuggestedSchedule
	if schedule == "" {
		// Base on readiness and activity
		if g.Metadata.ReadinessScore > 0.8 && g.Metadata.ActivityScore > 5.0 {
			schedule = ScheduleNow
		} else if g.Metadata.ReadinessScore > 0.6 {
			schedule = ScheduleNext
		} else {
			schedule = ScheduleLater
		}
	}
	
	// Priority suggestion
	priority = g.Metadata.SuggestedPriority
	if priority == "" {
		// Base on activity and completion scores
		combinedScore := (g.Metadata.ActivityScore/10.0 + g.Metadata.CompletionScore) / 2.0
		if combinedScore > 0.8 {
			priority = WorkPriorityHigh
		} else if combinedScore > 0.5 {
			priority = WorkPriorityMedium
		} else {
			priority = WorkPriorityLow
		}
	}
	
	return title, description, schedule, priority
}

// ShouldSplit returns true if this group should be split into smaller groups
func (g *Group) ShouldSplit() bool {
	return g.Metadata.ArtifactCount > 8 && 
		   g.HasMixedTypes() && 
		   g.Metadata.CohesionScore < 0.5
}

// ShouldMerge returns true if this group should be merged with another
func (g *Group) ShouldMerge() bool {
	return g.Metadata.ArtifactCount < 3 && 
		   g.Metadata.MergeCandidate != "" &&
		   g.Metadata.CohesionScore < 0.6
}

// GetConsolidationSummary returns a summary of what consolidating this group would create
func (g *Group) GetConsolidationSummary() string {
	title, _, schedule, priority := g.GenerateWorkSuggestion()
	
	return fmt.Sprintf("Would create %s priority Work '%s' in %s with %d supporting artifacts",
		priority, title, schedule, g.Metadata.ArtifactCount)
}