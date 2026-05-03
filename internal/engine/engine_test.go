package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/fuegoio/speedtest-exporter/internal/config"
	"github.com/fuegoio/speedtest-exporter/internal/model"
)

// ---------------------------------------------------------------------------
// Mock Cloudflare server
// ---------------------------------------------------------------------------

// mockCloudflareServer creates a test HTTP server that mimics the Cloudflare
// speed test endpoints: /meta, /__down, /__up, /__latency.
func mockCloudflareServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	// /meta — returns a fixed JSON payload matching metaResponse
	mux.HandleFunc("/meta", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"clientIp":       "1.2.3.4",
			"asn":            12345,
			"asOrganization": "Test ISP",
			"country":        "US",
			"city":           "San Francisco",
			"region":         "California",
			"postalCode":     "94105",
			"latitude":       "37.7749",
			"longitude":      "-122.4194",
			"colo": map[string]any{
				"iata": "SFO",
			},
		})
	})

	// /__down?bytes=N — streams exactly N zero bytes
	mux.HandleFunc("/__down", func(w http.ResponseWriter, r *http.Request) {
		n, _ := strconv.ParseInt(r.URL.Query().Get("bytes"), 10, 64)
		if n <= 0 {
			n = 1024
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", n))
		io.Copy(w, io.LimitReader(bytes.NewReader(make([]byte, n)), n))
	})

	// /__up — reads the request body and returns 200
	mux.HandleFunc("/__up", func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
		w.WriteHeader(http.StatusOK)
	})

	// /__latency — returns a tiny response immediately
	mux.HandleFunc("/__latency", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("ok"))
	})

	return httptest.NewServer(mux)
}

// newTestEngine creates a CloudflareSpeedtest pointed at the given base URL.
func newTestEngine(baseURL string) *CloudflareSpeedtest {
	cfg := config.ExporterConfig{
		BaseURL:        baseURL,
		ProbeTimeoutMs: 5 * time.Second,
	}
	return NewCloudflareSpeedtest(cfg)
}

// ---------------------------------------------------------------------------
// Pure math helpers
// ---------------------------------------------------------------------------

func TestComputeMedian(t *testing.T) {
	cases := []struct {
		samples  []float64
		expected float64
	}{
		{[]float64{}, 0},
		{[]float64{5}, 5},
		{[]float64{1, 3}, 2},
		{[]float64{1, 2, 3}, 2},
		{[]float64{1, 2, 3, 4}, 2.5},
		// computeMedian does NOT sort — caller must sort first
		{[]float64{1, 2, 3, 4, 5}, 3},
	}
	for _, tc := range cases {
		got := computeMedian(tc.samples)
		if got != tc.expected {
			t.Errorf("computeMedian(%v) = %v, want %v", tc.samples, got, tc.expected)
		}
	}
}

func TestComputePercentile(t *testing.T) {
	cases := []struct {
		samples    []float64
		percentile float64
		expected   float64
	}{
		{[]float64{}, 0.5, 0},
		{[]float64{1, 2, 3, 4, 5}, 0.25, 2},
		{[]float64{1, 2, 3, 4, 5}, 0.50, 3},
		{[]float64{1, 2, 3, 4, 5}, 0.75, 4},
		{[]float64{10}, 0.99, 10},
	}
	for _, tc := range cases {
		got := computePercentile(tc.samples, tc.percentile)
		if got != tc.expected {
			t.Errorf("computePercentile(%v, %.2f) = %v, want %v", tc.samples, tc.percentile, got, tc.expected)
		}
	}
}

func TestComputeMean(t *testing.T) {
	cases := []struct {
		samples  []float64
		expected float64
	}{
		{[]float64{}, 0},
		{[]float64{5}, 5},
		{[]float64{1, 3}, 2},
		{[]float64{1, 2, 3}, 2},
		{[]float64{0, 0, 0}, 0},
	}
	for _, tc := range cases {
		got := computeMean(tc.samples)
		if got != tc.expected {
			t.Errorf("computeMean(%v) = %v, want %v", tc.samples, got, tc.expected)
		}
	}
}

