package models

import (
	"time"
)

// MarkdownWorkItem represents a work item stored as markdown with YAML frontmatter
type MarkdownWorkItem struct {
	// Frontmatter fields
	ID             string     `yaml:"id" json:"id"`
	Type           string     `yaml:"type" json:"type"`                     // plan|proposal|analysis|update|decision
	Summary        string     `yaml:"summary" json:"summary"`               // Tweet-length summary
	Schedule       string     `yaml:"schedule" json:"schedule"`             // now|next|later
	TechnicalTags  []string   `yaml:"technical_tags" json:"technical_tags"` // [design, frontend, api, etc]
	SessionNumber  string     `yaml:"session_number" json:"session_number"`
	CreatedAt      time.Time  `yaml:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `yaml:"updated_at" json:"updated_at"`
	GitContext     GitContext `yaml:"git_context" json:"git_context"`
	
	// Metadata based on type
	Metadata MarkdownMetadata `yaml:"metadata" json:"metadata"`
	
	// Content is the markdown body after frontmatter
	Content string `yaml:"-" json:"content"`
	
	// Derived fields
	Filename string `yaml:"-" json:"filename"`
	Filepath string `yaml:"-" json:"filepath"`
}

// MarkdownMetadata contains type-specific metadata
type MarkdownMetadata struct {
	// Common fields
	Status         string   `yaml:"status,omitempty" json:"status,omitempty"`                 // draft|active|completed|archived
	RelatedItems   []string `yaml:"related_items,omitempty" json:"related_items,omitempty"`   // IDs of related work items
	
	// Plan-specific
	ImplementationStatus string   `yaml:"implementation_status,omitempty" json:"implementation_status,omitempty"` // not_started|in_progress|completed
	Phases              []string `yaml:"phases,omitempty" json:"phases,omitempty"`                             // List of plan phases
	EstimatedEffort     string   `yaml:"estimated_effort,omitempty" json:"estimated_effort,omitempty"`          // low|medium|high
	
	// Decision-specific  
	EnforcementActive    bool     `yaml:"enforcement_active,omitempty" json:"enforcement_active,omitempty"`     // Is this decision being enforced?
	Supersedes          []string `yaml:"supersedes,omitempty" json:"supersedes,omitempty"`                   // IDs of decisions this replaces
	AlternativesConsidered []string `yaml:"alternatives_considered,omitempty" json:"alternatives_considered,omitempty"`
	ReviewDate          *time.Time `yaml:"review_date,omitempty" json:"review_date,omitempty"`                // When to review this decision
	
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

// WorkItemType constants for the 5-type system
const (
	TypePlan     = "plan"
	TypeProposal = "proposal"
	TypeAnalysis = "analysis"
	TypeUpdate   = "update"
	TypeDecision = "decision"
)

// Schedule constants for NOW/NEXT/LATER
const (
	ScheduleNow   = "now"
	ScheduleNext  = "next"
	ScheduleLater = "later"
)

// Status constants
const (
	StatusDraft     = "draft"
	StatusActive    = "active"
	StatusCompleted = "completed"
	StatusArchived  = "archived"
)

// Helper methods

// GetSchedulePriority returns a numeric priority based on schedule
func (m *MarkdownWorkItem) GetSchedulePriority() int {
	switch m.Schedule {
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

// GetDisplayType returns a formatted type string
func (m *MarkdownWorkItem) GetDisplayType() string {
	switch m.Type {
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
		return m.Type
	}
}

// GetDisplaySchedule returns a formatted schedule string
func (m *MarkdownWorkItem) GetDisplaySchedule() string {
	switch m.Schedule {
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

// IsActive returns true if this item should be actively worked on
func (m *MarkdownWorkItem) IsActive() bool {
	return m.Schedule == ScheduleNow && m.Metadata.Status != StatusCompleted
}

// NeedsReview returns true if this item needs review (decisions past review date, etc)
func (m *MarkdownWorkItem) NeedsReview() bool {
	if m.Type == TypeDecision && m.Metadata.ReviewDate != nil {
		return time.Now().After(*m.Metadata.ReviewDate)
	}
	return false
}

// ToLegacyWorkItem converts to the legacy WorkItem format for compatibility
func (m *MarkdownWorkItem) ToLegacyWorkItem() WorkItem {
	// Map the new type system to legacy types
	legacyType := m.Type
	if m.Type == TypeUpdate {
		legacyType = "todo"
	} else if m.Type == TypeAnalysis {
		legacyType = "finding"
	}
	
	return WorkItem{
		ID:        m.ID,
		Type:      legacyType,
		Content:   m.Summary, // Use summary for legacy content field
		Status:    m.Metadata.Status,
		Context:   m.GitContext,
		SessionID: m.SessionNumber,
		Timestamp: m.CreatedAt.Format(time.RFC3339),
		Metadata: &WorkItemMetadata{
			Tags:     m.TechnicalTags,
			Priority: m.Schedule, // Map schedule to priority for legacy
		},
	}
}