package views

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"claude-work-tracker-ui/internal/data"
	"claude-work-tracker-ui/internal/models"
	"claude-work-tracker-ui/internal/renderer"
)

// WorkItem implements list.Item interface for Work items
type WorkItem struct {
	*models.Work
}


func (w WorkItem) FilterValue() string {
	return w.Title
}

// Custom item delegate for fancy list rendering
type ItemDelegate struct {
	showDetail bool
	glamour    *glamour.TermRenderer
}

func (d ItemDelegate) Height() int {
	return 4 // Consistent height for all items to prevent layout issues
}

func (d ItemDelegate) Spacing() int { return 2 }

func (d ItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	workItem, ok := listItem.(WorkItem)
	if !ok {
		return
	}

	item := workItem.Work
	isSelected := index == m.Index()

	// Base styles
	var (
		statusStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("242")).
				Foreground(lipgloss.Color("250")).
				Bold(true).
				Padding(0, 1)
		
		titleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("246")).
				Bold(true)
		
		overviewStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("242")).
				PaddingLeft(0).
				PaddingTop(1)
		
		metadataStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				PaddingTop(1)
		
		selectedStatusStyle = statusStyle.Copy().
				Background(lipgloss.Color("62")).
				Foreground(lipgloss.Color("15"))
		
		selectedTitleStyle = titleStyle.Copy().
				Foreground(lipgloss.Color("170"))
		
		selectedOverviewStyle = overviewStyle.Copy().
				Foreground(lipgloss.Color("252"))
		
		selectedMetadataStyle = metadataStyle.Copy().
				Foreground(lipgloss.Color("245"))
	)

	// Apply selected styles
	if isSelected {
		statusStyle = selectedStatusStyle
		titleStyle = selectedTitleStyle
		overviewStyle = selectedOverviewStyle
		metadataStyle = selectedMetadataStyle
	}

	// Status/priority badge with inverted background
	var badgeText string
	switch item.Metadata.Status {
	case models.WorkStatusInProgress:
		badgeText = "IN_PROGRESS"
	case models.WorkStatusBlocked:
		badgeText = "BLOCKED"
	case models.WorkStatusCompleted:
		badgeText = "COMPLETED"
	default:
		badgeText = strings.ToUpper(item.Metadata.Priority)
		if badgeText == "" {
			badgeText = "WORK"
		}
	}
	
	statusBadge := statusStyle.Render(badgeText)
	
	// Title line with status badge + title
	titleLine := lipgloss.JoinHorizontal(lipgloss.Center, statusBadge, " ", titleStyle.Render(item.Title))

	content := titleLine

	// Show description with task summary for better overview
	var overviewText string
	if item.Description != "" {
		overviewText = item.Description
	}
	
	// Add task summary from content if available
	if item.Content != "" {
		// Use a simple processor instance for display (performance optimization)
		processor := renderer.NewMarkdownProcessor("")
		taskSummary := processor.GetTaskSummary(item.Content)
		if taskSummary != "" {
			if overviewText != "" {
				overviewText = overviewText + "\n" + taskSummary
			} else {
				overviewText = taskSummary
			}
		}
	}
	
	if overviewText != "" {
		// Limit overview length for list display
		if len(overviewText) > 120 {
			overviewText = overviewText[:120] + "..."
		}
		renderedOverview := overviewStyle.Render(overviewText)
		content = lipgloss.JoinVertical(lipgloss.Left, content, renderedOverview)
	}

	// Add metadata line
	var metaParts []string
	
	// Add progress if available
	if item.Metadata.ProgressPercent > 0 {
		metaParts = append(metaParts, fmt.Sprintf("progress:%d%%", item.Metadata.ProgressPercent))
	}
	
	// Add artifact count
	if item.Metadata.ArtifactCount > 0 {
		metaParts = append(metaParts, fmt.Sprintf("artifacts:%d", item.Metadata.ArtifactCount))
	}
	
	// Add technical tags
	if len(item.TechnicalTags) > 0 {
		metaParts = append(metaParts, strings.Join(item.TechnicalTags, ", "))
	}
	
	// Add git context
	if item.GitContext.Branch != "" {
		metaParts = append(metaParts, "branch:"+item.GitContext.Branch)
	}
	if item.GitContext.Worktree != "" {
		// Extract just the worktree name (last part of path)
		worktreeName := item.GitContext.Worktree
		if lastSlash := strings.LastIndex(worktreeName, "/"); lastSlash != -1 {
			worktreeName = worktreeName[lastSlash+1:]
		}
		metaParts = append(metaParts, "wt:"+worktreeName)
	}
	
	// Show last update time
	lastUpdate := item.GetLastUpdateTime()
	if lastUpdate.Year() > 1 {
		metaParts = append(metaParts, "updated: " + formatRelativeTime(lastUpdate))
	}
	
	if len(metaParts) > 0 {
		metadata := metadataStyle.Render(strings.Join(metaParts, " • "))
		content = lipgloss.JoinVertical(lipgloss.Left, content, metadata)
	}

	// Render the complete item and let it flow naturally
	fmt.Fprint(w, content)
}

