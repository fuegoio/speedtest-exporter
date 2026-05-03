package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfigDefaults(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("PORT")
	os.Unsetenv("TEST_INTERVAL_MS")
	os.Unsetenv("BASE_URL")

	cfg := LoadConfig()

	if cfg.Port != 9537 {
		t.Errorf("Expected default port 9537, got %d", cfg.Port)
	}

	if cfg.BaseURL != "https://speed.cloudflare.com" {
		t.Errorf("Expected default BaseURL 'https://speed.cloudflare.com', got %s", cfg.BaseURL)
	}

	if cfg.TestIntervalMs != 1*time.Hour {
		t.Errorf("Expected default TestIntervalMs 1h, got %v", cfg.TestIntervalMs)
	}

	if cfg.Concurrency != 6 {
		t.Errorf("Expected default Concurrency 6, got %d", cfg.Concurrency)
	}

	if cfg.SkipDns != false {
		t.Errorf("Expected default SkipDns false, got %v", cfg.SkipDns)
	}

	if cfg.SkipTls != false {
		t.Errorf("Expected default SkipTls false, got %v", cfg.SkipTls)
	}

	if cfg.DnsRuns != 10 {
		t.Errorf("Expected default DnsRuns 10, got %d", cfg.DnsRuns)
	}

	if cfg.TlsRuns != 10 {
		t.Errorf("Expected default TlsRuns 10, got %d", cfg.TlsRuns)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("PORT", "8080")
	os.Setenv("BASE_URL", "https://example.com")
	os.Setenv("CONCURRENCY", "10")
	os.Setenv("SKIP_DNS", "true")
	os.Setenv("SKIP_TLS", "true")
	os.Setenv("DNS_RUNS", "5")
	os.Setenv("TLS_RUNS", "3")

	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("CONCURRENCY")
		os.Unsetenv("SKIP_DNS")
		os.Unsetenv("SKIP_TLS")
		os.Unsetenv("DNS_RUNS")
		os.Unsetenv("TLS_RUNS")
	}()

	cfg := LoadConfig()

	if cfg.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", cfg.Port)
	}

	if cfg.BaseURL != "https://example.com" {
		t.Errorf("Expected BaseURL 'https://example.com', got %s", cfg.BaseURL)
	}

	if cfg.Concurrency != 10 {
		t.Errorf("Expected Concurrency 10, got %d", cfg.Concurrency)
	}

	if cfg.SkipDns != true {
		t.Errorf("Expected SkipDns true, got %v", cfg.SkipDns)
	}

	if cfg.SkipTls != true {
		t.Errorf("Expected SkipTls true, got %v", cfg.SkipTls)
	}

	if cfg.DnsRuns != 5 {
		t.Errorf("Expected DnsRuns 5, got %d", cfg.DnsRuns)
	}

	if cfg.TlsRuns != 3 {
		t.Errorf("Expected TlsRuns 3, got %d", cfg.TlsRuns)
	}
}

func TestLoadConfigNetworkOverrides(t *testing.T) {
	os.Setenv("ASN", "12345")
	os.Setenv("AS_ORG", "Test ISP")
	os.Setenv("INTERFACE_NAME", "eth0")
	os.Setenv("LOCAL_IPV4", "192.168.1.1")

	defer func() {
		os.Unsetenv("ASN")
		os.Unsetenv("AS_ORG")
		os.Unsetenv("INTERFACE_NAME")
		os.Unsetenv("LOCAL_IPV4")
	}()

	cfg := LoadConfig()

	if cfg.Asn == nil || *cfg.Asn != "12345" {
		t.Errorf("Expected ASN '12345', got %v", cfg.Asn)
	}

	if cfg.AsOrg == nil || *cfg.AsOrg != "Test ISP" {
		t.Errorf("Expected AS_ORG 'Test ISP', got %v", cfg.AsOrg)
	}

	if cfg.InterfaceName == nil || *cfg.InterfaceName != "eth0" {
		t.Errorf("Expected INTERFACE_NAME 'eth0', got %v", cfg.InterfaceName)
	}

	if cfg.LocalIpv4 == nil || *cfg.LocalIpv4 != "192.168.1.1" {
		t.Errorf("Expected LOCAL_IPV4 '192.168.1.1', got %v", cfg.LocalIpv4)
	}
}
