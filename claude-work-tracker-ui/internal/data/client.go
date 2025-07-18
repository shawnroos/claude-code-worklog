package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"claude-work-tracker-ui/internal/models"
)

// Client provides access to work tracker data
type Client struct {
	baseDir           string
	todosDir          string
	findingsDir       string
	workStateDir      string
	projectsDir       string
	localWorkDir      string
	historyDir        string
	activeDir         string
	futureDir         string
	futureItemsDir    string
	futureGroupsDir   string
	currentWorkingDir string
	scanner           *ProjectScanner
}

// NewClient creates a new data client
func NewClient() *Client {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/tmp"
	}
	
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "/"
	}

	// Always use the project root .claude-work directory for global project state
	// This ensures the TUI shows the same view regardless of where it's run from
	scanner := NewProjectScanner()
	projectRoot := scanner.GetAbsoluteProjectRoot()
	
	// CRITICAL: Primary work directory must ALWAYS be at the absolute project root
	// Never allow subdirectories or worktrees to create their own .claude-work
	primaryWorkDir := filepath.Join(projectRoot, ".claude-work")
	
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(primaryWorkDir, 0755); err != nil {
		// If we can't create at project root, something is seriously wrong
		// Log the error but continue with the path (it might be read-only)
		fmt.Fprintf(os.Stderr, "Warning: Could not create .claude-work at project root %s: %v\n", projectRoot, err)
	}

	baseDir := filepath.Join(homeDir, ".claude")

	client := &Client{
		baseDir:           baseDir,
		todosDir:          filepath.Join(baseDir, "todos"),
		findingsDir:       filepath.Join(baseDir, "findings"),
		workStateDir:      filepath.Join(baseDir, "work-state"),
		projectsDir:       filepath.Join(baseDir, "work-state", "projects"),
		localWorkDir:      primaryWorkDir,
		historyDir:        filepath.Join(primaryWorkDir, "history"),
		activeDir:         filepath.Join(primaryWorkDir, "active"),
		futureDir:         filepath.Join(primaryWorkDir, "future"),
		futureItemsDir:    filepath.Join(primaryWorkDir, "future", "items"),
		futureGroupsDir:   filepath.Join(primaryWorkDir, "future", "groups"),
		currentWorkingDir: cwd,
		scanner:           scanner,
	}
	
	return client
}

// GetLocalWorkDir returns the local work directory path
func (c *Client) GetLocalWorkDir() string {
	return c.localWorkDir
}

// GetCurrentWorkState returns the current work state
func (c *Client) GetCurrentWorkState() (*models.WorkState, error) {
	activeTodos, err := c.LoadActiveTodos()
	if err != nil {
		return nil, fmt.Errorf("failed to load active todos: %w", err)
	}

	recentFindings, err := c.LoadRecentFindings()
	if err != nil {
		return nil, fmt.Errorf("failed to load recent findings: %w", err)
	}

	workState := &models.WorkState{
		CurrentSession: fmt.Sprintf("session_%d", time.Now().Unix()),
		ActiveTodos:    activeTodos,
		RecentFindings: recentFindings,
	}

	return workState, nil
}

// LoadActiveTodos loads active todos from the pending todos file
func (c *Client) LoadActiveTodos() ([]models.WorkItem, error) {
	pendingTodosPath := filepath.Join(c.localWorkDir, "PENDING_TODOS.json")
	
	if !fileExists(pendingTodosPath) {
		return []models.WorkItem{}, nil
	}

	data, err := ioutil.ReadFile(pendingTodosPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read pending todos: %w", err)
	}

	var rawTodos []map[string]interface{}
	if err := json.Unmarshal(data, &rawTodos); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pending todos: %w", err)
	}

	var todos []models.WorkItem
	for _, rawTodo := range rawTodos {
		todo := models.WorkItem{
			ID:        getStringValue(rawTodo, "id"),
			Type:      "todo",
			Content:   getStringValue(rawTodo, "content"),
			Status:    getStringValue(rawTodo, "status"),
			SessionID: getStringValue(rawTodo, "session_id"),
			Timestamp: getStringValue(rawTodo, "saved_at"),
		}

		if todo.Timestamp == "" {
			todo.Timestamp = time.Now().Format(time.RFC3339)
		}

		// Set priority in metadata
		priority := getStringValue(rawTodo, "priority")
		if priority == "" {
			priority = "medium"
		}
		todo.Metadata = &models.WorkItemMetadata{
			Priority: priority,
		}

		todos = append(todos, todo)
	}

	return todos, nil
}

