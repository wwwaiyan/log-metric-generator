package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Simulator   SimulatorConfig   `yaml:"simulator"`
	Generators  GeneratorsConfig  `yaml:"generators"`
	Metrics     MetricsConfig     `yaml:"metrics"`
	Output      OutputConfig      `yaml:"output"`
	HealthCheck HealthCheckConfig `yaml:"health_check"`
}

type SimulatorConfig struct {
	InstanceID    string `yaml:"instance_id"`
	RunDuration   int    `yaml:"run_duration_seconds"`
	FlushInterval int    `yaml:"flush_interval_ms"`
}

type GeneratorsConfig struct {
	WebServer  WebServerConfig  `yaml:"web_server"`
	ErrorLogs  ErrorLogsConfig  `yaml:"error_logs"`
	CustomLogs CustomLogsConfig `yaml:"custom_logs"`
}

type WebServerConfig struct {
	Enabled bool     `yaml:"enabled"`
	RPS     int      `yaml:"rps"`
	Paths   []string `yaml:"paths"`
}

type ErrorLogsConfig struct {
	Enabled   bool    `yaml:"enabled"`
	ErrorRate float64 `yaml:"error_rate"`
}

type CustomLogsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Format  string `yaml:"format"`
}

type MetricsConfig struct {
	Namespace       string  `yaml:"namespace"`
	IntervalSeconds int     `yaml:"interval_seconds"`
	CPUBase         float64 `yaml:"cpu_base"`
	MemoryBase      float64 `yaml:"memory_base"`
	RequestCount    int     `yaml:"request_count"`
	LatencyP50Ms    int     `yaml:"latency_p50_ms"`
	LatencyP99Ms    int     `yaml:"latency_p99_ms"`
}

type OutputConfig struct {
	Mode       string           `yaml:"mode"`
	CloudWatch CloudWatchConfig `yaml:"cloudwatch"`
}

type CloudWatchConfig struct {
	Region    string `yaml:"region"`
	Endpoint  string `yaml:"endpoint"`
	LogGroup  string `yaml:"log_group"`
	LogStream string `yaml:"log_stream"`
}

type HealthCheckConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	applyEnvOverrides(&cfg)

	if cfg.Simulator.InstanceID == "" {
		cfg.Simulator.InstanceID = generateInstanceID()
	}
	if cfg.Simulator.FlushInterval == 0 {
		cfg.Simulator.FlushInterval = 5000
	}
	if cfg.Output.CloudWatch.LogGroup == "" {
		cfg.Output.CloudWatch.LogGroup = "/ecs/simulator"
	}
	if cfg.Output.CloudWatch.LogStream == "" {
		cfg.Output.CloudWatch.LogStream = "app-" + cfg.Simulator.InstanceID
	}

	return &cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("SIMULATOR_LOG_GROUP"); v != "" {
		cfg.Output.CloudWatch.LogGroup = v
	}
	if v := os.Getenv("SIMULATOR_LOG_STREAM"); v != "" {
		cfg.Output.CloudWatch.LogStream = v
	}
	if v := os.Getenv("SIMULATOR_INSTANCE_ID"); v != "" {
		cfg.Simulator.InstanceID = v
	}
	if v := os.Getenv("AWS_REGION"); v != "" {
		cfg.Output.CloudWatch.Region = v
	}

	if v := os.Getenv("GENERATOR_WEB_ENABLED"); v != "" {
		cfg.Generators.WebServer.Enabled = parseBool(v)
	}
	if v := os.Getenv("GENERATOR_WEB_RPS"); v != "" {
		if rps, err := strconv.Atoi(v); err == nil {
			cfg.Generators.WebServer.RPS = rps
		}
	}

	if v := os.Getenv("GENERATOR_ERROR_ENABLED"); v != "" {
		cfg.Generators.ErrorLogs.Enabled = parseBool(v)
	}
	if v := os.Getenv("GENERATOR_ERROR_RATE"); v != "" {
		if rate, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.Generators.ErrorLogs.ErrorRate = rate
		}
	}

	if v := os.Getenv("GENERATOR_CUSTOM_ENABLED"); v != "" {
		cfg.Generators.CustomLogs.Enabled = parseBool(v)
	}

	if v := os.Getenv("METRICS_NAMESPACE"); v != "" {
		cfg.Metrics.Namespace = v
	}
	if v := os.Getenv("METRICS_INTERVAL"); v != "" {
		if interval, err := strconv.Atoi(v); err == nil {
			cfg.Metrics.IntervalSeconds = interval
		}
	}
	if v := os.Getenv("METRICS_CPU_BASE"); v != "" {
		if cpu, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.Metrics.CPUBase = cpu
		}
	}
	if v := os.Getenv("METRICS_MEMORY_BASE"); v != "" {
		if mem, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.Metrics.MemoryBase = mem
		}
	}

	if v := os.Getenv("OUTPUT_MODE"); v != "" {
		cfg.Output.Mode = v
	}

	if v := os.Getenv("HEALTH_CHECK_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.HealthCheck.Port = port
		}
	}
}

func parseBool(s string) bool {
	return strings.ToLower(s) == "true" || s == "1"
}

func generateInstanceID() string {
	return "sim-" + strconv.FormatInt(time.Now().UnixNano()%1000000, 10)
}

func DefaultConfig() *Config {
	return &Config{
		Simulator: SimulatorConfig{
			FlushInterval: 5000,
		},
		Generators: GeneratorsConfig{
			WebServer: WebServerConfig{
				Enabled: true,
				RPS:     10,
				Paths:   []string{"/api/users", "/api/orders", "/health", "/api/products"},
			},
			ErrorLogs: ErrorLogsConfig{
				Enabled:   true,
				ErrorRate: 0.05,
			},
			CustomLogs: CustomLogsConfig{
				Enabled: true,
				Format:  "json",
			},
		},
		Metrics: MetricsConfig{
			Namespace:       "TestApp/Metrics",
			IntervalSeconds: 60,
			CPUBase:         45.0,
			MemoryBase:      65.0,
			RequestCount:    1000,
			LatencyP50Ms:    120,
			LatencyP99Ms:    500,
		},
		Output: OutputConfig{
			Mode: "stdout",
			CloudWatch: CloudWatchConfig{
				Region:    "us-east-1",
				LogGroup:  "/ecs/simulator",
				LogStream: "app-default",
			},
		},
		HealthCheck: HealthCheckConfig{
			Enabled: true,
			Port:    8080,
		},
	}
}
