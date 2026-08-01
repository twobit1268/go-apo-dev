#!/usr/bin/env bash
# One-command local dev: starts Postgres + Pub/Sub emulator, runs
# migrations, then starts the API, worker, and web dev server together.
# Ctrl+C stops everything it started (the Go processes and the web dev
# server - the docker-compose infra is left running so the next `dev.sh`
# run doesn't have to re-pull/re-migrate; stop it yourself with
# `make dev-down` when you're done for the day).
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

LOG_DIR="$ROOT_DIR/.dev-logs"
mkdir -p "$LOG_DIR"

export DATABASE_URL="${DATABASE_URL:-postgres://postgres:postgres@localhost:5432/autoparts?sslmode=disable}"
export PUBSUB_EMULATOR_HOST="${PUBSUB_EMULATOR_HOST:-localhost:8085}"
export GCP_PROJECT_ID="${GCP_PROJECT_ID:-auto-parts-local}"

echo "==> Starting Postgres + Pub/Sub emulator (docker compose)..."
docker compose up -d postgres pubsub-emulator

echo "==> Waiting for Postgres..."
until docker compose exec -T postgres pg_isready -U postgres >/dev/null 2>&1; do
  sleep 1
done

echo "==> Running migrations..."
(cd backend && go run ./cmd/migrate up)

if [ ! -d web/node_modules ]; then
  echo "==> Installing web dependencies (first run only)..."
  (cd web && npm install)
fi

PIDS=()

cleanup() {
  echo ""
  echo "==> Stopping api/worker/web (docker infra left running - 'make dev-down' to stop it too)..."
  for pid in "${PIDS[@]:-}"; do
    kill "$pid" >/dev/null 2>&1 || true
  done
  wait 2>/dev/null || true
}
trap cleanup EXIT INT TERM

echo "==> Starting API (logs: .dev-logs/api.log)..."
(cd backend && go run ./cmd/api) >"$LOG_DIR/api.log" 2>&1 &
PIDS+=("$!")

echo "==> Starting worker (logs: .dev-logs/worker.log)..."
(cd backend && go run ./cmd/worker) >"$LOG_DIR/worker.log" 2>&1 &
PIDS+=("$!")

echo "==> Starting web dev server (logs: .dev-logs/web.log)..."
(cd web && npm run dev -- --host) >"$LOG_DIR/web.log" 2>&1 &
PIDS+=("$!")

cat <<EOF

--------------------------------------------------------------
  API:    http://localhost:8080
  Web:    http://localhost:5173
  Logs:   $LOG_DIR/{api,worker,web}.log
--------------------------------------------------------------
Press Ctrl+C to stop the app (docker infra keeps running).
EOF

wait
