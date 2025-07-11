package themes

import "github.com/charmbracelet/lipgloss"

// Color schemes for the TUI
type ColorScheme struct {
	Primary     string
	Secondary   string
	Accent      string
	Background  string
	Text        string
	Muted       string
	Success     string
	Warning     string
	Error       string
	Border      string
}

var (
	// Default theme (current)
	DefaultTheme = ColorScheme{
		Primary:    "86",  // Cyan
		Secondary:  "62",  // Green
		Accent:     "170", // Purple
		Background: "235", // Dark gray
		Text:       "255", // White
		Muted:      "244", // Light gray
		Success:    "46",  // Green
		Warning:    "208", // Orange
		Error:      "196", // Red
		Border:     "62",  // Green
	}

	// Dark theme
	DarkTheme = ColorScheme{
		Primary:    "39",  // Blue
		Secondary:  "33",  // Darker blue
		Accent:     "207", // Pink
		Background: "236", // Very dark gray
		Text:       "252", // Light gray
		Muted:      "240", // Medium gray
		Success:    "82",  // Bright green
		Warning:    "214", // Yellow
		Error:      "203", // Light red
		Border:     "240", // Medium gray
	}

	// Light theme
	LightTheme = ColorScheme{
		Primary:    "21",  // Dark blue
		Secondary:  "18",  // Navy
		Accent:     "162", // Purple
		Background: "255", // White
		Text:       "16",  // Black
		Muted:      "244", // Gray
		Success:    "22",  // Dark green
		Warning:    "172", // Dark orange
		Error:      "124", // Dark red
		Border:     "244", // Gray
	}

	// Retro theme
	RetroTheme = ColorScheme{
		Primary:    "214", // Yellow
		Secondary:  "208", // Orange
		Accent:     "196", // Red
		Background: "16",  // Black
		Text:       "46",  // Green
		Muted:      "244", // Gray
		Success:    "46",  // Green
		Warning:    "214", // Yellow
		Error:      "196", // Red
		Border:     "214", // Yellow
	}
)

// Helper functions to create styles with theme
func (c ColorScheme) HeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.Primary)).
		Bold(true).
		Padding(0, 1)
}

func (c ColorScheme) SectionStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(c.Border)).
		Padding(1).
		Margin(0, 1, 1, 0)
}

func (c ColorScheme) ItemStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.Text)).
		Padding(0, 1)
}

func (c ColorScheme) AccentStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.Accent)).
		Bold(true).
		Padding(0, 1)
}

func (c ColorScheme) SuccessStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.Success)).
		Bold(true)
}

func (c ColorScheme) WarningStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.Warning)).
		Bold(true)
}

func (c ColorScheme) ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.Error)).
		Bold(true)
}