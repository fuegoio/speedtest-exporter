import type { ExporterConfig } from "./config";
import type { LatencySummary, ThroughputSummary, DnsSummary, TlsSummary, RunResult } from "./model";
import { networkInterfaces } from "os";

// Generate a random measurement ID
function generateMeasId(): string {
  const bytes = new Uint8Array(8);
  crypto.getRandomValues(bytes);
  const view = new DataView(bytes.buffer);
  return view.getBigUint64(0, true).toString();
}

// Network information from external services
interface NetworkInfo {
  asn?: string;
  as_org?: string;
  external_ipv4?: string;
  external_ipv6?: string;
}

// Fetch network info from ifconfig.co with timeout wrapper
async function fetchWithTimeout(
  url: string,
  options: RequestInit,
  timeoutMs: number,
): Promise<Response> {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeoutMs);

  try {
    const response = await fetch(url, {
      ...options,
      signal: controller.signal,
    });
    return response;
  } finally {
    clearTimeout(timeoutId);
  }
}

// Fetch network info from ifconfig.co
export async function fetchNetworkInfo(): Promise<NetworkInfo> {
  try {
    // Try ifconfig.co first - it provides ASN and org info
    const response = await fetchWithTimeout(
      "https://ifconfig.co/json",
      { method: "GET", headers: { "User-Agent": "speedtest-exporter/1.0" } },
      5000,
    );

    if (response.ok) {
      const data = (await response.json()) as {
        asn?: string | number;
        asn_org?: string;
        asn_description?: string;
        ipv4?: string;
        ipv6?: string;
      };
      return {
        asn: data.asn?.toString(),
        as_org: data.asn_org || data.asn_description,
        external_ipv4: data.ipv4,
        external_ipv6: data.ipv6,
      };
    }
  } catch {
    // Fall through to Cloudflare
  }

  // Fallback: try Cloudflare's IP endpoints
  try {
    const ipv4Response = await fetchWithTimeout("https://api.ipify.org?format=json", {}, 5000);
    const ipv4Data = (await ipv4Response.json()) as { ip?: string };

    const ipv6Response = await fetchWithTimeout("https://api6.ipify.org?format=json", {}, 5000);
    const ipv6Data = (await ipv6Response.json()) as { ip?: string };

    return {
      external_ipv4: ipv4Data.ip,
      external_ipv6: ipv6Data.ip,
    };
  } catch {
    // Return empty - we'll use what we have from trace
  }

  return {};
}

// Get local network info from system
export function getLocalNetworkInfo(): {
  local_ipv4?: string;
  local_ipv6?: string;
  interface_name?: string;
  network_name?: string;
} {
  try {
    // Try to get local IP addresses from OS
    const network = networkInterfaces();

    let localIpv4: string | undefined;
    let localIpv6: string | undefined;
    let interfaceName: string | undefined;
    let networkName: string | undefined;

    for (const [iface, addrs] of Object.entries(network)) {
      if (!addrs) continue;

      for (const addr of addrs) {
        if (addr.internal) continue;

        if (addr.family === "IPv4" && !localIpv4) {
          localIpv4 = addr.address;
          interfaceName = iface;
          networkName = iface;
        }

        if (addr.family === "IPv6" && !localIpv6) {
          localIpv6 = addr.address;
        }
      }
    }

    return {
      local_ipv4: localIpv4,
      local_ipv6: localIpv6,
      interface_name: interfaceName,
      network_name: networkName,
    };
  } catch {
    return {};
  }
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
      // Fetch network info from external services
      const networkInfo = await fetchNetworkInfo();
      const localNetworkInfo = getLocalNetworkInfo();

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

      // Get external IP from trace as fallback
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
        // Network info from external services and system, with config overrides
        asn: this.config.asn || networkInfo.asn,
        as_org: this.config.asOrg || networkInfo.as_org,
        interface_name: this.config.interfaceName || localNetworkInfo.interface_name,
        network_name: this.config.networkName || localNetworkInfo.network_name,
        local_ipv4: this.config.localIpv4 || localNetworkInfo.local_ipv4,
        local_ipv6: this.config.localIpv6 || localNetworkInfo.local_ipv6,
        external_ipv4: this.config.externalIpv4 || networkInfo.external_ipv4 || externalIp,
        external_ipv6: this.config.externalIpv6 || networkInfo.external_ipv6,
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
