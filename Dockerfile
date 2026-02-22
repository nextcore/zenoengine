# Stage 1: Builder
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
# -ldflags="-w -s" to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o zeno cmd/zeno/zeno.go

# Stage 2: Runner
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies (ca-certificates for HTTPS)
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/zeno .

# Copy application assets/scripts required for runtime
COPY --from=builder /app/public ./public
COPY --from=builder /app/views ./views
COPY --from=builder /app/src ./src
COPY --from=builder /app/database ./database

# Expose port
EXPOSE 3000

# Set environment variables
ENV APP_ENV=production
ENV APP_PORT=:3000

# Run the application
CMD ["./zeno"]
