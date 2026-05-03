import { describe, it, expect, beforeEach } from "bun:test";
import { CloudflareSpeedtest } from "../src/engine";
import type { ExporterConfig } from "../src/config";

describe("CloudflareSpeedtest", () => {
  let config: ExporterConfig;
  let engine: CloudflareSpeedtest;

  beforeEach(() => {
    config = {
      port: 9537,
      testIntervalMs: 300000,
      baseUrl: "https://speed.cloudflare.com",
      downloadDurationMs: 1000,
      uploadDurationMs: 1000,
      idleLatencyDurationMs: 500,
      concurrency: 2,
      downloadBytesPerReq: 1000000,
      uploadBytesPerReq: 500000,
      probeIntervalMs: 250,
      probeTimeoutMs: 800,
      skipDiagnostics: true,
      traceroute: false,
    };
    engine = new CloudflareSpeedtest(config);
  });

  describe("computeMedian", () => {
    it("should compute median of odd-length array", () => {
      const samples = [1, 2, 3, 4, 5];
      // @ts-expect-error - private method
      const result = engine.computeMedian(samples);
      expect(result).toBe(3);
    });

    it("should compute median of even-length array", () => {
      const samples = [1, 2, 3, 4];
      // @ts-expect-error - private method
      const result = engine.computeMedian(samples);
      expect(result).toBe(2.5);
    });

    it("should handle single element", () => {
      const samples = [5];
      // @ts-expect-error - private method
      const result = engine.computeMedian(samples);
      expect(result).toBe(5);
    });
  });

  describe("computePercentile", () => {
    it("should compute 25th percentile", () => {
      const samples = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];
      // @ts-expect-error - private method
      const result = engine.computePercentile(samples, 0.25);
      expect(result).toBeCloseTo(3.25);
    });

    it("should compute 50th percentile (median)", () => {
      const samples = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];
      // @ts-expect-error - private method
      const result = engine.computePercentile(samples, 0.5);
      expect(result).toBe(5.5);
    });

    it("should compute 75th percentile", () => {
      const samples = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];
      // @ts-expect-error - private method
      const result = engine.computePercentile(samples, 0.75);
      expect(result).toBeCloseTo(7.75);
    });
  });

  describe("computeJitter", () => {
    it("should return 0 for single sample", () => {
      const samples = [10];
      // @ts-expect-error - private method
      const result = engine.computeJitter(samples);
      expect(result).toBe(0);
    });

    it("should compute jitter for multiple samples", () => {
      const samples = [10, 12, 8, 15, 11];
      // @ts-expect-error - private method
      const result = engine.computeJitter(samples);
      // Jitter is standard deviation
      expect(result).toBeGreaterThan(0);
    });
  });

  describe("computeLatencySummary", () => {
    it("should handle empty samples", () => {
      const samples: number[] = [];
      // @ts-expect-error - private method
      const result = engine.computeLatencySummary(samples);
      expect(result.sent).toBe(0);
      expect(result.received).toBe(0);
      expect(result.loss).toBe(0);
    });

    it("should handle samples with losses", () => {
      const samples = [10, 15, -1, 20, -1]; // 2 losses out of 5
      // @ts-expect-error - private method
      const result = engine.computeLatencySummary(samples);
      expect(result.sent).toBe(5);
      expect(result.received).toBe(3);
      expect(result.loss).toBeCloseTo(0.4);
    });

    it("should compute all statistics", () => {
      const samples = [10, 15, 20, 25, 30];
      // @ts-expect-error - private method
      const result = engine.computeLatencySummary(samples);
      expect(result.sent).toBe(5);
      expect(result.received).toBe(5);
      expect(result.loss).toBe(0);
      expect(result.min_ms).toBe(10);
      expect(result.max_ms).toBe(30);
      expect(result.mean_ms).toBe(20);
      expect(result.median_ms).toBe(20);
    });
  });
});
