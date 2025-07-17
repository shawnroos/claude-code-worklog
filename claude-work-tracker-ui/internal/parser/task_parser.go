package parser

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"claude-work-tracker-ui/internal/models"
)

// TaskParser handles extraction and parsing of tasks from markdown content
type TaskParser struct {
	// Regex patterns for different task formats
	taskPattern   *regexp.Regexp
	phasePattern  *regexp.Regexp
	headerPattern *regexp.Regexp
}

// NewTaskParser creates a new task parser
func NewTaskParser() *TaskParser {
	return &TaskParser{
		// Matches: - [x] Task title, - [ ] Task title, etc.
		taskPattern: regexp.MustCompile(`^(\s*)-\s*(\[[x\s…!-]\])\s*(.+)$`),
		// Matches: ### Phase 1: Something, ## Phase 2: Something, etc.
		phasePattern: regexp.MustCompile(`^#{1,6}\s*(?:Phase\s+\d+|phase\s+\d+):\s*(.+)$`),
		// Matches: ### Something, ## Something (general headers for categories)
		headerPattern: regexp.MustCompile(`^(#{1,6})\s*(.+)$`),
	}
}

// ParsedTask represents a task extracted from markdown
type ParsedTask struct {
	Task       *models.Task
	RawLine    string
	LineNumber int
	Indentation int
}

// TaskExtractionResult contains all tasks found in markdown content
type TaskExtractionResult struct {
	Tasks  []ParsedTask
	Phases []Phase
}

// Phase represents a detected phase or category in the markdown
type Phase struct {
	Name        string
	Level       int // Header level (1-6)
	LineNumber  int
	TaskIndices []int // Indices of tasks under this phase
}

// ExtractTasksFromMarkdown parses markdown content and extracts all tasks
func (p *TaskParser) ExtractTasksFromMarkdown(content string, source string) *TaskExtractionResult {
	lines := strings.Split(content, "\n")
	result := &TaskExtractionResult{
		Tasks:  []ParsedTask{},
		Phases: []Phase{},
	}
	
	var currentPhase *Phase
	
	for i, line := range lines {
		lineNum := i + 1
		
		// Check for phase/category headers
		if phase := p.parsePhase(line, lineNum); phase != nil {
			// Save previous phase
			if currentPhase != nil {
				result.Phases = append(result.Phases, *currentPhase)
			}
			currentPhase = phase
			continue
		}
		
		// Check for tasks
		if task := p.parseTask(line, lineNum, source); task != nil {
			taskIndex := len(result.Tasks)
			
			// Associate with current phase
			if currentPhase != nil {
				task.Task.Phase = currentPhase.Name
				currentPhase.TaskIndices = append(currentPhase.TaskIndices, taskIndex)
			}
			
			result.Tasks = append(result.Tasks, *task)
		}
	}
	
	// Don't forget the last phase
	if currentPhase != nil {
		result.Phases = append(result.Phases, *currentPhase)
	}
	
	return result
}

