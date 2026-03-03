# ── Build stage ────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /shipping-service ./cmd/server

# ── Run stage ─────────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /shipping-service .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/.env.example .env.example

EXPOSE 8080

ENTRYPOINT ["./shipping-service"]
