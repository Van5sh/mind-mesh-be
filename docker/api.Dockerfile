# =========================
# Build stage
# =========================
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Download dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire Go backend
COPY . .

# Build API server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /app/server ./cmd/server


# =========================
# Runtime stage
# =========================
FROM alpine:3.22

WORKDIR /app

# CA certificates are required for HTTPS,
# OAuth providers, AWS, etc.
RUN apk add --no-cache ca-certificates

# Copy compiled Go application
COPY --from=builder /app/server /app/server

EXPOSE 8080

CMD ["/app/server"]