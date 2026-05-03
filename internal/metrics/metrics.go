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

// Histogram buckets for packet loss (in percent)
var lossBuckets = []float64{
	0.1, 0.5, 1, 2, 5, 10, 15, 20, 25, 30, 40, 50, 75, 100,
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
		[]string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "country", "city", "region", "postal_code", "latitude", "longitude", "size"},
	)

	// Upload histogram (replaces UploadMbps gauge)
	UploadMbps = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "speedtest_upload_mbps",
			Help:    "Upload speed in Mbps (histogram)",
			Buckets: throughputBuckets,
		},
		[]string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "country", "city", "region", "postal_code", "latitude", "longitude", "size"},
	)

	// Latency histogram (idle, download, upload)
	LatencyMs = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "speedtest_latency_ms",
			Help:    "Latency in milliseconds",
			Buckets: latencyBuckets,
		},
		[]string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "country", "city", "region", "postal_code", "latitude", "longitude", "during"},
	)

	LatencyJitterMs = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "speedtest_latency_jitter_ms",
			Help:    "Latency jitter in milliseconds (histogram)",
			Buckets: latencyBuckets,
		},
		[]string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "country", "city", "region", "postal_code", "latitude", "longitude", "during"},
	)

	// Packet loss metrics
	LatencyLossPercent = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "speedtest_latency_loss_percent",
			Help:    "Packet loss percentage (histogram)",
			Buckets: lossBuckets,
		},
		[]string{"server", "colo", "asn", "as_org", "interface", "network", "ip_version", "country", "city", "region", "postal_code", "latitude", "longitude", "during"},
	)

	// DNS metrics
	DnsResolutionTimeMs = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "speedtest_dns_resolution_time_ms",
			Help:    "DNS resolution time in milliseconds (histogram over 10 runs)",
			Buckets: latencyBuckets,
		},
		[]string{"hostname", "dns_server"},
	)

	// TLS metrics
	TlsHandshakeTimeMs = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "speedtest_tls_handshake_time_ms",
			Help:    "TLS handshake time in milliseconds (histogram over 10 runs)",
			Buckets: latencyBuckets,
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
		DownloadMbps,
		UploadMbps,
		LatencyMs, LatencyJitterMs, LatencyLossPercent,
		DnsResolutionTimeMs,
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

// Handler returns the HTTP handler for the metrics endpoint using the given registry
func Handler(reg prometheus.Gatherer) http.Handler {
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
}

// UpdateMetrics updates all Prometheus metrics from a RunResult
func UpdateMetrics(result *model.RunResult) {
	labels := map[string]string{
		"server":      derefString(result.Server, "unknown"),
		"colo":        derefString(result.Colo, "unknown"),
		"asn":         derefString(result.Asn, "unknown"),
		"as_org":      derefString(result.AsOrg, "unknown"),
		"interface":   derefString(result.InterfaceName, "unknown"),
		"network":     derefString(result.NetworkName, "unknown"),
		"ip_version":  getIPVersion(result),
		"country":     derefString(result.Country, "unknown"),
		"city":        derefString(result.City, "unknown"),
		"region":      derefString(result.Region, "unknown"),
		"postal_code": derefString(result.PostalCode, "unknown"),
		"latitude":    derefString(result.Latitude, "unknown"),
		"longitude":   derefString(result.Longitude, "unknown"),
	}

	// Increment test run counter
	TestRunsTotal.WithLabelValues("success").Inc()

	// Set timestamp
	TestTimestamp.Set(float64(time.Now().Unix()))

	// Calculate total test duration
	totalDuration := result.Download.DurationMs + result.Upload.DurationMs
	TestDurationTotalMs.Set(float64(totalDuration))

	geo := []string{labels["country"], labels["city"], labels["region"], labels["postal_code"], labels["latitude"], labels["longitude"]}
	base := []string{labels["server"], labels["colo"], labels["asn"], labels["as_org"], labels["interface"], labels["network"], labels["ip_version"]}
	baseGeo := append(base, geo...)

	// Download metrics (histogram)
	DownloadMbps.WithLabelValues(append(baseGeo, "max")...).Observe(result.Download.Mbps)
	// Upload metrics (histogram)
	UploadMbps.WithLabelValues(append(baseGeo, "max")...).Observe(result.Upload.Mbps)

	// Latency metrics
	LatencyMs.WithLabelValues(append(baseGeo, "idle")...).Observe(derefFloat64(result.IdleLatency.MedianMs, 0))
	LatencyJitterMs.WithLabelValues(append(baseGeo, "idle")...).Observe(derefFloat64(result.IdleLatency.JitterMs, 0))
	LatencyJitterMs.WithLabelValues(append(baseGeo, "download")...).Observe(derefFloat64(result.LoadedLatencyDownload.JitterMs, 0))
	LatencyJitterMs.WithLabelValues(append(baseGeo, "upload")...).Observe(derefFloat64(result.LoadedLatencyUpload.JitterMs, 0))
	LatencyLossPercent.WithLabelValues(append(baseGeo, "idle")...).Observe(result.IdleLatency.Loss * 100)

	LatencyMs.WithLabelValues(append(baseGeo, "download")...).Observe(derefFloat64(result.LoadedLatencyDownload.MedianMs, 0))
	LatencyLossPercent.WithLabelValues(append(baseGeo, "download")...).Observe(result.LoadedLatencyDownload.Loss * 100)

	LatencyMs.WithLabelValues(append(baseGeo, "upload")...).Observe(derefFloat64(result.LoadedLatencyUpload.MedianMs, 0))
	LatencyLossPercent.WithLabelValues(append(baseGeo, "upload")...).Observe(result.LoadedLatencyUpload.Loss * 100)

	// DNS metrics
	if result.Dns != nil {
		for _, ms := range result.Dns.ResolutionTimeSamples {
			DnsResolutionTimeMs.WithLabelValues(
				result.Dns.Hostname,
				strings.Join(result.Dns.DnsServers, ","),
			).Observe(ms)
		}
	}

	// TLS metrics
	if result.Tls != nil {
		tlsLabels := []string{
			derefString(result.Tls.ProtocolVersion, "unknown"),
			derefString(result.Tls.CipherSuite, "unknown"),
		}
		for _, ms := range result.Tls.HandshakeTimeSamples {
			TlsHandshakeTimeMs.WithLabelValues(tlsLabels...).Observe(ms)
		}
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
