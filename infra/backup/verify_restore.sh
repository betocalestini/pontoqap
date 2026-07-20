#!/usr/bin/env bash
set -euo pipefail

# Valida backup + restore (BK-1116) em banco descartável.
# Uso: DATABASE_URL=postgres://store:store@localhost:5432/store?sslmode=disable ./verify_restore.sh

ROOT="$(cd "$(dirname "$0")" && pwd)"
SRC="${DATABASE_URL:-postgres://store:store@localhost:5432/store?sslmode=disable}"
TMP="$ROOT/../../backups/.verify"
mkdir -p "$TMP"
VERIFY_DB="${VERIFY_DB_NAME:-store_restore_verify}"

export PGPASSWORD="${PGPASSWORD:-store}"

psql "$SRC" -v ON_ERROR_STOP=1 -c "SELECT 1" >/dev/null

FILE="$TMP/verify_$(date +%s).sql.gz"
DATABASE_URL="$SRC" BACKUP_DIR="$TMP" "$ROOT/backup.sh"
LATEST="$(ls -t "$TMP"/store_*.sql.gz | head -1)"

# Banco temporário no mesmo servidor
psql "$SRC" -v ON_ERROR_STOP=1 -tc "SELECT 1 FROM pg_database WHERE datname = '${VERIFY_DB}'" | grep -q 1 && \
  psql "$SRC" -v ON_ERROR_STOP=1 -c "DROP DATABASE ${VERIFY_DB} WITH (FORCE);"
psql "$SRC" -v ON_ERROR_STOP=1 -c "CREATE DATABASE ${VERIFY_DB};"

TARGET="${SRC%%\?*}"
TARGET="${TARGET%/*}/${VERIFY_DB}?sslmode=disable"

DATABASE_URL="$TARGET" "$ROOT/restore.sh" "$LATEST"
COUNT="$(psql "$TARGET" -tAc "SELECT COUNT(*) FROM users")"
if [[ "$COUNT" == "0" ]]; then
  echo "Falha: usuários não restaurados" >&2
  psql "$SRC" -c "DROP DATABASE IF EXISTS ${VERIFY_DB} WITH (FORCE);" >/dev/null || true
  exit 1
fi

psql "$SRC" -c "DROP DATABASE IF EXISTS ${VERIFY_DB} WITH (FORCE);" >/dev/null
rm -f "$TMP"/store_*.sql.gz
echo "Restore verificado com sucesso (banco ${VERIFY_DB})."
