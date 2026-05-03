import client from "prom-client";

// Create a registry for all metrics
export const register = new client.Registry();

// ============================================================================
// Gauge Metrics - Current values that can go up and down
// ============================================================================

// Download metrics
export const downloadMbps = new client.Gauge({
  name: "speedtest_download_mbps",
  help: "Current download speed in Mbps",
  labelNames: ["server", "colo", "asn", "as_org", "interface", "network", "ip_version"],
});

export const downloadBytesTotal = new client.Gauge({
  name: "speedtest_download_bytes_total",
  help: "Total bytes downloaded in the last test",
  labelNames: ["server", "colo", "asn", "as_org"],
});

export const downloadDurationMs = new client.Gauge({
  name: "speedtest_download_duration_ms",
  help: "Duration of download test in milliseconds",
  labelNames: ["server", "colo"],
});

// Upload metrics
export const uploadMbps = new client.Gauge({
  name: "speedtest_upload_mbps",
  help: "Current upload speed in Mbps",
  labelNames: ["server", "colo", "asn", "as_org", "interface", "network", "ip_version"],
});

export const uploadBytesTotal = new client.Gauge({
  name: "speedtest_upload_bytes_total",
  help: "Total bytes uploaded in the last test",
  labelNames: ["server", "colo", "asn", "as_org"],
});

export const uploadDurationMs = new client.Gauge({
  name: "speedtest_upload_duration_ms",
  help: "Duration of upload test in milliseconds",
  labelNames: ["server", "colo"],
});

// Latency metrics (Idle)
export const idleLatencyMs = new client.Gauge({
  name: "speedtest_idle_latency_ms",
  help: "Idle latency in milliseconds",
  labelNames: ["server", "colo", "asn", "as_org", "interface", "network", "ip_version", "type"],
});

export const idleLatencyMinMs = new client.Gauge({
  name: "speedtest_idle_latency_min_ms",
  help: "Minimum idle latency in milliseconds",
  labelNames: ["server", "colo"],
});

export const idleLatencyMeanMs = new client.Gauge({
  name: "speedtest_idle_latency_mean_ms",
  help: "Mean idle latency in milliseconds",
  labelNames: ["server", "colo"],
});

export const idleLatencyMedianMs = new client.Gauge({
  name: "speedtest_idle_latency_median_ms",
  help: "Median idle latency in milliseconds",
  labelNames: ["server", "colo"],
});

export const idleLatencyP25Ms = new client.Gauge({
  name: "speedtest_idle_latency_p25_ms",
  help: "25th percentile idle latency in milliseconds",
  labelNames: ["server", "colo"],
});

export const idleLatencyP75Ms = new client.Gauge({
  name: "speedtest_idle_latency_p75_ms",
  help: "75th percentile idle latency in milliseconds",
  labelNames: ["server", "colo"],
});

export const idleLatencyMaxMs = new client.Gauge({
  name: "speedtest_idle_latency_max_ms",
  help: "Maximum idle latency in milliseconds",
  labelNames: ["server", "colo"],
});

export const idleLatencyJitterMs = new client.Gauge({
  name: "speedtest_idle_latency_jitter_ms",
  help: "Idle latency jitter in milliseconds",
  labelNames: ["server", "colo"],
});

// Loaded latency (during download)
export const loadedLatencyDownloadMs = new client.Gauge({
  name: "speedtest_loaded_latency_download_ms",
  help: "Loaded latency during download in milliseconds",
  labelNames: ["server", "colo", "asn", "as_org", "interface", "network", "ip_version", "type"],
});

export const loadedLatencyDownloadMinMs = new client.Gauge({
  name: "speedtest_loaded_latency_download_min_ms",
  help: "Minimum loaded latency during download in milliseconds",
  labelNames: ["server", "colo"],
});

export const loadedLatencyDownloadMeanMs = new client.Gauge({
  name: "speedtest_loaded_latency_download_mean_ms",
  help: "Mean loaded latency during download in milliseconds",
  labelNames: ["server", "colo"],
});

export const loadedLatencyDownloadMedianMs = new client.Gauge({
  name: "speedtest_loaded_latency_download_median_ms",
  help: "Median loaded latency during download in milliseconds",
  labelNames: ["server", "colo"],
});

export const loadedLatencyDownloadMaxMs = new client.Gauge({
  name: "speedtest_loaded_latency_download_max_ms",
  help: "Maximum loaded latency during download in milliseconds",
  labelNames: ["server", "colo"],
});

// Loaded latency (during upload)
export const loadedLatencyUploadMs = new client.Gauge({
  name: "speedtest_loaded_latency_upload_ms",
  help: "Loaded latency during upload in milliseconds",
  labelNames: ["server", "colo", "asn", "as_org", "interface", "network", "ip_version", "type"],
});

export const loadedLatencyUploadMinMs = new client.Gauge({
  name: "speedtest_loaded_latency_upload_min_ms",
  help: "Minimum loaded latency during upload in milliseconds",
  labelNames: ["server", "colo"],
});

export const loadedLatencyUploadMeanMs = new client.Gauge({
  name: "speedtest_loaded_latency_upload_mean_ms",
  help: "Mean loaded latency during upload in milliseconds",
  labelNames: ["server", "colo"],
});

export const loadedLatencyUploadMedianMs = new client.Gauge({
  name: "speedtest_loaded_latency_upload_median_ms",
  help: "Median loaded latency during upload in milliseconds",
  labelNames: ["server", "colo"],
});

export const loadedLatencyUploadMaxMs = new client.Gauge({
  name: "speedtest_loaded_latency_upload_max_ms",
  help: "Maximum loaded latency during upload in milliseconds",
  labelNames: ["server", "colo"],
});

