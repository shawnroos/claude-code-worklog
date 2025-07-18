package data

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"claude-work-tracker-ui/internal/models"
	"gopkg.in/yaml.v3"
)

// MarkdownIO handles reading and writing markdown work items, artifacts, and work containers
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

// === Work Container Methods ===

// ReadWork reads a Work container from a markdown file
func (m *MarkdownIO) ReadWork(filepath string) (*models.Work, error) {
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
	var work models.Work
	if err := yaml.Unmarshal(frontmatter, &work); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Set content and file info
	work.Content = markdownContent
	work.Filepath = filepath
	work.Filename = filepath[strings.LastIndex(filepath, "/")+1:]

	return &work, nil
}

// WriteWork writes a Work container to a file
func (m *MarkdownIO) WriteWork(work *models.Work) error {
	// Generate filename if not set
	if work.Filename == "" {
		work.Filename = m.generateWorkFilename(work)
	}

	// Determine directory based on schedule
	dir := m.getWorkDirectory(work.Schedule)
	fullPath := filepath.Join(dir, work.Filename)

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate markdown content
	content, err := m.generateWorkContent(work)
	if err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	// Write file
	if err := ioutil.WriteFile(fullPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	work.Filepath = fullPath
	return nil
}

// generateWorkContent creates the full markdown file content for Work with frontmatter
func (m *MarkdownIO) generateWorkContent(work *models.Work) ([]byte, error) {
	var buf bytes.Buffer

	// Write frontmatter
	buf.WriteString("---\n")
	
	// Use a custom YAML encoder to control formatting
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	
	// Create a map to control field order for Work
	frontmatter := map[string]interface{}{
		"id":             work.ID,
		"title":          work.Title,
		"description":    work.Description,
		"schedule":       work.Schedule,
		"created_at":     work.CreatedAt,
		"updated_at":     work.UpdatedAt,
		"git_context":    work.GitContext,
		"session_number": work.SessionNumber,
		"technical_tags": work.TechnicalTags,
		"artifact_refs":  work.ArtifactRefs,
		"metadata":       work.Metadata,
	}
	
	// Handle optional pointer fields
	if work.StartedAt != nil {
		frontmatter["started_at"] = *work.StartedAt
	}
	if work.CompletedAt != nil {
		frontmatter["completed_at"] = *work.CompletedAt
	}
	if work.GroupID != "" {
		frontmatter["group_id"] = work.GroupID
	}
	if work.OverviewUpdated != nil {
		frontmatter["overview_updated"] = *work.OverviewUpdated
	}
	if work.UpdatesRef != "" {
		frontmatter["updates_ref"] = work.UpdatesRef
	}
	
	if err := encoder.Encode(frontmatter); err != nil {
		return nil, fmt.Errorf("failed to encode frontmatter: %w", err)
	}
	
	buf.WriteString("---\n\n")
	
	// Write content
	buf.WriteString(work.Content)
	if !strings.HasSuffix(work.Content, "\n") {
		buf.WriteString("\n")
	}

	return buf.Bytes(), nil
}

// generateWorkFilename creates a filename for Work items: work-{brief-description}-{date}-{short-id}.md
func (m *MarkdownIO) generateWorkFilename(work *models.Work) string {
	// Extract brief description from title (first few words, sanitized)
	title := strings.ToLower(work.Title)
	words := strings.Fields(title)
	if len(words) > 4 {
		words = words[:4]
	}
	description := strings.Join(words, "-")
	
	// Sanitize description for filename
	description = regexp.MustCompile(`[^a-z0-9-]+`).ReplaceAllString(description, "-")
	description = strings.Trim(description, "-")
	
	// Get date
	date := work.CreatedAt.Format("2006-01-02")
	
	// Get short ID (last 6 chars or generate)
	shortID := work.ID
	if len(shortID) > 6 {
		shortID = shortID[len(shortID)-6:]
	}
	
	return fmt.Sprintf("work-%s-%s-%s.md", description, date, shortID)
}

// getWorkDirectory returns the appropriate directory for a Work item
func (m *MarkdownIO) getWorkDirectory(schedule string) string {
	switch schedule {
	case models.ScheduleNow:
		return filepath.Join(m.baseDir, "now")
	case models.ScheduleNext:
		return filepath.Join(m.baseDir, "next")
	case models.ScheduleLater:
		return filepath.Join(m.baseDir, "later")
	case models.ScheduleClosed:
		return filepath.Join(m.baseDir, "closed")
	default:
		return filepath.Join(m.baseDir, "work", "unscheduled")
	}
}

// === Artifact Methods ===

// ReadArtifact reads an Artifact from a markdown file
func (m *MarkdownIO) ReadArtifact(filepath string) (*models.Artifact, error) {
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
	var artifact models.Artifact
	if err := yaml.Unmarshal(frontmatter, &artifact); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Set content and file info
	artifact.Content = markdownContent
	artifact.Filepath = filepath
	artifact.Filename = filepath[strings.LastIndex(filepath, "/")+1:]

	return &artifact, nil
}

// WriteArtifact writes an Artifact to a file
func (m *MarkdownIO) WriteArtifact(artifact *models.Artifact) error {
	// Generate filename if not set
	if artifact.Filename == "" {
		artifact.Filename = m.generateArtifactFilename(artifact)
	}

	// Determine directory based on type
	dir := m.getArtifactDirectory(artifact.Type)
	fullPath := filepath.Join(dir, artifact.Filename)

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate markdown content
	content, err := m.generateArtifactContent(artifact)
	if err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	// Write file
	if err := ioutil.WriteFile(fullPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	artifact.Filepath = fullPath
	return nil
}

// generateArtifactContent creates the full markdown file content for Artifact with frontmatter
func (m *MarkdownIO) generateArtifactContent(artifact *models.Artifact) ([]byte, error) {
	var buf bytes.Buffer

	// Write frontmatter
	buf.WriteString("---\n")
	
	// Use a custom YAML encoder to control formatting
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	
	// Create a map to control field order for Artifact
	frontmatter := map[string]interface{}{
		"id":                artifact.ID,
		"type":              artifact.Type,
		"summary":           artifact.Summary,
		"technical_tags":    artifact.TechnicalTags,
		"session_number":    artifact.SessionNumber,
		"created_at":        artifact.CreatedAt,
		"updated_at":        artifact.UpdatedAt,
		"git_context":       artifact.GitContext,
		"related_artifacts": artifact.RelatedArtifacts,
		"work_refs":         artifact.WorkRefs,
		"group_id":          artifact.GroupID,
		"metadata":          artifact.Metadata,
	}
	
	if err := encoder.Encode(frontmatter); err != nil {
		return nil, fmt.Errorf("failed to encode frontmatter: %w", err)
	}
	
	buf.WriteString("---\n\n")
	
	// Write content
	buf.WriteString(artifact.Content)
	if !strings.HasSuffix(artifact.Content, "\n") {
		buf.WriteString("\n")
	}

	return buf.Bytes(), nil
}

// generateArtifactFilename creates a filename for Artifacts: {type}-{brief-description}-{date}-{short-id}.md
func (m *MarkdownIO) generateArtifactFilename(artifact *models.Artifact) string {
	// Extract brief description from summary (first few words, sanitized)
	summary := strings.ToLower(artifact.Summary)
	words := strings.Fields(summary)
	if len(words) > 4 {
		words = words[:4]
	}
	description := strings.Join(words, "-")
	
	// Sanitize description for filename
	description = regexp.MustCompile(`[^a-z0-9-]+`).ReplaceAllString(description, "-")
	description = strings.Trim(description, "-")
	
	// Get date
	date := artifact.CreatedAt.Format("2006-01-02")
	
	// Get short ID (last 6 chars or generate)
	shortID := artifact.ID
	if len(shortID) > 6 {
		shortID = shortID[len(shortID)-6:]
	}
	
	return fmt.Sprintf("%s-%s-%s-%s.md", artifact.Type, description, date, shortID)
}

// getArtifactDirectory returns the appropriate directory for an Artifact
func (m *MarkdownIO) getArtifactDirectory(artifactType string) string {
	switch artifactType {
	case models.TypePlan:
		return filepath.Join(m.baseDir, "artifacts", "plans")
	case models.TypeProposal:
		return filepath.Join(m.baseDir, "artifacts", "proposals")
	case models.TypeAnalysis:
		return filepath.Join(m.baseDir, "artifacts", "analysis")
	case models.TypeUpdate:
		return filepath.Join(m.baseDir, "artifacts", "updates")
	case models.TypeDecision:
		return filepath.Join(m.baseDir, "artifacts", "decisions")
	default:
		return filepath.Join(m.baseDir, "artifacts", "plans") // default to plans
	}
}

// === Enhanced Listing Methods ===

// ListWork returns all Work items from a specific schedule directory
func (m *MarkdownIO) ListWork(schedule string) ([]*models.Work, error) {
	dir := m.getWorkDirectory(schedule)
	return m.listWorkFromDir(dir)
}

// ListAllWork returns all Work items from all schedule directories
func (m *MarkdownIO) ListAllWork() ([]*models.Work, error) {
	var allWork []*models.Work

	// List from all schedule directories
	schedules := []string{models.ScheduleNow, models.ScheduleNext, models.ScheduleLater, models.ScheduleClosed}
	for _, schedule := range schedules {
		items, err := m.ListWork(schedule)
		if err != nil {
			continue // Directory might not exist yet
		}
		allWork = append(allWork, items...)
	}

	return allWork, nil
}

// listWorkFromDir reads all Work markdown files from a directory
func (m *MarkdownIO) listWorkFromDir(dir string) ([]*models.Work, error) {
	log.Printf("listWorkFromDir: Reading directory: %s", dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("listWorkFromDir: Directory does not exist: %s", dir)
			return []*models.Work{}, nil
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	log.Printf("listWorkFromDir: Found %d files in %s", len(files), dir)
	var items []*models.Work
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		filepath := filepath.Join(dir, file.Name())
		log.Printf("listWorkFromDir: Reading file: %s", filepath)
		work, err := m.ReadWork(filepath)
		if err != nil {
			log.Printf("listWorkFromDir: Error reading %s: %v", filepath, err)
			continue // Skip files that can't be parsed
		}

		log.Printf("listWorkFromDir: Successfully parsed work item: %s", work.Title)
		items = append(items, work)
	}

	log.Printf("listWorkFromDir: Returning %d work items", len(items))
	return items, nil
}

// ListArtifacts returns all Artifacts from a specific type directory
func (m *MarkdownIO) ListArtifacts(artifactType string) ([]*models.Artifact, error) {
	dir := m.getArtifactDirectory(artifactType)
	return m.listArtifactsFromDir(dir)
}

// ListAllArtifacts returns all Artifacts from all type directories
func (m *MarkdownIO) ListAllArtifacts() ([]*models.Artifact, error) {
	var allArtifacts []*models.Artifact

	// List from all artifact type directories
	types := []string{models.TypePlan, models.TypeProposal, models.TypeAnalysis, models.TypeUpdate, models.TypeDecision}
	for _, artifactType := range types {
		items, err := m.ListArtifacts(artifactType)
		if err != nil {
			continue // Directory might not exist yet
		}
		allArtifacts = append(allArtifacts, items...)
	}

	return allArtifacts, nil
}

// listArtifactsFromDir reads all Artifact markdown files from a directory
func (m *MarkdownIO) listArtifactsFromDir(dir string) ([]*models.Artifact, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*models.Artifact{}, nil
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var items []*models.Artifact
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		filepath := filepath.Join(dir, file.Name())
		artifact, err := m.ReadArtifact(filepath)
		if err != nil {
			continue // Skip files that can't be parsed
		}

		items = append(items, artifact)
	}

	return items, nil
}

// === Search Methods ===

// SearchWork searches for Work items containing the query
func (m *MarkdownIO) SearchWork(query string) ([]*models.Work, error) {
	allWork, err := m.ListAllWork()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []*models.Work
	
	for _, work := range allWork {
		// Search in title, description, content, and tags
		if strings.Contains(strings.ToLower(work.Title), query) ||
			strings.Contains(strings.ToLower(work.Description), query) ||
			strings.Contains(strings.ToLower(work.Content), query) ||
			containsTag(work.TechnicalTags, query) {
			results = append(results, work)
		}
	}

	return results, nil
}

// SearchArtifacts searches for Artifacts containing the query
func (m *MarkdownIO) SearchArtifacts(query string) ([]*models.Artifact, error) {
	allArtifacts, err := m.ListAllArtifacts()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []*models.Artifact
	
	for _, artifact := range allArtifacts {
		// Search in summary, content, and tags
		if strings.Contains(strings.ToLower(artifact.Summary), query) ||
			strings.Contains(strings.ToLower(artifact.Content), query) ||
			containsTag(artifact.TechnicalTags, query) {
			results = append(results, artifact)
		}
	}

	return results, nil
}