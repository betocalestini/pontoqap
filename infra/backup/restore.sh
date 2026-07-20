#!/usr/bin/env bash
set -euo pipefail

# Restaura backup gerado por backup.sh (SQL gzip).
# Uso: DATABASE_URL=... ./restore.sh backups/store_YYYYMMDD_HHMMSS.sql.gz

if [[ $# -lt 1 ]]; then
  echo "Uso: $0 <arquivo.sql.gz>" >&2
  exit 1
fi
FILE="$1"
if [[ ! -f "$FILE" ]]; then
  echo "Arquivo não encontrado: $FILE" >&2
  exit 1
fi
if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "DATABASE_URL é obrigatório" >&2
  exit 1
fi

echo "Restaurando $FILE em $DATABASE_URL ..."
gunzip -c "$FILE" | sed '/^SET transaction_timeout/d' | psql "$DATABASE_URL" -v ON_ERROR_STOP=1
echo "Restauração concluída."
