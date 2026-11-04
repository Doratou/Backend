#!/usr/bin/env bash
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

RUN_DIR="$(run_dir)"
mkdir -p "$RUN_DIR"

if [ -f "$RUN_DIR/app.pid" ] && kill -0 "$(cat "$RUN_DIR/app.pid")" 2>/dev/null; then
  echo "Application is already running"
  exit 1
fi

go build -o "$RUN_DIR/app" ./cmd/app
echo $$ > "$RUN_DIR/app.pid"
exec "$RUN_DIR/app"
