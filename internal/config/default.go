package config

import "docksphinx/internal/monitor"

const (
	defaultIntervalSeconds         = 5
	defaultResourceIntervalSeconds = 15
	defaultGRPCAddress             = "127.0.0.1:50051"
	defaultGRPCTimeoutSeconds      = 30
	defaultLogLevel                = "info"
	defaultEventHistoryMax         = 1000
	defaultPIDFile                 = "/tmp/docksphinxd.pid"
)

// Default returns default configuration.
func Default() *Config {
	return &Config{
		Monitor: MonitorConfig{
			Interval:         defaultIntervalSeconds,
			ResourceInterval: defaultResourceIntervalSeconds,
			Filters: FilterConfig{
				ContainerNames: nil,
				ImageNames:     nil,
			},
			Thresholds: monitor.DefaultThresholdConfig(),
		},
		GRPC: GRPCConfig{
			Address:          defaultGRPCAddress,
			Timeout:          defaultGRPCTimeoutSeconds,
			EnableReflection: false,
		},
		Log: LogConfig{
			Level: defaultLogLevel,
			File:  "",
		},
		Event: EventConfig{
			MaxHistory: defaultEventHistoryMax,
		},
		Daemon: DaemonConfig{
			PIDFile: defaultPIDFile,
		},
	}
}
