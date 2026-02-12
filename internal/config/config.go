package config

import (
	"time"

	"docksphinx/internal/monitor"
)

// Config is the root configuration for docksphinxd.
type Config struct {
	GRPCAddress string `yaml:"grpc_address" json:"grpc_address" toml:"grpc_address"`
	PidFile     string `yaml:"pid_file" json:"pid_file" toml:"pid_file"`
	LogLevel    string `yaml:"log_level" json:"log_level" toml:"log_level"`

	Monitor MonitorConfig `yaml:"monitor" json:"monitor" toml:"monitor"`

	EventHistoryMax int `yaml:"event_history_max" json:"event_history_max" toml:"event_history_max"`
}

// MonitorConfig holds monitor engine settings.
type MonitorConfig struct {
	Interval             time.Duration           `yaml:"interval" json:"interval" toml:"interval"`
	ContainerNamePattern string                  `yaml:"container_name_pattern" json:"container_name_pattern" toml:"container_name_pattern"`
	ImageNamePattern     string                  `yaml:"image_name_pattern" json:"image_name_pattern" toml:"image_name_pattern"`
	Thresholds           monitor.ThresholdConfig `yaml:"thresholds" json:"thresholds" toml:"thresholds"`
}