func TestComputeJitter(t *testing.T) {
	cases := []struct {
		samples  []float64
		wantZero bool
	}{
		{[]float64{}, true},
		{[]float64{5}, true},
		{[]float64{5, 5, 5, 5}, true},
		{[]float64{1, 2, 3, 4, 5}, false},
	}
	for _, tc := range cases {
		got := computeJitter(tc.samples)
		if tc.wantZero && got != 0 {
			t.Errorf("computeJitter(%v) = %v, want 0", tc.samples, got)
		}
		if !tc.wantZero && got == 0 {
			t.Errorf("computeJitter(%v) = 0, want non-zero", tc.samples)
		}
	}
}

// ---------------------------------------------------------------------------
// computeLatencySummary
// ---------------------------------------------------------------------------

func TestComputeLatencySummary(t *testing.T) {
	c := &CloudflareSpeedtest{config: config.ExporterConfig{}}

	cases := []struct {
		name         string
		samples      []float64
		wantSent     int
		wantReceived int
		wantLoss     float64
		wantNilStats bool // true when all samples are loss
	}{
		{"empty", []float64{}, 0, 0, 0, true},
		{"all valid", []float64{10, 20, 30, 40, 50}, 5, 5, 0, false},
		{"with loss", []float64{10, -1, 30, -1, 50}, 5, 3, 0.4, false},
		{"all loss", []float64{-1, -1, -1}, 3, 0, 1.0, true},
		{"single valid", []float64{42}, 1, 1, 0, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := c.computeLatencySummary(tc.samples)
			if got.Sent != tc.wantSent {
				t.Errorf("Sent = %d, want %d", got.Sent, tc.wantSent)
			}
			if got.Received != tc.wantReceived {
				t.Errorf("Received = %d, want %d", got.Received, tc.wantReceived)
			}
			if got.Loss != tc.wantLoss {
				t.Errorf("Loss = %v, want %v", got.Loss, tc.wantLoss)
			}
			if tc.wantNilStats {
				if got.MinMs != nil || got.MaxMs != nil || got.MeanMs != nil {
					t.Error("expected nil stats for all-loss or empty input")
				}
			} else {
				if got.MinMs == nil || got.MaxMs == nil || got.MeanMs == nil || got.MedianMs == nil {
					t.Error("expected non-nil stats for valid samples")
				}
				// Min <= Mean <= Max
				if *got.MinMs > *got.MeanMs || *got.MeanMs > *got.MaxMs {
					t.Errorf("stats ordering violated: min=%v mean=%v max=%v", *got.MinMs, *got.MeanMs, *got.MaxMs)
				}
			}
		})
	}
}

func TestComputeLatencySummaryOrdering(t *testing.T) {
	// Verify that unsorted input still produces correct min/max
	c := &CloudflareSpeedtest{config: config.ExporterConfig{}}
	got := c.computeLatencySummary([]float64{50, 10, 30, 20, 40})
	if *got.MinMs != 10 {
		t.Errorf("MinMs = %v, want 10", *got.MinMs)
	}
	if *got.MaxMs != 50 {
		t.Errorf("MaxMs = %v, want 50", *got.MaxMs)
	}
}

// ---------------------------------------------------------------------------
// generateMeasID
// ---------------------------------------------------------------------------

func TestGenerateMeasID(t *testing.T) {
	id1 := generateMeasID()
	id2 := generateMeasID()

	if len(id1) != 8 {
		t.Errorf("expected ID length 8, got %d", len(id1))
	}
	if id1 == id2 {
		t.Error("two generated IDs should differ")
	}
}

// ---------------------------------------------------------------------------
// getHostname
// ---------------------------------------------------------------------------

func TestGetHostname(t *testing.T) {
	cases := []struct {
		url      string
		expected string
	}{
		{"https://speed.cloudflare.com", "speed.cloudflare.com"},
		{"http://localhost:8080", "localhost"},
		{"https://example.com/path?q=1", "example.com"},
		// url.Parse parses "not-a-url" as a relative path with empty host
		{"not-a-url", ""},
	}
	for _, tc := range cases {
		got := getHostname(tc.url)
		if got != tc.expected {
			t.Errorf("getHostname(%q) = %q, want %q", tc.url, got, tc.expected)
		}
	}
}

