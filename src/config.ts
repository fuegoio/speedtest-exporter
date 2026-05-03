import { z } from "zod";

// Configuration schema using Zod
export const configSchema = z.object({
  // Server configuration
  port: z.coerce.number().int().positive().default(9537),

  // Test timing configuration
  testIntervalMs: z.coerce.number().int().positive().default(3600000), // 1 hour
  baseUrl: z.url().default("https://speed.cloudflare.com"),

  // Test duration configuration
  downloadDurationMs: z.coerce.number().int().positive().default(10000),
  uploadDurationMs: z.coerce.number().int().positive().default(10000),
  idleLatencyDurationMs: z.coerce.number().int().positive().default(2000),

  // Concurrency and request sizes
  concurrency: z.coerce.number().int().positive().default(6),
  downloadBytesPerReq: z.coerce.number().int().positive().default(10000000),
  uploadBytesPerReq: z.coerce.number().int().positive().default(5000000),

  // Probe configuration
  probeIntervalMs: z.coerce.number().int().positive().default(250),
  probeTimeoutMs: z.coerce.number().int().positive().default(800),

  // Feature flags
  skipDiagnostics: z
    .preprocess((v) => (typeof v === "string" ? v.toLowerCase() === "true" : v), z.boolean())
    .default(false),
  traceroute: z
    .preprocess((v) => (typeof v === "string" ? v.toLowerCase() === "true" : v), z.boolean())
    .default(false),
});

// ExporterConfig type derived from schema
export type ExporterConfig = z.infer<typeof configSchema>;

// Parse and validate environment variables
export function loadConfig(): ExporterConfig {
  return configSchema.parse({
    port: process.env.PORT,
    testIntervalMs: process.env.TEST_INTERVAL_MS,
    baseUrl: process.env.BASE_URL,
    downloadDurationMs: process.env.DOWNLOAD_DURATION_MS,
    uploadDurationMs: process.env.UPLOAD_DURATION_MS,
    idleLatencyDurationMs: process.env.IDLE_LATENCY_DURATION_MS,
    concurrency: process.env.CONCURRENCY,
    downloadBytesPerReq: process.env.DOWNLOAD_BYTES_PER_REQ,
    uploadBytesPerReq: process.env.UPLOAD_BYTES_PER_REQ,
    probeIntervalMs: process.env.PROBE_INTERVAL_MS,
    probeTimeoutMs: process.env.PROBE_TIMEOUT_MS,
    skipDiagnostics: process.env.SKIP_DIAGNOSTICS,
    traceroute: process.env.TRACEROUTE,
  });
}
