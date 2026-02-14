package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"sync"
	"time"

	"docksphinx/internal/docker"
	"docksphinx/internal/event"
)

// EngineConfig represents monitoring engine configuration
type EngineConfig struct {
	Interval time.Duration // Collection interval
	// ResourceInterval is refresh interval for images/networks/volumes.
	ResourceInterval time.Duration

	// Filters
	ContainerNamePattern string // Regex pattern for container names
	ImageNamePattern     string // Regex pattern for image names

	// Thresholds
	Thresholds ThresholdConfig

	// EventHistoryMax is max in-memory events retained.
	EventHistoryMax int
}

// Engine is the main monitoring engine
type Engine struct {
	config       EngineConfig
	dockerClient *docker.Client
	stateManager *StateManager
	detector     *Detector
	thresholdMon *ThresholdMonitor
	history      *event.History
	logger       *slog.Logger
	resourceMu   sync.Mutex
	lastResource time.Time

	// Event channel for publishing events
	eventChan chan *event.Event

	// Control
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	mu      sync.RWMutex
	running bool
}

// NewEngine creates a new monitoring engine
func NewEngine(config EngineConfig, dockerClient *docker.Client) (*Engine, error) {
	if config.Interval <= 0 {
		config.Interval = 5 * time.Second
	}
	if config.ResourceInterval <= 0 {
		config.ResourceInterval = 15 * time.Second
	}
	if config.EventHistoryMax <= 0 {
		config.EventHistoryMax = 1000
	}

	stateManager := NewStateManager()
	detector := NewDetector(stateManager)
	thresholdMon := NewThresholdMonitor(config.Thresholds)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	ctx, cancel := context.WithCancel(context.Background())

	return &Engine{
		config:       config,
		dockerClient: dockerClient,
		stateManager: stateManager,
		detector:     detector,
		thresholdMon: thresholdMon,
		history:      event.NewHistory(config.EventHistoryMax),
		logger:       logger,
		eventChan:    make(chan *event.Event, 100),
		ctx:          ctx,
		cancel:       cancel,
		running:      false,
	}, nil
}

// Start starts the monitoring engine
func (e *Engine) Start() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return fmt.Errorf("monitoring engine is already running")
	}
	if e.dockerClient == nil {
		return fmt.Errorf("docker client is nil")
	}

	e.running = true
	e.wg.Add(1)
	go e.monitorLoop()

	return nil
}

// Stop stops the monitoring engine
func (e *Engine) Stop() {
	e.mu.Lock()

	if !e.running {
		e.mu.Unlock()
		return
	}

	e.running = false
	e.cancel()
	e.mu.Unlock()

	e.wg.Wait()
	close(e.eventChan)
}

// monitorLoop is the main monitoring loop
func (e *Engine) monitorLoop() {
	defer e.wg.Done()

	ticker := time.NewTicker(e.config.Interval)
	defer ticker.Stop()

	e.collectAndDetect()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.collectAndDetect()
		}
	}
}

