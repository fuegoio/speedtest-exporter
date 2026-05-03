package engine

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alexis/speedtest-exporter/internal/config"
	"github.com/alexis/speedtest-exporter/internal/model"
)

// NetworkInfo contains network information from external services
type NetworkInfo struct {
	Asn          *string
	AsOrg        *string
	ExternalIPv4 *string
	ExternalIPv6 *string
}

// generateMeasID generates a random measurement ID
func generateMeasID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "0"
	}
	return string(buf[:])
}

// computeMedian computes the median of a slice of floats
func computeMedian(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	sorted := make([]float64, len(samples))
	copy(sorted, samples)
	sort.Float64s(sorted)

	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

// computePercentile computes the percentile of a sorted slice
func computePercentile(samples []float64, percentile float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	sorted := make([]float64, len(samples))
	copy(sorted, samples)
	sort.Float64s(sorted)

	pos := (float64(len(sorted)-1)) * percentile
	base := int(math.Floor(pos))
	rest := pos - float64(base)

	if base+1 < len(sorted) {
		return sorted[base] + rest*(sorted[base+1]-sorted[base])
	}
	return sorted[base]
}

// computeJitter computes the jitter (standard deviation) of samples
func computeJitter(samples []float64) float64 {
	if len(samples) < 2 {
		return 0
	}
	mean := 0.0
	for _, s := range samples {
		mean += s
	}
	mean /= float64(len(samples))

	variance := 0.0
	for _, s := range samples {
		variance += math.Pow(s-mean, 2)
	}
	variance /= float64(len(samples) - 1)

	return math.Sqrt(variance)
}

// computeMean computes the mean of samples
func computeMean(samples []float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	sum := 0.0
	for _, s := range samples {
		sum += s
	}
	return sum / float64(len(samples))
}

// floatPtr returns a pointer to a float64
func floatPtr(f float64) *float64 {
	return &f
}

// fetchNetworkInfo fetches network info from external services
func fetchNetworkInfo() NetworkInfo {
	client := &http.Client{Timeout: 5 * time.Second}

	// Try ifconfig.co first
	var info NetworkInfo
	resp, err := client.Get("https://ifconfig.co/json")
	if err == nil && resp.StatusCode == http.StatusOK {
		var data struct {
			ASN          interface{} `json:"asn"`
			ASNOrg       string      `json:"asn_org"`
			ASNDescription string    `json:"asn_description"`
			IPv4         string      `json:"ipv4"`
			IPv6         string      `json:"ipv6"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
			if data.ASN != nil {
				asn := fmt.Sprintf("%v", data.ASN)
				info.Asn = &asn
			}
			if data.ASNOrg != "" {
				info.AsOrg = &data.ASNOrg
			} else if data.ASNDescription != "" {
				info.AsOrg = &data.ASNDescription
			}
			if data.IPv4 != "" {
				info.ExternalIPv4 = &data.IPv4
			}
			if data.IPv6 != "" {
				info.ExternalIPv6 = &data.IPv6
			}
		}
		resp.Body.Close()
		return info
	}
	if resp != nil {
		resp.Body.Close()
	}

	// Fallback to ipify
	if info.ExternalIPv4 == nil {
		resp, err := client.Get("https://api.ipify.org?format=json")
		if err == nil && resp.StatusCode == http.StatusOK {
			var data struct {
				IP string `json:"ip"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&data); err == nil && data.IP != "" {
				info.ExternalIPv4 = &data.IP
			}
			resp.Body.Close()
		}
	}

	return info
}

// getLocalNetworkInfo gets local network information
func getLocalNetworkInfo() (localIpv4, localIpv6, interfaceName, networkName string) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil {
				continue
			}

			if ip.To4() != nil && localIpv4 == "" {
				localIpv4 = ip.String()
				interfaceName = iface.Name
				networkName = iface.Name
			}
			if ip.To16() != nil && ip.To4() == nil && localIpv6 == "" {
				localIpv6 = ip.String()
			}
		}
	}

	return
}

