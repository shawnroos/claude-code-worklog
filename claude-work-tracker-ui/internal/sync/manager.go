package sync

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"claude-work-tracker-ui/internal/models"
)

// SyncEvent represents a file system event for work items
type SyncEvent struct {
	Type     string                `json:"type"`     // "created", "modified", "deleted"
	ItemID   string                `json:"item_id"`
	FilePath string                `json:"file_path"`
	Content  *models.MarkdownWorkItem `json:"content,omitempty"`
	Timestamp time.Time             `json:"timestamp"`
}

// SyncListener is a function that handles sync events
type SyncListener func(event SyncEvent)

// SyncManager handles real-time synchronization of work items
type SyncManager struct {
	watcher      *fsnotify.Watcher
	eventBus     chan SyncEvent
	listeners    []SyncListener
	watchDir     string
	isRunning    bool
	stopChan     chan bool
	mu           sync.RWMutex
	
	// Debouncing to handle rapid file changes
	debounceMap  map[string]*time.Timer
	debounceMu   sync.Mutex
	debounceTime time.Duration
}

// NewSyncManager creates a new sync manager for the given directory
func NewSyncManager(watchDir string) (*SyncManager, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	sm := &SyncManager{
		watcher:      watcher,
		eventBus:     make(chan SyncEvent, 100), // Buffer for events
		listeners:    make([]SyncListener, 0),
		watchDir:     watchDir,
		stopChan:     make(chan bool),
		debounceMap:  make(map[string]*time.Timer),
		debounceTime: 100 * time.Millisecond, // 100ms debounce
	}

	return sm, nil
}

// AddListener adds a listener for sync events
func (sm *SyncManager) AddListener(listener SyncListener) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.listeners = append(sm.listeners, listener)
}

// Start begins watching for file changes
func (sm *SyncManager) Start() error {
	sm.mu.Lock()
	if sm.isRunning {
		sm.mu.Unlock()
		return fmt.Errorf("sync manager is already running")
	}
	sm.isRunning = true
	sm.mu.Unlock()

	// Add the watch directory and subdirectories
	if err := sm.addWatchPaths(); err != nil {
		return fmt.Errorf("failed to add watch paths: %w", err)
	}

	// Start the event processing goroutines
	go sm.processFileEvents()
	go sm.processSyncEvents()

	log.Printf("Sync manager started, watching: %s", sm.watchDir)
	return nil
}

// Stop stops the sync manager
func (sm *SyncManager) Stop() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.isRunning {
		return
	}

	sm.isRunning = false
	close(sm.stopChan)
	close(sm.eventBus)
	sm.watcher.Close()

	// Clear debounce timers
	sm.debounceMu.Lock()
	for _, timer := range sm.debounceMap {
		timer.Stop()
	}
	sm.debounceMap = make(map[string]*time.Timer)
	sm.debounceMu.Unlock()

	log.Println("Sync manager stopped")
}

// addWatchPaths adds all relevant directories to the watcher
func (sm *SyncManager) addWatchPaths() error {
	// Watch main .claude-work directory
	if err := sm.watcher.Add(sm.watchDir); err != nil {
		return err
	}

	// Watch subdirectories
	subdirs := []string{
		"items",
		"items/now",
		"items/next", 
		"items/later",
		"decisions",
		"decisions/active",
	}

	for _, subdir := range subdirs {
		path := filepath.Join(sm.watchDir, subdir)
		if err := sm.watcher.Add(path); err != nil {
			// Directory might not exist yet, which is okay
			log.Printf("Warning: Could not watch %s: %v", path, err)
		}
	}

	return nil
}

// processFileEvents handles raw file system events
func (sm *SyncManager) processFileEvents() {
	for {
		select {
		case event, ok := <-sm.watcher.Events:
			if !ok {
				return
			}
			sm.handleFileEvent(event)

		case err, ok := <-sm.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("File watcher error: %v", err)

		case <-sm.stopChan:
			return
		}
	}
}

// handleFileEvent processes a single file system event
func (sm *SyncManager) handleFileEvent(event fsnotify.Event) {
	// Only process markdown files
	if !strings.HasSuffix(event.Name, ".md") {
		return
	}

	// Debounce rapid changes to the same file
	sm.debounceMu.Lock()
	if timer, exists := sm.debounceMap[event.Name]; exists {
		timer.Stop()
	}

	sm.debounceMap[event.Name] = time.AfterFunc(sm.debounceTime, func() {
		sm.processEventAfterDebounce(event)
		
		sm.debounceMu.Lock()
		delete(sm.debounceMap, event.Name)
		sm.debounceMu.Unlock()
	})
	sm.debounceMu.Unlock()
}

// processEventAfterDebounce handles the event after debouncing
func (sm *SyncManager) processEventAfterDebounce(event fsnotify.Event) {
	var eventType string
	var content *models.MarkdownWorkItem

	// Determine event type
	if event.Op&fsnotify.Create == fsnotify.Create {
		eventType = "created"
	} else if event.Op&fsnotify.Write == fsnotify.Write {
		eventType = "modified"
	} else if event.Op&fsnotify.Remove == fsnotify.Remove {
		eventType = "deleted"
	} else {
		return // Ignore other event types
	}

	// Extract item ID from filename
	itemID := sm.extractItemIDFromPath(event.Name)
	if itemID == "" {
		return
	}

	// For create/modify events, try to load the content
	if eventType != "deleted" {
		// We would need access to MarkdownIO here to load the content
		// For now, we'll create the event without content and let listeners load it
	}

	syncEvent := SyncEvent{
		Type:      eventType,
		ItemID:    itemID,
		FilePath:  event.Name,
		Content:   content,
		Timestamp: time.Now(),
	}

	// Send to event bus
	select {
	case sm.eventBus <- syncEvent:
	default:
		log.Printf("Event bus full, dropping event for %s", event.Name)
	}
}

// processSyncEvents distributes sync events to listeners
func (sm *SyncManager) processSyncEvents() {
	for event := range sm.eventBus {
		sm.mu.RLock()
		listeners := make([]SyncListener, len(sm.listeners))
		copy(listeners, sm.listeners)
		sm.mu.RUnlock()

		// Notify all listeners
		for _, listener := range listeners {
			go func(l SyncListener, e SyncEvent) {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Listener panic: %v", r)
					}
				}()
				l(e)
			}(listener, event)
		}
	}
}

// extractItemIDFromPath extracts the work item ID from a file path
func (sm *SyncManager) extractItemIDFromPath(path string) string {
	filename := filepath.Base(path)
	
	// Remove .md extension
	if strings.HasSuffix(filename, ".md") {
		filename = filename[:len(filename)-3]
	}

	// Extract ID from filename pattern: {type}-{description}-{date}-{id}.md
	parts := strings.Split(filename, "-")
	if len(parts) >= 4 {
		// The ID is the last part
		return parts[len(parts)-1]
	}

	return ""
}

// IsRunning returns whether the sync manager is currently running
func (sm *SyncManager) IsRunning() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.isRunning
}