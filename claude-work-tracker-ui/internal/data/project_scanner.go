package data

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// ProjectScanner finds work directories within a project tree
type ProjectScanner struct {
	currentDir string
	projectRoot string
	workDirs   []string
}

// NewProjectScanner creates a new project scanner
func NewProjectScanner() *ProjectScanner {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "/"
	}

	scanner := &ProjectScanner{
		currentDir: cwd,
		workDirs:   []string{},
	}

	scanner.findProjectRoot()
	scanner.scanForWorkDirectories()

	return scanner
}

// findProjectRoot walks up the directory tree to find the project root
func (s *ProjectScanner) findProjectRoot() {
	current := s.currentDir

	// Look for common project indicators
	for {
		// Check for git root - but handle worktrees properly
		gitPath := filepath.Join(current, ".git")
		if dirExists(gitPath) {
			s.projectRoot = current
			return
		}
		
		// Check for git worktree (file pointing to main repo)
		if fileExists(gitPath) {
			// This is a git worktree, find the main repository
			if mainRepo := s.findMainRepositoryFromWorktree(gitPath); mainRepo != "" {
				s.projectRoot = mainRepo
				return
			}
			// If we can't find main repo, use current directory
			s.projectRoot = current
			return
		}

		// Check for package.json, go.mod, etc.
		for _, indicator := range []string{"package.json", "go.mod", "Cargo.toml", ".project", "pyproject.toml"} {
			if fileExists(filepath.Join(current, indicator)) {
				s.projectRoot = current
				return
			}
		}

		// Check for .claude-work directory
		if dirExists(filepath.Join(current, ".claude-work")) {
			s.projectRoot = current
			return
		}

		parent := filepath.Dir(current)
		if parent == current || parent == "/" {
			// Reached filesystem root, use current directory
			s.projectRoot = s.currentDir
			return
		}
		current = parent
	}
}

// scanForWorkDirectories finds all .claude-work directories within the project
func (s *ProjectScanner) scanForWorkDirectories() {
	if s.projectRoot == "" {
		return
	}

	// Walk the project tree looking for .claude-work directories
	filepath.Walk(s.projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on errors
		}

		// Skip hidden directories except .claude-work
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != ".claude-work" {
			return filepath.SkipDir
		}

		// Skip common ignore patterns
		if info.IsDir() {
			switch info.Name() {
			case "node_modules", "vendor", "target", "__pycache__", ".next", "dist", "build":
				return filepath.SkipDir
			}
		}

		// Found a .claude-work directory
		if info.IsDir() && info.Name() == ".claude-work" {
			s.workDirs = append(s.workDirs, path)
		}

		return nil
	})
}

// GetProjectRoot returns the detected project root
func (s *ProjectScanner) GetProjectRoot() string {
	return s.projectRoot
}

// GetWorkDirectories returns all found .claude-work directories
func (s *ProjectScanner) GetWorkDirectories() []string {
	return s.workDirs
}

// GetPrimaryWorkDirectory returns the most relevant work directory
func (s *ProjectScanner) GetPrimaryWorkDirectory() string {
	if len(s.workDirs) == 0 {
		return ""
	}

	// Prefer work directory in current directory or below
	for _, dir := range s.workDirs {
		if strings.HasPrefix(dir, s.currentDir) {
			return dir
		}
	}
	
	// Next, prefer work directory in current path (parent contains current)
	for _, dir := range s.workDirs {
		if strings.HasPrefix(s.currentDir, filepath.Dir(dir)) {
			return dir
		}
	}

	// Prefer work directory at project root
	for _, dir := range s.workDirs {
		if filepath.Dir(dir) == s.projectRoot {
			return dir
		}
	}

	// Return first found
	return s.workDirs[0]
}

// GetAllWorkDirectories returns all work directories with metadata
func (s *ProjectScanner) GetAllWorkDirectories() []WorkDirectoryInfo {
	var result []WorkDirectoryInfo

	for _, dir := range s.workDirs {
		parentDir := filepath.Dir(dir)
		relPath, _ := filepath.Rel(s.projectRoot, parentDir)
		if relPath == "." {
			relPath = "Project Root"
		}

		info := WorkDirectoryInfo{
			Path:        dir,
			ParentDir:   parentDir,
			RelativePath: relPath,
			IsActive:    strings.HasPrefix(s.currentDir, parentDir),
			HasActive:   dirExists(filepath.Join(dir, "active")),
			HasHistory:  dirExists(filepath.Join(dir, "history")),
			HasFuture:   dirExists(filepath.Join(dir, "future")),
			HasPending:  fileExists(filepath.Join(dir, "PENDING_TODOS.json")),
		}

		result = append(result, info)
	}

	return result
}

// WorkDirectoryInfo contains metadata about a work directory
type WorkDirectoryInfo struct {
	Path         string
	ParentDir    string
	RelativePath string
	IsActive     bool
	HasActive    bool
	HasHistory   bool
	HasFuture    bool
	HasPending   bool
}

// findMainRepositoryFromWorktree reads a git worktree file to find the main repository
func (s *ProjectScanner) findMainRepositoryFromWorktree(gitFilePath string) string {
	file, err := os.Open(gitFilePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "gitdir: ") {
			gitdir := strings.TrimPrefix(line, "gitdir: ")
			// gitdir points to something like: /path/to/main/repo/.git/worktrees/name
			// We want to extract /path/to/main/repo
			if strings.Contains(gitdir, "/.git/worktrees/") {
				parts := strings.Split(gitdir, "/.git/worktrees/")
				if len(parts) > 0 {
					return parts[0]
				}
			}
		}
	}
	return ""
}