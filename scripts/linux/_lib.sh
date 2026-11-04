#!/usr/bin/env bash
# Shared helpers for the Linux/macOS scripts. Sourced by the other scripts.
set -euo pipefail

LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$LIB_DIR/../.." && pwd)"
cd "$ROOT_DIR"

COMPOSE_PATH="${COMPOSE_PATH:-docker/docker-compose.yaml}"
ENV_FILE="${ENV_FILE:-.env}"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-$ROOT_DIR/migrations}"

die() { echo "ERROR: $*" >&2; exit 1; }

require_docker() {
  command -v docker >/dev/null 2>&1 || die "docker is not installed or not in PATH."
  docker info >/dev/null 2>&1 || die "Docker daemon is not running."
}

# env_get KEY DEFAULT : shell env first, then .env file, then default.
env_get() {
  local key="$1" default="${2:-}" val="${!1:-}"
  if [ -z "$val" ] && [ -f "$ENV_FILE" ]; then
    val="$(grep -E "^[[:space:]]*${key}=" "$ENV_FILE" | tail -n1 | cut -d= -f2- | tr -d '\r' || true)"
    val="${val%\"}"; val="${val#\"}"; val="${val%\'}"; val="${val#\'}"
  fi
  printf '%s' "${val:-$default}"
}

compose() {
  docker compose --env-file "$ENV_FILE" -f "$COMPOSE_PATH" "$@"
}

db_container() {
  local id
  id="$(compose ps -q "$DB_SERVICE" | head -n1)"
  [ -n "$id" ] || die "Database container '$DB_SERVICE' is not running. Run 'make up' first."
  printf '%s' "$id"
}

# run_migrate <goose args...>
# Uses the locally installed Goose CLI. The database must be reachable at
run_migrate() {
  command -v goose >/dev/null 2>&1 || die "goose is not installed or not in PATH."
  [ -d "$MIGRATIONS_DIR" ] || die "Migrations folder not found: $MIGRATIONS_DIR"
  local user pass name port url
  user="$(env_get PDB_USER admin)"
  pass="$(env_get PDB_PASSWORD admin)"
  name="$(env_get PDB_NAME doratou)"
  port="$(env_get PDB_PORT 5432)"
  url="postgres://${user}:${pass}@127.0.0.1:${port}/${name}?sslmode=disable"
  goose -dir "$MIGRATIONS_DIR" postgres "$url" "$@"
}

# Per-checkout runtime dir (pid file + built binary), same idea as the old Makefile.
run_dir() {
  echo "/tmp/doratou-$(id -u)-$(pwd | cksum | cut -d' ' -f1)"
}