// collectAndDetect collects container information and detects events
func (e *Engine) collectAndDetect() {
	if e.dockerClient == nil {
		e.logger.Error("docker client is nil; skipping collection")
		return
	}
	ctx, cancel := context.WithTimeout(e.ctx, 30*time.Second)
	defer cancel()
	now := time.Now()

	opts := docker.ListContainersOptions{
		All:          true,
		NamePattern:  e.config.ContainerNamePattern,
		ImagePattern: e.config.ImageNamePattern,
	}

	containers, err := e.dockerClient.ListContainers(ctx, opts)
	if err != nil {
		e.logger.Error("list containers failed", "error", err)
		return
	}

	seenContainers := make(map[string]bool)

	for _, container := range containers {
		seenContainers[container.ID] = true

		oldState, exists := e.stateManager.GetState(container.ID)

		newState := &ContainerState{
			ContainerID:    container.ID,
			ContainerName:  container.Name,
			ImageName:      container.Image,
			State:          container.State,
			Status:         container.Status,
			LastSeen:       now,
			ComposeProject: container.Labels["com.docker.compose.project"],
			ComposeService: container.Labels["com.docker.compose.service"],
			NetworkNames:   append([]string(nil), container.NetworkNames...),
		}

		if exists {
			newState.StartedAt = oldState.StartedAt
			newState.RestartCount = oldState.RestartCount
			newState.VolumeMountCount = oldState.VolumeMountCount
		}

		if e.needsContainerDetails(exists, oldState, newState) {
			e.fillContainerDetails(ctx, container.ID, newState)
		}
		if !newState.StartedAt.IsZero() {
			newState.UptimeSeconds = int64(now.Sub(newState.StartedAt).Seconds())
			if newState.UptimeSeconds < 0 {
				newState.UptimeSeconds = 0
			}
		}

		if container.State == "running" {
			stats, err := e.dockerClient.GetContainerStats(ctx, container.ID)
			if err == nil {
				newState.CPUPercent = stats.CPUPercent
				newState.MemoryUsage = stats.MemoryUsage
				newState.MemoryLimit = stats.MemoryLimit
				newState.MemoryPercent = stats.MemoryPercent
				newState.NetworkRx = stats.NetworkRx
				newState.NetworkTx = stats.NetworkTx
			}
		}

		if exists {
			newState.CPUThresholdCount = oldState.CPUThresholdCount
			newState.MemoryThresholdCount = oldState.MemoryThresholdCount
		}

		// Detect state changes before update (detector uses GetState, which still has old state)
		stateChanged := !exists && container.State == "running" ||
			(exists && oldState.State != container.State)
		if stateChanged {
			events := e.detector.DetectStateChange(
				container.ID,
				container.Name,
				container.Image,
				container.State,
			)
			for _, evt := range events {
				e.publishEvent(evt)
			}
		}

		e.stateManager.UpdateState(container.ID, newState)

		if container.State == "running" {
			thresholdEvents := e.thresholdMon.CheckThresholds(
				container.ID,
				container.Name,
				container.Image,
				newState.CPUPercent,
				newState.MemoryPercent,
				newState,
			)
			for _, evt := range thresholdEvents {
				e.publishEvent(evt)
			}
		}
	}

	allStates := e.stateManager.GetAllStates()
	for containerID := range allStates {
		if !seenContainers[containerID] {
			e.stateManager.RemoveState(containerID)
		}
	}

	e.collectResources(ctx, now)
}

// GetEventChannel returns the event channel
func (e *Engine) GetEventChannel() <-chan *event.Event {
	return e.eventChan
}

// GetStateManager returns the state manager
func (e *Engine) GetStateManager() *StateManager {
	return e.stateManager
}

// SetLogger overrides engine logger.
func (e *Engine) SetLogger(logger *slog.Logger) {
	if logger == nil {
		return
	}
	e.logger = logger
}

// GetRecentEvents returns recent events from newest to oldest.
func (e *Engine) GetRecentEvents(limit int) []*event.Event {
	return e.history.Recent(limit)
}

func (e *Engine) publishEvent(evt *event.Event) {
	if evt == nil {
		return
	}
	e.history.Add(evt)
	select {
	case e.eventChan <- evt:
	default:
		e.logger.Warn("event channel full; dropping event", "event_type", evt.Type, "container", evt.ContainerName)
	}
}

func (e *Engine) needsContainerDetails(exists bool, oldState *ContainerState, newState *ContainerState) bool {
	if !exists {
		return true
	}
	if oldState == nil {
		return true
	}
	if oldState.State != newState.State {
		return true
	}
	if oldState.StartedAt.IsZero() {
		return true
	}
	if oldState.VolumeMountCount == 0 {
		return true
	}
	return false
}

