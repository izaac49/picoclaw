# Build stage
FROM golang:1.25.7 AS builder

WORKDIR /app

# Force Go to use the public proxy
ENV GOPROXY=https://proxy.golang.org,direct

# Copy everything
COPY . .

# Clean up and fetch dependencies
RUN go mod tidy

# Build binary from cmd/picoclaw
RUN go build -tags netgo -ldflags "-s -w" -o app ./cmd/picoclaw

# Final stage: minimal runtime image
FROM debian:bookworm-slim
WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/app .
COPY --from=builder /app/config.json /app/config.json

# Use standard gateway port
EXPOSE 18790

CMD ["./app", "gateway", "--config", "/app/config.json"]
