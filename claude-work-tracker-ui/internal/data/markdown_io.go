package data

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"claude-work-tracker-ui/internal/models"
	"gopkg.in/yaml.v3"
)

// MarkdownIO handles reading and writing markdown work items
type MarkdownIO struct {
	baseDir string
}

// NewMarkdownIO creates a new markdown IO handler
func NewMarkdownIO(baseDir string) *MarkdownIO {
	return &MarkdownIO{
		baseDir: baseDir,
	}
}

// frontmatterRegex matches YAML frontmatter in markdown files
var frontmatterRegex = regexp.MustCompile(`(?s)^---\n(.*?)\n---\n(.*)$`)

// ReadMarkdownWorkItem reads a markdown work item from a file
func (m *MarkdownIO) ReadMarkdownWorkItem(filepath string) (*models.MarkdownWorkItem, error) {
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	matches := frontmatterRegex.FindSubmatch(content)
	if len(matches) < 3 {
		return nil, fmt.Errorf("invalid markdown format: no frontmatter found")
	}

	frontmatter := matches[1]
	markdownContent := strings.TrimSpace(string(matches[2]))

	// Parse YAML frontmatter
	var item models.MarkdownWorkItem
	if err := yaml.Unmarshal(frontmatter, &item); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Set content and file info
	item.Content = markdownContent
	item.Filepath = filepath
	item.Filename = filepath[strings.LastIndex(filepath, "/")+1:]

	return &item, nil
}

