FROM golang:1.24-bullseye AS builder

WORKDIR /app

# Устанавливаем зависимости для confluent-kafka-go
RUN apt-get update && apt-get install -y git build-essential librdkafka-dev && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o processing-core ./cmd

# Runtime
FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /app/processing-core /app/processing-core

EXPOSE 9090 9100
USER nonroot:nonroot
ENTRYPOINT ["/app/processing-core"]
