# Pure Go Dockerfile for AlumniGo
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Copy and download dependencies first (layer caching)
COPY app/src/go.mod app/src/go.sum ./
RUN go mod download

# Copy Go source files
COPY app/src/*.go ./

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o algo-server .

# Final minimal image
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy the binary
COPY --from=builder /build/algo-server ./algo-server

# Copy static assets and templates
COPY app/static ./app/static
COPY app/templates ./app/templates

EXPOSE 8080

CMD ["./algo-server"]