func (e *Engine) fillContainerDetails(ctx context.Context, containerID string, state *ContainerState) {
	detail, err := e.dockerClient.GetContainerDetails(ctx, containerID)
	if err != nil {
		e.logger.Debug("inspect container failed", "container_id", containerID, "error", err)
		return
	}

	if t, err := time.Parse(time.RFC3339Nano, detail.StartedAt); err == nil {
		state.StartedAt = t
	}
	state.RestartCount = detail.RestartCount

	volumeCount := 0
	for _, m := range detail.Mounts {
		if m.Type == "volume" {
			volumeCount++
		}
	}
	state.VolumeMountCount = volumeCount

	if detail.NetworkSettings != nil {
		networkNames := make([]string, 0, len(detail.NetworkSettings.Networks))
		for name := range detail.NetworkSettings.Networks {
			networkNames = append(networkNames, name)
		}
		sort.Strings(networkNames)
		state.NetworkNames = networkNames
	}

	if detail.Config != nil {
		if project := detail.Config.Labels["com.docker.compose.project"]; project != "" {
			state.ComposeProject = project
		}
		if service := detail.Config.Labels["com.docker.compose.service"]; service != "" {
			state.ComposeService = service
		}
	}
}

func (e *Engine) collectResources(ctx context.Context, now time.Time) {
	e.resourceMu.Lock()
	defer e.resourceMu.Unlock()

	if !e.lastResource.IsZero() && now.Sub(e.lastResource) < e.config.ResourceInterval {
		return
	}

	images, err := e.dockerClient.ListImages(ctx)
	if err != nil {
		e.logger.Warn("list images failed", "error", err)
	}
	networks, err := e.dockerClient.ListNetworks(ctx)
	if err != nil {
		e.logger.Warn("list networks failed", "error", err)
	}
	volumes, err := e.dockerClient.ListVolumes(ctx)
	if err != nil {
		e.logger.Warn("list volumes failed", "error", err)
	}

	for i := range networks {
		inspect, err := e.dockerClient.GetNetwork(ctx, networks[i].ID)
		if err != nil {
			continue
		}
		networks[i].ContainerCount = len(inspect.Containers)
	}

	for i := range volumes {
		inspect, err := e.dockerClient.GetVolume(ctx, volumes[i].Name)
		if err != nil {
			continue
		}
		if inspect.UsageData != nil {
			volumes[i].RefCount = inspect.UsageData.RefCount
		}
	}

	states := e.stateManager.GetAllStates()
	groups := buildComposeGroups(states)
	e.stateManager.UpdateResources(ResourceState{
		Images:   images,
		Networks: networks,
		Volumes:  volumes,
		Groups:   groups,
	})
	e.lastResource = now
}

func buildComposeGroups(states map[string]*ContainerState) []ComposeGroup {
	type groupAcc struct {
		group ComposeGroup
		nets  map[string]struct{}
	}

	groups := map[string]*groupAcc{}
	for id, st := range states {
		if st == nil {
			continue
		}
		project := st.ComposeProject
		service := st.ComposeService
		if project == "" && service == "" {
			for _, netName := range st.NetworkNames {
				if isSystemNetwork(netName) {
					continue
				}
				project = "network:" + netName
				service = "(heuristic)"
				break
			}
		}
		if project == "" {
			project = "(ungrouped)"
		}
		if service == "" {
			service = "(service)"
		}
		key := project + "|" + service
		acc, ok := groups[key]
		if !ok {
			acc = &groupAcc{
				group: ComposeGroup{
					Project:        project,
					Service:        service,
					ContainerIDs:   []string{},
					ContainerNames: []string{},
					NetworkNames:   []string{},
				},
				nets: map[string]struct{}{},
			}
			groups[key] = acc
		}
		acc.group.ContainerIDs = append(acc.group.ContainerIDs, id)
		acc.group.ContainerNames = append(acc.group.ContainerNames, st.ContainerName)
		for _, netName := range st.NetworkNames {
			if netName == "" {
				continue
			}
			acc.nets[netName] = struct{}{}
		}
	}

	out := make([]ComposeGroup, 0, len(groups))
	for _, acc := range groups {
		for netName := range acc.nets {
			acc.group.NetworkNames = append(acc.group.NetworkNames, netName)
		}
		sort.Strings(acc.group.NetworkNames)
		sort.Strings(acc.group.ContainerNames)
		out = append(out, acc.group)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Project == out[j].Project {
			return out[i].Service < out[j].Service
		}
		return out[i].Project < out[j].Project
	})
	return out
}

func isSystemNetwork(name string) bool {
	switch name {
	case "", "bridge", "host", "none":
		return true
	default:
		return false
	}
}
