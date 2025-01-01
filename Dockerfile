FROM golang:1.24-rc-bullseye AS builder
WORKDIR /app
COPY . .
RUN go build -o mqttcli ./cmd/mqttcli

FROM debian:bullseye-slim
WORKDIR /app
COPY --from=builder /app/mqttcli /app/
ENTRYPOINT ["./mqttcli"]
