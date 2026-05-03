package metrics

import (
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/fuegoio/speedtest-exporter/internal/model"
)

// Histogram buckets for latency metrics (in milliseconds)
var latencyBuckets = []float64{
	0.1, 0.5, 1, 2.5, 5, 10, 25, 50, 75, 100, 150, 200, 250, 300, 400, 500, 750, 1000, 1500, 2000, 3000, 5000,
}

// Histogram buckets for throughput metrics (in Mbps)
var throughputBuckets = []float64{
	0.01, 0.1, 1, 5, 10, 25, 50, 75, 100, 150, 200, 250, 300, 400, 500, 750, 1000, 1500, 2000,
}

// Histogram buckets for duration metrics (in milliseconds)
var durationBuckets = []float64{
	1, 5, 10, 50, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000, 15000, 20000, 30000,
}

// Metrics holds all Prometheus metrics
var (
	// Download histogram (replaces DownloadMbps gauge)
	DownloadMbps = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "speedtest_download_mbps",
			Help:    "Download speed in Mbps (histogram)",
			Buckets: throughputBuckets,
		},
		[]string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "size"},
	)

	// Download duration as histogram
	DownloadDurationMs = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "speedtest_download_duration_ms",
			Help:    "Duration of download test in milliseconds",
			Buckets: durationBuckets,
		},
		[]string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "size"},
	)

	// Upload histogram (replaces UploadMbps gauge)
	UploadMbps = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "speedtest_upload_mbps",
			Help:    "Upload speed in Mbps (histogram)",
			Buckets: throughputBuckets,
		},
		[]string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "size"},
	)

	// Upload duration as histogram
	UploadDurationMs = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "speedtest_upload_duration_ms",
			Help:    "Duration of upload test in milliseconds",
			Buckets: durationBuckets,
		},
		[]string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "size"},
	)

	// Idle latency histogram
	IdleLatencyMs = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "speedtest_idle_latency_ms",
			Help:    "Idle latency in milliseconds",
			Buckets: latencyBuckets,
		},
		[]string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "type"},
	)

	IdleLatencyJitterMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_idle_latency_jitter_ms",
			Help: "Idle latency jitter in milliseconds",
		},
		[]string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version"},
	)

	// Loaded latency (download) histogram
	LoadedLatencyDownloadMs = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "speedtest_loaded_latency_download_ms",
			Help:    "Loaded latency during download in milliseconds",
			Buckets: latencyBuckets,
		},
		[]string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "type"},
	)

	// Loaded latency (upload) histogram
	LoadedLatencyUploadMs = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "speedtest_loaded_latency_upload_ms",
			Help:    "Loaded latency during upload in milliseconds",
			Buckets: latencyBuckets,
		},
		[]string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "type"},
	)

	// Packet loss metrics
	IdleLatencyLossPercent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_idle_latency_loss_percent",
			Help: "Packet loss percentage during idle latency test",
		},
		[]string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version"},
	)

	LoadedLatencyDownloadLossPercent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_loaded_latency_download_loss_percent",
			Help: "Packet loss percentage during download test",
		},
		[]string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version"},
	)

	LoadedLatencyUploadLossPercent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_loaded_latency_upload_loss_percent",
			Help: "Packet loss percentage during upload test",
		},
		[]string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version"},
	)

	// DNS metrics
	DnsResolutionTimeMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_dns_resolution_time_ms",
			Help: "DNS resolution time in milliseconds",
		},
		[]string{"hostname", "dns_server"},
	)

	DnsIpv4Count = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_dns_ipv4_count",
			Help: "Number of IPv4 addresses resolved",
		},
		[]string{"hostname"},
	)

	DnsIpv6Count = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_dns_ipv6_count",
			Help: "Number of IPv6 addresses resolved",
		},
		[]string{"hostname"},
	)

	// TLS metrics
	TlsHandshakeTimeMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_tls_handshake_time_ms",
			Help: "TLS handshake time in milliseconds",
		},
		[]string{"protocol", "cipher_suite"},
	)

	// Network information metrics
	LocalIpv4 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_local_ipv4",
			Help: "Local IPv4 address (1 if present, 0 otherwise)",
		},
		[]string{"address"},
	)

	LocalIpv6 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_local_ipv6",
			Help: "Local IPv6 address (1 if present, 0 otherwise)",
		},
		[]string{"address"},
	)

	ExternalIpv4 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_external_ipv4",
			Help: "External IPv4 address (1 if present, 0 otherwise)",
		},
		[]string{"address"},
	)

	ExternalIpv6 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_external_ipv6",
			Help: "External IPv6 address (1 if present, 0 otherwise)",
		},
		[]string{"address"},
	)

	// Test metadata
	TestTimestamp = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "speedtest_test_timestamp",
			Help: "Timestamp of the last test in Unix seconds",
		},
	)

	TestDurationTotalMs = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "speedtest_test_duration_total_ms",
			Help: "Total duration of the test in milliseconds",
		},
	)

	// Error tracking
	TestErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "speedtest_test_errors_total",
			Help: "Total number of test errors",
		},
		[]string{"error_type"},
	)

	TestRunsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "speedtest_test_runs_total",
			Help: "Total number of test runs",
		},
		[]string{"status"},
	)
)

