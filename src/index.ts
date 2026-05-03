import { serve } from "bun";
import { loadConfig } from "./config";
import { CloudflareSpeedtest } from "./engine";
import { register, testErrorsTotal, testRunsTotal } from "./metrics";
import { updateMetrics } from "./metricsUpdater";

// Main exporter class
export class SpeedtestExporter {
  private config: ReturnType<typeof loadConfig>;
  private speedtest: CloudflareSpeedtest;
  private running: boolean = false;
  private timer: ReturnType<typeof setInterval> | null = null;

  constructor() {
    this.config = loadConfig();
    this.speedtest = new CloudflareSpeedtest(this.config);
  }

  async start() {
    console.log(`Starting Speedtest Exporter on port ${this.config.port}`);
    console.log(`Test interval: ${this.config.testIntervalMs}ms`);

    // Start HTTP server for Prometheus metrics
    serve({
      port: this.config.port,
      fetch: async (req) => {
        const url = new URL(req.url);

        if (url.pathname === "/metrics") {
          const metrics = await register.metrics();
          return new Response(metrics, {
            headers: { "Content-Type": register.contentType },
          });
        }

        if (url.pathname === "/health") {
          return new Response("OK", { status: 200 });
        }

        if (url.pathname === "/run") {
          // Manually trigger a test
          try {
            const result = await this.speedtest.runDirectTest();
            updateMetrics(result);
            testRunsTotal.inc({ status: "success" });
            return new Response(JSON.stringify(result, null, 2), {
              headers: { "Content-Type": "application/json" },
            });
          } catch (error) {
            testErrorsTotal.inc({ error_type: String(error) });
            testRunsTotal.inc({ status: "failed" });
            return new Response(JSON.stringify({ error: String(error) }), {
              status: 500,
              headers: { "Content-Type": "application/json" },
            });
          }
        }

        return new Response(
          "Speedtest Exporter\n\nEndpoints:\n/metrics - Prometheus metrics\n/health - Health check\n/run - Run test manually",
          {
            headers: { "Content-Type": "text/plain" },
          },
        );
      },
    });

    console.log(`Server started on http://localhost:${this.config.port}`);

    // Start periodic tests
    this.startPeriodicTests();
  }

  startPeriodicTests() {
    if (this.running) return;
    this.running = true;

    console.log(`Starting periodic speed tests every ${this.config.testIntervalMs}ms`);

    // Run initial test immediately
    this.runTest();

    // Then run on interval
    this.timer = setInterval(() => {
      this.runTest();
    }, this.config.testIntervalMs);
  }

  stopPeriodicTests() {
    if (!this.running) return;
    this.running = false;
    if (this.timer) {
      clearInterval(this.timer);
      this.timer = null;
    }
    console.log("Stopped periodic speed tests");
  }

  async runTest() {
    try {
      console.log(`Running speed test at ${new Date().toISOString()}`);
      const result = await this.speedtest.runDirectTest();
      updateMetrics(result);
      console.log(
        `Test completed: Download=${result.download.mbps.toFixed(2)} Mbps, Upload=${result.upload.mbps.toFixed(2)} Mbps`,
      );
    } catch (error) {
      console.error(`Test failed: ${error}`);
      testErrorsTotal.inc({ error_type: String(error) });
      testRunsTotal.inc({ status: "failed" });
    }
  }

  async shutdown() {
    this.stopPeriodicTests();
    console.log("Exporter shutdown");
  }
}

// Main entry point
async function main() {
  const exporter = new SpeedtestExporter();

  // Handle graceful shutdown
  process.on("SIGINT", async () => {
    await exporter.shutdown();
    process.exit(0);
  });

  process.on("SIGTERM", async () => {
    await exporter.shutdown();
    process.exit(0);
  });

  await exporter.start();
}

main().catch((error) => {
  console.error("Fatal error:", error);
  process.exit(1);
});

export default SpeedtestExporter;
