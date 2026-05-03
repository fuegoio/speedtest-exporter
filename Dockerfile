# Syntax: docker/dockerfile:1.4
# Lightweight build for speedtest-exporter using Bun compile

# Build stage: Compile with Bun
FROM oven/bun:1-alpine AS builder

WORKDIR /app

# Copy package files first for better caching
COPY package.json bun.lock ./

# Install production dependencies only
RUN bun install --production

# Copy application code
COPY src/ ./src/

# Compile the application to a standalone executable
# --compile bundles the Bun runtime and all dependencies into a single binary
RUN bun build --compile --outfile /app/dist/index src/index.ts

# Final stage: Minimal runtime image
FROM alpine:3.20

# Install only curl for health check
RUN apk add --no-cache curl && rm -rf /var/cache/apk/*

WORKDIR /app

# Copy the compiled standalone executable from builder stage
COPY --from=builder /app/dist/index /app/speedtest-exporter

# Make executable
RUN chmod +x /app/speedtest-exporter

# Create non-root user for security
RUN adduser -D appuser
USER appuser

# Ensure binary has correct permissions
RUN chown appuser:appuser /app/speedtest-exporter

# Expose port
EXPOSE 9537

# Health check - curl the HTTP endpoint
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:9537/health || exit 1

# Run the compiled standalone binary
ENTRYPOINT ["/app/speedtest-exporter"]
