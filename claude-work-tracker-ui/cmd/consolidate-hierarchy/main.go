package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"claude-work-tracker-ui/internal/data"
	"claude-work-tracker-ui/internal/models"
)

// ConsolidationEngine handles finding and creating Work from Artifacts
type ConsolidationEngine struct {
	client         *data.EnhancedClient
	associationMgr *data.AssociationManager
	groupMgr       *data.GroupManager
}

// NewConsolidationEngine creates a new hierarchy-aware consolidation engine
func NewConsolidationEngine() *ConsolidationEngine {
	client := data.NewEnhancedClient()
	
	return &ConsolidationEngine{
		client:         client,
		associationMgr: client.GetAssociationManager(),
		groupMgr:       client.GetGroupManager(),
	}
}

// ArtifactCluster represents a group of related artifacts that could become Work
type ArtifactCluster struct {
	Artifacts        []*models.Artifact `json:"artifacts"`
	SimilarityScore  float64           `json:"similarity_score"`
	CohesionScore    float64           `json:"cohesion_score"`
	Theme            string            `json:"theme"`
	SuggestedTitle   string            `json:"suggested_title"`
	SuggestedSchedule string           `json:"suggested_schedule"`
	SuggestedPriority string           `json:"suggested_priority"`
	Reason           string            `json:"reason"`
	GroupID          string            `json:"group_id,omitempty"`
}

// FindArtifactClusters identifies groups of artifacts that should become Work
func (c *ConsolidationEngine) FindArtifactClusters() ([]ArtifactCluster, error) {
	// Get all orphaned artifacts (not assigned to Work)
	orphanedArtifacts, err := c.client.GetOrphanedArtifacts()
	if err != nil {
		return nil, fmt.Errorf("failed to get orphaned artifacts: %w", err)
	}

	if len(orphanedArtifacts) < 2 {
		return []ArtifactCluster{}, nil
	}

	var clusters []ArtifactCluster

	// 1. Find clusters by shared tags
	tagClusters := c.findTagBasedClusters(orphanedArtifacts)
	clusters = append(clusters, tagClusters...)

	// 2. Find clusters by content similarity
	contentClusters := c.findContentBasedClusters(orphanedArtifacts)
	clusters = append(clusters, contentClusters...)

	// 3. Find clusters by type and theme
	typeClusters := c.findTypeBasedClusters(orphanedArtifacts)
	clusters = append(clusters, typeClusters...)

	// Remove duplicates and overlaps
	clusters = c.deduplicateClusters(clusters)

	// Sort by potential impact (cohesion * count)
	sort.Slice(clusters, func(i, j int) bool {
		scoreI := clusters[i].CohesionScore * float64(len(clusters[i].Artifacts))
		scoreJ := clusters[j].CohesionScore * float64(len(clusters[j].Artifacts))
		return scoreI > scoreJ
	})

	return clusters, nil
}

// findTagBasedClusters groups artifacts with overlapping technical tags
func (c *ConsolidationEngine) findTagBasedClusters(artifacts []*models.Artifact) []ArtifactCluster {
	var clusters []ArtifactCluster
	
	// Group by shared tags (need at least 2 shared tags)
	tagGroups := make(map[string][]*models.Artifact)
	
	for _, artifact := range artifacts {
		for _, tag := range artifact.TechnicalTags {
			if _, exists := tagGroups[tag]; !exists {
				tagGroups[tag] = []*models.Artifact{}
			}
			tagGroups[tag] = append(tagGroups[tag], artifact)
		}
	}
	
	// Find groups with multiple artifacts
	for tag, group := range tagGroups {
		if len(group) >= 2 {
			cluster := ArtifactCluster{
				Artifacts:      group,
				Theme:         tag,
				SuggestedTitle: fmt.Sprintf("Work on %s", tag),
				SuggestedSchedule: c.determineScheduleFromArtifacts(group),
				SuggestedPriority: c.determinePriorityFromArtifacts(group),
				Reason:        fmt.Sprintf("Shared tag: %s", tag),
			}
			
			cluster.SimilarityScore = c.calculateTagSimilarity(group)
			cluster.CohesionScore = c.calculateCohesion(group)
			
			if cluster.CohesionScore > 0.5 {
				clusters = append(clusters, cluster)
			}
		}
	}
	
	return clusters
}

