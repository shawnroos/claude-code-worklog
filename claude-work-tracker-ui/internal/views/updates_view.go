package views

import (
	"fmt"
	"strings"

	"claude-work-tracker-ui/internal/models"
	"claude-work-tracker-ui/internal/renderer"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	updateHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Padding(0, 1)

	updateTimeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)

	updateContentStyle = lipgloss.NewStyle().
		Padding(0, 2)
)

type UpdatesView struct {
	updates   []models.Update
	viewport  viewport.Model
	width     int
	height    int
	ready     bool
	processor *renderer.MarkdownProcessor
	baseDir   string
}

func NewUpdatesView() UpdatesView {
	baseDir := ".claude-work"
	return UpdatesView{
		updates:   []models.Update{},
		processor: renderer.NewMarkdownProcessor(baseDir),
		baseDir:   baseDir,
	}
}

func (v UpdatesView) Init() tea.Cmd {
	return nil
}

func (v UpdatesView) Update(msg tea.Msg) (UpdatesView, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !v.ready {
			v.viewport = viewport.New(msg.Width-4, msg.Height-6)
			v.viewport.YPosition = 0
			v.viewport.HighPerformanceRendering = false
			v.ready = true
		} else {
			v.viewport.Width = msg.Width - 4
			v.viewport.Height = msg.Height - 6
		}
		v.width = msg.Width
		v.height = msg.Height
		v.renderContent()

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			v.viewport.LineUp(1)
		case "down", "j":
			v.viewport.LineDown(1)
		case "pgup":
			v.viewport.ViewUp()
		case "pgdn":
			v.viewport.ViewDown()
		case "home":
			v.viewport.GotoTop()
		case "end":
			v.viewport.GotoBottom()
		}
	}

	v.viewport, cmd = v.viewport.Update(msg)
	return v, cmd
}

func (v UpdatesView) View() string {
	if !v.ready {
		return "\n  Initializing updates view..."
	}

	header := updateHeaderStyle.Render("📝 Updates Timeline")
	
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(fmt.Sprintf("\n %d updates • %d%% ", len(v.updates), int(v.viewport.ScrollPercent()*100)))

	return fmt.Sprintf("%s\n\n%s%s", header, v.viewport.View(), footer)
}

func (v *UpdatesView) SetUpdates(updates []models.Update) {
	v.updates = updates
	v.renderContent()
}

func (v *UpdatesView) renderContent() {
	if !v.ready {
		return
	}

	var content strings.Builder

	if len(v.updates) == 0 {
		content.WriteString("\n  No updates yet.\n")
	} else {
		for i, update := range v.updates {
			if i > 0 {
				content.WriteString("\n" + strings.Repeat("─", v.width-4) + "\n\n")
			}

			// Update header with time
			timeStr := update.Timestamp.Format("Jan 2, 15:04")
			header := fmt.Sprintf("Update #%d • %s", i+1, updateTimeStyle.Render(timeStr))
			content.WriteString(header + "\n\n")

			// Process and render markdown content
			updateContent := fmt.Sprintf("**%s**\n\n%s", update.Title, update.Summary)
			processed := v.processor.ProcessForLightRendering(updateContent)
			content.WriteString(updateContentStyle.Render(processed))

			// Add task changes if present
			if len(update.TasksCompleted) > 0 {
				content.WriteString("\n\n✅ Completed: " + strings.Join(update.TasksCompleted, ", "))
			}
			if len(update.TasksAdded) > 0 {
				content.WriteString("\n\n➕ Added: " + strings.Join(update.TasksAdded, ", "))
			}
		}
	}

	v.viewport.SetContent(content.String())
}

// Helper to create updates from a Work item
func (v *UpdatesView) LoadFromWork(work *models.Work) {
	// In a real implementation, this would load updates from the updates file
	// For now, we'll create a sample update
	if work.UpdatesRef != "" {
		sampleUpdate := models.Update{
			ID:        "update-001",
			WorkID:    work.ID,
			Timestamp: work.UpdatedAt,
			Title:     "Initial Implementation Progress",
			Summary:   "Work item created and initial tasks defined. TUI modification is progressing well with core implementation complete.",
			Author:    "Claude",
			UpdateType: "automatic",
			TasksCompleted: []string{"Modified WorkItem struct", "Updated data loading", "Fixed rendering"},
		}
		v.SetUpdates([]models.Update{sampleUpdate})
	}
}