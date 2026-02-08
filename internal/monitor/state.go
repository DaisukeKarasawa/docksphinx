package monitor

import (
	"sync"
	"time"
)

// ContainerState represents the current state of a container
type ContainerState struct {
	// Container identification
	ContainerID   string
	ContainerName string
	ImageName     string

	// State information
	State    string    // "running", "exited", "restarting", etc.
	Status   string    // Human-readable status string
	LastSeen time.Time // When this state was last observed

	// Metrics
	CPUPercent    float64
	MemoryUsage   int64
	MemoryLimit   int64
	MemoryPercent float64
	NetworkRx     int64
	NetworkTx     int64

	// For threshold detection
	CPUThresholdCount    int // Consecutive CPU threshold violations
	MemoryThresholdCount int // Consecutive memory threshold violations

	// Previous state for comparison
	PreviousState string
	PreviousCPU   float64
	PreviousMem   float64
}

// StateManager manages container states
type StateManager struct {
	mu     sync.RWMutex
	states map[string]*ContainerState // Key: ContainerID
}

// NewStateManager creates a new state manager
func NewStateManager() *StateManager {
	return &StateManager{
		states: make(map[string]*ContainerState),
	}
}

// GetState retrieves the current state of a container
func (sm *StateManager) GetState(containerID string) (*ContainerState, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	state, exists := sm.states[containerID]
	return state, exists
}

// UpdateState updates the state of a container
// Returns true if the state changed (for event detection)
func (sm *StateManager) UpdateState(containerID string, state *ContainerState) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	oldState, exists := sm.states[containerID]

	// Store previous values for comparison
	if exists {
		state.PreviousState = oldState.State
		state.PreviousCPU = oldState.CPUPercent
		state.PreviousMem = oldState.MemoryPercent
	}

	sm.states[containerID] = state

	// Check if state changed
	if !exists {
		return true // New container
	}

	return oldState.State != state.State
}

// RemoveState removes a container from state tracking
// Called when a container is removed
func (sm *StateManager) RemoveState(containerID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.states, containerID)
}

// GetAllStates returns all current container states
func (sm *StateManager) GetAllStates() map[string]*ContainerState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Create a copy to avoid race conditions
	result := make(map[string]*ContainerState)
	for id, state := range sm.states {
		result[id] = state
	}

	return result
}

// Clear removes all states (useful for testing or reset)
func (sm *StateManager) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.states = make(map[string]*ContainerState)
}
