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

type FutureWorkViewType int

const (
	FutureItemsView FutureWorkViewType = iota
	FutureGroupsView
)

type FutureWorkModel struct {
	dataClient    *data.Client
	futureItems   []models.FutureWorkItem
	futureGroups  []models.FutureWorkGroup
	itemsList     list.Model
	groupsList    list.Model
	currentView   FutureWorkViewType
	width         int
	height        int
	loading       bool
	error         error
	selectedItem  *models.FutureWorkItem
	selectedGroup *models.FutureWorkGroup
	showingDetails bool
}

type FutureWorkLoadedMsg struct {
	FutureItems  []models.FutureWorkItem
	FutureGroups []models.FutureWorkGroup
	Error        error
}

type futureItemListItem struct {
	item models.FutureWorkItem
}

func (i futureItemListItem) Title() string {
	return i.item.Content
}

func (i futureItemListItem) Description() string {
	return fmt.Sprintf("%s • %s • %s", 
		i.item.OriginalType,
		i.item.GroupingStatus,
		i.item.PriorityWhenPromoted)
}

func (i futureItemListItem) FilterValue() string {
	return i.item.Content + " " + i.item.OriginalType
}

type futureGroupListItem struct {
	group models.FutureWorkGroup
}

func (i futureGroupListItem) Title() string {
	return i.group.Name
}

func (i futureGroupListItem) Description() string {
	return fmt.Sprintf("%s • %d items • %s", 
		i.group.Description,
		len(i.group.Items),
		i.group.StrategicValue)
}

func (i futureGroupListItem) FilterValue() string {
	return i.group.Name + " " + i.group.Description
}

var (
	futureWorkStyle = lipgloss.NewStyle().
			Padding(1, 2)

	futureHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Bold(true)

	futureTabStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Margin(0, 1)

	futureActiveTabStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("205")).
				Foreground(lipgloss.Color("230")).
				Bold(true).
				Padding(0, 1).
				Margin(0, 1)

	futureDetailStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("205")).
				Padding(1)
)

func NewFutureWorkModel(dataClient *data.Client) *FutureWorkModel {
	// Create items list
	itemsDelegate := list.NewDefaultDelegate()
	itemsDelegate.Styles.SelectedTitle = itemsDelegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("205")).
		Bold(true)
	itemsDelegate.Styles.SelectedDesc = itemsDelegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("244"))

	itemsList := list.New([]list.Item{}, itemsDelegate, 0, 0)
	itemsList.Title = "Future Work Items"
	itemsList.SetFilteringEnabled(true)
	itemsList.SetShowStatusBar(true)
	itemsList.SetShowPagination(true)
	itemsList.SetShowHelp(false)

	// Create groups list
	groupsDelegate := list.NewDefaultDelegate()
	groupsDelegate.Styles.SelectedTitle = groupsDelegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("205")).
		Bold(true)
	groupsDelegate.Styles.SelectedDesc = groupsDelegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("244"))

	groupsList := list.New([]list.Item{}, groupsDelegate, 0, 0)
	groupsList.Title = "Future Work Groups"
	groupsList.SetFilteringEnabled(true)
	groupsList.SetShowStatusBar(true)
	groupsList.SetShowPagination(true)
	groupsList.SetShowHelp(false)

	return &FutureWorkModel{
		dataClient:  dataClient,
		itemsList:   itemsList,
		groupsList:  groupsList,
		currentView: FutureItemsView,
		loading:     true,
	}
}

func (m *FutureWorkModel) Init() tea.Cmd {
	return m.loadFutureWork()
}

func (m *FutureWorkModel) Update(msg tea.Msg) (*FutureWorkModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.itemsList.SetWidth(msg.Width - 4)
		m.itemsList.SetHeight(msg.Height - 12)
		m.groupsList.SetWidth(msg.Width - 4)
		m.groupsList.SetHeight(msg.Height - 12)
		return m, nil

	case FutureWorkLoadedMsg:
		m.loading = false
		m.futureItems = msg.FutureItems
		m.futureGroups = msg.FutureGroups
		m.error = msg.Error
		
		if m.error == nil {
			m.updateLists()
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
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			if m.currentView == FutureItemsView {
				m.currentView = FutureGroupsView
			} else {
				m.currentView = FutureItemsView
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if m.currentView == FutureItemsView {
				if item, ok := m.itemsList.SelectedItem().(futureItemListItem); ok {
					m.selectedItem = &item.item
					m.showingDetails = true
					return m, nil
				}
			} else {
				if group, ok := m.groupsList.SelectedItem().(futureGroupListItem); ok {
					m.selectedGroup = &group.group
					m.showingDetails = true
					return m, nil
				}
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("r", "R"))):
			m.loading = true
			return m, m.loadFutureWork()
		}
	}

	// Forward to appropriate list
	var cmd tea.Cmd
	if m.currentView == FutureItemsView {
		m.itemsList, cmd = m.itemsList.Update(msg)
	} else {
		m.groupsList, cmd = m.groupsList.Update(msg)
	}
	return m, cmd
}