// FancyListView provides a list-based interface for work items
type FancyListView struct {
	dataClient       *data.EnhancedClient
	list             list.Model
	tabs             []Tab
	activeTab        int
	workItems        map[string][]*models.Work
	glamour          *glamour.TermRenderer
	markdownProcessor *renderer.MarkdownProcessor
	showDetail       bool
	showFullPost     bool
	selectedItem     *models.Work
	viewport         viewport.Model   // For scrollable content
	width            int
	height           int
	ready            bool
	keys             FancyKeyMap
	renderCache      map[string]string // Cache rendered markdown
	embeddedCache    map[string]string // Cache embedded content
	embeddingStates  map[string]embeddingState // Track embedding loading states
	lastWidth        int               // Track width changes for cache invalidation
}

// embeddingState tracks the state of embedded content
type embeddingState struct {
	loading  bool
	loaded   bool
	content  string
	workID   string
	spinner  spinner.Model
}

// embeddingLoadedMsg is sent when embedded content finishes loading
type embeddingLoadedMsg struct {
	workID    string
	reference string
	content   string
	err       error
}

type FancyKeyMap struct {
	NextTab       key.Binding
	PrevTab       key.Binding
	ToggleDetail  key.Binding
	ViewFullPost  key.Binding
	Back          key.Binding
	NextItem      key.Binding
	PrevItem      key.Binding
	Quit          key.Binding
}

