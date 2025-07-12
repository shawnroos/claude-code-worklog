package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"claude-work-tracker-ui/internal/data"
	"claude-work-tracker-ui/internal/models"
)

// ConsolidationCandidate represents a potential merge opportunity
type ConsolidationCandidate struct {
	Item1          *models.MarkdownWorkItem
	Item2          *models.MarkdownWorkItem
	SimilarityScore float64
	Reason         string
	MergeStrategy  string
}

// ConsolidationEngine handles finding and merging similar work items
type ConsolidationEngine struct {
	markdownIO *data.MarkdownIO
	workDir    string
}

// NewConsolidationEngine creates a new consolidation engine
func NewConsolidationEngine(workDir string) *ConsolidationEngine {
	return &ConsolidationEngine{
		markdownIO: data.NewMarkdownIO(workDir),
		workDir:    workDir,
	}
}

// FindCandidates identifies work items that could be consolidated
func (c *ConsolidationEngine) FindCandidates() ([]ConsolidationCandidate, error) {
	allItems, err := c.markdownIO.ListAllWorkItems()
	if err != nil {
		return nil, fmt.Errorf("failed to load work items: %w", err)
	}

	var candidates []ConsolidationCandidate

	// Compare each item with every other item
	for i := 0; i < len(allItems); i++ {
		for j := i + 1; j < len(allItems); j++ {
			item1 := allItems[i]
			item2 := allItems[j]

			// Skip if different types
			if item1.Type != item2.Type {
				continue
			}

			// Skip if one is completed
			if item1.Metadata.Status == models.StatusCompleted || 
			   item2.Metadata.Status == models.StatusCompleted {
				continue
			}

			candidate := c.analyzeSimilarity(item1, item2)
			if candidate.SimilarityScore > 0.6 { // 60% threshold
				candidates = append(candidates, candidate)
			}
		}
	}

	// Sort by similarity score (highest first)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].SimilarityScore > candidates[j].SimilarityScore
	})

	return candidates, nil
}

// analyzeSimilarity computes similarity between two work items
func (c *ConsolidationEngine) analyzeSimilarity(item1, item2 *models.MarkdownWorkItem) ConsolidationCandidate {
	var totalScore float64
	var reasons []string

	// 1. Summary similarity (40% weight)
	summaryScore := c.computeTextSimilarity(item1.Summary, item2.Summary)
	totalScore += summaryScore * 0.4
	if summaryScore > 0.7 {
		reasons = append(reasons, fmt.Sprintf("Similar summaries (%.1f%%)", summaryScore*100))
	}

	// 2. Technical tags overlap (25% weight)
	tagScore := c.computeTagSimilarity(item1.TechnicalTags, item2.TechnicalTags)
	totalScore += tagScore * 0.25
	if tagScore > 0.5 {
		reasons = append(reasons, fmt.Sprintf("Overlapping tags (%.1f%%)", tagScore*100))
	}

	// 3. Content similarity (25% weight)
	contentScore := c.computeTextSimilarity(item1.Content, item2.Content)
	totalScore += contentScore * 0.25
	if contentScore > 0.5 {
		reasons = append(reasons, fmt.Sprintf("Similar content (%.1f%%)", contentScore*100))
	}

	// 4. Schedule compatibility (10% weight)
	scheduleScore := c.computeScheduleCompatibility(item1.Schedule, item2.Schedule)
	totalScore += scheduleScore * 0.1
	if scheduleScore > 0.5 {
		reasons = append(reasons, "Compatible schedules")
	}

	// Determine merge strategy
	mergeStrategy := c.determineMergeStrategy(item1, item2, totalScore)

	return ConsolidationCandidate{
		Item1:          item1,
		Item2:          item2,
		SimilarityScore: totalScore,
		Reason:         strings.Join(reasons, ", "),
		MergeStrategy:  mergeStrategy,
	}
}

