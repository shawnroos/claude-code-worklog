package main

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"claude-work-tracker-ui/internal/app"
)

func main() {
	// Create the application
	app := app.NewApp()
	
	// Create the program
	p := tea.NewProgram(app, tea.WithAltScreen())
	
	// Run the program
	if _, err := p.Run(); err != nil {
		log.Fatal("Error running program:", err)
		os.Exit(1)
	}
}