func DefaultFancyKeyMap() FancyKeyMap {
	return FancyKeyMap{
		NextTab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next tab"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("shift+tab"), 
			key.WithHelp("shift+tab", "prev tab"),
		),
		ToggleDetail: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "toggle detail"),
		),
		ViewFullPost: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "view full post"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back to list"),
		),
		NextItem: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("→", "next item"),
		),
		PrevItem: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("←", "prev item"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// Styles for fancy list with connected tabs
var (
	fancyInactiveTabBorder = fancyTabBorderWithBottom("┴", "─", "┴")
	fancyActiveTabBorder   = fancyTabBorderWithBottom("┘", " ", "└")
	fancyHighlightColor    = lipgloss.AdaptiveColor{Light: "235", Dark: "241"}
	fancyInactiveTabStyle  = lipgloss.NewStyle().Border(fancyInactiveTabBorder, true).BorderForeground(fancyHighlightColor).Padding(0, 1)
	fancyActiveTabStyle    = lipgloss.NewStyle().Border(fancyActiveTabBorder, true).BorderForeground(fancyHighlightColor).Padding(0, 1)
	fancyDocStyle          = lipgloss.NewStyle().Padding(1, 2, 1, 2)
)

// fancyTabBorderWithBottom creates a custom border with specified bottom characters for fancy list
func fancyTabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

func NewFancyListView(dataClient *data.EnhancedClient) *FancyListView {
	// Initialize glamour renderer
	glamourRenderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(60),
	)

	// Create tabs
	tabs := []Tab{
		{Name: "NOW", Schedule: models.ScheduleNow, Active: true},
		{Name: "NEXT", Schedule: models.ScheduleNext, Active: false},
		{Name: "LATER", Schedule: models.ScheduleLater, Active: false},
		{Name: "CLOSED", Schedule: models.ScheduleClosed, Active: false},
	}

	// Create list with custom delegate
	delegate := ItemDelegate{
		showDetail: true,
		glamour:    glamourRenderer,
	}
	
	workList := list.New([]list.Item{}, delegate, 0, 0)
	workList.SetShowStatusBar(false)
	workList.SetShowPagination(true)  // Enable pagination to handle overflow
	workList.SetShowHelp(false)
	workList.SetFilteringEnabled(false)
	workList.Styles.Title = lipgloss.NewStyle() // Remove default title styling

	// Initialize viewport for scrollable content
	vp := viewport.New(80, 24)
	vp.KeyMap = viewport.KeyMap{
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", " "),
			key.WithHelp("pgdn/space", "page down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "½ page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "½ page down"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
	}


	// Initialize markdown processor with the data client's work directory
	markdownProcessor := renderer.NewMarkdownProcessor(dataClient.GetLocalWorkDir())

	return &FancyListView{
		dataClient:        dataClient,
		list:              workList,
		tabs:              tabs,
		activeTab:         0,
		workItems:         make(map[string][]*models.Work),
		glamour:           glamourRenderer,
		markdownProcessor: markdownProcessor,
		showDetail:        true,
		showFullPost:      false,
		selectedItem:      nil,
		viewport:          vp,
		keys:              DefaultFancyKeyMap(),
		renderCache:       make(map[string]string),
		embeddedCache:     make(map[string]string),
		embeddingStates:   make(map[string]embeddingState),
		lastWidth:         0,
	}
}

func (f *FancyListView) Init() tea.Cmd {
	return f.loadWorkItems()
}

func (f *FancyListView) loadWorkItems() tea.Cmd {
	return tea.Batch(
		f.loadScheduleItems(models.ScheduleNow),
		f.loadScheduleItems(models.ScheduleNext),
		f.loadScheduleItems(models.ScheduleLater),
		f.loadScheduleItems(models.ScheduleClosed),
	)
}

func (f *FancyListView) loadScheduleItems(schedule string) tea.Cmd {
	return func() tea.Msg {
		items, err := f.dataClient.GetWorkBySchedule(schedule)
		if err != nil {
			// Debug: log the error
			// log.Printf("Error loading %s items: %v", schedule, err)
			return errMsg{err: err}
		}
		// Debug: log successful load
		// log.Printf("Loaded %d items for %s", len(items), schedule)
		return scheduleItemsLoadedMsg{schedule: schedule, items: items}
	}
}

func (f *FancyListView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Always update dimensions
		f.width = msg.Width
		f.height = msg.Height
		
		if f.showFullPost {
			// Update viewport size for full post or embedded view with proper margins
			viewportHeight := f.height - 6 // Reserve space for pagination, help, and margins
			viewportWidth := f.width - 4   // Reserve space for side margins
			
			// Ensure minimum sizes
			if viewportHeight < 5 {
				viewportHeight = 5
			}
			if viewportWidth < 20 {
				viewportWidth = 20
			}
			
			// Update viewport dimensions immediately
			if f.showFullPost {
				f.viewport.Width = viewportWidth
				f.viewport.Height = viewportHeight
				
				// Force re-render content if we have a selected item and width changed significantly
				if f.selectedItem != nil && abs(f.width - f.lastWidth) > 5 {
					// Clear cache to force re-render with new width
					f.renderCache = make(map[string]string)
					f.embeddingStates = make(map[string]embeddingState) // Clear embedding states on resize
					f.updateViewportContent()
				}
			}
		} else {
			// Calculate available space for list view with better margins
			tabHeight := 3      // Space for tabs and borders
			helpHeight := 2     // Space for help text
			borderHeight := 2   // Space for list borders
			
			listWidth := msg.Width - 6   // Better margin calculation
			listHeight := msg.Height - tabHeight - helpHeight - borderHeight - 1
			
			// Ensure minimum sizes
			if listWidth < 20 {
				listWidth = 20
			}
			if listHeight < 5 {
				listHeight = 5
			}
			
			f.list.SetSize(listWidth, listHeight)
		}
		
		// Update last width for cache invalidation
		f.lastWidth = f.width
		
		if !f.ready {
			f.ready = true
			// Load data synchronously when window size is set
			schedules := []string{models.ScheduleNow, models.ScheduleNext, models.ScheduleLater, models.ScheduleClosed}
			for _, schedule := range schedules {
				if items, err := f.dataClient.GetWorkBySchedule(schedule); err == nil {
					f.workItems[schedule] = items
				}
			}
			f.updateListItems()
		}

	case scheduleItemsLoadedMsg:
		f.workItems[msg.schedule] = msg.items
		if !f.ready {
			f.ready = true
			// Update list immediately when we become ready
			f.updateListItems()
		} else if msg.schedule == f.getCurrentSchedule() {
			f.updateListItems()
		}

	case embeddingLoadedMsg:
		// Handle async embedding content loading with spinners
		if f.selectedItem != nil && f.selectedItem.ID == msg.workID {
			// Update the embedding state
			if state, exists := f.embeddingStates[msg.reference]; exists {
				state.loaded = true
				state.loading = false
				state.content = msg.content
				f.embeddingStates[msg.reference] = state
			}
			
			// Store the loaded content in cache
			if msg.content != "" {
				f.embeddedCache[msg.reference] = msg.content
			}
			
			// Clear render cache to force regeneration
			f.renderCache = make(map[string]string)
			
			// Regenerate viewport content with the new embedding
			f.updateViewportContentWithEmbeddings()
		}

	case tea.KeyMsg:
		// Handle global quit key first
		if key.Matches(msg, f.keys.Quit) {
			return f, tea.Quit
		}

		if f.showFullPost {
			// Full post view navigation
			switch {
			case key.Matches(msg, f.keys.Back):
				f.showFullPost = false
				f.selectedItem = nil
			case key.Matches(msg, f.keys.NextItem):
				f.navigateToNextItem()
				f.updateViewportContent() // Update viewport with new content
			case key.Matches(msg, f.keys.PrevItem):
				f.navigateToPrevItem()
				f.updateViewportContent() // Update viewport with new content
			default:
				// Let viewport handle scrolling keys
				f.viewport, cmd = f.viewport.Update(msg)
			}
		} else {
			// List view navigation
			switch {
			case key.Matches(msg, f.keys.NextTab):
				f.nextTab()
				f.updateListItems()
			case key.Matches(msg, f.keys.PrevTab):
				f.prevTab()
				f.updateListItems()
			case key.Matches(msg, f.keys.ToggleDetail):
				f.showDetail = !f.showDetail
				f.updateDelegate()
			case key.Matches(msg, f.keys.ViewFullPost):
				if selectedItem := f.list.SelectedItem(); selectedItem != nil {
					if workItem, ok := selectedItem.(WorkItem); ok {
						f.selectedItem = workItem.Work
						f.showFullPost = true
						f.updateViewportContent() // Load content into viewport
						
						// Auto-start loading embeddings if the item has them
						if f.hasEmbeddings(workItem.Work) {
							return f, f.startAutoEmbeddingLoad()
						}
					}
				}
			default:
				// Let the list handle up/down arrow keys and other navigation
				f.list, cmd = f.list.Update(msg)
			}
		}
	}

	// Update spinners for loading embeddings
	var spinnerCmds []tea.Cmd
	for ref, state := range f.embeddingStates {
		if state.loading {
			var spinnerCmd tea.Cmd
			state.spinner, spinnerCmd = state.spinner.Update(msg)
			f.embeddingStates[ref] = state
			if spinnerCmd != nil {
				spinnerCmds = append(spinnerCmds, spinnerCmd)
			}
		}
	}
	
	if len(spinnerCmds) > 0 {
		if cmd != nil {
			spinnerCmds = append(spinnerCmds, cmd)
		}
		return f, tea.Batch(spinnerCmds...)
	}

	return f, cmd
}

