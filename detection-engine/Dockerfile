FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git for go mod
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o detection-engine ./cmd/main.go

# Final image
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Create directories
RUN mkdir -p /opt/rules_storage /var/log/detection-engine

# Copy binary
COPY --from=builder /app/detection-engine .

# Copy default rules
COPY rules/ /opt/rules_storage/

# Environment variables with defaults
ENV DB_HOST=localhost \
    DB_PORT=9000 \
    DB_USER=default \
    DB_PASSWORD= \
    DB_NAME=detection_db \
    LOG_LEVEL=info \
    LOG_TABLE=log_edr_v2 \
    MATCHING_TABLE=matching \
    ALERT_TABLE=alert \
    LOOKUP_TABLE=matching_lists \
    TIME_QUERY_WINDOW=2h \
    RULES_STORAGE_PATH=/opt/rules_storage \
    LOG_FILE_PATH=/var/log/detection-engine/engine.log \
    API_PORT=8000

EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8000/health || exit 1

CMD ["./detection-engine"]
