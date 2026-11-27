#!/usr/bin/env bash
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

RUN_DIR="$(run_dir)"
mkdir -p "$RUN_DIR"

WORKERS=("email" "sms" "ws")
PIDS=()

cleanup() {
  echo ""
  echo "Shutting down all workers..."
  for pid in "${PIDS[@]}"; do
    if kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null
    fi
  done
  for w in "${WORKERS[@]}"; do
    rm -f "$RUN_DIR/worker-$w.pid"
  done
  echo "All workers stopped cleanly."
  exit 0
}

trap cleanup SIGINT SIGTERM EXIT

echo "Building Doratou notification workers (email, sms, ws)..."
for w in "${WORKERS[@]}"; do
  go build -o "$RUN_DIR/worker-$w" "./cmd/worker/$w"
done

echo "Starting all workers (email, sms, ws)... Press Ctrl+C to stop."
for w in "${WORKERS[@]}"; do
  "$RUN_DIR/worker-$w" &
  pid=$!
  PIDS+=("$pid")
  echo "$pid" > "$RUN_DIR/worker-$w.pid"
  echo "  -> Started $w worker (pid $pid)"
done

wait
