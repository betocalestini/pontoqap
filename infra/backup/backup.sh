#!/usr/bin/env bash
set -euo pipefail

# Backup lógico do PostgreSQL (uso em cron ou stack de produção).
# Variáveis: DATABASE_URL ou PG* padrão; BACKUP_DIR (default ./backups)

BACKUP_DIR="${BACKUP_DIR:-./backups}"
STAMP="$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

FILE="$BACKUP_DIR/store_${STAMP}.sql.gz"
echo "Gerando $FILE ..."
pg_dump "${DATABASE_URL}" | gzip -9 > "$FILE"
echo "Concluído."
