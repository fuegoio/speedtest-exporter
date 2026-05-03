import { describe, it, expect } from "bun:test";
import { register } from "../src/metrics";
import { updateMetrics } from "../src/metricsUpdater";
import type { RunResult } from "../src/model";

describe("metricsUpdater", () => {
  const createMockResult = (overrides: Partial<RunResult> = {}): RunResult => ({
    version: "1.0.0",
    timestamp_utc: "2024-01-01T00:00:00Z",
    base_url: "https://speed.cloudflare.com",
    meas_id: "test-123",
    server: "FRA",
    colo: "FRA",
    ip: "1.2.3.4",
    asn: "12345",
    as_org: "Test Org",
    interface_name: "eth0",
    network_name: "Test Network",
    idle_latency: {
      sent: 10,
      received: 10,
      loss: 0,
      min_ms: 5,
      mean_ms: 6,
      median_ms: 6,
      p25_ms: 5.5,
      p75_ms: 6.5,
      max_ms: 7,
      jitter_ms: 0.5,
    },
    download: {
      bytes: 1000000,
      duration_ms: 1000,
      mbps: 8,
      mean_mbps: 8.5,
      median_mbps: 8.2,
      p25_mbps: 7.8,
      p75_mbps: 9.2,
    },
    upload: {
      bytes: 500000,
      duration_ms: 1000,
      mbps: 4,
      mean_mbps: 4.5,
      median_mbps: 4.2,
      p25_mbps: 3.8,
      p75_mbps: 5.2,
    },
    loaded_latency_download: {
      sent: 10,
      received: 10,
      loss: 0,
      min_ms: 6,
      mean_ms: 7,
      median_ms: 7,
      p25_ms: 6.5,
      p75_ms: 7.5,
      max_ms: 8,
      jitter_ms: 0.6,
    },
    loaded_latency_upload: {
      sent: 10,
      received: 10,
      loss: 0,
      min_ms: 7,
      mean_ms: 8,
      median_ms: 8,
      p25_ms: 7.5,
      p75_ms: 8.5,
      max_ms: 9,
      jitter_ms: 0.7,
    },
    ...overrides,
  });

  it("should update all basic metrics", async () => {
    const result = createMockResult();
    updateMetrics(result);

    const metrics = await register.metrics();
    expect(metrics.includes("speedtest_download_mbps")).toBe(true);
    expect(metrics.includes("speedtest_upload_mbps")).toBe(true);
    expect(metrics.includes("speedtest_idle_latency_ms")).toBe(true);
  });

  it("should handle missing optional fields", () => {
    const result: RunResult = {
      base_url: "https://speed.cloudflare.com",
      meas_id: "test-123",
      idle_latency: { sent: 10, received: 10, loss: 0 },
      download: { bytes: 1000000, duration_ms: 1000, mbps: 8 },
      upload: { bytes: 500000, duration_ms: 1000, mbps: 4 },
      loaded_latency_download: { sent: 10, received: 10, loss: 0 },
      loaded_latency_upload: { sent: 10, received: 10, loss: 0 },
    };

    // Should not throw
    expect(() => updateMetrics(result)).not.toThrow();
  });
});
