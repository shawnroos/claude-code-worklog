package data

import (
	"fmt"
	"sort"

	"claude-work-tracker-ui/internal/models"
)

// AssociationManager handles tracking and resolution of relationships between Work and Artifacts
type AssociationManager struct {
	markdownIO *MarkdownIO
}

// NewAssociationManager creates a new association manager
func NewAssociationManager(markdownIO *MarkdownIO) *AssociationManager {
	return &AssociationManager{
		markdownIO: markdownIO,
	}
}

// AssociationGraph represents the full relationship graph
type AssociationGraph struct {
	WorkItems         []*models.Work     `json:"work_items"`
	Artifacts         []*models.Artifact `json:"artifacts"`
	WorkToArtifacts   map[string][]string `json:"work_to_artifacts"`   // Work ID -> Artifact IDs
	ArtifactToWork    map[string][]string `json:"artifact_to_work"`    // Artifact ID -> Work IDs
	ArtifactToArtifact map[string][]string `json:"artifact_to_artifact"` // Artifact ID -> Related Artifact IDs
	OrphanedArtifacts []string           `json:"orphaned_artifacts"`   // Artifact IDs with no connections
	TagClusters       map[string][]string `json:"tag_clusters"`        // Tag -> Item IDs
}

// BuildAssociationGraph creates a complete graph of all relationships
func (am *AssociationManager) BuildAssociationGraph() (*AssociationGraph, error) {
	// Load all work and artifacts
	allWork, err := am.markdownIO.ListAllWork()
	if err != nil {
		return nil, fmt.Errorf("failed to load work: %w", err)
	}

	allArtifacts, err := am.markdownIO.ListAllArtifacts()
	if err != nil {
		return nil, fmt.Errorf("failed to load artifacts: %w", err)
	}

	graph := &AssociationGraph{
		WorkItems:          allWork,
		Artifacts:          allArtifacts,
		WorkToArtifacts:    make(map[string][]string),
		ArtifactToWork:     make(map[string][]string),
		ArtifactToArtifact: make(map[string][]string),
		OrphanedArtifacts:  []string{},
		TagClusters:        make(map[string][]string),
	}

	// Build Work -> Artifact relationships
	for _, work := range allWork {
		graph.WorkToArtifacts[work.ID] = work.ArtifactRefs
		
		// Build reverse mapping
		for _, artifactID := range work.ArtifactRefs {
			if graph.ArtifactToWork[artifactID] == nil {
				graph.ArtifactToWork[artifactID] = []string{}
			}
			graph.ArtifactToWork[artifactID] = append(graph.ArtifactToWork[artifactID], work.ID)
		}
		
		// Add to tag clusters
		for _, tag := range work.TechnicalTags {
			if graph.TagClusters[tag] == nil {
				graph.TagClusters[tag] = []string{}
			}
			graph.TagClusters[tag] = append(graph.TagClusters[tag], work.ID)
		}
	}

	// Build Artifact -> Work and Artifact -> Artifact relationships
	for _, artifact := range allArtifacts {
		// Work refs from artifacts
		for _, workID := range artifact.WorkRefs {
			if graph.ArtifactToWork[artifact.ID] == nil {
				graph.ArtifactToWork[artifact.ID] = []string{}
			}
			// Avoid duplicates
			found := false
			for _, existingWorkID := range graph.ArtifactToWork[artifact.ID] {
				if existingWorkID == workID {
					found = true
					break
				}
			}
			if !found {
				graph.ArtifactToWork[artifact.ID] = append(graph.ArtifactToWork[artifact.ID], workID)
			}
		}

		// Artifact to artifact relationships
		graph.ArtifactToArtifact[artifact.ID] = artifact.RelatedArtifacts
		
		// Add to tag clusters
		for _, tag := range artifact.TechnicalTags {
			if graph.TagClusters[tag] == nil {
				graph.TagClusters[tag] = []string{}
			}
			graph.TagClusters[tag] = append(graph.TagClusters[tag], artifact.ID)
		}
		
		// Check if orphaned
		if len(artifact.WorkRefs) == 0 && 
		   len(artifact.RelatedArtifacts) == 0 && 
		   artifact.GroupID == "" &&
		   artifact.Metadata.ReferenceCount == 0 {
			graph.OrphanedArtifacts = append(graph.OrphanedArtifacts, artifact.ID)
		}
	}

	return graph, nil
}

