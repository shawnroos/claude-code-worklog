package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ExternalStorage manages work data outside of git repositories
type ExternalStorage struct {
	BaseDir      string
	ProjectsDir  string
	WorkDir      string
	ArtifactsDir string
	ConfigDir    string
}

// NewExternalStorage creates a new external storage manager
func NewExternalStorage() (*ExternalStorage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, ".claude", "work-data")
	
	storage := &ExternalStorage{
		BaseDir:      baseDir,
		ProjectsDir:  filepath.Join(baseDir, "projects"),
		WorkDir:      filepath.Join(baseDir, "work"),
		ArtifactsDir: filepath.Join(baseDir, "artifacts"),
		ConfigDir:    filepath.Join(homeDir, ".claude", "config"),
	}

	// Ensure all directories exist
	dirs := []string{
		storage.BaseDir,
		storage.ProjectsDir,
		storage.WorkDir,
		storage.ArtifactsDir,
		storage.ConfigDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return storage, nil
}

// Project represents a registered project
type Project struct {
	ID           string    `json:"id"`
	Path         string    `json:"path"`
	Name         string    `json:"name"`
	RemoteURL    string    `json:"remote_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	LastAccess   time.Time `json:"last_access"`
	WorkItems    int       `json:"work_items"`
	ActiveBranch string    `json:"active_branch,omitempty"`
}

// ProjectRegistry manages all registered projects
type ProjectRegistry struct {
	Projects map[string]*Project `json:"projects"`
	storage  *ExternalStorage
}

// LoadProjectRegistry loads or creates the project registry
func LoadProjectRegistry(storage *ExternalStorage) (*ProjectRegistry, error) {
	registryPath := filepath.Join(storage.ProjectsDir, "project-index.json")
	
	registry := &ProjectRegistry{
		Projects: make(map[string]*Project),
		storage:  storage,
	}

	// Load existing registry if it exists
	if data, err := ioutil.ReadFile(registryPath); err == nil {
		if err := json.Unmarshal(data, registry); err != nil {
			return nil, fmt.Errorf("failed to parse project registry: %w", err)
		}
	}

	return registry, nil
}

// Save persists the project registry
func (r *ProjectRegistry) Save() error {
	registryPath := filepath.Join(r.storage.ProjectsDir, "project-index.json")
	
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
	}

	if err := ioutil.WriteFile(registryPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write registry: %w", err)
	}

	return nil
}

// RegisterProject registers a project or updates last access time
func (r *ProjectRegistry) RegisterProject(projectPath string) (*Project, error) {
	// Generate consistent project ID
	id := generateProjectID(projectPath)
	
	// Check if project already registered
	if project, exists := r.Projects[id]; exists {
		project.LastAccess = time.Now()
		project.Path = projectPath // Update path in case it moved
		
		// Update branch info
		if branch := getCurrentGitBranch(projectPath); branch != "" {
			project.ActiveBranch = branch
		}
		
		return project, r.Save()
	}

	// Create new project registration
	project := &Project{
		ID:         id,
		Path:       projectPath,
		Name:       filepath.Base(projectPath),
		RemoteURL:  getGitRemoteURL(projectPath),
		CreatedAt:  time.Now(),
		LastAccess: time.Now(),
	}

	// Get current branch
	if branch := getCurrentGitBranch(projectPath); branch != "" {
		project.ActiveBranch = branch
	}

	r.Projects[id] = project
	
	// Create project work directories
	projectDirs := []string{
		filepath.Join(r.storage.WorkDir, id, "now"),
		filepath.Join(r.storage.WorkDir, id, "next"),
		filepath.Join(r.storage.WorkDir, id, "later"),
		filepath.Join(r.storage.WorkDir, id, "closed"),
		filepath.Join(r.storage.ArtifactsDir, id),
	}

	for _, dir := range projectDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create project directory %s: %w", dir, err)
		}
	}

	return project, r.Save()
}

// GetProject retrieves a project by ID
func (r *ProjectRegistry) GetProject(id string) (*Project, bool) {
	project, exists := r.Projects[id]
	return project, exists
}

// ListProjects returns all registered projects
func (r *ProjectRegistry) ListProjects() []*Project {
	projects := make([]*Project, 0, len(r.Projects))
	for _, project := range r.Projects {
		projects = append(projects, project)
	}
	return projects
}

// generateProjectID creates a consistent ID for a project
func generateProjectID(projectPath string) string {
	// Try to get git remote URL first
	if remoteURL := getGitRemoteURL(projectPath); remoteURL != "" {
		// Use remote URL for consistent ID across clones
		return hashString(remoteURL)
	}
	
	// Fallback to absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		absPath = projectPath
	}
	
	// Include hostname for uniqueness across machines
	hostname, _ := os.Hostname()
	uniqueID := fmt.Sprintf("%s:%s", hostname, absPath)
	
	return hashString(uniqueID)
}

// hashString creates a short hash of a string
func hashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:8]) // Use first 8 bytes for shorter ID
}

// getGitRemoteURL gets the origin remote URL for a git repository
func getGitRemoteURL(projectPath string) string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = projectPath
	
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	
	return strings.TrimSpace(string(output))
}

// getCurrentGitBranch gets the current git branch
func getCurrentGitBranch(projectPath string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = projectPath
	
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	
	return strings.TrimSpace(string(output))
}

// GetProjectWorkDir returns the work directory for a specific project
func (s *ExternalStorage) GetProjectWorkDir(projectID string) string {
	return filepath.Join(s.WorkDir, projectID)
}

// GetProjectArtifactsDir returns the artifacts directory for a specific project
func (s *ExternalStorage) GetProjectArtifactsDir(projectID string) string {
	return filepath.Join(s.ArtifactsDir, projectID)
}

// MigrateFromRepository migrates work items from repository .claude-work to external storage
func (s *ExternalStorage) MigrateFromRepository(repoPath string, projectID string) error {
	oldWorkDir := filepath.Join(repoPath, ".claude-work", "work")
	if _, err := os.Stat(oldWorkDir); os.IsNotExist(err) {
		return nil // Nothing to migrate
	}

	newWorkDir := s.GetProjectWorkDir(projectID)
	
	// Copy all work items
	schedules := []string{"now", "next", "later", "closed"}
	for _, schedule := range schedules {
		oldDir := filepath.Join(oldWorkDir, schedule)
		newDir := filepath.Join(newWorkDir, schedule)
		
		if err := copyDirectory(oldDir, newDir); err != nil {
			return fmt.Errorf("failed to migrate %s: %w", schedule, err)
		}
	}

	// Backup old directory instead of deleting
	backupPath := filepath.Join(repoPath, ".claude-work.backup")
	if err := os.Rename(filepath.Join(repoPath, ".claude-work"), backupPath); err != nil {
		// Log but don't fail
		fmt.Fprintf(os.Stderr, "Warning: Could not backup old .claude-work: %v\n", err)
	}

	return nil
}

// copyDirectory recursively copies a directory
func copyDirectory(src, dst string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil // Source doesn't exist, nothing to copy
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	
	return ioutil.WriteFile(dst, data, 0644)
}