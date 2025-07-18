package app

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"claude-work-tracker-ui/internal/models"
	"claude-work-tracker-ui/internal/storage"
	"claude-work-tracker-ui/internal/views"
)

// CentralizedApp is the main application using external centralized storage
type CentralizedApp struct {
	client          *storage.CentralizedClient
	currentView     ViewType
	fancyListView   *views.FancyListView
	width           int
	height          int
	quitting        bool
	projectSwitcher *ProjectSwitcherModel
	showProjects    bool
}

// NewCentralizedApp creates a new app with centralized storage
func NewCentralizedApp() (*CentralizedApp, error) {
	// Initialize centralized client
	client, err := storage.NewCentralizedClient()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize centralized client: %w", err)
	}

	// Log current project
	project := client.GetCurrentProject()
	log.Printf("Using centralized storage for project: %s (ID: %s)", project.Name, project.ID)
	log.Printf("Work directory: %s", client.GetWorkDir())

	// Create fancy list view with work adapter
	adapter := &CentralizedWorkAdapter{client: client}
	fancyListView := views.NewFancyListViewWithAdapter(adapter)

	// Create project switcher
	projectSwitcher := NewProjectSwitcherModel(client)

	app := &CentralizedApp{
		client:          client,
		currentView:     FancyListView,
		fancyListView:   fancyListView,
		projectSwitcher: projectSwitcher,
		showProjects:    false,
	}

	// Cleanup old repository storage
	if err := client.CleanupOldRepositoryStorage(); err != nil {
		log.Printf("Warning: Could not cleanup old repository storage: %v", err)
	}

	return app, nil
}

func (a *CentralizedApp) Init() tea.Cmd {
	return a.fancyListView.Init()
}

