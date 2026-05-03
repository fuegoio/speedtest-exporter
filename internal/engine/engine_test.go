package engine

import (
	"testing"

	"github.com/alexis/speedtest-exporter/internal/config"
	"github.com/alexis/speedtest-exporter/internal/model"
)

func TestGenerateMeasID(t *testing.T) {
	id1 := generateMeasID()
	id2 := generateMeasID()

	// IDs should be 8 bytes (as string)
	if len(id1) != 8 {
		t.Errorf("Expected ID length 8, got %d", len(id1))
	}

	// IDs should be different (very unlikely to be the same)
	if id1 == id2 {
		t.Error("Generated IDs should be different")
	}
}

func TestComputeMedian(t *testing.T) {
	tests := []struct {
		name     string
		samples  []float64
		expected float64
	}{
		{"empty", []float64{}, 0},
		{"single", []float64{5}, 5},
		{"two", []float64{1, 3}, 2},
		{"three", []float64{1, 2, 3}, 2},
		{"four", []float64{1, 2, 3, 4}, 2.5},
		{"unsorted", []float64{4, 1, 3, 2}, 2.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeMedian(tt.samples)
			if result != tt.expected {
				t.Errorf("computeMedian(%v) = %v, expected %v", tt.samples, result, tt.expected)
			}
		})
	}
}

func TestComputePercentile(t *testing.T) {
	tests := []struct {
		name      string
		samples   []float64
		percentile float64
		expected  float64
	}{
		{"empty", []float64{}, 0.5, 0},
		{"p25", []float64{1, 2, 3, 4, 5}, 0.25, 2},
		{"p50", []float64{1, 2, 3, 4, 5}, 0.5, 3},
		{"p75", []float64{1, 2, 3, 4, 5}, 0.75, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computePercentile(tt.samples, tt.percentile)
			if result != tt.expected {
				t.Errorf("computePercentile(%v, %v) = %v, expected %v", tt.samples, tt.percentile, result, tt.expected)
			}
		})
	}
}

func TestComputeJitter(t *testing.T) {
	tests := []struct {
		name     string
		samples  []float64
		wantZero bool
	}{
		{"empty", []float64{}, true},
		{"single", []float64{5}, true},
		{"constant", []float64{5, 5, 5, 5}, true},
		{"varying", []float64{1, 2, 3, 4, 5}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeJitter(tt.samples)
			if tt.wantZero && result != 0 {
				t.Errorf("computeJitter(%v) = %v, expected 0", tt.samples, result)
			}
			if !tt.wantZero && result == 0 {
				t.Errorf("computeJitter(%v) = 0, expected non-zero", tt.samples)
			}
		})
	}
}

func TestComputeMean(t *testing.T) {
	tests := []struct {
		name     string
		samples  []float64
		expected float64
	}{
		{"empty", []float64{}, 0},
		{"single", []float64{5}, 5},
		{"two", []float64{1, 3}, 2},
		{"three", []float64{1, 2, 3}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeMean(tt.samples)
			if result != tt.expected {
				t.Errorf("computeMean(%v) = %v, expected %v", tt.samples, result, tt.expected)
			}
		})
	}
}

func TestComputeLatencySummary(t *testing.T) {
	c := &CloudflareSpeedtest{
		config: config.ExporterConfig{},
	}

	tests := []struct {
		name     string
		samples  []float64
		wantSent int
		wantLoss float64
	}{
		{"empty", []float64{}, 0, 0},
		{"all valid", []float64{1, 2, 3, 4, 5}, 5, 0},
		{"with loss", []float64{1, -1, 3, -1, 5}, 5, 0.4},
		{"all loss", []float64{-1, -1, -1}, 3, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.computeLatencySummary(tt.samples)
			if result.Sent != tt.wantSent {
				t.Errorf("Sent = %v, expected %v", result.Sent, tt.wantSent)
			}
			if result.Loss != tt.wantLoss {
				t.Errorf("Loss = %v, expected %v", result.Loss, tt.wantLoss)
			}
		})
	}
}

func TestNewCloudflareSpeedtest(t *testing.T) {
	cfg := config.ExporterConfig{
		BaseURL: "https://speed.cloudflare.com",
	}

	speedtest := NewCloudflareSpeedtest(cfg)

	if speedtest == nil {
		t.Error("NewCloudflareSpeedtest returned nil")
	}

	if speedtest.config.BaseURL != cfg.BaseURL {
		t.Errorf("BaseURL = %v, expected %v", speedtest.config.BaseURL, cfg.BaseURL)
	}

	if speedtest.client == nil {
		t.Error("HTTP client is nil")
	}
}

func TestLatencySummaryJSON(t *testing.T) {
	summary := &model.LatencySummary{
		Sent:     10,
		Received: 8,
		Loss:     0.2,
		MinMs:    floatPtr(1.0),
		MaxMs:    floatPtr(10.0),
	}

	// Test that we can create a summary
	if summary.Sent != 10 {
		t.Errorf("Sent = %v, expected 10", summary.Sent)
	}

	if summary.Loss != 0.2 {
		t.Errorf("Loss = %v, expected 0.2", summary.Loss)
	}
}

func TestThroughputSummaryJSON(t *testing.T) {
	summary := &model.ThroughputSummary{
		Bytes:      1000000,
		DurationMs: 1000,
		Mbps:       8.0,
	}

	if summary.Bytes != 1000000 {
		t.Errorf("Bytes = %v, expected 1000000", summary.Bytes)
	}

	if summary.Mbps != 8.0 {
		t.Errorf("Mbps = %v, expected 8.0", summary.Mbps)
	}
}
