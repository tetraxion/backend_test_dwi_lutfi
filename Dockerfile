# ── Stage 1: Build ──────────────────────────────────────────────────────────
FROM golang:1.21-alpine AS builder

# Install git (required by go mod download for some modules).
RUN apk add --no-cache git

WORKDIR /app

# Cache dependency downloads as a separate layer.
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a statically-linked binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o task-tracker .

# ── Stage 2: Runtime ─────────────────────────────────────────────────────────
FROM alpine:latest

# Add CA certificates so the service can make HTTPS calls if needed.
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Pull only the compiled binary from the builder stage.
COPY --from=builder /app/task-tracker .

# Expose the application port.
EXPOSE 8080

# Run as a non-root user for security.
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

ENTRYPOINT ["./task-tracker"]
