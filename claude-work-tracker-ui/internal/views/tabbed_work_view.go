package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"claude-work-tracker-ui/internal/data"
	"claude-work-tracker-ui/internal/models"
)

// Tab represents a schedule category
type Tab struct {
	Name     string
	Schedule string
	Active   bool
}

// TabbedWorkView provides a tabbed interface for NOW/NEXT/LATER work items
type TabbedWorkView struct {
	dataClient    *data.EnhancedClient
	tabs          []Tab
	activeTab     int
	viewport      viewport.Model
	workItems     map[string][]*models.Work // keyed by schedule
	selectedItem  int
	width         int
	height        int
	glamourRender *glamour.TermRenderer
	showDetail    bool
	showUpdates   bool
	updatesView   UpdatesView
	ready         bool
	keys          KeyMap
}

// KeyMap defines keyboard shortcuts
type KeyMap struct {
	NextTab     key.Binding
	PrevTab     key.Binding
	NextItem    key.Binding
	PrevItem    key.Binding
	ViewDetail  key.Binding
	ViewUpdates key.Binding
	Back        key.Binding
	Quit        key.Binding
}

// DefaultKeyMap returns default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		NextTab: key.NewBinding(
			key.WithKeys("tab", "right"),
			key.WithHelp("tab/→", "next tab"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("shift+tab", "left"),
			key.WithHelp("shift+tab/←", "prev tab"),
		),
		NextItem: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "next item"),
		),
		PrevItem: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "prev item"),
		),
		ViewDetail: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "view detail"),
		),
		ViewUpdates: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "view updates"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back to list"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// Styles for the tabbed interface (based on bubbletea tabs example)
var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle          = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor    = lipgloss.AdaptiveColor{Light: "235", Dark: "241"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = lipgloss.NewStyle().Border(activeTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Center).Border(lipgloss.NormalBorder()).UnsetBorderTop()

	contentStyle = lipgloss.NewStyle().
			Padding(1, 2)

	itemListStyle = lipgloss.NewStyle().
			Padding(0, 1)

	selectedItemStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("62")).
				Foreground(lipgloss.Color("15")).
				Padding(0, 1)

	normalItemStyle = lipgloss.NewStyle().
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(1, 0, 0, 2)
)

// tabBorderWithBottom creates a custom border with specified bottom characters
func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

// NewTabbedWorkView creates a new tabbed work view
func NewTabbedWorkView(dataClient *data.EnhancedClient) *TabbedWorkView {
	// Initialize glamour renderer
	glamourRenderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)

	tabs := []Tab{
		{Name: "NOW", Schedule: models.ScheduleNow, Active: true},
		{Name: "NEXT", Schedule: models.ScheduleNext, Active: false},
		{Name: "LATER", Schedule: models.ScheduleLater, Active: false},
	}

	vp := viewport.New(78, 20)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)

	return &TabbedWorkView{
		dataClient:    dataClient,
		tabs:          tabs,
		activeTab:     0,
		viewport:      vp,
		workItems:     make(map[string][]*models.Work),
		selectedItem:  0,
		glamourRender: glamourRenderer,
		showDetail:    false,
		showUpdates:   false,
		updatesView:   NewUpdatesView(),
		keys:          DefaultKeyMap(),
	}
}

// Init initializes the tabbed work view
func (t *TabbedWorkView) Init() tea.Cmd {
	return t.loadWorkItems()
}

// loadWorkItems loads work items for all schedules
func (t *TabbedWorkView) loadWorkItems() tea.Cmd {
	return tea.Batch(
		t.loadScheduleItems(models.ScheduleNow),
		t.loadScheduleItems(models.ScheduleNext),
		t.loadScheduleItems(models.ScheduleLater),
	)
}

// errMsg represents an error message
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

// loadScheduleItems loads work items for a specific schedule
func (t *TabbedWorkView) loadScheduleItems(schedule string) tea.Cmd {
	return func() tea.Msg {
		items, err := t.dataClient.GetWorkBySchedule(schedule)
		if err != nil {
			return errMsg{err}
		}
		return scheduleItemsLoadedMsg{schedule: schedule, items: items}
	}
}

