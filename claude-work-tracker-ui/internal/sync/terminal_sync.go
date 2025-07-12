package sync

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// TerminalSyncMessage represents a message between terminal instances
type TerminalSyncMessage struct {
	Type        string    `json:"type"`         // "work_item_created", "work_item_updated", "work_item_deleted"
	InstanceID  string    `json:"instance_id"`  // Unique ID for this terminal instance
	ItemID      string    `json:"item_id"`
	FilePath    string    `json:"file_path"`
	Timestamp   time.Time `json:"timestamp"`
	SessionInfo string    `json:"session_info"` // Optional session context
}

// TerminalSync handles synchronization between multiple terminal instances
type TerminalSync struct {
	instanceID    string
	syncDir       string
	messageQueue  chan TerminalSyncMessage
	subscribers   []func(TerminalSyncMessage)
	isRunning     bool
	stopChan      chan bool
	mu            sync.RWMutex
	lastProcessed time.Time
}

// NewTerminalSync creates a new terminal synchronization manager
func NewTerminalSync(baseDir string) (*TerminalSync, error) {
	instanceID := fmt.Sprintf("instance-%d-%d", os.Getpid(), time.Now().UnixNano())
	syncDir := filepath.Join(baseDir, ".sync")
	
	// Create sync directory if it doesn't exist
	if err := os.MkdirAll(syncDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sync directory: %w", err)
	}
	
	ts := &TerminalSync{
		instanceID:    instanceID,
		syncDir:       syncDir,
		messageQueue:  make(chan TerminalSyncMessage, 50),
		subscribers:   make([]func(TerminalSyncMessage), 0),
		stopChan:      make(chan bool),
		lastProcessed: time.Now(),
	}
	
	return ts, nil
}

// Subscribe adds a callback for terminal sync messages
func (ts *TerminalSync) Subscribe(callback func(TerminalSyncMessage)) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.subscribers = append(ts.subscribers, callback)
}

// Start begins terminal synchronization
func (ts *TerminalSync) Start() error {
	ts.mu.Lock()
	if ts.isRunning {
		ts.mu.Unlock()
		return fmt.Errorf("terminal sync is already running")
	}
	ts.isRunning = true
	ts.mu.Unlock()
	
	// Start message processing
	go ts.processMessages()
	go ts.pollForMessages()
	
	log.Printf("Terminal sync started with instance ID: %s", ts.instanceID)
	return nil
}

// Stop stops terminal synchronization
func (ts *TerminalSync) Stop() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	if !ts.isRunning {
		return
	}
	
	ts.isRunning = false
	close(ts.stopChan)
	close(ts.messageQueue)
	
	// Clean up instance files
	ts.cleanupInstanceFiles()
	
	log.Println("Terminal sync stopped")
}

// BroadcastMessage sends a message to all other terminal instances
func (ts *TerminalSync) BroadcastMessage(msgType, itemID, filePath string) error {
	message := TerminalSyncMessage{
		Type:       msgType,
		InstanceID: ts.instanceID,
		ItemID:     itemID,
		FilePath:   filePath,
		Timestamp:  time.Now(),
	}
	
	// Write message to sync directory
	filename := fmt.Sprintf("msg-%s-%d.json", ts.instanceID, time.Now().UnixNano())
	filepath := filepath.Join(ts.syncDir, filename)
	
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	if err := ioutil.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write message file: %w", err)
	}
	
	log.Printf("Broadcast message: %s for item %s", msgType, itemID)
	return nil
}

// processMessages handles outgoing message queue
func (ts *TerminalSync) processMessages() {
	for message := range ts.messageQueue {
		ts.mu.RLock()
		subscribers := make([]func(TerminalSyncMessage), len(ts.subscribers))
		copy(subscribers, ts.subscribers)
		ts.mu.RUnlock()
		
		// Notify all subscribers
		for _, callback := range subscribers {
			go func(cb func(TerminalSyncMessage), msg TerminalSyncMessage) {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Terminal sync callback panic: %v", r)
					}
				}()
				cb(msg)
			}(callback, message)
		}
	}
}

