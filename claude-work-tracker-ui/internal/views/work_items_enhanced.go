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

// EnhancedWorkItemsView represents the view mode
type EnhancedWorkItemsView int

const (
	ViewBySchedule EnhancedWorkItemsView = iota
	ViewByType
)

// EnhancedWorkItemsModel represents the enhanced work items view
type EnhancedWorkItemsModel struct {
	dataClient       *data.EnhancedClient
	markdownItems    []*models.MarkdownWorkItem
	legacyItems      []models.WorkItem
	filteredItems    []list.Item
	list             list.Model
	viewMode         EnhancedWorkItemsView
	selectedSchedule string
	selectedType     string
	width            int
	height           int
	loading          bool
	error            error
	selectedItem     interface{} // Can be *models.MarkdownWorkItem or *models.WorkItem
	showingDetails   bool
}

type enhancedWorkItemListItem struct {
	markdown *models.MarkdownWorkItem
	legacy   *models.WorkItem
}

func (i enhancedWorkItemListItem) Title() string {
	if i.markdown != nil {
		return i.markdown.Summary
	}
	return i.legacy.Content
}

func (i enhancedWorkItemListItem) Description() string {
	if i.markdown != nil {
		return fmt.Sprintf("%s • %s • %s",
			i.markdown.GetDisplayType(),
			i.markdown.GetDisplaySchedule(),
			strings.Join(i.markdown.TechnicalTags, ", "))
	}
	return fmt.Sprintf("%s • %s • %s",
		i.legacy.GetDisplayType(),
		i.legacy.GetDisplayStatus(),
		i.legacy.GetPriority())
}

func (i enhancedWorkItemListItem) FilterValue() string {
	if i.markdown != nil {
		return i.markdown.Summary + " " + i.markdown.Type + " " + strings.Join(i.markdown.TechnicalTags, " ")
	}
	return i.legacy.Content + " " + i.legacy.Type + " " + i.legacy.Status
}

func NewEnhancedWorkItemsModel(dataClient *data.EnhancedClient) *EnhancedWorkItemsModel {
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

	return &EnhancedWorkItemsModel{
		dataClient:       dataClient,
		list:             l,
		viewMode:         ViewBySchedule,
		selectedSchedule: models.ScheduleNow,
		selectedType:     models.TypePlan,
		loading:          true,
	}
}

func (m *EnhancedWorkItemsModel) Init() tea.Cmd {
	return m.loadWorkItems()
}

func (m *EnhancedWorkItemsModel) Update(msg tea.Msg) (*EnhancedWorkItemsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width - 4)
		m.list.SetHeight(msg.Height - 8)
		return m, nil

	case WorkItemsLoadedMsg:
		m.loading = false
		m.error = msg.Error
		
		if m.error == nil {
			// Store legacy items
			m.legacyItems = msg.WorkItems
			m.applyFilter()
		}
		
		return m, nil

	case MarkdownItemsLoadedMsg:
		m.loading = false
		m.markdownItems = msg.MarkdownItems
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
			if item, ok := m.list.SelectedItem().(enhancedWorkItemListItem); ok {
				if item.markdown != nil {
					m.selectedItem = item.markdown
				} else {
					m.selectedItem = item.legacy
				}
				m.showingDetails = true
				return m, nil
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("r", "R"))):
			m.loading = true
			return m, m.loadWorkItems()

		case key.Matches(msg, key.NewBinding(key.WithKeys("v", "V"))):
			// Toggle view mode
			if m.viewMode == ViewBySchedule {
				m.viewMode = ViewByType
			} else {
				m.viewMode = ViewBySchedule
			}
			m.applyFilter()
			return m, nil

		// Schedule filters (when in schedule view)
		case key.Matches(msg, key.NewBinding(key.WithKeys("1"))) && m.viewMode == ViewBySchedule:
			m.selectedSchedule = models.ScheduleNow
			m.applyFilter()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("2"))) && m.viewMode == ViewBySchedule:
			m.selectedSchedule = models.ScheduleNext
			m.applyFilter()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("3"))) && m.viewMode == ViewBySchedule:
			m.selectedSchedule = models.ScheduleLater
			m.applyFilter()
			return m, nil

		// Type filters (when in type view)
		case key.Matches(msg, key.NewBinding(key.WithKeys("1"))) && m.viewMode == ViewByType:
			m.selectedType = models.TypePlan
			m.applyFilter()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("2"))) && m.viewMode == ViewByType:
			m.selectedType = models.TypeProposal
			m.applyFilter()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("3"))) && m.viewMode == ViewByType:
			m.selectedType = models.TypeAnalysis
			m.applyFilter()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("4"))) && m.viewMode == ViewByType:
			m.selectedType = models.TypeUpdate
			m.applyFilter()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("5"))) && m.viewMode == ViewByType:
			m.selectedType = models.TypeDecision
			m.applyFilter()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *EnhancedWorkItemsModel) View() string {
	if m.loading {
		return "Loading work items..."
	}

	if m.error != nil {
		return fmt.Sprintf("Error loading work items: %s", m.error.Error())
	}

	if m.showingDetails {
		return m.renderWorkItemDetails()
	}

	// Filter tabs
	filterTabs := m.renderFilterTabs()

	// Help text
	helpText := m.renderHelpText()

	return filterTabs + "\n" + m.list.View() + "\n" + helpText
}

