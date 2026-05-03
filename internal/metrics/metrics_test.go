package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"github.com/fuegoio/speedtest-exporter/internal/model"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// ptr helpers
func strPtr(s string) *string  { return &s }
func f64Ptr(f float64) *float64 { return &f }

// freshMetrics replaces all package-level metric vars with new instances and
// returns a registry containing them. This isolates each test from others.
func freshMetrics(t *testing.T) *prometheus.Registry {
	t.Helper()

	DownloadMbps = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "speedtest_download_mbps", Help: ".", Buckets: throughputBuckets,
	}, []string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "country", "city", "region", "postal_code", "latitude", "longitude", "size"})

	UploadMbps = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "speedtest_upload_mbps", Help: ".", Buckets: throughputBuckets,
	}, []string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "country", "city", "region", "postal_code", "latitude", "longitude", "size"})

	IdleLatencyMs = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "speedtest_idle_latency_ms", Help: ".", Buckets: latencyBuckets,
	}, []string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "country", "city", "region", "postal_code", "latitude", "longitude", "type"})

	IdleLatencyJitterMs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "speedtest_idle_latency_jitter_ms", Help: ".",
	}, []string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "country", "city", "region", "postal_code", "latitude", "longitude"})

	LoadedLatencyDownloadMs = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "speedtest_loaded_latency_download_ms", Help: ".", Buckets: latencyBuckets,
	}, []string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "country", "city", "region", "postal_code", "latitude", "longitude", "type"})

	LoadedLatencyUploadMs = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "speedtest_loaded_latency_upload_ms", Help: ".", Buckets: latencyBuckets,
	}, []string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "country", "city", "region", "postal_code", "latitude", "longitude", "type"})

	IdleLatencyLossPercent = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "speedtest_idle_latency_loss_percent", Help: ".",
	}, []string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "country", "city", "region", "postal_code", "latitude", "longitude"})

	LoadedLatencyDownloadLossPercent = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "speedtest_loaded_latency_download_loss_percent", Help: ".",
	}, []string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "country", "city", "region", "postal_code", "latitude", "longitude"})

	LoadedLatencyUploadLossPercent = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "speedtest_loaded_latency_upload_loss_percent", Help: ".",
	}, []string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "country", "city", "region", "postal_code", "latitude", "longitude"})

	DnsResolutionTimeMs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "speedtest_dns_resolution_time_ms", Help: ".",
	}, []string{"hostname", "dns_server"})

	DnsIpv4Count = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "speedtest_dns_ipv4_count", Help: ".",
	}, []string{"hostname"})

	DnsIpv6Count = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "speedtest_dns_ipv6_count", Help: ".",
	}, []string{"hostname"})

	TlsHandshakeTimeMs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "speedtest_tls_handshake_time_ms", Help: ".",
	}, []string{"protocol", "cipher_suite"})

	LocalIpv4 = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "speedtest_local_ipv4", Help: ".",
	}, []string{"address"})

	LocalIpv6 = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "speedtest_local_ipv6", Help: ".",
	}, []string{"address"})

	ExternalIpv4 = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "speedtest_external_ipv4", Help: ".",
	}, []string{"address"})

	ExternalIpv6 = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "speedtest_external_ipv6", Help: ".",
	}, []string{"address"})

	TestTimestamp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "speedtest_test_timestamp", Help: ".",
	})

	TestDurationTotalMs = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "speedtest_test_duration_total_ms", Help: ".",
	})

	TestErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "speedtest_test_errors_total", Help: ".",
	}, []string{"error_type"})

	TestRunsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "speedtest_test_runs_total", Help: ".",
	}, []string{"status"})

	reg := prometheus.NewRegistry()
	if err := RegisterAll(reg); err != nil {
		t.Fatalf("RegisterAll failed: %v", err)
	}
	return reg
}

// gather returns the named metric family from the registry, or nil.
func gather(t *testing.T, reg *prometheus.Registry, name string) *dto.MetricFamily {
	t.Helper()
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}
	for _, mf := range mfs {
		if mf.GetName() == name {
			return mf
		}
	}
	return nil
}

