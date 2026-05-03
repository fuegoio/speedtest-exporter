# Speedtest Exporter

A golang Prometheus exporter for network performance metrics. This exporter runs periodic speed tests against Cloudflare's speed test infrastructure (`speed.cloudflare.com`) and exposes comprehensive metrics in Prometheus format.

While it uses Cloudflare's public speed test endpoints as the default target, the exporter measures general network performance metrics including throughput, latency, packet loss, DNS resolution, and TLS handshake times.

## Features

- **Download/Upload Speed**: Measures download and upload speeds in Mbps
- **Latency Tests**: Measures idle latency, loaded latency (during download/upload)
- **DNS Resolution**: Measures DNS resolution time and IP counts
- **TLS Handshake**: Measures TLS handshake time and protocol/cipher info
- **Network Information**: Captures local/external IPs, ASN, ISP info
- **Prometheus Metrics**: Exposes all data as Prometheus-compatible metrics

## Metrics

### Download Metrics

| Metric                           | Labels                                                    | Description                               |
| -------------------------------- | --------------------------------------------------------- | ----------------------------------------- |
| `speedtest_download_mbps`        | server, colo, asn, as_org, interface, network, ip_version | Current download speed in Mbps            |
| `speedtest_download_bytes_total` | server, colo, asn, as_org                                 | Total bytes downloaded in last test       |
| `speedtest_download_duration_ms` | server, colo                                              | Duration of download test in milliseconds |

### Upload Metrics

| Metric                         | Labels                                                    | Description                             |
| ------------------------------ | --------------------------------------------------------- | --------------------------------------- |
| `speedtest_upload_mbps`        | server, colo, asn, as_org, interface, network, ip_version | Current upload speed in Mbps            |
| `speedtest_upload_bytes_total` | server, colo, asn, as_org                                 | Total bytes uploaded in last test       |
| `speedtest_upload_duration_ms` | server, colo                                              | Duration of upload test in milliseconds |

### Latency Metrics (Idle)

| Metric                                | Labels                                                          | Description                                     |
| ------------------------------------- | --------------------------------------------------------------- | ----------------------------------------------- |
| `speedtest_idle_latency_ms`           | server, colo, asn, as_org, interface, network, ip_version, type | Idle latency in milliseconds                    |
| `speedtest_idle_latency_min_ms`       | server, colo                                                    | Minimum idle latency                            |
| `speedtest_idle_latency_mean_ms`      | server, colo                                                    | Mean idle latency                               |
| `speedtest_idle_latency_median_ms`    | server, colo                                                    | Median idle latency                             |
| `speedtest_idle_latency_p25_ms`       | server, colo                                                    | 25th percentile idle latency                    |
| `speedtest_idle_latency_p75_ms`       | server, colo                                                    | 75th percentile idle latency                    |
| `speedtest_idle_latency_max_ms`       | server, colo                                                    | Maximum idle latency                            |
| `speedtest_idle_latency_jitter_ms`    | server, colo                                                    | Idle latency jitter                             |
| `speedtest_idle_latency_loss_percent` | server, colo                                                    | Packet loss percentage during idle latency test |

### Loaded Latency Metrics (During Download/Upload)

Similar metrics to idle latency but measured during download/upload tests:

- `speedtest_loaded_latency_download_*`
- `speedtest_loaded_latency_upload_*`

### DNS Metrics

| Metric                             | Labels               | Description                         |
| ---------------------------------- | -------------------- | ----------------------------------- |
| `speedtest_dns_resolution_time_ms` | hostname, dns_server | DNS resolution time in milliseconds |
| `speedtest_dns_ipv4_count`         | hostname             | Number of IPv4 addresses resolved   |
| `speedtest_dns_ipv6_count`         | hostname             | Number of IPv6 addresses resolved   |

### TLS Metrics

| Metric                            | Labels                 | Description                        |
| --------------------------------- | ---------------------- | ---------------------------------- |
| `speedtest_tls_handshake_time_ms` | protocol, cipher_suite | TLS handshake time in milliseconds |

### Network Information Metrics

| Metric                    | Labels  | Description                          |
| ------------------------- | ------- | ------------------------------------ |
| `speedtest_local_ipv4`    | address | Local IPv4 address (1 if present)    |
| `speedtest_local_ipv6`    | address | Local IPv6 address (1 if present)    |
| `speedtest_external_ipv4` | address | External IPv4 address (1 if present) |
| `speedtest_external_ipv6` | address | External IPv6 address (1 if present) |

### Test Metadata

| Metric                             | Description                                                   |
| ---------------------------------- | ------------------------------------------------------------- |
| `speedtest_test_timestamp`         | Timestamp of last test in Unix seconds                        |
| `speedtest_test_duration_total_ms` | Total duration of last test in milliseconds                   |
| `speedtest_test_errors_total`      | Total number of test errors (labeled by error_type)           |
| `speedtest_test_runs_total`        | Total number of test runs (labeled by status: success/failed) |

## Quick Start

### Using Docker

```bash
# Run with default configuration
docker run -d -p 9537:9537 --name speedtest-exporter ghcr.io/alexis/speedtest-exporter:latest

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

| Variable                   | Default                      | Description                                        |
| -------------------------- | ---------------------------- | -------------------------------------------------- |
| `PORT`                     | 9537                         | HTTP server port                                   |
| `BASE_URL`                 | https://speed.cloudflare.com | Cloudflare speedtest base URL                      |
| `TEST_INTERVAL_MS`         | 3600000 (1 hour)             | Interval between tests in milliseconds             |
| `DOWNLOAD_DURATION_MS`     | 10000 (10s)                  | Download test duration                             |
| `UPLOAD_DURATION_MS`       | 10000 (10s)                  | Upload test duration                               |
| `IDLE_LATENCY_DURATION_MS` | 2000 (2s)                    | Idle latency test duration                         |
| `CONCURRENCY`              | 6                            | Number of concurrent requests for throughput tests |
| `DOWNLOAD_BYTES_PER_REQ`   | 10000000 (10MB)              | Bytes per download request                         |
| `UPLOAD_BYTES_PER_REQ`     | 5000000 (5MB)                | Bytes per upload request                           |
| `PROBE_INTERVAL_MS`        | 250                          | Interval between latency probes                    |
| `PROBE_TIMEOUT_MS`         | 800                          | Timeout for individual probes                      |
| `SKIP_DIAGNOSTICS`         | false                        | Skip DNS and TLS diagnostics                       |
| `TRACEROUTE`               | false                        | Enable traceroute (not implemented)                |
| `ASN`                      | -                            | Override ASN                                       |
| `AS_ORG`                   | -                            | Override AS organization                           |
| `INTERFACE_NAME`           | -                            | Override interface name                            |
| `NETWORK_NAME`             | -                            | Override network name                              |
| `LOCAL_IPV4`               | -                            | Override local IPv4                                |
| `LOCAL_IPV6`               | -                            | Override local IPv6                                |
| `EXTERNAL_IPV4`            | -                            | Override external IPv4                             |
| `EXTERNAL_IPV6`            | -                            | Override external IPv6                             |

### Example with custom configuration

```bash
docker run -d -p 9537:9537 \
  -e TEST_INTERVAL_MS=300000 \  # 5 minutes
  -e CONCURRENCY=4 \
  -e DOWNLOAD_DURATION_MS=5000 \
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
