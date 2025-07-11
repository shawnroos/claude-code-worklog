package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"claude-work-tracker-ui/internal/data"
	"claude-work-tracker-ui/internal/models"
)

type WorkItemsFilterType int

const (
	FilterAll WorkItemsFilterType = iota
	FilterTodos
	FilterPlans
	FilterProposals
	FilterFindings
	FilterReports
	FilterSummaries
)

type WorkItemsModel struct {
	dataClient     *data.Client
	workItems      []models.WorkItem
	filteredItems  []models.WorkItem
	list           list.Model
	filter         WorkItemsFilterType
	width          int
	height         int
	loading        bool
	error          error
	selectedItem   *models.WorkItem
	showingDetails bool
}

type WorkItemsLoadedMsg struct {
	WorkItems []models.WorkItem
	Error     error
}

type workItemListItem struct {
	item models.WorkItem
}

func (i workItemListItem) Title() string {
	return i.item.Content
}

func (i workItemListItem) Description() string {
	return fmt.Sprintf("%s • %s • %s", 
		i.item.GetDisplayType(),
		i.item.GetDisplayStatus(),
		i.item.GetPriority())
}

func (i workItemListItem) FilterValue() string {
	return i.item.Content + " " + i.item.Type + " " + i.item.Status
}

func NewWorkItemsModel(dataClient *data.Client) *WorkItemsModel {
	// Create list with custom styles
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("170")).
		Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("244"))

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Work Items"
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(true)
	l.SetShowPagination(true)
	l.SetShowHelp(true)

	return &WorkItemsModel{
		dataClient: dataClient,
		list:       l,
		filter:     FilterAll,
		loading:    true,
	}
}

func (m *WorkItemsModel) Init() tea.Cmd {
	return m.loadWorkItems()
}

func (m *WorkItemsModel) Update(msg tea.Msg) (*WorkItemsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width - 4)
		m.list.SetHeight(msg.Height - 8)
		return m, nil

	case WorkItemsLoadedMsg:
		m.loading = false
		m.workItems = msg.WorkItems
		m.error = msg.Error
		
		if m.error == nil {
			m.applyFilter()
		}
		
		return m, nil

	case tea.KeyMsg:
		if m.showingDetails {
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q"))):
				m.showingDetails = false
				return m, nil
			}
			return m, nil
		}

		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if item, ok := m.list.SelectedItem().(workItemListItem); ok {
				m.selectedItem = &item.item
				m.showingDetails = true
				return m, nil
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("r", "R"))):
			m.loading = true
			return m, m.loadWorkItems()

		case key.Matches(msg, key.NewBinding(key.WithKeys("1"))):
			m.filter = FilterAll
			m.applyFilter()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("2"))):
			m.filter = FilterTodos
			m.applyFilter()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("3"))):
			m.filter = FilterPlans
			m.applyFilter()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("4"))):
			m.filter = FilterProposals
			m.applyFilter()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("5"))):
			m.filter = FilterFindings
			m.applyFilter()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("6"))):
			m.filter = FilterReports
			m.applyFilter()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("7"))):
			m.filter = FilterSummaries
			m.applyFilter()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *WorkItemsModel) View() string {
	if m.loading {
		return "Loading work items..."
	}

	if m.error != nil {
		return fmt.Sprintf("Error loading work items: %s", m.error.Error())
	}

	if m.showingDetails && m.selectedItem != nil {
		return m.renderWorkItemDetails()
	}

	// Filter tabs
	filterTabs := m.renderFilterTabs()

	// Help text
	helpText := m.renderHelpText()

	return filterTabs + "\n" + m.list.View() + "\n" + helpText
}

func (m *WorkItemsModel) renderFilterTabs() string {
	filters := []struct {
		name   string
		filter WorkItemsFilterType
	}{
		{"All", FilterAll},
		{"Todos", FilterTodos},
		{"Plans", FilterPlans},
		{"Proposals", FilterProposals},
		{"Findings", FilterFindings},
		{"Reports", FilterReports},
		{"Summaries", FilterSummaries},
	}

	var tabs []string
	for i, f := range filters {
		style := lipgloss.NewStyle().
			Padding(0, 1).
			Margin(0, 1)

		if f.filter == m.filter {
			style = style.
				Background(lipgloss.Color("62")).
				Foreground(lipgloss.Color("230")).
				Bold(true)
		} else {
			style = style.
				Foreground(lipgloss.Color("244"))
		}

		count := m.getFilteredCount(f.filter)
		tab := fmt.Sprintf("%d:%s(%d)", i+1, f.name, count)
		tabs = append(tabs, style.Render(tab))
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, tabs...)
}