// ---------------------------------------------------------------------------
// getMetaInfo — uses mock server
// ---------------------------------------------------------------------------

func TestGetMetaInfo(t *testing.T) {
	srv := mockCloudflareServer(t)
	defer srv.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	externalIP, colo, asn, asOrg, country, city, region, postalCode, lat, lon, err :=
		getMetaInfo(srv.URL, client)

	if err != nil {
		t.Fatalf("getMetaInfo returned error: %v", err)
	}

	checks := []struct{ got, want string }{
		{externalIP, "1.2.3.4"},
		{colo, "SFO"},
		{asn, "AS12345"},
		{asOrg, "Test ISP"},
		{country, "US"},
		{city, "San Francisco"},
		{region, "California"},
		{postalCode, "94105"},
		{lat, "37.7749"},
		{lon, "-122.4194"},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("got %q, want %q", c.got, c.want)
		}
	}
}

func TestGetMetaInfoError(t *testing.T) {
	// Point at a server that immediately closes
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	client := &http.Client{Timeout: 2 * time.Second}
	_, _, _, _, _, _, _, _, _, _, err := getMetaInfo(srv.URL, client)
	if err == nil {
		t.Error("expected error for invalid JSON response, got nil")
	}
}

// ---------------------------------------------------------------------------
// measureSingleFetch — uses mock server
// ---------------------------------------------------------------------------

func TestMeasureSingleFetch(t *testing.T) {
	srv := mockCloudflareServer(t)
	defer srv.Close()

	engine := newTestEngine(srv.URL)

	cases := []struct {
		bytes int64
	}{
		{1024},
		{100 * 1024},
		{1024 * 1024},
	}

	for _, tc := range cases {
		url := fmt.Sprintf("%s/__down?bytes=%d", srv.URL, tc.bytes)
		result := engine.measureSingleFetch(url)

		if result.Bytes != tc.bytes {
			t.Errorf("bytes=%d: got Bytes=%d, want %d", tc.bytes, result.Bytes, tc.bytes)
		}
		if result.DurationMs < 0 {
			t.Errorf("bytes=%d: negative DurationMs %d", tc.bytes, result.DurationMs)
		}
		if result.Mbps < 0 {
			t.Errorf("bytes=%d: negative Mbps %f", tc.bytes, result.Mbps)
		}
		// Mbps = (bytes * 8) / (durationSec * 1e6) — must be positive for non-zero duration
		if result.DurationMs > 0 && result.Mbps == 0 {
			t.Errorf("bytes=%d: expected non-zero Mbps when duration > 0", tc.bytes)
		}
	}
}

func TestMeasureSingleFetchBadURL(t *testing.T) {
	engine := newTestEngine("http://127.0.0.1:1") // nothing listening
	result := engine.measureSingleFetch("http://127.0.0.1:1/__down?bytes=1024")
	// Should return empty summary, not panic
	if result == nil {
		t.Fatal("expected non-nil result even on error")
	}
	if result.Bytes != 0 || result.Mbps != 0 {
		t.Errorf("expected zero result on connection error, got %+v", result)
	}
}

// ---------------------------------------------------------------------------
// measureSingleUpload — uses mock server
// ---------------------------------------------------------------------------

func TestMeasureSingleUpload(t *testing.T) {
	srv := mockCloudflareServer(t)
	defer srv.Close()

	engine := newTestEngine(srv.URL)
	url := fmt.Sprintf("%s/__up?bytes=102400", srv.URL)

	cases := []int64{1024, 100 * 1024, 1024 * 1024}
	for _, size := range cases {
		result := engine.measureSingleUpload(url, size)
		if result.Bytes != size {
			t.Errorf("size=%d: got Bytes=%d, want %d", size, result.Bytes, size)
		}
		if result.DurationMs < 0 {
			t.Errorf("size=%d: negative DurationMs", size)
		}
		if result.Mbps < 0 {
			t.Errorf("size=%d: negative Mbps", size)
		}
	}
}