// fetchWithTimeout performs an HTTP request with a timeout
func fetchWithTimeout(client *http.Client, url string, timeout time.Duration) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	return client.Do(req)
}

// getExternalIP gets the external IP from Cloudflare trace
func getExternalIP(baseURL string, client *http.Client) (string, error) {
	url := baseURL + "/cdn-cgi/trace"
	resp, err := fetchWithTimeout(client, url, 5*time.Second)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var externalIP string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ip=") {
			externalIP = strings.TrimPrefix(line, "ip=")
			break
		}
	}
	return externalIP, nil
}

// getServerAndColo gets server and colo from Cloudflare trace
func getServerAndColo(baseURL string, client *http.Client) (server, colo string, err error) {
	url := baseURL + "/cdn-cgi/trace"
	resp, err := fetchWithTimeout(client, url, 5*time.Second)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "loc=") {
			server = strings.TrimPrefix(line, "loc=")
		}
		if strings.HasPrefix(line, "colo=") {
			colo = strings.TrimPrefix(line, "colo=")
		}
	}
	return
}

// getHostname extracts hostname from URL
func getHostname(baseURL string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}
	return u.Hostname()
}

// CloudflareSpeedtest implements the speedtest engine
type CloudflareSpeedtest struct {
	config config.ExporterConfig
	client *http.Client
}

// NewCloudflareSpeedtest creates a new speedtest engine
func NewCloudflareSpeedtest(cfg config.ExporterConfig) *CloudflareSpeedtest {
	return &CloudflareSpeedtest{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
			},
		},
	}
}

// RunDirectTest runs a direct speed test
func (c *CloudflareSpeedtest) RunDirectTest() (*model.RunResult, error) {
	measID := generateMeasID()

	// Get network info
	networkInfo := fetchNetworkInfo()
	localIpv4, localIpv6, interfaceName, networkName := getLocalNetworkInfo()

	// Get server and colo info
	server, colo, err := getServerAndColo(c.config.BaseURL, c.client)
	if err != nil {
		return nil, fmt.Errorf("failed to get server info: %w", err)
	}

	// Get external IP
	externalIP, err := getExternalIP(c.config.BaseURL, c.client)
	if err != nil {
		// Use networkInfo as fallback
		if networkInfo.ExternalIPv4 != nil {
			externalIP = *networkInfo.ExternalIPv4
		}
	}

	// Run idle latency test
	idleLatency := c.measureLatency(
		c.config.BaseURL+"/__latency",
		c.config.IdleLatencyDurationMs,
		c.config.ProbeIntervalMs,
	)

	// Run download test
	download := c.measureThroughput(
		c.config.BaseURL+"/__down",
		c.config.DownloadDurationMs,
		c.config.DownloadBytesPerReq,
		c.config.Concurrency,
	)

	// Run upload test
	upload := c.measureThroughput(
		c.config.BaseURL+"/__up",
		c.config.UploadDurationMs,
		c.config.UploadBytesPerReq,
		c.config.Concurrency,
	)

	// Measure loaded latency during download
	loadedLatencyDownload := c.measureLatency(
		c.config.BaseURL+"/__latency?phase=download",
		c.config.DownloadDurationMs,
		c.config.ProbeIntervalMs,
	)

	// Measure loaded latency during upload
	loadedLatencyUpload := c.measureLatency(
		c.config.BaseURL+"/__latency?phase=upload",
		c.config.UploadDurationMs,
		c.config.ProbeIntervalMs,
	)

	// DNS measurement
	var dns *model.DnsSummary
	if !c.config.SkipDiagnostics {
		dns = c.measureDns()
	}

	// TLS measurement
	var tls *model.TlsSummary
	if !c.config.SkipDiagnostics {
		tls = c.measureTls()
	}

	// Build result
	result := &model.RunResult{
		Version:              "1.0.0",
		TimestampUTC:         time.Now().UTC().Format(time.RFC3339),
		BaseURL:              c.config.BaseURL,
		MeasID:               measID,
		Server:               &server,
		Colo:                 &colo,
		IP:                   &externalIP,
		IdleLatency:          *idleLatency,
		Download:             *download,
		Upload:               *upload,
		LoadedLatencyDownload: *loadedLatencyDownload,
		LoadedLatencyUpload:   *loadedLatencyUpload,
		Dns:                  dns,
		Tls:                  tls,
	}

	// Set network info with config overrides
	if c.config.Asn != nil {
		result.Asn = c.config.Asn
	} else if networkInfo.Asn != nil {
		result.Asn = networkInfo.Asn
	}
	if c.config.AsOrg != nil {
		result.AsOrg = c.config.AsOrg
	} else if networkInfo.AsOrg != nil {
		result.AsOrg = networkInfo.AsOrg
	}
	if c.config.InterfaceName != nil {
		result.InterfaceName = c.config.InterfaceName
	} else if interfaceName != "" {
		result.InterfaceName = &interfaceName
	}
	if c.config.NetworkName != nil {
		result.NetworkName = c.config.NetworkName
	} else if networkName != "" {
		result.NetworkName = &networkName
	}
	if c.config.LocalIpv4 != nil {
		result.LocalIpv4 = c.config.LocalIpv4
	} else if localIpv4 != "" {
		result.LocalIpv4 = &localIpv4
	}
	if c.config.LocalIpv6 != nil {
		result.LocalIpv6 = c.config.LocalIpv6
	} else if localIpv6 != "" {
		result.LocalIpv6 = &localIpv6
	}
	if c.config.ExternalIpv4 != nil {
		result.ExternalIpv4 = c.config.ExternalIpv4
	} else if networkInfo.ExternalIPv4 != nil {
		result.ExternalIpv4 = networkInfo.ExternalIPv4
	} else if externalIP != "" {
		result.ExternalIpv4 = &externalIP
	}
	if c.config.ExternalIpv6 != nil {
		result.ExternalIpv6 = c.config.ExternalIpv6
	} else if networkInfo.ExternalIPv6 != nil {
		result.ExternalIpv6 = networkInfo.ExternalIPv6
	}

	return result, nil
}

