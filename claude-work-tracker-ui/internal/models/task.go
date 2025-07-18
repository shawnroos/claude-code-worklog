package models

import (
	"time"
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusTodo        TaskStatus = "todo"        // [ ]
	TaskStatusInProgress  TaskStatus = "in_progress" // […]
	TaskStatusCompleted   TaskStatus = "completed"   // [x]
	TaskStatusBlocked     TaskStatus = "blocked"     // [!]
	TaskStatusCancelled   TaskStatus = "cancelled"   // [-]
)

// Task represents a high-level task within a Work item
type Task struct {
	// Core identification
	ID       string     `yaml:"id" json:"id"`
	Title    string     `yaml:"title" json:"title"`
	Status   TaskStatus `yaml:"status" json:"status"`
	
	// Organization
	Phase    string `yaml:"phase,omitempty" json:"phase,omitempty"`       // Phase or category
	Category string `yaml:"category,omitempty" json:"category,omitempty"` // Alternative grouping
	
	// Source tracking
	Source     string `yaml:"source" json:"source"`           // artifact_id or "manual"
	LineNumber int    `yaml:"line_number" json:"line_number"` // Line in markdown content
	
	// Timestamps
	CreatedAt   time.Time  `yaml:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `yaml:"updated_at" json:"updated_at"`
	CompletedAt *time.Time `yaml:"completed_at,omitempty" json:"completed_at,omitempty"`
	
	// Optional metadata
	AssignedTo   string   `yaml:"assigned_to,omitempty" json:"assigned_to,omitempty"`
	Dependencies []string `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	Tags         []string `yaml:"tags,omitempty" json:"tags,omitempty"`
	Notes        string   `yaml:"notes,omitempty" json:"notes,omitempty"`
}

// Update represents a progress update on a Work item
type Update struct {
	// Core identification
	ID        string    `yaml:"id" json:"id"`
	WorkID    string    `yaml:"work_id" json:"work_id"`
	Timestamp time.Time `yaml:"timestamp" json:"timestamp"`
	
	// Content
	Title   string `yaml:"title" json:"title"`
	Summary string `yaml:"summary" json:"summary"`
	
	// Context
	Author     string `yaml:"author" json:"author"`           // "Claude" or user name
	SessionID  string `yaml:"session_id,omitempty" json:"session_id,omitempty"`
	UpdateType string `yaml:"update_type" json:"update_type"` // "automatic", "manual"
	
	// Status changes
	TasksCompleted []string `yaml:"tasks_completed,omitempty" json:"tasks_completed,omitempty"`
	TasksAdded     []string `yaml:"tasks_added,omitempty" json:"tasks_added,omitempty"`
	
	// Progress indicators
	ProgressBefore int `yaml:"progress_before,omitempty" json:"progress_before,omitempty"`
	ProgressAfter  int `yaml:"progress_after,omitempty" json:"progress_after,omitempty"`
}

// TaskStatusFromMarkdown converts markdown checkbox syntax to TaskStatus
func TaskStatusFromMarkdown(checkbox string) TaskStatus {
	switch checkbox {
	case "[ ]":
		return TaskStatusTodo
	case "[…]", "[...]":
		return TaskStatusInProgress
	case "[x]", "[X]", "[✓]":
		return TaskStatusCompleted
	case "[!]", "[⚠]", "[⚠️]":
		return TaskStatusBlocked
	case "[-]":
		return TaskStatusCancelled
	default:
		return TaskStatusTodo
	}
}

// TaskStatusToMarkdown converts TaskStatus to markdown checkbox syntax
func TaskStatusToMarkdown(status TaskStatus) string {
	switch status {
	case TaskStatusTodo:
		return "[ ]"
	case TaskStatusInProgress:
		return "[…]"
	case TaskStatusCompleted:
		return "[x]"
	case TaskStatusBlocked:
		return "[!]"
	case TaskStatusCancelled:
		return "[-]"
	default:
		return "[ ]"
	}
}

// IsCompleted returns true if the task is completed
func (t *Task) IsCompleted() bool {
	return t.Status == TaskStatusCompleted
}

// IsActive returns true if the task is in progress or todo
func (t *Task) IsActive() bool {
	return t.Status == TaskStatusTodo || t.Status == TaskStatusInProgress
}

// IsBlocked returns true if the task is blocked
func (t *Task) IsBlocked() bool {
	return t.Status == TaskStatusBlocked
}

// MarkAsCompleted marks the task as completed with timestamp
func (t *Task) MarkAsCompleted() {
	t.Status = TaskStatusCompleted
	now := time.Now()
	t.UpdatedAt = now
	t.CompletedAt = &now
}

// MarkAsBlocked marks the task as blocked
func (t *Task) MarkAsBlocked() {
	t.Status = TaskStatusBlocked
	t.UpdatedAt = time.Now()
}

// MarkAsInProgress marks the task as in progress
func (t *Task) MarkAsInProgress() {
	t.Status = TaskStatusInProgress
	t.UpdatedAt = time.Now()
}

// GetDisplayStatus returns a human-readable status
func (t *Task) GetDisplayStatus() string {
	switch t.Status {
	case TaskStatusTodo:
		return "To Do"
	case TaskStatusInProgress:
		return "In Progress"
	case TaskStatusCompleted:
		return "Completed"
	case TaskStatusBlocked:
		return "Blocked"
	case TaskStatusCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// GetStatusIcon returns an icon for the task status
func (t *Task) GetStatusIcon() string {
	switch t.Status {
	case TaskStatusTodo:
		return "○"
	case TaskStatusInProgress:
		return "◐"
	case TaskStatusCompleted:
		return "●"
	case TaskStatusBlocked:
		return "⚠"
	case TaskStatusCancelled:
		return "⊘"
	default:
		return "○"
	}
}