func (f *FancyListView) View() string {
	if !f.ready {
		return "Loading work items... (press any key if stuck)"
	}

	if f.showFullPost && f.selectedItem != nil {
		return f.renderFullPost()
	}


	// Render connected tab bar and list
	tabBar := f.renderConnectedTabBar()
	listContent := f.renderConnectedList()
	help := f.renderHelp()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		tabBar,
		listContent,
		help,
	)
}

// renderConnectedTabBar creates the top tab bar that connects to list content
func (f *FancyListView) renderConnectedTabBar() string {
	var renderedTabs []string
	
	for i, tab := range f.tabs {
		var style lipgloss.Style
		isActive := i == f.activeTab
		if isActive {
			style = fancyActiveTabStyle.Copy()
		} else {
			style = fancyInactiveTabStyle.Copy()
		}
		
		// Add item count and unicode symbols
		count := len(f.workItems[tab.Schedule])
		var symbol string
		switch tab.Schedule {
		case models.ScheduleNow:
			symbol = "●"
		case models.ScheduleNext:
			symbol = "○"
		case models.ScheduleLater:
			symbol = "⊖"
		case models.ScheduleClosed:
			symbol = "✓"
		}
		
		tabText := fmt.Sprintf("%s %s (%d)", symbol, tab.Name, count)
		renderedTabs = append(renderedTabs, style.Render(tabText))
	}
	
	// Join tabs
	tabRow := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	
	// Add underline that extends across terminal with same color as tab borders
	tabWidth := lipgloss.Width(tabRow)
	underlineWidth := f.width - 4
	if underlineWidth > tabWidth {
		underlineStyle := lipgloss.NewStyle().Foreground(fancyHighlightColor)
		underline := underlineStyle.Render(strings.Repeat("─", underlineWidth-tabWidth))
		tabRow = lipgloss.JoinHorizontal(lipgloss.Bottom, tabRow, underline)
	}
	
	return lipgloss.NewStyle().
		Padding(0, 1).
		Render(tabRow)
}

