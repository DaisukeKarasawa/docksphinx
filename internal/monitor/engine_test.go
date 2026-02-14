package monitor

import (
	"context"
	"testing"
	"time"

	"docksphinx/internal/docker"
	"docksphinx/internal/event"
)

func TestStateManager(t *testing.T) {
	sm := NewStateManager()

	state := &ContainerState{
		ContainerID:   "test-container",
		ContainerName: "test",
		ImageName:     "test-image",
		State:         "running",
		Status:        "Up 1 hour",
		LastSeen:      time.Now(),
		NetworkNames:  []string{"net-a"},
	}

	changed := sm.UpdateState("test-container", state)
	if !changed {
		t.Error("Expected state change for new container")
	}

	retrieved, exists := sm.GetState("test-container")
	if !exists {
		t.Error("Expected state to exist")
	}
	if retrieved.ContainerName != "test" {
		t.Errorf("Expected container name 'test', got '%s'", retrieved.ContainerName)
	}
	retrieved.ContainerName = "mutated"
	again, exists := sm.GetState("test-container")
	if !exists {
		t.Error("Expected state to exist")
	}
	if again.ContainerName != "test" {
		t.Fatalf("Expected cloned state read to keep original value, got '%s'", again.ContainerName)
	}
	retrieved.NetworkNames[0] = "net-mutated"
	again2, exists := sm.GetState("test-container")
	if !exists {
		t.Error("Expected state to exist")
	}
	if again2.NetworkNames[0] != "net-a" {
		t.Fatalf("Expected cloned state network names to keep original value, got '%s'", again2.NetworkNames[0])
	}

	all := sm.GetAllStates()
	delete(all, "test-container")
	allAfterDelete := sm.GetAllStates()
	if _, ok := allAfterDelete["test-container"]; !ok {
		t.Fatal("Expected deleting from returned map not to affect internal state")
	}

	state2 := &ContainerState{
		ContainerID:   "test-container",
		ContainerName: "test",
		ImageName:     "test-image",
		State:         "running",
		Status:        "Up 1 hour",
		LastSeen:      time.Now(),
	}

	changed = sm.UpdateState("test-container", state2)
	if changed {
		t.Error("Expected no state change for same state")
	}

	state3 := &ContainerState{
		ContainerID:   "test-container",
		ContainerName: "test",
		ImageName:     "test-image",
		State:         "exited",
		Status:        "Exited (0) 1 minute ago",
		LastSeen:      time.Now(),
	}

	changed = sm.UpdateState("test-container", state3)
	if !changed {
		t.Error("Expected state change")
	}

	sm.RemoveState("test-container")
	_, exists = sm.GetState("test-container")
	if exists {
		t.Error("Expected state to be removed")
	}

	sm.UpdateResources(ResourceState{
		Images: []docker.Image{
			{ID: "img1", Repository: "repo1", Tag: "latest"},
		},
		Groups: []ComposeGroup{
			{
				Project:        "proj",
				Service:        "svc",
				ContainerIDs:   []string{"c1"},
				ContainerNames: []string{"web"},
				NetworkNames:   []string{"net1"},
			},
		},
	})

	resources := sm.GetResources()
	if len(resources.Images) != 1 || resources.Images[0].ID != "img1" {
		t.Fatalf("Expected resource images to be copied")
	}
	// mutate returned copy and ensure manager state is isolated
	resources.Images[0].ID = "mutated"
	resources.Groups[0].ContainerIDs[0] = "mutated"
	resources.Groups[0].ContainerNames[0] = "mutated-name"
	resources.Groups[0].NetworkNames[0] = "mutated-net"

	resources2 := sm.GetResources()
	if resources2.Images[0].ID != "img1" {
		t.Fatalf("Expected internal images to remain unchanged, got %s", resources2.Images[0].ID)
	}
	if resources2.Groups[0].ContainerIDs[0] != "c1" {
		t.Fatalf("Expected internal groups to remain unchanged, got %s", resources2.Groups[0].ContainerIDs[0])
	}
	if resources2.Groups[0].ContainerNames[0] != "web" {
		t.Fatalf("Expected internal group container names to remain unchanged, got %s", resources2.Groups[0].ContainerNames[0])
	}
	if resources2.Groups[0].NetworkNames[0] != "net1" {
		t.Fatalf("Expected internal group network names to remain unchanged, got %s", resources2.Groups[0].NetworkNames[0])
	}
}

