# Syntax: docker/dockerfile:1.4
# Lightweight multi-stage build for speedtest-exporter

# Stage 1: Build with Bun alpine image
FROM oven/bun:1-alpine AS builder

WORKDIR /app

# Copy package files first for better caching
COPY package.json bun.lockb ./

# Install production dependencies only
RUN bun install --production

# Copy application code
COPY src/ ./src/

# Stage 2: Ultra-lightweight runtime image
FROM alpine:3.20

# Install minimal dependencies
RUN apk add --no-cache \
    curl \
    dumb-init \
    libgcc \
    && rm -rf /var/cache/apk/*

WORKDIR /app

# Install Bun (static binary, ~50MB)
RUN curl -fsSL https://bun.sh/install | bash -s "bun-linux-x64" > /dev/null 2>&1 \
    && mv /root/.bun/bin/bun /usr/local/bin/bun \
    && rm -rf /root/.bun

# Copy from builder stage
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/src/ ./src/
COPY --from=builder /app/package.json ./
COPY --from=builder /app/bun.lockb ./

# Create non-root user for security
RUN adduser -D appuser
USER appuser

# Ensure node_modules has correct permissions
RUN chown -R appuser:appuser /app

# Expose port
EXPOSE 9537

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:9537/health || exit 1

# Run the application with dumb-init for proper signal handling
ENTRYPOINT ["dumb-init", "--"]
CMD ["bun", "run", "--production", "src/index.ts"]
