package config

import (
	"net/url"
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
	TestIntervalMs time.Duration
	BaseURL        string
	Concurrency    int

	// Probe configuration
	ProbeIntervalMs   time.Duration
	ProbeTimeoutMs    time.Duration
	LatencyDurationMs time.Duration

	// DNS diagnostics configuration
	DnsHostname string
	DnsRuns     int
	SkipDns     bool

	// TLS diagnostics configuration
	TlsRuns       int
	SkipTls       bool
	TLSSkipVerify bool // skip TLS certificate verification (for testing only)

	// Network information overrides
	Asn           *string
	AsOrg         *string
	InterfaceName *string
	NetworkName   *string
	LocalIpv4     *string
	LocalIpv6     *string
	ExternalIpv4  *string
	ExternalIpv6  *string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() ExporterConfig {
	baseURL := getStringEnv("BASE_URL", "https://speed.cloudflare.com")
	defaultDnsHostname := getHostnameFromURL(baseURL)

	config := ExporterConfig{
		Port:              getIntEnv("PORT", 9537),
		TestIntervalMs:    getDurationEnv("TEST_INTERVAL_MS", 1*time.Hour),
		BaseURL:           baseURL,
		Concurrency:       getIntEnv("CONCURRENCY", 6),
		ProbeIntervalMs:   getDurationEnv("PROBE_INTERVAL_MS", 250*time.Millisecond),
		ProbeTimeoutMs:    getDurationEnv("PROBE_TIMEOUT_MS", 800*time.Millisecond),
		LatencyDurationMs: getDurationEnv("LATENCY_DURATION_MS", 10*time.Second),
		DnsHostname:       getStringEnv("DNS_HOSTNAME", defaultDnsHostname),
		DnsRuns:           getIntEnv("DNS_RUNS", 10),
		SkipDns:           getBoolEnv("SKIP_DNS", false),
		TlsRuns:           getIntEnv("TLS_RUNS", 10),
		SkipTls:           getBoolEnv("SKIP_TLS", false),
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

func getHostnameFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return u.Hostname()
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