// Packet loss metrics
export const idleLatencyLossPercent = new client.Gauge({
  name: "speedtest_idle_latency_loss_percent",
  help: "Packet loss percentage during idle latency test",
  labelNames: ["server", "colo"],
});

export const loadedLatencyDownloadLossPercent = new client.Gauge({
  name: "speedtest_loaded_latency_download_loss_percent",
  help: "Packet loss percentage during download test",
  labelNames: ["server", "colo"],
});

export const loadedLatencyUploadLossPercent = new client.Gauge({
  name: "speedtest_loaded_latency_upload_loss_percent",
  help: "Packet loss percentage during upload test",
  labelNames: ["server", "colo"],
});

// Diagnostic metrics - DNS
export const dnsResolutionTimeMs = new client.Gauge({
  name: "speedtest_dns_resolution_time_ms",
  help: "DNS resolution time in milliseconds",
  labelNames: ["hostname", "dns_server"],
});

export const dnsIpv4Count = new client.Gauge({
  name: "speedtest_dns_ipv4_count",
  help: "Number of IPv4 addresses resolved",
  labelNames: ["hostname"],
});

export const dnsIpv6Count = new client.Gauge({
  name: "speedtest_dns_ipv6_count",
  help: "Number of IPv6 addresses resolved",
  labelNames: ["hostname"],
});

// Diagnostic metrics - TLS
export const tlsHandshakeTimeMs = new client.Gauge({
  name: "speedtest_tls_handshake_time_ms",
  help: "TLS handshake time in milliseconds",
  labelNames: ["protocol", "cipher_suite"],
});

// Traceroute metrics
export const tracerouteHopsCount = new client.Gauge({
  name: "speedtest_traceroute_hops_count",
  help: "Number of hops in traceroute",
  labelNames: ["destination"],
});

export const tracerouteCompleted = new client.Gauge({
  name: "speedtest_traceroute_completed",
  help: "Whether traceroute completed successfully (1 = yes, 0 = no)",
  labelNames: ["destination"],
});

// Network information metrics
export const localIpv4 = new client.Gauge({
  name: "speedtest_local_ipv4",
  help: "Local IPv4 address (1 if present, 0 otherwise)",
  labelNames: ["address"],
});

export const localIpv6 = new client.Gauge({
  name: "speedtest_local_ipv6",
  help: "Local IPv6 address (1 if present, 0 otherwise)",
  labelNames: ["address"],
});

export const externalIpv4 = new client.Gauge({
  name: "speedtest_external_ipv4",
  help: "External IPv4 address (1 if present, 0 otherwise)",
  labelNames: ["address"],
});

export const externalIpv6 = new client.Gauge({
  name: "speedtest_external_ipv6",
  help: "External IPv6 address (1 if present, 0 otherwise)",
  labelNames: ["address"],
});

// Test metadata
export const testTimestamp = new client.Gauge({
  name: "speedtest_test_timestamp",
  help: "Timestamp of the last test in Unix seconds",
});

export const testDurationTotalMs = new client.Gauge({
  name: "speedtest_test_duration_total_ms",
  help: "Total duration of the test in milliseconds",
});

// Error tracking
export const testErrorsTotal = new client.Counter({
  name: "speedtest_test_errors_total",
  help: "Total number of test errors",
  labelNames: ["error_type"],
});

export const testRunsTotal = new client.Counter({
  name: "speedtest_test_runs_total",
  help: "Total number of test runs",
  labelNames: ["status"],
});

// Register all metrics individually to avoid type issues
register.registerMetric(downloadMbps);
register.registerMetric(downloadBytesTotal);
register.registerMetric(downloadDurationMs);
register.registerMetric(uploadMbps);
register.registerMetric(uploadBytesTotal);
register.registerMetric(uploadDurationMs);
register.registerMetric(idleLatencyMs);
register.registerMetric(idleLatencyMinMs);
register.registerMetric(idleLatencyMeanMs);
register.registerMetric(idleLatencyMedianMs);
register.registerMetric(idleLatencyP25Ms);
register.registerMetric(idleLatencyP75Ms);
register.registerMetric(idleLatencyMaxMs);
register.registerMetric(idleLatencyJitterMs);
register.registerMetric(loadedLatencyDownloadMs);
register.registerMetric(loadedLatencyDownloadMinMs);
register.registerMetric(loadedLatencyDownloadMeanMs);
register.registerMetric(loadedLatencyDownloadMedianMs);
register.registerMetric(loadedLatencyDownloadMaxMs);
register.registerMetric(loadedLatencyUploadMs);
register.registerMetric(loadedLatencyUploadMinMs);
register.registerMetric(loadedLatencyUploadMeanMs);
register.registerMetric(loadedLatencyUploadMedianMs);
register.registerMetric(loadedLatencyUploadMaxMs);
register.registerMetric(idleLatencyLossPercent);
register.registerMetric(loadedLatencyDownloadLossPercent);
register.registerMetric(loadedLatencyUploadLossPercent);
register.registerMetric(dnsResolutionTimeMs);
register.registerMetric(dnsIpv4Count);
register.registerMetric(dnsIpv6Count);
register.registerMetric(tlsHandshakeTimeMs);
register.registerMetric(tracerouteHopsCount);
register.registerMetric(tracerouteCompleted);
register.registerMetric(localIpv4);
register.registerMetric(localIpv6);
register.registerMetric(externalIpv4);
register.registerMetric(externalIpv6);
register.registerMetric(testTimestamp);
register.registerMetric(testDurationTotalMs);
register.registerMetric(testErrorsTotal);
register.registerMetric(testRunsTotal);