// ResolveWorkArtifacts returns all artifacts associated with a work item
func (am *AssociationManager) ResolveWorkArtifacts(workID string) ([]*models.Artifact, error) {
	// Find the work item
	allWork, err := am.markdownIO.ListAllWork()
	if err != nil {
		return nil, fmt.Errorf("failed to load work: %w", err)
	}

	var targetWork *models.Work
	for _, work := range allWork {
		if work.ID == workID {
			targetWork = work
			break
		}
	}

	if targetWork == nil {
		return nil, fmt.Errorf("work item not found: %s", workID)
	}

	// Load all artifacts
	allArtifacts, err := am.markdownIO.ListAllArtifacts()
	if err != nil {
		return nil, fmt.Errorf("failed to load artifacts: %w", err)
	}

	// Build artifact lookup map
	artifactMap := make(map[string]*models.Artifact)
	for _, artifact := range allArtifacts {
		artifactMap[artifact.ID] = artifact
	}

	// Resolve artifact references
	var resolvedArtifacts []*models.Artifact
	for _, artifactID := range targetWork.ArtifactRefs {
		if artifact, exists := artifactMap[artifactID]; exists {
			resolvedArtifacts = append(resolvedArtifacts, artifact)
		}
	}

	// Sort by creation date (most recent first)
	sort.Slice(resolvedArtifacts, func(i, j int) bool {
		return resolvedArtifacts[i].CreatedAt.After(resolvedArtifacts[j].CreatedAt)
	})

	return resolvedArtifacts, nil
}

// ResolveArtifactWork returns all work items associated with an artifact
func (am *AssociationManager) ResolveArtifactWork(artifactID string) ([]*models.Work, error) {
	// Find the artifact
	allArtifacts, err := am.markdownIO.ListAllArtifacts()
	if err != nil {
		return nil, fmt.Errorf("failed to load artifacts: %w", err)
	}

	var targetArtifact *models.Artifact
	for _, artifact := range allArtifacts {
		if artifact.ID == artifactID {
			targetArtifact = artifact
			break
		}
	}

	if targetArtifact == nil {
		return nil, fmt.Errorf("artifact not found: %s", artifactID)
	}

	// Load all work
	allWork, err := am.markdownIO.ListAllWork()
	if err != nil {
		return nil, fmt.Errorf("failed to load work: %w", err)
	}

	// Build work lookup map
	workMap := make(map[string]*models.Work)
	for _, work := range allWork {
		workMap[work.ID] = work
	}

	// Resolve work references
	var resolvedWork []*models.Work
	for _, workID := range targetArtifact.WorkRefs {
		if work, exists := workMap[workID]; exists {
			resolvedWork = append(resolvedWork, work)
		}
	}

	// Sort by schedule priority (NOW -> NEXT -> LATER)
	sort.Slice(resolvedWork, func(i, j int) bool {
		return resolvedWork[i].GetSchedulePriority() < resolvedWork[j].GetSchedulePriority()
	})

	return resolvedWork, nil
}

// ResolveRelatedArtifacts returns artifacts related to the given artifact
func (am *AssociationManager) ResolveRelatedArtifacts(artifactID string) ([]*models.Artifact, error) {
	// Find the artifact
	allArtifacts, err := am.markdownIO.ListAllArtifacts()
	if err != nil {
		return nil, fmt.Errorf("failed to load artifacts: %w", err)
	}

	var targetArtifact *models.Artifact
	for _, artifact := range allArtifacts {
		if artifact.ID == artifactID {
			targetArtifact = artifact
			break
		}
	}

	if targetArtifact == nil {
		return nil, fmt.Errorf("artifact not found: %s", artifactID)
	}

	// Build artifact lookup map
	artifactMap := make(map[string]*models.Artifact)
	for _, artifact := range allArtifacts {
		artifactMap[artifact.ID] = artifact
	}

	// Resolve related artifact references
	var relatedArtifacts []*models.Artifact
	for _, relatedID := range targetArtifact.RelatedArtifacts {
		if artifact, exists := artifactMap[relatedID]; exists {
			relatedArtifacts = append(relatedArtifacts, artifact)
		}
	}

	return relatedArtifacts, nil
}

