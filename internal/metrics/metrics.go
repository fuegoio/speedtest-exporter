package metrics

import (
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/alexis/speedtest-exporter/internal/model"
)

// Metrics holds all Prometheus metrics
var (
	// Download metrics
	DownloadMbps = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_download_mbps",
			Help: "Current download speed in Mbps",
		},
		[]string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version"},
	)

	DownloadBytesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_download_bytes_total",
			Help: "Total bytes downloaded in the last test",
		},
		[]string{"server", "colo", "asn", "as_org"},
	)

	DownloadDurationMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_download_duration_ms",
			Help: "Duration of download test in milliseconds",
		},
		[]string{"server", "colo"},
	)

	// Upload metrics
	UploadMbps = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_upload_mbps",
			Help: "Current upload speed in Mbps",
		},
		[]string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version"},
	)

	UploadBytesTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_upload_bytes_total",
			Help: "Total bytes uploaded in the last test",
		},
		[]string{"server", "colo", "asn", "as_org"},
	)

	UploadDurationMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_upload_duration_ms",
			Help: "Duration of upload test in milliseconds",
		},
		[]string{"server", "colo"},
	)

	// Idle latency metrics
	IdleLatencyMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_idle_latency_ms",
			Help: "Idle latency in milliseconds",
		},
		[]string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "type"},
	)

	IdleLatencyMinMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_idle_latency_min_ms",
			Help: "Minimum idle latency in milliseconds",
		},
		[]string{"server", "colo"},
	)

	IdleLatencyMeanMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_idle_latency_mean_ms",
			Help: "Mean idle latency in milliseconds",
		},
		[]string{"server", "colo"},
	)

	IdleLatencyMedianMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_idle_latency_median_ms",
			Help: "Median idle latency in milliseconds",
		},
		[]string{"server", "colo"},
	)

	IdleLatencyP25Ms = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_idle_latency_p25_ms",
			Help: "25th percentile idle latency in milliseconds",
		},
		[]string{"server", "colo"},
	)

	IdleLatencyP75Ms = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_idle_latency_p75_ms",
			Help: "75th percentile idle latency in milliseconds",
		},
		[]string{"server", "colo"},
	)

	IdleLatencyMaxMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_idle_latency_max_ms",
			Help: "Maximum idle latency in milliseconds",
		},
		[]string{"server", "colo"},
	)

	IdleLatencyJitterMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_idle_latency_jitter_ms",
			Help: "Idle latency jitter in milliseconds",
		},
		[]string{"server", "colo"},
	)

	// Loaded latency (download) metrics
	LoadedLatencyDownloadMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_loaded_latency_download_ms",
			Help: "Loaded latency during download in milliseconds",
		},
		[]string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "type"},
	)

	LoadedLatencyDownloadMinMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_loaded_latency_download_min_ms",
			Help: "Minimum loaded latency during download in milliseconds",
		},
		[]string{"server", "colo"},
	)

	LoadedLatencyDownloadMeanMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_loaded_latency_download_mean_ms",
			Help: "Mean loaded latency during download in milliseconds",
		},
		[]string{"server", "colo"},
	)

	LoadedLatencyDownloadMedianMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_loaded_latency_download_median_ms",
			Help: "Median loaded latency during download in milliseconds",
		},
		[]string{"server", "colo"},
	)

	LoadedLatencyDownloadMaxMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_loaded_latency_download_max_ms",
			Help: "Maximum loaded latency during download in milliseconds",
		},
		[]string{"server", "colo"},
	)

	// Loaded latency (upload) metrics
	LoadedLatencyUploadMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_loaded_latency_upload_ms",
			Help: "Loaded latency during upload in milliseconds",
		},
		[]string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "type"},
	)

	LoadedLatencyUploadMinMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_loaded_latency_upload_min_ms",
			Help: "Minimum loaded latency during upload in milliseconds",
		},
		[]string{"server", "colo"},
	)

	LoadedLatencyUploadMeanMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_loaded_latency_upload_mean_ms",
			Help: "Mean loaded latency during upload in milliseconds",
		},
		[]string{"server", "colo"},
	)

	LoadedLatencyUploadMedianMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_loaded_latency_upload_median_ms",
			Help: "Median loaded latency during upload in milliseconds",
		},
		[]string{"server", "colo"},
	)

	LoadedLatencyUploadMaxMs = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_loaded_latency_upload_max_ms",
			Help: "Maximum loaded latency during upload in milliseconds",
		},
		[]string{"server", "colo"},
	)

	// Packet loss metrics
	IdleLatencyLossPercent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_idle_latency_loss_percent",
			Help: "Packet loss percentage during idle latency test",
		},
		[]string{"server", "colo"},
	)

	LoadedLatencyDownloadLossPercent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_loaded_latency_download_loss_percent",
			Help: "Packet loss percentage during download test",
		},
		[]string{"server", "colo"},
	)

	LoadedLatencyUploadLossPercent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "speedtest_loaded_latency_upload_loss_percent",
			Help: "Packet loss percentage during upload test",
		},
		[]string{"server", "colo"},
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
		DownloadMbps, DownloadBytesTotal, DownloadDurationMs,
		UploadMbps, UploadBytesTotal, UploadDurationMs,
		IdleLatencyMs, IdleLatencyMinMs, IdleLatencyMeanMs, IdleLatencyMedianMs,
		IdleLatencyP25Ms, IdleLatencyP75Ms, IdleLatencyMaxMs, IdleLatencyJitterMs,
		LoadedLatencyDownloadMs, LoadedLatencyDownloadMinMs, LoadedLatencyDownloadMeanMs,
		LoadedLatencyDownloadMedianMs, LoadedLatencyDownloadMaxMs,
		LoadedLatencyUploadMs, LoadedLatencyUploadMinMs, LoadedLatencyUploadMeanMs,
		LoadedLatencyUploadMedianMs, LoadedLatencyUploadMaxMs,
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
		"server":    derefString(result.Server, "unknown"),
		"colo":     derefString(result.Colo, "unknown"),
		"asn":      derefString(result.Asn, "unknown"),
		"as_org":   derefString(result.AsOrg, "unknown"),
		"interface": derefString(result.InterfaceName, "unknown"),
		"network":  derefString(result.NetworkName, "unknown"),
		"ip_version": getIPVersion(result),
	}

	// Increment test run counter
	TestRunsTotal.WithLabelValues("success").Inc()

	// Set timestamp
	TestTimestamp.Set(float64(time.Now().Unix()))

	// Calculate total test duration
	totalDuration := result.Download.DurationMs + result.Upload.DurationMs
	TestDurationTotalMs.Set(float64(totalDuration))

	// Download metrics
	DownloadMbps.With(labels).Set(result.Download.Mbps)
	DownloadBytesTotal.WithLabelValues(
		labels["server"], labels["colo"], labels["asn"], labels["as_org"],
	).Set(float64(result.Download.Bytes))
	DownloadDurationMs.WithLabelValues(
		labels["server"], labels["colo"],
	).Set(float64(result.Download.DurationMs))

	// Upload metrics
	UploadMbps.With(labels).Set(result.Upload.Mbps)
	UploadBytesTotal.WithLabelValues(
		labels["server"], labels["colo"], labels["asn"], labels["as_org"],
	).Set(float64(result.Upload.Bytes))
	UploadDurationMs.WithLabelValues(
		labels["server"], labels["colo"],
	).Set(float64(result.Upload.DurationMs))

	// Idle latency metrics
	IdleLatencyMs.With(labels).Set(derefFloat64(result.IdleLatency.MedianMs, 0))
	IdleLatencyMinMs.WithLabelValues(
		labels["server"], labels["colo"],
	).Set(derefFloat64(result.IdleLatency.MinMs, 0))
	IdleLatencyMeanMs.WithLabelValues(
		labels["server"], labels["colo"],
	).Set(derefFloat64(result.IdleLatency.MeanMs, 0))
	IdleLatencyMedianMs.WithLabelValues(
		labels["server"], labels["colo"],
	).Set(derefFloat64(result.IdleLatency.MedianMs, 0))
	IdleLatencyP25Ms.WithLabelValues(
		labels["server"], labels["colo"],
	).Set(derefFloat64(result.IdleLatency.P25Ms, 0))
	IdleLatencyP75Ms.WithLabelValues(
		labels["server"], labels["colo"],
	).Set(derefFloat64(result.IdleLatency.P75Ms, 0))
	IdleLatencyMaxMs.WithLabelValues(
		labels["server"], labels["colo"],
	).Set(derefFloat64(result.IdleLatency.MaxMs, 0))
	IdleLatencyJitterMs.WithLabelValues(
		labels["server"], labels["colo"],
	).Set(derefFloat64(result.IdleLatency.JitterMs, 0))
	IdleLatencyLossPercent.WithLabelValues(
		labels["server"], labels["colo"],
	).Set(result.IdleLatency.Loss * 100)

	// Loaded latency (download) metrics
	loadedLabels := labels
	loadedLabels["type"] = "current"
	LoadedLatencyDownloadMs.With(loadedLabels).Set(derefFloat64(result.LoadedLatencyDownload.MedianMs, 0))
	LoadedLatencyDownloadMinMs.WithLabelValues(
		labels["server"], labels["colo"],
	).Set(derefFloat64(result.LoadedLatencyDownload.MinMs, 0))
	LoadedLatencyDownloadMeanMs.WithLabelValues(
		labels["server"], labels["colo"],
	).Set(derefFloat64(result.LoadedLatencyDownload.MeanMs, 0))
	LoadedLatencyDownloadMedianMs.WithLabelValues(
		labels["server"], labels["colo"],
	).Set(derefFloat64(result.LoadedLatencyDownload.MedianMs, 0))
	LoadedLatencyDownloadMaxMs.WithLabelValues(
		labels["server"], labels["colo"],
	).Set(derefFloat64(result.LoadedLatencyDownload.MaxMs, 0))
	LoadedLatencyDownloadLossPercent.WithLabelValues(
		labels["server"], labels["colo"],
	).Set(result.LoadedLatencyDownload.Loss * 100)

	// Loaded latency (upload) metrics
	LoadedLatencyUploadMs.With(loadedLabels).Set(derefFloat64(result.LoadedLatencyUpload.MedianMs, 0))
	LoadedLatencyUploadMinMs.WithLabelValues(
		labels["server"], labels["colo"],
	).Set(derefFloat64(result.LoadedLatencyUpload.MinMs, 0))
	LoadedLatencyUploadMeanMs.WithLabelValues(
		labels["server"], labels["colo"],
	).Set(derefFloat64(result.LoadedLatencyUpload.MeanMs, 0))
	LoadedLatencyUploadMedianMs.WithLabelValues(
		labels["server"], labels["colo"],
	).Set(derefFloat64(result.LoadedLatencyUpload.MedianMs, 0))
	LoadedLatencyUploadMaxMs.WithLabelValues(
		labels["server"], labels["colo"],
	).Set(derefFloat64(result.LoadedLatencyUpload.MaxMs, 0))
	LoadedLatencyUploadLossPercent.WithLabelValues(
		labels["server"], labels["colo"],
	).Set(result.LoadedLatencyUpload.Loss * 100)

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
