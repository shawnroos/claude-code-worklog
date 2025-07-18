package data

import (
	"fmt"
	"sort"
	"time"

	"claude-work-tracker-ui/internal/models"
)

// LifecycleManager handles decay logic and cleanup suggestions for Work, Artifacts, and Groups
type LifecycleManager struct {
	markdownIO         *MarkdownIO
	associationManager *AssociationManager
	groupManager       *GroupManager
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager(markdownIO *MarkdownIO, associationMgr *AssociationManager, groupMgr *GroupManager) *LifecycleManager {
	return &LifecycleManager{
		markdownIO:         markdownIO,
		associationManager: associationMgr,
		groupManager:       groupMgr,
	}
}

// DecayAnalysis represents the decay status of items in the system
type DecayAnalysis struct {
	OrphanedArtifacts    []*models.Artifact  `json:"orphaned_artifacts"`
	StaleWork           []*models.Work       `json:"stale_work"`
	StaleArtifacts      []*models.Artifact   `json:"stale_artifacts"`
	StaleGroups         []*models.Group      `json:"stale_groups"`
	UnsupportedWork     []*models.Work       `json:"unsupported_work"`
	RecommendedActions  []CleanupAction      `json:"recommended_actions"`
	Summary             DecaySummary         `json:"summary"`
}

// CleanupAction represents a recommended cleanup action
type CleanupAction struct {
	Type        string      `json:"type"`         // archive, merge, review, consolidate
	Priority    string      `json:"priority"`     // high, medium, low
	ItemID      string      `json:"item_id"`
	ItemType    string      `json:"item_type"`    // work, artifact, group
	Reason      string      `json:"reason"`
	Details     string      `json:"details"`
	AutoSafe    bool        `json:"auto_safe"`    // Can be safely auto-executed
}

// DecaySummary provides overview statistics
type DecaySummary struct {
	TotalItems           int     `json:"total_items"`
	HealthyItems         int     `json:"healthy_items"`
	ItemsNeedingReview   int     `json:"items_needing_review"`
	ItemsNeedingAction   int     `json:"items_needing_action"`
	OrphanedCount        int     `json:"orphaned_count"`
	StaleCount           int     `json:"stale_count"`
	OverallHealthScore   float64 `json:"overall_health_score"`
}

// AnalyzeDecay performs comprehensive decay analysis across the entire system
func (lm *LifecycleManager) AnalyzeDecay() (*DecayAnalysis, error) {
	analysis := &DecayAnalysis{
		OrphanedArtifacts:   []*models.Artifact{},
		StaleWork:          []*models.Work{},
		StaleArtifacts:     []*models.Artifact{},
		StaleGroups:        []*models.Group{},
		UnsupportedWork:    []*models.Work{},
		RecommendedActions: []CleanupAction{},
	}

	// Get all items
	allWork, err := lm.markdownIO.ListAllWork()
	if err != nil {
		return nil, fmt.Errorf("failed to load work: %w", err)
	}

	allArtifacts, err := lm.markdownIO.ListAllArtifacts()
	if err != nil {
		return nil, fmt.Errorf("failed to load artifacts: %w", err)
	}

	allGroups, err := lm.groupManager.ListAllGroups()
	if err != nil {
		return nil, fmt.Errorf("failed to load groups: %w", err)
	}

	// Analyze Work items
	for _, work := range allWork {
		if work.ShouldDecay() {
			analysis.StaleWork = append(analysis.StaleWork, work)
			analysis.RecommendedActions = append(analysis.RecommendedActions, lm.createWorkDecayAction(work))
		}

		// Check for unsupported work (no artifacts)
		if len(work.ArtifactRefs) == 0 {
			analysis.UnsupportedWork = append(analysis.UnsupportedWork, work)
			analysis.RecommendedActions = append(analysis.RecommendedActions, lm.createUnsupportedWorkAction(work))
		}
	}

	// Analyze Artifacts
	for _, artifact := range allArtifacts {
		if artifact.IsOrphaned() {
			analysis.OrphanedArtifacts = append(analysis.OrphanedArtifacts, artifact)
			analysis.RecommendedActions = append(analysis.RecommendedActions, lm.createOrphanedArtifactAction(artifact))
		}

		if artifact.ShouldDecay() {
			analysis.StaleArtifacts = append(analysis.StaleArtifacts, artifact)
			analysis.RecommendedActions = append(analysis.RecommendedActions, lm.createArtifactDecayAction(artifact))
		}
	}

	// Analyze Groups
	for _, group := range allGroups {
		if lm.isGroupStale(group) {
			analysis.StaleGroups = append(analysis.StaleGroups, group)
			analysis.RecommendedActions = append(analysis.RecommendedActions, lm.createGroupDecayAction(group))
		}
	}

	// Sort actions by priority
	sort.Slice(analysis.RecommendedActions, func(i, j int) bool {
		priorityOrder := map[string]int{"high": 3, "medium": 2, "low": 1}
		return priorityOrder[analysis.RecommendedActions[i].Priority] > priorityOrder[analysis.RecommendedActions[j].Priority]
	})

	// Generate summary
	analysis.Summary = lm.generateDecaySummary(allWork, allArtifacts, allGroups, analysis)

	return analysis, nil
}

// createWorkDecayAction creates a cleanup action for stale work
func (lm *LifecycleManager) createWorkDecayAction(work *models.Work) CleanupAction {
	daysSinceActivity := 0.0
	if work.Metadata.LastActivityAt != nil {
		daysSinceActivity = time.Since(*work.Metadata.LastActivityAt).Hours() / 24
	}

	priority := "medium"
	autoSafe := false
	reason := fmt.Sprintf("No activity for %.0f days", daysSinceActivity)

	// Determine severity
	if work.Schedule == models.ScheduleNow && daysSinceActivity > 14 {
		priority = "high"
		reason = fmt.Sprintf("NOW item inactive for %.0f days", daysSinceActivity)
	} else if daysSinceActivity > 90 {
		priority = "high"
		autoSafe = true // Very old items can be safely archived
		reason = fmt.Sprintf("Very old item (%.0f days)", daysSinceActivity)
	}

	actionType := "review"
	if autoSafe {
		actionType = "archive"
	}

	return CleanupAction{
		Type:     actionType,
		Priority: priority,
		ItemID:   work.ID,
		ItemType: "work",
		Reason:   reason,
		Details:  fmt.Sprintf("Work '%s' in %s schedule. Progress: %d%%, Activity Score: %.1f", 
			work.Title, work.Schedule, work.Metadata.ProgressPercent, work.Metadata.ActivityScore),
		AutoSafe: autoSafe,
	}
}

// createUnsupportedWorkAction creates an action for work with no artifacts
func (lm *LifecycleManager) createUnsupportedWorkAction(work *models.Work) CleanupAction {
	daysSinceCreated := time.Since(work.CreatedAt).Hours() / 24
	
	priority := "medium"
	if daysSinceCreated > 7 {
		priority = "high"
	}

	return CleanupAction{
		Type:     "review",
		Priority: priority,
		ItemID:   work.ID,
		ItemType: "work",
		Reason:   fmt.Sprintf("No supporting artifacts for %.0f days", daysSinceCreated),
		Details:  fmt.Sprintf("Work '%s' has no artifacts. Consider adding documentation or archiving.", work.Title),
		AutoSafe: false,
	}
}

// createOrphanedArtifactAction creates an action for orphaned artifacts
func (lm *LifecycleManager) createOrphanedArtifactAction(artifact *models.Artifact) CleanupAction {
	daysSinceCreated := time.Since(artifact.CreatedAt).Hours() / 24
	
	priority := "low"
	autoSafe := false
	actionType := "consolidate"

	// Different handling based on type
	switch artifact.Type {
	case models.TypeDecision:
		if artifact.Metadata.EnforcementActive {
			priority = "medium"
			actionType = "review"
		} else {
			priority = "low"
			autoSafe = daysSinceCreated > 30
		}
	case models.TypeUpdate:
		// Updates become less relevant over time
		if daysSinceCreated > 30 {
			priority = "medium"
			autoSafe = true
			actionType = "archive"
		}
	case models.TypePlan, models.TypeProposal:
		if daysSinceCreated > 14 {
			priority = "medium"
		}
	}

	if autoSafe {
		actionType = "archive"
	}

	return CleanupAction{
		Type:     actionType,
		Priority: priority,
		ItemID:   artifact.ID,
		ItemType: "artifact",
		Reason:   fmt.Sprintf("Orphaned %s for %.0f days", artifact.Type, daysSinceCreated),
		Details:  fmt.Sprintf("'%s' has no work assignments or references", artifact.Summary),
		AutoSafe: autoSafe,
	}
}

// createArtifactDecayAction creates an action for stale artifacts
func (lm *LifecycleManager) createArtifactDecayAction(artifact *models.Artifact) CleanupAction {
	daysSinceActivity := 0.0
	if artifact.Metadata.LastActivityAt != nil {
		daysSinceActivity = time.Since(*artifact.Metadata.LastActivityAt).Hours() / 24
	}

	return CleanupAction{
		Type:     "archive",
		Priority: "low",
		ItemID:   artifact.ID,
		ItemType: "artifact",
		Reason:   fmt.Sprintf("Stale %s (%.0f days inactive)", artifact.Type, daysSinceActivity),
		Details:  fmt.Sprintf("'%s' activity score: %.1f", artifact.Summary, artifact.Metadata.ActivityScore),
		AutoSafe: daysSinceActivity > 90,
	}
}

// createGroupDecayAction creates an action for stale groups
func (lm *LifecycleManager) createGroupDecayAction(group *models.Group) CleanupAction {
	daysSinceModified := 0.0
	if group.Metadata.LastModified != nil {
		daysSinceModified = time.Since(*group.Metadata.LastModified).Hours() / 24
	}

	actionType := "review"
	if group.IsReadyForWork() {
		actionType = "consolidate"
	}

	return CleanupAction{
		Type:     actionType,
		Priority: "medium",
		ItemID:   group.ID,
		ItemType: "group",
		Reason:   fmt.Sprintf("Stale group (%.0f days)", daysSinceModified),
		Details:  fmt.Sprintf("Group '%s' with %d artifacts, readiness: %.1f%%", 
			group.Name, group.Metadata.ArtifactCount, group.Metadata.ReadinessScore*100),
		AutoSafe: false,
	}
}

// isGroupStale determines if a group should be considered stale
func (lm *LifecycleManager) isGroupStale(group *models.Group) bool {
	// Groups are stale if:
	// 1. No modification for 30+ days and not ready for work
	// 2. Very low activity score
	// 3. Empty or nearly empty

	if group.Metadata.ArtifactCount == 0 {
		return true
	}

	if group.Metadata.LastModified != nil {
		daysSinceModified := time.Since(*group.Metadata.LastModified).Hours() / 24
		if daysSinceModified > 30 && !group.IsReadyForWork() {
			return true
		}
	}

	if group.Metadata.ActivityScore < 1.0 {
		return true
	}

	return false
}

// generateDecaySummary creates a summary of the decay analysis
func (lm *LifecycleManager) generateDecaySummary(allWork []*models.Work, allArtifacts []*models.Artifact, allGroups []*models.Group, analysis *DecayAnalysis) DecaySummary {
	totalItems := len(allWork) + len(allArtifacts) + len(allGroups)
	staleCount := len(analysis.StaleWork) + len(analysis.StaleArtifacts) + len(analysis.StaleGroups)
	orphanedCount := len(analysis.OrphanedArtifacts)
	unsupportedCount := len(analysis.UnsupportedWork)

	itemsNeedingReview := 0
	itemsNeedingAction := 0
	for _, action := range analysis.RecommendedActions {
		if action.Type == "review" {
			itemsNeedingReview++
		} else {
			itemsNeedingAction++
		}
	}

	healthyItems := totalItems - staleCount - orphanedCount - unsupportedCount
	if healthyItems < 0 {
		healthyItems = 0
	}

	// Calculate overall health score (0-1)
	healthScore := 1.0
	if totalItems > 0 {
		problemItems := float64(staleCount + orphanedCount + unsupportedCount)
		healthScore = 1.0 - (problemItems / float64(totalItems))
		if healthScore < 0 {
			healthScore = 0
		}
	}

	return DecaySummary{
		TotalItems:         totalItems,
		HealthyItems:       healthyItems,
		ItemsNeedingReview: itemsNeedingReview,
		ItemsNeedingAction: itemsNeedingAction,
		OrphanedCount:      orphanedCount,
		StaleCount:         staleCount,
		OverallHealthScore: healthScore,
	}
}

// ExecuteCleanupAction performs a cleanup action
func (lm *LifecycleManager) ExecuteCleanupAction(action CleanupAction) error {
	switch action.ItemType {
	case "work":
		return lm.executeWorkCleanup(action)
	case "artifact":
		return lm.executeArtifactCleanup(action)
	case "group":
		return lm.executeGroupCleanup(action)
	default:
		return fmt.Errorf("unknown item type: %s", action.ItemType)
	}
}

// executeWorkCleanup performs cleanup action on a Work item
func (lm *LifecycleManager) executeWorkCleanup(action CleanupAction) error {
	allWork, err := lm.markdownIO.ListAllWork()
	if err != nil {
		return err
	}

	for _, work := range allWork {
		if work.ID == action.ItemID {
			switch action.Type {
			case "archive":
				return lm.archiveWork(work)
			case "review":
				// Mark for review but don't auto-action
				work.Metadata.DecayWarning = true
				return lm.markdownIO.WriteWork(work)
			default:
				return fmt.Errorf("unsupported work action: %s", action.Type)
			}
		}
	}

	return fmt.Errorf("work not found: %s", action.ItemID)
}

// executeArtifactCleanup performs cleanup action on an Artifact
func (lm *LifecycleManager) executeArtifactCleanup(action CleanupAction) error {
	allArtifacts, err := lm.markdownIO.ListAllArtifacts()
	if err != nil {
		return err
	}

	for _, artifact := range allArtifacts {
		if artifact.ID == action.ItemID {
			switch action.Type {
			case "archive":
				return lm.archiveArtifact(artifact)
			case "consolidate":
				// Mark as needing consolidation
				artifact.Metadata.DecayWarning = true
				return lm.markdownIO.WriteArtifact(artifact)
			default:
				return fmt.Errorf("unsupported artifact action: %s", action.Type)
			}
		}
	}

	return fmt.Errorf("artifact not found: %s", action.ItemID)
}

// executeGroupCleanup performs cleanup action on a Group
func (lm *LifecycleManager) executeGroupCleanup(action CleanupAction) error {
	group, err := lm.groupManager.GetGroupByID(action.ItemID)
	if err != nil {
		return err
	}

	switch action.Type {
	case "consolidate":
		// Auto-consolidate ready groups
		if group.IsReadyForWork() {
			_, err := lm.groupManager.ConsolidateGroupToWork(action.ItemID, "auto-decay")
			return err
		}
		return fmt.Errorf("group not ready for consolidation")
	case "archive":
		return lm.archiveGroup(group)
	default:
		return fmt.Errorf("unsupported group action: %s", action.Type)
	}
}

// archiveWork moves a Work item to archive
func (lm *LifecycleManager) archiveWork(work *models.Work) error {
	work.Metadata.Status = models.WorkStatusArchived
	now := time.Now()
	work.CompletedAt = &now
	work.UpdatedAt = now

	// Add archive notation to content
	archiveNote := fmt.Sprintf("\n\n---\n**ARCHIVED**: %s\n*Reason: Lifecycle management - stale item*\n", 
		now.Format("2006-01-02 15:04:05"))
	work.Content += archiveNote

	return lm.markdownIO.WriteWork(work)
}

// archiveArtifact moves an Artifact to archive
func (lm *LifecycleManager) archiveArtifact(artifact *models.Artifact) error {
	artifact.Metadata.Status = models.ArtifactStatusArchived
	artifact.UpdatedAt = time.Now()

	// Add archive notation to content
	archiveNote := fmt.Sprintf("\n\n---\n**ARCHIVED**: %s\n*Reason: Lifecycle management - orphaned/stale item*\n", 
		time.Now().Format("2006-01-02 15:04:05"))
	artifact.Content += archiveNote

	return lm.markdownIO.WriteArtifact(artifact)
}

// archiveGroup archives a Group
func (lm *LifecycleManager) archiveGroup(group *models.Group) error {
	group.Metadata.Status = models.GroupStatusArchived
	group.UpdatedAt = time.Now()

	return lm.groupManager.WriteGroup(group)
}

// RefreshAllActivityScores updates activity scores for all items to ensure accurate decay detection
func (lm *LifecycleManager) RefreshAllActivityScores() error {
	return lm.associationManager.RefreshAllActivityScores()
}

// GetHealthMetrics returns current system health metrics
func (lm *LifecycleManager) GetHealthMetrics() (*SystemHealthMetrics, error) {
	analysis, err := lm.AnalyzeDecay()
	if err != nil {
		return nil, err
	}

	// Calculate trend scores (would be enhanced with historical data)
	return &SystemHealthMetrics{
		OverallHealth:         analysis.Summary.OverallHealthScore,
		TotalItems:           analysis.Summary.TotalItems,
		HealthyItems:         analysis.Summary.HealthyItems,
		ProblematicItems:     analysis.Summary.StaleCount + analysis.Summary.OrphanedCount,
		PendingActions:       len(analysis.RecommendedActions),
		HighPriorityActions:  lm.countHighPriorityActions(analysis.RecommendedActions),
		AutoSafeActions:      lm.countAutoSafeActions(analysis.RecommendedActions),
		LastAnalyzed:         time.Now(),
		HealthTrend:          "stable", // Would be calculated from historical data
	}, nil
}

// SystemHealthMetrics provides system health overview
type SystemHealthMetrics struct {
	OverallHealth       float64   `json:"overall_health"`
	TotalItems          int       `json:"total_items"`
	HealthyItems        int       `json:"healthy_items"`
	ProblematicItems    int       `json:"problematic_items"`
	PendingActions      int       `json:"pending_actions"`
	HighPriorityActions int       `json:"high_priority_actions"`
	AutoSafeActions     int       `json:"auto_safe_actions"`
	LastAnalyzed        time.Time `json:"last_analyzed"`
	HealthTrend         string    `json:"health_trend"` // improving, stable, declining
}

// countHighPriorityActions counts actions with high priority
func (lm *LifecycleManager) countHighPriorityActions(actions []CleanupAction) int {
	count := 0
	for _, action := range actions {
		if action.Priority == "high" {
			count++
		}
	}
	return count
}

// countAutoSafeActions counts actions that can be safely auto-executed
func (lm *LifecycleManager) countAutoSafeActions(actions []CleanupAction) int {
	count := 0
	for _, action := range actions {
		if action.AutoSafe {
			count++
		}
	}
	return count
}

// AutoCleanup executes all auto-safe cleanup actions
func (lm *LifecycleManager) AutoCleanup() (*AutoCleanupResult, error) {
	analysis, err := lm.AnalyzeDecay()
	if err != nil {
		return nil, err
	}

	result := &AutoCleanupResult{
		ActionsExecuted: []CleanupAction{},
		ActionsFailed:   []CleanupFailure{},
		Summary:         "No auto-safe actions found",
	}

	executed := 0
	failed := 0

	for _, action := range analysis.RecommendedActions {
		if action.AutoSafe {
			if err := lm.ExecuteCleanupAction(action); err != nil {
				result.ActionsFailed = append(result.ActionsFailed, CleanupFailure{
					Action: action,
					Error:  err.Error(),
				})
				failed++
			} else {
				result.ActionsExecuted = append(result.ActionsExecuted, action)
				executed++
			}
		}
	}

	result.Summary = fmt.Sprintf("Executed %d auto-safe actions, %d failed", executed, failed)
	return result, nil
}

// AutoCleanupResult represents the result of auto cleanup
type AutoCleanupResult struct {
	ActionsExecuted []CleanupAction   `json:"actions_executed"`
	ActionsFailed   []CleanupFailure  `json:"actions_failed"`
	Summary         string           `json:"summary"`
}

// CleanupFailure represents a failed cleanup action
type CleanupFailure struct {
	Action CleanupAction `json:"action"`
	Error  string        `json:"error"`
}