func TestStateManagerNilSafetyContracts(t *testing.T) {
	var sm *StateManager

	if state, ok := sm.GetState("x"); ok || state != nil {
		t.Fatalf("expected nil,false for nil receiver GetState, got %#v %v", state, ok)
	}
	if changed := sm.UpdateState("x", &ContainerState{ContainerID: "x"}); changed {
		t.Fatalf("expected nil receiver UpdateState to return false")
	}
	sm.RemoveState("x") // should not panic
	if got := sm.GetAllStates(); got != nil {
		t.Fatalf("expected nil receiver GetAllStates to return nil, got %#v", got)
	}
	sm.Clear() // should not panic
	sm.UpdateResources(ResourceState{
		Images: []docker.Image{{ID: "img"}},
	}) // should not panic
	if got := sm.GetResources(); len(got.Images) != 0 || len(got.Networks) != 0 || len(got.Volumes) != 0 || len(got.Groups) != 0 {
		t.Fatalf("expected zero-value resources for nil receiver, got %#v", got)
	}

	alive := NewStateManager()
	if changed := alive.UpdateState("x", nil); changed {
		t.Fatalf("expected UpdateState with nil state to return false")
	}
	if got := alive.GetAllStates(); len(got) != 0 {
		t.Fatalf("expected UpdateState(nil) not to mutate states, got %#v", got)
	}
}

func TestDetector(t *testing.T) {
	sm := NewStateManager()
	detector := NewDetector(sm)

	events := detector.DetectStateChange("new-container", "new", "new-image", "running")
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	if events[0].Type != "started" {
		t.Errorf("Expected 'started' event, got '%s'", events[0].Type)
	}

	state := &ContainerState{
		ContainerID:   "test-container",
		ContainerName: "test",
		ImageName:     "test-image",
		State:         "running",
		Status:        "Up 1 hour",
		LastSeen:      time.Now(),
	}
	sm.UpdateState("test-container", state)

	events = detector.DetectStateChange("test-container", "test", "test-image", "exited")
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	if events[0].Type != "stopped" {
		t.Errorf("Expected 'stopped' event, got '%s'", events[0].Type)
	}
}

func TestDetectorNilSafetyContracts(t *testing.T) {
	var nilDetector *Detector
	if events := nilDetector.DetectStateChange("cid", "cname", "img", "running"); len(events) != 0 {
		t.Fatalf("expected nil receiver detector to return empty events, got %#v", events)
	}

	detector := NewDetector(nil)
	if events := detector.DetectStateChange("cid", "cname", "img", "running"); len(events) != 0 {
		t.Fatalf("expected detector with nil state manager to return empty events, got %#v", events)
	}
}

func TestThresholdMonitor(t *testing.T) {
	config := DefaultThresholdConfig()
	th := NewThresholdMonitor(config)

	state := &ContainerState{
		ContainerID:          "test-container",
		CPUThresholdCount:    0,
		MemoryThresholdCount: 0,
	}

	events := th.CheckThresholds("test-container", "test", "test-image", 50.0, 50.0, state)
	if len(events) != 0 {
		t.Errorf("Expected no events, got %d", len(events))
	}
	if state.CPUThresholdCount != 0 {
		t.Errorf("Expected CPU count to be 0, got %d", state.CPUThresholdCount)
	}

	events = th.CheckThresholds("test-container", "test", "test-image", 75.0, 50.0, state)
	if len(events) != 0 {
		t.Errorf("Expected no events yet, got %d", len(events))
	}
	if state.CPUThresholdCount != 1 {
		t.Errorf("Expected CPU count to be 1, got %d", state.CPUThresholdCount)
	}

	_ = th.CheckThresholds("test-container", "test", "test-image", 75.0, 50.0, state)
	events = th.CheckThresholds("test-container", "test", "test-image", 75.0, 50.0, state)
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	if events[0].Type != "cpu_threshold" {
		t.Errorf("Expected 'cpu_threshold' event, got '%s'", events[0].Type)
	}
	if state.CPUThresholdCount != 0 {
		t.Errorf("Expected CPU count to be reset to 0, got %d", state.CPUThresholdCount)
	}
}

func TestThresholdMonitorNilSafetyContracts(t *testing.T) {
	var tm *ThresholdMonitor
	state := &ContainerState{}
	if events := tm.CheckThresholds("c1", "name", "image", 95, 95, state); len(events) != 0 {
		t.Fatalf("expected nil monitor to return empty events, got %#v", events)
	}
	if state.CPUThresholdCount != 0 || state.MemoryThresholdCount != 0 {
		t.Fatalf("expected nil monitor not to mutate counters, state=%#v", state)
	}

	monitor := NewThresholdMonitor(DefaultThresholdConfig())
	if events := monitor.CheckThresholds("c1", "name", "image", 95, 95, nil); len(events) != 0 {
		t.Fatalf("expected nil state to return empty events, got %#v", events)
	}

	zeroValue := &ThresholdMonitor{
		config: ThresholdConfig{
			CPU: CPUThresholdConfig{
				Warning:          70,
				Critical:         90,
				ConsecutiveCount: 1,
			},
			Memory: MemoryThresholdConfig{
				Warning:          80,
				Critical:         95,
				ConsecutiveCount: 1,
			},
			CooldownSeconds: 30,
		},
		// lastEmit intentionally nil to verify lazy init path
	}
	s := &ContainerState{}
	events := zeroValue.CheckThresholds("cid", "cname", "img", 95, 96, s)
	if len(events) != 2 {
		t.Fatalf("expected zero-value monitor to emit 2 events without panic, got %d", len(events))
	}
	if zeroValue.lastEmit == nil || len(zeroValue.lastEmit) != 2 {
		t.Fatalf("expected lastEmit map to be initialized and populated, got %#v", zeroValue.lastEmit)
	}
}

