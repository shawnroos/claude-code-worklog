package data

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"claude-work-tracker-ui/internal/models"
)

// EnhancedClient provides access to work tracker data with hierarchical Work + Artifacts support
type EnhancedClient struct {
	*Client               // Embed existing client for backward compatibility
	markdownIO            *MarkdownIO
	associationManager    *AssociationManager
	groupManager          *GroupManager
	lifecycleManager      *LifecycleManager
	useMarkdown           bool
	useHierarchy          bool // Enable new Work + Artifacts hierarchy
}

// NewEnhancedClient creates a new enhanced data client
func NewEnhancedClient() *EnhancedClient {
	client := NewClient()
	
	// Initialize markdown IO with local work directory
	markdownIO := NewMarkdownIO(client.localWorkDir)
	
	// Initialize hierarchy managers
	associationManager := NewAssociationManager(markdownIO)
	groupManager := NewGroupManager(markdownIO, client.localWorkDir)
	lifecycleManager := NewLifecycleManager(markdownIO, associationManager, groupManager)
	
	return &EnhancedClient{
		Client:             client,
		markdownIO:         markdownIO,
		associationManager: associationManager,
		groupManager:       groupManager,
		lifecycleManager:   lifecycleManager,
		useMarkdown:        true,  // Default to markdown format
		useHierarchy:       true,  // Default to new hierarchy
	}
}

// GetAllWorkItemsEnhanced returns work items from both markdown and legacy JSON sources
func (c *EnhancedClient) GetAllWorkItemsEnhanced() ([]models.WorkItem, error) {
	var allItems []models.WorkItem
	
	if c.useMarkdown {
		// Get markdown items
		markdownItems, err := c.markdownIO.ListAllWorkItems()
		if err == nil {
			// Convert markdown items to legacy format for UI compatibility
			for _, mdItem := range markdownItems {
				allItems = append(allItems, mdItem.ToLegacyWorkItem())
			}
		}
	}
	
	// Also get legacy JSON items for backward compatibility
	legacyItems, err := c.GetAllWorkItems()
	if err == nil {
		allItems = append(allItems, legacyItems...)
	}
	
	// Sort by timestamp (newest first)
	sort.Slice(allItems, func(i, j int) bool {
		timeI, errI := time.Parse(time.RFC3339, allItems[i].Timestamp)
		timeJ, errJ := time.Parse(time.RFC3339, allItems[j].Timestamp)
		if errI != nil || errJ != nil {
			return false
		}
		return timeI.After(timeJ)
	})
	
	// Remove duplicates (prefer markdown version)
	seen := make(map[string]bool)
	var dedupedItems []models.WorkItem
	for _, item := range allItems {
		if !seen[item.ID] {
			seen[item.ID] = true
			dedupedItems = append(dedupedItems, item)
		}
	}
	
	return dedupedItems, nil
}

// GetWorkItemsBySchedule returns work items filtered by NOW/NEXT/LATER schedule
func (c *EnhancedClient) GetWorkItemsBySchedule(schedule string) ([]*models.MarkdownWorkItem, error) {
	if !c.useMarkdown {
		return []*models.MarkdownWorkItem{}, nil
	}
	
	return c.markdownIO.ListWorkItems(strings.ToLower(schedule))
}

// GetWorkItemsByTypeAndSchedule returns items filtered by both type and schedule
func (c *EnhancedClient) GetWorkItemsByTypeAndSchedule(itemType, schedule string) ([]*models.MarkdownWorkItem, error) {
	items, err := c.GetWorkItemsBySchedule(schedule)
	if err != nil {
		return nil, err
	}
	
	var filtered []*models.MarkdownWorkItem
	for _, item := range items {
		if item.Type == itemType {
			filtered = append(filtered, item)
		}
	}
	
	return filtered, nil
}

