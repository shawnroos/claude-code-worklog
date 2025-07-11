package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"claude-work-tracker-ui/internal/data"
	"claude-work-tracker-ui/internal/models"
)

var (
	dashboardStyle = lipgloss.NewStyle().
			Padding(1, 2)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Padding(0, 1)

	sectionStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1).
			Margin(0, 1, 1, 0)

	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(0, 1)

	activeItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("27")).
			Bold(true)

	priorityHighStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true)

	priorityMediumStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("208")).
				Bold(true)

	priorityLowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Bold(true)
)

type DashboardModel struct {
	dataClient   *data.Client
	workState    *models.WorkState
	width        int
	height       int
	loading      bool
	error        error
	lastRefresh  time.Time
}

type DashboardLoadedMsg struct {
	WorkState *models.WorkState
	Error     error
}

func NewDashboardModel(dataClient *data.Client) *DashboardModel {
	return &DashboardModel{
		dataClient:  dataClient,
		loading:     true,
		lastRefresh: time.Now(),
	}
}

func (m *DashboardModel) Init() tea.Cmd {
	return m.loadWorkState()
}

func (m *DashboardModel) Update(msg tea.Msg) (*DashboardModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case DashboardLoadedMsg:
		m.loading = false
		m.workState = msg.WorkState
		m.error = msg.Error
		m.lastRefresh = time.Now()
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("r", "R"))):
			m.loading = true
			return m, m.loadWorkState()
		}
	}

	return m, nil
}

func (m *DashboardModel) View() string {
	if m.loading {
		return "Loading dashboard..."
	}

	if m.error != nil {
		return fmt.Sprintf("Error loading dashboard: %s", m.error.Error())
	}

	if m.workState == nil {
		return "No work state available"
	}

	var sections []string

	// Header
	header := headerStyle.Render("📊 Current Directory Work")
	sections = append(sections, header)

	// Session Info
	sessionInfo := m.renderSessionInfo()
	sections = append(sections, sessionInfo)

	// Active Todos
	activeTodos := m.renderActiveTodos()
	sections = append(sections, activeTodos)

	// Recent Findings
	recentFindings := m.renderRecentFindings()
	sections = append(sections, recentFindings)

	// Help
	help := m.renderHelp()
	sections = append(sections, help)

	return strings.Join(sections, "\n")
}

func (m *DashboardModel) renderSessionInfo() string {
	// Get current directory info
	dirInfo := m.dataClient.GetCurrentDirectoryInfo()
	workingDir := dirInfo["working_directory"].(string)
	projectRoot := dirInfo["project_root"].(string)
	workDirCount := dirInfo["work_dir_count"].(int)
	
	// Shorten paths for display
	if len(workingDir) > 35 {
		parts := strings.Split(workingDir, "/")
		if len(parts) > 2 {
			workingDir = ".../" + strings.Join(parts[len(parts)-2:], "/")
		}
	}
	
	if len(projectRoot) > 35 {
		parts := strings.Split(projectRoot, "/")
		if len(parts) > 2 {
			projectRoot = ".../" + strings.Join(parts[len(parts)-2:], "/")
		}
	}
	
	content := []string{
		fmt.Sprintf("Current: %s", workingDir),
		fmt.Sprintf("Project: %s", projectRoot),
		fmt.Sprintf("Session: %s", m.workState.CurrentSession),
		fmt.Sprintf("Last Refresh: %s", m.lastRefresh.Format("15:04:05")),
	}
	
	hasWorkTracking, _ := dirInfo["has_work_tracking"].(bool)
	if hasWorkTracking {
		if workDirCount > 1 {
			content = append(content, fmt.Sprintf("📁 %d work directories found", workDirCount))
		} else {
			content = append(content, "📁 Work tracking active")
		}
	} else {
		content = append(content, "📁 No work tracking found")
	}

	return sectionStyle.Width(55).Render(
		headerStyle.Render("Project Context") + "\n" +
			strings.Join(content, "\n"),
	)
}

func (m *DashboardModel) renderActiveTodos() string {
	todos := m.workState.ActiveTodos
	
	if len(todos) == 0 {
		return sectionStyle.Width(60).Render(
			headerStyle.Render("Active Todos (0)") + "\n" +
				itemStyle.Render("No active todos"),
		)
	}

	var items []string
	for i, todo := range todos {
		if i >= 5 { // Show only first 5
			items = append(items, itemStyle.Render(fmt.Sprintf("... and %d more", len(todos)-5)))
			break
		}

		status := m.getStatusSymbol(todo.Status)
		priority := m.getPriorityString(todo.GetPriority())
		content := todo.Content
		if len(content) > 50 {
			content = content[:47] + "..."
		}

		item := fmt.Sprintf("%s %s %s", status, priority, content)
		items = append(items, itemStyle.Render(item))
	}

	return sectionStyle.Width(60).Render(
		headerStyle.Render(fmt.Sprintf("Active Todos (%d)", len(todos))) + "\n" +
			strings.Join(items, "\n"),
	)
}

func (m *DashboardModel) renderRecentFindings() string {
	findings := m.workState.RecentFindings
	
	if len(findings) == 0 {
		return sectionStyle.Width(60).Render(
			headerStyle.Render("Recent Findings (0)") + "\n" +
				itemStyle.Render("No recent findings"),
		)
	}

	var items []string
	for i, finding := range findings {
		if i >= 5 { // Show only first 5
			items = append(items, itemStyle.Render(fmt.Sprintf("... and %d more", len(findings)-5)))
			break
		}

		typeStr := strings.Title(finding.Type)
		content := finding.Content
		if len(content) > 40 {
			content = content[:37] + "..."
		}

		item := fmt.Sprintf("[%s] %s", typeStr, content)
		items = append(items, itemStyle.Render(item))
	}

	return sectionStyle.Width(60).Render(
		headerStyle.Render(fmt.Sprintf("Recent Findings (%d)", len(findings))) + "\n" +
			strings.Join(items, "\n"),
	)
}

func (m *DashboardModel) renderHelp() string {
	help := []string{
		"r/R: Refresh dashboard",
		"Use sidebar to navigate",
	}

	return itemStyle.Render(strings.Join(help, " • "))
}

func (m *DashboardModel) getStatusSymbol(status string) string {
	switch status {
	case "completed":
		return "✓"
	case "in_progress":
		return "○"
	case "pending":
		return "●"
	default:
		return "?"
	}
}

func (m *DashboardModel) getPriorityString(priority string) string {
	switch priority {
	case "high":
		return priorityHighStyle.Render("[H]")
	case "medium":
		return priorityMediumStyle.Render("[M]")
	case "low":
		return priorityLowStyle.Render("[L]")
	default:
		return "[?]"
	}
}

func (m *DashboardModel) loadWorkState() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		workState, err := m.dataClient.GetCurrentWorkState()
		return DashboardLoadedMsg{
			WorkState: workState,
			Error:     err,
		}
	})
}

// Helper methods for the app to access dashboard state
func (m *DashboardModel) HasWorkItems() bool {
	return m.workState != nil && len(m.workState.ActiveTodos) > 0
}

func (m *DashboardModel) HasFindings() bool {
	return m.workState != nil && len(m.workState.RecentFindings) > 0
}