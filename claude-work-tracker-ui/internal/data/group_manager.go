package data

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"claude-work-tracker-ui/internal/models"
	"gopkg.in/yaml.v3"
)

// GroupManager handles lifecycle tracking and operations for Groups
type GroupManager struct {
	markdownIO *MarkdownIO
	baseDir    string
}

// NewGroupManager creates a new group manager
func NewGroupManager(markdownIO *MarkdownIO, baseDir string) *GroupManager {
	return &GroupManager{
		markdownIO: markdownIO,
		baseDir:    baseDir,
	}
}

// ReadGroup reads a Group from a markdown file
func (gm *GroupManager) ReadGroup(filepath string) (*models.Group, error) {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Use the same frontmatter regex from MarkdownIO
	frontmatterRegex := regexp.MustCompile(`(?s)^---\n(.*?)\n---\n(.*)$`)
	
	matches := frontmatterRegex.FindSubmatch(content)
	if len(matches) < 3 {
		return nil, fmt.Errorf("invalid markdown format: no frontmatter found")
	}

	frontmatter := matches[1]

	// Parse YAML frontmatter
	var group models.Group
	if err := yaml.Unmarshal(frontmatter, &group); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Set file info
	group.Filepath = filepath
	group.Filename = filepath[strings.LastIndex(filepath, "/")+1:]

	return &group, nil
}

// WriteGroup writes a Group to a file
func (gm *GroupManager) WriteGroup(group *models.Group) error {
	// Generate filename if not set
	if group.Filename == "" {
		group.Filename = gm.generateGroupFilename(group)
	}

	// Determine directory
	dir := filepath.Join(gm.baseDir, "groups")
	fullPath := filepath.Join(dir, group.Filename)

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate markdown content
	content, err := gm.generateGroupContent(group)
	if err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	// Write file
	if err := ioutil.WriteFile(fullPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	group.Filepath = fullPath
	return nil
}

// generateGroupContent creates the full markdown file content for Group with frontmatter
func (gm *GroupManager) generateGroupContent(group *models.Group) ([]byte, error) {
	var buf strings.Builder

	// Write frontmatter
	buf.WriteString("---\n")
	
	// Create a map to control field order for Group
	frontmatter := map[string]interface{}{
		"id":             group.ID,
		"name":           group.Name,
		"description":    group.Description,
		"theme":          group.Theme,
		"created_at":     group.CreatedAt,
		"updated_at":     group.UpdatedAt,
		"git_context":    group.GitContext,
		"session_number": group.SessionNumber,
		"artifact_ids":   group.ArtifactIDs,
		"work_refs":      group.WorkRefs,
		"technical_tags": group.TechnicalTags,
		"metadata":       group.Metadata,
	}
	
	// Use YAML encoder
	yamlBytes, err := yaml.Marshal(frontmatter)
	if err != nil {
		return nil, fmt.Errorf("failed to encode frontmatter: %w", err)
	}
	
	buf.Write(yamlBytes)
	buf.WriteString("---\n\n")
	
	// Write content (group description or generated summary)
	content := gm.generateGroupDescription(group)
	buf.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		buf.WriteString("\n")
	}

	return []byte(buf.String()), nil
}

// generateGroupDescription creates descriptive content for the group
func (gm *GroupManager) generateGroupDescription(group *models.Group) string {
	var buf strings.Builder
	
	buf.WriteString(fmt.Sprintf("# %s\n\n", group.Name))
	
	if group.Description != "" {
		buf.WriteString(fmt.Sprintf("%s\n\n", group.Description))
	}
	
	buf.WriteString(fmt.Sprintf("## Group Overview\n\n"))
	buf.WriteString(fmt.Sprintf("- **Theme**: %s\n", group.Theme))
	buf.WriteString(fmt.Sprintf("- **Artifacts**: %d items\n", group.Metadata.ArtifactCount))
	buf.WriteString(fmt.Sprintf("- **Status**: %s\n", group.Metadata.Status))
	buf.WriteString(fmt.Sprintf("- **Readiness Score**: %.2f\n", group.Metadata.ReadinessScore))
	buf.WriteString(fmt.Sprintf("- **Cohesion Score**: %.2f\n", group.Metadata.CohesionScore))
	
	if len(group.Metadata.TypeDistribution) > 0 {
		buf.WriteString("\n### Artifact Types\n\n")
		for artifactType, count := range group.Metadata.TypeDistribution {
			buf.WriteString(fmt.Sprintf("- %s: %d\n", artifactType, count))
		}
	}
	
	if len(group.ArtifactIDs) > 0 {
		buf.WriteString("\n### Included Artifacts\n\n")
		for _, artifactID := range group.ArtifactIDs {
			buf.WriteString(fmt.Sprintf("- %s\n", artifactID))
		}
	}
	
	if group.IsReadyForWork() {
		title, description, schedule, priority := group.GenerateWorkSuggestion()
		buf.WriteString("\n### Consolidation Suggestion\n\n")
		buf.WriteString(fmt.Sprintf("**Ready for Work Creation:**\n"))
		buf.WriteString(fmt.Sprintf("- Title: %s\n", title))
		buf.WriteString(fmt.Sprintf("- Schedule: %s\n", schedule))
		buf.WriteString(fmt.Sprintf("- Priority: %s\n", priority))
		buf.WriteString(fmt.Sprintf("- Description: %s\n", description))
	}
	
	return buf.String()
}

