FROM golang:1.24-bullseye AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o antifraud ./cmd

# Runtime
FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /app/antifraud /app/antifraud

EXPOSE 9090 9100
USER nonroot:nonroot
ENTRYPOINT ["/app/antifraud"]
