# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bridgr ./cmd/bridgr

# Final stage
FROM gcr.io/distroless/static-debian11

# Copy the binary from builder
COPY --from=builder /app/bridgr /usr/local/bin/bridgr


# Set working directory
WORKDIR /etc/bridgr

# Expose port
EXPOSE 8080

# Set non-root user
USER nonroot:nonroot

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
ENTRYPOINT ["/usr/local/bin/bridgr"] 