// generateGroupFilename creates a filename for Group: group-{brief-name}-{date}-{short-id}.md
func (gm *GroupManager) generateGroupFilename(group *models.Group) string {
	// Extract brief name from group name (first few words, sanitized)
	name := strings.ToLower(group.Name)
	words := strings.Fields(name)
	if len(words) > 3 {
		words = words[:3]
	}
	description := strings.Join(words, "-")
	
	// Sanitize description for filename
	description = strings.ReplaceAll(description, " ", "-")
	description = strings.ToLower(description)
	
	// Get date
	date := group.CreatedAt.Format("2006-01-02")
	
	// Get short ID (last 6 chars or generate)
	shortID := group.ID
	if len(shortID) > 6 {
		shortID = shortID[len(shortID)-6:]
	}
	
	return fmt.Sprintf("group-%s-%s-%s.md", description, date, shortID)
}

// ListAllGroups returns all Groups
func (gm *GroupManager) ListAllGroups() ([]*models.Group, error) {
	dir := filepath.Join(gm.baseDir, "groups")
	return gm.listGroupsFromDir(dir)
}

// listGroupsFromDir reads all Group markdown files from a directory
func (gm *GroupManager) listGroupsFromDir(dir string) ([]*models.Group, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*models.Group{}, nil
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var groups []*models.Group
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		filepath := filepath.Join(dir, file.Name())
		group, err := gm.ReadGroup(filepath)
		if err != nil {
			continue // Skip files that can't be parsed
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// CreateGroup creates a new group with the given artifacts
func (gm *GroupManager) CreateGroup(name, description, theme string, artifactIDs []string, technicalTags []string) (*models.Group, error) {
	now := time.Now()
	
	group := &models.Group{
		ID:            fmt.Sprintf("group-%s-%s", theme, now.Format("2006-01-02-150405")),
		Name:          name,
		Description:   description,
		Theme:         theme,
		CreatedAt:     now,
		UpdatedAt:     now,
		ArtifactIDs:   artifactIDs,
		TechnicalTags: technicalTags,
		Metadata: models.GroupMetadata{
			Status:           models.GroupStatusActive,
			ArtifactCount:    len(artifactIDs),
			TypeDistribution: make(map[string]int),
			ConfidenceScore:  0.8, // Default for manually created groups
		},
	}
	
	// Update scores
	group.CalculateScores()
	
	// Update type distribution if we have access to artifacts
	if allArtifacts, err := gm.markdownIO.ListAllArtifacts(); err == nil {
		group.UpdateTypeDistribution(allArtifacts)
	}
	
	// Save the group
	if err := gm.WriteGroup(group); err != nil {
		return nil, fmt.Errorf("failed to save group: %w", err)
	}
	
	return group, nil
}

// UpdateGroup updates an existing group
func (gm *GroupManager) UpdateGroup(group *models.Group) error {
	group.UpdatedAt = time.Now()
	
	// Recalculate scores
	group.CalculateScores()
	
	// Update type distribution
	if allArtifacts, err := gm.markdownIO.ListAllArtifacts(); err == nil {
		group.UpdateTypeDistribution(allArtifacts)
	}
	
	return gm.WriteGroup(group)
}

// DeleteGroup removes a group file
func (gm *GroupManager) DeleteGroup(groupID string) error {
	groups, err := gm.ListAllGroups()
	if err != nil {
		return fmt.Errorf("failed to list groups: %w", err)
	}
	
	for _, group := range groups {
		if group.ID == groupID {
			return os.Remove(group.Filepath)
		}
	}
	
	return fmt.Errorf("group not found: %s", groupID)
}

// GetGroupByID finds a group by its ID
func (gm *GroupManager) GetGroupByID(groupID string) (*models.Group, error) {
	groups, err := gm.ListAllGroups()
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}
	
	for _, group := range groups {
		if group.ID == groupID {
			return group, nil
		}
	}
	
	return nil, fmt.Errorf("group not found: %s", groupID)
}

// GetReadyGroups returns groups that are ready for work consolidation
func (gm *GroupManager) GetReadyGroups() ([]*models.Group, error) {
	allGroups, err := gm.ListAllGroups()
	if err != nil {
		return nil, err
	}
	
	var readyGroups []*models.Group
	for _, group := range allGroups {
		if group.IsReadyForWork() {
			readyGroups = append(readyGroups, group)
		}
	}
	
	// Sort by readiness score (highest first)
	sort.Slice(readyGroups, func(i, j int) bool {
		return readyGroups[i].Metadata.ReadinessScore > readyGroups[j].Metadata.ReadinessScore
	})
	
	return readyGroups, nil
}

// GetCandidateGroups returns groups that are candidates for operations (merge, split, etc.)
func (gm *GroupManager) GetCandidateGroups() (*GroupCandidates, error) {
	allGroups, err := gm.ListAllGroups()
	if err != nil {
		return nil, err
	}
	
	candidates := &GroupCandidates{
		ForMerging:  []*models.Group{},
		ForSplitting: []*models.Group{},
		ForReview:   []*models.Group{},
		Stale:       []*models.Group{},
	}
	
	for _, group := range allGroups {
		if group.ShouldMerge() {
			candidates.ForMerging = append(candidates.ForMerging, group)
		}
		
		if group.ShouldSplit() {
			candidates.ForSplitting = append(candidates.ForSplitting, group)
		}
		
		if group.Metadata.RecommendedForWork {
			candidates.ForReview = append(candidates.ForReview, group)
		}
		
		// Check for stale groups (no activity for extended period)
		if group.Metadata.LastModified != nil {
			daysSinceModified := time.Since(*group.Metadata.LastModified).Hours() / 24
			if daysSinceModified > 30 && group.Metadata.Status == models.GroupStatusActive {
				candidates.Stale = append(candidates.Stale, group)
			}
		}
	}
	
	return candidates, nil
}

// GroupCandidates represents different types of group operations that might be needed
type GroupCandidates struct {
	ForMerging   []*models.Group `json:"for_merging"`
	ForSplitting []*models.Group `json:"for_splitting"`
	ForReview    []*models.Group `json:"for_review"`
	Stale        []*models.Group `json:"stale"`
}

// ConsolidateGroupToWork creates a Work item from a Group
func (gm *GroupManager) ConsolidateGroupToWork(groupID, method string) (*models.Work, error) {
	group, err := gm.GetGroupByID(groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to find group: %w", err)
	}
	
	if !group.IsReadyForWork() {
		return nil, fmt.Errorf("group is not ready for work consolidation")
	}
	
	// Generate work suggestion
	title, description, schedule, priority := group.GenerateWorkSuggestion()
	
	// Create Work item
	now := time.Now()
	work := &models.Work{
		ID:            fmt.Sprintf("work-%s-%s", strings.ToLower(strings.ReplaceAll(title, " ", "-")), now.Format("2006-01-02-150405")),
		Title:         title,
		Description:   description,
		Schedule:      schedule,
		CreatedAt:     now,
		UpdatedAt:     now,
		GitContext:    group.GitContext,
		SessionNumber: group.SessionNumber,
		TechnicalTags: group.TechnicalTags,
		ArtifactRefs:  group.ArtifactIDs,
		GroupID:       group.ID,
		Metadata: models.WorkMetadata{
			Status:          models.WorkStatusActive,
			Priority:        priority,
			EstimatedEffort: models.WorkEffortMedium, // Default
			ArtifactCount:   len(group.ArtifactIDs),
		},
	}
	
	// Set started time if schedule is NOW
	if schedule == models.ScheduleNow {
		work.StartedAt = &now
		work.Metadata.Status = models.WorkStatusInProgress
	}
	
	// Calculate initial activity score
	work.CalculateActivityScore()
	
	// Save the work item
	if err := gm.markdownIO.WriteWork(work); err != nil {
		return nil, fmt.Errorf("failed to save work: %w", err)
	}
	
	// Mark group as consolidated
	group.MarkAsConsolidated(work.ID, method)
	if err := gm.UpdateGroup(group); err != nil {
		return nil, fmt.Errorf("failed to update group: %w", err)
	}
	
	// Update artifacts to reference the new work
	allArtifacts, err := gm.markdownIO.ListAllArtifacts()
	if err != nil {
		return nil, fmt.Errorf("failed to load artifacts: %w", err)
	}
	
	for _, artifact := range allArtifacts {
		// Check if this artifact is in the group
		for _, artifactID := range group.ArtifactIDs {
			if artifact.ID == artifactID {
				artifact.AssignToWork(work.ID)
				if err := gm.markdownIO.WriteArtifact(artifact); err != nil {
					// Log error but don't fail the whole operation
					fmt.Printf("Warning: failed to update artifact %s: %v\n", artifact.ID, err)
				}
				break
			}
		}
	}
	
	return work, nil
}

// AnalyzeGroupHealth analyzes the health of all groups and returns recommendations
func (gm *GroupManager) AnalyzeGroupHealth() (*GroupHealthReport, error) {
	allGroups, err := gm.ListAllGroups()
	if err != nil {
		return nil, err
	}
	
	report := &GroupHealthReport{
		TotalGroups:      len(allGroups),
		ActiveGroups:     0,
		ConsolidatedGroups: 0,
		ReadyForWork:     0,
		NeedAttention:    0,
		AverageReadiness: 0.0,
		AverageCohesion:  0.0,
		Recommendations:  []string{},
	}
	
	totalReadiness := 0.0
	totalCohesion := 0.0
	
	for _, group := range allGroups {
		switch group.Metadata.Status {
		case models.GroupStatusActive:
			report.ActiveGroups++
		case models.GroupStatusConsolidated:
			report.ConsolidatedGroups++
		}
		
		if group.IsReadyForWork() {
			report.ReadyForWork++
		}
		
		if group.ShouldMerge() || group.ShouldSplit() || group.Metadata.ActivityScore < 2.0 {
			report.NeedAttention++
		}
		
		totalReadiness += group.Metadata.ReadinessScore
		totalCohesion += group.Metadata.CohesionScore
	}
	
	if len(allGroups) > 0 {
		report.AverageReadiness = totalReadiness / float64(len(allGroups))
		report.AverageCohesion = totalCohesion / float64(len(allGroups))
	}
	
	// Generate recommendations
	if report.ReadyForWork > 0 {
		report.Recommendations = append(report.Recommendations, 
			fmt.Sprintf("%d groups are ready for work consolidation", report.ReadyForWork))
	}
	
	if report.NeedAttention > 0 {
		report.Recommendations = append(report.Recommendations, 
			fmt.Sprintf("%d groups need attention (merge/split/review)", report.NeedAttention))
	}
	
	if report.AverageReadiness < 0.5 {
		report.Recommendations = append(report.Recommendations, 
			"Overall group readiness is low - consider adding more artifacts or improving grouping")
	}
	
	if report.AverageCohesion < 0.6 {
		report.Recommendations = append(report.Recommendations, 
			"Overall group cohesion is low - review grouping criteria and consider splits/merges")
	}
	
	return report, nil
}

// GroupHealthReport provides an overview of group system health
type GroupHealthReport struct {
	TotalGroups        int      `json:"total_groups"`
	ActiveGroups       int      `json:"active_groups"`
	ConsolidatedGroups int      `json:"consolidated_groups"`
	ReadyForWork       int      `json:"ready_for_work"`
	NeedAttention      int      `json:"need_attention"`
	AverageReadiness   float64  `json:"average_readiness"`
	AverageCohesion    float64  `json:"average_cohesion"`
	Recommendations    []string `json:"recommendations"`
}

// RefreshAllGroupScores recalculates scores for all groups
func (gm *GroupManager) RefreshAllGroupScores() error {
	allGroups, err := gm.ListAllGroups()
	if err != nil {
		return fmt.Errorf("failed to load groups: %w", err)
	}
	
	for _, group := range allGroups {
		if err := gm.UpdateGroup(group); err != nil {
			return fmt.Errorf("failed to update group %s: %w", group.ID, err)
		}
	}
	
	return nil
}