# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o orchestrator main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests and timezone data
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/orchestrator .

# Copy templates directory
COPY --from=builder /app/templates ./templates

# Create necessary directories
RUN mkdir -p /tmp/orchestrator

# Set environment variables
ENV GIN_MODE=release
ENV ORCHESTRATOR_PORT=8080
ENV ORCHESTRATOR_HOST=0.0.0.0
ENV ORCHESTRATOR_WORKSPACE=/tmp/orchestrator
ENV ORCHESTRATOR_TEMPLATES=./templates

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./orchestrator"]