// LoadRecentFindings loads recent findings from the findings directory
func (c *Client) LoadRecentFindings() ([]models.Finding, error) {
	if !dirExists(c.findingsDir) {
		return []models.Finding{}, nil
	}

	files, err := ioutil.ReadDir(c.findingsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read findings directory: %w", err)
	}

	var findings []models.Finding
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		finding, err := c.loadFinding(filepath.Join(c.findingsDir, file.Name()))
		if err != nil {
			continue // Skip files that can't be loaded
		}

		findings = append(findings, *finding)
	}

	// Sort by timestamp (newest first)
	sort.Slice(findings, func(i, j int) bool {
		timeI, errI := time.Parse(time.RFC3339, findings[i].Timestamp)
		timeJ, errJ := time.Parse(time.RFC3339, findings[j].Timestamp)
		if errI != nil || errJ != nil {
			return false
		}
		return timeI.After(timeJ)
	})

	// Return only the 20 most recent
	if len(findings) > 20 {
		findings = findings[:20]
	}

	return findings, nil
}

// GetAllWorkItems returns work items from the primary work directory
func (c *Client) GetAllWorkItems() ([]models.WorkItem, error) {
	var workItems []models.WorkItem
	
	// Get active todos from primary work directory
	activeTodos, err := c.LoadActiveTodos()
	if err == nil {
		workItems = append(workItems, activeTodos...)
	}
	
	// Get historical work items from primary work directory
	if dirExists(c.historyDir) {
		files, err := ioutil.ReadDir(c.historyDir)
		if err == nil {
			for _, file := range files {
				if !strings.HasSuffix(file.Name(), ".json") {
					continue
				}

				workItem, err := c.loadWorkItem(filepath.Join(c.historyDir, file.Name()))
				if err != nil {
					continue // Skip files that can't be loaded
				}

				workItems = append(workItems, *workItem)
			}
		}
	}
	
	// Get active work items from primary work directory
	if dirExists(c.activeDir) {
		files, err := ioutil.ReadDir(c.activeDir)
		if err == nil {
			for _, file := range files {
				if !strings.HasSuffix(file.Name(), ".json") {
					continue
				}

				workItem, err := c.loadWorkItem(filepath.Join(c.activeDir, file.Name()))
				if err != nil {
					continue // Skip files that can't be loaded
				}

				workItems = append(workItems, *workItem)
			}
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(workItems, func(i, j int) bool {
		timeI, errI := time.Parse(time.RFC3339, workItems[i].Timestamp)
		timeJ, errJ := time.Parse(time.RFC3339, workItems[j].Timestamp)
		if errI != nil || errJ != nil {
			return false
		}
		return timeI.After(timeJ)
	})

	return workItems, nil
}

// GetProjectWorkItems returns work items from all work directories in the project
func (c *Client) GetProjectWorkItems() ([]models.WorkItem, error) {
	var allWorkItems []models.WorkItem
	
	// Get all work directories in the project
	workDirs := c.scanner.GetAllWorkDirectories()
	
	for _, workDirInfo := range workDirs {
		workDir := workDirInfo.Path
		
		// Load PENDING_TODOS.json if it exists
		pendingTodosPath := filepath.Join(workDir, "PENDING_TODOS.json")
		if fileExists(pendingTodosPath) {
			data, err := ioutil.ReadFile(pendingTodosPath)
			if err == nil {
				var rawTodos []map[string]interface{}
				if json.Unmarshal(data, &rawTodos) == nil {
					for _, rawTodo := range rawTodos {
						todo := models.WorkItem{
							ID:        getStringValue(rawTodo, "id"),
							Type:      "todo",
							Content:   getStringValue(rawTodo, "content"),
							Status:    getStringValue(rawTodo, "status"),
							SessionID: getStringValue(rawTodo, "session_id"),
							Timestamp: getStringValue(rawTodo, "saved_at"),
						}

						if todo.Timestamp == "" {
							todo.Timestamp = time.Now().Format(time.RFC3339)
						}

						priority := getStringValue(rawTodo, "priority")
						if priority == "" {
							priority = "medium"
						}
						todo.Metadata = &models.WorkItemMetadata{
							Priority: priority,
						}

						allWorkItems = append(allWorkItems, todo)
					}
				}
			}
		}
		
		// Load from history directory
		historyDir := filepath.Join(workDir, "history")
		if dirExists(historyDir) {
			files, err := ioutil.ReadDir(historyDir)
			if err == nil {
				for _, file := range files {
					if !strings.HasSuffix(file.Name(), ".json") {
						continue
					}

					workItem, err := c.loadWorkItem(filepath.Join(historyDir, file.Name()))
					if err != nil {
						continue
					}

					allWorkItems = append(allWorkItems, *workItem)
				}
			}
		}
		
		// Load from active directory
		activeDir := filepath.Join(workDir, "active")
		if dirExists(activeDir) {
			files, err := ioutil.ReadDir(activeDir)
			if err == nil {
				for _, file := range files {
					if !strings.HasSuffix(file.Name(), ".json") {
						continue
					}

					workItem, err := c.loadWorkItem(filepath.Join(activeDir, file.Name()))
					if err != nil {
						continue
					}

					allWorkItems = append(allWorkItems, *workItem)
				}
			}
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(allWorkItems, func(i, j int) bool {
		timeI, errI := time.Parse(time.RFC3339, allWorkItems[i].Timestamp)
		timeJ, errJ := time.Parse(time.RFC3339, allWorkItems[j].Timestamp)
		if errI != nil || errJ != nil {
			return false
		}
		return timeI.After(timeJ)
	})

	return allWorkItems, nil
}

// GetCurrentDirectoryInfo returns information about the current working directory and project
func (c *Client) GetCurrentDirectoryInfo() map[string]interface{} {
	info := make(map[string]interface{})
	
	info["working_directory"] = c.currentWorkingDir
	info["project_root"] = c.scanner.GetProjectRoot()
	info["primary_work_dir"] = c.localWorkDir
	info["has_work_tracking"] = dirExists(c.localWorkDir)
	
	// Get all work directories in project
	workDirs := c.scanner.GetAllWorkDirectories()
	info["work_directories"] = workDirs
	info["work_dir_count"] = len(workDirs)
	
	// Count items in different categories
	if dirExists(c.localWorkDir) {
		info["has_active_dir"] = dirExists(c.activeDir)
		info["has_history_dir"] = dirExists(c.historyDir)
		info["has_future_dir"] = dirExists(c.futureDir)
		
		// Count pending todos
		pendingTodosPath := filepath.Join(c.localWorkDir, "PENDING_TODOS.json")
		info["has_pending_todos"] = fileExists(pendingTodosPath)
	}
	
	return info
}

// GetWorkItemsByType returns work items filtered by type
func (c *Client) GetWorkItemsByType(itemType string) ([]models.WorkItem, error) {
	allItems, err := c.GetAllWorkItems()
	if err != nil {
		return nil, err
	}

	var filtered []models.WorkItem
	for _, item := range allItems {
		if item.Type == itemType {
			filtered = append(filtered, item)
		}
	}

	return filtered, nil
}

// SearchWorkItems searches for work items containing the query string
func (c *Client) SearchWorkItems(query string) ([]models.WorkItem, error) {
	allItems, err := c.GetAllWorkItems()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []models.WorkItem
	for _, item := range allItems {
		if strings.Contains(strings.ToLower(item.Content), query) ||
			strings.Contains(strings.ToLower(item.Type), query) ||
			strings.Contains(strings.ToLower(item.Status), query) {
			results = append(results, item)
		}
	}

	return results, nil
}

// GetFutureWorkItems returns future work items from the primary work directory
func (c *Client) GetFutureWorkItems() ([]models.FutureWorkItem, error) {
	// Return future items from primary work directory
	if !dirExists(c.localWorkDir) || !dirExists(c.futureItemsDir) {
		return []models.FutureWorkItem{}, nil
	}

	files, err := ioutil.ReadDir(c.futureItemsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read future items directory: %w", err)
	}

	var futureItems []models.FutureWorkItem
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		futureItem, err := c.loadFutureWorkItem(filepath.Join(c.futureItemsDir, file.Name()))
		if err != nil {
			continue // Skip files that can't be loaded
		}

		futureItems = append(futureItems, *futureItem)
	}

	// Sort by creation date (newest first)
	sort.Slice(futureItems, func(i, j int) bool {
		timeI, errI := time.Parse(time.RFC3339, futureItems[i].CreatedAt)
		timeJ, errJ := time.Parse(time.RFC3339, futureItems[j].CreatedAt)
		if errI != nil || errJ != nil {
			return false
		}
		return timeI.After(timeJ)
	})

	return futureItems, nil
}

// GetFutureWorkGroups returns future work groups from current directory only
func (c *Client) GetFutureWorkGroups() ([]models.FutureWorkGroup, error) {
	// Only return future groups if we're in a directory with local work tracking
	if !dirExists(c.localWorkDir) || !dirExists(c.futureGroupsDir) {
		return []models.FutureWorkGroup{}, nil
	}

	files, err := ioutil.ReadDir(c.futureGroupsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read future groups directory: %w", err)
	}

	var futureGroups []models.FutureWorkGroup
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		futureGroup, err := c.loadFutureWorkGroup(filepath.Join(c.futureGroupsDir, file.Name()))
		if err != nil {
			continue // Skip files that can't be loaded
		}

		futureGroups = append(futureGroups, *futureGroup)
	}

	// Sort by last updated (newest first)
	sort.Slice(futureGroups, func(i, j int) bool {
		timeI, errI := time.Parse(time.RFC3339, futureGroups[i].LastUpdated)
		timeJ, errJ := time.Parse(time.RFC3339, futureGroups[j].LastUpdated)
		if errI != nil || errJ != nil {
			return false
		}
		return timeI.After(timeJ)
	})

	return futureGroups, nil
}

// Helper methods

func (c *Client) loadWorkItem(filePath string) (*models.WorkItem, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var workItem models.WorkItem
	if err := json.Unmarshal(data, &workItem); err != nil {
		return nil, err
	}

	return &workItem, nil
}

func (c *Client) loadFinding(filePath string) (*models.Finding, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var finding models.Finding
	if err := json.Unmarshal(data, &finding); err != nil {
		return nil, err
	}

	return &finding, nil
}

func (c *Client) loadFutureWorkItem(filePath string) (*models.FutureWorkItem, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var futureItem models.FutureWorkItem
	if err := json.Unmarshal(data, &futureItem); err != nil {
		return nil, err
	}

	return &futureItem, nil
}

func (c *Client) loadFutureWorkGroup(filePath string) (*models.FutureWorkGroup, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var futureGroup models.FutureWorkGroup
	if err := json.Unmarshal(data, &futureGroup); err != nil {
		return nil, err
	}

	return &futureGroup, nil
}

// Utility functions

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}