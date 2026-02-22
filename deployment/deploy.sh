#!/bin/bash
#
# Podman deployment script for IdeaArmy War Room Server (server-v2)
# Deploys the web server in a Podman pod on Debian Linux or macOS.
#
# Each user provides their own LLM token via the web UI.
# No API keys need to be set at deployment time.
#
# Usage: ./deploy.sh <command>
# Run:   ./deploy.sh help
#

set -e

# ============== Path Resolution ==============
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

# ============== Configuration ==============
POD_NAME="ideaarmy-pod"
APP_NAME="ideaarmy-warroom"
TAG="${TAG:-dev}"

APP_PORT="${APP_PORT:-8080}"

# Container / image names
APP_IMAGE="localhost/${APP_NAME}:${TAG}"
APP_CONTAINER="${APP_NAME}-${TAG}"

# LLM configuration — defaults for server; each user provides their own token via the web UI
# LLMPROXY_KEY is optional at deployment time (only used as fallback)
LLMPROXY_KEY="${LLMPROXY_KEY:-}"
LLM_MODEL="${LLM_MODEL:-gpt-4o}"
LLM_BASE_URL="${LLM_BASE_URL:-https://llm-proxy-api.ai.eng.netapp.com/v1}"
HTTPS_PROXY="${HTTPS_PROXY:-http://10.251.20.33:3128}"

# ============== Helpers ==============
info()    { echo -e "\033[1;34m[INFO]\033[0m $1"; }
success() { echo -e "\033[1;32m[OK]\033[0m $1"; }
warn()    { echo -e "\033[1;33m[WARN]\033[0m $1"; }
error()   { echo -e "\033[1;31m[ERROR]\033[0m $1"; exit 1; }

cleanup_container() {
    local name=$1
    if podman ps -a --format "{{.Names}}" | grep -q "^${name}$"; then
        info "Stopping and removing container: $name"
        podman stop "$name" 2>/dev/null || true
        podman rm -f "$name" 2>/dev/null || true
    fi
}

cleanup_pod() {
    if podman pod exists "$POD_NAME" 2>/dev/null; then
        info "Removing existing pod: $POD_NAME"
        podman pod rm -f "$POD_NAME" 2>/dev/null || true
    fi
}

cleanup_orphaned() {
    podman system prune -f 2>/dev/null || true
}

build_image() {
    info "Building image: $APP_IMAGE"
    podman build -t "$APP_IMAGE" -f "$REPO_ROOT/Dockerfile" "$REPO_ROOT"
}

start_container() {
    info "Starting War Room server: $APP_CONTAINER"

    local env_args=(
        -e LLM_MODEL="$LLM_MODEL"
        -e LLM_BASE_URL="$LLM_BASE_URL"
        -e HTTPS_PROXY="$HTTPS_PROXY"
        -e PORT="$APP_PORT"
        -e SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt
    )
    # Only pass LLMPROXY_KEY if set (optional fallback)
    if [ -n "$LLMPROXY_KEY" ]; then
        env_args+=(-e LLMPROXY_KEY="$LLMPROXY_KEY")
    fi

    podman run -d \
        --pod "$POD_NAME" \
        --name "$APP_CONTAINER" \
        "${env_args[@]}" \
        "$APP_IMAGE"
}

# ============== Commands ==============

deploy_all() {
    info "Deploying $POD_NAME (port ${APP_PORT})..."

    cleanup_container "$APP_CONTAINER"
    cleanup_pod
    cleanup_orphaned

    info "Creating pod: $POD_NAME (host networking)"
    podman pod create \
        --name "$POD_NAME" \
        --network host

    build_image
    start_container

    success "Deployment complete!"
    echo ""
    info "⚔️  War Room available at: http://localhost:${APP_PORT}"
    echo ""
    status
}

restart() {
    info "Rebuilding and restarting War Room server..."
    cleanup_container "$APP_CONTAINER"
    build_image
    start_container
    success "Server restarted."
}

start_all() {
    info "Starting pod: $POD_NAME"
    podman pod start "$POD_NAME" || error "Failed to start pod (has it been deployed?)"
    success "Pod started."
}

stop_all() {
    info "Stopping pod: $POD_NAME"
    podman pod stop "$POD_NAME" 2>/dev/null || warn "Pod not running or not found."
    success "Pod stopped."
}

status() {
    info "Pod status:"
    podman pod ps --filter "name=${POD_NAME}" 2>/dev/null || true
    echo ""
    info "Container status:"
    podman ps -a --filter "pod=${POD_NAME}" 2>/dev/null || true
}

show_logs() {
    info "Showing logs for ${APP_CONTAINER} (Ctrl+C to stop)..."
    podman logs -f "$APP_CONTAINER"
}

destroy() {
    warn "This will remove the pod and all containers."
    read -p "Are you sure? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        cleanup_container "$APP_CONTAINER"
        cleanup_pod
        success "All resources removed."
    else
        info "Aborted."
    fi
}

usage() {
    cat << EOF
Usage: $0 <command>

Commands:
  deploy      Build image, create pod, and start the War Room server
  restart     Rebuild image and restart container (preserves pod)
  start       Start a stopped pod
  stop        Stop the running pod
  status      Show pod and container status
  logs        Follow container logs
  destroy     Remove pod and all containers
  help        Show this help message

Required Environment Variables:
  LLMPROXY_KEY        LLM proxy credentials (format: user=<name>&key=<token>)

Optional Environment Variables:
  TAG                 Image tag (default: dev)
  APP_PORT            Server port (default: 8080)
  LLM_MODEL           LLM model name (default: gpt-4o)
  LLM_BASE_URL        LLM API endpoint (default: https://llm-proxy-api.ai.eng.netapp.com/v1)
  HTTPS_PROXY         NetApp proxy (default: http://10.251.20.33:3128)

Examples:
  $0 deploy                      # Full deployment on port 8080
  $0 logs                        # Follow server logs
  $0 status                      # Check pod/container status
  APP_PORT=9090 $0 deploy        # Deploy on a different port
  LLM_MODEL=gpt-4.1 $0 deploy   # Deploy with a specific model
  TAG=prod $0 deploy             # Deploy with production tag
EOF
}

# ============== Entry Point ==============
case "${1:-deploy}" in
    deploy)             deploy_all ;;
    restart)            restart ;;
    start)              start_all ;;
    stop)               stop_all ;;
    status)             status ;;
    logs)               show_logs ;;
    destroy)            destroy ;;
    help|--help|-h)     usage ;;
    *)
        error "Unknown command: $1"
        usage
        exit 1
        ;;
esac