// RegisterAll registers all metrics with the given registry
func RegisterAll(reg prometheus.Registerer) error {
	collectors := []prometheus.Collector{
		DownloadMbps, DownloadDurationMs,
		UploadMbps, UploadDurationMs,
		IdleLatencyMs, IdleLatencyJitterMs,
		LoadedLatencyDownloadMs,
		LoadedLatencyUploadMs,
		IdleLatencyLossPercent, LoadedLatencyDownloadLossPercent, LoadedLatencyUploadLossPercent,
		DnsResolutionTimeMs, DnsIpv4Count, DnsIpv6Count,
		TlsHandshakeTimeMs,
		LocalIpv4, LocalIpv6, ExternalIpv4, ExternalIpv6,
		TestTimestamp, TestDurationTotalMs,
		TestErrorsTotal, TestRunsTotal,
	}

	for _, c := range collectors {
		if err := reg.Register(c); err != nil {
			return err
		}
	}
	return nil
}

// Handler returns the HTTP handler for the metrics endpoint
func Handler() http.Handler {
	return promhttp.Handler()
}

// UpdateMetrics updates all Prometheus metrics from a RunResult
func UpdateMetrics(result *model.RunResult) {
	labels := map[string]string{
		"server":     derefString(result.Server, "unknown"),
		"colo":       derefString(result.Colo, "unknown"),
		"asn":        derefString(result.Asn, "unknown"),
		"as_org":     derefString(result.AsOrg, "unknown"),
		"interface":  derefString(result.InterfaceName, "unknown"),
		"network":    derefString(result.NetworkName, "unknown"),
		"ip_version": getIPVersion(result),
	}

	// Increment test run counter
	TestRunsTotal.WithLabelValues("success").Inc()

	// Set timestamp
	TestTimestamp.Set(float64(time.Now().Unix()))

	// Calculate total test duration
	totalDuration := result.Download.DurationMs + result.Upload.DurationMs
	TestDurationTotalMs.Set(float64(totalDuration))

	// Download metrics (histogram)
	DownloadMbps.WithLabelValues(labels["server"], labels["colo"], labels["asn"], labels["as_org"], labels["interface"], labels["network"], labels["ip_version"], "total").Observe(result.Download.Mbps)
	DownloadDurationMs.WithLabelValues(labels["server"], labels["colo"], labels["asn"], labels["as_org"], labels["interface"], labels["network"], labels["ip_version"], "total").Observe(float64(result.Download.DurationMs))

	// Upload metrics (histogram)
	UploadMbps.WithLabelValues(labels["server"], labels["colo"], labels["asn"], labels["as_org"], labels["interface"], labels["network"], labels["ip_version"], "total").Observe(result.Upload.Mbps)
	UploadDurationMs.WithLabelValues(labels["server"], labels["colo"], labels["asn"], labels["as_org"], labels["interface"], labels["network"], labels["ip_version"], "total").Observe(float64(result.Upload.DurationMs))

	// Idle latency metrics (histogram)
	idleLabels := labels
	idleLabels["type"] = "idle"
	IdleLatencyMs.WithLabelValues(labels["server"], labels["colo"], labels["asn"], labels["as_org"], labels["interface"], labels["network"], labels["ip_version"], "idle").Observe(derefFloat64(result.IdleLatency.MedianMs, 0))
	IdleLatencyJitterMs.WithLabelValues(labels["server"], labels["colo"], labels["asn"], labels["as_org"], labels["interface"], labels["network"], labels["ip_version"]).Set(derefFloat64(result.IdleLatency.JitterMs, 0))
	IdleLatencyLossPercent.WithLabelValues(labels["server"], labels["colo"], labels["asn"], labels["as_org"], labels["interface"], labels["network"], labels["ip_version"]).Set(result.IdleLatency.Loss * 100)

	// Loaded latency (download) metrics (histogram)
	loadedLabels := labels
	loadedLabels["type"] = "download"
	LoadedLatencyDownloadMs.WithLabelValues(labels["server"], labels["colo"], labels["asn"], labels["as_org"], labels["interface"], labels["network"], labels["ip_version"], "download").Observe(derefFloat64(result.LoadedLatencyDownload.MedianMs, 0))
	LoadedLatencyDownloadLossPercent.WithLabelValues(labels["server"], labels["colo"], labels["asn"], labels["as_org"], labels["interface"], labels["network"], labels["ip_version"]).Set(result.LoadedLatencyDownload.Loss * 100)

	// Loaded latency (upload) metrics (histogram)
	loadedLabelsUpload := labels
	loadedLabelsUpload["type"] = "upload"
	LoadedLatencyUploadMs.WithLabelValues(labels["server"], labels["colo"], labels["asn"], labels["as_org"], labels["interface"], labels["network"], labels["ip_version"], "upload").Observe(derefFloat64(result.LoadedLatencyUpload.MedianMs, 0))
	LoadedLatencyUploadLossPercent.WithLabelValues(labels["server"], labels["colo"], labels["asn"], labels["as_org"], labels["interface"], labels["network"], labels["ip_version"]).Set(result.LoadedLatencyUpload.Loss * 100)

	// DNS metrics
	if result.Dns != nil {
		DnsResolutionTimeMs.WithLabelValues(
			result.Dns.Hostname,
			strings.Join(result.Dns.DnsServers, ","),
		).Set(result.Dns.ResolutionTimeMs)
		DnsIpv4Count.WithLabelValues(result.Dns.Hostname).Set(float64(result.Dns.Ipv4Count))
		DnsIpv6Count.WithLabelValues(result.Dns.Hostname).Set(float64(result.Dns.Ipv6Count))
	}

	// TLS metrics
	if result.Tls != nil {
		TlsHandshakeTimeMs.WithLabelValues(
			derefString(result.Tls.ProtocolVersion, "unknown"),
			derefString(result.Tls.CipherSuite, "unknown"),
		).Set(result.Tls.HandshakeTimeMs)
	}

	// Network information
	if result.LocalIpv4 != nil {
		LocalIpv4.WithLabelValues(*result.LocalIpv4).Set(1)
	}
	if result.LocalIpv6 != nil {
		LocalIpv6.WithLabelValues(*result.LocalIpv6).Set(1)
	}
	if result.ExternalIpv4 != nil {
		ExternalIpv4.WithLabelValues(*result.ExternalIpv4).Set(1)
	}
	if result.ExternalIpv6 != nil {
		ExternalIpv6.WithLabelValues(*result.ExternalIpv6).Set(1)
	}
}

// derefString returns the value or default if nil
func derefString(s *string, defaultValue string) string {
	if s != nil {
		return *s
	}
	return defaultValue
}

// derefFloat64 returns the value or default if nil
func derefFloat64(f *float64, defaultValue float64) float64 {
	if f != nil {
		return *f
	}
	return defaultValue
}

// getIPVersion determines the IP version based on available addresses
func getIPVersion(result *model.RunResult) string {
	if result.LocalIpv4 != nil && result.LocalIpv6 == nil {
		return "ipv4"
	} else if result.LocalIpv4 == nil && result.LocalIpv6 != nil {
		return "ipv6"
	}
	return "both"
}

// IncrementError increments the error counter
func IncrementError(errorType string) {
	TestErrorsTotal.WithLabelValues(errorType).Inc()
}

// IncrementRun increments the run counter with status
func IncrementRun(status string) {
	TestRunsTotal.WithLabelValues(status).Inc()
}
