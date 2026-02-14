package monitor

import (
	"context"
	"testing"
	"time"

	"docksphinx/internal/docker"
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