func (m *WorkItemsModel) renderHelpText() string {
	help := []string{
		"↑/↓: Navigate",
		"Enter: View details",
		"r: Refresh",
		"1-7: Filter",
		"/: Search",
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Render(strings.Join(help, " • "))
}

func (m *WorkItemsModel) renderWorkItemDetails() string {
	item := m.selectedItem
	if item == nil {
		return "No item selected"
	}

	var sections []string

	// Header
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true).
		Render(fmt.Sprintf("📋 %s Details", item.GetDisplayType()))
	sections = append(sections, header)

	// Basic info
	basicInfo := []string{
		fmt.Sprintf("ID: %s", item.ID),
		fmt.Sprintf("Type: %s", item.GetDisplayType()),
		fmt.Sprintf("Status: %s", item.GetDisplayStatus()),
		fmt.Sprintf("Priority: %s", item.GetPriority()),
		fmt.Sprintf("Created: %s", item.GetTimestamp().Format("2006-01-02 15:04:05")),
	}

	basicSection := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1).
		Render(strings.Join(basicInfo, "\n"))
	sections = append(sections, basicSection)

	// Content
	contentSection := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1).
		Render(fmt.Sprintf("Content:\n%s", item.Content))
	sections = append(sections, contentSection)

	// Context info
	if item.Context.Branch != "" {
		contextInfo := []string{
			fmt.Sprintf("Branch: %s", item.Context.Branch),
			fmt.Sprintf("Worktree: %s", item.Context.Worktree),
			fmt.Sprintf("Directory: %s", item.Context.WorkingDirectory),
		}

		contextSection := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1).
			Render(fmt.Sprintf("Git Context:\n%s", strings.Join(contextInfo, "\n")))
		sections = append(sections, contextSection)
	}

	// Smart references
	if item.HasSmartReferences() {
		refInfo := []string{
			fmt.Sprintf("Smart References: %d", item.GetSmartReferenceCount()),
		}

		for _, ref := range item.Metadata.SmartReferences {
			refInfo = append(refInfo, fmt.Sprintf("• %s (%.2f, %s)", 
				ref.TargetID, ref.SimilarityScore, ref.RelationshipType))
		}

		refSection := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1).
			Render(strings.Join(refInfo, "\n"))
		sections = append(sections, refSection)
	}

	// Help
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Render("Press ESC or q to go back")
	sections = append(sections, help)

	return strings.Join(sections, "\n\n")
}

func (m *WorkItemsModel) getFilteredCount(filter WorkItemsFilterType) int {
	if len(m.workItems) == 0 {
		return 0
	}

	count := 0
	for _, item := range m.workItems {
		if m.matchesFilter(item, filter) {
			count++
		}
	}
	return count
}

func (m *WorkItemsModel) matchesFilter(item models.WorkItem, filter WorkItemsFilterType) bool {
	switch filter {
	case FilterAll:
		return true
	case FilterTodos:
		return item.Type == "todo"
	case FilterPlans:
		return item.Type == "plan"
	case FilterProposals:
		return item.Type == "proposal"
	case FilterFindings:
		return item.Type == "finding"
	case FilterReports:
		return item.Type == "report"
	case FilterSummaries:
		return item.Type == "summary"
	default:
		return true
	}
}

func (m *WorkItemsModel) applyFilter() {
	m.filteredItems = []models.WorkItem{}
	
	for _, item := range m.workItems {
		if m.matchesFilter(item, m.filter) {
			m.filteredItems = append(m.filteredItems, item)
		}
	}

	// Convert to list items
	var listItems []list.Item
	for _, item := range m.filteredItems {
		listItems = append(listItems, workItemListItem{item: item})
	}

	// Update list title
	filterName := m.getFilterName(m.filter)
	m.list.Title = fmt.Sprintf("Work Items - %s (%d)", filterName, len(m.filteredItems))
	
	// Set items
	m.list.SetItems(listItems)
}

func (m *WorkItemsModel) getFilterName(filter WorkItemsFilterType) string {
	switch filter {
	case FilterAll:
		return "All"
	case FilterTodos:
		return "Todos"
	case FilterPlans:
		return "Plans"
	case FilterProposals:
		return "Proposals"
	case FilterFindings:
		return "Findings"
	case FilterReports:
		return "Reports"
	case FilterSummaries:
		return "Summaries"
	default:
		return "Unknown"
	}
}

func (m *WorkItemsModel) loadWorkItems() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		workItems, err := m.dataClient.GetAllWorkItems()
		return WorkItemsLoadedMsg{
			WorkItems: workItems,
			Error:     err,
		}
	})
}