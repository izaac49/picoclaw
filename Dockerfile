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
COPY --from=builder /app/app .
CMD ["./app"]
