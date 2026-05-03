# Syntax: docker/dockerfile:1.4
FROM oven/bun:1-debian

WORKDIR /app

# Copy package files first for better caching
COPY package.json bun.lock ./

# Install production dependencies only
RUN bun install --production

# Copy application code
COPY src/ ./src/

# Create non-root user
RUN useradd -r appuser && chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 9537

# Health check - curl the HTTP endpoint
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:9537/health || exit 1

# Run with Bun
CMD ["bun", "run", "src/index.ts"]
