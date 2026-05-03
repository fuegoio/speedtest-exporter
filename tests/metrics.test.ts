import { describe, it, expect } from "bun:test";
import { register } from "../src/metrics";

describe("metrics", () => {
  describe("Registry", () => {
    it("should contain all registered metrics", async () => {
      const metrics = await register.metrics();
      expect(metrics.includes("speedtest_download_mbps")).toBe(true);
      expect(metrics.includes("speedtest_upload_mbps")).toBe(true);
      expect(metrics.includes("speedtest_idle_latency_ms")).toBe(true);
      expect(metrics.includes("speedtest_test_runs_total")).toBe(true);
      expect(metrics.includes("speedtest_test_errors_total")).toBe(true);
    });
  });
});
