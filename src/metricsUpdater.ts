import type { RunResult } from "./model";
import {
  downloadMbps,
  downloadBytesTotal,
  downloadDurationMs,
  uploadMbps,
  uploadBytesTotal,
  uploadDurationMs,
  idleLatencyMs,
  idleLatencyMinMs,
  idleLatencyMeanMs,
  idleLatencyMedianMs,
  idleLatencyP25Ms,
  idleLatencyP75Ms,
  idleLatencyMaxMs,
  idleLatencyJitterMs,
  loadedLatencyDownloadMs,
  loadedLatencyDownloadMinMs,
  loadedLatencyDownloadMeanMs,
  loadedLatencyDownloadMedianMs,
  loadedLatencyDownloadMaxMs,
  loadedLatencyUploadMs,
  loadedLatencyUploadMinMs,
  loadedLatencyUploadMeanMs,
  loadedLatencyUploadMedianMs,
  loadedLatencyUploadMaxMs,
  idleLatencyLossPercent,
  loadedLatencyDownloadLossPercent,
  loadedLatencyUploadLossPercent,
  dnsResolutionTimeMs,
  dnsIpv4Count,
  dnsIpv6Count,
  tlsHandshakeTimeMs,
  tracerouteHopsCount,
  tracerouteCompleted,
  localIpv4,
  localIpv6,
  externalIpv4,
  externalIpv6,
  testTimestamp,
  testDurationTotalMs,
  testRunsTotal,
} from "./metrics";

// Update Prometheus metrics from RunResult
export function updateMetrics(result: RunResult) {
  const labels = {
    server: result.server || "unknown",
    colo: result.colo || "unknown",
    asn: result.asn || "unknown",
    as_org: result.as_org || "unknown",
    interface: result.interface_name || "unknown",
    network: result.network_name || "unknown",
    ip_version: "both",
  };

  // Increment test run counter
  testRunsTotal.inc({ status: "success" });

  // Set timestamp
  testTimestamp.set({}, Date.now() / 1000);

  // Calculate total test duration
  const totalDuration = result.download.duration_ms + result.upload.duration_ms;
  testDurationTotalMs.set({}, totalDuration);

  // Download metrics
  downloadMbps.set(labels, result.download.mbps);
  downloadBytesTotal.set(
    { server: labels.server, colo: labels.colo, asn: labels.asn, as_org: labels.as_org },
    result.download.bytes,
  );
  downloadDurationMs.set({ server: labels.server, colo: labels.colo }, result.download.duration_ms);

  // Upload metrics
  uploadMbps.set(labels, result.upload.mbps);
  uploadBytesTotal.set(
    { server: labels.server, colo: labels.colo, asn: labels.asn, as_org: labels.as_org },
    result.upload.bytes,
  );
  uploadDurationMs.set({ server: labels.server, colo: labels.colo }, result.upload.duration_ms);

  // Idle latency metrics
  idleLatencyMs.set({ ...labels, type: "current" }, result.idle_latency.median_ms || 0);
  idleLatencyMinMs.set(
    { server: labels.server, colo: labels.colo },
    result.idle_latency.min_ms || 0,
  );
  idleLatencyMeanMs.set(
    { server: labels.server, colo: labels.colo },
    result.idle_latency.mean_ms || 0,
  );
  idleLatencyMedianMs.set(
    { server: labels.server, colo: labels.colo },
    result.idle_latency.median_ms || 0,
  );
  idleLatencyP25Ms.set(
    { server: labels.server, colo: labels.colo },
    result.idle_latency.p25_ms || 0,
  );
  idleLatencyP75Ms.set(
    { server: labels.server, colo: labels.colo },
    result.idle_latency.p75_ms || 0,
  );
  idleLatencyMaxMs.set(
    { server: labels.server, colo: labels.colo },
    result.idle_latency.max_ms || 0,
  );
  idleLatencyJitterMs.set(
    { server: labels.server, colo: labels.colo },
    result.idle_latency.jitter_ms || 0,
  );
  idleLatencyLossPercent.set(
    { server: labels.server, colo: labels.colo },
    result.idle_latency.loss * 100,
  );

  // Loaded latency (download) metrics
  loadedLatencyDownloadMs.set(
    { ...labels, type: "current" },
    result.loaded_latency_download.median_ms || 0,
  );
  loadedLatencyDownloadMinMs.set(
    { server: labels.server, colo: labels.colo },
    result.loaded_latency_download.min_ms || 0,
  );
  loadedLatencyDownloadMeanMs.set(
    { server: labels.server, colo: labels.colo },
    result.loaded_latency_download.mean_ms || 0,
  );
  loadedLatencyDownloadMedianMs.set(
    { server: labels.server, colo: labels.colo },
    result.loaded_latency_download.median_ms || 0,
  );
  loadedLatencyDownloadMaxMs.set(
    { server: labels.server, colo: labels.colo },
    result.loaded_latency_download.max_ms || 0,
  );
  loadedLatencyDownloadLossPercent.set(
    { server: labels.server, colo: labels.colo },
    result.loaded_latency_download.loss * 100,
  );

  // Loaded latency (upload) metrics
  loadedLatencyUploadMs.set(
    { ...labels, type: "current" },
    result.loaded_latency_upload.median_ms || 0,
  );
  loadedLatencyUploadMinMs.set(
    { server: labels.server, colo: labels.colo },
    result.loaded_latency_upload.min_ms || 0,
  );
  loadedLatencyUploadMeanMs.set(
    { server: labels.server, colo: labels.colo },
    result.loaded_latency_upload.mean_ms || 0,
  );
  loadedLatencyUploadMedianMs.set(
    { server: labels.server, colo: labels.colo },
    result.loaded_latency_upload.median_ms || 0,
  );
  loadedLatencyUploadMaxMs.set(
    { server: labels.server, colo: labels.colo },
    result.loaded_latency_upload.max_ms || 0,
  );
  loadedLatencyUploadLossPercent.set(
    { server: labels.server, colo: labels.colo },
    result.loaded_latency_upload.loss * 100,
  );

  // DNS metrics
  if (result.dns) {
    dnsResolutionTimeMs.set(
      {
        hostname: result.dns.hostname,
        dns_server: result.dns.dns_servers?.join(",") || "unknown",
      },
      result.dns.resolution_time_ms,
    );
    dnsIpv4Count.set({ hostname: result.dns.hostname }, result.dns.ipv4_count);
    dnsIpv6Count.set({ hostname: result.dns.hostname }, result.dns.ipv6_count);
  }

  // TLS metrics
  if (result.tls) {
    tlsHandshakeTimeMs.set(
      {
        protocol: result.tls.protocol_version || "unknown",
        cipher_suite: result.tls.cipher_suite || "unknown",
      },
      result.tls.handshake_time_ms,
    );
  }

  // Traceroute metrics
  if (result.traceroute) {
    tracerouteHopsCount.set(
      { destination: result.traceroute.destination },
      result.traceroute.hops.length,
    );
    tracerouteCompleted.set(
      { destination: result.traceroute.destination },
      result.traceroute.completed ? 1 : 0,
    );
  }

  // Network information
  if (result.local_ipv4) {
    localIpv4.set({ address: result.local_ipv4 }, 1);
  }
  if (result.local_ipv6) {
    localIpv6.set({ address: result.local_ipv6 }, 1);
  }
  if (result.external_ipv4) {
    externalIpv4.set({ address: result.external_ipv4 }, 1);
  }
  if (result.external_ipv6) {
    externalIpv6.set({ address: result.external_ipv6 }, 1);
  }
}