// computeTextSimilarity calculates similarity between two text strings
func (c *ConsolidationEngine) computeTextSimilarity(text1, text2 string) float64 {
	if text1 == "" || text2 == "" {
		return 0.0
	}

	// Normalize texts
	words1 := c.normalizeText(text1)
	words2 := c.normalizeText(text2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	// Convert to sets for intersection/union
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, word := range words1 {
		set1[word] = true
	}
	for _, word := range words2 {
		set2[word] = true
	}

	// Calculate Jaccard similarity (intersection / union)
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

// normalizeText extracts meaningful words from text
func (c *ConsolidationEngine) normalizeText(text string) []string {
	// Convert to lowercase and split on non-alphanumeric characters
	text = strings.ToLower(text)
	words := strings.FieldsFunc(text, func(c rune) bool {
		return !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9'))
	})

	// Filter out common stop words and short words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true, "do": true,
		"does": true, "did": true, "will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "must": true, "can": true, "this": true, "that": true,
		"these": true, "those": true, "we": true, "us": true, "our": true, "you": true,
		"your": true, "i": true, "me": true, "my": true, "it": true, "its": true,
	}

	var filtered []string
	for _, word := range words {
		if len(word) >= 3 && !stopWords[word] {
			filtered = append(filtered, word)
		}
	}

	return filtered
}

// computeTagSimilarity calculates overlap between tag arrays
func (c *ConsolidationEngine) computeTagSimilarity(tags1, tags2 []string) float64 {
	if len(tags1) == 0 && len(tags2) == 0 {
		return 1.0 // Both empty, perfectly similar
	}
	if len(tags1) == 0 || len(tags2) == 0 {
		return 0.0 // One empty, no similarity
	}

	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, tag := range tags1 {
		set1[strings.ToLower(tag)] = true
	}
	for _, tag := range tags2 {
		set2[strings.ToLower(tag)] = true
	}

	intersection := 0
	union := make(map[string]bool)

	for tag := range set1 {
		union[tag] = true
		if set2[tag] {
			intersection++
		}
	}
	for tag := range set2 {
		union[tag] = true
	}

	return float64(intersection) / float64(len(union))
}

// computeScheduleCompatibility checks if schedules are compatible for merging
func (c *ConsolidationEngine) computeScheduleCompatibility(schedule1, schedule2 string) float64 {
	if schedule1 == schedule2 {
		return 1.0
	}

	// Define compatibility matrix
	compatibility := map[string]map[string]float64{
		models.ScheduleNow: {
			models.ScheduleNext:  0.7, // NOW and NEXT are somewhat compatible
			models.ScheduleLater: 0.3, // NOW and LATER less so
		},
		models.ScheduleNext: {
			models.ScheduleNow:   0.7,
			models.ScheduleLater: 0.8, // NEXT and LATER are quite compatible
		},
		models.ScheduleLater: {
			models.ScheduleNow:  0.3,
			models.ScheduleNext: 0.8,
		},
	}

	if comp, exists := compatibility[schedule1][schedule2]; exists {
		return comp
	}
	return 0.0
}

// determineMergeStrategy decides how to merge two items
func (c *ConsolidationEngine) determineMergeStrategy(item1, item2 *models.MarkdownWorkItem, score float64) string {
	if score > 0.9 {
		return "merge_content" // Very similar, merge all content
	} else if score > 0.8 {
		return "combine_detailed" // Similar, combine with detailed merge
	} else if score > 0.7 {
		return "combine_summary" // Somewhat similar, combine with summary
	} else {
		return "reference_only" // Just add references to each other
	}
}

// PerformConsolidation executes the consolidation for a candidate
func (c *ConsolidationEngine) PerformConsolidation(candidate ConsolidationCandidate, userApproved bool) error {
	if !userApproved {
		return fmt.Errorf("consolidation not approved by user")
	}

	switch candidate.MergeStrategy {
	case "merge_content":
		return c.mergeContent(candidate)
	case "combine_detailed":
		return c.combineDetailed(candidate)
	case "combine_summary":
		return c.combineSummary(candidate)
	case "reference_only":
		return c.addReferences(candidate)
	default:
		return fmt.Errorf("unknown merge strategy: %s", candidate.MergeStrategy)
	}
}

// mergeContent combines two very similar items into one
func (c *ConsolidationEngine) mergeContent(candidate ConsolidationCandidate) error {
	primary := candidate.Item1
	secondary := candidate.Item2

	// Combine summaries
	if len(secondary.Summary) > len(primary.Summary) {
		primary.Summary = secondary.Summary
	}

	// Merge technical tags
	tagSet := make(map[string]bool)
	for _, tag := range primary.TechnicalTags {
		tagSet[tag] = true
	}
	for _, tag := range secondary.TechnicalTags {
		tagSet[tag] = true
	}
	
	var mergedTags []string
	for tag := range tagSet {
		mergedTags = append(mergedTags, tag)
	}
	sort.Strings(mergedTags)
	primary.TechnicalTags = mergedTags

	// Combine content
	combinedContent := fmt.Sprintf(`%s

## Merged Content

The following content was merged from a similar work item:

%s

---
*Merged from item: %s*
*Original item archived with references preserved*`,
		primary.Content,
		secondary.Content,
		secondary.ID)

	primary.Content = combinedContent

	// Choose the more urgent schedule
	if c.getSchedulePriority(secondary.Schedule) < c.getSchedulePriority(primary.Schedule) {
		primary.Schedule = secondary.Schedule
	}

	// Add reference to merged item
	if primary.Metadata.RelatedItems == nil {
		primary.Metadata.RelatedItems = []string{}
	}
	primary.Metadata.RelatedItems = append(primary.Metadata.RelatedItems, secondary.ID)

	// Write updated primary item
	if err := c.markdownIO.WriteMarkdownWorkItem(primary); err != nil {
		return fmt.Errorf("failed to write merged item: %w", err)
	}

	// Archive secondary item
	if err := c.archiveMergedItem(secondary, primary.ID); err != nil {
		return fmt.Errorf("failed to archive secondary item: %w", err)
	}

	return nil
}

// combineDetailed creates a comprehensive combined item
func (c *ConsolidationEngine) combineDetailed(candidate ConsolidationCandidate) error {
	// Similar to merge_content but preserves more structure
	return c.mergeContent(candidate) // For now, use same implementation
}

// combineSummary creates a lighter combination
func (c *ConsolidationEngine) combineSummary(candidate ConsolidationCandidate) error {
	primary := candidate.Item1
	secondary := candidate.Item2

	// Add reference in content
	referenceText := fmt.Sprintf(`

## Related Work

This item is related to: %s
- Summary: %s
- Schedule: %s
- ID: %s`,
		secondary.Summary,
		secondary.Summary,
		secondary.Schedule,
		secondary.ID)

	primary.Content += referenceText

	// Add to related items
	if primary.Metadata.RelatedItems == nil {
		primary.Metadata.RelatedItems = []string{}
	}
	primary.Metadata.RelatedItems = append(primary.Metadata.RelatedItems, secondary.ID)

	// Write updated primary item
	if err := c.markdownIO.WriteMarkdownWorkItem(primary); err != nil {
		return fmt.Errorf("failed to write updated item: %w", err)
	}

	// Update secondary item with reference back
	referenceBackText := fmt.Sprintf(`

## Related Work

This item is related to: %s (ID: %s)`,
		primary.Summary,
		primary.ID)

	secondary.Content += referenceBackText
	if secondary.Metadata.RelatedItems == nil {
		secondary.Metadata.RelatedItems = []string{}
	}
	secondary.Metadata.RelatedItems = append(secondary.Metadata.RelatedItems, primary.ID)

	return c.markdownIO.WriteMarkdownWorkItem(secondary)
}

// addReferences just adds cross-references without merging
func (c *ConsolidationEngine) addReferences(candidate ConsolidationCandidate) error {
	return c.combineSummary(candidate) // Use same logic as summary combination
}

// archiveMergedItem moves a merged item to archive with merge notation
func (c *ConsolidationEngine) archiveMergedItem(item *models.MarkdownWorkItem, mergedIntoID string) error {
	// Update status and add merge notation
	item.Metadata.Status = models.StatusArchived
	item.Content = fmt.Sprintf(`# MERGED ITEM - ARCHIVED

This item was merged into: %s

Original content preserved below:

---

%s`, mergedIntoID, item.Content)

	// Move to archive
	archiveDir := filepath.Join(c.workDir, "items", "merged")
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return err
	}

	// Generate new filename for archive
	newFilename := fmt.Sprintf("merged-%s", item.Filename)
	newPath := filepath.Join(archiveDir, newFilename)
	item.Filepath = newPath
	item.Filename = newFilename

	return c.markdownIO.WriteMarkdownWorkItem(item)
}

