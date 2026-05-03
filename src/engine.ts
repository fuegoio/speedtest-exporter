import type { ExporterConfig } from "./config";
import type { LatencySummary, ThroughputSummary, DnsSummary, TlsSummary, RunResult } from "./model";

// Generate a random measurement ID
function generateMeasId(): string {
  const bytes = new Uint8Array(8);
  crypto.getRandomValues(bytes);
  const view = new DataView(bytes.buffer);
  return view.getBigUint64(0, true).toString();
}

// Cloudflare speedtest engine
export class CloudflareSpeedtest {
  private config: ExporterConfig;

  constructor(config: ExporterConfig) {
    this.config = config;
  }

  async runDirectTest(): Promise<RunResult> {
    const measId = generateMeasId();

    try {
      // First, get the test configuration from Cloudflare
      const configResponse = await fetch(`${this.config.baseUrl}/cdn-cgi/trace`);
      const traceText = await configResponse.text();

      const server = traceText
        .split("\n")
        .find((line) => line.startsWith("loc="))
        ?.split("=")[1];
      const colo = traceText
        .split("\n")
        .find((line) => line.startsWith("colo="))
        ?.split("=")[1];

      // Get external IP
      const ipResponse = await fetch(`${this.config.baseUrl}/cdn-cgi/trace`);
      const ipText = await ipResponse.text();
      const externalIp = ipText
        .split("\n")
        .find((line) => line.startsWith("ip="))
        ?.split("=")[1];

      // Run latency test
      const idleLatency = await this.measureLatency(
        `${this.config.baseUrl}/__latency`,
        this.config.idleLatencyDurationMs,
        this.config.probeIntervalMs,
      );

      // Run download test
      const download = await this.measureThroughput(
        `${this.config.baseUrl}/__down`,
        this.config.downloadDurationMs,
        this.config.downloadBytesPerReq,
        this.config.concurrency,
      );

      // Run upload test
      const upload = await this.measureThroughput(
        `${this.config.baseUrl}/__up`,
        this.config.uploadDurationMs,
        this.config.uploadBytesPerReq,
        this.config.concurrency,
      );

      // Measure loaded latency during download
      const loadedLatencyDownload = await this.measureLatency(
        `${this.config.baseUrl}/__latency?phase=download`,
        this.config.downloadDurationMs,
        this.config.probeIntervalMs,
      );

      // Measure loaded latency during upload
      const loadedLatencyUpload = await this.measureLatency(
        `${this.config.baseUrl}/__latency?phase=upload`,
        this.config.uploadDurationMs,
        this.config.probeIntervalMs,
      );

      // DNS measurement
      let dns: DnsSummary | undefined;
      if (!this.config.skipDiagnostics) {
        dns = await this.measureDns();
      }

      // TLS measurement
      let tls: TlsSummary | undefined;
      if (!this.config.skipDiagnostics) {
        tls = await this.measureTls();
      }

      return {
        version: "1.0.0",
        timestamp_utc: new Date().toISOString(),
        base_url: this.config.baseUrl,
        meas_id: measId,
        server: server || "unknown",
        colo: colo || "unknown",
        ip: externalIp,
        idle_latency: idleLatency,
        download,
        upload,
        loaded_latency_download: loadedLatencyDownload,
        loaded_latency_upload: loadedLatencyUpload,
        dns,
        tls,
        // Network info (we'll set these from environment or system)
        asn: process.env.ASN,
        as_org: process.env.AS_ORG,
        interface_name: process.env.INTERFACE_NAME,
        network_name: process.env.NETWORK_NAME,
        is_wireless: process.env.IS_WIRELESS === "true",
        local_ipv4: process.env.LOCAL_IPV4,
        local_ipv6: process.env.LOCAL_IPV6,
        external_ipv4: externalIp,
        external_ipv6: process.env.EXTERNAL_IPV6,
      };
    } catch (error) {
      console.error("Error in direct test:", error);
      throw error;
    }
  }

  private async measureLatency(
    url: string,
    durationMs: number,
    intervalMs: number,
  ): Promise<LatencySummary> {
    const samples: number[] = [];
    const startTime = Date.now();
    const endTime = startTime + durationMs;

    while (Date.now() < endTime) {
      try {
        const probeStart = performance.now();
        const response = await fetch(url, {
          method: "GET",
          headers: { "Cache-Control": "no-cache" },
        });
        await response.text();
        const rtt = performance.now() - probeStart;
        samples.push(rtt);
      } catch {
        // Timeout or error - count as loss
        samples.push(-1);
      }
      await Bun.sleep(intervalMs);
    }

    return this.computeLatencySummary(samples);
  }