// measureLatency measures latency to a URL
func (c *CloudflareSpeedtest) measureLatency(url string, durationMs time.Duration, intervalMs time.Duration) *model.LatencySummary {
	samples := []float64{}
	startTime := time.Now()
	endTime := startTime.Add(durationMs)

	for time.Now().Before(endTime) {
		start := time.Now()
		resp, err := fetchWithTimeout(c.client, url, c.config.ProbeTimeoutMs)
		if err != nil {
			samples = append(samples, -1) // Mark as loss
		} else {
			io.ReadAll(resp.Body)
			resp.Body.Close()
			elapsed := float64(time.Since(start).Milliseconds())
			samples = append(samples, elapsed)
		}

		// Sleep for the interval, but ensure we don't exceed the total duration
		remaining := endTime.Sub(time.Now())
		if remaining <= 0 {
			break
		}
		sleepTime := intervalMs
		if sleepTime > remaining {
			sleepTime = remaining
		}
		time.Sleep(sleepTime)
	}

	return c.computeLatencySummary(samples)
}

// measureThroughput measures throughput to a URL
func (c *CloudflareSpeedtest) measureThroughput(url string, durationMs time.Duration, bytesPerRequest int64, concurrency int) *model.ThroughputSummary {
	startTime := time.Now()
	endTime := startTime.Add(durationMs)

	var mu sync.Mutex
	bytesTotal := int64(0)
	speeds := []float64{}

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for time.Now().Before(endTime) {
				start := time.Now()
				resp, err := fetchWithTimeout(c.client, url, c.config.ProbeTimeoutMs)
				if err != nil {
					continue
				}
				data, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				elapsed := time.Since(start).Seconds()

				mu.Lock()
				bytesTotal += int64(len(data))
				if elapsed > 0 {
					bps := (float64(len(data)) * 8) / elapsed
					speeds = append(speeds, bps)
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	actualDuration := time.Since(startTime).Milliseconds()
	if actualDuration <= 0 {
		actualDuration = 1
	}

	totalMbps := (float64(bytesTotal) * 8) / (float64(actualDuration) * 1000)

	// Convert speeds from bps to Mbps
	mbpsSpeeds := make([]float64, len(speeds))
	for i, s := range speeds {
		mbpsSpeeds[i] = s / 1000000
	}

	return &model.ThroughputSummary{
		Bytes:        bytesTotal,
		DurationMs:   actualDuration,
		Mbps:         totalMbps,
		MeanMbps:     floatPtr(computeMedian(mbpsSpeeds)),
		MedianMbps:   floatPtr(computeMedian(mbpsSpeeds)),
		P25Mbps:      floatPtr(computePercentile(mbpsSpeeds, 0.25)),
		P75Mbps:      floatPtr(computePercentile(mbpsSpeeds, 0.75)),
	}
}

// measureDns measures DNS resolution time
func (c *CloudflareSpeedtest) measureDns() *model.DnsSummary {
	hostname := getHostname(c.config.BaseURL)
	start := time.Now()

	// Simple DNS resolution
	_, err := net.LookupHost(hostname)
	elapsed := time.Since(start).Milliseconds()

	// Count IPv4 and IPv6
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil
	}

	ipv4Count := 0
	ipv6Count := 0
	for _, ip := range ips {
		if ip.To4() != nil {
			ipv4Count++
		} else {
			ipv6Count++
		}
	}

	return &model.DnsSummary{
		Hostname:          hostname,
		ResolutionTimeMs: float64(elapsed),
		ResolvedIPs:       make([]string, len(ips)),
		Ipv4Count:        ipv4Count,
		Ipv6Count:        ipv6Count,
	}
}

// measureTls measures TLS handshake time
func (c *CloudflareSpeedtest) measureTls() *model.TlsSummary {
	start := time.Now()

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}
	client := &http.Client{Transport: transport, Timeout: 10 * time.Second}

	resp, err := client.Get(c.config.BaseURL)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	elapsed := time.Since(start).Milliseconds()

	// Get TLS info from connection state
	tlsState := resp.TLS
	var protocol, cipher string
	if tlsState != nil {
		protocol = tls.VersionName(tlsState.Version)
		cipher = tls.CipherSuiteName(tlsState.CipherSuite)
	}

	return &model.TlsSummary{
		HandshakeTimeMs: float64(elapsed),
		ProtocolVersion: &protocol,
		CipherSuite:     &cipher,
	}
}

