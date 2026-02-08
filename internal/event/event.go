package event

import (
	"fmt"
	"time"
)

// EventType represents the type of event
type EventType string

const (
	// Container lifecycle events
	EventTypeStarted   EventType = "started"   // container started
	EventTypeStopped   EventType = "stopped"   // container stopped
	EventTypeRestarted EventType = "restarted" // Container restarted
	EventTypeDied      EventType = "died"      // Container died (abnormal exit)

	// Resource threshold events
	EventTypeCPUThreshold EventType = "cpu_threshold" // CPU usage exceeded threshold
	EventTypeMemThreshold EventType = "mem_threshold" // Memory usage exceeded threshold
)

// Event represents a monitoring event
type Event struct {
	// Event identification
	ID        string    // Unique event ID
	Type      EventType // Event type
	Timestamp time.Time // When the event occurred

	// Container information
	ContainerID   string // Container ID
	ContainerName string // Container name
	ImageName     string // Image name

	// Event-specific data
	// For threshold events, this contains the threshold value and actual value
	Data map[string]interface{}

	// Message for human-readable description
	Message string
}

// NewEvent creates a new event
func NewEvent(eventType EventType, containerID, containerName, imageName string) *Event {
	return &Event{
		ID:            generateEventID(),
		Type:          eventType,
		Timestamp:     time.Now(),
		ContainerID:   containerID,
		ContainerName: containerName,
		ImageName:     imageName,
		Data:          make(map[string]interface{}),
	}
}

// generateEventID generates a unique event ID
func generateEventID() string {
	// Simple implementation using timestamp and random number
	// In production, you might want to use UUID
	return time.Now().Format("20060102150405") + "-" +
		fmt.Sprintf("%d", time.Now().UnixNano()%1000000)
}
