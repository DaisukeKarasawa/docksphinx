package monitor

import (
	"fmt"
	"time"

	"docksphinx/internal/event"
)

// Detector detects container state changes and generates events
type Detector struct {
	stateManager *StateManager
}

// NewDetector creates a new event detector
func NewDetector(stateManager *StateManager) *Detector {
	return &Detector{
		stateManager: stateManager,
	}
}

// DetectStateChange detects state changes and returns events
// This is called after updating container states
func (d *Detector) DetectStateChange(containerID, containerName, imageName, currentState string) []*event.Event {
	if d == nil || d.stateManager == nil {
		return nil
	}
	oldState, exists := d.stateManager.GetState(containerID)

	var events []*event.Event

	if !exists {
		// New container detected
		if currentState == "running" {
			evt := event.NewEvent(event.EventTypeStarted, containerID, containerName, imageName)
			evt.Message = fmt.Sprintf("Container %s started", containerName)
			events = append(events, evt)
		}
		return events
	}

	// State changed
	if oldState.State != currentState {
		switch currentState {
		case "running":
			// Check if this is a restart
			if oldState.State == "exited" || oldState.State == "dead" {
				// Check if the container was recently stopped (within 10 seconds)
				timeSinceLastSeen := time.Since(oldState.LastSeen)
				if timeSinceLastSeen < 10*time.Second {
					// Likely a restart
					evt := event.NewEvent(event.EventTypeRestarted, containerID, containerName, imageName)
					evt.Message = fmt.Sprintf("Container %s restarted", containerName)
					evt.Data["previous_state"] = oldState.State
					evt.Data["time_since_stop"] = timeSinceLastSeen.Seconds()
					events = append(events, evt)
				} else {
					// New start after a long time
					evt := event.NewEvent(event.EventTypeStarted, containerID, containerName, imageName)
					evt.Message = fmt.Sprintf("Container %s started", containerName)
					evt.Data["previous_state"] = oldState.State
					events = append(events, evt)
				}
			} else {
				// Transition from other state to running
				evt := event.NewEvent(event.EventTypeStarted, containerID, containerName, imageName)
				evt.Message = fmt.Sprintf("Container %s started", containerName)
				evt.Data["previous_state"] = oldState.State
				events = append(events, evt)
			}

		case "exited":
			// Container stopped normally
			evt := event.NewEvent(event.EventTypeStopped, containerID, containerName, imageName)
			evt.Message = fmt.Sprintf("Container %s stopped", containerName)
			evt.Data["previous_state"] = oldState.State
			events = append(events, evt)

		case "dead":
			// Container died (abnormal exit)
			evt := event.NewEvent(event.EventTypeDied, containerID, containerName, imageName)
			evt.Message = fmt.Sprintf("Container %s died (abnormal exit)", containerName)
			evt.Data["previous_state"] = oldState.State
			events = append(events, evt)
		}
	}

	return events
}
