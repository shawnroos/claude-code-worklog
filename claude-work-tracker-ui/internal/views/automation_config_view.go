package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"claude-work-tracker-ui/internal/automation"
)

// AutomationConfigView provides a Bubble Tea interface for configuring automation rules
type AutomationConfigView struct {
	config        *automation.TransitionConfig
	activityConfig *automation.ActivityConfig
	inputs        []textinput.Model
	focusIndex    int
	width         int
	height        int
	showHelp      bool
	errorMsg      string
	successMsg    string
}

// ConfigKeyMap defines key bindings for the config view
type ConfigKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Tab      key.Binding
	Toggle   key.Binding
	Save     key.Binding
	Cancel   key.Binding
	Help     key.Binding
}

// DefaultConfigKeyMap returns default key bindings
func DefaultConfigKeyMap() ConfigKeyMap {
	return ConfigKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next field"),
		),
		Toggle: key.NewBinding(
			key.WithKeys(" ", "enter"),
			key.WithHelp("space/enter", "toggle"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

// NewAutomationConfigView creates a new configuration view
func NewAutomationConfigView(transitionConfig *automation.TransitionConfig, activityConfig *automation.ActivityConfig) *AutomationConfigView {
	// Create text inputs for numeric values
	inputs := make([]textinput.Model, 5)
	
	// Stale threshold days
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "7"
	inputs[0].SetValue(fmt.Sprintf("%d", transitionConfig.StaleThresholdDays))
	inputs[0].CharLimit = 3
	inputs[0].Width = 10
	
	// Inactivity threshold hours
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "48"
	inputs[1].SetValue(fmt.Sprintf("%d", transitionConfig.InactivityThresholdHours))
	inputs[1].CharLimit = 4
	inputs[1].Width = 10
	
	// Auto-archive threshold days
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "90"
	inputs[2].SetValue(fmt.Sprintf("%d", transitionConfig.AutoArchiveThresholdDays))
	inputs[2].CharLimit = 4
	inputs[2].Width = 10
	
	// Focus threshold minutes
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "10"
	inputs[3].SetValue(fmt.Sprintf("%d", activityConfig.FocusThresholdMinutes))
	inputs[3].CharLimit = 3
	inputs[3].Width = 10
	
	// Focus min events
	inputs[4] = textinput.New()
	inputs[4].Placeholder = "3"
	inputs[4].SetValue(fmt.Sprintf("%d", activityConfig.FocusMinEvents))
	inputs[4].CharLimit = 2
	inputs[4].Width = 10
	
	// Focus on first input
	inputs[0].Focus()
	
	return &AutomationConfigView{
		config:         transitionConfig,
		activityConfig: activityConfig,
		inputs:        inputs,
		focusIndex:    0,
	}
}

// Init implements tea.Model
func (v *AutomationConfigView) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model
func (v *AutomationConfigView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil
		
	case tea.KeyMsg:
		keyMap := DefaultConfigKeyMap()
		
		switch {
		case key.Matches(msg, keyMap.Help):
			v.showHelp = !v.showHelp
			return v, nil
			
		case key.Matches(msg, keyMap.Save):
			return v, v.saveConfig()
			
		case key.Matches(msg, keyMap.Cancel):
			return v, tea.Quit
			
		case key.Matches(msg, keyMap.Up):
			v.focusPrevious()
			return v, nil
			
		case key.Matches(msg, keyMap.Down), key.Matches(msg, keyMap.Tab):
			v.focusNext()
			return v, nil
			
		case key.Matches(msg, keyMap.Toggle):
			return v, v.toggleBoolean()
		}
	}
	
	// Update focused input
	var cmd tea.Cmd
	if v.focusIndex < len(v.inputs) {
		v.inputs[v.focusIndex], cmd = v.inputs[v.focusIndex].Update(msg)
	}
	
	return v, cmd
}

// View implements tea.Model
func (v *AutomationConfigView) View() string {
	var s strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true).
		MarginBottom(1)
	
	s.WriteString(headerStyle.Render("Automation Configuration"))
	s.WriteString("\n\n")
	
	// Transition rules section
	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	
	s.WriteString(sectionStyle.Render("Transition Rules"))
	s.WriteString("\n\n")
	
	// Boolean options
	checkbox := func(enabled bool) string {
		if enabled {
			return "[▪]"
		}
		return "[ ]"
	}
	
	labelStyle := lipgloss.NewStyle().Width(40)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	
	// Automation enabled
	enabledLabel := "Enable automation"
	if v.focusIndex == 5 {
		enabledLabel = "> " + enabledLabel
	}
	s.WriteString(fmt.Sprintf("%s %s\n", checkbox(v.config.Enabled), labelStyle.Render(enabledLabel)))
	
	// Require confirmation for NOW
	confirmLabel := "Require confirmation for NOW transitions"
	if v.focusIndex == 6 {
		confirmLabel = "> " + confirmLabel
	}
	s.WriteString(fmt.Sprintf("%s %s\n\n", checkbox(v.config.RequireUserConfirmation), labelStyle.Render(confirmLabel)))
	
	// Numeric inputs
	inputLabel := func(label string, index int, value string) string {
		if v.focusIndex == index {
			label = "> " + label
		}
		return fmt.Sprintf("%-40s %s", labelStyle.Render(label), valueStyle.Render(value))
	}
	
	s.WriteString(inputLabel("Stale threshold (days):", 0, v.inputs[0].View()) + "\n")
	s.WriteString(inputLabel("Inactivity threshold (hours):", 1, v.inputs[1].View()) + "\n")
	s.WriteString(inputLabel("Auto-archive threshold (days):", 2, v.inputs[2].View()) + "\n\n")
	
	// Activity detection section
	s.WriteString(sectionStyle.Render("Activity Detection"))
	s.WriteString("\n\n")
	
	s.WriteString(inputLabel("Focus session threshold (minutes):", 3, v.inputs[3].View()) + "\n")
	s.WriteString(inputLabel("Minimum events for focus:", 4, v.inputs[4].View()) + "\n")
	
	// Auto-promote on focus
	promoteLabel := "Auto-promote to NOW on focus detection"
	if v.focusIndex == 7 {
		promoteLabel = "> " + promoteLabel
	}
	s.WriteString(fmt.Sprintf("%s %s\n\n", checkbox(v.activityConfig.AutoPromoteOnFocus), labelStyle.Render(promoteLabel)))
	
	// Messages
	if v.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		s.WriteString(errorStyle.Render("Error: " + v.errorMsg) + "\n\n")
	}
	
	if v.successMsg != "" {
		successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
		s.WriteString(successStyle.Render(v.successMsg) + "\n\n")
	}
	
	// Help
	if v.showHelp {
		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("Controls:"))
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("  ↑/↓/tab - Navigate fields"))
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("  space   - Toggle checkboxes"))
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("  ctrl+s  - Save configuration"))
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("  esc     - Cancel"))
		s.WriteString("\n")
		s.WriteString(helpStyle.Render("  ?       - Toggle help"))
		s.WriteString("\n")
	} else {
		hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		s.WriteString("\n")
		s.WriteString(hintStyle.Render("Press ? for help"))
	}
	
	return s.String()
}

