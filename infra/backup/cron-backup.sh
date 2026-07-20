#!/usr/bin/env bash
# Cron diário no servidor: 0 3 * * * /opt/store-platform/infra/backup/cron-backup.sh
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)/.."
cd "$ROOT"
ENV_FILE="${ENV_FILE:-infra/compose/.env.production}"
set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a
export PGPASSWORD="${PGPASSWORD:-$POSTGRES_PASSWORD}"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/store-platform}"
mkdir -p "$BACKUP_DIR"
DATABASE_URL="${DATABASE_URL}" BACKUP_DIR="$BACKUP_DIR" ./infra/backup/backup.sh
find "$BACKUP_DIR" -name 'store_*.sql.gz' -mtime +14 -delete