func (a *CentralizedApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		
		// Update views
		if model, cmd := a.fancyListView.Update(msg); model != nil {
			a.fancyListView = model.(*views.FancyListView)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		
		a.projectSwitcher.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		// Global hotkeys
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+p"))):
			// Toggle project switcher
			a.showProjects = !a.showProjects
			if a.showProjects {
				cmds = append(cmds, a.projectSwitcher.Init())
			}
			return a, tea.Batch(cmds...)
			
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			if a.showProjects {
				a.showProjects = false
				return a, nil
			}
			a.quitting = true
			return a, tea.Quit
		}

		// Handle project switcher input
		if a.showProjects {
			m, cmd := a.projectSwitcher.Update(msg)
			a.projectSwitcher = m.(*ProjectSwitcherModel)
			
			// Check if project was selected
			if selectedProject := a.projectSwitcher.GetSelectedProject(); selectedProject != nil {
				// Switch to selected project
				if err := a.client.SwitchProject(selectedProject.ID); err != nil {
					log.Printf("Error switching project: %v", err)
				} else {
					// Recreate views with new project
					adapter := &CentralizedWorkAdapter{client: a.client}
					a.fancyListView = views.NewFancyListViewWithAdapter(adapter)
					cmds = append(cmds, a.fancyListView.Init())
				}
				a.showProjects = false
			}
			
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		} else {
			// Forward to fancy list view
			if model, cmd := a.fancyListView.Update(msg); model != nil {
				a.fancyListView = model.(*views.FancyListView)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}

	default:
		// Forward other messages to current view
		if !a.showProjects {
			if model, cmd := a.fancyListView.Update(msg); model != nil {
				a.fancyListView = model.(*views.FancyListView)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
	}

	return a, tea.Batch(cmds...)
}

func (a *CentralizedApp) View() string {
	if a.quitting {
		return "Goodbye! 👋\n"
	}

	// Show project switcher overlay
	if a.showProjects {
		return a.projectSwitcher.View()
	}

	// Show current view with project info header
	project := a.client.GetCurrentProject()
	
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Width(a.width)
		
	header := headerStyle.Render(fmt.Sprintf("📁 %s • %s", project.Name, project.ActiveBranch))
	
	content := a.fancyListView.View()
	
	return lipgloss.JoinVertical(lipgloss.Top, header, content)
}

// CentralizedWorkAdapter adapts the centralized client to the WorkDataProvider interface
type CentralizedWorkAdapter struct {
	client *storage.CentralizedClient
}

func (a *CentralizedWorkAdapter) GetWorkBySchedule(schedule string) ([]*models.Work, error) {
	return a.client.GetWorkBySchedule(schedule)
}

func (a *CentralizedWorkAdapter) UpdateWorkSchedule(workID, newSchedule string) error {
	// Get the work item
	allWork, err := a.client.GetAllWork()
	if err != nil {
		return err
	}
	
	for _, work := range allWork {
		if work.ID == workID {
			work.Schedule = newSchedule
			return a.client.UpdateWork(work)
		}
	}
	
	return fmt.Errorf("work item not found: %s", workID)
}

func (a *CentralizedWorkAdapter) CompleteWork(workID string) error {
	// Get the work item
	allWork, err := a.client.GetAllWork()
	if err != nil {
		return err
	}
	
	for _, work := range allWork {
		if work.ID == workID {
			work.MarkAsCompleted()
			return a.client.UpdateWork(work)
		}
	}
	
	return fmt.Errorf("work item not found: %s", workID)
}

func (a *CentralizedWorkAdapter) SearchWork(query string) ([]*models.Work, error) {
	// For now, search only in current project
	// TODO: Add cross-project search support
	allWork, err := a.client.GetAllWork()
	if err != nil {
		return nil, err
	}
	
	var results []*models.Work
	for _, work := range allWork {
		if work.MatchesSearch(query) {
			results = append(results, work)
		}
	}
	
	return results, nil
}

// ProjectSwitcherModel allows switching between projects
type ProjectSwitcherModel struct {
	client          *storage.CentralizedClient
	projects        []*storage.Project
	cursor          int
	selectedProject *storage.Project
	width           int
	height          int
}

func NewProjectSwitcherModel(client *storage.CentralizedClient) *ProjectSwitcherModel {
	return &ProjectSwitcherModel{
		client:   client,
		projects: client.GetAllProjects(),
	}
}

func (m *ProjectSwitcherModel) Init() tea.Cmd {
	// Refresh project list
	m.projects = m.client.GetAllProjects()
	return nil
}

func (m *ProjectSwitcherModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.projects)-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor < len(m.projects) {
				m.selectedProject = m.projects[m.cursor]
			}
		case "esc", "q":
			m.selectedProject = nil
		}
	}
	
	return m, nil
}

func (m *ProjectSwitcherModel) View() string {
	if len(m.projects) == 0 {
		return "No projects found"
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		MarginBottom(1)
		
	itemStyle := lipgloss.NewStyle().
		PaddingLeft(2)
		
	selectedStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		Foreground(lipgloss.Color("86")).
		Background(lipgloss.Color("235"))

	var s strings.Builder
	s.WriteString(titleStyle.Render("🗂  Select Project"))
	s.WriteString("\n\n")

	for i, project := range m.projects {
		cursor := "  "
		if i == m.cursor {
			cursor = "▸ "
		}
		
		line := fmt.Sprintf("%s%s", cursor, project.Name)
		if project.ID == m.client.GetCurrentProject().ID {
			line += " (current)"
		}
		
		if i == m.cursor {
			s.WriteString(selectedStyle.Render(line))
		} else {
			s.WriteString(itemStyle.Render(line))
		}
		
		if i < len(m.projects)-1 {
			s.WriteString("\n")
		}
	}
	
	s.WriteString("\n\n")
	s.WriteString(lipgloss.NewStyle().Faint(true).Render("↑/↓: Navigate • Enter: Select • Esc: Cancel"))

	// Center in window
	content := s.String()
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m *ProjectSwitcherModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *ProjectSwitcherModel) GetSelectedProject() *storage.Project {
	selected := m.selectedProject
	m.selectedProject = nil // Reset after retrieval
	return selected
}