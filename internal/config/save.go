package config

import (
	"fmt"
	"os"
	"path/filepath"

	yaml "go.yaml.in/yaml/v4"
)

// Save validates and writes config as YAML.
func (c *Config) Save(path string) error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}
	c.normalize()
	if err := c.Validate(); err != nil {
		return err
	}

	expanded, err := expandPath(path)
	if err != nil {
		return err
	}
	if expanded == "" {
		return fmt.Errorf("save path is empty")
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(expanded), 0o750); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	if err := os.WriteFile(expanded, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
