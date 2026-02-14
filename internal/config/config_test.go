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

func TestSaveAndLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "saved.yaml")

	cfg := Default()
	cfg.Monitor.Interval = 7
	cfg.Event.MaxHistory = 777
	cfg.Daemon.PIDFile = filepath.Join(dir, "docksphinxd.pid")

	if err := cfg.Save(path); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat saved file failed: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected file mode 0600, got %o", info.Mode().Perm())
	}

	loaded, resolved, err := Load(path)
	if err != nil {
		t.Fatalf("load after save failed: %v", err)
	}
	if resolved != path {
		t.Fatalf("resolved path mismatch: got %s, want %s", resolved, path)
	}
	if loaded.Monitor.Interval != 7 {
		t.Fatalf("expected monitor.interval=7, got %d", loaded.Monitor.Interval)
	}
	if loaded.Event.MaxHistory != 777 {
		t.Fatalf("expected event.max_history=777, got %d", loaded.Event.MaxHistory)
	}
}

func TestResolveConfigPathFindsWorkingDirectoryConfig(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "configs")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create configs dir: %v", err)
	}
	configPath := filepath.Join(configDir, "docksphinx.yaml")
	if err := os.WriteFile(configPath, []byte("monitor:\n  interval: 5\n"), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	defer func() { _ = os.Chdir(oldWD) }()

	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg-empty"))
	got, err := ResolveConfigPath("")
	if err != nil {
		t.Fatalf("ResolveConfigPath returned error: %v", err)
	}
	if got != configPath {
		t.Fatalf("expected %s, got %s", configPath, got)
	}
}

func TestResolveConfigPathExplicitMissingReturnsError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.yaml")
	_, err := ResolveConfigPath(missing)
	if err == nil {
		t.Fatal("expected error for missing explicit config path")
	}
}