// scheduleItemsLoadedMsg carries loaded work items for a schedule
type scheduleItemsLoadedMsg struct {
	schedule string
	items    []*models.Work
}

// Update handles messages and user input
func (t *TabbedWorkView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height
		
		// Update viewport size with conservative margins
		headerHeight := 4 // Space for tabs 
		footerHeight := 2 // Space for help text
		borderSpace := 4  // Space for borders and padding
		
		viewportWidth := msg.Width - borderSpace - 4 // Extra margin
		viewportHeight := msg.Height - headerHeight - footerHeight - 4 // Extra margin
		
		if viewportWidth < 10 {
			viewportWidth = 10
		}
		if viewportHeight < 3 {
			viewportHeight = 3
		}
		
		t.viewport.Width = viewportWidth
		t.viewport.Height = viewportHeight
		
		if !t.ready {
			t.ready = true
		}

	case scheduleItemsLoadedMsg:
		t.workItems[msg.schedule] = msg.items
		if msg.schedule == t.getCurrentSchedule() {
			t.selectedItem = 0
			t.updateViewportContent()
		}

	case tea.KeyMsg:
		if t.showUpdates {
			// Handle updates view navigation
			switch {
			case key.Matches(msg, t.keys.Back):
				t.showUpdates = false
				t.updateViewportContent()
			default:
				var updatesCmd tea.Cmd
				t.updatesView, updatesCmd = t.updatesView.Update(msg)
				return t, updatesCmd
			}
		} else if t.showDetail {
			// Handle detail view navigation
			switch {
			case key.Matches(msg, t.keys.Back):
				t.showDetail = false
				t.updateViewportContent()
			case key.Matches(msg, t.keys.ViewUpdates):
				t.showUpdates = true
				// Load updates for current work item
				currentWork := t.getCurrentWork()
				if currentWork != nil {
					t.updatesView.LoadFromWork(currentWork)
					t.updatesView, cmd = t.updatesView.Update(tea.WindowSizeMsg{
						Width:  t.width,
						Height: t.height,
					})
				}
			default:
				t.viewport, cmd = t.viewport.Update(msg)
			}
		} else {
			// Handle list view navigation
			switch {
			case key.Matches(msg, t.keys.NextTab):
				t.nextTab()
				t.updateViewportContent()
			case key.Matches(msg, t.keys.PrevTab):
				t.prevTab()
				t.updateViewportContent()
			case key.Matches(msg, t.keys.NextItem):
				t.nextItem()
				t.updateViewportContent()
			case key.Matches(msg, t.keys.PrevItem):
				t.prevItem()
				t.updateViewportContent()
			case key.Matches(msg, t.keys.ViewDetail):
				t.showDetail = true
				t.updateViewportContent()
			}
		}
	}

	return t, cmd
}

// View renders the tabbed work view
func (t *TabbedWorkView) View() string {
	if !t.ready {
		return "Loading..."
	}

	// Show updates view if active
	if t.showUpdates {
		return t.updatesView.View()
	}

	// Render tab bar and content as connected elements
	tabBar := t.renderTabBar()
	content := t.renderConnectedContent()
	help := t.renderHelp()
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		tabBar,
		content,
		help,
	)
}

// renderConnectedContent renders content that connects directly to the tabs
func (t *TabbedWorkView) renderConnectedContent() string {
	// Create a content area that connects to the tabs without double borders
	// Use more conservative sizing
	contentWidth := t.width - 6  // More margin
	contentHeight := t.height - 10 // More space for tabs and help
	
	if contentWidth < 20 {
		contentWidth = 20 // Minimum width
	}
	if contentHeight < 5 {
		contentHeight = 5 // Minimum height
	}
	
	contentStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(highlightColor).
		BorderTop(false). // No top border to connect with tabs
		Padding(1, 1).    // Reduced padding
		Width(contentWidth).
		Height(contentHeight)
	
	return contentStyle.Render(t.viewport.View())
}