// renderConnectedList renders list content that connects directly to the tabs
func (f *FancyListView) renderConnectedList() string {
	// Create constrained list container to prevent overflow
	maxHeight := f.height - 8 // Reserve space for tabs and help
	if maxHeight < 5 {
		maxHeight = 5
	}
	
	listStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(fancyHighlightColor).
		BorderTop(false).  // No top border to connect with tabs
		Padding(1, 2).     // Normal padding
		Width(f.width - 8). // More conservative width
		MaxHeight(maxHeight) // Constrain height to prevent overflow
	
	return listStyle.Render(f.list.View())
}

func (f *FancyListView) renderPagination() string {
	if !f.showFullPost || f.selectedItem == nil {
		return ""
	}

	schedule := f.getCurrentSchedule()
	items := f.workItems[schedule]
	itemCount := len(items)

	if itemCount <= 1 {
		return "" // No pagination needed for 0 or 1 items
	}

	// Find current item index
	currentIndex := 0
	if f.selectedItem != nil {
		for i, item := range items {
			if item.ID == f.selectedItem.ID {
				currentIndex = i
				break
			}
		}
	}

	// Style similar to bubbletea paginator
	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"})
	
	activeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).
		Bold(true)

	// Create dot-based pagination
	var dots []string
	for i := 0; i < itemCount; i++ {
		if i == currentIndex {
			dots = append(dots, activeStyle.Render("●"))
		} else {
			dots = append(dots, normalStyle.Render("○"))
		}
	}

	// Add numerical indicator
	pageInfo := fmt.Sprintf("%d/%d", currentIndex+1, itemCount)
	pageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "240", Dark: "245"}).
		Padding(0, 1)

	// Combine dots and page info
	pagination := strings.Join(dots, " ") + " " + pageStyle.Render(pageInfo)

	// Center the pagination with current width
	paginationStyle := lipgloss.NewStyle().
		Width(f.width).
		Align(lipgloss.Center).
		Foreground(lipgloss.Color("241"))

	return paginationStyle.Render(pagination)
}

