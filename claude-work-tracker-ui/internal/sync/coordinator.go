package sync

import (
	"fmt"
	"log"
	"path/filepath"
	"sync"

	"claude-work-tracker-ui/internal/data"
	"claude-work-tracker-ui/internal/models"
)

// UIUpdateCallback is called when the UI should refresh
type UIUpdateCallback func(eventType string, item *models.MarkdownWorkItem)

// SyncCoordinator coordinates between file watching and UI updates
type SyncCoordinator struct {
	syncManager  *SyncManager
	terminalSync *TerminalSync
	dataClient   *data.EnhancedClient
	uiCallback   UIUpdateCallback
	watchDir     string
	mu           sync.RWMutex
}

// NewSyncCoordinator creates a new sync coordinator
func NewSyncCoordinator(watchDir string, dataClient *data.EnhancedClient) (*SyncCoordinator, error) {
	syncManager, err := NewSyncManager(watchDir)
	if err != nil {
		return nil, err
	}

	// Initialize terminal sync for multi-instance coordination
	terminalSync, err := NewTerminalSync(watchDir)
	if err != nil {
		log.Printf("Failed to initialize terminal sync: %v", err)
		// Continue without terminal sync - it's an optional feature
	}

	coordinator := &SyncCoordinator{
		syncManager:  syncManager,
		terminalSync: terminalSync,
		dataClient:   dataClient,
		watchDir:     watchDir,
	}

	// Add coordinator as a listener to sync events
	syncManager.AddListener(coordinator.handleSyncEvent)
	
	// Add coordinator as a subscriber to terminal sync messages
	if terminalSync != nil {
		terminalSync.Subscribe(coordinator.handleTerminalSyncMessage)
	}

	return coordinator, nil
}

// SetUICallback sets the callback function for UI updates
func (sc *SyncCoordinator) SetUICallback(callback UIUpdateCallback) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.uiCallback = callback
}

// Start starts the sync coordinator
func (sc *SyncCoordinator) Start() error {
	// Start file sync manager
	if err := sc.syncManager.Start(); err != nil {
		return err
	}
	
	// Start terminal sync if available
	if sc.terminalSync != nil {
		if err := sc.terminalSync.Start(); err != nil {
			log.Printf("Failed to start terminal sync: %v", err)
			// Continue without terminal sync
		}
	}
	
	return nil
}

// Stop stops the sync coordinator
func (sc *SyncCoordinator) Stop() {
	sc.syncManager.Stop()
	
	if sc.terminalSync != nil {
		sc.terminalSync.Stop()
	}
}

// IsRunning returns whether the coordinator is running
func (sc *SyncCoordinator) IsRunning() bool {
	return sc.syncManager.IsRunning()
}

// handleSyncEvent processes sync events and triggers UI updates
func (sc *SyncCoordinator) handleSyncEvent(event SyncEvent) {
	log.Printf("Sync event: %s for item %s (%s)", event.Type, event.ItemID, filepath.Base(event.FilePath))

	// Broadcast to other terminal instances
	if sc.terminalSync != nil {
		go func() {
			msgType := fmt.Sprintf("work_item_%s", event.Type)
			if err := sc.terminalSync.BroadcastMessage(msgType, event.ItemID, event.FilePath); err != nil {
				log.Printf("Failed to broadcast terminal sync message: %v", err)
			}
		}()
	}

	// Get UI callback
	sc.mu.RLock()
	callback := sc.uiCallback
	sc.mu.RUnlock()

	if callback == nil {
		return // No UI to update
	}

	var item *models.MarkdownWorkItem

	// Load the item content for created/modified events
	if event.Type == "created" || event.Type == "modified" {
		// Try to load the item from the file
		if loadedItem, err := sc.loadItemFromFile(event.FilePath); err == nil {
			item = loadedItem
		} else {
			log.Printf("Failed to load item from %s: %v", event.FilePath, err)
			return
		}
	}

	// Call UI callback
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("UI callback panic: %v", r)
			}
		}()
		callback(event.Type, item)
	}()
}

// handleTerminalSyncMessage processes messages from other terminal instances
func (sc *SyncCoordinator) handleTerminalSyncMessage(message TerminalSyncMessage) {
	log.Printf("Terminal sync message: %s for item %s from %s", message.Type, message.ItemID, message.InstanceID)

	// Get UI callback
	sc.mu.RLock()
	callback := sc.uiCallback
	sc.mu.RUnlock()

	if callback == nil {
		return // No UI to update
	}

	// Convert terminal sync message type to UI event type
	var eventType string
	switch message.Type {
	case "work_item_created":
		eventType = "created"
	case "work_item_modified":
		eventType = "modified"  
	case "work_item_deleted":
		eventType = "deleted"
	default:
		eventType = "refresh" // Generic refresh for unknown types
	}

	var item *models.MarkdownWorkItem

	// Load the item content for created/modified events
	if eventType == "created" || eventType == "modified" {
		if message.FilePath != "" {
			if loadedItem, err := sc.loadItemFromFile(message.FilePath); err == nil {
				item = loadedItem
			} else {
				log.Printf("Failed to load item from terminal sync message: %v", err)
			}
		}
	}

	// Call UI callback
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Terminal sync UI callback panic: %v", r)
			}
		}()
		callback(eventType, item)
	}()
}

// loadItemFromFile loads a work item from a markdown file
func (sc *SyncCoordinator) loadItemFromFile(filePath string) (*models.MarkdownWorkItem, error) {
	// Use the enhanced client's MarkdownIO to load the item
	markdownIO := sc.dataClient.GetMarkdownIO()
	return markdownIO.ReadMarkdownWorkItem(filePath)
}

// TriggerManualRefresh can be called to manually trigger a UI refresh
func (sc *SyncCoordinator) TriggerManualRefresh() {
	sc.mu.RLock()
	callback := sc.uiCallback
	sc.mu.RUnlock()

	if callback != nil {
		go callback("refresh", nil)
	}
}

// GetStats returns statistics about the sync manager
func (sc *SyncCoordinator) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"running":    sc.syncManager.IsRunning(),
		"watch_dir":  sc.watchDir,
		"listeners":  len(sc.syncManager.listeners),
	}
	
	// Add terminal sync stats
	if sc.terminalSync != nil {
		stats["terminal_sync_running"] = sc.terminalSync.IsRunning()
		
		if activeInstances, err := sc.terminalSync.GetActiveInstances(); err == nil {
			stats["active_instances"] = len(activeInstances)
			stats["other_instances"] = activeInstances
		}
	} else {
		stats["terminal_sync_running"] = false
		stats["active_instances"] = 0
	}
	
	return stats
}