// buildResult returns a fully-populated RunResult for testing.
func buildResult() *model.RunResult {
	proto := "TLSv1.3"
	cipher := "TLS_AES_128_GCM_SHA256"
	return &model.RunResult{
		Server:        strPtr("US"),
		Colo:          strPtr("SFO"),
		Asn:           strPtr("AS12345"),
		AsOrg:         strPtr("Test ISP"),
		InterfaceName: strPtr("eth0"),
		NetworkName:   strPtr("home"),
		Country:       strPtr("US"),
		City:          strPtr("San Francisco"),
		Region:        strPtr("California"),
		PostalCode:    strPtr("94105"),
		Latitude:      strPtr("37.7749"),
		Longitude:     strPtr("-122.4194"),
		LocalIpv4:     strPtr("192.168.1.1"),
		LocalIpv6:     nil,
		ExternalIpv4:  strPtr("1.2.3.4"),
		ExternalIpv6:  nil,
		Download: model.ThroughputSummary{
			Bytes:      10_000_000,
			DurationMs: 1000,
			Mbps:       80.0,
		},
		Upload: model.ThroughputSummary{
			Bytes:      5_000_000,
			DurationMs: 1000,
			Mbps:       40.0,
		},
		IdleLatency: model.LatencySummary{
			Sent:     10,
			Received: 10,
			Loss:     0,
			MinMs:    f64Ptr(5.0),
			MeanMs:   f64Ptr(10.0),
			MedianMs: f64Ptr(9.5),
			MaxMs:    f64Ptr(20.0),
			JitterMs: f64Ptr(2.0),
		},
		LoadedLatencyDownload: model.LatencySummary{
			Sent:     10,
			Received: 9,
			Loss:     0.1,
			MedianMs: f64Ptr(25.0),
			JitterMs: f64Ptr(5.0),
		},
		LoadedLatencyUpload: model.LatencySummary{
			Sent:     10,
			Received: 8,
			Loss:     0.2,
			MedianMs: f64Ptr(30.0),
			JitterMs: f64Ptr(7.0),
		},
		Dns: &model.DnsSummary{
			Hostname:         "speed.cloudflare.com",
			ResolutionTimeMs: 12.5,
			Ipv4Count:        2,
			Ipv6Count:        1,
			DnsServers:       []string{"8.8.8.8"},
		},
		Tls: &model.TlsSummary{
			HandshakeTimeMs: 45.0,
			ProtocolVersion: &proto,
			CipherSuite:     &cipher,
		},
	}
}

// ---------------------------------------------------------------------------
// RegisterAll
// ---------------------------------------------------------------------------

func TestRegisterAll(t *testing.T) {
	reg := prometheus.NewRegistry()
	freshMetrics(t) // reset globals
	if err := RegisterAll(reg); err != nil {
		t.Fatalf("RegisterAll failed: %v", err)
	}
	// Registering again to the same registry should fail (duplicate)
	if err := RegisterAll(reg); err == nil {
		t.Error("expected error on duplicate registration")
	}
}

// ---------------------------------------------------------------------------
// UpdateMetrics — throughput
// ---------------------------------------------------------------------------

func TestUpdateMetricsDownload(t *testing.T) {
	reg := freshMetrics(t)
	UpdateMetrics(buildResult())

	mf := gather(t, reg, "speedtest_download_mbps")
	if mf == nil {
		t.Fatal("speedtest_download_mbps not found")
	}
	h := mf.GetMetric()[0].GetHistogram()
	if h.GetSampleCount() != 1 {
		t.Errorf("sample count = %d, want 1", h.GetSampleCount())
	}
	if h.GetSampleSum() != 80.0 {
		t.Errorf("download sum = %v, want 80.0", h.GetSampleSum())
	}
}

func TestUpdateMetricsUpload(t *testing.T) {
	reg := freshMetrics(t)
	UpdateMetrics(buildResult())

	mf := gather(t, reg, "speedtest_upload_mbps")
	if mf == nil {
		t.Fatal("speedtest_upload_mbps not found")
	}
	h := mf.GetMetric()[0].GetHistogram()
	if h.GetSampleCount() != 1 {
		t.Errorf("sample count = %d, want 1", h.GetSampleCount())
	}
	if h.GetSampleSum() != 40.0 {
		t.Errorf("upload sum = %v, want 40.0", h.GetSampleSum())
	}
}

// ---------------------------------------------------------------------------
// UpdateMetrics — latency
// ---------------------------------------------------------------------------

