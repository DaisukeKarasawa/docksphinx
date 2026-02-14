package config

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"docksphinx/internal/monitor"
)

// Config is the root YAML configuration schema.
type Config struct {
	Monitor MonitorConfig `yaml:"monitor"`
	GRPC    GRPCConfig    `yaml:"grpc"`
	Log     LogConfig     `yaml:"log"`
	Event   EventConfig   `yaml:"event"`
	Daemon  DaemonConfig  `yaml:"daemon"`
}

type MonitorConfig struct {
	// Interval is polling interval in seconds.
	Interval int `yaml:"interval"`

	// ResourceInterval is a lighter refresh interval (seconds) for images/networks/volumes.
	// 0 means fallback to Interval.
	ResourceInterval int `yaml:"resource_interval"`

	Filters    FilterConfig            `yaml:"filters"`
	Thresholds monitor.ThresholdConfig `yaml:"thresholds"`
}

type FilterConfig struct {
	ContainerNames []string `yaml:"container_names"`
	ImageNames     []string `yaml:"image_names"`
}

type GRPCConfig struct {
	Address          string `yaml:"address"`
	Timeout          int    `yaml:"timeout"`
	EnableReflection bool   `yaml:"enable_reflection"`
}

type LogConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

type EventConfig struct {
	MaxHistory int `yaml:"max_history"`
}

type DaemonConfig struct {
	PIDFile string `yaml:"pid_file"`
}

// EngineConfig converts config to monitor.EngineConfig.
func (c *Config) EngineConfig() monitor.EngineConfig {
	containerPattern := joinRegexOr(c.Monitor.Filters.ContainerNames)
	imagePattern := joinRegexOr(c.Monitor.Filters.ImageNames)

	interval := time.Duration(c.Monitor.Interval) * time.Second
	resourceInterval := time.Duration(c.Monitor.ResourceInterval) * time.Second
	if c.Monitor.ResourceInterval <= 0 {
		resourceInterval = interval
	}

	return monitor.EngineConfig{
		Interval:             interval,
		ResourceInterval:     resourceInterval,
		ContainerNamePattern: containerPattern,
		ImageNamePattern:     imagePattern,
		Thresholds:           c.Monitor.Thresholds,
	}
}

func joinRegexOr(patterns []string) string {
	parts := make([]string, 0, len(patterns))
	for _, p := range patterns {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("(?:%s)", trimmed))
	}
	return strings.Join(parts, "|")
}

func validateRegexList(patterns []string) error {
	for _, p := range patterns {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		if _, err := regexp.Compile(trimmed); err != nil {
			return fmt.Errorf("invalid regex %q: %w", trimmed, err)
		}
	}
	return nil
}
