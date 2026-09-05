# syntax=docker/dockerfile:1

# ---- Build stage ---------------------------------------------------------
FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/server ./cmd/server

# ---- Runtime stage -------------------------------------------------------
FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S app \
    && adduser -S app -G app

WORKDIR /app

COPY --from=build /out/server /app/server
# Migrations must be available at runtime for the `migrate` subcommand.
COPY migrations /app/migrations

RUN mkdir -p /app/storage && chown -R app:app /app

USER app
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/health/live || exit 1

ENTRYPOINT ["/app/server"]