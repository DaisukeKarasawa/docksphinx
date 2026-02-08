package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"docksphinx/internal/docker"
	"docksphinx/internal/event"
)

// EngineConfig represents monitoring engine configuration
type EngineConfig struct {
	Interval time.Duration // Collection interval

	// Filters
	ContainerNamePattern string // Regex pattern for container names
	ImageNamePattern     string // Regex pattern for image names

	// Thresholds
	Thresholds ThresholdConfig
}

// Engine is the main monitoring engine
type Engine struct {
	config       EngineConfig
	dockerClient *docker.Client
	stateManager *StateManager
	detector     *Detector
	thresholdMon *ThresholdMonitor

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
	stateManager := NewStateManager()
	detector := NewDetector(stateManager)
	thresholdMon := NewThresholdMonitor(config.Thresholds)

	ctx, cancel := context.WithCancel(context.Background())

	return &Engine{
		config:       config,
		dockerClient: dockerClient,
		stateManager: stateManager,
		detector:     detector,
		thresholdMon: thresholdMon,
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

	e.running = true
	e.wg.Add(1)
	go e.monitorLoop()

	return nil
}

// Stop stops the monitoring engine
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running {
		return
	}

	e.running = false
	e.cancel()
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
	ctx, cancel := context.WithTimeout(e.ctx, 30*time.Second)
	defer cancel()

	opts := docker.ListContainersOptions{
		All:          true,
		NamePattern:  e.config.ContainerNamePattern,
		ImagePattern: e.config.ImageNamePattern,
	}

	containers, err := e.dockerClient.ListContainers(ctx, opts)
	if err != nil {
		fmt.Printf("Error listing containers: %v\n", err)
		return
	}

	seenContainers := make(map[string]bool)

	for _, container := range containers {
		seenContainers[container.ID] = true

		oldState, exists := e.stateManager.GetState(container.ID)

		newState := &ContainerState{
			ContainerID:   container.ID,
			ContainerName: container.Name,
			ImageName:     container.Image,
			State:         container.State,
			Status:        container.Status,
			LastSeen:      time.Now(),
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
				select {
				case e.eventChan <- evt:
				default:
					fmt.Printf("Warning: event channel is full, dropping event\n")
				}
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
				select {
				case e.eventChan <- evt:
				default:
					fmt.Printf("Warning: event channel is full, dropping event\n")
				}
			}
		}
	}

	allStates := e.stateManager.GetAllStates()
	for containerID := range allStates {
		if !seenContainers[containerID] {
			e.stateManager.RemoveState(containerID)
		}
	}
}

// GetEventChannel returns the event channel
func (e *Engine) GetEventChannel() <-chan *event.Event {
	return e.eventChan
}

// GetStateManager returns the state manager
func (e *Engine) GetStateManager() *StateManager {
	return e.stateManager
}
