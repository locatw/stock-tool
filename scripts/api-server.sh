#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

PORT="${PORT:-8080}"
PID_FILE="/tmp/stock-tool-api.pid"
LOG_FILE="/tmp/stock-tool-api.log"

start() {
    if [ -f "$PID_FILE" ]; then
        local old_pid
        old_pid=$(cat "$PID_FILE")
        if kill -0 "$old_pid" 2>/dev/null; then
            echo "ERROR: API server already running (PID $old_pid)" >&2
            exit 1
        fi
        rm -f "$PID_FILE"
    fi

    if lsof -i ":$PORT" -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo "ERROR: Port $PORT is already in use" >&2
        exit 1
    fi

    # shellcheck source=../.env
    set -a
    source "$REPO_ROOT/.env"
    set +a

    cd "$REPO_ROOT/backend"
    PORT="$PORT" go run ./cmd/api/ > "$LOG_FILE" 2>&1 &
    local pid=$!
    echo "$pid" > "$PID_FILE"
    echo "API server started (PID $pid, PORT $PORT, LOG $LOG_FILE)"
}

wait_ready() {
    local url="http://localhost:${PORT}/health"
    local max_wait=30
    local elapsed=0

    echo "Waiting for API server at $url ..."
    while [ "$elapsed" -lt "$max_wait" ]; do
        if curl -sf "$url" > /dev/null 2>&1; then
            echo "API server is ready"
            return 0
        fi
        sleep 1
        elapsed=$((elapsed + 1))
    done

    echo "ERROR: API server did not become ready within ${max_wait}s" >&2
    echo "--- Last 30 lines of $LOG_FILE ---" >&2
    tail -30 "$LOG_FILE" >&2 2>/dev/null || true
    return 1
}

stop() {
    if [ ! -f "$PID_FILE" ]; then
        echo "No PID file found; server may not be running"
        return 0
    fi

    local pid
    pid=$(cat "$PID_FILE")

    if kill -0 "$pid" 2>/dev/null; then
        echo "Stopping API server (PID $pid) ..."
        kill "$pid" 2>/dev/null || true
        local waited=0
        while [ "$waited" -lt 5 ] && kill -0 "$pid" 2>/dev/null; do
            sleep 1
            waited=$((waited + 1))
        done
        if kill -0 "$pid" 2>/dev/null; then
            echo "Force killing PID $pid"
            kill -9 "$pid" 2>/dev/null || true
        fi
    fi

    rm -f "$PID_FILE"
    echo "API server stopped"
}

case "${1:-}" in
    start)      start ;;
    wait)       wait_ready ;;
    stop)       stop ;;
    *)
        echo "Usage: $0 {start|wait|stop}" >&2
        exit 1
        ;;
esac
