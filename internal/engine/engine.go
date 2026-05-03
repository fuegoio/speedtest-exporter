package engine

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/fuegoio/speedtest-exporter/internal/config"
	"github.com/fuegoio/speedtest-exporter/internal/metrics"
	"github.com/fuegoio/speedtest-exporter/internal/model"
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
	mid := len(samples) / 2
	if len(samples)%2 == 0 {
		return (samples[mid-1] + samples[mid]) / 2
	}
	return samples[mid]
}

// computePercentile computes the percentile of a sorted slice
func computePercentile(samples []float64, percentile float64) float64 {
	if len(samples) == 0 {
		return 0
	}
	pos := (float64(len(samples)-1)) * percentile
	base := int(math.Floor(pos))
	rest := pos - float64(base)
	if base+1 < len(samples) {
		return samples[base] + rest*(samples[base+1]-samples[base])
	}
	return samples[base]
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

// floatPtr returns a pointer to a float64
func floatPtr(f float64) *float64 {
	return &f
}

// derefString returns the value or default if nil
func derefString(s *string, defaultValue string) string {
	if s != nil {
		return *s
	}
	return defaultValue
}

// derefFloat64 returns the value or default if nil
func derefFloat64(f *float64, defaultValue float64) float64 {
	if f != nil {
		return *f
	}
	return defaultValue
}

// fetchNetworkInfo fetches network info from Cloudflare trace
func fetchNetworkInfo() NetworkInfo {
	return NetworkInfo{}
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

// metaResponse is the JSON response from https://speed.cloudflare.com/meta
type metaResponse struct {
	ClientIp        string `json:"clientIp"`
	Asn             int    `json:"asn"`
	AsOrganization  string `json:"asOrganization"`
	Country         string `json:"country"`
	City            string `json:"city"`
	Region          string `json:"region"`
	PostalCode      string `json:"postalCode"`
	Latitude        string `json:"latitude"`
	Longitude       string `json:"longitude"`
	Colo            struct {
		IATA string `json:"iata"`
	} `json:"colo"`
}

// getMetaInfo fetches network and geo info from https://speed.cloudflare.com/meta
func getMetaInfo(baseURL string, client *http.Client) (externalIP, colo, asn, asOrg, country, city, region, postalCode, latitude, longitude string, err error) {
	metaURL := baseURL + "/meta"
	resp, err := fetchWithTimeout(client, metaURL, 5*time.Second)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var meta metaResponse
	if err = json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return
	}

	externalIP = meta.ClientIp
	colo = meta.Colo.IATA
	asn = fmt.Sprintf("AS%d", meta.Asn)
	asOrg = meta.AsOrganization
	country = meta.Country
	city = meta.City
	region = meta.Region
	postalCode = meta.PostalCode
	latitude = meta.Latitude
	longitude = meta.Longitude
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

// customTransport wraps http.Transport to add browser-like headers
type customTransport struct {
	transport http.RoundTripper
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add browser-like headers to avoid rate limiting
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", "https://speed.cloudflare.com/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	return t.transport.RoundTrip(req)
}

// NewCloudflareSpeedtest creates a new speedtest engine
func NewCloudflareSpeedtest(cfg config.ExporterConfig) *CloudflareSpeedtest {
	return &CloudflareSpeedtest{
		config: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &customTransport{
				transport: &http.Transport{
					MaxIdleConns:        100,
					MaxIdleConnsPerHost: 100,
				},
			},
		},
	}
}

// runSequentialTests runs sequential tests with specific sizes and iterations
func (c *CloudflareSpeedtest) runSequentialTests(
	testType string,
	sizes []struct {
		size       int64
		iterations int
	},
	server, colo, asn, asOrg, interfaceName, networkName, ipVersion,
	country, city, region, postalCode, latitude, longitude string,
) []model.ThroughputSummary {
	var results []model.ThroughputSummary

	for _, test := range sizes {
		sizeLabel := fmt.Sprintf("%d", test.size)
		for i := 0; i < test.iterations; i++ {
			var url string
			if testType == "download" {
				url = fmt.Sprintf("%s/__down?bytes=%d", c.config.BaseURL, test.size)
			} else {
				url = fmt.Sprintf("%s/__up?bytes=%d", c.config.BaseURL, test.size)
			}

			result := c.measureSingleFetch(url)

			// Update metrics for this specific test
			if testType == "download" {
				metrics.DownloadMbps.WithLabelValues(server, colo, asn, asOrg, interfaceName, networkName, ipVersion, country, city, region, postalCode, latitude, longitude, sizeLabel).Observe(result.Mbps)
				metrics.DownloadDurationMs.WithLabelValues(server, colo, asn, asOrg, interfaceName, networkName, ipVersion, country, city, region, postalCode, latitude, longitude, sizeLabel).Observe(float64(result.DurationMs))
			} else {
				metrics.UploadMbps.WithLabelValues(server, colo, asn, asOrg, interfaceName, networkName, ipVersion, country, city, region, postalCode, latitude, longitude, sizeLabel).Observe(result.Mbps)
				metrics.UploadDurationMs.WithLabelValues(server, colo, asn, asOrg, interfaceName, networkName, ipVersion, country, city, region, postalCode, latitude, longitude, sizeLabel).Observe(float64(result.DurationMs))
			}

			log.Printf("[Speedtest] %s %s: %.2f Mbps (%d bytes in %v)",
				testType, sizeLabel, result.Mbps, result.Bytes, time.Duration(result.DurationMs)*time.Millisecond)

			results = append(results, *result)
		}
	}

	return results
}

// measureIdleLatency runs idle latency tests
func (c *CloudflareSpeedtest) measureIdleLatency(baseURL, server, colo string) *model.LatencySummary {
	latencyURL := fmt.Sprintf("%s/__latency", baseURL)
	log.Printf("[Speedtest] Running idle latency tests...")
	return c.measureLatency(latencyURL, 10*time.Second, 100*time.Millisecond)
}

// measureLoadedLatencyDownload runs latency tests while download is running in background
func (c *CloudflareSpeedtest) measureLoadedLatencyDownload(baseURL, server, colo string, durationMs time.Duration) *model.LatencySummary {
	latencyURL := fmt.Sprintf("%s/__latency", baseURL)
	downloadURL := fmt.Sprintf("%s/__down?bytes=%d", baseURL, 100*1024) // 100kB
	log.Printf("[Speedtest] Running loaded latency tests during download...")
	
	// Start background download
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.measureThroughput(downloadURL, durationMs)
	}()

	// Run latency tests
	result := c.measureLatency(latencyURL, durationMs, 100*time.Millisecond)
	wg.Wait()
	return result
}