func (f *FancyListView) renderHelp() string {
	var helpText string
	if f.showFullPost {
		schedule := f.getCurrentSchedule()
		itemCount := len(f.workItems[schedule])
		if itemCount > 1 {
			helpText = "←/→: navigate items • esc: back • q: quit"
		} else {
			helpText = "esc: back • q: quit"
		}
	} else {
		helpText = "tab/shift+tab: switch tabs • ↑/↓: navigate • enter: view full • d: detail • q: quit"
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(1, 2).
		Render(helpText)
}

func (f *FancyListView) renderScrollableHelp() string {
	schedule := f.getCurrentSchedule()
	itemCount := len(f.workItems[schedule])
	
	var helpText string
	if itemCount > 1 {
		helpText = "↑/↓/j/k: scroll • space/pgdn: page down • pgup: page up • ←/→: items • esc: back • q: quit"
	} else {
		helpText = "↑/↓/j/k: scroll • space/pgdn: page down • pgup: page up • esc: back • q: quit"
	}
	
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(1, 2).
		Render(helpText)
}

func (f *FancyListView) renderFullPost() string {
	if f.selectedItem == nil {
		return "No item selected"
	}

	// Use viewport for scrollable content - no truncation!
	content := f.viewport.View()

	// Build pagination indicator
	pagination := f.renderPagination()

	// Build help text with scrolling instructions
	help := f.renderScrollableHelp()

	// Join viewport content, pagination, and help
	return content + "\n" + pagination + "\n" + help
}

func (f *FancyListView) updateListItems() {
	schedule := f.getCurrentSchedule()
	items := f.workItems[schedule]
	
	var listItems []list.Item
	for _, item := range items {
		// All filtering is now done at the data layer
		listItems = append(listItems, WorkItem{item})
	}
	
	f.list.SetItems(listItems)
	
	// Remove redundant title - tabs already show the schedule
	f.list.Title = ""
}

func (f *FancyListView) updateDelegate() {
	delegate := ItemDelegate{
		showDetail: f.showDetail,
		glamour:    f.glamour,
	}
	f.list.SetDelegate(delegate)
}

func (f *FancyListView) nextTab() {
	f.activeTab = (f.activeTab + 1) % len(f.tabs)
}

func (f *FancyListView) prevTab() {
	f.activeTab = (f.activeTab - 1 + len(f.tabs)) % len(f.tabs)
}

func (f *FancyListView) getCurrentSchedule() string {
	if f.activeTab < len(f.tabs) {
		return f.tabs[f.activeTab].Schedule
	}
	return models.ScheduleNow
}

func (f *FancyListView) navigateToNextItem() {
	schedule := f.getCurrentSchedule()
	items := f.workItems[schedule]
	if len(items) == 0 {
		return
	}

	if len(items) == 1 {
		// Only one item, nothing to navigate to
		return
	}

	// Find current item index
	currentIndex := 0
	if f.selectedItem != nil {
		for i, item := range items {
			if item.ID == f.selectedItem.ID {
				currentIndex = i
				break
			}
		}
	}

	// Move to next item (wrap around)
	nextIndex := (currentIndex + 1) % len(items)
	f.selectedItem = items[nextIndex]
}

func (f *FancyListView) navigateToPrevItem() {
	schedule := f.getCurrentSchedule()
	items := f.workItems[schedule]
	if len(items) == 0 {
		return
	}

	if len(items) == 1 {
		// Only one item, nothing to navigate to
		return
	}

	// Find current item index
	currentIndex := 0
	if f.selectedItem != nil {
		for i, item := range items {
			if item.ID == f.selectedItem.ID {
				currentIndex = i
				break
			}
		}
	}

	// Move to previous item (wrap around)
	prevIndex := (currentIndex - 1 + len(items)) % len(items)
	f.selectedItem = items[prevIndex]
}

func (f *FancyListView) updateViewportContent() {
	if f.selectedItem == nil {
		return
	}

	item := f.selectedItem

	// Check if we need to invalidate cache due to width change
	if f.width != f.lastWidth {
		f.renderCache = make(map[string]string) // Clear cache on width change
		f.lastWidth = f.width
	}

	// Generate cache key
	glamourWidth := f.viewport.Width - 4
	if glamourWidth < 20 {
		glamourWidth = 20
	}
	cacheKey := fmt.Sprintf("%s_%d", item.ID, glamourWidth)

	// Check cache first, but skip cache if embeddings are loading for this item
	hasLoadingEmbeddings := false
	for _, state := range f.embeddingStates {
		if state.workID == item.ID && state.loading {
			hasLoadingEmbeddings = true
			break
		}
	}
	
	renderedContent, exists := f.renderCache[cacheKey]
	if !exists || hasLoadingEmbeddings {
		// For Work items, we combine Description and Content
		var fullContent string
		if item.Description != "" {
			fullContent = "# " + item.Title + "\n\n" + item.Description
			if item.Content != "" {
				fullContent = fullContent + "\n\n" + item.Content
			}
		} else if item.Content != "" {
			fullContent = "# " + item.Title + "\n\n" + item.Content
		} else {
			fullContent = "# " + item.Title + "\n\nNo detailed content available."
		}
		
		if fullContent != "" {
			var processedContent string
			
			if hasLoadingEmbeddings {
				// Collect loaded embeddings and loading states for live rendering
				loadedEmbeddings := make(map[string]string)
				loadingStates := make(map[string]string)
				
				for ref, state := range f.embeddingStates {
					if state.workID == item.ID {
						if state.loaded {
							loadedEmbeddings[ref] = state.content
						} else if state.loading {
							loadingStates[ref] = state.spinner.View()
						}
					}
				}
				
				// Use async processing to show loading states
				processedContent = f.markdownProcessor.ProcessWithAsyncEmbeddings(fullContent, loadedEmbeddings, loadingStates)
			} else {
				// Use lightweight processing for initial fast rendering
				processedContent = f.markdownProcessor.ProcessForLightRendering(fullContent)
			}
			
			// Render with Glamour using cached renderer if possible
			var renderer *glamour.TermRenderer
			var err error
			
			// Use existing glamour renderer if available and width matches
			if f.glamour != nil {
				renderer = f.glamour
			} else {
				renderer, err = glamour.NewTermRenderer(
					glamour.WithAutoStyle(),
					glamour.WithWordWrap(glamourWidth),
				)
			}
			
			if err == nil {
				if rendered, err := renderer.Render(processedContent); err == nil {
					renderedContent = rendered
				} else {
					renderedContent = processedContent // Fallback to processed content
				}
			} else {
				renderedContent = processedContent // Fallback to processed content
			}
		} else {
			renderedContent = "No content available"
		}
		
		// Cache the result only if we're not loading embeddings
		if !hasLoadingEmbeddings {
			f.renderCache[cacheKey] = renderedContent
		}
	}

	// Set content in viewport (no truncation - full scrollable content)
	f.viewport.SetContent(renderedContent)
	f.viewport.GotoTop() // Start at top when switching items
}


// Helper functions
func getWorkItemIcon(itemType string) string {
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

func getScheduleIcon(schedule string) string {
	switch schedule {
	case models.ScheduleNow:
		return "🔥"
	case models.ScheduleNext:
		return "⏳"
	case models.ScheduleLater:
		return "📅"
	default:
		return "📋"
	}
}

// extractFirstTwoLines gets just the first 2 content lines
func extractFirstTwoLines(content string) string {
	lines := strings.Split(content, "\n")
	var contentLines []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip empty lines and headers
		if line != "" && !strings.HasPrefix(line, "#") {
			// Clean markdown formatting
			cleanLine := strings.ReplaceAll(line, "**", "")
			cleanLine = strings.ReplaceAll(cleanLine, "*", "")
			cleanLine = strings.ReplaceAll(cleanLine, "_", "")
			cleanLine = strings.TrimSpace(cleanLine)
			
			if cleanLine != "" {
				contentLines = append(contentLines, cleanLine)
				if len(contentLines) >= 2 {
					break
				}
			}
		}
	}
	
	if len(contentLines) == 0 {
		return ""
	}
	
	result := strings.Join(contentLines, " ")
	// Limit total length
	if len(result) > 120 {
		result = result[:120] + "..."
	}
	
	return result
}

// abs returns absolute value of integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// formatRelativeTime formats a time as a relative string (e.g., "2h ago", "3d ago")
func formatRelativeTime(t time.Time) string {
	duration := time.Since(t)
	
	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", minutes)
	case duration < 24*time.Hour:
		hours := int(duration.Hours())
		if hours == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", hours)
	case duration < 7*24*time.Hour:
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", days)
	case duration < 30*24*time.Hour:
		weeks := int(duration.Hours() / (24 * 7))
		if weeks == 1 {
			return "1w ago"
		}
		return fmt.Sprintf("%dw ago", weeks)
	case duration < 365*24*time.Hour:
		months := int(duration.Hours() / (24 * 30))
		if months == 1 {
			return "1mo ago"
		}
		return fmt.Sprintf("%dmo ago", months)
	default:
		years := int(duration.Hours() / (24 * 365))
		if years == 1 {
			return "1y ago"
		}
		return fmt.Sprintf("%dy ago", years)
	}
}