func TestEngineIntegration(t *testing.T) {
	dockerClient, err := docker.NewClient()
	if err != nil {
		t.Skip("Docker client not available, skipping integration test")
	}
	defer dockerClient.Close()

	ctx := context.Background()
	if err := dockerClient.Ping(ctx); err != nil {
		t.Skip("Docker daemon not available, skipping integration test")
	}

	config := EngineConfig{
		Interval:   2 * time.Second,
		Thresholds: DefaultThresholdConfig(),
	}

	engine, err := NewEngine(config, dockerClient)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	if err := engine.Start(); err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}
	defer engine.Stop()

	time.Sleep(3 * time.Second)

	select {
	case evt := <-engine.GetEventChannel():
		t.Logf("Received event: %s - %s", evt.Type, evt.Message)
	case <-time.After(5 * time.Second):
		t.Log("No events received (this is OK if no containers are running)")
	}

	states := engine.GetStateManager().GetAllStates()
	t.Logf("Tracking %d containers", len(states))
}

func TestEngineStartFailsWithNilDockerClient(t *testing.T) {
	config := EngineConfig{
		Interval:   1 * time.Second,
		Thresholds: DefaultThresholdConfig(),
	}
	engine, err := NewEngine(config, nil)
	if err != nil {
		t.Fatalf("NewEngine should not fail for nil client setup test: %v", err)
	}

	if err := engine.Start(); err == nil {
		t.Fatal("expected Start to fail when docker client is nil")
	}
}

func TestEngineCannotRestartAfterStop(t *testing.T) {
	config := EngineConfig{
		Interval:   1 * time.Second,
		Thresholds: DefaultThresholdConfig(),
	}
	engine, err := NewEngine(config, nil)
	if err != nil {
		t.Fatalf("NewEngine should not fail for restart guard test: %v", err)
	}

	// Simulate an already-running loop to exercise Stop path safely without Docker.
	engine.mu.Lock()
	engine.running = true
	engine.mu.Unlock()
	engine.wg.Add(1)
	go func() {
		defer engine.wg.Done()
		<-engine.ctx.Done()
	}()

	engine.Stop()
	if err := engine.Start(); err == nil {
		t.Fatal("expected Start to fail after engine has been stopped once")
	}
}

func TestEngineNilSafetyContracts(t *testing.T) {
	var engine *Engine

	if err := engine.Start(); err == nil {
		t.Fatal("expected nil engine start to fail")
	}

	engine.Stop() // should not panic

	if ch := engine.GetEventChannel(); ch != nil {
		t.Fatalf("expected nil engine event channel to be nil, got %#v", ch)
	}
	if sm := engine.GetStateManager(); sm != nil {
		t.Fatalf("expected nil engine state manager to be nil, got %#v", sm)
	}
	engine.SetLogger(nil) // should not panic

	if events := engine.GetRecentEvents(10); events != nil {
		t.Fatalf("expected nil engine recent events to be nil, got %#v", events)
	}
}

func TestEngineLoggerNilSafetyOnInternalPaths(t *testing.T) {
	t.Run("collectAndDetect early-return path", func(t *testing.T) {
		engine := &Engine{
			dockerClient: nil,
			logger:       nil,
		}
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("collectAndDetect should not panic with nil logger: %v", r)
			}
		}()
		engine.collectAndDetect()
	})

	t.Run("publishEvent full-channel path", func(t *testing.T) {
		engine := &Engine{
			history:   event.NewHistory(10),
			eventChan: make(chan *event.Event), // unbuffered => send falls back to default case
			logger:    nil,
		}
		ev := event.NewEvent(event.EventTypeStarted, "cid", "cname", "img")
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("publishEvent should not panic with nil logger: %v", r)
			}
		}()
		engine.publishEvent(ev)
		if got := engine.history.Recent(1); len(got) != 1 || got[0] == nil || got[0].ID != ev.ID {
			t.Fatalf("expected event to be recorded in history, got %#v", got)
		}
	})
}