// measureLoadedLatencyUpload runs latency tests while upload is running in background
func (c *CloudflareSpeedtest) measureLoadedLatencyUpload(baseURL, server, colo string, durationMs time.Duration) *model.LatencySummary {
	latencyURL := fmt.Sprintf("%s/__latency", baseURL)
	uploadURL := fmt.Sprintf("%s/__up?bytes=%d", baseURL, 100*1024) // 100kB
	log.Printf("[Speedtest] Running loaded latency tests during upload...")
	
	// Start background upload
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.measureThroughput(uploadURL, durationMs)
	}()

	// Run latency tests
	result := c.measureLatency(latencyURL, durationMs, 100*time.Millisecond)
	wg.Wait()
	return result
}

// RunDirectTest runs a direct speed test with sequential downloads and uploads
func (c *CloudflareSpeedtest) RunDirectTest() (*model.RunResult, error) {
	measID := generateMeasID()
	log.Printf("[Speedtest] Starting test with measurement ID: %s", measID)

	// Get local network info
	log.Printf("[Speedtest] Fetching network information...")
	localIpv4, localIpv6, interfaceName, networkName := getLocalNetworkInfo()
	log.Printf("[Speedtest] Local IPv4: %s, IPv6: %s, Interface: %s, Network: %s",
		localIpv4, localIpv6, interfaceName, networkName)

	// Get meta info (colo, external IP, ASN, ASN Org, geo) from Cloudflare /meta
	log.Printf("[Speedtest] Getting server and network information from Cloudflare /meta...")
	externalIP, colo, asn, asOrg, country, city, region, postalCode, latitude, longitude, err := getMetaInfo(c.config.BaseURL, c.client)
	if err != nil {
		return nil, fmt.Errorf("failed to get meta info: %w", err)
	}
	server := country
	log.Printf("[Speedtest] Colo: %s, External IP: %s, Country: %s, City: %s", colo, externalIP, country, city)
	if asn != "" {
		log.Printf("[Speedtest] ASN: %s, ASN Org: %s", asn, asOrg)
	}

	// Run idle latency tests (10 pings)
	idleLatency := c.measureIdleLatency(c.config.BaseURL, server, colo)

	// Get IP version for labels
	ipVersion := "both"
	if localIpv4 != "" && localIpv6 == "" {
		ipVersion = "ipv4"
	} else if localIpv4 == "" && localIpv6 != "" {
		ipVersion = "ipv6"
	}

	// Run sequential download tests
	log.Printf("[Speedtest] Running sequential download tests...")
	downloadResults := c.runSequentialTests(
		"download",
		[]struct {
			size       int64
			iterations int
		}{
			{100 * 1024, 10},    // 100kB, 10 times
			{1 * 1024 * 1024, 8}, // 1MB, 8 times
			{10 * 1024 * 1024, 6}, // 10MB, 6 times
			{25 * 1024 * 1024, 4}, // 25MB, 4 times
			{100 * 1024 * 1024, 3}, // 100MB, 3 times
			{250 * 1024 * 1024, 2}, // 250MB, 2 times
		},
		server, colo, asn, asOrg, interfaceName, networkName, ipVersion,
		country, city, region, postalCode, latitude, longitude,
	)

	// Run loaded latency tests during download (10 pings)
	loadedLatencyDownload := c.measureLoadedLatencyDownload(c.config.BaseURL, server, colo, 10*time.Second)

	// Run sequential upload tests
	log.Printf("[Speedtest] Running sequential upload tests...")
	uploadResults := c.runSequentialTests(
		"upload",
		[]struct {
			size       int64
			iterations int
		}{
			{100 * 1024, 8},    // 100kB, 8 times
			{1 * 1024 * 1024, 6}, // 1MB, 6 times
			{10 * 1024 * 1024, 4}, // 10MB, 4 times
			{25 * 1024 * 1024, 4}, // 25MB, 4 times
			{50 * 1024 * 1024, 3}, // 50MB, 3 times
		},
		server, colo, asn, asOrg, interfaceName, networkName, ipVersion,
		country, city, region, postalCode, latitude, longitude,
	)

	// Run loaded latency tests during upload (10 pings)
	loadedLatencyUpload := c.measureLoadedLatencyUpload(c.config.BaseURL, server, colo, 10*time.Second)

	// Use the last results for backward compatibility
	var download *model.ThroughputSummary
	var upload *model.ThroughputSummary
	if len(downloadResults) > 0 {
		download = &downloadResults[len(downloadResults)-1]
	} else {
		download = &model.ThroughputSummary{}
	}
	if len(uploadResults) > 0 {
		upload = &uploadResults[len(uploadResults)-1]
	} else {
		upload = &model.ThroughputSummary{}
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
		Country:              &country,
		City:                 &city,
		Region:               &region,
		PostalCode:           &postalCode,
		Latitude:             &latitude,
		Longitude:            &longitude,
	}

	log.Printf("[Speedtest] Test completed successfully with measurement ID: %s", measID)

	// Set network info with config overrides
	if c.config.Asn != nil {
		result.Asn = c.config.Asn
	} else if asn != "" {
		result.Asn = &asn
	}
	if c.config.AsOrg != nil {
		result.AsOrg = c.config.AsOrg
	} else if asOrg != "" {
		result.AsOrg = &asOrg
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
	} else if externalIP != "" {
		result.ExternalIpv4 = &externalIP
	}
	if c.config.ExternalIpv6 != nil {
		result.ExternalIpv6 = c.config.ExternalIpv6
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

// measureSingleFetch performs a single HTTP fetch and returns a ThroughputSummary.
func (c *CloudflareSpeedtest) measureSingleFetch(url string) *model.ThroughputSummary {
	start := time.Now()
	resp, err := fetchWithTimeout(c.client, url, 60*time.Second)
	if err != nil {
		return &model.ThroughputSummary{}
	}
	data, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	dur := time.Since(start)
	elapsed := dur.Seconds()

	bytes := int64(len(data))
	mbps := 0.0
	if elapsed > 0 {
		mbps = (float64(bytes) * 8) / (elapsed * 1_000_000)
	}

	return &model.ThroughputSummary{
		Bytes:      bytes,
		DurationMs: dur.Milliseconds(),
		Mbps:       mbps,
	}
}

// measureThroughput fetches url in a loop for the given duration (used for loaded latency background load).
func (c *CloudflareSpeedtest) measureThroughput(url string, duration time.Duration) {
	endTime := time.Now().Add(duration)
	for time.Now().Before(endTime) {
		resp, err := fetchWithTimeout(c.client, url, duration)
		if err != nil {
			continue
		}
		io.ReadAll(resp.Body)
		resp.Body.Close()
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
