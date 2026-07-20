#!/usr/bin/env bash
# Deploy no servidor (staging ou produção).
# Variáveis:
#   REPO_DIR          — raiz do clone (default /opt/store-platform)
#   DEPLOY_REF        — branch ou tag (default master)
#   COMPOSE_FILE      — default infra/compose/compose.traefik.yaml (Traefik); use compose.production.yaml para Caddy dedicado
#   ENV_FILE          — default infra/compose/.env.production
set -euo pipefail

REPO_DIR="${REPO_DIR:-/opt/store-platform}"
DEPLOY_REF="${DEPLOY_REF:-master}"
COMPOSE_FILE="${COMPOSE_FILE:-infra/compose/compose.traefik.yaml}"
ENV_FILE="${ENV_FILE:-infra/compose/.env.production}"

cd "$REPO_DIR"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Arquivo $ENV_FILE não encontrado. Copie de .env.production.example e preencha secrets." >&2
  exit 1
fi

echo "==> Atualizando código ($DEPLOY_REF)"
git fetch origin --tags
git checkout "$DEPLOY_REF"
git pull origin "$DEPLOY_REF" --ff-only

echo "==> Build e subida dos containers"
set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a
if [[ "${USE_REGISTRY_IMAGES:-false}" == "true" ]]; then
  docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" pull
fi
docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" build --pull
docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d --remove-orphans

echo "==> Limpando imagens antigas"
docker image prune -f >/dev/null || true

HEALTH_URL="http://127.0.0.1/health"
if [[ -n "${DOMAIN:-}" ]]; then
  HEALTH_URL="https://${DOMAIN}/health"
fi

echo "==> Health check: $HEALTH_URL"
for i in 1 2 3 4 5 6 7 8 9 10; do
  if curl -fsS "$HEALTH_URL" >/dev/null 2>&1; then
    echo "Deploy concluído com sucesso."
    exit 0
  fi
  sleep 3
done

echo "Health check falhou após deploy. Verifique: docker compose -f $COMPOSE_FILE --env-file $ENV_FILE logs api store-web traefik" >&2
exit 1
