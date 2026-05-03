import { describe, it, expect } from "bun:test";
import type {
  LatencySummary,
  ThroughputSummary,
  DnsSummary,
  TlsSummary,
  TracerouteHop,
  TracerouteSummary,
  RunResult,
} from "../src/model";

describe("model", () => {
  describe("LatencySummary", () => {
    it("should have all required fields", () => {
      const summary: LatencySummary = {
        sent: 10,
        received: 8,
        loss: 0.2,
      };
      expect(summary.sent).toBe(10);
      expect(summary.received).toBe(8);
      expect(summary.loss).toBe(0.2);
    });

    it("should have optional statistic fields", () => {
      const summary: LatencySummary = {
        sent: 10,
        received: 10,
        loss: 0,
        min_ms: 10,
        mean_ms: 15,
        median_ms: 15,
        p25_ms: 12,
        p75_ms: 18,
        max_ms: 20,
        jitter_ms: 2,
      };
      expect(summary.min_ms).toBe(10);
      expect(summary.max_ms).toBe(20);
      expect(summary.jitter_ms).toBe(2);
    });
  });

  describe("ThroughputSummary", () => {
    it("should have all required fields", () => {
      const summary: ThroughputSummary = {
        bytes: 1000000,
        duration_ms: 1000,
        mbps: 8,
      };
      expect(summary.bytes).toBe(1000000);
      expect(summary.duration_ms).toBe(1000);
      expect(summary.mbps).toBe(8);
    });

    it("should have optional statistic fields", () => {
      const summary: ThroughputSummary = {
        bytes: 1000000,
        duration_ms: 1000,
        mbps: 8,
        mean_mbps: 8.5,
        median_mbps: 8.2,
        p25_mbps: 7.8,
        p75_mbps: 9.2,
      };
      expect(summary.mean_mbps).toBe(8.5);
      expect(summary.p25_mbps).toBe(7.8);
    });
  });

  describe("DnsSummary", () => {
    it("should have all required fields", () => {
      const summary: DnsSummary = {
        hostname: "example.com",
        resolution_time_ms: 10,
        resolved_ips: ["1.2.3.4"],
        ipv4_count: 1,
        ipv6_count: 0,
      };
      expect(summary.hostname).toBe("example.com");
      expect(summary.resolution_time_ms).toBe(10);
      expect(summary.ipv4_count).toBe(1);
      expect(summary.ipv6_count).toBe(0);
    });

    it("should have optional dns_servers field", () => {
      const summary: DnsSummary = {
        hostname: "example.com",
        resolution_time_ms: 10,
        resolved_ips: ["1.2.3.4"],
        ipv4_count: 1,
        ipv6_count: 0,
        dns_servers: ["8.8.8.8", "8.8.4.4"],
      };
      expect(summary.dns_servers).toEqual(["8.8.8.8", "8.8.4.4"]);
    });
  });

  describe("TlsSummary", () => {
    it("should have all required fields", () => {
      const summary: TlsSummary = {
        handshake_time_ms: 50,
      };
      expect(summary.handshake_time_ms).toBe(50);
    });

    it("should have optional protocol and cipher fields", () => {
      const summary: TlsSummary = {
        handshake_time_ms: 50,
        protocol_version: "TLSv1.3",
        cipher_suite: "AES_256_GCM",
      };
      expect(summary.protocol_version).toBe("TLSv1.3");
      expect(summary.cipher_suite).toBe("AES_256_GCM");
    });
  });

  describe("TracerouteHop", () => {
    it("should have all required fields", () => {
      const hop: TracerouteHop = {
        hop_number: 1,
        rtt_ms: [10, 12, 11],
        timeout: false,
      };
      expect(hop.hop_number).toBe(1);
      expect(hop.rtt_ms).toEqual([10, 12, 11]);
      expect(hop.timeout).toBe(false);
    });

    it("should have optional ip and hostname fields", () => {
      const hop: TracerouteHop = {
        hop_number: 1,
        ip_address: "1.2.3.4",
        hostname: "router.example.com",
        rtt_ms: [10, 12, 11],
        timeout: false,
      };
      expect(hop.ip_address).toBe("1.2.3.4");
      expect(hop.hostname).toBe("router.example.com");
    });
  });

  describe("TracerouteSummary", () => {
    it("should have all required fields", () => {
      const summary: TracerouteSummary = {
        destination: "example.com",
        hops: [],
        completed: false,
      };
      expect(summary.destination).toBe("example.com");
      expect(summary.hops).toEqual([]);
      expect(summary.completed).toBe(false);
    });
  });

  describe("RunResult", () => {
    it("should have all required fields", () => {
      const result: RunResult = {
        base_url: "https://speed.cloudflare.com",
        meas_id: "test-123",
        idle_latency: { sent: 10, received: 10, loss: 0 },
        download: { bytes: 1000000, duration_ms: 1000, mbps: 8 },
        upload: { bytes: 500000, duration_ms: 1000, mbps: 4 },
        loaded_latency_download: { sent: 10, received: 10, loss: 0 },
        loaded_latency_upload: { sent: 10, received: 10, loss: 0 },
      };
      expect(result.base_url).toBe("https://speed.cloudflare.com");
      expect(result.meas_id).toBe("test-123");
      expect(result.download.mbps).toBe(8);
      expect(result.upload.mbps).toBe(4);
    });

    it("should have optional fields", () => {
      const result: RunResult = {
        version: "1.0.0",
        timestamp_utc: "2024-01-01T00:00:00Z",
        base_url: "https://speed.cloudflare.com",
        meas_id: "test-123",
        comments: "Test run",
        server: "FRA",
        colo: "FRA",
        ip: "1.2.3.4",
        asn: "12345",
        as_org: "Example Org",
        interface_name: "eth0",
        network_name: "Example Network",
        local_ipv4: "192.168.1.1",
        local_ipv6: "::1",
        external_ipv4: "1.2.3.4",
        external_ipv6: "::2",
        idle_latency: {
          sent: 10,
          received: 10,
          loss: 0,
          min_ms: 5,
          mean_ms: 6,
          median_ms: 6,
          max_ms: 7,
        },
        download: { bytes: 1000000, duration_ms: 1000, mbps: 8 },
        upload: { bytes: 500000, duration_ms: 1000, mbps: 4 },
        loaded_latency_download: { sent: 10, received: 10, loss: 0 },
        loaded_latency_upload: { sent: 10, received: 10, loss: 0 },
      };
      expect(result.version).toBe("1.0.0");
      expect(result.server).toBe("FRA");
      expect(result.asn).toBe("12345");
    });
  });
});