func TestMeasureSingleUploadBadURL(t *testing.T) {
	engine := newTestEngine("http://127.0.0.1:1")
	result := engine.measureSingleUpload("http://127.0.0.1:1/__up", 1024)
	if result == nil {
		t.Fatal("expected non-nil result even on error")
	}
	if result.Bytes != 0 || result.Mbps != 0 {
		t.Errorf("expected zero result on connection error, got %+v", result)
	}
}

// ---------------------------------------------------------------------------
// measureLatency — uses mock server
// ---------------------------------------------------------------------------

func TestMeasureLatency(t *testing.T) {
	srv := mockCloudflareServer(t)
	defer srv.Close()

	engine := newTestEngine(srv.URL)
	url := fmt.Sprintf("%s/__latency", srv.URL)

	// Run for 500ms with 100ms interval → expect ~5 samples
	result := engine.measureLatency(url, 500*time.Millisecond, 100*time.Millisecond)

	if result == nil {
		t.Fatal("expected non-nil latency result")
	}
	if result.Sent == 0 {
		t.Error("expected at least one probe sent")
	}
	if result.Loss != 0 {
		t.Errorf("expected 0 loss against local mock, got %v", result.Loss)
	}
	if result.MinMs == nil || *result.MinMs < 0 {
		t.Error("expected non-negative MinMs")
	}
	if result.MaxMs == nil || *result.MaxMs < *result.MinMs {
		t.Error("MaxMs should be >= MinMs")
	}
}

func TestMeasureLatencyAllLoss(t *testing.T) {
	// Server that always errors
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Hijack and close to force a client error
		hj, ok := w.(http.Hijacker)
		if !ok {
			w.WriteHeader(500)
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close()
	}))
	defer srv.Close()

	engine := newTestEngine(srv.URL)
	url := fmt.Sprintf("%s/__latency", srv.URL)

	result := engine.measureLatency(url, 300*time.Millisecond, 100*time.Millisecond)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Loss != 1.0 {
		t.Errorf("expected 100%% loss, got %.2f", result.Loss)
	}
	if result.MinMs != nil {
		t.Error("expected nil MinMs for all-loss result")
	}
}

// ---------------------------------------------------------------------------
// NewCloudflareSpeedtest
// ---------------------------------------------------------------------------

func TestNewCloudflareSpeedtest(t *testing.T) {
	cfg := config.ExporterConfig{
		BaseURL:        "https://speed.cloudflare.com",
		ProbeTimeoutMs: 10 * time.Second,
	}
	st := NewCloudflareSpeedtest(cfg)

	if st == nil {
		t.Fatal("NewCloudflareSpeedtest returned nil")
	}
	if st.config.BaseURL != cfg.BaseURL {
		t.Errorf("BaseURL = %q, want %q", st.config.BaseURL, cfg.BaseURL)
	}
	if st.client == nil {
		t.Error("HTTP client is nil")
	}
}

// ---------------------------------------------------------------------------
// RunDirectTest — full end-to-end with mock server
// ---------------------------------------------------------------------------

