# ⚡ Speedtest Exporter

A golang Prometheus exporter for network performance metrics. This exporter runs periodic speed tests against Cloudflare's speed test infrastructure (`speed.cloudflare.com`) and exposes comprehensive metrics in Prometheus format.

While it uses Cloudflare's public speed test endpoints as the default target, the exporter measures general network performance metrics including throughput, latency, packet loss, DNS resolution, and TLS handshake times.

## Features

- **Download/Upload Speed**: Measures download and upload speeds in Mbps
- **Latency Tests**: Measures idle latency, loaded latency (during download/upload)
- **DNS Resolution**: Measures DNS resolution time as a histogram over 10 runs
- **TLS Handshake**: Measures TLS handshake time and protocol/cipher info
- **Network Information**: Captures local/external IPs, ASN, ISP info
- **Prometheus Metrics**: Exposes all data as Prometheus-compatible metrics

## Metrics

### Common Labels

Most metrics share a common set of labels that describe the test context:

| Label         | Description                                                                 |
| ------------- | --------------------------------------------------------------------------- |
| `server`      | Hostname of the speed test server                                           |
| `colo`        | Cloudflare colocation (PoP) code closest to the server                      |
| `asn`         | Autonomous System Number of the client's network                            |
| `as_org`      | Name of the AS organization (ISP)                                           |
| `interface`   | Network interface name used for the test (overridable via `INTERFACE_NAME`) |
| `network`     | Logical network name (overridable via `NETWORK_NAME`)                       |
| `ip_version`  | IP version used: `ipv4`, `ipv6`, or `both`                                  |
| `country`     | Country of the server                                                       |
| `city`        | City of the server                                                          |
| `region`      | Region of the server                                                        |
| `postal_code` | Postal code of the server                                                   |
| `latitude`    | Latitude of the server                                                      |
| `longitude`   | Longitude of the server                                                     |

### Download Metrics

| Metric                    | Additional Labels | Description                        |
| ------------------------- | ----------------- | ---------------------------------- |
| `speedtest_download_mbps` | `size`            | Download speed in Mbps (histogram) |

- `size`: payload size used for the measurement (e.g. `max`)

**Note**: Histogram metric. Use `histogram_quantile()` in PromQL for percentile calculations.

### Upload Metrics

| Metric                  | Additional Labels | Description                      |
| ----------------------- | ----------------- | -------------------------------- |
| `speedtest_upload_mbps` | `size`            | Upload speed in Mbps (histogram) |

- `size`: payload size used for the measurement (e.g. `max`)

**Note**: Histogram metric. Use `histogram_quantile()` in PromQL for percentile calculations.

### Latency Metrics

| Metric                           | Additional Labels | Description                         |
| -------------------------------- | ----------------- | ----------------------------------- |
| `speedtest_latency_ms`           | `during`          | Latency in milliseconds (histogram) |
| `speedtest_latency_jitter_ms`    | `during`          | Jitter in milliseconds (histogram)  |
| `speedtest_latency_loss_percent` | `during`          | Packet loss percentage (histogram)  |

- `during`: phase of the test when the measurement was taken — `idle`, `download`, or `upload`

**Note**: Histogram metrics. Use `histogram_quantile()` in PromQL for percentile calculations.

### DNS Metrics

| Metric                             | Labels                   | Description                                     |
| ---------------------------------- | ------------------------ | ----------------------------------------------- |
| `speedtest_dns_resolution_time_ms` | `hostname`, `dns_server` | DNS resolution time in milliseconds (histogram) |

DNS metrics do not carry the common labels.

- `hostname`: the hostname being resolved (configurable via `DNS_HOSTNAME`)
- `dns_server`: comma-separated list of DNS servers used

**Note**: Histogram metric. Each test run contributes `DNS_RUNS` observations (default 10). Use `histogram_quantile()` in PromQL for percentile calculations.

### TLS Metrics

| Metric                            | Labels                     | Description                                    |
| --------------------------------- | -------------------------- | ---------------------------------------------- |
| `speedtest_tls_handshake_time_ms` | `protocol`, `cipher_suite` | TLS handshake time in milliseconds (histogram) |

TLS metrics do not carry the common labels.

- `protocol`: TLS protocol version negotiated (e.g. `TLSv1.3`)
- `cipher_suite`: cipher suite negotiated (e.g. `TLS_AES_128_GCM_SHA256`)

**Note**: Histogram metric. Each test run contributes `TLS_RUNS` observations (default 10). Use `histogram_quantile()` in PromQL for percentile calculations.

