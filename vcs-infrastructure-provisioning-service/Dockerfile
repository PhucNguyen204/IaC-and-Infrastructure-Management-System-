# syntax = docker/dockerfile:1.7

FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o provisioning-service ./cmd/main.go

FROM alpine:latest

# Install dependencies: ca-certificates, docker-cli, curl, kubectl, k3d
RUN apk --no-cache add ca-certificates curl docker-cli

# Install kubectl
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/

# Install k3d
RUN curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | TAG=v5.6.0 sh

WORKDIR /app

COPY --from=builder /app/provisioning-service .

RUN mkdir -p /app/logs

EXPOSE 8083 50051

CMD ["./provisioning-service"]