func (m *FutureWorkModel) View() string {
	if m.loading {
		return "Loading future work..."
	}

	if m.error != nil {
		return fmt.Sprintf("Error loading future work: %s", m.error.Error())
	}

	// Check if we have any data at all
	if len(m.futureItems) == 0 && len(m.futureGroups) == 0 {
		return futureWorkStyle.Render(
			futureHeaderStyle.Render("🚀 Future Work Management") + "\n\n" +
			"No future work items found in current directory.\n" +
			"Future work is only shown for directories with local work tracking enabled.\n\n" +
			"To enable local work tracking, create a .claude-work directory or\n" +
			"use Claude Code to start a work session in this directory.")
	}

	if m.showingDetails {
		if m.selectedItem != nil {
			return m.renderItemDetails()
		}
		if m.selectedGroup != nil {
			return m.renderGroupDetails()
		}
	}

	var sections []string

	// Header
	header := futureHeaderStyle.Render("🚀 Future Work Management")
	sections = append(sections, header)

	// Tabs
	tabs := m.renderTabs()
	sections = append(sections, tabs)

	// Current view
	if m.currentView == FutureItemsView {
		sections = append(sections, m.itemsList.View())
	} else {
		sections = append(sections, m.groupsList.View())
	}

	// Help
	help := m.renderHelp()
	sections = append(sections, help)

	return futureWorkStyle.Render(strings.Join(sections, "\n"))
}

func (m *FutureWorkModel) renderTabs() string {
	var tabs []string

	// Items tab
	itemsCount := len(m.futureItems)
	itemsText := fmt.Sprintf("Items (%d)", itemsCount)
	if m.currentView == FutureItemsView {
		tabs = append(tabs, futureActiveTabStyle.Render(itemsText))
	} else {
		tabs = append(tabs, futureTabStyle.Render(itemsText))
	}

	// Groups tab
	groupsCount := len(m.futureGroups)
	groupsText := fmt.Sprintf("Groups (%d)", groupsCount)
	if m.currentView == FutureGroupsView {
		tabs = append(tabs, futureActiveTabStyle.Render(groupsText))
	} else {
		tabs = append(tabs, futureTabStyle.Render(groupsText))
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, tabs...)
}