### Network Information Metrics

| Metric                    | Labels    | Description                          |
| ------------------------- | --------- | ------------------------------------ |
| `speedtest_local_ipv4`    | `address` | Local IPv4 address (1 if present)    |
| `speedtest_local_ipv6`    | `address` | Local IPv6 address (1 if present)    |
| `speedtest_external_ipv4` | `address` | External IPv4 address (1 if present) |
| `speedtest_external_ipv6` | `address` | External IPv6 address (1 if present) |

- `address`: the IP address as a string

### Test Metadata

| Metric                             | Labels       | Description                                       |
| ---------------------------------- | ------------ | ------------------------------------------------- |
| `speedtest_test_timestamp`         | —            | Timestamp of last test in Unix seconds            |
| `speedtest_test_duration_total_ms` | —            | Total duration of last test in milliseconds       |
| `speedtest_test_errors_total`      | `error_type` | Total number of test errors, by error type        |
| `speedtest_test_runs_total`        | `status`     | Total number of test runs (`success` or `failed`) |

## Quick Start

### Using Docker Compose

```yaml
services:
  speedtest-exporter:
    image: ghcr.io/fuegoio/speedtest-exporter:latest
    ports:
      - "9537:9537"
    environment:
      TEST_INTERVAL_MS: 3600000
    restart: unless-stopped
```

```bash
docker compose up -d
```

### Using Docker

```bash
# Run with default configuration
docker run -d -p 9537:9537 --name speedtest-exporter ghcr.io/fuegoio/speedtest-exporter:latest

# Or build from source
docker build -t speedtest-exporter .
docker run -d -p 9537:9537 --name speedtest-exporter speedtest-exporter
```

### Using Go

```bash
# Build and run
go build ./cmd/speedtest-exporter/
./speedtest-exporter

# Or install
go install ./cmd/speedtest-exporter/
```

## Endpoints

| Endpoint   | Description                   |
| ---------- | ----------------------------- |
| `/metrics` | Prometheus metrics            |
| `/health`  | Health check (returns "OK")   |
| `/run`     | Manually trigger a speed test |
| `/`        | Help text with endpoint list  |

## Configuration

All configuration is done via environment variables:

| Variable            | Default                      | Description                            |
| ------------------- | ---------------------------- | -------------------------------------- |
| `PORT`              | 9537                         | HTTP server port                       |
| `BASE_URL`          | https://speed.cloudflare.com | Cloudflare speedtest base URL          |
| `TEST_INTERVAL_MS`  | 3600000 (1 hour)             | Interval between tests in milliseconds |
| `PROBE_INTERVAL_MS`   | 250                          | Interval between latency probes in milliseconds  |
| `PROBE_TIMEOUT_MS`    | 800                          | Timeout for individual probes in milliseconds    |
| `LATENCY_DURATION_MS` | 10000 (10 seconds)           | Duration of each latency measurement phase       |
| `DNS_HOSTNAME`      | hostname from `BASE_URL`     | Hostname to resolve in DNS tests       |
| `DNS_RUNS`          | 10                           | Number of DNS resolution runs per test |
| `SKIP_DNS`          | false                        | Skip DNS diagnostics                   |
| `TLS_RUNS`          | 10                           | Number of TLS handshake runs per test  |
| `SKIP_TLS`          | false                        | Skip TLS diagnostics                   |
| `ASN`               | -                            | Override ASN                           |
| `AS_ORG`            | -                            | Override AS organization               |
| `INTERFACE_NAME`    | -                            | Override interface name                |
| `NETWORK_NAME`      | -                            | Override network name                  |
| `LOCAL_IPV4`        | -                            | Override local IPv4                    |
| `LOCAL_IPV6`        | -                            | Override local IPv6                    |
| `EXTERNAL_IPV4`     | -                            | Override external IPv4                 |
| `EXTERNAL_IPV6`     | -                            | Override external IPv6                 |

### Example with custom configuration

```bash
docker run -d -p 9537:9537 \
  -e TEST_INTERVAL_MS=300000 \
  --name speedtest-exporter \
  speedtest-exporter
```

## Prometheus Configuration

```yaml
scrape_configs:
  - job_name: "speedtest"
    static_configs:
      - targets: ["localhost:9537"]
    scrape_interval: 30s
```

````

## Development

```bash
# Run tests
go test ./...

# Build
go build ./cmd/speedtest-exporter/

# Run locally
PORT=9537 TEST_INTERVAL_MS=60000 ./speedtest-exporter
````

## License

MIT License