func TestUpdateMetricsIdleLatency(t *testing.T) {
	reg := freshMetrics(t)
	UpdateMetrics(buildResult())

	mf := gather(t, reg, "speedtest_idle_latency_ms")
	if mf == nil {
		t.Fatal("speedtest_idle_latency_ms not found")
	}
	h := mf.GetMetric()[0].GetHistogram()
	if h.GetSampleSum() != 9.5 {
		t.Errorf("idle latency sum = %v, want 9.5 (median)", h.GetSampleSum())
	}

	mfJ := gather(t, reg, "speedtest_idle_latency_jitter_ms")
	if mfJ == nil {
		t.Fatal("speedtest_idle_latency_jitter_ms not found")
	}
	if mfJ.GetMetric()[0].GetGauge().GetValue() != 2.0 {
		t.Errorf("jitter = %v, want 2.0", mfJ.GetMetric()[0].GetGauge().GetValue())
	}

	mfL := gather(t, reg, "speedtest_idle_latency_loss_percent")
	if mfL == nil {
		t.Fatal("speedtest_idle_latency_loss_percent not found")
	}
	if mfL.GetMetric()[0].GetGauge().GetValue() != 0 {
		t.Errorf("loss = %v, want 0", mfL.GetMetric()[0].GetGauge().GetValue())
	}
}

func TestUpdateMetricsLoadedLatencyDownload(t *testing.T) {
	reg := freshMetrics(t)
	UpdateMetrics(buildResult())

	mf := gather(t, reg, "speedtest_loaded_latency_download_ms")
	if mf == nil {
		t.Fatal("speedtest_loaded_latency_download_ms not found")
	}
	h := mf.GetMetric()[0].GetHistogram()
	if h.GetSampleSum() != 25.0 {
		t.Errorf("loaded download latency sum = %v, want 25.0", h.GetSampleSum())
	}

	mfL := gather(t, reg, "speedtest_loaded_latency_download_loss_percent")
	if mfL == nil {
		t.Fatal("speedtest_loaded_latency_download_loss_percent not found")
	}
	if mfL.GetMetric()[0].GetGauge().GetValue() != 10.0 {
		t.Errorf("download loss = %v, want 10.0", mfL.GetMetric()[0].GetGauge().GetValue())
	}
}

func TestUpdateMetricsLoadedLatencyUpload(t *testing.T) {
	reg := freshMetrics(t)
	UpdateMetrics(buildResult())

	mf := gather(t, reg, "speedtest_loaded_latency_upload_ms")
	if mf == nil {
		t.Fatal("speedtest_loaded_latency_upload_ms not found")
	}
	h := mf.GetMetric()[0].GetHistogram()
	if h.GetSampleSum() != 30.0 {
		t.Errorf("loaded upload latency sum = %v, want 30.0", h.GetSampleSum())
	}

	mfL := gather(t, reg, "speedtest_loaded_latency_upload_loss_percent")
	if mfL == nil {
		t.Fatal("speedtest_loaded_latency_upload_loss_percent not found")
	}
	if mfL.GetMetric()[0].GetGauge().GetValue() != 20.0 {
		t.Errorf("upload loss = %v, want 20.0", mfL.GetMetric()[0].GetGauge().GetValue())
	}
}

// ---------------------------------------------------------------------------
// UpdateMetrics — DNS
// ---------------------------------------------------------------------------

func TestUpdateMetricsDns(t *testing.T) {
	reg := freshMetrics(t)
	UpdateMetrics(buildResult())

	mf := gather(t, reg, "speedtest_dns_resolution_time_ms")
	if mf == nil {
		t.Fatal("speedtest_dns_resolution_time_ms not found")
	}
	if mf.GetMetric()[0].GetGauge().GetValue() != 12.5 {
		t.Errorf("dns resolution time = %v, want 12.5", mf.GetMetric()[0].GetGauge().GetValue())
	}

	mfV4 := gather(t, reg, "speedtest_dns_ipv4_count")
	if mfV4 == nil {
		t.Fatal("speedtest_dns_ipv4_count not found")
	}
	if mfV4.GetMetric()[0].GetGauge().GetValue() != 2 {
		t.Errorf("ipv4 count = %v, want 2", mfV4.GetMetric()[0].GetGauge().GetValue())
	}

	mfV6 := gather(t, reg, "speedtest_dns_ipv6_count")
	if mfV6 == nil {
		t.Fatal("speedtest_dns_ipv6_count not found")
	}
	if mfV6.GetMetric()[0].GetGauge().GetValue() != 1 {
		t.Errorf("ipv6 count = %v, want 1", mfV6.GetMetric()[0].GetGauge().GetValue())
	}
}

