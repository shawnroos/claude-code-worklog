package views

import "claude-work-tracker-ui/internal/models"

// WorkDataProvider is an interface for providing work data
type WorkDataProvider interface {
	GetWorkBySchedule(schedule string) ([]*models.Work, error)
	UpdateWorkSchedule(workID, newSchedule string) error
	CompleteWork(workID string) error
	SearchWork(query string) ([]*models.Work, error)
}