// findContentBasedClusters groups artifacts with similar content
func (c *ConsolidationEngine) findContentBasedClusters(artifacts []*models.Artifact) []ArtifactCluster {
	var clusters []ArtifactCluster
	
	// Compare each artifact with others for content similarity
	for i := 0; i < len(artifacts); i++ {
		var similarGroup []*models.Artifact
		similarGroup = append(similarGroup, artifacts[i])
		
		for j := i + 1; j < len(artifacts); j++ {
			similarity := c.calculateContentSimilarity(artifacts[i], artifacts[j])
			if similarity > 0.6 {
				similarGroup = append(similarGroup, artifacts[j])
			}
		}
		
		if len(similarGroup) >= 2 {
			theme := c.extractTheme(similarGroup)
			cluster := ArtifactCluster{
				Artifacts:      similarGroup,
				Theme:         theme,
				SuggestedTitle: fmt.Sprintf("Work on %s", theme),
				SuggestedSchedule: c.determineScheduleFromArtifacts(similarGroup),
				SuggestedPriority: c.determinePriorityFromArtifacts(similarGroup),
				Reason:        "Similar content",
			}
			
			cluster.SimilarityScore = c.calculateGroupSimilarity(similarGroup)
			cluster.CohesionScore = c.calculateCohesion(similarGroup)
			
			if cluster.CohesionScore > 0.6 {
				clusters = append(clusters, cluster)
			}
		}
	}
	
	return clusters
}

// findTypeBasedClusters groups artifacts of the same type with related themes
func (c *ConsolidationEngine) findTypeBasedClusters(artifacts []*models.Artifact) []ArtifactCluster {
	var clusters []ArtifactCluster
	
	// Group by type
	typeGroups := make(map[string][]*models.Artifact)
	for _, artifact := range artifacts {
		typeGroups[artifact.Type] = append(typeGroups[artifact.Type], artifact)
	}
	
	// For each type, look for thematic groupings
	for artifactType, group := range typeGroups {
		if len(group) >= 2 {
			// Try to find thematic subgroups
			subGroups := c.findThematicSubgroups(group)
			for _, subGroup := range subGroups {
				if len(subGroup) >= 2 {
					theme := c.extractTheme(subGroup)
					cluster := ArtifactCluster{
						Artifacts:      subGroup,
						Theme:         theme,
						SuggestedTitle: fmt.Sprintf("%s: %s", strings.Title(artifactType), theme),
						SuggestedSchedule: c.determineScheduleFromArtifacts(subGroup),
						SuggestedPriority: c.determinePriorityFromArtifacts(subGroup),
						Reason:        fmt.Sprintf("Same type (%s) with related themes", artifactType),
					}
					
					cluster.SimilarityScore = c.calculateGroupSimilarity(subGroup)
					cluster.CohesionScore = c.calculateCohesion(subGroup)
					
					if cluster.CohesionScore > 0.4 {
						clusters = append(clusters, cluster)
					}
				}
			}
		}
	}
	
	return clusters
}

// calculateContentSimilarity computes similarity between two artifacts
func (c *ConsolidationEngine) calculateContentSimilarity(a1, a2 *models.Artifact) float64 {
	// Combine summary and content for comparison
	text1 := strings.ToLower(a1.Summary + " " + a1.Content)
	text2 := strings.ToLower(a2.Summary + " " + a2.Content)
	
	words1 := c.extractKeywords(text1)
	words2 := c.extractKeywords(text2)
	
	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}
	
	// Calculate Jaccard similarity
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)
	
	for _, word := range words1 {
		set1[word] = true
	}
	for _, word := range words2 {
		set2[word] = true
	}
	
	intersection := 0
	union := make(map[string]bool)
	
	for word := range set1 {
		union[word] = true
		if set2[word] {
			intersection++
		}
	}
	for word := range set2 {
		union[word] = true
	}
	
	if len(union) == 0 {
		return 0.0
	}
	
	return float64(intersection) / float64(len(union))
}

