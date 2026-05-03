# Syntax: docker/dockerfile:1.4
# Ultra-minimal image using scratch

# Build stage: Compile with Bun on Alpine
FROM oven/bun:1-alpine AS builder

WORKDIR /app

COPY package.json bun.lock ./
RUN bun install --production
COPY src/ ./src/

# Compile to a standalone executable
RUN bun build --compile --outfile /app/speedtest-exporter src/index.ts

# Final stage: Scratch with only the binary and required libraries
FROM scratch

# Copy the compiled binary
COPY --from=builder /app/speedtest-exporter /speedtest-exporter

# Copy required musl libc libraries from Alpine (create dirs during copy)
COPY --from=builder /lib/ld-musl-aarch64.so.1 /lib/
COPY --from=builder /usr/lib/libstdc++.so.6 /usr/lib/
COPY --from=builder /usr/lib/libgcc_s.so.1 /usr/lib/

# Create non-root user (uid 1000)
USER 1000

EXPOSE 9537

ENTRYPOINT ["/speedtest-exporter"]
