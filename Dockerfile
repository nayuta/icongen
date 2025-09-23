# Multi-stage build for icongen
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Final stage - minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS (if needed in future)
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh icongen

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/icongen /usr/local/bin/icongen

# Change ownership
RUN chown icongen:icongen /usr/local/bin/icongen

# Switch to non-root user
USER icongen

# Create directory for input/output
WORKDIR /workspace

# Set entrypoint
ENTRYPOINT ["icongen"]

# Default command shows help
CMD ["--help"]