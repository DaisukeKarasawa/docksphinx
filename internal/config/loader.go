package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "go.yaml.in/yaml/v4"
)

// Load resolves config path, loads YAML if present, applies defaults, and validates values.
func Load(configPath string) (*Config, string, error) {
	resolved, err := ResolveConfigPath(configPath)
	if err != nil {
		return nil, "", err
	}

	cfg := Default()
	if resolved == "" {
		if err := cfg.Validate(); err != nil {
			return nil, "", err
		}
		return cfg, "", nil
	}

	data, err := os.ReadFile(resolved)
	if err != nil {
		return nil, "", fmt.Errorf("read config %q: %w", resolved, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, "", fmt.Errorf("decode yaml %q: %w", resolved, err)
	}

	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return nil, "", fmt.Errorf("validate config %q: %w", resolved, err)
	}
	return cfg, resolved, nil
}

// ResolveConfigPath resolves a config file path from explicit path or default candidates.
// Returns empty string when no config file exists (caller should use defaults).
func ResolveConfigPath(input string) (string, error) {
	candidates := make([]string, 0, 4)
	if strings.TrimSpace(input) != "" {
		expanded, err := expandPath(input)
		if err != nil {
			return "", err
		}
		candidates = append(candidates, expanded)
	} else {
		defaults, err := DefaultConfigCandidates()
		if err != nil {
			return "", err
		}
		candidates = append(candidates, defaults...)
	}

	for _, p := range candidates {
		info, err := os.Stat(p)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return "", fmt.Errorf("stat config %q: %w", p, err)
		}
		if info.IsDir() {
			return "", fmt.Errorf("config path %q is a directory", p)
		}
		return p, nil
	}

	if strings.TrimSpace(input) != "" {
		return "", fmt.Errorf("config file not found: %s", input)
	}
	return "", nil
}

func DefaultConfigCandidates() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home directory: %w", err)
	}

	xdg := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME"))
	if xdg == "" {
		xdg = filepath.Join(home, ".config")
	}

	return []string{
		"/workspace/configs/docksphinx.yaml",
		"/workspace/configs/docksphinx.yml",
		filepath.Join(xdg, "docksphinx", "config.yaml"),
		filepath.Join(xdg, "docksphinx", "config.yml"),
	}, nil
}

func expandPath(p string) (string, error) {
	trimmed := strings.TrimSpace(p)
	if trimmed == "" {
		return "", nil
	}

	if strings.HasPrefix(trimmed, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home for %q: %w", p, err)
		}
		if trimmed == "~" {
			return home, nil
		}
		return filepath.Join(home, strings.TrimPrefix(trimmed, "~/")), nil
	}
	return trimmed, nil
}

func (c *Config) normalize() {
	if c == nil {
		return
	}
	c.Log.File = strings.TrimSpace(c.Log.File)
	c.Log.Level = strings.TrimSpace(strings.ToLower(c.Log.Level))
	c.GRPC.Address = strings.TrimSpace(c.GRPC.Address)
	c.Daemon.PIDFile = strings.TrimSpace(c.Daemon.PIDFile)
}
