FROM golang:1.25.7 AS builder

WORKDIR /app
COPY . .

# Force Go proxy
ENV GOPROXY=https://proxy.golang.org,direct

RUN go mod tidy
RUN go build -tags netgo -ldflags "-s -w" -o app ./cmd/picoclaw

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/app .
CMD ["./app"]
