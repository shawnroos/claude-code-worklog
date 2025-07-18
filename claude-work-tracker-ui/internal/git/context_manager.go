package git

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/shawnroos/claude-work-tracker-ui/internal/model"
)

// ContextManager manages Git context for work items
type ContextManager struct {
	mu         sync.RWMutex
	cache      map[string]*model.GitContext
	cacheTTL   time.Duration
	lastUpdate map[string]time.Time
}

// NewContextManager creates a new Git context manager
func NewContextManager() *ContextManager {
	return &ContextManager{
		cache:      make(map[string]*model.GitContext),
		cacheTTL:   5 * time.Minute,
		lastUpdate: make(map[string]time.Time),
	}
}

// GetContext retrieves Git context for a directory
func (cm *ContextManager) GetContext(ctx context.Context, workingDir string) (*model.GitContext, error) {
	cm.mu.RLock()
	if cached, ok := cm.cache[workingDir]; ok {
		if time.Since(cm.lastUpdate[workingDir]) < cm.cacheTTL {
			cm.mu.RUnlock()
			return cached, nil
		}
	}
	cm.mu.RUnlock()

	// Fetch fresh context
	gitCtx, err := cm.fetchGitContext(ctx, workingDir)
	if err != nil {
		return nil, err
	}

	// Update cache
	cm.mu.Lock()
	cm.cache[workingDir] = gitCtx
	cm.lastUpdate[workingDir] = time.Now()
	cm.mu.Unlock()

	return gitCtx, nil
}

// fetchGitContext retrieves Git information for a directory
func (cm *ContextManager) fetchGitContext(ctx context.Context, workingDir string) (*model.GitContext, error) {
	gitCtx := &model.GitContext{
		WorkingDirectory: workingDir,
	}

	// Check if directory is in a Git repository
	if !cm.isGitRepo(ctx, workingDir) {
		return gitCtx, nil
	}

	// Get current branch
	branch, err := cm.execGitCommand(ctx, workingDir, "rev-parse", "--abbrev-ref", "HEAD")
	if err == nil {
		gitCtx.Branch = strings.TrimSpace(branch)
	}

	// Get worktree name
	worktree, err := cm.getWorktreeName(ctx, workingDir)
	if err == nil && worktree != "" {
		gitCtx.Worktree = worktree
	}

	// Get remote URL
	remoteURL, err := cm.execGitCommand(ctx, workingDir, "remote", "get-url", "origin")
	if err == nil {
		gitCtx.RemoteURL = strings.TrimSpace(remoteURL)
	}

	return gitCtx, nil
}

// UpdateWorkItemContext updates a work item with current Git context
func (cm *ContextManager) UpdateWorkItemContext(ctx context.Context, work *model.Work) error {
	if work.GitContext.WorkingDirectory == "" {
		// Try to determine working directory from current location
		wd, err := os.Getwd()
		if err == nil {
			work.GitContext.WorkingDirectory = wd
		}
	}

	if work.GitContext.WorkingDirectory != "" {
		gitCtx, err := cm.GetContext(ctx, work.GitContext.WorkingDirectory)
		if err != nil {
			return fmt.Errorf("failed to get git context: %w", err)
		}
		work.GitContext = *gitCtx
	}

	return nil
}

// GetCommitInfo retrieves information about the latest commit
func (cm *ContextManager) GetCommitInfo(ctx context.Context, workingDir string) (map[string]string, error) {
	info := make(map[string]string)

	// Get latest commit hash
	hash, err := cm.execGitCommand(ctx, workingDir, "rev-parse", "HEAD")
	if err != nil {
		return info, err
	}
	info["hash"] = strings.TrimSpace(hash)

	// Get commit message
	message, err := cm.execGitCommand(ctx, workingDir, "log", "-1", "--pretty=%B")
	if err == nil {
		info["message"] = strings.TrimSpace(message)
	}

	// Get author
	author, err := cm.execGitCommand(ctx, workingDir, "log", "-1", "--pretty=%an")
	if err == nil {
		info["author"] = strings.TrimSpace(author)
	}

	// Get timestamp
	timestamp, err := cm.execGitCommand(ctx, workingDir, "log", "-1", "--pretty=%at")
	if err == nil {
		info["timestamp"] = strings.TrimSpace(timestamp)
	}

	return info, nil
}