// WriteMarkdownWorkItem writes a markdown work item to a file
func (m *MarkdownIO) WriteMarkdownWorkItem(item *models.MarkdownWorkItem) error {
	// Generate filename if not set
	if item.Filename == "" {
		item.Filename = m.generateFilename(item)
	}

	// Determine directory based on schedule
	dir := m.getDirectoryForSchedule(item.Schedule, item.Type)
	fullPath := filepath.Join(dir, item.Filename)

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate markdown content
	content, err := m.generateMarkdownContent(item)
	if err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	// Write file
	if err := ioutil.WriteFile(fullPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	item.Filepath = fullPath
	return nil
}

// generateMarkdownContent creates the full markdown file content with frontmatter
func (m *MarkdownIO) generateMarkdownContent(item *models.MarkdownWorkItem) ([]byte, error) {
	var buf bytes.Buffer

	// Write frontmatter
	buf.WriteString("---\n")
	
	// Use a custom YAML encoder to control formatting
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	
	// Create a map to control field order
	frontmatter := map[string]interface{}{
		"id":             item.ID,
		"type":           item.Type,
		"summary":        item.Summary,
		"schedule":       item.Schedule,
		"technical_tags": item.TechnicalTags,
		"session_number": item.SessionNumber,
		"created_at":     item.CreatedAt,
		"updated_at":     item.UpdatedAt,
		"git_context":    item.GitContext,
		"metadata":       item.Metadata,
	}
	
	if err := encoder.Encode(frontmatter); err != nil {
		return nil, fmt.Errorf("failed to encode frontmatter: %w", err)
	}
	
	buf.WriteString("---\n\n")
	
	// Write content
	buf.WriteString(item.Content)
	if !strings.HasSuffix(item.Content, "\n") {
		buf.WriteString("\n")
	}

	return buf.Bytes(), nil
}

// generateFilename creates a filename following the pattern: {type}-{brief-description}-{date}-{short-id}.md
func (m *MarkdownIO) generateFilename(item *models.MarkdownWorkItem) string {
	// Extract brief description from summary (first few words, sanitized)
	summary := strings.ToLower(item.Summary)
	words := strings.Fields(summary)
	if len(words) > 4 {
		words = words[:4]
	}
	description := strings.Join(words, "-")
	
	// Sanitize description for filename
	description = regexp.MustCompile(`[^a-z0-9-]+`).ReplaceAllString(description, "-")
	description = strings.Trim(description, "-")
	
	// Get date
	date := item.CreatedAt.Format("2006-01-02")
	
	// Get short ID (last 6 chars or generate)
	shortID := item.ID
	if len(shortID) > 6 {
		shortID = shortID[len(shortID)-6:]
	}
	
	return fmt.Sprintf("%s-%s-%s-%s.md", item.Type, description, date, shortID)
}

// getDirectoryForSchedule returns the appropriate directory for a work item
func (m *MarkdownIO) getDirectoryForSchedule(schedule, itemType string) string {
	if itemType == models.TypeDecision {
		// Decisions have their own directory structure
		return filepath.Join(m.baseDir, "decisions", "active")
	}
	
	switch schedule {
	case models.ScheduleNow:
		return filepath.Join(m.baseDir, "items", "now")
	case models.ScheduleNext:
		return filepath.Join(m.baseDir, "items", "next")
	case models.ScheduleLater:
		return filepath.Join(m.baseDir, "items", "later")
	default:
		return filepath.Join(m.baseDir, "items", "unscheduled")
	}
}

// MoveToCompleted moves a work item to the completed directory
func (m *MarkdownIO) MoveToCompleted(item *models.MarkdownWorkItem) error {
	if item.Filepath == "" {
		return fmt.Errorf("item has no filepath")
	}

	// Determine completed directory with year-month subdirectory
	yearMonth := time.Now().Format("2006-01")
	completedDir := filepath.Join(m.baseDir, "items", "completed", yearMonth)
	
	// Ensure directory exists
	if err := os.MkdirAll(completedDir, 0755); err != nil {
		return fmt.Errorf("failed to create completed directory: %w", err)
	}

	// Move file
	newPath := filepath.Join(completedDir, item.Filename)
	if err := os.Rename(item.Filepath, newPath); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	item.Filepath = newPath
	return nil
}

// UpdateSchedule changes the schedule of a work item and moves it to appropriate directory
func (m *MarkdownIO) UpdateSchedule(item *models.MarkdownWorkItem, newSchedule string) error {
	oldPath := item.Filepath
	item.Schedule = newSchedule
	item.UpdatedAt = time.Now()

	// Write to new location
	if err := m.WriteMarkdownWorkItem(item); err != nil {
		return fmt.Errorf("failed to write to new location: %w", err)
	}

	// Remove old file if path changed
	if oldPath != "" && oldPath != item.Filepath {
		os.Remove(oldPath)
	}

	return nil
}

// ListWorkItems returns all work items from a specific schedule directory
func (m *MarkdownIO) ListWorkItems(schedule string) ([]*models.MarkdownWorkItem, error) {
	dir := filepath.Join(m.baseDir, "items", schedule)
	return m.listWorkItemsFromDir(dir)
}

// ListAllWorkItems returns all work items from all directories
func (m *MarkdownIO) ListAllWorkItems() ([]*models.MarkdownWorkItem, error) {
	var allItems []*models.MarkdownWorkItem

	// List from all schedule directories
	schedules := []string{"now", "next", "later"}
	for _, schedule := range schedules {
		items, err := m.ListWorkItems(schedule)
		if err != nil {
			continue // Directory might not exist yet
		}
		allItems = append(allItems, items...)
	}

	// List active decisions
	decisionsDir := filepath.Join(m.baseDir, "decisions", "active")
	decisions, err := m.listWorkItemsFromDir(decisionsDir)
	if err == nil {
		allItems = append(allItems, decisions...)
	}

	return allItems, nil
}

// listWorkItemsFromDir reads all markdown files from a directory
func (m *MarkdownIO) listWorkItemsFromDir(dir string) ([]*models.MarkdownWorkItem, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*models.MarkdownWorkItem{}, nil
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var items []*models.MarkdownWorkItem
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		filepath := filepath.Join(dir, file.Name())
		item, err := m.ReadMarkdownWorkItem(filepath)
		if err != nil {
			continue // Skip files that can't be parsed
		}

		items = append(items, item)
	}

	return items, nil
}

// SearchWorkItems searches for work items containing the query
func (m *MarkdownIO) SearchWorkItems(query string) ([]*models.MarkdownWorkItem, error) {
	allItems, err := m.ListAllWorkItems()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []*models.MarkdownWorkItem
	
	for _, item := range allItems {
		// Search in summary, content, and tags
		if strings.Contains(strings.ToLower(item.Summary), query) ||
			strings.Contains(strings.ToLower(item.Content), query) ||
			containsTag(item.TechnicalTags, query) {
			results = append(results, item)
		}
	}

	return results, nil
}

// containsTag checks if a tag list contains a query string
func containsTag(tags []string, query string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}