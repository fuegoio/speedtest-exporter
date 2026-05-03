package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// ExporterConfig contains all configuration for the speedtest exporter
type ExporterConfig struct {
	// Server configuration
	Port int

	// Test timing configuration
	TestIntervalMs        time.Duration
	BaseURL              string
	DownloadDurationMs   time.Duration
	UploadDurationMs     time.Duration
	IdleLatencyDurationMs time.Duration

	// Concurrency and request sizes
	Concurrency        int
	DownloadBytesPerReq int64
	UploadBytesPerReq   int64

	// Probe configuration
	ProbeIntervalMs time.Duration
	ProbeTimeoutMs  time.Duration

	// Feature flags
	SkipDiagnostics bool
	Traceroute      bool

	// Network information overrides
	Asn           *string
	AsOrg        *string
	InterfaceName *string
	NetworkName  *string
	LocalIpv4    *string
	LocalIpv6    *string
	ExternalIpv4 *string
	ExternalIpv6 *string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() ExporterConfig {
	config := ExporterConfig{
		Port:               getIntEnv("PORT", 9537),
		TestIntervalMs:     getDurationEnv("TEST_INTERVAL_MS", 1*time.Hour),
		BaseURL:            getStringEnv("BASE_URL", "https://speed.cloudflare.com"),
		DownloadDurationMs: getDurationEnv("DOWNLOAD_DURATION_MS", 10*time.Second),
		UploadDurationMs:   getDurationEnv("UPLOAD_DURATION_MS", 10*time.Second),
		IdleLatencyDurationMs: getDurationEnv("IDLE_LATENCY_DURATION_MS", 2*time.Second),
		Concurrency:        getIntEnv("CONCURRENCY", 6),
		DownloadBytesPerReq: int64(getIntEnv("DOWNLOAD_BYTES_PER_REQ", 10000000)),
		UploadBytesPerReq:   int64(getIntEnv("UPLOAD_BYTES_PER_REQ", 5000000)),
		ProbeIntervalMs:    getDurationEnv("PROBE_INTERVAL_MS", 250*time.Millisecond),
		ProbeTimeoutMs:     getDurationEnv("PROBE_TIMEOUT_MS", 800*time.Millisecond),
		SkipDiagnostics:    getBoolEnv("SKIP_DIAGNOSTICS", false),
		Traceroute:         getBoolEnv("TRACEROUTE", false),
	}

	// Network information overrides
	if v := os.Getenv("ASN"); v != "" {
		config.Asn = &v
	}
	if v := os.Getenv("AS_ORG"); v != "" {
		config.AsOrg = &v
	}
	if v := os.Getenv("INTERFACE_NAME"); v != "" {
		config.InterfaceName = &v
	}
	if v := os.Getenv("NETWORK_NAME"); v != "" {
		config.NetworkName = &v
	}
	if v := os.Getenv("LOCAL_IPV4"); v != "" {
		config.LocalIpv4 = &v
	}
	if v := os.Getenv("LOCAL_IPV6"); v != "" {
		config.LocalIpv6 = &v
	}
	if v := os.Getenv("EXTERNAL_IPV4"); v != "" {
		config.ExternalIpv4 = &v
	}
	if v := os.Getenv("EXTERNAL_IPV6"); v != "" {
		config.ExternalIpv6 = &v
	}

	return config
}

func getStringEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if v := os.Getenv(key); v != "" {
		return strings.ToLower(v) == "true"
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v + "ms"); err == nil {
			return d
		}
		// Try parsing as seconds
		if d, err := time.ParseDuration(v + "s"); err == nil {
			return d
		}
	}
	return defaultValue
}