func TestUpdateMetricsDnsNil(t *testing.T) {
	reg := freshMetrics(t)
	result := buildResult()
	result.Dns = nil
	UpdateMetrics(result)

	mf := gather(t, reg, "speedtest_dns_resolution_time_ms")
	if mf != nil && len(mf.GetMetric()) > 0 {
		t.Error("expected no DNS metric when Dns is nil")
	}
}

// ---------------------------------------------------------------------------
// UpdateMetrics — TLS
// ---------------------------------------------------------------------------

func TestUpdateMetricsTls(t *testing.T) {
	reg := freshMetrics(t)
	UpdateMetrics(buildResult())

	mf := gather(t, reg, "speedtest_tls_handshake_time_ms")
	if mf == nil {
		t.Fatal("speedtest_tls_handshake_time_ms not found")
	}
	if mf.GetMetric()[0].GetGauge().GetValue() != 45.0 {
		t.Errorf("tls handshake = %v, want 45.0", mf.GetMetric()[0].GetGauge().GetValue())
	}
}

func TestUpdateMetricsTlsNil(t *testing.T) {
	reg := freshMetrics(t)
	result := buildResult()
	result.Tls = nil
	UpdateMetrics(result)

	mf := gather(t, reg, "speedtest_tls_handshake_time_ms")
	if mf != nil && len(mf.GetMetric()) > 0 {
		t.Error("expected no TLS metric when Tls is nil")
	}
}

// ---------------------------------------------------------------------------
// UpdateMetrics — network addresses
// ---------------------------------------------------------------------------

func TestUpdateMetricsNetworkAddresses(t *testing.T) {
	reg := freshMetrics(t)
	UpdateMetrics(buildResult())

	mfV4 := gather(t, reg, "speedtest_local_ipv4")
	if mfV4 == nil {
		t.Fatal("speedtest_local_ipv4 not found")
	}
	if mfV4.GetMetric()[0].GetGauge().GetValue() != 1 {
		t.Errorf("local_ipv4 = %v, want 1", mfV4.GetMetric()[0].GetGauge().GetValue())
	}

	// LocalIpv6 is nil — metric should be absent
	mfV6 := gather(t, reg, "speedtest_local_ipv6")
	if mfV6 != nil && len(mfV6.GetMetric()) > 0 {
		t.Error("expected no local_ipv6 metric when LocalIpv6 is nil")
	}

	mfExtV4 := gather(t, reg, "speedtest_external_ipv4")
	if mfExtV4 == nil {
		t.Fatal("speedtest_external_ipv4 not found")
	}
	if mfExtV4.GetMetric()[0].GetGauge().GetValue() != 1 {
		t.Errorf("external_ipv4 = %v, want 1", mfExtV4.GetMetric()[0].GetGauge().GetValue())
	}
}

// ---------------------------------------------------------------------------
// UpdateMetrics — counters and gauges
// ---------------------------------------------------------------------------

func TestUpdateMetricsRunCounter(t *testing.T) {
	reg := freshMetrics(t)
	UpdateMetrics(buildResult())
	UpdateMetrics(buildResult()) // second call

	mf := gather(t, reg, "speedtest_test_runs_total")
	if mf == nil {
		t.Fatal("speedtest_test_runs_total not found")
	}
	if mf.GetMetric()[0].GetCounter().GetValue() != 2 {
		t.Errorf("runs_total = %v, want 2", mf.GetMetric()[0].GetCounter().GetValue())
	}
}

func TestUpdateMetricsTimestamp(t *testing.T) {
	reg := freshMetrics(t)
	UpdateMetrics(buildResult())

	mf := gather(t, reg, "speedtest_test_timestamp")
	if mf == nil {
		t.Fatal("speedtest_test_timestamp not found")
	}
	if mf.GetMetric()[0].GetGauge().GetValue() == 0 {
		t.Error("expected non-zero timestamp")
	}
}