// Helper methods

func (v *AutomationConfigView) focusNext() {
	v.focusIndex = (v.focusIndex + 1) % 8
	v.updateInputFocus()
}

func (v *AutomationConfigView) focusPrevious() {
	v.focusIndex--
	if v.focusIndex < 0 {
		v.focusIndex = 7
	}
	v.updateInputFocus()
}

func (v *AutomationConfigView) updateInputFocus() {
	for i := range v.inputs {
		if i == v.focusIndex {
			v.inputs[i].Focus()
		} else {
			v.inputs[i].Blur()
		}
	}
}

func (v *AutomationConfigView) toggleBoolean() tea.Cmd {
	switch v.focusIndex {
	case 5: // Enable automation
		v.config.Enabled = !v.config.Enabled
	case 6: // Require confirmation
		v.config.RequireUserConfirmation = !v.config.RequireUserConfirmation
	case 7: // Auto-promote on focus
		v.activityConfig.AutoPromoteOnFocus = !v.activityConfig.AutoPromoteOnFocus
	}
	return nil
}

func (v *AutomationConfigView) saveConfig() tea.Cmd {
	// Parse and validate numeric inputs
	var err error
	
	if val := v.inputs[0].Value(); val != "" {
		if _, err = fmt.Sscanf(val, "%d", &v.config.StaleThresholdDays); err != nil {
			v.errorMsg = "Invalid stale threshold"
			return nil
		}
	}
	
	if val := v.inputs[1].Value(); val != "" {
		if _, err = fmt.Sscanf(val, "%d", &v.config.InactivityThresholdHours); err != nil {
			v.errorMsg = "Invalid inactivity threshold"
			return nil
		}
	}
	
	if val := v.inputs[2].Value(); val != "" {
		if _, err = fmt.Sscanf(val, "%d", &v.config.AutoArchiveThresholdDays); err != nil {
			v.errorMsg = "Invalid auto-archive threshold"
			return nil
		}
	}
	
	if val := v.inputs[3].Value(); val != "" {
		if _, err = fmt.Sscanf(val, "%d", &v.activityConfig.FocusThresholdMinutes); err != nil {
			v.errorMsg = "Invalid focus threshold"
			return nil
		}
	}
	
	if val := v.inputs[4].Value(); val != "" {
		if _, err = fmt.Sscanf(val, "%d", &v.activityConfig.FocusMinEvents); err != nil {
			v.errorMsg = "Invalid focus min events"
			return nil
		}
	}
	
	v.errorMsg = ""
	v.successMsg = "Configuration saved successfully!"
	
	// In a real implementation, you would persist this configuration
	// For now, it's held in memory
	
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tea.Quit
	})
}