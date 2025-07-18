package data

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"claude-work-tracker-ui/internal/models"
)

// UpdatesManager handles creation and management of updates documents
type UpdatesManager struct {
	baseDir string
}

// NewUpdatesManager creates a new updates manager
func NewUpdatesManager(baseDir string) *UpdatesManager {
	return &UpdatesManager{
		baseDir: baseDir,
	}
}

// CreateUpdate adds a new update to a Work item's updates document
func (um *UpdatesManager) CreateUpdate(workID string, update *models.Update) error {
	updatesPath := um.getUpdatesPath(workID)
	
	// Ensure updates directory exists
	if err := os.MkdirAll(filepath.Dir(updatesPath), 0755); err != nil {
		return fmt.Errorf("failed to create updates directory: %w", err)
	}
	
	// Read existing updates
	existingContent := ""
	if content, err := os.ReadFile(updatesPath); err == nil {
		existingContent = string(content)
	}
	
	// Generate new update content
	newUpdateContent := um.renderUpdate(update)
	
	// Prepend new update to existing content
	var finalContent string
	if existingContent == "" {
		// First update - create new file with frontmatter
		finalContent = fmt.Sprintf(`---
work_id: %s
---

%s`, workID, newUpdateContent)
	} else {
		// Prepend to existing content
		parts := strings.SplitN(existingContent, "---\n", 3)
		if len(parts) == 3 {
			// Has frontmatter
			frontmatter := fmt.Sprintf("---\n%s---\n", parts[1])
			existingUpdates := parts[2]
			finalContent = frontmatter + newUpdateContent + "\n---\n\n" + existingUpdates
		} else {
			// No frontmatter, add it
			finalContent = fmt.Sprintf(`---
work_id: %s
---

%s

---

%s`, workID, newUpdateContent, existingContent)
		}
	}
	
	// Write to file
	return os.WriteFile(updatesPath, []byte(finalContent), 0644)
}

// GetUpdates retrieves all updates for a Work item
func (um *UpdatesManager) GetUpdates(workID string) ([]*models.Update, error) {
	updatesPath := um.getUpdatesPath(workID)
	
	content, err := os.ReadFile(updatesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*models.Update{}, nil // No updates yet
		}
		return nil, fmt.Errorf("failed to read updates file: %w", err)
	}
	
	return um.parseUpdates(string(content), workID)
}

// GetUpdatesRef returns the relative path to the updates file for a Work item
func (um *UpdatesManager) GetUpdatesRef(workID string) string {
	return fmt.Sprintf("updates/%s.md", workID)
}

// getUpdatesPath returns the full path to the updates file
func (um *UpdatesManager) getUpdatesPath(workID string) string {
	return filepath.Join(um.baseDir, "updates", fmt.Sprintf("%s.md", workID))
}

// renderUpdate converts an Update to markdown format
func (um *UpdatesManager) renderUpdate(update *models.Update) string {
	var content strings.Builder
	
	// Header with timestamp and session
	content.WriteString(fmt.Sprintf("## Update %s", update.Timestamp.Format("2006-01-02 15:04")))
	if update.SessionID != "" {
		content.WriteString(fmt.Sprintf(" (Session: %s)", update.SessionID))
	}
	content.WriteString("\n")
	
	// Author and type
	content.WriteString(fmt.Sprintf("**Author**: %s", update.Author))
	if update.UpdateType != "" {
		content.WriteString(fmt.Sprintf(" | **Type**: %s", update.UpdateType))
	}
	content.WriteString("\n")
	
	// Progress change if available
	if update.ProgressBefore != update.ProgressAfter {
		content.WriteString(fmt.Sprintf("**Progress**: %d%% → %d%%\n", 
			update.ProgressBefore, update.ProgressAfter))
	}
	
	// Title if provided
	if update.Title != "" {
		content.WriteString(fmt.Sprintf("**Status**: %s\n", update.Title))
	}
	
	content.WriteString("\n")
	
	// Main summary
	content.WriteString(update.Summary)
	content.WriteString("\n")
	
	// Task changes
	if len(update.TasksCompleted) > 0 {
		content.WriteString("\n**Tasks Completed:**\n")
		for _, task := range update.TasksCompleted {
			content.WriteString(fmt.Sprintf("- ✅ %s\n", task))
		}
	}
	
	if len(update.TasksAdded) > 0 {
		content.WriteString("\n**Tasks Added:**\n")
		for _, task := range update.TasksAdded {
			content.WriteString(fmt.Sprintf("- ➕ %s\n", task))
		}
	}
	
	return content.String()
}

// parseUpdates extracts updates from markdown content
func (um *UpdatesManager) parseUpdates(content, workID string) ([]*models.Update, error) {
	var updates []*models.Update
	
	// Split content by update separators
	parts := strings.Split(content, "---\n")
	
	// Skip frontmatter (first part if it exists)
	startIndex := 0
	if strings.HasPrefix(content, "---\n") {
		startIndex = 2 // Skip frontmatter
	}
	
	for i := startIndex; i < len(parts); i += 2 {
		if i >= len(parts) {
			break
		}
		
		updateContent := strings.TrimSpace(parts[i])
		if updateContent == "" {
			continue
		}
		
		update := um.parseUpdate(updateContent, workID)
		if update != nil {
			updates = append(updates, update)
		}
	}
	
	return updates, nil
}

