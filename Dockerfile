# Syntax: docker/dockerfile:1.4
# Multi-stage build for minimal Go image

# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o speedtest-exporter ./cmd/speedtest-exporter/

# Final stage: Scratch with CA certificates for HTTPS
FROM scratch

# Copy CA certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the compiled binary
COPY --from=builder /app/speedtest-exporter /speedtest-exporter

# Create non-root user (uid 1000)
USER 1000

EXPOSE 9537

ENTRYPOINT ["/speedtest-exporter"]