// renderTabBar creates the top tab bar that connects to content
func (t *TabbedWorkView) renderTabBar() string {
	var renderedTabs []string
	
	for i, tab := range t.tabs {
		var style lipgloss.Style
		isActive := i == t.activeTab
		if isActive {
			style = activeTabStyle.Copy()
		} else {
			style = inactiveTabStyle.Copy()
		}
		
		// Add item count
		count := len(t.workItems[tab.Schedule])
		tabText := fmt.Sprintf("%s (%d)", tab.Name, count)
		
		renderedTabs = append(renderedTabs, style.Render(tabText))
	}
	
	// Join tabs and add a horizontal line to connect with content
	tabRow := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	
	// Create a bottom border line that extends to match content width
	availableWidth := t.width - 6 // Match content width calculation
	gap := availableWidth - lipgloss.Width(tabRow)
	if gap > 0 {
		borderLine := lipgloss.NewStyle().
			Foreground(highlightColor).
			Render(strings.Repeat("─", gap))
		tabRow = lipgloss.JoinHorizontal(lipgloss.Bottom, tabRow, borderLine)
	}
	
	return tabRow // Remove extra docStyle wrapping
}

// renderHelp creates the help text
func (t *TabbedWorkView) renderHelp() string {
	if t.showDetail {
		return helpStyle.Render("esc: back to list • ↑/↓: scroll • q: quit")
	}
	return helpStyle.Render("tab: switch tabs • ↑/↓: navigate • enter: view detail • q: quit")
}

// updateViewportContent updates the content shown in the viewport
func (t *TabbedWorkView) updateViewportContent() {
	schedule := t.getCurrentSchedule()
	items := t.workItems[schedule]
	
	if len(items) == 0 {
		t.viewport.SetContent(contentStyle.Render("No items in this schedule"))
		return
	}
	
	if t.showDetail {
		// Render detailed view of selected item
		if t.selectedItem < len(items) {
			t.renderItemDetail(items[t.selectedItem])
		}
	} else {
		// Render item list
		t.renderItemList(items)
	}
}

// renderItemList renders the list of work items
func (t *TabbedWorkView) renderItemList(items []*models.Work) {
	var content strings.Builder
	
	for i, item := range items {
		style := normalItemStyle
		if i == t.selectedItem {
			style = selectedItemStyle
		}
		
		// Create item summary
		statusIcon := t.getStatusIcon(item.Metadata.Status)
		summary := fmt.Sprintf("%s %s", statusIcon, item.Title)
		
		// Add tags if any
		if len(item.TechnicalTags) > 0 {
			tagStr := strings.Join(item.TechnicalTags, ", ")
			summary += fmt.Sprintf(" [%s]", tagStr)
		}
		
		content.WriteString(style.Render(summary))
		content.WriteString("\n")
	}
	
	t.viewport.SetContent(content.String())
}