// getSchedulePriority returns numeric priority for schedule comparison
func (c *ConsolidationEngine) getSchedulePriority(schedule string) int {
	switch schedule {
	case models.ScheduleNow:
		return 1
	case models.ScheduleNext:
		return 2
	case models.ScheduleLater:
		return 3
	default:
		return 4
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: consolidate <command> [options]")
		fmt.Println("Commands:")
		fmt.Println("  analyze   - Find consolidation candidates")
		fmt.Println("  merge <id1> <id2> - Merge two specific items")
		fmt.Println("  interactive - Interactive consolidation mode")
		os.Exit(1)
	}

	workDir := ".claude-work"
	if wd := os.Getenv("WORK_DIR"); wd != "" {
		workDir = wd
	}

	engine := NewConsolidationEngine(workDir)

	switch os.Args[1] {
	case "analyze":
		candidates, err := engine.FindCandidates()
		if err != nil {
			log.Fatalf("Failed to find candidates: %v", err)
		}

		if len(candidates) == 0 {
			fmt.Println("✅ No consolidation candidates found")
			fmt.Println("   All work items appear to be sufficiently distinct")
			return
		}

		fmt.Printf("🔍 Found %d consolidation candidates:\n\n", len(candidates))
		for i, candidate := range candidates {
			fmt.Printf("%d. Similarity: %.1f%% (%s)\n", i+1, candidate.SimilarityScore*100, candidate.Reason)
			fmt.Printf("   Item 1: [%s/%s] %s\n", candidate.Item1.Type, candidate.Item1.Schedule, candidate.Item1.Summary)
			fmt.Printf("   Item 2: [%s/%s] %s\n", candidate.Item2.Type, candidate.Item2.Schedule, candidate.Item2.Summary)
			fmt.Printf("   Strategy: %s\n\n", candidate.MergeStrategy)
		}

		fmt.Println("💡 Use 'consolidate interactive' to review and merge items")

	case "interactive":
		candidates, err := engine.FindCandidates()
		if err != nil {
			log.Fatalf("Failed to find candidates: %v", err)
		}

		if len(candidates) == 0 {
			fmt.Println("✅ No consolidation candidates found")
			return
		}

		fmt.Printf("🔍 Found %d consolidation candidates\n\n", len(candidates))
		
		for i, candidate := range candidates {
			fmt.Printf("Candidate %d/%d: %.1f%% similarity\n", i+1, len(candidates), candidate.SimilarityScore*100)
			fmt.Printf("Reason: %s\n", candidate.Reason)
			fmt.Printf("Strategy: %s\n\n", candidate.MergeStrategy)
			
			fmt.Printf("Item 1: [%s/%s] %s\n", candidate.Item1.Type, candidate.Item1.Schedule, candidate.Item1.Summary)
			fmt.Printf("Item 2: [%s/%s] %s\n\n", candidate.Item2.Type, candidate.Item2.Schedule, candidate.Item2.Summary)
			
			fmt.Print("Consolidate these items? [y/N/s(kip all)]: ")
			var response string
			fmt.Scanln(&response)
			
			switch strings.ToLower(response) {
			case "y", "yes":
				if err := engine.PerformConsolidation(candidate, true); err != nil {
					fmt.Printf("❌ Failed to consolidate: %v\n", err)
				} else {
					fmt.Printf("✅ Consolidated successfully\n")
				}
			case "s", "skip":
				fmt.Println("⏭️  Skipping remaining candidates")
				return
			default:
				fmt.Println("⏭️  Skipped")
			}
			fmt.Println()
		}

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}