// CreateWorkItem creates a new work item in markdown format
func (c *EnhancedClient) CreateWorkItem(itemType, summary, content, schedule string, tags []string) (*models.MarkdownWorkItem, error) {
	if !c.useMarkdown {
		return nil, fmt.Errorf("markdown format not enabled")
	}
	
	// Generate ID
	id := fmt.Sprintf("%s-%d-%s", itemType, time.Now().UnixNano(), generateShortID())
	
	// Get git context
	gitContext := models.GitContext{
		Branch:           c.scanner.GetProjectRoot(),
		Worktree:         c.currentWorkingDir,
		WorkingDirectory: c.currentWorkingDir,
	}
	
	// Create work item
	item := &models.MarkdownWorkItem{
		ID:            id,
		Type:          itemType,
		Summary:       summary,
		Content:       content,
		Schedule:      schedule,
		TechnicalTags: tags,
		SessionNumber: fmt.Sprintf("session-%d", time.Now().Unix()),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		GitContext:    gitContext,
		Metadata: models.MarkdownMetadata{
			Status: models.StatusActive,
		},
	}
	
	// Set type-specific metadata
	switch itemType {
	case models.TypePlan:
		item.Metadata.ImplementationStatus = "not_started"
	case models.TypeDecision:
		item.Metadata.EnforcementActive = true
	case models.TypeProposal:
		item.Metadata.ApprovalStatus = "pending"
	}
	
	// Write to disk
	if err := c.markdownIO.WriteMarkdownWorkItem(item); err != nil {
		return nil, fmt.Errorf("failed to write work item: %w", err)
	}
	
	return item, nil
}

// UpdateWorkItemSchedule moves a work item between NOW/NEXT/LATER
func (c *EnhancedClient) UpdateWorkItemSchedule(itemID, newSchedule string) error {
	if !c.useMarkdown {
		return fmt.Errorf("markdown format not enabled")
	}
	
	// Find the item
	allItems, err := c.markdownIO.ListAllWorkItems()
	if err != nil {
		return err
	}
	
	for _, item := range allItems {
		if item.ID == itemID {
			return c.markdownIO.UpdateSchedule(item, newSchedule)
		}
	}
	
	return fmt.Errorf("work item not found: %s", itemID)
}

// CompleteWorkItem moves a work item to completed status
func (c *EnhancedClient) CompleteWorkItem(itemID string) error {
	if !c.useMarkdown {
		return fmt.Errorf("markdown format not enabled")
	}
	
	// Find the item
	allItems, err := c.markdownIO.ListAllWorkItems()
	if err != nil {
		return err
	}
	
	for _, item := range allItems {
		if item.ID == itemID {
			item.Metadata.Status = models.StatusCompleted
			item.UpdatedAt = time.Now()
			
			// Write updated item
			if err := c.markdownIO.WriteMarkdownWorkItem(item); err != nil {
				return err
			}
			
			// Move to completed directory
			return c.markdownIO.MoveToCompleted(item)
		}
	}
	
	return fmt.Errorf("work item not found: %s", itemID)
}

// SearchWorkItems searches across all work items
func (c *EnhancedClient) SearchWorkItems(query string) ([]models.WorkItem, error) {
	var results []models.WorkItem
	
	if c.useMarkdown {
		// Search markdown items
		markdownResults, err := c.markdownIO.SearchWorkItems(query)
		if err == nil {
			for _, mdItem := range markdownResults {
				results = append(results, mdItem.ToLegacyWorkItem())
			}
		}
	}
	
	// Also search legacy items
	legacyResults, err := c.Client.SearchWorkItems(query)
	if err == nil {
		results = append(results, legacyResults...)
	}
	
	return results, nil
}

// GetScheduleOverview returns counts of items in each schedule category
func (c *EnhancedClient) GetScheduleOverview() (map[string]int, error) {
	overview := map[string]int{
		models.ScheduleNow:   0,
		models.ScheduleNext:  0,
		models.ScheduleLater: 0,
	}
	
	if !c.useMarkdown {
		return overview, nil
	}
	
	allItems, err := c.markdownIO.ListAllWorkItems()
	if err != nil {
		return overview, err
	}
	
	for _, item := range allItems {
		if item.Metadata.Status != models.StatusCompleted {
			overview[item.Schedule]++
		}
	}
	
	return overview, nil
}