// CreateAssociation creates a new association between work and artifact
func (am *AssociationManager) CreateAssociation(workID, artifactID string) error {
	// Load work item
	allWork, err := am.markdownIO.ListAllWork()
	if err != nil {
		return fmt.Errorf("failed to load work: %w", err)
	}

	var targetWork *models.Work
	for _, work := range allWork {
		if work.ID == workID {
			targetWork = work
			break
		}
	}

	if targetWork == nil {
		return fmt.Errorf("work item not found: %s", workID)
	}

	// Load artifact
	allArtifacts, err := am.markdownIO.ListAllArtifacts()
	if err != nil {
		return fmt.Errorf("failed to load artifacts: %w", err)
	}

	var targetArtifact *models.Artifact
	for _, artifact := range allArtifacts {
		if artifact.ID == artifactID {
			targetArtifact = artifact
			break
		}
	}

	if targetArtifact == nil {
		return fmt.Errorf("artifact not found: %s", artifactID)
	}

	// Add association on both sides
	targetWork.AddArtifact(artifactID)
	targetArtifact.AssignToWork(workID)

	// Save both items
	if err := am.markdownIO.WriteWork(targetWork); err != nil {
		return fmt.Errorf("failed to save work: %w", err)
	}

	if err := am.markdownIO.WriteArtifact(targetArtifact); err != nil {
		return fmt.Errorf("failed to save artifact: %w", err)
	}

	return nil
}

// RemoveAssociation removes an association between work and artifact
func (am *AssociationManager) RemoveAssociation(workID, artifactID string) error {
	// Load work item
	allWork, err := am.markdownIO.ListAllWork()
	if err != nil {
		return fmt.Errorf("failed to load work: %w", err)
	}

	var targetWork *models.Work
	for _, work := range allWork {
		if work.ID == workID {
			targetWork = work
			break
		}
	}

	if targetWork == nil {
		return fmt.Errorf("work item not found: %s", workID)
	}

	// Load artifact
	allArtifacts, err := am.markdownIO.ListAllArtifacts()
	if err != nil {
		return fmt.Errorf("failed to load artifacts: %w", err)
	}

	var targetArtifact *models.Artifact
	for _, artifact := range allArtifacts {
		if artifact.ID == artifactID {
			targetArtifact = artifact
			break
		}
	}

	if targetArtifact == nil {
		return fmt.Errorf("artifact not found: %s", artifactID)
	}

	// Remove association on both sides
	targetWork.RemoveArtifact(artifactID)
	targetArtifact.UnassignFromWork(workID)

	// Save both items
	if err := am.markdownIO.WriteWork(targetWork); err != nil {
		return fmt.Errorf("failed to save work: %w", err)
	}

	if err := am.markdownIO.WriteArtifact(targetArtifact); err != nil {
		return fmt.Errorf("failed to save artifact: %w", err)
	}

	return nil
}

// FindSimilarByTags finds items with similar tags
func (am *AssociationManager) FindSimilarByTags(tags []string, excludeID string) ([]string, error) {
	graph, err := am.BuildAssociationGraph()
	if err != nil {
		return nil, err
	}

	similarItems := make(map[string]int) // Item ID -> tag overlap count
	
	for _, tag := range tags {
		if itemIDs, exists := graph.TagClusters[tag]; exists {
			for _, itemID := range itemIDs {
				if itemID != excludeID {
					similarItems[itemID]++
				}
			}
		}
	}

	// Sort by similarity (tag overlap count)
	type similarItem struct {
		ID    string
		Score int
	}

	var sortedItems []similarItem
	for itemID, score := range similarItems {
		sortedItems = append(sortedItems, similarItem{ID: itemID, Score: score})
	}

	sort.Slice(sortedItems, func(i, j int) bool {
		return sortedItems[i].Score > sortedItems[j].Score
	})

	var result []string
	for _, item := range sortedItems {
		result = append(result, item.ID)
	}

	return result, nil
}

// GetOrphanedArtifacts returns artifacts with no associations
func (am *AssociationManager) GetOrphanedArtifacts() ([]*models.Artifact, error) {
	graph, err := am.BuildAssociationGraph()
	if err != nil {
		return nil, err
	}

	var orphanedArtifacts []*models.Artifact
	
	// Build artifact lookup map
	artifactMap := make(map[string]*models.Artifact)
	for _, artifact := range graph.Artifacts {
		artifactMap[artifact.ID] = artifact
	}

	for _, orphanedID := range graph.OrphanedArtifacts {
		if artifact, exists := artifactMap[orphanedID]; exists {
			orphanedArtifacts = append(orphanedArtifacts, artifact)
		}
	}

	// Sort by creation date (oldest first - these need attention)
	sort.Slice(orphanedArtifacts, func(i, j int) bool {
		return orphanedArtifacts[i].CreatedAt.Before(orphanedArtifacts[j].CreatedAt)
	})

	return orphanedArtifacts, nil
}