// pollForMessages checks for new messages from other instances
func (ts *TerminalSync) pollForMessages() {
	ticker := time.NewTicker(500 * time.Millisecond) // Poll every 500ms
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			ts.checkForNewMessages()
		case <-ts.stopChan:
			return
		}
	}
}

// checkForNewMessages scans the sync directory for new messages
func (ts *TerminalSync) checkForNewMessages() {
	files, err := ioutil.ReadDir(ts.syncDir)
	if err != nil {
		log.Printf("Failed to read sync directory: %v", err)
		return
	}
	
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			filepath := filepath.Join(ts.syncDir, file.Name())
			
			// Skip files that are older than our last processed time
			if file.ModTime().Before(ts.lastProcessed) {
				continue
			}
			
			// Skip our own messages
			if ts.isOwnMessage(file.Name()) {
				continue
			}
			
			// Process the message
			if err := ts.processMessageFile(filepath); err != nil {
				log.Printf("Failed to process message file %s: %v", file.Name(), err)
			}
		}
	}
	
	// Update last processed time
	ts.lastProcessed = time.Now()
	
	// Clean up old message files (older than 1 minute)
	ts.cleanupOldMessages()
}

// isOwnMessage checks if a message file was created by this instance
func (ts *TerminalSync) isOwnMessage(filename string) bool {
	return filepath.HasPrefix(filename, fmt.Sprintf("msg-%s-", ts.instanceID))
}

// processMessageFile reads and processes a single message file
func (ts *TerminalSync) processMessageFile(filepath string) error {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read message file: %w", err)
	}
	
	var message TerminalSyncMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}
	
	// Ignore messages from ourselves
	if message.InstanceID == ts.instanceID {
		return nil
	}
	
	// Add to message queue for processing
	select {
	case ts.messageQueue <- message:
		log.Printf("Received terminal sync message: %s from %s", message.Type, message.InstanceID)
	default:
		log.Printf("Message queue full, dropping message from %s", message.InstanceID)
	}
	
	return nil
}

// cleanupOldMessages removes message files older than 1 minute
func (ts *TerminalSync) cleanupOldMessages() {
	files, err := ioutil.ReadDir(ts.syncDir)
	if err != nil {
		return
	}
	
	cutoff := time.Now().Add(-1 * time.Minute)
	
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			if file.ModTime().Before(cutoff) {
				filepath := filepath.Join(ts.syncDir, file.Name())
				os.Remove(filepath)
			}
		}
	}
}

// cleanupInstanceFiles removes files created by this instance
func (ts *TerminalSync) cleanupInstanceFiles() {
	files, err := ioutil.ReadDir(ts.syncDir)
	if err != nil {
		return
	}
	
	prefix := fmt.Sprintf("msg-%s-", ts.instanceID)
	
	for _, file := range files {
		if !file.IsDir() && filepath.HasPrefix(file.Name(), prefix) {
			filepath := filepath.Join(ts.syncDir, file.Name())
			os.Remove(filepath)
		}
	}
}

// GetActiveInstances returns a list of currently active terminal instances
func (ts *TerminalSync) GetActiveInstances() ([]string, error) {
	files, err := ioutil.ReadDir(ts.syncDir)
	if err != nil {
		return nil, err
	}
	
	instances := make(map[string]bool)
	cutoff := time.Now().Add(-2 * time.Minute) // Consider instances active if they've sent messages in the last 2 minutes
	
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" && file.ModTime().After(cutoff) {
			// Extract instance ID from filename
			parts := strings.SplitN(file.Name(), "-", 3)
			if len(parts) >= 3 && parts[0] == "msg" {
				instances[parts[1]] = true
			}
		}
	}
	
	var result []string
	for instanceID := range instances {
		if instanceID != ts.instanceID {
			result = append(result, instanceID)
		}
	}
	
	return result, nil
}

// IsRunning returns whether terminal sync is currently running
func (ts *TerminalSync) IsRunning() bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.isRunning
}