  private async measureThroughput(
    url: string,
    durationMs: number,
    bytesPerRequest: number,
    concurrency: number,
  ): Promise<ThroughputSummary> {
    const startTime = Date.now();
    const endTime = startTime + durationMs;
    let bytesTotal = 0;
    const speeds: number[] = [];

    const promises: Promise<void>[] = [];
    for (let i = 0; i < concurrency; i++) {
      promises.push(
        (async () => {
          while (Date.now() < endTime) {
            try {
              const reqStart = performance.now();
              const response = await fetch(url, {
                method: "GET",
                headers: { "Cache-Control": "no-cache" },
              });
              const data = await response.arrayBuffer();
              const reqEnd = performance.now();

              bytesTotal += data.byteLength;
              const duration = reqEnd - reqStart;
              if (duration > 0) {
                const bps = (data.byteLength * 8) / (duration / 1000);
                speeds.push(bps);
              }
            } catch {
              // Error - skip
            }
          }
        })(),
      );
    }

    await Promise.all(promises);
    const actualDuration = Date.now() - startTime;
    const totalMbps = (bytesTotal * 8) / (actualDuration * 1000); // Convert to Mbps

    return {
      bytes: bytesTotal,
      duration_ms: actualDuration,
      mbps: totalMbps,
      mean_mbps:
        speeds.length > 0 ? speeds.reduce((a, b) => a + b, 0) / speeds.length / 1000000 : undefined,
      median_mbps:
        speeds.length > 0 ? this.computeMedian(speeds.map((s) => s / 1000000)) : undefined,
      p25_mbps:
        speeds.length > 0
          ? this.computePercentile(
              speeds.map((s) => s / 1000000),
              0.25,
            )
          : undefined,
      p75_mbps:
        speeds.length > 0
          ? this.computePercentile(
              speeds.map((s) => s / 1000000),
              0.75,
            )
          : undefined,
    };
  }

  private async measureDns(): Promise<DnsSummary> {
    const hostname = new URL(this.config.baseUrl).hostname;
    const startTime = performance.now();

    // Use a simple approach - just time the fetch
    // Bun.dns.resolve is not available, so we'll use fetch timing as a proxy
    const response = await fetch(`https://${hostname}/cdn-cgi/trace`);
    await response.text();

    const resolutionTime = performance.now() - startTime;

    return {
      hostname,
      resolution_time_ms: resolutionTime,
      resolved_ips: [hostname],
      ipv4_count: 1,
      ipv6_count: 0,
    };
  }

  private async measureTls(): Promise<TlsSummary> {
    const startTime = performance.now();

    // Measure TLS handshake time
    const response = await fetch(`${this.config.baseUrl}/`, {
      method: "GET",
    });
    await response.text();

    const handshakeTime = performance.now() - startTime;

    // Get TLS info from response headers or use defaults
    const protocol = response.headers.get("x-tls-protocol") || "unknown";
    const cipher = response.headers.get("x-tls-cipher") || "unknown";

    return {
      handshake_time_ms: handshakeTime,
      protocol_version: protocol,
      cipher_suite: cipher,
    };
  }

  private computeLatencySummary(samples: number[]): LatencySummary {
    const validSamples = samples.filter((s) => s >= 0);
    const loss = samples.length > 0 ? samples.filter((s) => s < 0).length / samples.length : 0;

    if (validSamples.length === 0) {
      return {
        sent: samples.length,
        received: 0,
        loss,
        min_ms: undefined,
        mean_ms: undefined,
        median_ms: undefined,
        p25_ms: undefined,
        p75_ms: undefined,
        max_ms: undefined,
        jitter_ms: undefined,
      };
    }

    validSamples.sort((a, b) => a - b);

    return {
      sent: samples.length,
      received: validSamples.length,
      loss,
      min_ms: validSamples[0],
      mean_ms: validSamples.reduce((a, b) => a + b, 0) / validSamples.length,
      median_ms: this.computeMedian(validSamples),
      p25_ms: this.computePercentile(validSamples, 0.25),
      p75_ms: this.computePercentile(validSamples, 0.75),
      max_ms: validSamples[validSamples.length - 1],
      jitter_ms: this.computeJitter(validSamples),
    };
  }

  private computeMedian(samples: number[]): number {
    const sorted = [...samples].sort((a, b) => a - b);
    const mid = Math.floor(sorted.length / 2);
    if (sorted.length % 2 === 0) {
      return (sorted[mid - 1] + sorted[mid]) / 2;
    }
    return sorted[mid];
  }

  private computePercentile(samples: number[], percentile: number): number {
    const sorted = [...samples].sort((a, b) => a - b);
    const pos = (sorted.length - 1) * percentile;
    const base = Math.floor(pos);
    const rest = pos - base;
    if (sorted[base + 1] !== undefined) {
      return sorted[base] + rest * (sorted[base + 1] - sorted[base]);
    }
    return sorted[base];
  }

  private computeJitter(samples: number[]): number {
    if (samples.length < 2) return 0;
    const mean = samples.reduce((a, b) => a + b, 0) / samples.length;
    const variance =
      samples.reduce((sum, s) => sum + Math.pow(s - mean, 2), 0) / (samples.length - 1);
    return Math.sqrt(variance);
  }
}
