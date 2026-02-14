package config

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"
)

const (
	minIntervalSeconds         = 1
	minResourceIntervalSeconds = 1
	minHistory                 = 1
	maxHistory                 = 100000
)

// Validate checks config consistency and guardrails.
func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}

	if c.Monitor.Interval < minIntervalSeconds {
		return fmt.Errorf("monitor.interval must be >= %d", minIntervalSeconds)
	}
	if c.Monitor.ResourceInterval != 0 && c.Monitor.ResourceInterval < minResourceIntervalSeconds {
		return fmt.Errorf("monitor.resource_interval must be >= %d", minResourceIntervalSeconds)
	}

	if err := validateRegexList(c.Monitor.Filters.ContainerNames); err != nil {
		return fmt.Errorf("monitor.filters.container_names: %w", err)
	}
	if err := validateRegexList(c.Monitor.Filters.ImageNames); err != nil {
		return fmt.Errorf("monitor.filters.image_names: %w", err)
	}

	if err := validateThresholds(c); err != nil {
		return err
	}

	if err := validateGRPCAddress(c.GRPC.Address); err != nil {
		return err
	}
	if c.GRPC.Timeout <= 0 {
		return fmt.Errorf("grpc.timeout must be > 0")
	}

	if c.Event.MaxHistory < minHistory || c.Event.MaxHistory > maxHistory {
		return fmt.Errorf("event.max_history must be between %d and %d", minHistory, maxHistory)
	}

	if c.Log.Level == "" {
		return fmt.Errorf("log.level must not be empty")
	}
	switch c.Log.Level {
	case "debug", "info", "warn", "warning", "error":
	default:
		return fmt.Errorf("unsupported log.level %q", c.Log.Level)
	}

	if c.Log.File != "" {
		if !filepath.IsAbs(c.Log.File) {
			return fmt.Errorf("log.file must be an absolute path: %s", c.Log.File)
		}
	}

	if c.Daemon.PIDFile == "" {
		return fmt.Errorf("daemon.pid_file must not be empty")
	}
	if !filepath.IsAbs(c.Daemon.PIDFile) {
		return fmt.Errorf("daemon.pid_file must be an absolute path: %s", c.Daemon.PIDFile)
	}
	return nil
}

func validateThresholds(c *Config) error {
	cpu := c.Monitor.Thresholds.CPU
	if cpu.Warning <= 0 || cpu.Critical <= 0 {
		return fmt.Errorf("monitor.thresholds.cpu warning/critical must be > 0")
	}
	if cpu.Warning >= cpu.Critical {
		return fmt.Errorf("monitor.thresholds.cpu warning must be < critical")
	}
	if cpu.ConsecutiveCount <= 0 {
		return fmt.Errorf("monitor.thresholds.cpu.consecutive_count must be > 0")
	}

	mem := c.Monitor.Thresholds.Memory
	if mem.Warning <= 0 || mem.Critical <= 0 {
		return fmt.Errorf("monitor.thresholds.memory warning/critical must be > 0")
	}
	if mem.Warning >= mem.Critical {
		return fmt.Errorf("monitor.thresholds.memory warning must be < critical")
	}
	if mem.ConsecutiveCount <= 0 {
		return fmt.Errorf("monitor.thresholds.memory.consecutive_count must be > 0")
	}

	if c.Monitor.Thresholds.CooldownSeconds < 0 {
		return fmt.Errorf("monitor.thresholds.cooldown_seconds must be >= 0")
	}
	return nil
}

func validateGRPCAddress(addr string) error {
	trimmed := strings.TrimSpace(addr)
	if trimmed == "" {
		return fmt.Errorf("grpc.address must not be empty")
	}
	host, port, err := net.SplitHostPort(trimmed)
	if err != nil {
		return fmt.Errorf("grpc.address must be host:port, got %q", addr)
	}
	if port == "" {
		return fmt.Errorf("grpc.address port is empty")
	}
	if host == "" {
		return fmt.Errorf("grpc.address host is empty")
	}
	return nil
}