func extractOverview(content string) string {
	lines := strings.Split(content, "\n")
	var overview []string
	
	// Look for overview section or take first few paragraphs
	inOverview := false
	paragraphCount := 0
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Check for overview section
		if strings.Contains(strings.ToLower(line), "overview") && strings.HasPrefix(line, "#") {
			inOverview = true
			continue
		}
		
		// Stop at next section if we were in overview
		if inOverview && strings.HasPrefix(line, "#") {
			break
		}
		
		// Add content
		if inOverview || (!inOverview && line != "") {
			overview = append(overview, line)
			if line != "" && !strings.HasPrefix(line, "#") {
				paragraphCount++
				if paragraphCount >= 3 && !inOverview {
					break // Limit to first 3 paragraphs if no overview section
				}
			}
		}
	}
	
	result := strings.Join(overview, "\n")
	// Limit length
	if len(result) > 300 {
		result = result[:300] + "..."
	}
	
	return result
}

// hasEmbeddings checks if a Work item contains embedded content references
func (f *FancyListView) hasEmbeddings(work *models.Work) bool {
	content := work.Content
	if content == "" {
		content = work.Description
	}
	hasEmbeds := strings.Contains(content, "![[")
	// fmt.Printf("DEBUG hasEmbeddings: work=%s, contentLen=%d, hasEmbeds=%t\n", work.Title, len(content), hasEmbeds)
	return hasEmbeds
}

