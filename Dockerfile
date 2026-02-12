# Build stage
FROM golang:1.25.7 AS builder

WORKDIR /app

# Copy go.mod and go.sum first for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary from cmd/picoclaw
RUN go build -tags netgo -ldflags "-s -w" -o app ./cmd/picoclaw

# Final stage: minimal runtime image
FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/app .
CMD ["./app"]
