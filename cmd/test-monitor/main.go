package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"docksphinx/internal/docker"
	"docksphinx/internal/monitor"
)

func main() {
	dockerClient, err := docker.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}
	defer dockerClient.Close()

	ctx := context.Background()
	if err := dockerClient.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping Docker daemon: %v", err)
	}

	config := monitor.EngineConfig{
		Interval:   5 * time.Second,
		Thresholds: monitor.DefaultThresholdConfig(),
	}

	engine, err := monitor.NewEngine(config, dockerClient)
	if err != nil {
		log.Fatalf("Failed to create engine: %v", err)
	}

	if err := engine.Start(); err != nil {
		log.Fatalf("Failed to start engine: %v", err)
	}
	defer engine.Stop()

	fmt.Println("Monitoring engine started. Press Ctrl+C to stop.")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	eventChan := engine.GetEventChannel()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			fmt.Println("\nStopping monitoring engine...")
			return

		case evt := <-eventChan:
			fmt.Printf("[%s] %s: %s\n",
				evt.Timestamp.Format("15:04:05"),
				evt.Type,
				evt.Message,
			)

		case <-ticker.C:
			states := engine.GetStateManager().GetAllStates()
			fmt.Printf("\n[%s] Tracking %d containers:\n",
				time.Now().Format("15:04:05"),
				len(states),
			)
			for id, state := range states {
				idShort := id
				if len(id) > 12 {
					idShort = id[:12]
				}
				fmt.Printf("  - %s (%s): %s - CPU: %.2f%%, Mem: %.2f%%\n",
					state.ContainerName,
					idShort,
					state.State,
					state.CPUPercent,
					state.MemoryPercent,
				)
			}
		}
	}
}