func TestRunDirectTest(t *testing.T) {
	srv := mockCloudflareServer(t)
	defer srv.Close()

	cfg := config.ExporterConfig{
		BaseURL:        srv.URL,
		ProbeTimeoutMs: 2 * time.Second,
	}
	engine := NewCloudflareSpeedtest(cfg)

	result, err := engine.RunDirectTest()
	if err != nil {
		t.Fatalf("RunDirectTest failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Meta fields from mock
	if result.Server == nil || *result.Server != "US" {
		t.Errorf("Server = %v, want 'US'", result.Server)
	}
	if result.Colo == nil || *result.Colo != "SFO" {
		t.Errorf("Colo = %v, want 'SFO'", result.Colo)
	}
	if result.IP == nil || *result.IP != "1.2.3.4" {
		t.Errorf("IP = %v, want '1.2.3.4'", result.IP)
	}
	if result.Country == nil || *result.Country != "US" {
		t.Errorf("Country = %v, want 'US'", result.Country)
	}
	if result.City == nil || *result.City != "San Francisco" {
		t.Errorf("City = %v, want 'San Francisco'", result.City)
	}
	if result.Asn == nil || *result.Asn != "AS12345" {
		t.Errorf("Asn = %v, want 'AS12345'", result.Asn)
	}
	if result.AsOrg == nil || *result.AsOrg != "Test ISP" {
		t.Errorf("AsOrg = %v, want 'Test ISP'", result.AsOrg)
	}

	// Throughput
	if result.Download.Bytes == 0 {
		t.Error("expected non-zero download bytes")
	}
	if result.Upload.Bytes == 0 {
		t.Error("expected non-zero upload bytes")
	}

	// Latency
	if result.IdleLatency.Sent == 0 {
		t.Error("expected idle latency probes to be sent")
	}
	if result.IdleLatency.Loss != 0 {
		t.Errorf("expected 0 idle latency loss against mock, got %v", result.IdleLatency.Loss)
	}

	// MeasID is 8 bytes
	if len(result.MeasID) != 8 {
		t.Errorf("MeasID length = %d, want 8", len(result.MeasID))
	}

	// Timestamp is set
	if result.TimestampUTC == "" {
		t.Error("expected non-empty TimestampUTC")
	}
	if result.BaseURL != srv.URL {
		t.Errorf("BaseURL = %q, want %q", result.BaseURL, srv.URL)
	}
}

func TestRunDirectTestConfigOverrides(t *testing.T) {
	srv := mockCloudflareServer(t)
	defer srv.Close()

	customAsn := "AS99999"
	customOrg := "My ISP"
	customIface := "eth0"
	customNet := "home"
	customIPv4 := "10.0.0.1"
	customIPv6 := "::1"
	customExtIPv4 := "5.6.7.8"
	customExtIPv6 := "2001:db8::1"

	cfg := config.ExporterConfig{
		BaseURL:        srv.URL,
		ProbeTimeoutMs: 2 * time.Second,
		Asn:            &customAsn,
		AsOrg:          &customOrg,
		InterfaceName:  &customIface,
		NetworkName:    &customNet,
		LocalIpv4:      &customIPv4,
		LocalIpv6:      &customIPv6,
		ExternalIpv4:   &customExtIPv4,
		ExternalIpv6:   &customExtIPv6,
	}
	engine := NewCloudflareSpeedtest(cfg)

	result, err := engine.RunDirectTest()
	if err != nil {
		t.Fatalf("RunDirectTest failed: %v", err)
	}

	checks := []struct {
		field string
		got   *string
		want  string
	}{
		{"Asn", result.Asn, customAsn},
		{"AsOrg", result.AsOrg, customOrg},
		{"InterfaceName", result.InterfaceName, customIface},
		{"NetworkName", result.NetworkName, customNet},
		{"LocalIpv4", result.LocalIpv4, customIPv4},
		{"LocalIpv6", result.LocalIpv6, customIPv6},
		{"ExternalIpv4", result.ExternalIpv4, customExtIPv4},
		{"ExternalIpv6", result.ExternalIpv6, customExtIPv6},
	}
	for _, c := range checks {
		if c.got == nil || *c.got != c.want {
			t.Errorf("%s = %v, want %q", c.field, c.got, c.want)
		}
	}
}

func TestRunDirectTestMetaFailure(t *testing.T) {
	// Server that returns 500 for /meta
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/meta") {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error"))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := config.ExporterConfig{
		BaseURL:        srv.URL,
		ProbeTimeoutMs: 2 * time.Second,
	}
	engine := NewCloudflareSpeedtest(cfg)

	_, err := engine.RunDirectTest()
	if err == nil {
		t.Error("expected error when /meta returns invalid JSON")
	}
}

// ---------------------------------------------------------------------------
// measureTls — uses TLS mock server
// ---------------------------------------------------------------------------

