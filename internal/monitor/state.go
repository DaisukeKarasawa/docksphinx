package monitor

import (
	"sync"
	"time"

	"docksphinx/internal/docker"
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

	// Runtime details
	StartedAt        time.Time
	UptimeSeconds    int64
	RestartCount     int
	ComposeProject   string
	ComposeService   string
	VolumeMountCount int
	NetworkNames     []string

	// For threshold detection
	CPUThresholdCount    int // Consecutive CPU threshold violations
	MemoryThresholdCount int // Consecutive memory threshold violations

	// Previous state for comparison
	PreviousState string
	PreviousCPU   float64
	PreviousMem   float64
}

// ComposeGroup is a heuristic grouping based on compose labels and shared network.
type ComposeGroup struct {
	Project        string
	Service        string
	ContainerIDs   []string
	ContainerNames []string
	NetworkNames   []string
}

// ResourceState stores latest non-container resources.
type ResourceState struct {
	Images   []docker.Image
	Networks []docker.Network
	Volumes  []docker.Volume
	Groups   []ComposeGroup
}

// StateManager manages container states
type StateManager struct {
	mu        sync.RWMutex
	states    map[string]*ContainerState // Key: ContainerID
	resources ResourceState
}

// NewStateManager creates a new state manager
func NewStateManager() *StateManager {
	return &StateManager{
		states: make(map[string]*ContainerState),
		resources: ResourceState{
			Images:   nil,
			Networks: nil,
			Volumes:  nil,
			Groups:   nil,
		},
	}
}

// GetState retrieves the current state of a container
func (sm *StateManager) GetState(containerID string) (*ContainerState, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	state, exists := sm.states[containerID]
	if !exists || state == nil {
		return nil, exists
	}
	return cloneContainerState(state), true
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
		result[id] = cloneContainerState(state)
	}

	return result
}

// Clear removes all states (useful for testing or reset)
func (sm *StateManager) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.states = make(map[string]*ContainerState)
	sm.resources = ResourceState{}
}

// UpdateResources updates non-container resource snapshots.
func (sm *StateManager) UpdateResources(resources ResourceState) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.resources = ResourceState{
		Images:   append([]docker.Image(nil), resources.Images...),
		Networks: append([]docker.Network(nil), resources.Networks...),
		Volumes:  append([]docker.Volume(nil), resources.Volumes...),
		Groups:   cloneComposeGroups(resources.Groups),
	}
}

// GetResources returns latest non-container resource snapshot.
func (sm *StateManager) GetResources() ResourceState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return ResourceState{
		Images:   append([]docker.Image(nil), sm.resources.Images...),
		Networks: append([]docker.Network(nil), sm.resources.Networks...),
		Volumes:  append([]docker.Volume(nil), sm.resources.Volumes...),
		Groups:   cloneComposeGroups(sm.resources.Groups),
	}
}

func cloneComposeGroups(in []ComposeGroup) []ComposeGroup {
	if len(in) == 0 {
		return nil
	}
	out := make([]ComposeGroup, 0, len(in))
	for _, g := range in {
		out = append(out, ComposeGroup{
			Project:        g.Project,
			Service:        g.Service,
			ContainerIDs:   append([]string(nil), g.ContainerIDs...),
			ContainerNames: append([]string(nil), g.ContainerNames...),
			NetworkNames:   append([]string(nil), g.NetworkNames...),
		})
	}
	return out
}

func cloneContainerState(in *ContainerState) *ContainerState {
	if in == nil {
		return nil
	}
	out := *in
	out.NetworkNames = append([]string(nil), in.NetworkNames...)
	return &out
}
