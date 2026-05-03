# Speedtest Exporter

A Bun-powered Prometheus exporter for network performance metrics. This exporter runs periodic speed tests against Cloudflare's speed test infrastructure (`speed.cloudflare.com`) and exposes comprehensive metrics in Prometheus format.

While it uses Cloudflare's public speed test endpoints as the default target, the exporter measures general network performance metrics including throughput, latency, packet loss, DNS resolution, and TLS handshake times.

## Metrics Exposed

All metrics are prefixed with `speedtest_` for easy filtering in Prometheus.

### Throughput Metrics

- `speedtest_download_mbps` - Download speed in Mbps
- `speedtest_download_bytes_total` - Total bytes downloaded
- `speedtest_download_duration_ms` - Download test duration
- `speedtest_upload_mbps` - Upload speed in Mbps
- `speedtest_upload_bytes_total` - Total bytes uploaded
- `speedtest_upload_duration_ms` - Upload test duration

### Latency Metrics

- `speedtest_idle_latency_ms` - Current idle latency
- `speedtest_idle_latency_min_ms` - Minimum idle latency
- `speedtest_idle_latency_mean_ms` - Mean idle latency
- `speedtest_idle_latency_median_ms` - Median idle latency
- `speedtest_idle_latency_p25_ms` - 25th percentile idle latency
- `speedtest_idle_latency_p75_ms` - 75th percentile idle latency
- `speedtest_idle_latency_max_ms` - Maximum idle latency
- `speedtest_idle_latency_jitter_ms` - Idle latency jitter
- `speedtest_loaded_latency_download_*` - Latency during download
- `speedtest_loaded_latency_upload_*` - Latency during upload

### Packet Loss Metrics

- `speedtest_idle_latency_loss_percent` - Packet loss during idle test
- `speedtest_loaded_latency_download_loss_percent` - Packet loss during download
- `speedtest_loaded_latency_upload_loss_percent` - Packet loss during upload

### Diagnostic Metrics

- `speedtest_dns_resolution_time_ms` - DNS resolution time
- `speedtest_dns_ipv4_count` - Number of IPv4 addresses resolved
- `speedtest_dns_ipv6_count` - Number of IPv6 addresses resolved
- `speedtest_tls_handshake_time_ms` - TLS handshake time

### Traceroute Metrics

- `speedtest_traceroute_hops_count` - Number of hops
- `speedtest_traceroute_completed` - Whether traceroute completed

### Network Information

- `speedtest_local_ipv4` - Local IPv4 address presence
- `speedtest_local_ipv6` - Local IPv6 address presence
- `speedtest_external_ipv4` - External IPv4 address presence
- `speedtest_external_ipv6` - External IPv6 address presence

### Test Metadata

- `speedtest_test_timestamp` - Timestamp of last test
- `speedtest_test_duration_total_ms` - Total test duration
- `speedtest_test_runs_total` - Total test runs (with status label)
- `speedtest_test_errors_total` - Total test errors (with error_type label)

## Installation

### Prerequisites

- [Bun](https://bun.sh/) runtime installed
- Node.js 18+ (for development)

### Quick Start

```bash
# Install dependencies
bun install

# Start the exporter
bun run start
```

### Docker

The project includes a production-ready multi-stage Dockerfile that uses Alpine Linux for a minimal footprint.

Build and run:

```bash
docker build -t speedtest-exporter .
docker run -p 9537:9537 -e TEST_INTERVAL_MS=60000 speedtest-exporter
```

## Configuration

All configuration is done via environment variables:

| Variable                   | Default                      | Description                      |
| -------------------------- | ---------------------------- | -------------------------------- |
| `PORT`                     | 9537                         | HTTP server port                 |
| `TEST_INTERVAL_MS`         | 3600000 (1 hour)             | Test frequency in milliseconds   |
| `BASE_URL`                 | https://speed.cloudflare.com | Cloudflare speedtest base URL    |
| `DOWNLOAD_DURATION_MS`     | 10000                        | Download test duration           |
| `UPLOAD_DURATION_MS`       | 10000                        | Upload test duration             |
| `IDLE_LATENCY_DURATION_MS` | 2000                         | Idle latency test duration       |
| `CONCURRENCY`              | 6                            | Number of concurrent connections |
| `DOWNLOAD_BYTES_PER_REQ`   | 10000000                     | Bytes per download request       |
| `UPLOAD_BYTES_PER_REQ`     | 5000000                      | Bytes per upload request         |
| `PROBE_INTERVAL_MS`        | 250                          | Probe interval for latency tests |
| `PROBE_TIMEOUT_MS`         | 800                          | Probe timeout                    |
| `SKIP_DIAGNOSTICS`         | false                        | Skip DNS/TLS diagnostics         |
| `TRACEROUTE`               | false                        | Run traceroute                   |

### Network Information

Network information (ASN, AS organization, local/external IPs, interface details) is **automatically fetched** from external services and the local system. The exporter uses:

- **ifconfig.co** - Primary source for ASN, AS organization, and external IP addresses
- **api.ipify.org** - Fallback for external IPv4 address
- **api6.ipify.org** - Fallback for external IPv6 address
- **OS network interfaces** - Local IP addresses and interface information

If you need to override any of these values (e.g., in a containerized environment), you can still use the following environment variables:

| Variable         | Description                  |
| ---------------- | ---------------------------- |
| `ASN`            | ASN for labeling metrics     |
| `AS_ORG`         | AS organization for labeling |
| `INTERFACE_NAME` | Network interface name       |
| `NETWORK_NAME`   | Network name                 |
| `LOCAL_IPV4`     | Local IPv4 address           |
| `LOCAL_IPV6`     | Local IPv6 address           |
| `EXTERNAL_IPV4`  | External IPv4 address        |
| `EXTERNAL_IPV6`  | External IPv6 address        |

## Usage

### Start the exporter

```bash
bun run start
```

### Start with custom interval (1 minute)

```bash
TEST_INTERVAL_MS=60000 bun run start
```

### Start on custom port

```bash
PORT=8080 bun run start
```

### Access metrics

- Prometheus metrics: `http://localhost:9537/metrics`
- Health check: `http://localhost:9537/health`
- Manual test trigger: `http://localhost:9537/run`

### Prometheus Configuration

Add to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: "speedtest"
    static_configs:
      - targets: ["localhost:9537"]
    scrape_interval: 30s
```

## Development

```bash
# Install dependencies
bun install

# Run with hot reload
bun run dev

# Run type checking
bun run typecheck

# Run linting
bun run lint

# Run formatting
bun run format

# Run all tests
bun run test

# Run CI checks (lint, format, typecheck, test)
bun run ci
```

## License

MIT

## Inspiration

This exporter is inspired by [cloudflare-speed-cli](https://github.com/kavehtehrani/cloudflare-speed-cli) and provides a Prometheus-compatible way to monitor network performance using Cloudflare's speed test infrastructure.
