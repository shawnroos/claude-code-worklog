package views

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"claude-work-tracker-ui/internal/automation"
	"claude-work-tracker-ui/internal/models"
)

// AutomationAction represents an available automation action
type AutomationAction struct {
	ID          string
	Title       string
	Description string
	Icon        string // Unicode icon
	Type        ActionType
	Target      string // Target status or schedule
}

// ActionType defines the type of automation action
type ActionType int

const (
	ActionTransition ActionType = iota
	ActionConfirmPending
	ActionRejectPending
	ActionClearAutomation
	ActionRunRules
	ActionToggleFocus
)

// AutomationActionsMenu provides action menu items for automation
type AutomationActionsMenu struct {
	work          *models.Work
	transitions   []automation.TransitionSuggestion
	actions       []AutomationAction
	selectedIndex int
}

// NewAutomationActionsMenu creates menu actions for a work item
func NewAutomationActionsMenu(work *models.Work, transitions []automation.TransitionSuggestion) *AutomationActionsMenu {
	menu := &AutomationActionsMenu{
		work:        work,
		transitions: transitions,
	}
	
	// Build available actions
	menu.buildActions()
	return menu
}

// buildActions populates available actions based on work state
func (m *AutomationActionsMenu) buildActions() {
	m.actions = []AutomationAction{}
	
	// Check for pending transitions
	// Would need pending_transition field in WorkMetadata
	if false { // Disabled for now
		// Would need transition_reason field in WorkMetadata
		// reason := "Manual confirmation required"
		// if r, ok := m.work.Metadata.Metadata["transition_reason"].(string); ok {
		//     reason = r
		// }
		
		// Would create pending transition actions here
		// m.actions = append(m.actions, AutomationAction{
		//     ID:          "confirm_pending",
		//     Title:       fmt.Sprintf("Confirm: Move to %s", strings.ToUpper(pending)),
		//     Description: reason,
		//     Icon:        "✓", // Checkmark
		//     Type:        ActionConfirmPending,
		//     Target:      pending,
		// })
		
		// m.actions = append(m.actions, AutomationAction{
		//     ID:          "reject_pending",
		//     Title:       "Reject pending transition",
		//     Description: "Keep current state",
		//     Icon:        "✗", // X mark
		//     Type:        ActionRejectPending,
		// })
	}
	
	// Add suggested transitions
	for i, suggestion := range m.transitions {
		if !suggestion.AutoApply {
			var icon string
			switch suggestion.Type {
			case "schedule_change":
				icon = "⇄" // Arrows
			case "status_change":
				icon = "▶" // Triangle
			default:
				icon = "→" // Arrow
			}
			
			title := fmt.Sprintf("%s to %s", strings.Title(suggestion.Type), suggestion.Target)
			m.actions = append(m.actions, AutomationAction{
				ID:          fmt.Sprintf("transition_%d", i),
				Title:       title,
				Description: suggestion.Reason,
				Icon:        icon,
				Type:        ActionTransition,
				Target:      suggestion.Target,
			})
		}
	}
	
	// Manual status changes
	if m.work.Metadata.Status != models.WorkStatusCompleted {
		m.actions = append(m.actions, AutomationAction{
			ID:          "complete",
			Title:       "Complete work item",
			Description: "Mark as completed and move to CLOSED",
			Icon:        "✓", // Checkmark
			Type:        ActionTransition,
			Target:      "completed",
		})
	}
	
	if m.work.Metadata.Status != models.WorkStatusCanceled {
		m.actions = append(m.actions, AutomationAction{
			ID:          "cancel",
			Title:       "Cancel work item",
			Description: "Mark as canceled and move to CLOSED",
			Icon:        "✗", // X mark
			Type:        ActionTransition,
			Target:      "canceled",
		})
	}
	
	// Schedule changes
	if m.work.Schedule != models.ScheduleNow {
		m.actions = append(m.actions, AutomationAction{
			ID:          "move_now",
			Title:       "Move to NOW",
			Description: "Start working on this item",
			Icon:        "▶", // Play triangle
			Type:        ActionTransition,
			Target:      "now",
		})
	}
	
	if m.work.Schedule == models.ScheduleNow {
		m.actions = append(m.actions, AutomationAction{
			ID:          "move_next",
			Title:       "Move to NEXT",
			Description: "Defer to next session",
			Icon:        "⇨", // Right arrow
			Type:        ActionTransition,
			Target:      "next",
		})
	}
	
	// Automation management - would need auto_transitioned field in WorkMetadata
	if false { // Disabled for now
		m.actions = append(m.actions, AutomationAction{
			ID:          "clear_auto",
			Title:       "Clear automation flags",
			Description: "Remove auto-transition markers",
			Icon:        "⊙", // Circle with line
			Type:        ActionClearAutomation,
		})
	}
	
	// Always available
	m.actions = append(m.actions, AutomationAction{
		ID:          "run_rules",
		Title:       "Run automation rules",
		Description: "Check for applicable transitions",
		Icon:        "⚙", // Gear
		Type:        ActionRunRules,
	})
}

