package config

import (
	"time"

	"docksphinx/internal/monitor"
)

const (
	minMonitorInterval = 1 * time.Second
	minEventHistoryMax = 1
)

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		GRPCAddress:     "127.0.0.1:50051",
		PidFile:         "",
		LogLevel:        "info",
		EventHistoryMax: 1000,
		Monitor: MonitorConfig{
			Interval:             5 * time.Second,
			ContainerNamePattern: "",
			ImageNamePattern:     "",
			Thresholds:           monitor.DefaultThresholdConfig(),
		},
	}
}

// MergeWithDefault overlays non-zero values from file onto default.
func (c *Config) MergeWithDefault(defaults *Config) {
	if defaults == nil {
		return
	}
	if c.GRPCAddress == "" {
		c.GRPCAddress = defaults.GRPCAddress
	}
	if c.PidFile == "" {
		c.PidFile = defaults.PidFile
	}
	if c.LogLevel == "" {
		c.LogLevel = defaults.LogLevel
	}
	if c.EventHistoryMax < minEventHistoryMax {
		c.EventHistoryMax = defaults.EventHistoryMax
	}
	c.Monitor.MergeWithDefault(&defaults.Monitor)
}

// MergeWithDefault overlays monitor defaults for empty values.
func (m *MonitorConfig) MergeWithDefault(d *MonitorConfig) {
	if d == nil {
		return
	}
	if m.Interval < minMonitorInterval {
		m.Interval = d.Interval
	}
	if m.ContainerNamePattern == "" {
		m.ContainerNamePattern = d.ContainerNamePattern
	}
	if m.ImageNamePattern == "" {
		m.ImageNamePattern = d.ImageNamePattern
	}
	if m.Thresholds.CPU.Warning <= 0 {
		m.Thresholds.CPU.Warning = d.Thresholds.CPU.Warning
	}
	if m.Thresholds.CPU.Critical <= 0 {
		m.Thresholds.CPU.Critical = d.Thresholds.CPU.Critical
	}
	if m.Thresholds.CPU.ConsecutiveCount <= 0 {
		m.Thresholds.CPU.ConsecutiveCount = d.Thresholds.CPU.ConsecutiveCount
	}
	if m.Thresholds.Memory.Warning <= 0 {
		m.Thresholds.Memory.Warning = d.Thresholds.Memory.Warning
	}
	if m.Thresholds.Memory.Critical <= 0 {
		m.Thresholds.Memory.Critical = d.Thresholds.Memory.Critical
	}
	if m.Thresholds.Memory.ConsecutiveCount <= 0 {
		m.Thresholds.Memory.ConsecutiveCount = d.Thresholds.Memory.ConsecutiveCount
	}
}
