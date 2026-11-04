#!/usr/bin/env bash
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"

RUN_DIR="$(run_dir)"
PID_FILE="$RUN_DIR/app.pid"

if [ -f "$PID_FILE" ]; then
  pid="$(cat "$PID_FILE")"
  # Only kill the PID if it really is our binary (works on Linux and macOS).
  if ps -p "$pid" -o args= 2>/dev/null | grep -qF "$RUN_DIR/app"; then
    kill "$pid"
    echo "Stopped application (pid $pid)"
  else
    echo "Application is not running"
  fi
  rm -f "$PID_FILE"
else
  echo "Application is not running"
fi
