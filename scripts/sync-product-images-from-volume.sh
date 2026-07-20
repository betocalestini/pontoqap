#!/usr/bin/env bash
# Copia imagens do volume Docker legado (store_upload_data) para o catálogo estático no repo.
# Uso (na raiz do monorepo): ./scripts/sync-product-images-from-volume.sh
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEST="$ROOT/backend/internal/catalog/static/product-images"
mkdir -p "$DEST"
if ! docker volume inspect store_upload_data >/dev/null 2>&1; then
  echo "Volume store_upload_data não encontrado; nada a copiar."
  exit 0
fi
docker run --rm \
  -v store_upload_data:/from:ro \
  -v "$DEST:/to" \
  alpine:3.21 \
  sh -c 'if [ -d /from/product-images ]; then cp -av /from/product-images/. /to/; else cp -av /from/. /to/; fi'
echo "Imagens em $DEST"
