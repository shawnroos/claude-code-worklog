package storage

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// ProjectScanner finds the project root directory
type ProjectScanner struct {
	currentDir  string
	projectRoot string
}

// NewProjectScanner creates a new project scanner
func NewProjectScanner() *ProjectScanner {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "/"
	}

	scanner := &ProjectScanner{
		currentDir: cwd,
	}

	scanner.findProjectRoot()
	return scanner
}

// findProjectRoot walks up the directory tree to find the project root
func (s *ProjectScanner) findProjectRoot() {
	current := s.currentDir

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
		for _, indicator := range []string{"package.json", "go.mod", "Cargo.toml", ".project", "pyproject.toml", "setup.py", "requirements.txt"} {
			if fileExists(filepath.Join(current, indicator)) {
				s.projectRoot = current
				return
			}
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

// GetProjectRoot returns the detected project root
func (s *ProjectScanner) GetProjectRoot() string {
	return s.projectRoot
}

// GetCurrentDir returns the current working directory
func (s *ProjectScanner) GetCurrentDir() string {
	return s.currentDir
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

// Helper functions
func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}