// startAutoEmbeddingLoad automatically starts loading embeddings with spinners
func (f *FancyListView) startAutoEmbeddingLoad() tea.Cmd {
	if f.selectedItem == nil {
		return nil
	}
	
	// Extract embedding references from the work item content
	content := f.selectedItem.Content
	if content == "" {
		content = f.selectedItem.Description
	}
	
	// Debug: Check what content we're working with
	// fmt.Printf("DEBUG: Content length: %d, Has embedding: %t\n", len(content), strings.Contains(content, "![["))
	
	references := f.markdownProcessor.ExtractEmbeddingReferences(content)
	// fmt.Printf("DEBUG: Found %d embedding references: %v\n", len(references), references)
	if len(references) == 0 {
		return nil
	}
	
	// Start loading all embeddings with spinners
	var cmds []tea.Cmd
	for _, ref := range references {
		// Skip if already loaded
		if state, exists := f.embeddingStates[ref]; exists && (state.loaded || state.loading) {
			continue
		}
		
		// Create and start spinner
		s := spinner.New()
		s.Spinner = spinner.Dot
		s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		
		// Mark as loading with spinner
		f.embeddingStates[ref] = embeddingState{
			loading: true,
			loaded:  false,
			workID:  f.selectedItem.ID,
			spinner: s,
		}
		
		// Create async loading command
		cmds = append(cmds, s.Tick)
		cmds = append(cmds, f.loadSingleEmbedding(f.selectedItem.ID, ref))
	}
	
	// Clear render cache to show loading state
	f.renderCache = make(map[string]string)
	
	return tea.Batch(cmds...)
}

// loadSingleEmbedding loads a single embedding asynchronously
func (f *FancyListView) loadSingleEmbedding(workID, reference string) tea.Cmd {
	return func() tea.Msg {
		// Add a small delay to show the spinner
		time.Sleep(100 * time.Millisecond)
		content := f.markdownProcessor.ResolveReference(reference)
		return embeddingLoadedMsg{
			workID:    workID,
			reference: reference,
			content:   content,
			err:       nil,
		}
	}
}

// updateViewportContentWithEmbeddings updates viewport content with loaded embeddings
func (f *FancyListView) updateViewportContentWithEmbeddings() {
	if f.selectedItem == nil {
		return
	}
	
	item := f.selectedItem
	
	// Generate cache key including embedding state
	glamourWidth := f.viewport.Width - 4
	if glamourWidth < 20 {
		glamourWidth = 20
	}
	
	// Create content with loaded embeddings
	var fullContent string
	if item.Description != "" {
		fullContent = "# " + item.Title + "\n\n" + item.Description
		if item.Content != "" {
			fullContent = fullContent + "\n\n" + item.Content
		}
	} else if item.Content != "" {
		fullContent = "# " + item.Title + "\n\n" + item.Content
	} else {
		fullContent = "# " + item.Title + "\n\nNo detailed content available."
	}
	
	// Collect loaded embeddings and loading states
	loadedEmbeddings := make(map[string]string)
	loadingStates := make(map[string]string)
	
	for ref, state := range f.embeddingStates {
		if state.workID == item.ID {
			if state.loaded {
				loadedEmbeddings[ref] = state.content
			} else if state.loading {
				loadingStates[ref] = state.spinner.View()
			}
		}
	}
	
	// Process content with loaded embeddings and loading spinners
	processedContent := f.markdownProcessor.ProcessWithAsyncEmbeddings(fullContent, loadedEmbeddings, loadingStates)
	
	// Render with Glamour
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(glamourWidth),
	)
	
	var renderedContent string
	if err == nil {
		if rendered, err := renderer.Render(processedContent); err == nil {
			renderedContent = rendered
		} else {
			renderedContent = processedContent
		}
	} else {
		renderedContent = processedContent
	}
	
	// Update viewport content
	f.viewport.SetContent(renderedContent)
}