// GetTypeOverview returns counts of items by type
func (c *EnhancedClient) GetTypeOverview() (map[string]int, error) {
	overview := map[string]int{
		models.TypePlan:     0,
		models.TypeProposal: 0,
		models.TypeAnalysis: 0,
		models.TypeUpdate:   0,
		models.TypeDecision: 0,
	}
	
	if !c.useMarkdown {
		return overview, nil
	}
	
	allItems, err := c.markdownIO.ListAllWorkItems()
	if err != nil {
		return overview, err
	}
	
	for _, item := range allItems {
		if item.Metadata.Status != models.StatusCompleted {
			overview[item.Type]++
		}
	}
	
	return overview, nil
}

// GetMarkdownIO returns the markdown IO instance for direct access
func (c *EnhancedClient) GetMarkdownIO() *MarkdownIO {
	return c.markdownIO
}

// generateShortID creates a short unique identifier
func generateShortID() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// === Hierarchical Work + Artifacts Methods ===

// GetAllWork returns all Work items from the hierarchical system
func (c *EnhancedClient) GetAllWork() ([]*models.Work, error) {
	if !c.useHierarchy {
		return []*models.Work{}, nil
	}
	return c.markdownIO.ListAllWork()
}

// GetWorkBySchedule returns Work items filtered by schedule (NOW/NEXT/LATER)
func (c *EnhancedClient) GetWorkBySchedule(schedule string) ([]*models.Work, error) {
	if !c.useHierarchy {
		return []*models.Work{}, nil
	}
	return c.markdownIO.ListWork(strings.ToLower(schedule))
}

// GetAllArtifacts returns all Artifacts from the hierarchical system
func (c *EnhancedClient) GetAllArtifacts() ([]*models.Artifact, error) {
	if !c.useHierarchy {
		return []*models.Artifact{}, nil
	}
	return c.markdownIO.ListAllArtifacts()
}

// GetArtifactsByType returns Artifacts filtered by type
func (c *EnhancedClient) GetArtifactsByType(artifactType string) ([]*models.Artifact, error) {
	if !c.useHierarchy {
		return []*models.Artifact{}, nil
	}
	return c.markdownIO.ListArtifacts(artifactType)
}

// GetAllGroups returns all Groups
func (c *EnhancedClient) GetAllGroups() ([]*models.Group, error) {
	if !c.useHierarchy {
		return []*models.Group{}, nil
	}
	return c.groupManager.ListAllGroups()
}

// CreateWork creates a new Work item
func (c *EnhancedClient) CreateWork(title, description, schedule, priority string, tags []string, artifactRefs []string) (*models.Work, error) {
	if !c.useHierarchy {
		return nil, fmt.Errorf("hierarchy not enabled")
	}
	
	// Generate ID
	id := fmt.Sprintf("work-%s-%d", strings.ToLower(strings.ReplaceAll(title, " ", "-")), time.Now().UnixNano())
	
	// Get git context
	gitContext := models.GitContext{
		Branch:           c.scanner.GetProjectRoot(),
		Worktree:         c.currentWorkingDir,
		WorkingDirectory: c.currentWorkingDir,
	}
	
	now := time.Now()
	work := &models.Work{
		ID:            id,
		Title:         title,
		Description:   description,
		Schedule:      schedule,
		CreatedAt:     now,
		UpdatedAt:     now,
		GitContext:    gitContext,
		SessionNumber: fmt.Sprintf("session-%d", now.Unix()),
		TechnicalTags: tags,
		ArtifactRefs:  artifactRefs,
		Metadata: models.WorkMetadata{
			Status:          models.WorkStatusActive,
			Priority:        priority,
			EstimatedEffort: models.WorkEffortMedium,
			ArtifactCount:   len(artifactRefs),
		},
	}
	
	// Set started time if schedule is NOW
	if schedule == models.ScheduleNow {
		work.StartedAt = &now
		work.Metadata.Status = models.WorkStatusInProgress
	}
	
	// Calculate initial activity score
	work.CalculateActivityScore()
	
	// Write to disk
	if err := c.markdownIO.WriteWork(work); err != nil {
		return nil, fmt.Errorf("failed to write work: %w", err)
	}
	
	return work, nil
}