func TestMeasureTls(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := config.ExporterConfig{
		BaseURL:        srv.URL,
		ProbeTimeoutMs: 5 * time.Second,
		TLSSkipVerify:  true,
	}
	engine := NewCloudflareSpeedtest(cfg)

	result := engine.measureTls()
	if result == nil {
		t.Fatal("expected non-nil TlsSummary")
	}
	if len(result.HandshakeTimeSamples) != 10 {
		t.Errorf("HandshakeTimeSamples count = %d, want 10", len(result.HandshakeTimeSamples))
	}
	for i, ms := range result.HandshakeTimeSamples {
		if ms < 0 {
			t.Errorf("sample[%d] = %v, want >= 0", i, ms)
		}
	}
	if result.ProtocolVersion == nil || *result.ProtocolVersion == "" {
		t.Error("expected non-empty ProtocolVersion")
	}
	if result.CipherSuite == nil || *result.CipherSuite == "" {
		t.Error("expected non-empty CipherSuite")
	}
	// Mean should be within the range of samples
	if result.HandshakeTimeMs < 0 {
		t.Errorf("HandshakeTimeMs = %v, want >= 0", result.HandshakeTimeMs)
	}
}

func TestMeasureTlsBadURL(t *testing.T) {
	engine := newTestEngine("http://127.0.0.1:1")
	result := engine.measureTls()
	// All 10 attempts fail → nil
	if result != nil {
		t.Errorf("expected nil result for unreachable server, got %+v", result)
	}
}

// ---------------------------------------------------------------------------
// Model types
// ---------------------------------------------------------------------------

func TestLatencySummaryFields(t *testing.T) {
	min := 1.0
	max := 10.0
	summary := &model.LatencySummary{
		Sent:     10,
		Received: 8,
		Loss:     0.2,
		MinMs:    &min,
		MaxMs:    &max,
	}
	if summary.Sent != 10 {
		t.Errorf("Sent = %d, want 10", summary.Sent)
	}
	if summary.Loss != 0.2 {
		t.Errorf("Loss = %v, want 0.2", summary.Loss)
	}
	if *summary.MinMs != 1.0 {
		t.Errorf("MinMs = %v, want 1.0", *summary.MinMs)
	}
}

func TestThroughputSummaryFields(t *testing.T) {
	summary := &model.ThroughputSummary{
		Bytes:      1_000_000,
		DurationMs: 1000,
		Mbps:       8.0,
	}
	if summary.Bytes != 1_000_000 {
		t.Errorf("Bytes = %d, want 1000000", summary.Bytes)
	}
	if summary.Mbps != 8.0 {
		t.Errorf("Mbps = %v, want 8.0", summary.Mbps)
	}
}

// ---------------------------------------------------------------------------
// Integration test (network required, skipped in -short mode)
// ---------------------------------------------------------------------------

// TestSpeedtestIntegration tests the full speed test against real Cloudflare servers.
// Run with: go test -v -run TestSpeedtestIntegration -timeout 120s
func TestSpeedtestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cfg := config.ExporterConfig{
		BaseURL:         "https://speed.cloudflare.com",
		Concurrency:     4,
		ProbeTimeoutMs:  10 * time.Second,
		ProbeIntervalMs: 100 * time.Millisecond,
		SkipDiagnostics: true,
	}

	st := NewCloudflareSpeedtest(cfg)
	result, err := st.RunDirectTest()
	if err != nil {
		t.Fatalf("RunDirectTest failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Server == nil || *result.Server == "" {
		t.Error("expected server to be set")
	}
	if result.Colo == nil || *result.Colo == "" {
		t.Error("expected colo to be set")
	}
	if result.Download.Mbps <= 0 {
		t.Errorf("expected positive download speed, got %.2f Mbps", result.Download.Mbps)
	}
	if result.Upload.Mbps <= 0 {
		t.Errorf("expected positive upload speed, got %.2f Mbps", result.Upload.Mbps)
	}
	if result.IdleLatency.MedianMs == nil || *result.IdleLatency.MedianMs < 0 {
		t.Error("expected non-negative idle latency median")
	}

	t.Logf("Server: %s, Colo: %s", *result.Server, *result.Colo)
	t.Logf("Download: %.2f Mbps (%d bytes in %v)", result.Download.Mbps, result.Download.Bytes, time.Duration(result.Download.DurationMs)*time.Millisecond)
	t.Logf("Upload: %.2f Mbps (%d bytes in %v)", result.Upload.Mbps, result.Upload.Bytes, time.Duration(result.Upload.DurationMs)*time.Millisecond)
	t.Logf("Idle Latency: median=%.2fms, loss=%.2f%%", *result.IdleLatency.MedianMs, result.IdleLatency.Loss*100)
}
