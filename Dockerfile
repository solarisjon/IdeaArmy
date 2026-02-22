# ── Stage 1: Build the Go binary ──
FROM golang:1.24-bookworm AS builder

WORKDIR /build

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build the server-v2 binary (static, no CGO)
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/server-v2 ./cmd/server/main_v2.go

# ── Stage 2: Minimal runtime image ──
FROM debian:bookworm-slim

WORKDIR /app

# NetApp internal apt / cert configuration
COPY deployment/container-files/sources.list /etc/apt/sources.list
COPY deployment/container-files/ca-certificates.crt /tmp/netapp-ca.crt
COPY deployment/container-files/99insecure /etc/apt/apt.conf.d/99insecure
RUN cat /tmp/netapp-ca.crt >> /etc/ssl/certs/ca-certificates.crt && rm /tmp/netapp-ca.crt

RUN rm -rf /var/lib/apt/lists/* /etc/apt/sources.list.d/* \
    && update-ca-certificates 2>/dev/null || true

ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt
ENV REQUESTS_CA_BUNDLE=/etc/ssl/certs/ca-certificates.crt
ENV CURL_CA_BUNDLE=/etc/ssl/certs/ca-certificates.crt
ENV DEBIAN_FRONTEND=noninteractive

RUN apt update -y --no-install-recommends --allow-unauthenticated \
    && apt install -y --no-install-recommends ca-certificates \
    && apt clean \
    && rm -rf /var/lib/apt/lists/*

# Copy the compiled binary from builder
COPY --from=builder /build/server-v2 /app/server-v2

# Non-root user for security
RUN useradd -r -s /bin/false appuser
USER appuser

EXPOSE 8080

# Runtime env vars are injected via deploy.sh — never baked into the image
# LLMPROXY_KEY, LLM_MODEL, LLM_BASE_URL, PORT, HTTPS_PROXY
CMD ["/app/server-v2"]
