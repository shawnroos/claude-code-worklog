package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"claude-work-tracker-ui/internal/app"
)

func main() {
	// Add panic recovery for main
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC in main: %v", r)
			os.Exit(1)
		}
	}()
	
	// Parse command line flags
	_ = flag.Bool("centralized", true, "Use centralized storage (default: true)")
	useLegacy := flag.Bool("legacy", false, "Use legacy repository-based storage")
	flag.Parse()
	
	// Determine which storage mode to use
	var program *tea.Program
	
	if *useLegacy {
		// Use legacy repository-based storage
		log.Println("Using legacy repository-based storage")
		legacyApp := app.NewApp()
		program = tea.NewProgram(legacyApp, tea.WithAltScreen())
	} else {
		// Use new centralized storage (default)
		log.Println("Using centralized external storage")
		centralizedApp, err := app.NewCentralizedApp()
		if err != nil {
			fmt.Printf("Error initializing centralized app: %v\n", err)
			fmt.Println("\nFalling back to legacy storage...")
			legacyApp := app.NewApp()
			program = tea.NewProgram(legacyApp, tea.WithAltScreen())
		} else {
			program = tea.NewProgram(centralizedApp, tea.WithAltScreen())
		}
	}
	
	// Run the program
	if _, err := program.Run(); err != nil {
		log.Fatal("Error running program:", err)
		os.Exit(1)
	}
}