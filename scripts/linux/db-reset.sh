#!/usr/bin/env bash
# Wipes the whole database (all schemas objects, types, tables) and re-runs all migrations.
# Skip the confirmation with:  make db-reset YES=1
source "$(dirname "${BASH_SOURCE[0]}")/_lib.sh"
require_docker

user="$(env_get PDB_USER admin)"
name="$(env_get PDB_DB_NAME doratou)"

if [ "${YES:-}" != "1" ]; then
  echo "This will DELETE ALL DATA in database '$name' and re-run all migrations."
  read -r -p "Type 'yes' to continue: " answer
  [ "$answer" = "yes" ] || { echo "Aborted."; exit 1; }
fi

compose exec -T "$DB_SERVICE" psql -U "$user" -d "$name" -v ON_ERROR_STOP=1 \
  -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
run_migrate up
echo "Database '$name' was reset."