// CreateArtifact creates a new Artifact
func (c *EnhancedClient) CreateArtifact(artifactType, summary, content string, tags []string) (*models.Artifact, error) {
	if !c.useHierarchy {
		return nil, fmt.Errorf("hierarchy not enabled")
	}
	
	// Generate ID
	id := fmt.Sprintf("%s-%d-%s", artifactType, time.Now().UnixNano(), generateShortID())
	
	// Get git context
	gitContext := models.GitContext{
		Branch:           c.scanner.GetProjectRoot(),
		Worktree:         c.currentWorkingDir,
		WorkingDirectory: c.currentWorkingDir,
	}
	
	now := time.Now()
	artifact := &models.Artifact{
		ID:            id,
		Type:          artifactType,
		Summary:       summary,
		TechnicalTags: tags,
		CreatedAt:     now,
		UpdatedAt:     now,
		GitContext:    gitContext,
		SessionNumber: fmt.Sprintf("session-%d", now.Unix()),
		Content:       content,
		Metadata: models.ArtifactMetadata{
			Status: models.ArtifactStatusActive,
		},
	}
	
	// Set type-specific metadata
	switch artifactType {
	case models.TypePlan:
		artifact.Metadata.ImplementationStatus = "not_started"
	case models.TypeDecision:
		artifact.Metadata.EnforcementActive = true
	case models.TypeProposal:
		artifact.Metadata.ApprovalStatus = "pending"
	}
	
	// Calculate initial activity score
	artifact.CalculateActivityScore()
	
	// Write to disk
	if err := c.markdownIO.WriteArtifact(artifact); err != nil {
		return nil, fmt.Errorf("failed to write artifact: %w", err)
	}
	
	return artifact, nil
}

// CreateGroup creates a new Group
func (c *EnhancedClient) CreateGroup(name, description, theme string, artifactIDs []string, tags []string) (*models.Group, error) {
	if !c.useHierarchy {
		return nil, fmt.Errorf("hierarchy not enabled")
	}
	return c.groupManager.CreateGroup(name, description, theme, artifactIDs, tags)
}

// UpdateWorkSchedule moves a Work item between schedules
func (c *EnhancedClient) UpdateWorkSchedule(workID, newSchedule string) error {
	if !c.useHierarchy {
		return fmt.Errorf("hierarchy not enabled")
	}
	
	allWork, err := c.markdownIO.ListAllWork()
	if err != nil {
		return err
	}
	
	for _, work := range allWork {
		if work.ID == workID {
			work.Schedule = newSchedule
			work.UpdatedAt = time.Now()
			
			// Update status based on schedule
			if newSchedule == models.ScheduleNow && work.Metadata.Status == models.WorkStatusActive {
				work.Metadata.Status = models.WorkStatusInProgress
				now := time.Now()
				work.StartedAt = &now
			}
			
			return c.markdownIO.WriteWork(work)
		}
	}
	
	return fmt.Errorf("work not found: %s", workID)
}

// CompleteWork marks a Work item as completed
func (c *EnhancedClient) CompleteWork(workID string) error {
	if !c.useHierarchy {
		return fmt.Errorf("hierarchy not enabled")
	}
	
	allWork, err := c.markdownIO.ListAllWork()
	if err != nil {
		return err
	}
	
	for _, work := range allWork {
		if work.ID == workID {
			work.MarkAsCompleted()
			return c.markdownIO.WriteWork(work)
		}
	}
	
	return fmt.Errorf("work not found: %s", workID)
}

// CreateAssociation creates an association between Work and Artifact
func (c *EnhancedClient) CreateAssociation(workID, artifactID string) error {
	if !c.useHierarchy {
		return fmt.Errorf("hierarchy not enabled")
	}
	return c.associationManager.CreateAssociation(workID, artifactID)
}