func TestUpdateMetricsDuration(t *testing.T) {
	reg := freshMetrics(t)
	UpdateMetrics(buildResult())

	mf := gather(t, reg, "speedtest_test_duration_total_ms")
	if mf == nil {
		t.Fatal("speedtest_test_duration_total_ms not found")
	}
	// download 1000ms + upload 1000ms = 2000ms
	if mf.GetMetric()[0].GetGauge().GetValue() != 2000 {
		t.Errorf("duration = %v, want 2000", mf.GetMetric()[0].GetGauge().GetValue())
	}
}

// ---------------------------------------------------------------------------
// getIPVersion
// ---------------------------------------------------------------------------

func TestGetIPVersion(t *testing.T) {
	cases := []struct {
		name     string
		ipv4     *string
		ipv6     *string
		expected string
	}{
		{"both", strPtr("192.168.1.1"), strPtr("::1"), "both"},
		{"ipv4 only", strPtr("192.168.1.1"), nil, "ipv4"},
		{"ipv6 only", nil, strPtr("::1"), "ipv6"},
		{"neither", nil, nil, "both"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := &model.RunResult{LocalIpv4: tc.ipv4, LocalIpv6: tc.ipv6}
			got := getIPVersion(result)
			if got != tc.expected {
				t.Errorf("getIPVersion = %q, want %q", got, tc.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// IncrementError / IncrementRun
// ---------------------------------------------------------------------------

func TestIncrementError(t *testing.T) {
	reg := freshMetrics(t)
	IncrementError("network")
	IncrementError("network")
	IncrementError("timeout")

	mf := gather(t, reg, "speedtest_test_errors_total")
	if mf == nil {
		t.Fatal("speedtest_test_errors_total not found")
	}
	total := 0.0
	for _, m := range mf.GetMetric() {
		total += m.GetCounter().GetValue()
	}
	if total != 3 {
		t.Errorf("total errors = %v, want 3", total)
	}
}

func TestIncrementRun(t *testing.T) {
	reg := freshMetrics(t)
	IncrementRun("success")
	IncrementRun("failure")

	mf := gather(t, reg, "speedtest_test_runs_total")
	if mf == nil {
		t.Fatal("speedtest_test_runs_total not found")
	}
	total := 0.0
	for _, m := range mf.GetMetric() {
		total += m.GetCounter().GetValue()
	}
	if total != 2 {
		t.Errorf("total runs = %v, want 2", total)
	}
}

// ---------------------------------------------------------------------------
// derefString / derefFloat64
// ---------------------------------------------------------------------------

func TestDerefString(t *testing.T) {
	s := "hello"
	if derefString(&s, "default") != "hello" {
		t.Error("expected 'hello'")
	}
	if derefString(nil, "default") != "default" {
		t.Error("expected 'default'")
	}
}

func TestDerefFloat64(t *testing.T) {
	f := 3.14
	if derefFloat64(&f, 0) != 3.14 {
		t.Error("expected 3.14")
	}
	if derefFloat64(nil, 99.9) != 99.9 {
		t.Error("expected 99.9")
	}
}

// ---------------------------------------------------------------------------
// Label correctness
// ---------------------------------------------------------------------------

func TestUpdateMetricsLabels(t *testing.T) {
	reg := freshMetrics(t)
	UpdateMetrics(buildResult())

	mf := gather(t, reg, "speedtest_download_mbps")
	if mf == nil {
		t.Fatal("speedtest_download_mbps not found")
	}

	labelMap := map[string]string{}
	for _, lp := range mf.GetMetric()[0].GetLabel() {
		labelMap[lp.GetName()] = lp.GetValue()
	}

	expected := map[string]string{
		"server":      "US",
		"colo":        "SFO",
		"asn":         "AS12345",
		"as_org":      "Test ISP",
		"interface":   "eth0",
		"network":     "home",
		"ip_version":  "ipv4",
		"country":     "US",
		"city":        "San Francisco",
		"region":      "California",
		"postal_code": "94105",
		"latitude":    "37.7749",
		"longitude":   "-122.4194",
		"size":        "total",
	}
	for k, want := range expected {
		if got := labelMap[k]; got != want {
			t.Errorf("label %q = %q, want %q", k, got, want)
		}
	}
}