// extractKeywords extracts meaningful keywords from text
func (c *ConsolidationEngine) extractKeywords(text string) []string {
	words := strings.FieldsFunc(text, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	})
	
	// Filter out stop words and short words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true, "do": true,
		"does": true, "did": true, "will": true, "would": true, "could": true,
		"this": true, "that": true, "these": true, "those": true, "we": true,
		"implement": true, "add": true, "create": true, "update": true, "fix": true,
	}
	
	var keywords []string
	for _, word := range words {
		if len(word) >= 3 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}
	
	return keywords
}

// calculateTagSimilarity calculates similarity based on shared tags
func (c *ConsolidationEngine) calculateTagSimilarity(artifacts []*models.Artifact) float64 {
	if len(artifacts) < 2 {
		return 0.0
	}
	
	// Count tag frequencies
	tagCounts := make(map[string]int)
	for _, artifact := range artifacts {
		for _, tag := range artifact.TechnicalTags {
			tagCounts[tag]++
		}
	}
	
	// Calculate how many artifacts share tags
	sharedTags := 0
	for _, count := range tagCounts {
		if count > 1 {
			sharedTags++
		}
	}
	
	// Normalize by total number of unique tags
	if len(tagCounts) == 0 {
		return 0.0
	}
	
	return float64(sharedTags) / float64(len(tagCounts))
}

// calculateGroupSimilarity calculates overall similarity within a group
func (c *ConsolidationEngine) calculateGroupSimilarity(artifacts []*models.Artifact) float64 {
	if len(artifacts) < 2 {
		return 0.0
	}
	
	totalSimilarity := 0.0
	comparisons := 0
	
	for i := 0; i < len(artifacts); i++ {
		for j := i + 1; j < len(artifacts); j++ {
			similarity := c.calculateContentSimilarity(artifacts[i], artifacts[j])
			totalSimilarity += similarity
			comparisons++
		}
	}
	
	if comparisons == 0 {
		return 0.0
	}
	
	return totalSimilarity / float64(comparisons)
}

// calculateCohesion calculates how well the artifacts fit together
func (c *ConsolidationEngine) calculateCohesion(artifacts []*models.Artifact) float64 {
	if len(artifacts) < 2 {
		return 0.0
	}
	
	// Factors that contribute to cohesion:
	// 1. Shared tags
	// 2. Similar types
	// 3. Content similarity
	// 4. Temporal proximity
	
	tagSimilarity := c.calculateTagSimilarity(artifacts)
	contentSimilarity := c.calculateGroupSimilarity(artifacts)
	
	// Type consistency
	typeMap := make(map[string]int)
	for _, artifact := range artifacts {
		typeMap[artifact.Type]++
	}
	typeConsistency := 0.0
	if len(typeMap) == 1 {
		typeConsistency = 1.0
	} else if len(typeMap) == 2 {
		typeConsistency = 0.7
	} else {
		typeConsistency = 0.3
	}
	
	// Weighted combination
	cohesion := (tagSimilarity * 0.4) + (contentSimilarity * 0.4) + (typeConsistency * 0.2)
	
	return cohesion
}

// extractTheme attempts to identify a common theme from artifacts
func (c *ConsolidationEngine) extractTheme(artifacts []*models.Artifact) string {
	// Collect all keywords from summaries
	keywordCounts := make(map[string]int)
	
	for _, artifact := range artifacts {
		keywords := c.extractKeywords(strings.ToLower(artifact.Summary))
		for _, keyword := range keywords {
			keywordCounts[keyword]++
		}
	}
	
	// Find the most common keyword that appears in multiple artifacts
	maxCount := 0
	theme := "related work"
	
	for keyword, count := range keywordCounts {
		if count > 1 && count > maxCount {
			maxCount = count
			theme = keyword
		}
	}
	
	return theme
}

// findThematicSubgroups splits a group into thematic subgroups
func (c *ConsolidationEngine) findThematicSubgroups(artifacts []*models.Artifact) [][]*models.Artifact {
	// For now, return the whole group as one subgroup
	// In a more sophisticated implementation, we could use clustering algorithms
	return [][]*models.Artifact{artifacts}
}