// RemoveAssociation removes an association between Work and Artifact
func (c *EnhancedClient) RemoveAssociation(workID, artifactID string) error {
	if !c.useHierarchy {
		return fmt.Errorf("hierarchy not enabled")
	}
	return c.associationManager.RemoveAssociation(workID, artifactID)
}

// GetWorkArtifacts returns all artifacts associated with a Work item
func (c *EnhancedClient) GetWorkArtifacts(workID string) ([]*models.Artifact, error) {
	if !c.useHierarchy {
		return []*models.Artifact{}, nil
	}
	return c.associationManager.ResolveWorkArtifacts(workID)
}

// GetArtifactWork returns all Work items associated with an Artifact
func (c *EnhancedClient) GetArtifactWork(artifactID string) ([]*models.Work, error) {
	if !c.useHierarchy {
		return []*models.Work{}, nil
	}
	return c.associationManager.ResolveArtifactWork(artifactID)
}

// GetAssociationGraph returns the complete association graph
func (c *EnhancedClient) GetAssociationGraph() (*AssociationGraph, error) {
	if !c.useHierarchy {
		return nil, fmt.Errorf("hierarchy not enabled")
	}
	return c.associationManager.BuildAssociationGraph()
}

// ConsolidateGroupToWork converts a Group into a Work item
func (c *EnhancedClient) ConsolidateGroupToWork(groupID, method string) (*models.Work, error) {
	if !c.useHierarchy {
		return nil, fmt.Errorf("hierarchy not enabled")
	}
	return c.groupManager.ConsolidateGroupToWork(groupID, method)
}

// GetReadyGroups returns groups ready for Work consolidation
func (c *EnhancedClient) GetReadyGroups() ([]*models.Group, error) {
	if !c.useHierarchy {
		return []*models.Group{}, nil
	}
	return c.groupManager.GetReadyGroups()
}

// GetOrphanedArtifacts returns artifacts with no associations
func (c *EnhancedClient) GetOrphanedArtifacts() ([]*models.Artifact, error) {
	if !c.useHierarchy {
		return []*models.Artifact{}, nil
	}
	return c.associationManager.GetOrphanedArtifacts()
}

// GetStaleWork returns Work items flagged for decay
func (c *EnhancedClient) GetStaleWork() ([]*models.Work, error) {
	if !c.useHierarchy {
		return []*models.Work{}, nil
	}
	return c.associationManager.GetStaleWork()
}

// GetStaleArtifacts returns Artifacts flagged for decay
func (c *EnhancedClient) GetStaleArtifacts() ([]*models.Artifact, error) {
	if !c.useHierarchy {
		return []*models.Artifact{}, nil
	}
	return c.associationManager.GetStaleArtifacts()
}

// SearchWork searches Work items
func (c *EnhancedClient) SearchWork(query string) ([]*models.Work, error) {
	if !c.useHierarchy {
		return []*models.Work{}, nil
	}
	return c.markdownIO.SearchWork(query)
}

// SearchArtifacts searches Artifacts
func (c *EnhancedClient) SearchArtifacts(query string) ([]*models.Artifact, error) {
	if !c.useHierarchy {
		return []*models.Artifact{}, nil
	}
	return c.markdownIO.SearchArtifacts(query)
}

