package model

// LatencySummary contains latency test statistics
type LatencySummary struct {
	Sent       int     `json:"sent"`
	Received   int     `json:"received"`
	Loss       float64 `json:"loss"` // 0.0 to 1.0
	MinMs      *float64 `json:"min_ms,omitempty"`
	MeanMs     *float64 `json:"mean_ms,omitempty"`
	MedianMs   *float64 `json:"median_ms,omitempty"`
	P25Ms      *float64 `json:"p25_ms,omitempty"`
	P75Ms      *float64 `json:"p75_ms,omitempty"`
	MaxMs      *float64 `json:"max_ms,omitempty"`
	JitterMs   *float64 `json:"jitter_ms,omitempty"`
}

// ThroughputSummary contains throughput test statistics
type ThroughputSummary struct {
	Bytes        int64   `json:"bytes"`
	DurationMs   int64   `json:"duration_ms"`
	Mbps         float64 `json:"mbps"`
	MeanMbps     *float64 `json:"mean_mbps,omitempty"`
	MedianMbps   *float64 `json:"median_mbps,omitempty"`
	P25Mbps      *float64 `json:"p25_mbps,omitempty"`
	P75Mbps      *float64 `json:"p75_mbps,omitempty"`
}

// DnsSummary contains DNS resolution statistics
type DnsSummary struct {
	Hostname         string   `json:"hostname"`
	ResolutionTimeMs float64  `json:"resolution_time_ms"`
	ResolvedIPs      []string `json:"resolved_ips"`
	Ipv4Count       int      `json:"ipv4_count"`
	Ipv6Count       int      `json:"ipv6_count"`
	DnsServers       []string `json:"dns_servers,omitempty"`
}

// TlsSummary contains TLS handshake statistics
type TlsSummary struct {
	HandshakeTimeMs float64  `json:"handshake_time_ms"`
	ProtocolVersion *string  `json:"protocol_version,omitempty"`
	CipherSuite     *string  `json:"cipher_suite,omitempty"`
}

// RunResult contains the complete test result
type RunResult struct {
	Version              string           `json:"version,omitempty"`
	TimestampUTC         string           `json:"timestamp_utc,omitempty"`
	BaseURL              string           `json:"base_url"`
	MeasID               string           `json:"meas_id"`
	Server               *string          `json:"server,omitempty"`
	Colo                 *string          `json:"colo,omitempty"`
	IP                   *string          `json:"ip,omitempty"`
	IdleLatency          LatencySummary   `json:"idle_latency"`
	Download             ThroughputSummary `json:"download"`
	Upload               ThroughputSummary `json:"upload"`
	LoadedLatencyDownload LatencySummary   `json:"loaded_latency_download"`
	LoadedLatencyUpload   LatencySummary   `json:"loaded_latency_upload"`
	Dns                  *DnsSummary       `json:"dns,omitempty"`
	Tls                  *TlsSummary       `json:"tls,omitempty"`
	// Network information
	Asn           *string `json:"asn,omitempty"`
	AsOrg        *string `json:"as_org,omitempty"`
	InterfaceName *string `json:"interface_name,omitempty"`
	NetworkName  *string `json:"network_name,omitempty"`
	LocalIpv4    *string `json:"local_ipv4,omitempty"`
	LocalIpv6    *string `json:"local_ipv6,omitempty"`
	ExternalIpv4 *string `json:"external_ipv4,omitempty"`
	ExternalIpv6 *string `json:"external_ipv6,omitempty"`
	// Geo information
	Country    *string `json:"country,omitempty"`
	City       *string `json:"city,omitempty"`
	Region     *string `json:"region,omitempty"`
	PostalCode *string `json:"postal_code,omitempty"`
	Latitude   *string `json:"latitude,omitempty"`
	Longitude  *string `json:"longitude,omitempty"`
}
