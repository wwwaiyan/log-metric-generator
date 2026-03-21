package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/user/log-metric-generator/internal/config"
	"github.com/user/log-metric-generator/internal/simulator"
)

var (
	configPath = flag.String("config", "config.yaml", "Path to configuration file")
	version    = flag.Bool("version", false, "Show version")
)

func main() {
	flag.Parse()

	if *version {
		printVersion()
		return
	}

	cfg := loadConfig(*configPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sim, err := simulator.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create simulator: %v", err)
	}

	if err := sim.Start(ctx); err != nil {
		log.Fatalf("Failed to start simulator: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
		sim.Stop()
	}

	log.Println("Exiting...")
}

func loadConfig(path string) *config.Config {
	cfg, err := config.Load(path)
	if err != nil {
		log.Printf("Failed to load config from %s: %v, using defaults", path, err)
		cfg = config.DefaultConfig()
	}

	applyRuntimeConfig(cfg)

	return cfg
}

func applyRuntimeConfig(cfg *config.Config) {
	if cfg.Simulator.InstanceID == "" {
		hostname, _ := os.Hostname()
		if hostname != "" {
			cfg.Simulator.InstanceID = hostname
		} else {
			cfg.Simulator.InstanceID = "sim-" + time.Now().Format("150405")
		}
	}

	log.Printf("Configuration loaded:")
	log.Printf("  Instance ID: %s", cfg.Simulator.InstanceID)
	log.Printf("  Log Group: %s", cfg.Simulator.LogGroup)
	log.Printf("  Log Stream: %s", cfg.Simulator.LogStream)
	log.Printf("  Output Mode: %s", cfg.Output.Mode)
	log.Printf("  Web Server: enabled=%v, RPS=%d", cfg.Generators.WebServer.Enabled, cfg.Generators.WebServer.RPS)
	log.Printf("  Error Logs: enabled=%v, error_rate=%.2f%%", cfg.Generators.ErrorLogs.Enabled, cfg.Generators.ErrorLogs.ErrorRate*100)
	log.Printf("  Custom Logs: enabled=%v", cfg.Generators.CustomLogs.Enabled)
	log.Printf("  Metrics: namespace=%s, interval=%ds", cfg.Metrics.Namespace, cfg.Metrics.IntervalSeconds)
}

func printVersion() {
	log.Println("CloudWatch Log/Metric Simulator")
	log.Println("Version: 1.0.0")
	log.Println("")
	log.Println("Usage:")
	log.Println("  simulator [options]")
	log.Println("")
	log.Println("Options:")
	log.Println("  -config <path>   Path to configuration file (default: config.yaml)")
	log.Println("  -version         Show version information")
}