// GetHierarchyOverview returns a summary of the hierarchical system
func (c *EnhancedClient) GetHierarchyOverview() (*HierarchyOverview, error) {
	if !c.useHierarchy {
		return &HierarchyOverview{}, nil
	}
	
	// Get counts
	allWork, err := c.markdownIO.ListAllWork()
	if err != nil {
		return nil, err
	}
	
	allArtifacts, err := c.markdownIO.ListAllArtifacts()
	if err != nil {
		return nil, err
	}
	
	allGroups, err := c.groupManager.ListAllGroups()
	if err != nil {
		return nil, err
	}
	
	// Count by schedule
	scheduleOverview := map[string]int{
		models.ScheduleNow:   0,
		models.ScheduleNext:  0,
		models.ScheduleLater: 0,
	}
	
	for _, work := range allWork {
		if !work.IsCompleted() {
			scheduleOverview[work.Schedule]++
		}
	}
	
	// Count by artifact type
	typeOverview := map[string]int{
		models.TypePlan:     0,
		models.TypeProposal: 0,
		models.TypeAnalysis: 0,
		models.TypeUpdate:   0,
		models.TypeDecision: 0,
	}
	
	for _, artifact := range allArtifacts {
		if !artifact.IsStale() {
			typeOverview[artifact.Type]++
		}
	}
	
	// Get association summary
	associationSummary, err := c.associationManager.GetAssociationSummary()
	if err != nil {
		return nil, err
	}
	
	return &HierarchyOverview{
		WorkOverview:     scheduleOverview,
		ArtifactOverview: typeOverview,
		TotalWork:        len(allWork),
		TotalArtifacts:   len(allArtifacts),
		TotalGroups:      len(allGroups),
		Associations:     associationSummary,
	}, nil
}

// HierarchyOverview provides a high-level view of the hierarchical system
type HierarchyOverview struct {
	WorkOverview     map[string]int        `json:"work_overview"`
	ArtifactOverview map[string]int        `json:"artifact_overview"`
	TotalWork        int                   `json:"total_work"`
	TotalArtifacts   int                   `json:"total_artifacts"`
	TotalGroups      int                   `json:"total_groups"`
	Associations     *AssociationSummary   `json:"associations"`
}

// GetAssociationManager returns the association manager for direct access
func (c *EnhancedClient) GetAssociationManager() *AssociationManager {
	return c.associationManager
}

// GetGroupManager returns the group manager for direct access
func (c *EnhancedClient) GetGroupManager() *GroupManager {
	return c.groupManager
}

// GetLifecycleManager returns the lifecycle manager for direct access
func (c *EnhancedClient) GetLifecycleManager() *LifecycleManager {
	return c.lifecycleManager
}

// EnableHierarchy enables/disables the hierarchical system
func (c *EnhancedClient) EnableHierarchy(enable bool) {
	c.useHierarchy = enable
}

// IsHierarchyEnabled returns whether the hierarchical system is enabled
func (c *EnhancedClient) IsHierarchyEnabled() bool {
	return c.useHierarchy
}

// === Lifecycle Management Methods ===

// AnalyzeSystemHealth performs comprehensive decay analysis
func (c *EnhancedClient) AnalyzeSystemHealth() (*DecayAnalysis, error) {
	if !c.useHierarchy {
		return nil, fmt.Errorf("hierarchy not enabled")
	}
	return c.lifecycleManager.AnalyzeDecay()
}

// GetSystemHealthMetrics returns current system health metrics
func (c *EnhancedClient) GetSystemHealthMetrics() (*SystemHealthMetrics, error) {
	if !c.useHierarchy {
		return nil, fmt.Errorf("hierarchy not enabled")
	}
	return c.lifecycleManager.GetHealthMetrics()
}

// ExecuteCleanupAction performs a specific cleanup action
func (c *EnhancedClient) ExecuteCleanupAction(action CleanupAction) error {
	if !c.useHierarchy {
		return fmt.Errorf("hierarchy not enabled")
	}
	return c.lifecycleManager.ExecuteCleanupAction(action)
}

// AutoCleanup executes all auto-safe cleanup actions
func (c *EnhancedClient) AutoCleanup() (*AutoCleanupResult, error) {
	if !c.useHierarchy {
		return nil, fmt.Errorf("hierarchy not enabled")
	}
	return c.lifecycleManager.AutoCleanup()
}

// RefreshActivityScores updates activity scores for all items
func (c *EnhancedClient) RefreshActivityScores() error {
	if !c.useHierarchy {
		return fmt.Errorf("hierarchy not enabled")
	}
	return c.lifecycleManager.RefreshAllActivityScores()
}