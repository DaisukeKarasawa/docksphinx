package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "go.yaml.in/yaml/v4"
)

// LoadFile loads config from path.
// Empty path or non-existing path returns default config.
func LoadFile(path string) (*Config, error) {
	if strings.TrimSpace(path) == "" {
		return Default(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), nil
		}
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	cfg, err := unmarshal(data, filepath.Ext(path))
	if err != nil {
		return nil, fmt.Errorf("decode config %q: %w", path, err)
	}
	cfg.MergeWithDefault(Default())
	return cfg, nil
}

func unmarshal(data []byte, ext string) (*Config, error) {
	c := &Config{}
	switch strings.ToLower(ext) {
	case "", ".yaml", ".yml":
		if err := yaml.Unmarshal(data, c); err != nil {
			return nil, err
		}
	case ".json", ".toml":
		return nil, fmt.Errorf("config format %q is not implemented yet; use yaml", ext)
	default:
		return nil, fmt.Errorf("unsupported config extension %q", ext)
	}
	return c, nil
}

// Save writes the config to path as YAML.
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config %q: %w", path, err)
	}
	return nil
}
