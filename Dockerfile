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

FROM alpine:3.23 AS alpine
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/simulator .
COPY config.yaml .
EXPOSE 8080
ENV GIN_MODE=release
ENTRYPOINT ["./simulator"]
CMD ["-config", "config.yaml"]