// parseUpdate parses a single update from markdown content
func (um *UpdatesManager) parseUpdate(content, workID string) *models.Update {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return nil
	}
	
	update := &models.Update{
		ID:     fmt.Sprintf("update-%d", time.Now().UnixNano()),
		WorkID: workID,
	}
	
	// Parse header for timestamp and session
	headerLine := lines[0]
	if strings.Contains(headerLine, "## Update") {
		// Extract timestamp
		if timestampMatch := strings.Index(headerLine, "Update "); timestampMatch >= 0 {
			timestampStr := headerLine[timestampMatch+7:]
			if sessionMatch := strings.Index(timestampStr, " (Session:"); sessionMatch >= 0 {
				timestampStr = timestampStr[:sessionMatch]
			}
			
			if timestamp, err := time.Parse("2006-01-02 15:04", timestampStr); err == nil {
				update.Timestamp = timestamp
			}
		}
		
		// Extract session ID
		if sessionStart := strings.Index(headerLine, "(Session: "); sessionStart >= 0 {
			sessionEnd := strings.Index(headerLine[sessionStart:], ")")
			if sessionEnd > 0 {
				update.SessionID = headerLine[sessionStart+10 : sessionStart+sessionEnd]
			}
		}
	}
	
	// Parse metadata and summary
	inSummary := false
	var summaryLines []string
	
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "**Author**:") {
			update.Author = strings.TrimSpace(strings.Split(line, "|")[0][11:])
			if typeIndex := strings.Index(line, "**Type**:"); typeIndex >= 0 {
				update.UpdateType = strings.TrimSpace(line[typeIndex+9:])
			}
		} else if strings.HasPrefix(line, "**Status**:") {
			update.Title = strings.TrimSpace(line[11:])
		} else if strings.HasPrefix(line, "**Progress**:") {
			// Parse progress change
			progressStr := strings.TrimSpace(line[13:])
			if arrowIndex := strings.Index(progressStr, " → "); arrowIndex >= 0 {
				beforeStr := strings.TrimSpace(progressStr[:arrowIndex])
				afterStr := strings.TrimSpace(progressStr[arrowIndex+3:])
				
				if before, err := parsePercentage(beforeStr); err == nil {
					update.ProgressBefore = before
				}
				if after, err := parsePercentage(afterStr); err == nil {
					update.ProgressAfter = after
				}
			}
		} else if strings.HasPrefix(line, "**Tasks Completed:**") {
			// Parse completed tasks
			continue
		} else if strings.HasPrefix(line, "**Tasks Added:**") {
			// Parse added tasks
			continue
		} else if strings.HasPrefix(line, "- ✅") {
			update.TasksCompleted = append(update.TasksCompleted, strings.TrimSpace(line[4:]))
		} else if strings.HasPrefix(line, "- ➕") {
			update.TasksAdded = append(update.TasksAdded, strings.TrimSpace(line[4:]))
		} else if line != "" && !strings.HasPrefix(line, "**") {
			inSummary = true
			summaryLines = append(summaryLines, line)
		} else if inSummary && line == "" {
			break // End of summary
		}
	}
	
	update.Summary = strings.Join(summaryLines, "\n")
	
	return update
}

// CreateAutomaticUpdate creates an update from Claude session completion
func (um *UpdatesManager) CreateAutomaticUpdate(workID, sessionID string, summary string, tasksCompleted []string, progressBefore, progressAfter int) error {
	update := &models.Update{
		ID:             fmt.Sprintf("update-%d", time.Now().UnixNano()),
		WorkID:         workID,
		Timestamp:      time.Now(),
		Title:          "Session Update",
		Summary:        summary,
		Author:         "Claude",
		SessionID:      sessionID,
		UpdateType:     "automatic",
		TasksCompleted: tasksCompleted,
		ProgressBefore: progressBefore,
		ProgressAfter:  progressAfter,
	}
	
	return um.CreateUpdate(workID, update)
}

// CreateManualUpdate creates a manual update
func (um *UpdatesManager) CreateManualUpdate(workID, title, summary, author string) error {
	update := &models.Update{
		ID:         fmt.Sprintf("update-%d", time.Now().UnixNano()),
		WorkID:     workID,
		Timestamp:  time.Now(),
		Title:      title,
		Summary:    summary,
		Author:     author,
		UpdateType: "manual",
	}
	
	return um.CreateUpdate(workID, update)
}

// parsePercentage extracts percentage value from string like "50%"
func parsePercentage(s string) (int, error) {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "%") {
		s = s[:len(s)-1]
	}
	
	var percentage int
	if _, err := fmt.Sscanf(s, "%d", &percentage); err != nil {
		return 0, err
	}
	
	return percentage, nil
}