func (m *EnhancedWorkItemsModel) renderFilterTabs() string {
	var tabs []string
	
	if m.viewMode == ViewBySchedule {
		// Schedule tabs: NOW, NEXT, LATER
		schedules := []struct {
			name     string
			schedule string
		}{
			{"NOW", models.ScheduleNow},
			{"NEXT", models.ScheduleNext},
			{"LATER", models.ScheduleLater},
		}

		for i, s := range schedules {
			style := lipgloss.NewStyle().
				Padding(0, 1).
				Margin(0, 1)

			if s.schedule == m.selectedSchedule {
				style = style.
					Background(lipgloss.Color("62")).
					Foreground(lipgloss.Color("230")).
					Bold(true)
			} else {
				style = style.
					Foreground(lipgloss.Color("244"))
			}

			count := m.getScheduleCount(s.schedule)
			tab := fmt.Sprintf("%d:%s(%d)", i+1, s.name, count)
			tabs = append(tabs, style.Render(tab))
		}
	} else {
		// Type tabs: Plan, Proposal, Analysis, Update, Decision
		types := []struct {
			name     string
			itemType string
		}{
			{"Plan", models.TypePlan},
			{"Proposal", models.TypeProposal},
			{"Analysis", models.TypeAnalysis},
			{"Update", models.TypeUpdate},
			{"Decision", models.TypeDecision},
		}

		for i, t := range types {
			style := lipgloss.NewStyle().
				Padding(0, 1).
				Margin(0, 1)

			if t.itemType == m.selectedType {
				style = style.
					Background(lipgloss.Color("62")).
					Foreground(lipgloss.Color("230")).
					Bold(true)
			} else {
				style = style.
					Foreground(lipgloss.Color("244"))
			}

			count := m.getTypeCount(t.itemType)
			tab := fmt.Sprintf("%d:%s(%d)", i+1, t.name, count)
			tabs = append(tabs, style.Render(tab))
		}
	}

	// Add view mode indicator
	viewModeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true).
		Margin(0, 2)
	
	viewModeText := "View: Schedule"
	if m.viewMode == ViewByType {
		viewModeText = "View: Type"
	}
	
	tabs = append(tabs, viewModeStyle.Render(viewModeText))

	return lipgloss.JoinHorizontal(lipgloss.Left, tabs...)
}

func (m *EnhancedWorkItemsModel) renderHelpText() string {
	help := []string{
		"↑/↓: Navigate",
		"Enter: View details",
		"r: Refresh",
		"v: Toggle view",
		"/: Search",
	}
	
	if m.viewMode == ViewBySchedule {
		help = append(help, "1-3: Filter schedule")
	} else {
		help = append(help, "1-5: Filter type")
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Render(strings.Join(help, " • "))
}

func (m *EnhancedWorkItemsModel) renderWorkItemDetails() string {
	if m.selectedItem == nil {
		return "No item selected"
	}

	var sections []string
	
	// Check if it's a markdown item
	if mdItem, ok := m.selectedItem.(*models.MarkdownWorkItem); ok {
		// Header
		header := lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Render(fmt.Sprintf("📋 %s Details", mdItem.GetDisplayType()))
		sections = append(sections, header)

		// Basic info
		basicInfo := []string{
			fmt.Sprintf("ID: %s", mdItem.ID),
			fmt.Sprintf("Type: %s", mdItem.GetDisplayType()),
			fmt.Sprintf("Schedule: %s", mdItem.GetDisplaySchedule()),
			fmt.Sprintf("Status: %s", mdItem.Metadata.Status),
			fmt.Sprintf("Tags: %s", strings.Join(mdItem.TechnicalTags, ", ")),
			fmt.Sprintf("Created: %s", mdItem.CreatedAt.Format("2006-01-02 15:04:05")),
		}

		basicSection := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1).
			Render(strings.Join(basicInfo, "\n"))
		sections = append(sections, basicSection)

		// Summary
		summarySection := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1).
			Render(fmt.Sprintf("Summary:\n%s", mdItem.Summary))
		sections = append(sections, summarySection)

		// Full content (truncated for display)
		content := mdItem.Content
		if len(content) > 500 {
			content = content[:497] + "..."
		}
		contentSection := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1).
			Render(fmt.Sprintf("Content:\n%s", content))
		sections = append(sections, contentSection)
		
		// Type-specific metadata
		if mdItem.Type == models.TypePlan && mdItem.Metadata.ImplementationStatus != "" {
			metaInfo := []string{
				fmt.Sprintf("Implementation: %s", mdItem.Metadata.ImplementationStatus),
				fmt.Sprintf("Effort: %s", mdItem.Metadata.EstimatedEffort),
			}
			if len(mdItem.Metadata.Phases) > 0 {
				metaInfo = append(metaInfo, fmt.Sprintf("Phases: %s", strings.Join(mdItem.Metadata.Phases, ", ")))
			}
			
			metaSection := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(1).
				Render(strings.Join(metaInfo, "\n"))
			sections = append(sections, metaSection)
		}
	} else if legacyItem, ok := m.selectedItem.(*models.WorkItem); ok {
		// Render legacy item details (same as before)
		sections = append(sections, m.renderLegacyItemDetails(legacyItem)...)
	}

	// Help
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Render("Press ESC or q to go back")
	sections = append(sections, help)

	return strings.Join(sections, "\n\n")
}