// GetActions returns the list of available actions
func (m *AutomationActionsMenu) GetActions() []AutomationAction {
	return m.actions
}

// RenderActionsList renders actions as a Bubble Tea list
func (m *AutomationActionsMenu) RenderActionsList() string {
	var items []list.Item
	for _, action := range m.actions {
		items = append(items, actionItem{action})
	}
	
	l := list.New(items, actionDelegate{}, 50, len(items)*3)
	l.Title = "Automation Actions"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	
	return l.View()
}

// RenderCompactMenu renders a compact action menu
func (m *AutomationActionsMenu) RenderCompactMenu() string {
	var s strings.Builder
	
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true).
		MarginBottom(1)
	
	s.WriteString(titleStyle.Render("Actions"))
	s.WriteString("\n\n")
	
	for i, action := range m.actions {
		var itemStyle lipgloss.Style
		if i == m.selectedIndex {
			itemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Bold(true)
		} else {
			itemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("246"))
		}
		
		iconStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Width(3)
		
		s.WriteString(iconStyle.Render(action.Icon))
		s.WriteString(itemStyle.Render(action.Title))
		
		if action.Description != "" && i == m.selectedIndex {
			descStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				MarginLeft(3)
			s.WriteString("\n")
			s.WriteString(descStyle.Render(action.Description))
		}
		
		s.WriteString("\n")
	}
	
	return s.String()
}

// SelectNext moves to the next action
func (m *AutomationActionsMenu) SelectNext() {
	m.selectedIndex = (m.selectedIndex + 1) % len(m.actions)
}

// SelectPrevious moves to the previous action
func (m *AutomationActionsMenu) SelectPrevious() {
	m.selectedIndex--
	if m.selectedIndex < 0 {
		m.selectedIndex = len(m.actions) - 1
	}
}

// GetSelectedAction returns the currently selected action
func (m *AutomationActionsMenu) GetSelectedAction() *AutomationAction {
	if m.selectedIndex >= 0 && m.selectedIndex < len(m.actions) {
		return &m.actions[m.selectedIndex]
	}
	return nil
}

// actionItem implements list.Item for actions
type actionItem struct {
	action AutomationAction
}

func (i actionItem) FilterValue() string { return i.action.Title }

// actionDelegate implements list.ItemDelegate for actions
type actionDelegate struct{}

func (d actionDelegate) Height() int                             { return 2 }
func (d actionDelegate) Spacing() int                            { return 1 }
func (d actionDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d actionDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(actionItem)
	if !ok {
		return
	}
	
	action := i.action
	isSelected := index == m.Index()
	
	var titleStyle lipgloss.Style
	if isSelected {
		titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true)
	} else {
		titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("246"))
	}
	
	iconStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Width(3)
	
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		MarginLeft(3)
	
	fmt.Fprintf(w, "%s%s\n", iconStyle.Render(action.Icon), titleStyle.Render(action.Title))
	if action.Description != "" {
		fmt.Fprintf(w, "%s\n", descStyle.Render(action.Description))
	}
}