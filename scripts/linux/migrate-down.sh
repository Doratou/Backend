#!/usr/bin/env bash
# Rolls back STEPS migrations (default 1).  Usage: make migrate-down STEPS=2
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"
steps="${STEPS:-1}"
case "$steps" in ''|*[!0-9]*) die "STEPS must be a positive number, got '$steps'";; esac
run_migrate down "$steps"
