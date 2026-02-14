package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfigIsValid(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("default config should be valid: %v", err)
	}
}

func TestValidateRejectsInvalidGRPCAddress(t *testing.T) {
	cfg := Default()
	cfg.GRPC.Address = "invalid"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for invalid grpc address")
	}
}

func TestValidateRejectsInvalidRegex(t *testing.T) {
	cfg := Default()
	cfg.Monitor.Filters.ContainerNames = []string{"["}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for invalid regex")
	}
}

func TestLoadExplicitPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte(`
monitor:
  interval: 3
  resource_interval: 10
  filters:
    container_names: ["^web"]
    image_names: []
  thresholds:
    cpu:
      warning: 70
      critical: 90
      consecutive_count: 2
    memory:
      warning: 80
      critical: 95
      consecutive_count: 2
    cooldown_seconds: 15
grpc:
  address: "127.0.0.1:50051"
  timeout: 20
  enable_reflection: false
log:
  level: "info"
  file: ""
event:
  max_history: 500
daemon:
  pid_file: "/tmp/docksphinxd-test.pid"
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	cfg, resolved, err := Load(path)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if resolved != path {
		t.Fatalf("resolved path mismatch: got %s, want %s", resolved, path)
	}
	if cfg.Monitor.Interval != 3 {
		t.Fatalf("expected monitor.interval=3, got %d", cfg.Monitor.Interval)
	}
	if cfg.Event.MaxHistory != 500 {
		t.Fatalf("expected event.max_history=500, got %d", cfg.Event.MaxHistory)
	}
}
