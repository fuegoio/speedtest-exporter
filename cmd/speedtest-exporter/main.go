package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/alexis/speedtest-exporter/internal/config"
	"github.com/alexis/speedtest-exporter/internal/engine"
	"github.com/alexis/speedtest-exporter/internal/metrics"
)

// SpeedtestExporter is the main exporter struct
type SpeedtestExporter struct {
	config      config.ExporterConfig
	speedtest   *engine.CloudflareSpeedtest
	running    bool
	stopChan   chan struct{}
	cancelFunc context.CancelFunc
}

// NewSpeedtestExporter creates a new exporter instance
func NewSpeedtestExporter(cfg config.ExporterConfig) *SpeedtestExporter {
	return &SpeedtestExporter{
		config:    cfg,
		speedtest: engine.NewCloudflareSpeedtest(cfg),
		running:  false,
		stopChan: make(chan struct{}),
	}
}

// Start starts the exporter
func (e *SpeedtestExporter) Start() error {
	if e.running {
		return nil
	}
	e.running = true

	log.Printf("Starting Speedtest Exporter on port %d", e.config.Port)
	log.Printf("Test interval: %v", e.config.TestIntervalMs)

	// Register metrics
	reg := prometheus.NewRegistry()
	if err := metrics.RegisterAll(reg); err != nil {
		return fmt.Errorf("failed to register metrics: %w", err)
	}

	// Create HTTP server
	mux := http.NewServeMux()
	mux.Handle("/metrics", metrics.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		// Manually trigger a test
		result, err := e.speedtest.RunDirectTest()
		if err != nil {
			metrics.IncrementError(err.Error())
			metrics.IncrementRun("failed")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		metrics.UpdateMetrics(result)
		metrics.IncrementRun("success")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Speedtest Exporter\n\nEndpoints:\n")
		fmt.Fprintf(w, "/metrics - Prometheus metrics\n")
		fmt.Fprintf(w, "/health - Health check\n")
		fmt.Fprintf(w, "/run - Run test manually\n")
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", e.config.Port),
		Handler: mux,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server started on http://localhost:%d", e.config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	// Start periodic tests
	e.startPeriodicTests()

	return nil
}

// startPeriodicTests starts the periodic speed tests
func (e *SpeedtestExporter) startPeriodicTests() {
	if e.running {
		return
	}
	e.running = true

	log.Printf("Starting periodic speed tests every %v", e.config.TestIntervalMs)

	// Run initial test immediately
	e.runTest()

	// Then run on interval
	ticker := time.NewTicker(e.config.TestIntervalMs)
	go func() {
		for {
			select {
			case <-ticker.C:
				e.runTest()
			case <-e.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

// runTest runs a single speed test
func (e *SpeedtestExporter) runTest() {
	log.Printf("Running speed test at %s", time.Now().Format(time.RFC3339))

	result, err := e.speedtest.RunDirectTest()
	if err != nil {
		log.Printf("Test failed: %v", err)
		metrics.IncrementError(err.Error())
		metrics.IncrementRun("failed")
		return
	}

	metrics.UpdateMetrics(result)
	log.Printf(
		"Test completed: Download=%.2f Mbps, Upload=%.2f Mbps",
		result.Download.Mbps,
		result.Upload.Mbps,
	)
}

// Shutdown stops the exporter
func (e *SpeedtestExporter) Shutdown() {
	if !e.running {
		return
	}
	e.running = false

	// Stop periodic tests
	close(e.stopChan)

	// Stop HTTP server if we have one
	if e.cancelFunc != nil {
		e.cancelFunc()
	}

	log.Println("Exporter shutdown")
}

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Create exporter
	exporter := NewSpeedtestExporter(cfg)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutdown signal received")
		exporter.Shutdown()
		os.Exit(0)
	}()

	// Start exporter
	if err := exporter.Start(); err != nil {
		log.Fatalf("Failed to start exporter: %v", err)
	}

	// Wait forever
	select {}
}
