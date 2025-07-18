package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"claude-work-tracker-ui/internal/models"
)

// AutomationIndicators provides Unicode-based visual indicators for automation status
type AutomationIndicators struct {
	// Unicode symbols for different automation states
	AutoTransitioned   string
	PendingTransition  string
	BlockedTransition  string
	FocusMode         string
	InactivityWarning string
	ActivityHigh      string
	ActivityMedium    string
	ActivityLow       string
	GitLinked         string
	GitDirty          string
	
	// Styles for indicators
	AutoStyle     lipgloss.Style
	PendingStyle  lipgloss.Style
	WarningStyle  lipgloss.Style
	FocusStyle    lipgloss.Style
	GitStyle      lipgloss.Style
}

// DefaultAutomationIndicators returns indicators with Unicode characters
func DefaultAutomationIndicators() *AutomationIndicators {
	return &AutomationIndicators{
		// Unicode indicators (no emojis)
		AutoTransitioned:   "◉",  // Filled circle for automated transitions
		PendingTransition:  "◎",  // Double circle for pending
		BlockedTransition:  "⊘",  // Circle with slash for blocked
		FocusMode:         "▶",  // Triangle for active focus
		InactivityWarning: "⚠",  // Warning sign
		ActivityHigh:      "▰▰▰", // Activity bars
		ActivityMedium:    "▰▰□",
		ActivityLow:       "▰□□",
		GitLinked:         "⎇",  // Branch symbol
		GitDirty:          "±",  // Plus-minus for uncommitted changes
		
		// Styles
		AutoStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("42")),  // Green
		PendingStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("220")), // Yellow
		WarningStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("208")), // Orange
		FocusStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("170")), // Purple
		GitStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("39")),  // Blue
	}
}

// GetWorkItemIndicators returns formatted indicators for a work item
func (ai *AutomationIndicators) GetWorkItemIndicators(work *models.Work) string {
	var indicators []string
	
	// Check for auto-transition - would need auto_transitioned field in WorkMetadata
	if false { // Disabled for now
		indicators = append(indicators, ai.AutoStyle.Render(ai.AutoTransitioned))
	}
	
	// Check for pending transition - would need pending_transition field in WorkMetadata
	if false { // Disabled for now
		indicators = append(indicators, ai.PendingStyle.Render(ai.PendingTransition))
	}
	
	// Check for blocked status
	if work.Metadata.Status == models.WorkStatusBlocked {
		indicators = append(indicators, ai.WarningStyle.Render(ai.BlockedTransition))
	}
	
	// Check for focus mode - would need focus_session field in WorkMetadata
	if false { // Disabled for now
		indicators = append(indicators, ai.FocusStyle.Render(ai.FocusMode))
	}
	
	// Check for inactivity warning - would need Warnings field in WorkMetadata
	if false { // Disabled for now
		indicators = append(indicators, ai.WarningStyle.Render(ai.InactivityWarning))
	}
	
	// Activity level indicator
	if activityScore := work.Metadata.ActivityScore; activityScore > 0 {
		var activityIndicator string
		switch {
		case activityScore >= 10:
			activityIndicator = ai.ActivityHigh
		case activityScore >= 5:
			activityIndicator = ai.ActivityMedium
		default:
			activityIndicator = ai.ActivityLow
		}
		indicators = append(indicators, lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(activityIndicator))
	}
	
	// Git status
	if work.GitContext.Branch != "" {
		indicators = append(indicators, ai.GitStyle.Render(ai.GitLinked))
	}
	
	if len(indicators) > 0 {
		return strings.Join(indicators, " ")
	}
	return ""
}

// GetTransitionTooltip returns a description of pending transitions
func (ai *AutomationIndicators) GetTransitionTooltip(work *models.Work) string {
	var tooltips []string
	
	// Would need pending_transition field in WorkMetadata
	if false { // Disabled for now
		tooltips = append(tooltips, "Pending transition")
	}
	
	// Would need auto_transitioned field in WorkMetadata
	if false { // Disabled for now
		tooltips = append(tooltips, "Auto transition")
	}
	
	return strings.Join(tooltips, " | ")
}

// RenderIndicatorLegend returns a legend explaining the indicators
func (ai *AutomationIndicators) RenderIndicatorLegend() string {
	legendStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	items := []string{
		fmt.Sprintf("%s Auto-transitioned", ai.AutoStyle.Render(ai.AutoTransitioned)),
		fmt.Sprintf("%s Pending transition", ai.PendingStyle.Render(ai.PendingTransition)),
		fmt.Sprintf("%s Blocked", ai.WarningStyle.Render(ai.BlockedTransition)),
		fmt.Sprintf("%s Focus mode", ai.FocusStyle.Render(ai.FocusMode)),
		fmt.Sprintf("%s Inactive", ai.WarningStyle.Render(ai.InactivityWarning)),
		fmt.Sprintf("%s Git linked", ai.GitStyle.Render(ai.GitLinked)),
	}
	return legendStyle.Render("Indicators: " + strings.Join(items, "  "))
}