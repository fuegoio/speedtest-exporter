import { describe, it, expect } from "bun:test";
import { configSchema, loadConfig } from "../src/config";

describe("config", () => {
  describe("configSchema", () => {
    it("should validate default values", () => {
      const result = configSchema.parse({});
      expect(result.port).toBe(9537);
      expect(result.testIntervalMs).toBe(3600000);
      expect(result.baseUrl).toBe("https://speed.cloudflare.com");
      expect(result.downloadDurationMs).toBe(10000);
      expect(result.uploadDurationMs).toBe(10000);
      expect(result.idleLatencyDurationMs).toBe(2000);
      expect(result.concurrency).toBe(6);
      expect(result.downloadBytesPerReq).toBe(10000000);
      expect(result.uploadBytesPerReq).toBe(5000000);
      expect(result.probeIntervalMs).toBe(250);
      expect(result.probeTimeoutMs).toBe(800);
      expect(result.skipDiagnostics).toBe(false);
      expect(result.traceroute).toBe(false);
    });

    it("should coerce numeric strings to numbers", () => {
      const result = configSchema.parse({
        port: "8080",
        testIntervalMs: "60000",
        downloadDurationMs: "5000",
      });
      expect(result.port).toBe(8080);
      expect(result.testIntervalMs).toBe(60000);
      expect(result.downloadDurationMs).toBe(5000);
    });

    it("should validate URL format", () => {
      const result = configSchema.parse({
        baseUrl: "https://example.com",
      });
      expect(result.baseUrl).toBe("https://example.com");
    });

    it("should reject invalid URL", () => {
      expect(() =>
        configSchema.parse({
          baseUrl: "not-a-url",
        }),
      ).toThrow();
    });

    it("should reject negative numbers", () => {
      expect(() =>
        configSchema.parse({
          port: -1,
        }),
      ).toThrow();
    });

    it("should reject non-integer numbers", () => {
      expect(() =>
        configSchema.parse({
          port: 8080.5,
        }),
      ).toThrow();
    });

    it("should transform boolean strings to booleans", () => {
      const result = configSchema.parse({
        skipDiagnostics: "true",
        traceroute: "false",
      });
      expect(result.skipDiagnostics).toBe(true);
      expect(result.traceroute).toBe(false);
    });
  });

  describe("loadConfig", () => {
    it("should load from environment variables", () => {
      process.env.PORT = "8080";
      process.env.TEST_INTERVAL_MS = "60000";
      process.env.BASE_URL = "https://custom.example.com";

      const config = loadConfig();
      expect(config.port).toBe(8080);
      expect(config.testIntervalMs).toBe(60000);
      expect(config.baseUrl).toBe("https://custom.example.com");

      // Clean up
      delete process.env.PORT;
      delete process.env.TEST_INTERVAL_MS;
      delete process.env.BASE_URL;
    });

    it("should use defaults when environment variables are not set", () => {
      const config = loadConfig();
      expect(config.port).toBe(9537);
      expect(config.testIntervalMs).toBe(3600000);
      expect(config.baseUrl).toBe("https://speed.cloudflare.com");
    });

    it("should throw on invalid environment variables", () => {
      process.env.PORT = "invalid";
      expect(() => loadConfig()).toThrow();
      delete process.env.PORT;
    });
  });
});