// computeLatencySummary computes summary statistics for latency samples
func (c *CloudflareSpeedtest) computeLatencySummary(samples []float64) *model.LatencySummary {
	validSamples := []float64{}
	lossCount := 0

	for _, s := range samples {
		if s >= 0 {
			validSamples = append(validSamples, s)
		} else {
			lossCount++
		}
	}

	loss := 0.0
	if len(samples) > 0 {
		loss = float64(lossCount) / float64(len(samples))
	}

	if len(validSamples) == 0 {
		return &model.LatencySummary{
			Sent:     len(samples),
			Received: 0,
			Loss:     loss,
		}
	}

	sort.Float64s(validSamples)

	return &model.LatencySummary{
		Sent:       len(samples),
		Received:   len(validSamples),
		Loss:       loss,
		MinMs:      floatPtr(validSamples[0]),
		MeanMs:     floatPtr(computeMean(validSamples)),
		MedianMs:   floatPtr(computeMedian(validSamples)),
		P25Ms:      floatPtr(computePercentile(validSamples, 0.25)),
		P75Ms:      floatPtr(computePercentile(validSamples, 0.75)),
		MaxMs:      floatPtr(validSamples[len(validSamples)-1]),
		JitterMs:   floatPtr(computeJitter(validSamples)),
	}
}