// GetStaleWork returns work items that should be reviewed for decay
func (am *AssociationManager) GetStaleWork() ([]*models.Work, error) {
	allWork, err := am.markdownIO.ListAllWork()
	if err != nil {
		return nil, err
	}

	var staleWork []*models.Work
	for _, work := range allWork {
		work.CalculateActivityScore()
		if work.ShouldDecay() {
			staleWork = append(staleWork, work)
		}
	}

	// Sort by activity score (lowest first - most stale)
	sort.Slice(staleWork, func(i, j int) bool {
		return staleWork[i].Metadata.ActivityScore < staleWork[j].Metadata.ActivityScore
	})

	return staleWork, nil
}

// GetStaleArtifacts returns artifacts that should be reviewed for decay
func (am *AssociationManager) GetStaleArtifacts() ([]*models.Artifact, error) {
	allArtifacts, err := am.markdownIO.ListAllArtifacts()
	if err != nil {
		return nil, err
	}

	var staleArtifacts []*models.Artifact
	for _, artifact := range allArtifacts {
		artifact.CalculateActivityScore()
		if artifact.ShouldDecay() {
			staleArtifacts = append(staleArtifacts, artifact)
		}
	}

	// Sort by activity score (lowest first - most stale)
	sort.Slice(staleArtifacts, func(i, j int) bool {
		return staleArtifacts[i].Metadata.ActivityScore < staleArtifacts[j].Metadata.ActivityScore
	})

	return staleArtifacts, nil
}

// UpdateReferenceCount updates the reference count for an artifact
func (am *AssociationManager) UpdateReferenceCount(artifactID string, count int) error {
	allArtifacts, err := am.markdownIO.ListAllArtifacts()
	if err != nil {
		return fmt.Errorf("failed to load artifacts: %w", err)
	}

	for _, artifact := range allArtifacts {
		if artifact.ID == artifactID {
			artifact.UpdateReferenceCount(count)
			return am.markdownIO.WriteArtifact(artifact)
		}
	}

	return fmt.Errorf("artifact not found: %s", artifactID)
}

// RefreshAllActivityScores recalculates activity scores for all items
func (am *AssociationManager) RefreshAllActivityScores() error {
	// Update all work items
	allWork, err := am.markdownIO.ListAllWork()
	if err != nil {
		return fmt.Errorf("failed to load work: %w", err)
	}

	for _, work := range allWork {
		work.CalculateActivityScore()
		if err := am.markdownIO.WriteWork(work); err != nil {
			return fmt.Errorf("failed to save work %s: %w", work.ID, err)
		}
	}

	// Update all artifacts
	allArtifacts, err := am.markdownIO.ListAllArtifacts()
	if err != nil {
		return fmt.Errorf("failed to load artifacts: %w", err)
	}

	for _, artifact := range allArtifacts {
		artifact.CalculateActivityScore()
		if err := am.markdownIO.WriteArtifact(artifact); err != nil {
			return fmt.Errorf("failed to save artifact %s: %w", artifact.ID, err)
		}
	}

	return nil
}

// GetAssociationSummary returns a summary of all associations
func (am *AssociationManager) GetAssociationSummary() (*AssociationSummary, error) {
	graph, err := am.BuildAssociationGraph()
	if err != nil {
		return nil, err
	}

	summary := &AssociationSummary{
		TotalWork:          len(graph.WorkItems),
		TotalArtifacts:     len(graph.Artifacts),
		TotalAssociations:  0,
		OrphanedArtifacts:  len(graph.OrphanedArtifacts),
		TagClusters:        len(graph.TagClusters),
		MostConnectedWork:  "",
		MostConnectedArtifact: "",
	}

	// Count total associations
	for _, artifactRefs := range graph.WorkToArtifacts {
		summary.TotalAssociations += len(artifactRefs)
	}

	// Find most connected work
	maxWorkConnections := 0
	for workID, artifactRefs := range graph.WorkToArtifacts {
		if len(artifactRefs) > maxWorkConnections {
			maxWorkConnections = len(artifactRefs)
			summary.MostConnectedWork = workID
		}
	}

	// Find most connected artifact
	maxArtifactConnections := 0
	for artifactID, workRefs := range graph.ArtifactToWork {
		if len(workRefs) > maxArtifactConnections {
			maxArtifactConnections = len(workRefs)
			summary.MostConnectedArtifact = artifactID
		}
	}

	return summary, nil
}

// AssociationSummary provides a high-level view of the association system
type AssociationSummary struct {
	TotalWork             int    `json:"total_work"`
	TotalArtifacts        int    `json:"total_artifacts"`
	TotalAssociations     int    `json:"total_associations"`
	OrphanedArtifacts     int    `json:"orphaned_artifacts"`
	TagClusters           int    `json:"tag_clusters"`
	MostConnectedWork     string `json:"most_connected_work"`
	MostConnectedArtifact string `json:"most_connected_artifact"`
}