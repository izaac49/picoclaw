FROM golang:1.25.7

WORKDIR /app
COPY . .

RUN go build -tags netgo -ldflags "-s -w" -o app

CMD ["./app"]
