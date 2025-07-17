package app

import (
	"log"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"claude-work-tracker-ui/internal/data"
	"claude-work-tracker-ui/internal/models"
	"claude-work-tracker-ui/internal/sync"
	"claude-work-tracker-ui/internal/views"
)

type ViewType int

const (
	DashboardView ViewType = iota
	WorkItemsView
	TabbedWorkView
	FancyListView
	ReferencesView
	FutureWorkView
)

// RefreshMsg is sent when UI needs to refresh due to file changes
type RefreshMsg struct {
	EventType string
	Item      *models.MarkdownWorkItem
}

type App struct {
	dataClient       *data.Client
	enhancedClient   *data.EnhancedClient
	syncCoordinator  *sync.SyncCoordinator
	currentView      ViewType
	sidebar          *views.SidebarModel
	dashboard        *views.DashboardModel
	workItems        *views.WorkItemsModel
	enhancedWorkItems *views.EnhancedWorkItemsModel
	tabbedWorkView   *views.TabbedWorkView
	fancyListView    *views.FancyListView
	references       *views.ReferencesModel
	futureWork       *views.FutureWorkModel
	width            int
	height           int
	quitting         bool
	sidebarWidth     int
	useEnhanced      bool
	syncEnabled      bool
	useTabbedView    bool
	useFancyList     bool
}

var (
	appStyle = lipgloss.NewStyle()

	contentStyle = lipgloss.NewStyle().
			Padding(1, 2)

	layoutStyle = lipgloss.NewStyle()
)

func NewApp() *App {
	dataClient := data.NewClient()
	enhancedClient := data.NewEnhancedClient()
	sidebar := views.NewSidebarModel()
	dashboard := views.NewDashboardModel(dataClient)
	workItems := views.NewWorkItemsModel(dataClient)
	enhancedWorkItems := views.NewEnhancedWorkItemsModel(enhancedClient)
	tabbedWorkView := views.NewTabbedWorkView(enhancedClient)
	fancyListView := views.NewFancyListView(enhancedClient)
	references := views.NewReferencesModel(dataClient)
	futureWork := views.NewFutureWorkModel(dataClient)

	app := &App{
		dataClient:        dataClient,
		enhancedClient:    enhancedClient,
		currentView:       FancyListView, // Default to fancy list view
		sidebar:           sidebar,
		dashboard:         dashboard,
		workItems:         workItems,
		enhancedWorkItems: enhancedWorkItems,
		tabbedWorkView:    tabbedWorkView,
		fancyListView:     fancyListView,
		references:        references,
		futureWork:        futureWork,
		sidebarWidth:      25,
		useEnhanced:       true, // Default to enhanced view
		syncEnabled:       true, // Enable real-time sync by default
		useTabbedView:     false, // Disable tabbed interface
		useFancyList:      true,  // Enable fancy list interface
	}

	// Initialize sync coordinator if enabled
	if app.syncEnabled {
		// Use the enhanced client's work directory (which is now always at project root)
		watchDir := enhancedClient.GetLocalWorkDir()

		syncCoordinator, err := sync.NewSyncCoordinator(watchDir, enhancedClient)
		if err != nil {
			log.Printf("Failed to initialize sync coordinator: %v", err)
			app.syncEnabled = false
		} else {
			app.syncCoordinator = syncCoordinator
			
			// Set up UI update callback
			syncCoordinator.SetUICallback(app.handleSyncEvent)
			
			// Start the sync coordinator
			if err := syncCoordinator.Start(); err != nil {
				log.Printf("Failed to start sync coordinator: %v", err)
				app.syncEnabled = false
			} else {
				log.Println("Real-time sync enabled")
			}
		}
	}

	return app
}

func (a *App) Init() tea.Cmd {
	if a.useFancyList {
		return a.fancyListView.Init()
	}
	if a.useTabbedView {
		return a.tabbedWorkView.Init()
	}
	return a.dashboard.Init()
}

