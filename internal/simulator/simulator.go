package simulator

import (
	"context"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/user/log-metric-generator/internal/config"
	"github.com/user/log-metric-generator/internal/generators"
	"github.com/user/log-metric-generator/internal/output"
)

type Simulator struct {
	cfg        *config.Config
	webGen     *generators.WebServerGenerator
	errorGen   *generators.ErrorGenerator
	customGen  *generators.CustomLogGenerator
	metricsGen *generators.MetricsGenerator
	writer     outputWriter
	stopCh     chan struct{}
	wg         sync.WaitGroup
}

type outputWriter interface {
	WriteWebServerLog(line string) error
	WriteErrorLog(line string) error
	WriteCustomLog(message string) error
	WriteMetricEMF(emf string) error
	Flush() error
	Close() error
}

func New(cfg *config.Config) (*Simulator, error) {
	s := &Simulator{
		cfg:       cfg,
		webGen:    generators.NewWebServerGenerator(cfg.Generators.WebServer.Paths),
		errorGen:  generators.NewErrorGenerator(),
		customGen: generators.NewCustomLogGenerator(cfg.Generators.CustomLogs.Format),
		metricsGen: generators.NewMetricsGenerator(
			cfg.Metrics.Namespace,
			cfg.Metrics.CPUBase,
			cfg.Metrics.MemoryBase,
			cfg.Metrics.RequestCount,
			cfg.Metrics.LatencyP50Ms,
			cfg.Metrics.LatencyP99Ms,
		),
		stopCh: make(chan struct{}),
	}

	switch cfg.Output.Mode {
	case "cloudwatch":
		cwCfg := output.CloudWatchConfig{
			Region:    cfg.Output.CloudWatch.Region,
			LogGroup:  cfg.Output.CloudWatch.LogGroup,
			LogStream: cfg.Output.CloudWatch.LogStream,
			Endpoint:  cfg.Output.CloudWatch.Endpoint,
			UseHTTP:   false,
		}
		writer, err := output.NewCloudWatchWriter(cwCfg)
		if err != nil {
			return nil, err
		}
		s.writer = writer
	default:
		s.writer = output.NewStdoutWriter()
	}

	return s, nil
}

func (s *Simulator) Start(ctx context.Context) error {
	log.Printf("Starting simulator...")
	log.Printf("Mode: %s", s.cfg.Output.Mode)

	if s.cfg.Generators.WebServer.Enabled {
		s.wg.Add(1)
		go s.runWebServerGenerator(ctx)
	}

	if s.cfg.Generators.ErrorLogs.Enabled {
		s.wg.Add(1)
		go s.runErrorGenerator(ctx)
	}

	if s.cfg.Generators.CustomLogs.Enabled {
		s.wg.Add(1)
		go s.runCustomLogGenerator(ctx)
	}

	if s.cfg.Metrics.Namespace != "" {
		s.wg.Add(1)
		go s.runMetricsGenerator(ctx)
	}

	return nil
}

func (s *Simulator) runWebServerGenerator(ctx context.Context) {
	defer s.wg.Done()

	interval := time.Second
	if s.cfg.Generators.WebServer.RPS > 0 {
		interval = time.Second / time.Duration(s.cfg.Generators.WebServer.RPS)
		if interval < time.Millisecond*10 {
			interval = time.Millisecond * 10
		}
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.generateWebServerLog()
		}
	}
}

func (s *Simulator) generateWebServerLog() {
	var latencyMs float64
	var statusCode int

	if s.cfg.Generators.ErrorLogs.Enabled && rand.Float64() < s.cfg.Generators.ErrorLogs.ErrorRate {
		statusCode = s.getErrorStatusCode()
		latencyMs = float64(rand.Intn(2000) + 500)
	} else {
		statusCode = 200
		latencyMs = float64(rand.Intn(s.cfg.Metrics.LatencyP99Ms)) + 10
	}

	logEntry := s.webGen.Generate(latencyMs, statusCode)
	line := logEntry.ToApacheFormat()

	if err := s.writer.WriteWebServerLog(line); err != nil {
		log.Printf("Error writing web server log: %v", err)
	}
}

func (s *Simulator) getErrorStatusCode() int {
	codes := []int{400, 401, 403, 404, 500, 502, 503}
	return codes[rand.Intn(len(codes))]
}

func (s *Simulator) runErrorGenerator(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.generateErrorLog()
		}
	}
}

func (s *Simulator) generateErrorLog() {
	errorLog := s.errorGen.Generate("")
	line := errorLog.ToPlainText()

	if err := s.writer.WriteErrorLog(line); err != nil {
		log.Printf("Error writing error log: %v", err)
	}
}

func (s *Simulator) runCustomLogGenerator(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(time.Second * 3)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			message := s.customGen.Generate()
			if err := s.writer.WriteCustomLog(message); err != nil {
				log.Printf("Error writing custom log: %v", err)
			}
		}
	}
}

func (s *Simulator) runMetricsGenerator(ctx context.Context) {
	defer s.wg.Done()

	interval := time.Duration(s.cfg.Metrics.IntervalSeconds) * time.Second
	if interval < time.Second {
		interval = time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	serviceName := "app-service"
	instanceID := s.cfg.Simulator.InstanceID

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			metrics := s.metricsGen.Generate(serviceName, instanceID)
			emf := s.metricsGen.ToEMF(metrics)

			if err := s.writer.WriteMetricEMF(emf); err != nil {
				log.Printf("Error writing metrics: %v", err)
			}

			log.Printf("Generated metrics: CPU=%.2f%%, Memory=%.2f%%, Requests=%d, Latency=%.2fms",
				metrics.Metrics[0].Value,
				metrics.Metrics[1].Value,
				int(metrics.Metrics[2].Value),
				metrics.Metrics[3].Value)
		}
	}
}

func (s *Simulator) Stop() {
	log.Println("Stopping simulator...")
	close(s.stopCh)
	s.wg.Wait()

	if err := s.writer.Close(); err != nil {
		log.Printf("Error closing writer: %v", err)
	}

	log.Println("Simulator stopped")
}