// determineScheduleFromArtifacts suggests a schedule based on artifact characteristics
func (c *ConsolidationEngine) determineScheduleFromArtifacts(artifacts []*models.Artifact) string {
	// Look for urgency indicators in content
	urgencyKeywords := []string{"urgent", "critical", "immediate", "asap", "blocker", "blocking"}
	nearTermKeywords := []string{"soon", "next", "upcoming", "ready"}
	
	for _, artifact := range artifacts {
		text := strings.ToLower(artifact.Summary + " " + artifact.Content)
		
		for _, keyword := range urgencyKeywords {
			if strings.Contains(text, keyword) {
				return models.ScheduleNow
			}
		}
		
		for _, keyword := range nearTermKeywords {
			if strings.Contains(text, keyword) {
				return models.ScheduleNext
			}
		}
	}
	
	// Check artifact types for schedule hints
	hasPlans := false
	hasDecisions := false
	
	for _, artifact := range artifacts {
		switch artifact.Type {
		case models.TypeDecision:
			hasDecisions = true
		case models.TypePlan:
			hasPlans = true
		}
	}
	
	// Decisions often need immediate action
	if hasDecisions {
		return models.ScheduleNext
	}
	
	// Plans might be longer term
	if hasPlans {
		return models.ScheduleLater
	}
	
	return models.ScheduleNext // Default
}

// determinePriorityFromArtifacts suggests priority based on artifact content
func (c *ConsolidationEngine) determinePriorityFromArtifacts(artifacts []*models.Artifact) string {
	highPriorityKeywords := []string{"critical", "urgent", "important", "blocker", "security", "performance"}
	
	for _, artifact := range artifacts {
		text := strings.ToLower(artifact.Summary + " " + artifact.Content)
		
		for _, keyword := range highPriorityKeywords {
			if strings.Contains(text, keyword) {
				return models.WorkPriorityHigh
			}
		}
	}
	
	// Check for proposal approval status
	for _, artifact := range artifacts {
		if artifact.Type == models.TypeProposal && 
		   artifact.Metadata.ApprovalStatus == "approved" {
			return models.WorkPriorityMedium
		}
	}
	
	return models.WorkPriorityMedium // Default
}

// deduplicateClusters removes overlapping clusters
func (c *ConsolidationEngine) deduplicateClusters(clusters []ArtifactCluster) []ArtifactCluster {
	// Simple deduplication: remove clusters where >50% of artifacts overlap
	var deduplicated []ArtifactCluster
	
	for i, cluster1 := range clusters {
		isOverlap := false
		
		for j, cluster2 := range deduplicated {
			if j == i {
				continue
			}
			
			overlap := c.calculateClusterOverlap(cluster1, cluster2)
			if overlap > 0.5 {
				isOverlap = true
				// Keep the cluster with higher cohesion
				if cluster1.CohesionScore > cluster2.CohesionScore {
					deduplicated[j] = cluster1
				}
				break
			}
		}
		
		if !isOverlap {
			deduplicated = append(deduplicated, cluster1)
		}
	}
	
	return deduplicated
}

// calculateClusterOverlap calculates the overlap ratio between two clusters
func (c *ConsolidationEngine) calculateClusterOverlap(c1, c2 ArtifactCluster) float64 {
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)
	
	for _, artifact := range c1.Artifacts {
		set1[artifact.ID] = true
	}
	for _, artifact := range c2.Artifacts {
		set2[artifact.ID] = true
	}
	
	intersection := 0
	for id := range set1 {
		if set2[id] {
			intersection++
		}
	}
	
	minSize := len(c1.Artifacts)
	if len(c2.Artifacts) < minSize {
		minSize = len(c2.Artifacts)
	}
	
	if minSize == 0 {
		return 0.0
	}
	
	return float64(intersection) / float64(minSize)
}

