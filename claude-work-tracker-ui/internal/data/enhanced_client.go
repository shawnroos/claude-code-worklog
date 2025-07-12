package data

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"claude-work-tracker-ui/internal/models"
)

// EnhancedClient provides access to work tracker data with markdown support
type EnhancedClient struct {
	*Client         // Embed existing client for backward compatibility
	markdownIO      *MarkdownIO
	useMarkdown     bool
}

// NewEnhancedClient creates a new enhanced data client
func NewEnhancedClient() *EnhancedClient {
	client := NewClient()
	
	// Initialize markdown IO with local work directory
	markdownIO := NewMarkdownIO(client.localWorkDir)
	
	return &EnhancedClient{
		Client:      client,
		markdownIO:  markdownIO,
		useMarkdown: true, // Default to markdown format
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

// generateShortID creates a short unique identifier
func generateShortID() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}