func (m *FutureWorkModel) renderHelp() string {
	help := []string{
		"Tab: Switch view",
		"Enter: View details",
		"r: Refresh",
		"/: Search",
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Render(strings.Join(help, " • "))
}

func (m *FutureWorkModel) renderItemDetails() string {
	item := m.selectedItem
	if item == nil {
		return "No item selected"
	}

	var sections []string

	// Header
	header := futureHeaderStyle.Render("🚀 Future Work Item Details")
	sections = append(sections, header)

	// Basic info
	basicInfo := []string{
		fmt.Sprintf("ID: %s", item.ID),
		fmt.Sprintf("Type: %s", item.OriginalType),
		fmt.Sprintf("Grouping Status: %s", item.GroupingStatus),
		fmt.Sprintf("Priority When Promoted: %s", item.PriorityWhenPromoted),
		fmt.Sprintf("Created: %s", item.CreatedAt),
	}

	basicSection := futureDetailStyle.Render(
		"Basic Information:\n" + strings.Join(basicInfo, "\n"))
	sections = append(sections, basicSection)

	// Content
	contentSection := futureDetailStyle.Render(
		fmt.Sprintf("Content:\n%s", item.Content))
	sections = append(sections, contentSection)

	// Context
	contextInfo := []string{
		fmt.Sprintf("Deprioritized From: %s", item.Context.DeprioritizedFrom),
		fmt.Sprintf("Deprioritized Date: %s", item.Context.DeprioritizedDate),
		fmt.Sprintf("Reason: %s", item.Context.DeprioritizedReason),
	}

	if item.Context.SuggestedGroup != "" {
		contextInfo = append(contextInfo, fmt.Sprintf("Suggested Group: %s", item.Context.SuggestedGroup))
	}

	contextSection := futureDetailStyle.Render(
		"Context:\n" + strings.Join(contextInfo, "\n"))
	sections = append(sections, contextSection)

	// Similarity metadata
	metadata := item.SimilarityMetadata
	metadataInfo := []string{
		fmt.Sprintf("Feature Domain: %s", metadata.FeatureDomain),
		fmt.Sprintf("Technical Domain: %s", metadata.TechnicalDomain),
		fmt.Sprintf("Strategic Theme: %s", metadata.StrategicTheme),
	}

	if len(metadata.Keywords) > 0 {
		metadataInfo = append(metadataInfo, fmt.Sprintf("Keywords: %s", strings.Join(metadata.Keywords, ", ")))
	}

	if len(metadata.CodeLocations) > 0 {
		metadataInfo = append(metadataInfo, fmt.Sprintf("Code Locations: %s", strings.Join(metadata.CodeLocations, ", ")))
	}

	metadataSection := futureDetailStyle.Render(
		"Similarity Metadata:\n" + strings.Join(metadataInfo, "\n"))
	sections = append(sections, metadataSection)

	// Help
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Render("Press ESC or q to go back")
	sections = append(sections, help)

	return strings.Join(sections, "\n\n")
}

func (m *FutureWorkModel) renderGroupDetails() string {
	group := m.selectedGroup
	if group == nil {
		return "No group selected"
	}

	var sections []string

	// Header
	header := futureHeaderStyle.Render("🚀 Future Work Group Details")
	sections = append(sections, header)

	// Basic info
	basicInfo := []string{
		fmt.Sprintf("ID: %s", group.ID),
		fmt.Sprintf("Name: %s", group.Name),
		fmt.Sprintf("Description: %s", group.Description),
		fmt.Sprintf("Strategic Value: %s", group.StrategicValue),
		fmt.Sprintf("Estimated Effort: %s", group.EstimatedEffort),
		fmt.Sprintf("Readiness Status: %s", group.ReadinessStatus),
		fmt.Sprintf("Created: %s", group.CreatedDate),
		fmt.Sprintf("Last Updated: %s", group.LastUpdated),
	}

	basicSection := futureDetailStyle.Render(
		"Basic Information:\n" + strings.Join(basicInfo, "\n"))
	sections = append(sections, basicSection)

	// Items
	itemsInfo := []string{
		fmt.Sprintf("Items Count: %d", len(group.Items)),
		fmt.Sprintf("Similarity Score: %.2f", group.SimilarityScore),
	}

	if len(group.Items) > 0 {
		itemsInfo = append(itemsInfo, "Item IDs:")
		for _, itemID := range group.Items {
			itemsInfo = append(itemsInfo, fmt.Sprintf("  • %s", itemID))
		}
	}

	itemsSection := futureDetailStyle.Render(
		"Items:\n" + strings.Join(itemsInfo, "\n"))
	sections = append(sections, itemsSection)

	// Help
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Render("Press ESC or q to go back")
	sections = append(sections, help)

	return strings.Join(sections, "\n\n")
}

func (m *FutureWorkModel) updateLists() {
	// Update items list
	var itemsListItems []list.Item
	for _, item := range m.futureItems {
		itemsListItems = append(itemsListItems, futureItemListItem{item: item})
	}
	m.itemsList.SetItems(itemsListItems)

	// Update groups list
	var groupsListItems []list.Item
	for _, group := range m.futureGroups {
		groupsListItems = append(groupsListItems, futureGroupListItem{group: group})
	}
	m.groupsList.SetItems(groupsListItems)
}

func (m *FutureWorkModel) loadFutureWork() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		futureItems, err := m.dataClient.GetFutureWorkItems()
		if err != nil {
			return FutureWorkLoadedMsg{Error: err}
		}

		futureGroups, err := m.dataClient.GetFutureWorkGroups()
		if err != nil {
			return FutureWorkLoadedMsg{Error: err}
		}

		return FutureWorkLoadedMsg{
			FutureItems:  futureItems,
			FutureGroups: futureGroups,
			Error:        nil,
		}
	})
}