// CreateWorkFromCluster creates a Work item from an artifact cluster
func (c *ConsolidationEngine) CreateWorkFromCluster(cluster ArtifactCluster) (*models.Work, error) {
	// Extract artifact IDs
	var artifactIDs []string
	var tags []string
	tagSet := make(map[string]bool)
	
	for _, artifact := range cluster.Artifacts {
		artifactIDs = append(artifactIDs, artifact.ID)
		for _, tag := range artifact.TechnicalTags {
			if !tagSet[tag] {
				tagSet[tag] = true
				tags = append(tags, tag)
			}
		}
	}
	
	// Create Work item
	work, err := c.client.CreateWork(
		cluster.SuggestedTitle,
		fmt.Sprintf("Work item consolidated from %d related artifacts", len(cluster.Artifacts)),
		cluster.SuggestedSchedule,
		cluster.SuggestedPriority,
		tags,
		artifactIDs,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create work: %w", err)
	}
	
	// Create associations
	for _, artifact := range cluster.Artifacts {
		if err := c.client.CreateAssociation(work.ID, artifact.ID); err != nil {
			fmt.Printf("Warning: failed to create association for %s: %v\n", artifact.ID, err)
		}
	}
	
	return work, nil
}

// CreateGroupFromCluster creates a Group from an artifact cluster (for later consolidation)
func (c *ConsolidationEngine) CreateGroupFromCluster(cluster ArtifactCluster) (*models.Group, error) {
	// Extract artifact IDs and tags
	var artifactIDs []string
	var tags []string
	tagSet := make(map[string]bool)
	
	for _, artifact := range cluster.Artifacts {
		artifactIDs = append(artifactIDs, artifact.ID)
		for _, tag := range artifact.TechnicalTags {
			if !tagSet[tag] {
				tagSet[tag] = true
				tags = append(tags, tag)
			}
		}
	}
	
	groupName := fmt.Sprintf("Group: %s", cluster.Theme)
	description := fmt.Sprintf("Auto-generated group of %d related artifacts. %s", len(cluster.Artifacts), cluster.Reason)
	
	group, err := c.client.CreateGroup(groupName, description, cluster.Theme, artifactIDs, tags)
	if err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}
	
	// Update cluster with group ID
	cluster.GroupID = group.ID
	
	return group, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: consolidate-hierarchy <command> [options]")
		fmt.Println("Commands:")
		fmt.Println("  analyze       - Find artifact clusters for Work creation")
		fmt.Println("  interactive   - Interactive consolidation mode")
		fmt.Println("  auto-group    - Automatically create groups from clusters")
		fmt.Println("  ready-groups  - Show groups ready for Work consolidation")
		fmt.Println("  consolidate <group-id> - Convert group to Work")
		os.Exit(1)
	}

	engine := NewConsolidationEngine()

	switch os.Args[1] {
	case "analyze":
		clusters, err := engine.FindArtifactClusters()
		if err != nil {
			log.Fatalf("Failed to find clusters: %v", err)
		}

		if len(clusters) == 0 {
			fmt.Println("✅ No consolidation opportunities found")
			fmt.Println("   All artifacts appear to be sufficiently distinct")
			return
		}

		fmt.Printf("🔍 Found %d potential Work items from artifact clusters:\n\n", len(clusters))
		for i, cluster := range clusters {
			fmt.Printf("%d. Theme: %s (Cohesion: %.1f%%)\n", i+1, cluster.Theme, cluster.CohesionScore*100)
			fmt.Printf("   Suggested Work: [%s/%s] %s\n", cluster.SuggestedSchedule, cluster.SuggestedPriority, cluster.SuggestedTitle)
			fmt.Printf("   Artifacts: %d items (%s)\n", len(cluster.Artifacts), cluster.Reason)
			
			fmt.Printf("   Details:\n")
			for _, artifact := range cluster.Artifacts {
				fmt.Printf("     - [%s] %s\n", artifact.Type, artifact.Summary)
			}
			fmt.Println()
		}

		fmt.Println("💡 Use 'consolidate-hierarchy interactive' to review and create Work items")

	case "interactive":
		clusters, err := engine.FindArtifactClusters()
		if err != nil {
			log.Fatalf("Failed to find clusters: %v", err)
		}

		if len(clusters) == 0 {
			fmt.Println("✅ No consolidation opportunities found")
			return
		}

		fmt.Printf("🔍 Found %d potential Work items\n\n", len(clusters))
		
		reader := bufio.NewReader(os.Stdin)
		
		for i, cluster := range clusters {
			fmt.Printf("Cluster %d/%d: %s (%.1f%% cohesion)\n", i+1, len(clusters), cluster.Theme, cluster.CohesionScore*100)
			fmt.Printf("Suggested Work: [%s/%s] %s\n", cluster.SuggestedSchedule, cluster.SuggestedPriority, cluster.SuggestedTitle)
			fmt.Printf("Reason: %s\n", cluster.Reason)
			fmt.Printf("Artifacts (%d):\n", len(cluster.Artifacts))
			
			for _, artifact := range cluster.Artifacts {
				fmt.Printf("  - [%s] %s\n", artifact.Type, artifact.Summary)
			}
			fmt.Println()
			
			fmt.Print("Action? [w=create Work, g=create Group, s=skip, q=quit]: ")
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			
			switch response {
			case "w", "work":
				work, err := engine.CreateWorkFromCluster(cluster)
				if err != nil {
					fmt.Printf("❌ Failed to create Work: %v\n", err)
				} else {
					fmt.Printf("✅ Created Work: %s (ID: %s)\n", work.Title, work.ID)
				}
			case "g", "group":
				group, err := engine.CreateGroupFromCluster(cluster)
				if err != nil {
					fmt.Printf("❌ Failed to create Group: %v\n", err)
				} else {
					fmt.Printf("✅ Created Group: %s (ID: %s)\n", group.Name, group.ID)
				}
			case "q", "quit":
				fmt.Println("👋 Goodbye!")
				return
			default:
				fmt.Println("⏭️  Skipped")
			}
			fmt.Println()
		}

	case "auto-group":
		clusters, err := engine.FindArtifactClusters()
		if err != nil {
			log.Fatalf("Failed to find clusters: %v", err)
		}

		created := 0
		for _, cluster := range clusters {
			if cluster.CohesionScore > 0.7 { // Only auto-create high-cohesion groups
				group, err := engine.CreateGroupFromCluster(cluster)
				if err != nil {
					fmt.Printf("❌ Failed to create group for %s: %v\n", cluster.Theme, err)
				} else {
					fmt.Printf("✅ Created group: %s\n", group.Name)
					created++
				}
			}
		}
		
		fmt.Printf("🎉 Created %d groups from high-cohesion clusters\n", created)

	case "ready-groups":
		readyGroups, err := engine.client.GetReadyGroups()
		if err != nil {
			log.Fatalf("Failed to get ready groups: %v", err)
		}

		if len(readyGroups) == 0 {
			fmt.Println("✅ No groups ready for Work consolidation")
			return
		}

		fmt.Printf("🎯 %d groups ready for Work consolidation:\n\n", len(readyGroups))
		for i, group := range readyGroups {
			title, _, schedule, priority := group.GenerateWorkSuggestion()
			fmt.Printf("%d. %s (Readiness: %.1f%%)\n", i+1, group.Name, group.Metadata.ReadinessScore*100)
			fmt.Printf("   Would create: [%s/%s] %s\n", schedule, priority, title)
			fmt.Printf("   Artifacts: %d items\n", group.Metadata.ArtifactCount)
			fmt.Printf("   ID: %s\n\n", group.ID)
		}

		fmt.Println("💡 Use 'consolidate-hierarchy consolidate <group-id>' to create Work from a group")

	case "consolidate":
		if len(os.Args) < 3 {
			fmt.Println("Usage: consolidate-hierarchy consolidate <group-id>")
			os.Exit(1)
		}

		groupID := os.Args[2]
		work, err := engine.client.ConsolidateGroupToWork(groupID, "manual")
		if err != nil {
			log.Fatalf("Failed to consolidate group: %v", err)
		}

		fmt.Printf("✅ Consolidated group into Work: %s\n", work.Title)
		fmt.Printf("   Schedule: %s\n", work.Schedule)
		fmt.Printf("   Priority: %s\n", work.Metadata.Priority)
		fmt.Printf("   Artifacts: %d\n", len(work.ArtifactRefs))
		fmt.Printf("   ID: %s\n", work.ID)

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}