package hooks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"claude-work-tracker-ui/internal/models"
)

// HookType represents different types of hooks in the system
type HookType string

const (
	// Status change hooks
	BeforeStatusChange HookType = "before_status_change"
	AfterStatusChange  HookType = "after_status_change"

	// Schedule change hooks
	BeforeScheduleChange HookType = "before_schedule_change"
	AfterScheduleChange  HookType = "after_schedule_change"

	// Progress hooks
	ProgressUpdated HookType = "progress_updated"

	// Activity hooks
	ActivityDetected HookType = "activity_detected"
	InactivityWarning HookType = "inactivity_warning"

	// Git hooks
	GitContextChanged HookType = "git_context_changed"
	CommitDetected    HookType = "commit_detected"
)

// HookContext contains information passed to hook handlers
type HookContext struct {
	WorkItem    *models.Work
	OldWorkItem *models.Work // For change events
	EventType   HookType
	Timestamp   time.Time
	Metadata    map[string]interface{}
}

// HookHandler is a function that processes hook events
type HookHandler func(ctx context.Context, hookCtx *HookContext) error

// HookResult contains the result of a hook execution
type HookResult struct {
	HookName  string
	Success   bool
	Error     error
	Duration  time.Duration
	Metadata  map[string]interface{}
}

// HookSystem manages hook registration and execution
type HookSystem struct {
	mu       sync.RWMutex
	handlers map[HookType][]namedHandler
	config   *HookConfig
}

type namedHandler struct {
	name    string
	handler HookHandler
}

// HookConfig contains configuration for the hook system
type HookConfig struct {
	Enabled            bool
	Timeout            time.Duration
	ContinueOnError    bool
	MaxConcurrentHooks int
}

// DefaultHookConfig returns a default configuration
func DefaultHookConfig() *HookConfig {
	return &HookConfig{
		Enabled:            true,
		Timeout:            5 * time.Second,
		ContinueOnError:    true,
		MaxConcurrentHooks: 10,
	}
}

// NewHookSystem creates a new hook system
func NewHookSystem(config *HookConfig) *HookSystem {
	if config == nil {
		config = DefaultHookConfig()
	}
	return &HookSystem{
		handlers: make(map[HookType][]namedHandler),
		config:   config,
	}
}

// Register adds a hook handler for a specific hook type
func (hs *HookSystem) Register(hookType HookType, name string, handler HookHandler) {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	hs.handlers[hookType] = append(hs.handlers[hookType], namedHandler{
		name:    name,
		handler: handler,
	})
}

// Execute runs all registered handlers for a hook type
func (hs *HookSystem) Execute(ctx context.Context, hookCtx *HookContext) ([]HookResult, error) {
	if !hs.config.Enabled {
		return nil, nil
	}

	hs.mu.RLock()
	handlers := hs.handlers[hookCtx.EventType]
	hs.mu.RUnlock()

	if len(handlers) == 0 {
		return nil, nil
	}

	results := make([]HookResult, 0, len(handlers))
	var wg sync.WaitGroup
	resultCh := make(chan HookResult, len(handlers))
	sem := make(chan struct{}, hs.config.MaxConcurrentHooks)

	for _, h := range handlers {
		wg.Add(1)
		go func(handler namedHandler) {
			defer wg.Done()
			sem <- struct{}{} // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			start := time.Now()
			ctxWithTimeout, cancel := context.WithTimeout(ctx, hs.config.Timeout)
			defer cancel()

			err := handler.handler(ctxWithTimeout, hookCtx)
			duration := time.Since(start)

			result := HookResult{
				HookName: handler.name,
				Success:  err == nil,
				Error:    err,
				Duration: duration,
			}

			resultCh <- result
		}(h)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var firstError error
	for result := range resultCh {
		results = append(results, result)
		if result.Error != nil && firstError == nil {
			firstError = result.Error
		}
	}

	if firstError != nil && !hs.config.ContinueOnError {
		return results, fmt.Errorf("hook execution failed: %w", firstError)
	}

	return results, nil
}

// ExecuteSync executes hooks synchronously in order
func (hs *HookSystem) ExecuteSync(ctx context.Context, hookCtx *HookContext) ([]HookResult, error) {
	if !hs.config.Enabled {
		return nil, nil
	}

	hs.mu.RLock()
	handlers := hs.handlers[hookCtx.EventType]
	hs.mu.RUnlock()

	if len(handlers) == 0 {
		return nil, nil
	}

	results := make([]HookResult, 0, len(handlers))

	for _, h := range handlers {
		start := time.Now()
		ctxWithTimeout, cancel := context.WithTimeout(ctx, hs.config.Timeout)
		err := h.handler(ctxWithTimeout, hookCtx)
		cancel()
		duration := time.Since(start)

		result := HookResult{
			HookName: h.name,
			Success:  err == nil,
			Error:    err,
			Duration: duration,
		}

		results = append(results, result)

		if err != nil && !hs.config.ContinueOnError {
			return results, fmt.Errorf("hook %s failed: %w", h.name, err)
		}
	}

	return results, nil
}

// Clear removes all registered handlers
func (hs *HookSystem) Clear() {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	hs.handlers = make(map[HookType][]namedHandler)
}

// GetHandlerCount returns the number of handlers for a hook type
func (hs *HookSystem) GetHandlerCount(hookType HookType) int {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	return len(hs.handlers[hookType])
}