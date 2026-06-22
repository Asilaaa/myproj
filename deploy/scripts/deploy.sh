#!/usr/bin/env bash
set -Eeuo pipefail

APP_DIR="${APP_DIR:-/opt/myproj}"
APP_USER="${APP_USER:-myproj}"
APP_GROUP="${APP_GROUP:-$APP_USER}"
BACKEND_DIR="${BACKEND_DIR:-$APP_DIR/backend}"
FRONTEND_DIR="${FRONTEND_DIR:-$APP_DIR/frontend}"
API_BINARY="${API_BINARY:-$APP_DIR/bin/myproj-api}"
API_SERVICE="${API_SERVICE:-myproj-api}"
FRONTEND_SERVICE="${FRONTEND_SERVICE:-myproj-frontend}"
GO_CACHE="${GO_CACHE:-$APP_DIR/.cache/go-build}"
BACKEND_HEALTHCHECK_URL="${BACKEND_HEALTHCHECK_URL:-http://127.0.0.1:8080/health}"

log() {
  printf '\n[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$*"
}

have_service() {
  systemctl list-unit-files --type=service --no-legend 2>/dev/null | awk '{print $1}' | grep -Fxq "$1.service"
}

run_as_app() {
  sudo -u "$APP_USER" env HOME="$APP_DIR" "$@"
}

log "Preparing directories"
sudo mkdir -p "$APP_DIR/bin" "$GO_CACHE"
sudo chown -R "$APP_USER:$APP_GROUP" "$APP_DIR"

log "Building backend"
run_as_app /usr/bin/env GOCACHE="$GO_CACHE" /bin/bash -lc "cd '$BACKEND_DIR' && go build -o '$API_BINARY' ./cmd/api"

log "Installing frontend dependencies"
run_as_app /bin/bash -lc "cd '$FRONTEND_DIR' && npm ci"

log "Building frontend"
run_as_app /bin/bash -lc "cd '$FRONTEND_DIR' && npm run build"

if have_service "$API_SERVICE"; then
  log "Restarting $API_SERVICE"
  sudo systemctl restart "$API_SERVICE"
  sudo systemctl status "$API_SERVICE" --no-pager -l
else
  log "Skipping API restart; service $API_SERVICE.service not found"
fi

if have_service "$FRONTEND_SERVICE"; then
  log "Restarting $FRONTEND_SERVICE"
  sudo systemctl restart "$FRONTEND_SERVICE"
  sudo systemctl status "$FRONTEND_SERVICE" --no-pager -l
else
  log "Skipping frontend restart; service $FRONTEND_SERVICE.service not found"
fi

if command -v curl >/dev/null 2>&1; then
  log "Running backend healthcheck"
  curl --fail --silent --show-error "$BACKEND_HEALTHCHECK_URL" >/dev/null
fi

log "Deploy completed"
