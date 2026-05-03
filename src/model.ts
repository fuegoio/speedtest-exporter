// Interfaces for Cloudflare speedtest data

// Latency summary with statistics
export interface LatencySummary {
  sent: number;
  received: number;
  loss: number; // 0.0 to 1.0
  min_ms?: number;
  mean_ms?: number;
  median_ms?: number;
  p25_ms?: number;
  p75_ms?: number;
  max_ms?: number;
  jitter_ms?: number;
}

// Throughput summary with statistics
export interface ThroughputSummary {
  bytes: number;
  duration_ms: number;
  mbps: number;
  mean_mbps?: number;
  median_mbps?: number;
  p25_mbps?: number;
  p75_mbps?: number;
}

// DNS resolution summary
export interface DnsSummary {
  hostname: string;
  resolution_time_ms: number;
  resolved_ips: string[];
  ipv4_count: number;
  ipv6_count: number;
  dns_servers?: string[];
}

// TLS handshake summary
export interface TlsSummary {
  handshake_time_ms: number;
  protocol_version?: string;
  cipher_suite?: string;
}

// Traceroute hop information
export interface TracerouteHop {
  hop_number: number;
  ip_address?: string;
  hostname?: string;
  rtt_ms: number[];
  timeout: boolean;
}

// Traceroute summary
export interface TracerouteSummary {
  destination: string;
  hops: TracerouteHop[];
  completed: boolean;
}

// TURN server information
export interface TurnInfo {
  urls: string[];
  username?: string;
  credential?: string;
}

// Complete test result
export interface RunResult {
  version?: string;
  timestamp_utc?: string;
  base_url: string;
  meas_id: string;
  comments?: string;
  meta?: unknown;
  server?: string;
  idle_latency: LatencySummary;
  download: ThroughputSummary;
  upload: ThroughputSummary;
  loaded_latency_download: LatencySummary;
  loaded_latency_upload: LatencySummary;
  turn?: TurnInfo;
  // Network information
  ip?: string;
  colo?: string;
  asn?: string;
  as_org?: string;
  interface_name?: string;
  network_name?: string;
  is_wireless?: boolean;
  interface_mac?: string;
  local_ipv4?: string;
  local_ipv6?: string;
  external_ipv4?: string;
  external_ipv6?: string;
  // Diagnostic results
  dns?: DnsSummary;
  tls?: TlsSummary;
  traceroute?: TracerouteSummary;
}