// renderItemDetail renders the detailed view of a work item
func (t *TabbedWorkView) renderItemDetail(item *models.Work) {
	// Create full markdown content
	var content strings.Builder
	
	// Header with metadata
	content.WriteString(fmt.Sprintf("# %s\n\n", item.Title))
	content.WriteString(fmt.Sprintf("**Status:** %s | **Schedule:** %s\n\n", 
		strings.Title(item.Metadata.Status), strings.ToUpper(item.Schedule)))
	
	if item.Metadata.Priority != "" {
		content.WriteString(fmt.Sprintf("**Priority:** %s\n\n", strings.Title(item.Metadata.Priority)))
	}
	
	if len(item.TechnicalTags) > 0 {
		content.WriteString(fmt.Sprintf("**Tags:** %s\n\n", 
			strings.Join(item.TechnicalTags, ", ")))
	}
	
	if item.Metadata.ProgressPercent > 0 {
		content.WriteString(fmt.Sprintf("**Progress:** %d%%\n\n", item.Metadata.ProgressPercent))
	}
	
	if item.Metadata.ArtifactCount > 0 {
		content.WriteString(fmt.Sprintf("**Artifacts:** %d\n\n", item.Metadata.ArtifactCount))
	}
	
	// Add separator
	content.WriteString("---\n\n")
	
	// Add description
	if item.Description != "" {
		content.WriteString(item.Description)
		content.WriteString("\n\n")
	}
	
	// Add the main content
	if item.Content != "" {
		content.WriteString(item.Content)
		content.WriteString("\n\n")
	}
	
	// Fetch and render associated artifacts automatically
	if len(item.ArtifactRefs) > 0 {
		artifacts, err := t.dataClient.GetWorkArtifacts(item.ID)
		if err == nil && len(artifacts) > 0 {
			content.WriteString("## Associated Artifacts\n\n")
			
			for i, artifact := range artifacts {
				// Add artifact separator
				if i > 0 {
					content.WriteString("\n---\n\n")
				}
				
				// Artifact header
				content.WriteString(fmt.Sprintf("### %s (%s)\n\n", 
					artifact.Summary, strings.Title(artifact.Type)))
				
				// Artifact metadata
				if len(artifact.TechnicalTags) > 0 {
					content.WriteString(fmt.Sprintf("**Tags:** %s\n\n", 
						strings.Join(artifact.TechnicalTags, ", ")))
				}
				
				if artifact.Metadata.Status != "" {
					content.WriteString(fmt.Sprintf("**Status:** %s\n\n", 
						strings.Title(artifact.Metadata.Status)))
				}
				
				// Artifact content
				if artifact.Content != "" {
					content.WriteString(artifact.Content)
					content.WriteString("\n\n")
				}
			}
		}
	}
	
	// Render with glamour
	rendered, err := t.glamourRender.Render(content.String())
	if err != nil {
		// Fallback to plain text if rendering fails
		t.viewport.SetContent(content.String())
	} else {
		t.viewport.SetContent(rendered)
	}
}

// getTypeIcon returns an icon for the work item type
func (t *TabbedWorkView) getStatusIcon(status string) string {
	switch status {
	case models.WorkStatusInProgress:
		return "🔄"
	case models.WorkStatusCompleted:
		return "✅"
	case models.WorkStatusBlocked:
		return "🚫"
	case models.WorkStatusActive:
		return "🎯"
	default:
		return "📋"
	}
}

func (t *TabbedWorkView) getTypeIcon(itemType string) string {
	switch itemType {
	case models.TypePlan:
		return "📋"
	case models.TypeProposal:
		return "💡"
	case models.TypeAnalysis:
		return "🔍"
	case models.TypeUpdate:
		return "📝"
	case models.TypeDecision:
		return "⚖️"
	default:
		return "📄"
	}
}

// Navigation helpers (following bubbletea tabs example pattern)
func (t *TabbedWorkView) nextTab() {
	t.activeTab = min(t.activeTab+1, len(t.tabs)-1)
	t.selectedItem = 0
}

func (t *TabbedWorkView) prevTab() {
	t.activeTab = max(t.activeTab-1, 0)
	t.selectedItem = 0
}

// Helper functions for min/max
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (t *TabbedWorkView) nextItem() {
	items := t.workItems[t.getCurrentSchedule()]
	if len(items) > 0 {
		t.selectedItem = (t.selectedItem + 1) % len(items)
	}
}

func (t *TabbedWorkView) prevItem() {
	items := t.workItems[t.getCurrentSchedule()]
	if len(items) > 0 {
		t.selectedItem = (t.selectedItem - 1 + len(items)) % len(items)
	}
}

func (t *TabbedWorkView) getCurrentSchedule() string {
	if t.activeTab < len(t.tabs) {
		return t.tabs[t.activeTab].Schedule
	}
	return models.ScheduleNow
}

// getCurrentWork returns the currently selected work item
func (t *TabbedWorkView) getCurrentWork() *models.Work {
	schedule := t.getCurrentSchedule()
	items := t.workItems[schedule]
	if t.selectedItem >= 0 && t.selectedItem < len(items) {
		return items[t.selectedItem]
	}
	return nil
}