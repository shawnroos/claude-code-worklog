package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"claude-work-tracker-ui/internal/data"
	"claude-work-tracker-ui/internal/models"
)

// CentralizedClient provides access to work data stored outside repositories
type CentralizedClient struct {
	storage    *ExternalStorage
	registry   *ProjectRegistry
	project    *Project
	markdownIO *data.MarkdownIO
	scanner    *ProjectScanner
}

// NewCentralizedClient creates a new centralized data client
func NewCentralizedClient() (*CentralizedClient, error) {
	// Initialize external storage
	storage, err := NewExternalStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize external storage: %w", err)
	}

	// Load project registry
	registry, err := LoadProjectRegistry(storage)
	if err != nil {
		return nil, fmt.Errorf("failed to load project registry: %w", err)
	}

	// Find current project
	scanner := NewProjectScanner()
	projectRoot := scanner.GetProjectRoot()
	
	// Register or update project
	project, err := registry.RegisterProject(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to register project: %w", err)
	}

	// Initialize markdown IO with external storage path
	projectWorkDir := storage.GetProjectWorkDir(project.ID)
	markdownIO := data.NewMarkdownIO(projectWorkDir)

	// Attempt migration from repository
	if err := storage.MigrateFromRepository(projectRoot, project.ID); err != nil {
		// Log but don't fail
		fmt.Fprintf(os.Stderr, "Warning: Migration failed: %v\n", err)
	}

	client := &CentralizedClient{
		storage:    storage,
		registry:   registry,
		project:    project,
		markdownIO: markdownIO,
		scanner:    scanner,
	}

	return client, nil
}

// GetWorkDir returns the external work directory for the current project
func (c *CentralizedClient) GetWorkDir() string {
	return c.storage.GetProjectWorkDir(c.project.ID)
}

// GetArtifactsDir returns the external artifacts directory for the current project
func (c *CentralizedClient) GetArtifactsDir() string {
	return c.storage.GetProjectArtifactsDir(c.project.ID)
}

// GetCurrentProject returns the current project info
func (c *CentralizedClient) GetCurrentProject() *Project {
	return c.project
}

// GetAllProjects returns all registered projects
func (c *CentralizedClient) GetAllProjects() []*Project {
	return c.registry.ListProjects()
}

// GetProjectByID retrieves a project by its ID
func (c *CentralizedClient) GetProjectByID(id string) (*Project, error) {
	project, exists := c.registry.GetProject(id)
	if !exists {
		return nil, fmt.Errorf("project not found: %s", id)
	}
	return project, nil
}

// SwitchProject switches the client to work with a different project
func (c *CentralizedClient) SwitchProject(projectID string) error {
	project, exists := c.registry.GetProject(projectID)
	if !exists {
		return fmt.Errorf("project not found: %s", projectID)
	}

	c.project = project
	projectWorkDir := c.storage.GetProjectWorkDir(project.ID)
	c.markdownIO = data.NewMarkdownIO(projectWorkDir)
	
	return nil
}

// GetWorkBySchedule returns work items for the current project by schedule
func (c *CentralizedClient) GetWorkBySchedule(schedule string) ([]*models.Work, error) {
	return c.markdownIO.ListWork(strings.ToLower(schedule))
}

// GetAllWork returns all work items for the current project
func (c *CentralizedClient) GetAllWork() ([]*models.Work, error) {
	return c.markdownIO.ListAllWork()
}

// CreateWork creates a new work item in external storage
func (c *CentralizedClient) CreateWork(work *models.Work) error {
	// Ensure git context includes project info
	work.GitContext.ProjectID = c.project.ID
	work.GitContext.ProjectPath = c.project.Path
	
	return c.markdownIO.WriteWork(work)
}

// UpdateWork updates an existing work item
func (c *CentralizedClient) UpdateWork(work *models.Work) error {
	return c.markdownIO.WriteWork(work)
}

// GetCrossProjectWork returns work items across all projects
func (c *CentralizedClient) GetCrossProjectWork(schedule string) (map[string][]*models.Work, error) {
	results := make(map[string][]*models.Work)
	
	for _, project := range c.registry.ListProjects() {
		projectWorkDir := c.storage.GetProjectWorkDir(project.ID)
		projectIO := data.NewMarkdownIO(projectWorkDir)
		
		work, err := projectIO.ListWork(strings.ToLower(schedule))
		if err != nil {
			continue // Skip projects with errors
		}
		
		if len(work) > 0 {
			results[project.Name] = work
		}
	}
	
	return results, nil
}

// SearchAcrossProjects searches for work items across all projects
func (c *CentralizedClient) SearchAcrossProjects(query string) (map[string][]*models.Work, error) {
	results := make(map[string][]*models.Work)
	
	for _, project := range c.registry.ListProjects() {
		projectWorkDir := c.storage.GetProjectWorkDir(project.ID)
		projectIO := data.NewMarkdownIO(projectWorkDir)
		
		work, err := projectIO.SearchWork(query)
		if err != nil {
			continue // Skip projects with errors
		}
		
		if len(work) > 0 {
			results[project.Name] = work
		}
	}
	
	return results, nil
}

// CleanupOldRepositoryStorage removes .claude-work from the repository
func (c *CentralizedClient) CleanupOldRepositoryStorage() error {
	repoWorkDir := filepath.Join(c.scanner.GetProjectRoot(), ".claude-work")
	
	// Check if it exists
	if _, err := os.Stat(repoWorkDir); os.IsNotExist(err) {
		return nil // Already clean
	}

	// Check if it's already backed up
	backupPath := repoWorkDir + ".backup"
	if _, err := os.Stat(backupPath); err == nil {
		// Backup exists, safe to remove original
		return os.RemoveAll(repoWorkDir)
	}

	// Create backup first
	if err := os.Rename(repoWorkDir, backupPath); err != nil {
		return fmt.Errorf("failed to backup old storage: %w", err)
	}

	return nil
}

// GetStorageStats returns statistics about the external storage
func (c *CentralizedClient) GetStorageStats() (*StorageStats, error) {
	stats := &StorageStats{
		ProjectCount: len(c.registry.Projects),
		Projects:     make(map[string]*ProjectStats),
	}

	for _, project := range c.registry.ListProjects() {
		projectStats := &ProjectStats{
			Project: project,
			WorkCounts: map[string]int{
				"now":    0,
				"next":   0,
				"later":  0,
				"closed": 0,
			},
		}

		projectWorkDir := c.storage.GetProjectWorkDir(project.ID)
		projectIO := data.NewMarkdownIO(projectWorkDir)

		// Count work items by schedule
		for _, schedule := range []string{"now", "next", "later", "closed"} {
			work, _ := projectIO.ListWork(schedule)
			projectStats.WorkCounts[schedule] = len(work)
			projectStats.TotalWork += len(work)
			stats.TotalWork += len(work)
		}

		stats.Projects[project.ID] = projectStats
	}

	return stats, nil
}

// StorageStats contains statistics about the external storage
type StorageStats struct {
	ProjectCount int                       `json:"project_count"`
	TotalWork    int                       `json:"total_work"`
	Projects     map[string]*ProjectStats  `json:"projects"`
}

// ProjectStats contains statistics for a single project
type ProjectStats struct {
	Project    *Project       `json:"project"`
	TotalWork  int            `json:"total_work"`
	WorkCounts map[string]int `json:"work_counts"`
}