func (m *EnhancedWorkItemsModel) renderLegacyItemDetails(item *models.WorkItem) []string {
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

	return sections
}

func (m *EnhancedWorkItemsModel) getScheduleCount(schedule string) int {
	count := 0
	for _, item := range m.markdownItems {
		if item.Schedule == schedule && item.Metadata.Status != models.StatusCompleted {
			count++
		}
	}
	return count
}

func (m *EnhancedWorkItemsModel) getTypeCount(itemType string) int {
	count := 0
	for _, item := range m.markdownItems {
		if item.Type == itemType && item.Metadata.Status != models.StatusCompleted {
			count++
		}
	}
	// Also count legacy items
	for _, item := range m.legacyItems {
		if item.Type == itemType && item.Status != "completed" {
			count++
		}
	}
	return count
}

func (m *EnhancedWorkItemsModel) applyFilter() {
	m.filteredItems = []list.Item{}
	
	if m.viewMode == ViewBySchedule {
		// Filter by schedule
		for _, item := range m.markdownItems {
			if item.Schedule == m.selectedSchedule && item.Metadata.Status != models.StatusCompleted {
				m.filteredItems = append(m.filteredItems, enhancedWorkItemListItem{markdown: item})
			}
		}
		
		// Update list title
		m.list.Title = fmt.Sprintf("Work Items - %s (%d)", strings.ToUpper(m.selectedSchedule), len(m.filteredItems))
	} else {
		// Filter by type
		for _, item := range m.markdownItems {
			if item.Type == m.selectedType && item.Metadata.Status != models.StatusCompleted {
				m.filteredItems = append(m.filteredItems, enhancedWorkItemListItem{markdown: item})
			}
		}
		
		// Also include legacy items
		for _, item := range m.legacyItems {
			if item.Type == m.selectedType && item.Status != "completed" {
				itemCopy := item // Create a copy to avoid pointer issues
				m.filteredItems = append(m.filteredItems, enhancedWorkItemListItem{legacy: &itemCopy})
			}
		}
		
		// Update list title
		m.list.Title = fmt.Sprintf("Work Items - %s (%d)", m.getTypeDisplayName(m.selectedType), len(m.filteredItems))
	}
	
	// Set items
	m.list.SetItems(m.filteredItems)
}

func (m *EnhancedWorkItemsModel) getTypeDisplayName(itemType string) string {
	switch itemType {
	case models.TypePlan:
		return "Plans"
	case models.TypeProposal:
		return "Proposals"
	case models.TypeAnalysis:
		return "Analyses"
	case models.TypeUpdate:
		return "Updates"
	case models.TypeDecision:
		return "Decisions"
	default:
		return "Unknown"
	}
}

func (m *EnhancedWorkItemsModel) loadWorkItems() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		// Load all schedules
		allMarkdown := []*models.MarkdownWorkItem{}
		schedules := []string{models.ScheduleNow, models.ScheduleNext, models.ScheduleLater}
		for _, schedule := range schedules {
			items, err := m.dataClient.GetWorkItemsBySchedule(schedule)
			if err == nil {
				allMarkdown = append(allMarkdown, items...)
			}
		}
		
		// Also load decisions
		decisions, err := m.dataClient.GetWorkItemsByTypeAndSchedule(models.TypeDecision, "")
		if err == nil {
			allMarkdown = append(allMarkdown, decisions...)
		}
		
		if len(allMarkdown) > 0 {
			return MarkdownItemsLoadedMsg{
				MarkdownItems: allMarkdown,
				Error:         nil,
			}
		}
		
		// Fallback to legacy items if no markdown items found
		workItems, err := m.dataClient.GetAllWorkItemsEnhanced()
		return WorkItemsLoadedMsg{
			WorkItems: workItems,
			Error:     err,
		}
	})
}

// MarkdownItemsLoadedMsg represents loaded markdown items
type MarkdownItemsLoadedMsg struct {
	MarkdownItems []*models.MarkdownWorkItem
	Error         error
}