// GetFileChanges retrieves uncommitted changes in the repository
func (cm *ContextManager) GetFileChanges(ctx context.Context, workingDir string) ([]string, error) {
	output, err := cm.execGitCommand(ctx, workingDir, "status", "--porcelain")
	if err != nil {
		return nil, err
	}

	var changes []string
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 3 {
			changes = append(changes, strings.TrimSpace(line[3:]))
		}
	}

	return changes, nil
}

// GetBranchActivity checks for recent activity on a branch
func (cm *ContextManager) GetBranchActivity(ctx context.Context, workingDir string, branch string) (map[string]interface{}, error) {
	activity := make(map[string]interface{})

	// Get commit count
	count, err := cm.execGitCommand(ctx, workingDir, "rev-list", "--count", branch)
	if err == nil {
		var commitCount int
		fmt.Sscanf(strings.TrimSpace(count), "%d", &commitCount)
		activity["commit_count"] = commitCount
	}

	// Get last commit date
	lastDate, err := cm.execGitCommand(ctx, workingDir, "log", "-1", "--pretty=%at", branch)
	if err == nil {
		var timestamp int64
		fmt.Sscanf(strings.TrimSpace(lastDate), "%d", &timestamp)
		activity["last_commit"] = time.Unix(timestamp, 0)
	}

	// Check if branch has upstream
	upstream, err := cm.execGitCommand(ctx, workingDir, "rev-parse", "--abbrev-ref", branch+"@{upstream}")
	activity["has_upstream"] = err == nil && upstream != ""

	// Get ahead/behind count if has upstream
	if activity["has_upstream"].(bool) {
		aheadBehind, err := cm.execGitCommand(ctx, workingDir, "rev-list", "--left-right", "--count", branch+"..."+strings.TrimSpace(upstream))
		if err == nil {
			var ahead, behind int
			fmt.Sscanf(strings.TrimSpace(aheadBehind), "%d\t%d", &ahead, &behind)
			activity["ahead"] = ahead
			activity["behind"] = behind
		}
	}

	return activity, nil
}

// Helper methods

func (cm *ContextManager) isGitRepo(ctx context.Context, dir string) bool {
	_, err := cm.execGitCommand(ctx, dir, "rev-parse", "--git-dir")
	return err == nil
}

func (cm *ContextManager) getWorktreeName(ctx context.Context, dir string) (string, error) {
	// Try to get worktree name from git worktree list
	output, err := cm.execGitCommand(ctx, dir, "worktree", "list", "--porcelain")
	if err != nil {
		return "", err
	}

	// Parse worktree list to find current directory
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	var currentWorktree string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "worktree ") {
			currentWorktree = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "branch ") && currentWorktree == absDir {
			// Extract branch name as worktree name
			branch := strings.TrimPrefix(line, "branch refs/heads/")
			return branch, nil
		}
	}

	// If not in a worktree, return empty
	return "", nil
}

func (cm *ContextManager) execGitCommand(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil // Ignore stderr for cleaner output

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

// WatchForChanges monitors Git repository for changes
func (cm *ContextManager) WatchForChanges(ctx context.Context, workingDir string, callback func(event string)) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	var lastCommit string
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Check for new commits
			commitInfo, err := cm.GetCommitInfo(ctx, workingDir)
			if err == nil {
				currentCommit := commitInfo["hash"]
				if lastCommit != "" && lastCommit != currentCommit {
					callback("new_commit")
				}
				lastCommit = currentCommit
			}

			// Check for uncommitted changes
			changes, err := cm.GetFileChanges(ctx, workingDir)
			if err == nil && len(changes) > 0 {
				callback("uncommitted_changes")
			}
		}
	}
}