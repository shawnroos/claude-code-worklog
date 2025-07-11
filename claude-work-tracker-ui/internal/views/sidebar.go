package views

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SidebarItem struct {
	Label       string
	Description string
	Icon        string
	Key         string
	ViewType    int
}

type SidebarModel struct {
	items        []SidebarItem
	selected     int
	width        int
	height       int
	activeView   int
}

var (
	sidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 1).
			Height(20)

	sidebarItemStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Margin(0, 0, 1, 0)

	sidebarSelectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("62")).
				Foreground(lipgloss.Color("230")).
				Bold(true).
				Padding(0, 1).
				Margin(0, 0, 1, 0)

	sidebarHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true).
				Padding(0, 0, 1, 0).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(lipgloss.Color("244"))

	sidebarHelpStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("244")).
				Margin(1, 0, 0, 0)
)

func NewSidebarModel() *SidebarModel {
	items := []SidebarItem{
		{
			Label:       "Dashboard",
			Description: "Overview & summary",
			Icon:        "📊",
			Key:         "d",
			ViewType:    0, // DashboardView
		},
		{
			Label:       "Work Items",
			Description: "Todos, plans & tasks",
			Icon:        "📋",
			Key:         "w",
			ViewType:    1, // WorkItemsView
		},
		{
			Label:       "References",
			Description: "Smart connections",
			Icon:        "🔗",
			Key:         "r",
			ViewType:    2, // ReferencesView
		},
		{
			Label:       "Future Work",
			Description: "Deferred items",
			Icon:        "🚀",
			Key:         "u",
			ViewType:    3, // FutureWorkView
		},
	}

	return &SidebarModel{
		items:    items,
		selected: 0,
	}
}

func (m *SidebarModel) Init() tea.Cmd {
	return nil
}

func (m *SidebarModel) Update(msg tea.Msg) (*SidebarModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = 25 // Fixed sidebar width
		m.height = msg.Height
		sidebarStyle = sidebarStyle.Height(msg.Height - 4)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
			if m.selected < len(m.items)-1 {
				m.selected++
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
			if m.selected > 0 {
				m.selected--
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			// Return the selected view type
			return m, tea.Cmd(func() tea.Msg {
				return SidebarSelectMsg{ViewType: m.items[m.selected].ViewType}
			})
		}
	}

	return m, nil
}

func (m *SidebarModel) View() string {
	var sections []string

	// Header
	header := sidebarHeaderStyle.Width(m.width - 4).Render("🗂️ Navigation")
	sections = append(sections, header)

	// Menu items
	for i, item := range m.items {
		var style lipgloss.Style
		if i == m.selected {
			style = sidebarSelectedStyle
		} else {
			style = sidebarItemStyle
		}

		// Format item
		itemText := item.Icon + " " + item.Label
		if i == m.selected {
			itemText += "\n  " + item.Description
		}

		renderedItem := style.Width(m.width - 4).Render(itemText)
		sections = append(sections, renderedItem)
	}

	// Help
	help := sidebarHelpStyle.Width(m.width - 4).Render(
		"↑/↓ Navigate\n" +
		"Enter Select\n" +
		"q Quit")
	sections = append(sections, help)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return sidebarStyle.Width(m.width).Render(content)
}

func (m *SidebarModel) SetActiveView(viewType int) {
	m.activeView = viewType
	// Find and select the corresponding sidebar item
	for i, item := range m.items {
		if item.ViewType == viewType {
			m.selected = i
			break
		}
	}
}

func (m *SidebarModel) GetSelectedView() int {
	if m.selected >= 0 && m.selected < len(m.items) {
		return m.items[m.selected].ViewType
	}
	return 0
}

// Message types
type SidebarSelectMsg struct {
	ViewType int
}