func TestDidStateChangeHandlesNilOldState(t *testing.T) {
	t.Run("new running container is a change", func(t *testing.T) {
		if !didStateChange(false, nil, "running") {
			t.Fatal("expected state change for new running container")
		}
	})

	t.Run("new non-running container is not treated as start event", func(t *testing.T) {
		if didStateChange(false, nil, "exited") {
			t.Fatal("expected no state change event for new non-running container")
		}
	})

	t.Run("missing oldState with exists flag does not panic", func(t *testing.T) {
		if !didStateChange(true, nil, "running") {
			t.Fatal("expected running state to be treated as change when old state is missing")
		}
		if didStateChange(true, nil, "exited") {
			t.Fatal("expected exited state not to be treated as start change when old state is missing")
		}
	})

	t.Run("existing state compares previous and current values", func(t *testing.T) {
		old := &ContainerState{State: "running"}
		if didStateChange(true, old, "running") {
			t.Fatal("expected no change when state is unchanged")
		}
		if !didStateChange(true, old, "exited") {
			t.Fatal("expected change when state transitions")
		}
	})
}

func TestEngineCollectAndDetectHandlesNilContext(t *testing.T) {
	engine := &Engine{
		dockerClient: &docker.Client{}, // zero-value client returns explicit errors
		logger:       nil,
		ctx:          nil, // partial initialization edge case
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("collectAndDetect should not panic when context is nil: %v", r)
		}
	}()

	engine.collectAndDetect()
}

func TestEngineStopHandlesNilCancelAndEventChannel(t *testing.T) {
	engine := &Engine{
		running:   true,
		cancel:    nil,
		eventChan: nil,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Stop should not panic for partially initialized engine: %v", r)
		}
	}()

	engine.Stop()
	if engine.running {
		t.Fatal("expected running=false after Stop")
	}
	if !engine.terminated {
		t.Fatal("expected terminated=true after Stop")
	}
}

func TestEngineStartInitializesMissingRuntimeDependencies(t *testing.T) {
	engine := &Engine{
		config: EngineConfig{
			Interval:         0,
			ResourceInterval: 0,
			EventHistoryMax:  0,
			Thresholds:       DefaultThresholdConfig(),
		},
		dockerClient: &docker.Client{},
		// runtime dependencies intentionally nil to verify lazy initialization in Start
		ctx:       nil,
		cancel:    nil,
		eventChan: nil,
	}

	if err := engine.Start(); err != nil {
		t.Fatalf("expected Start to initialize missing runtime dependencies: %v", err)
	}
	defer engine.Stop()

	if engine.ctx == nil || engine.cancel == nil {
		t.Fatalf("expected context and cancel to be initialized, got ctx=%#v cancel=%#v", engine.ctx, engine.cancel)
	}
	if engine.eventChan == nil {
		t.Fatal("expected event channel to be initialized")
	}
	if engine.stateManager == nil {
		t.Fatal("expected state manager to be initialized")
	}
	if engine.detector == nil || engine.detector.stateManager == nil {
		t.Fatalf("expected detector to be initialized with state manager, got %#v", engine.detector)
	}
	if engine.thresholdMon == nil {
		t.Fatal("expected threshold monitor to be initialized")
	}
	if engine.history == nil {
		t.Fatal("expected event history to be initialized")
	}
	if engine.config.Interval <= 0 {
		t.Fatalf("expected interval to be defaulted, got %v", engine.config.Interval)
	}
	if engine.config.ResourceInterval <= 0 {
		t.Fatalf("expected resource interval to be defaulted, got %v", engine.config.ResourceInterval)
	}
	if engine.config.EventHistoryMax <= 0 {
		t.Fatalf("expected event history max to be defaulted, got %d", engine.config.EventHistoryMax)
	}
}

func TestEngineStartRebindsDetectorToCurrentStateManager(t *testing.T) {
	primaryStateManager := NewStateManager()
	staleStateManager := NewStateManager()

	engine := &Engine{
		config: EngineConfig{
			Interval:         time.Second,
			ResourceInterval: time.Second,
			EventHistoryMax:  10,
			Thresholds:       DefaultThresholdConfig(),
		},
		dockerClient: &docker.Client{},
		stateManager: primaryStateManager,
		detector:     NewDetector(staleStateManager), // intentionally inconsistent
	}

	if err := engine.Start(); err != nil {
		t.Fatalf("expected Start to recover detector/state manager mismatch: %v", err)
	}
	defer engine.Stop()

	if engine.detector == nil {
		t.Fatal("expected detector to be initialized")
	}
	if engine.detector.stateManager != primaryStateManager {
		t.Fatalf("expected detector to bind current state manager, got %#v want %#v", engine.detector.stateManager, primaryStateManager)
	}
}
