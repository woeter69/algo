# Multi-stage Dockerfile for integrated Python + Go deployment
FROM golang:1.22-alpine AS go-builder

# Build Go WebSocket server
WORKDIR /app

# Initialize Go module and copy dependencies
COPY app/src/go.mod ./go.mod

# Copy Go source code
COPY app/src/*.go ./

# Download dependencies and update go.sum
RUN go mod tidy

# Build the Go binary
RUN go build -o websocket-server .

# Final stage with Python + Go + Nginx
FROM python:3.11-slim

# Set environment variables to avoid debconf warnings
ENV DEBIAN_FRONTEND=noninteractive
ENV DEBCONF_NONINTERACTIVE_SEEN=true

# Install system dependencies
RUN apt-get update && apt-get install -y \
    supervisor \
    nginx \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy Python requirements and install
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy Python application
COPY app/ ./app/

# Copy Go binary from builder stage
COPY --from=go-builder /app/websocket-server ./app/src/
RUN chmod +x ./app/src/websocket-server

# Copy nginx configuration
COPY deploy/nginx.conf /etc/nginx/nginx.conf

# Create supervisor configuration
RUN mkdir -p /var/log/supervisor
COPY deploy/supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# Expose port (Render will map this to 443/80)
EXPOSE 10000

# Start all services with supervisor
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
