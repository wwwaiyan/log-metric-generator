# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o simulator \
    ./cmd/simulator

# Runtime stage - Alpine (small, ~15MB)
FROM alpine:3.23 AS alpine

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/simulator .
COPY config.yaml .

EXPOSE 8080

ENV GIN_MODE=release

ENTRYPOINT ["./simulator"]
CMD ["-config", "config.yaml"]

# Runtime stage - Scratch (smallest, ~10MB, no shell)
FROM scratch AS scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/simulator .
COPY config.yaml .

ENTRYPOINT ["./simulator"]
CMD ["-config", "config.yaml"]