// handleSyncEvent is called when files change and converts to bubbletea messages
func (a *App) handleSyncEvent(eventType string, item *models.MarkdownWorkItem) {
	// This would typically use a bubbletea Program.Send method
	// For now, we'll handle it synchronously in the UI
	log.Printf("Sync event: %s", eventType)
	// Note: In a real implementation, we'd need to pass the tea.Program
	// and use program.Send(RefreshMsg{EventType: eventType, Item: item})
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case RefreshMsg:
		// Handle real-time refresh events
		log.Printf("UI refresh triggered: %s", msg.EventType)
		
		// Refresh the current view
		switch a.currentView {
		case WorkItemsView:
			if a.useEnhanced {
				cmds = append(cmds, a.enhancedWorkItems.Init())
			}
		case TabbedWorkView:
			cmds = append(cmds, a.tabbedWorkView.Init())
		case FancyListView:
			cmds = append(cmds, a.fancyListView.Init())
		case DashboardView:
			cmds = append(cmds, a.dashboard.Init())
		}
		
		return a, tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		
		// Update sidebar
		a.sidebar, _ = a.sidebar.Update(msg)
		
		// Update content views with adjusted width
		contentMsg := tea.WindowSizeMsg{
			Width:  msg.Width - a.sidebarWidth,
			Height: msg.Height,
		}
		
		a.dashboard, _ = a.dashboard.Update(contentMsg)
		a.workItems, _ = a.workItems.Update(contentMsg)
		a.enhancedWorkItems, _ = a.enhancedWorkItems.Update(contentMsg)
		if model, cmd := a.tabbedWorkView.Update(msg); model != nil {
			a.tabbedWorkView = model.(*views.TabbedWorkView)
			_ = cmd
		}
		if model, cmd := a.fancyListView.Update(msg); model != nil {
			a.fancyListView = model.(*views.FancyListView)
			_ = cmd
		}
		a.references, _ = a.references.Update(contentMsg)
		a.futureWork, _ = a.futureWork.Update(contentMsg)
		
		return a, nil

	case views.SidebarSelectMsg:
		// Handle sidebar selection
		a.currentView = ViewType(msg.ViewType)
		a.sidebar.SetActiveView(msg.ViewType)
		
		// Initialize the selected view
		switch a.currentView {
		case WorkItemsView:
			if a.useEnhanced {
				cmds = append(cmds, a.enhancedWorkItems.Init())
			} else {
				cmds = append(cmds, a.workItems.Init())
			}
		case TabbedWorkView:
			cmds = append(cmds, a.tabbedWorkView.Init())
		case FancyListView:
			cmds = append(cmds, a.fancyListView.Init())
		case ReferencesView:
			cmds = append(cmds, a.references.Init())
		case FutureWorkView:
			cmds = append(cmds, a.futureWork.Init())
		}
		
		return a, tea.Batch(cmds...)

	case tea.KeyMsg:
		// Handle fancy list view first if enabled to allow it to handle keys
		if a.useFancyList {
			// Forward directly to fancy list view, no sidebar needed
			if model, cmd := a.fancyListView.Update(msg); model != nil {
				a.fancyListView = model.(*views.FancyListView)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
			return a, tea.Batch(cmds...)
		}
		
		// Default app-level key handling for non-fancy-list views
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			a.quitting = true
			
			// Clean up sync coordinator
			if a.syncCoordinator != nil {
				a.syncCoordinator.Stop()
			}
			
			return a, tea.Quit
		}
		
		// Handle tabbed view if enabled (and not fancy list)
		if a.useTabbedView {
			// Forward directly to tabbed view, no sidebar needed
			if model, cmd := a.tabbedWorkView.Update(msg); model != nil {
				a.tabbedWorkView = model.(*views.TabbedWorkView)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
			return a, tea.Batch(cmds...)
		}
	}

	// Handle sidebar-based views
	if !a.useFancyList && !a.useTabbedView {
		// Update sidebar first
		var sidebarCmd tea.Cmd
		a.sidebar, sidebarCmd = a.sidebar.Update(msg)
		if sidebarCmd != nil {
			cmds = append(cmds, sidebarCmd)
		}

		// Forward to current view
		var viewCmd tea.Cmd
		switch a.currentView {
		case DashboardView:
			a.dashboard, viewCmd = a.dashboard.Update(msg)
		case WorkItemsView:
			if a.useEnhanced {
				a.enhancedWorkItems, viewCmd = a.enhancedWorkItems.Update(msg)
			} else {
				a.workItems, viewCmd = a.workItems.Update(msg)
			}
		case TabbedWorkView:
			if model, cmd := a.tabbedWorkView.Update(msg); model != nil {
				a.tabbedWorkView = model.(*views.TabbedWorkView)
				viewCmd = cmd
			}
		case FancyListView:
			if model, cmd := a.fancyListView.Update(msg); model != nil {
				a.fancyListView = model.(*views.FancyListView)
				viewCmd = cmd
			}
		case ReferencesView:
			a.references, viewCmd = a.references.Update(msg)
		case FutureWorkView:
			a.futureWork, viewCmd = a.futureWork.Update(msg)
		}
		
		if viewCmd != nil {
			cmds = append(cmds, viewCmd)
		}
	}

	return a, tea.Batch(cmds...)
}

func (a *App) View() string {
	if a.quitting {
		return "Goodbye! 👋\n"
	}

	// Use fancy list view if enabled (full screen, no sidebar)
	if a.useFancyList {
		return appStyle.Render(a.fancyListView.View())
	}
	
	// Use tabbed view if enabled (full screen, no sidebar)
	if a.useTabbedView {
		return appStyle.Render(a.tabbedWorkView.View())
	}

	// Traditional sidebar + content layout
	// Get sidebar
	sidebar := a.sidebar.View()

	// Get main content
	var content string
	switch a.currentView {
	case DashboardView:
		content = a.dashboard.View()
	case WorkItemsView:
		if a.useEnhanced {
			content = a.enhancedWorkItems.View()
		} else {
			content = a.workItems.View()
		}
	case TabbedWorkView:
		content = a.tabbedWorkView.View()
	case FancyListView:
		content = a.fancyListView.View()
	case ReferencesView:
		content = a.references.View()
	case FutureWorkView:
		content = a.futureWork.View()
	}

	// Style the main content area
	contentWidth := a.width - a.sidebarWidth
	if contentWidth < 0 {
		contentWidth = 40 // minimum width
	}
	
	styledContent := contentStyle.
		Width(contentWidth - 4).
		Height(a.height - 2).
		Render(content)

	// Combine sidebar and content
	layout := lipgloss.JoinHorizontal(
		lipgloss.Top,
		sidebar,
		styledContent,
	)

	return appStyle.Render(layout)
}