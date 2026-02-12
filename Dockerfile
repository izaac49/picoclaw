FROM golang:1.25.7 AS builder
WORKDIR /app
COPY . .
ENV GOPROXY=https://proxy.golang.org,direct
RUN go mod tidy
RUN go build -tags netgo -ldflags "-s -w" -o app ./cmd/picoclaw

FROM debian:bookworm-slim
WORKDIR /app
# ca-certificates is required for HTTPS requests to APIs
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/app .
COPY config.json /app/config.json
CMD ["./app", "gateway", "--config", "/app/config.json"]
