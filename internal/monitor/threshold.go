package monitor

import (
	"fmt"
	"time"

	"docksphinx/internal/event"
)

// ThresholdConfig represents threshold configuration
type ThresholdConfig struct {
	CPU    CPUThresholdConfig
	Memory MemoryThresholdConfig
	// CooldownSeconds suppresses repeated same-level threshold events for the same container.
	CooldownSeconds int `yaml:"cooldown_seconds"`
}

// CPUThresholdConfig represents CPU threshold configuration
type CPUThresholdConfig struct {
	Warning          float64 // Warning threshold (%)
	Critical         float64 // Critical threshold (%)
	ConsecutiveCount int     // Number of consecutive violations before generating event
}

// MemoryThresholdConfig represents memory threshold configuration
type MemoryThresholdConfig struct {
	Warning          float64 // Warning threshold (%)
	Critical         float64 // Critical threshold (%)
	ConsecutiveCount int     // Number of consecutive violations before generating event
}

// DefaultThresholdConfig returns default threshold configuration
func DefaultThresholdConfig() ThresholdConfig {
	return ThresholdConfig{
		CPU: CPUThresholdConfig{
			Warning:          70.0,
			Critical:         90.0,
			ConsecutiveCount: 3,
		},
		Memory: MemoryThresholdConfig{
			Warning:          80.0,
			Critical:         95.0,
			ConsecutiveCount: 3,
		},
		CooldownSeconds: 30,
	}
}

// ThresholdMonitor monitors resource usage and detects threshold violations
type ThresholdMonitor struct {
	config   ThresholdConfig
	lastEmit map[string]time.Time
}

// NewThresholdMonitor creates a new threshold monitor
func NewThresholdMonitor(config ThresholdConfig) *ThresholdMonitor {
	return &ThresholdMonitor{
		config:   config,
		lastEmit: make(map[string]time.Time),
	}
}

// CheckThresholds checks if resource usage exceeds thresholds
// Returns events if thresholds are exceeded for consecutive times
func (tm *ThresholdMonitor) CheckThresholds(
	containerID, containerName, imageName string,
	cpuPercent, memoryPercent float64,
	state *ContainerState,
) []*event.Event {
	var events []*event.Event

	// Check CPU threshold
	if cpuPercent >= tm.config.CPU.Critical {
		state.CPUThresholdCount++
		if state.CPUThresholdCount >= tm.config.CPU.ConsecutiveCount {
			if tm.shouldEmit(containerID, event.EventTypeCPUThreshold, "critical") {
				evt := event.NewEvent(event.EventTypeCPUThreshold, containerID, containerName, imageName)
				evt.Message = fmt.Sprintf("Container %s CPU usage critical: %.2f%% (threshold: %.2f%%)",
					containerName, cpuPercent, tm.config.CPU.Critical)
				evt.Data["cpu_percent"] = cpuPercent
				evt.Data["threshold"] = tm.config.CPU.Critical
				evt.Data["level"] = "critical"
				evt.Data["consecutive_count"] = state.CPUThresholdCount
				events = append(events, evt)
			}
			state.CPUThresholdCount = 0
		}
	} else if cpuPercent >= tm.config.CPU.Warning {
		state.CPUThresholdCount++
		if state.CPUThresholdCount >= tm.config.CPU.ConsecutiveCount {
			if tm.shouldEmit(containerID, event.EventTypeCPUThreshold, "warning") {
				evt := event.NewEvent(event.EventTypeCPUThreshold, containerID, containerName, imageName)
				evt.Message = fmt.Sprintf("Container %s CPU usage warning: %.2f%% (threshold: %.2f%%)",
					containerName, cpuPercent, tm.config.CPU.Warning)
				evt.Data["cpu_percent"] = cpuPercent
				evt.Data["threshold"] = tm.config.CPU.Warning
				evt.Data["level"] = "warning"
				evt.Data["consecutive_count"] = state.CPUThresholdCount
				events = append(events, evt)
			}
			state.CPUThresholdCount = 0
		}
	} else {
		state.CPUThresholdCount = 0
	}

	// Check memory threshold
	if memoryPercent >= tm.config.Memory.Critical {
		state.MemoryThresholdCount++
		if state.MemoryThresholdCount >= tm.config.Memory.ConsecutiveCount {
			if tm.shouldEmit(containerID, event.EventTypeMemThreshold, "critical") {
				evt := event.NewEvent(event.EventTypeMemThreshold, containerID, containerName, imageName)
				evt.Message = fmt.Sprintf("Container %s memory usage critical: %.2f%% (threshold: %.2f%%)",
					containerName, memoryPercent, tm.config.Memory.Critical)
				evt.Data["memory_percent"] = memoryPercent
				evt.Data["threshold"] = tm.config.Memory.Critical
				evt.Data["level"] = "critical"
				evt.Data["consecutive_count"] = state.MemoryThresholdCount
				events = append(events, evt)
			}
			state.MemoryThresholdCount = 0
		}
	} else if memoryPercent >= tm.config.Memory.Warning {
		state.MemoryThresholdCount++
		if state.MemoryThresholdCount >= tm.config.Memory.ConsecutiveCount {
			if tm.shouldEmit(containerID, event.EventTypeMemThreshold, "warning") {
				evt := event.NewEvent(event.EventTypeMemThreshold, containerID, containerName, imageName)
				evt.Message = fmt.Sprintf("Container %s memory usage warning: %.2f%% (threshold: %.2f%%)",
					containerName, memoryPercent, tm.config.Memory.Warning)
				evt.Data["memory_percent"] = memoryPercent
				evt.Data["threshold"] = tm.config.Memory.Warning
				evt.Data["level"] = "warning"
				evt.Data["consecutive_count"] = state.MemoryThresholdCount
				events = append(events, evt)
			}
			state.MemoryThresholdCount = 0
		}
	} else {
		state.MemoryThresholdCount = 0
	}

	return events
}

func (tm *ThresholdMonitor) shouldEmit(containerID string, typ event.EventType, level string) bool {
	if tm.config.CooldownSeconds <= 0 {
		return true
	}
	key := fmt.Sprintf("%s|%s|%s", containerID, typ, level)
	now := time.Now()
	last, exists := tm.lastEmit[key]
	if exists && now.Sub(last) < time.Duration(tm.config.CooldownSeconds)*time.Second {
		return false
	}
	tm.lastEmit[key] = now
	return true
}