// parseTask extracts a task from a single line
func (p *TaskParser) parseTask(line string, lineNum int, source string) *ParsedTask {
	matches := p.taskPattern.FindStringSubmatch(line)
	if len(matches) != 4 {
		return nil
	}
	
	indentation := len(matches[1])
	checkbox := matches[2]
	title := strings.TrimSpace(matches[3])
	
	// Generate task ID
	taskID := fmt.Sprintf("task-%d-%s", time.Now().UnixNano(), generateShortID())
	
	task := &models.Task{
		ID:         taskID,
		Title:      title,
		Status:     models.TaskStatusFromMarkdown(checkbox),
		Source:     source,
		LineNumber: lineNum,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	// Set completion time if completed
	if task.Status == models.TaskStatusCompleted {
		now := time.Now()
		task.CompletedAt = &now
	}
	
	return &ParsedTask{
		Task:        task,
		RawLine:     line,
		LineNumber:  lineNum,
		Indentation: indentation,
	}
}

// parsePhase extracts phase information from a header line
func (p *TaskParser) parsePhase(line string, lineNum int) *Phase {
	// First try explicit phase pattern
	matches := p.phasePattern.FindStringSubmatch(line)
	if len(matches) == 2 {
		return &Phase{
			Name:        strings.TrimSpace(matches[1]),
			Level:       strings.Count(line, "#"),
			LineNumber:  lineNum,
			TaskIndices: []int{},
		}
	}
	
	// Try general header pattern as category
	matches = p.headerPattern.FindStringSubmatch(line)
	if len(matches) == 3 {
		level := len(matches[1])
		name := strings.TrimSpace(matches[2])
		
		// Skip very generic headers
		if level <= 2 && !isLikelyPhaseHeader(name) {
			return nil
		}
		
		return &Phase{
			Name:        name,
			Level:       level,
			LineNumber:  lineNum,
			TaskIndices: []int{},
		}
	}
	
	return nil
}

// isLikelyPhaseHeader checks if a header looks like it should group tasks
func isLikelyPhaseHeader(name string) bool {
	lowerName := strings.ToLower(name)
	
	// Common phase/category indicators
	phaseWords := []string{
		"phase", "step", "stage", "part", "section",
		"frontend", "backend", "testing", "deployment",
		"setup", "implementation", "cleanup", "documentation",
		"requirements", "design", "development", "validation",
		"tasks", "todo", "work", "items",
	}
	
	for _, word := range phaseWords {
		if strings.Contains(lowerName, word) {
			return true
		}
	}
	
	// Numbered sections
	if regexp.MustCompile(`\b\d+\b`).MatchString(name) {
		return true
	}
	
	return false
}

// UpdateTaskInMarkdown updates a task's status in markdown content
func (p *TaskParser) UpdateTaskInMarkdown(content string, taskID string, newStatus models.TaskStatus) string {
	lines := strings.Split(content, "\n")
	
	for i, line := range lines {
		if task := p.parseTask(line, i+1, ""); task != nil && task.Task.ID == taskID {
			// Replace the checkbox
			newCheckbox := models.TaskStatusToMarkdown(newStatus)
			lines[i] = p.taskPattern.ReplaceAllStringFunc(line, func(match string) string {
				parts := p.taskPattern.FindStringSubmatch(match)
				return fmt.Sprintf("%s- %s %s", parts[1], newCheckbox, parts[3])
			})
			break
		}
	}
	
	return strings.Join(lines, "\n")
}

// RenderTasksAsMarkdown converts tasks back to markdown format
func (p *TaskParser) RenderTasksAsMarkdown(tasks []models.Task, phases []Phase) string {
	var result strings.Builder
	
	// Group tasks by phase
	tasksByPhase := make(map[string][]models.Task)
	var unphased []models.Task
	
	for _, task := range tasks {
		if task.Phase != "" {
			tasksByPhase[task.Phase] = append(tasksByPhase[task.Phase], task)
		} else {
			unphased = append(unphased, task)
		}
	}
	
	// Render phases with their tasks
	for _, phase := range phases {
		result.WriteString(fmt.Sprintf("### %s\n", phase.Name))
		
		if phaseTasks, exists := tasksByPhase[phase.Name]; exists {
			for _, task := range phaseTasks {
				checkbox := models.TaskStatusToMarkdown(task.Status)
				result.WriteString(fmt.Sprintf("- %s %s\n", checkbox, task.Title))
			}
		}
		result.WriteString("\n")
	}
	
	// Render unphased tasks
	if len(unphased) > 0 {
		result.WriteString("### Tasks\n")
		for _, task := range unphased {
			checkbox := models.TaskStatusToMarkdown(task.Status)
			result.WriteString(fmt.Sprintf("- %s %s\n", checkbox, task.Title))
		}
		result.WriteString("\n")
	}
	
	return result.String()
}

// generateShortID creates a short unique identifier
func generateShortID() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}