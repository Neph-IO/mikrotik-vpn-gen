# === STAGE 1: Build the Go binary ===
FROM golang:1.24.2 AS builder

WORKDIR /app

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o mikrotik-vpn-gen ./cmd/mikrotik-vpn-gen

# === STAGE 2: Minimal runtime ===
FROM alpine:3.20

# Install SSH client for SFTP support
RUN apk add --no-cache openssh-client

# Create app directory
WORKDIR /app

# Copy the binary and necessary files
COPY --from=builder /app/mikrotik-vpn-gen .
COPY template/ template/
COPY config.exemple.yaml config.yaml
RUN mkdir -p output

# Expose the API port
EXPOSE 8081

# Start the app
CMD ["./mikrotik-vpn-gen"]
