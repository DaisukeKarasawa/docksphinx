package monitor

import (
	"testing"
	"time"
)

func TestThresholdMonitorCooldownSuppressesRepeatedEvents(t *testing.T) {
	cfg := ThresholdConfig{
		CPU: CPUThresholdConfig{
			Warning:          70,
			Critical:         90,
			ConsecutiveCount: 2,
		},
		Memory: MemoryThresholdConfig{
			Warning:          80,
			Critical:         95,
			ConsecutiveCount: 2,
		},
		CooldownSeconds: 30,
	}

	monitor := NewThresholdMonitor(cfg)
	state := &ContainerState{}

	events := monitor.CheckThresholds("c1", "web", "img", 95, 10, state)
	if len(events) != 0 {
		t.Fatalf("expected no event at first violation, got %d", len(events))
	}
	events = monitor.CheckThresholds("c1", "web", "img", 95, 10, state)
	if len(events) != 1 {
		t.Fatalf("expected event after second consecutive violation, got %d", len(events))
	}

	_ = monitor.CheckThresholds("c1", "web", "img", 95, 10, state)
	events = monitor.CheckThresholds("c1", "web", "img", 95, 10, state)
	if len(events) != 0 {
		t.Fatalf("expected cooldown suppression, got %d events", len(events))
	}
}

func TestThresholdMonitorCooldownExpiryAllowsEvent(t *testing.T) {
	cfg := ThresholdConfig{
		CPU: CPUThresholdConfig{
			Warning:          70,
			Critical:         90,
			ConsecutiveCount: 2,
		},
		Memory: MemoryThresholdConfig{
			Warning:          80,
			Critical:         95,
			ConsecutiveCount: 2,
		},
		CooldownSeconds: 1,
	}

	monitor := NewThresholdMonitor(cfg)
	state := &ContainerState{}

	_ = monitor.CheckThresholds("c1", "web", "img", 95, 10, state)
	events := monitor.CheckThresholds("c1", "web", "img", 95, 10, state)
	if len(events) != 1 {
		t.Fatalf("expected first event, got %d", len(events))
	}

	time.Sleep(1100 * time.Millisecond)
	_ = monitor.CheckThresholds("c1", "web", "img", 95, 10, state)
	events = monitor.CheckThresholds("c1", "web", "img", 95, 10, state)
	if len(events) != 1 {
		t.Fatalf("expected event after cooldown expiration, got %d", len(events))
	}
}
