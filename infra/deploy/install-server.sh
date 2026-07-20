#!/usr/bin/env bash
# Preparação única do servidor Linux (Ubuntu/Debian).
# Execute como root ou com sudo: curl ... | bash  OU  sudo ./install-server.sh
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq ca-certificates curl git ufw

if ! command -v docker >/dev/null; then
  curl -fsSL https://get.docker.com | sh
fi

if ! docker compose version >/dev/null 2>&1; then
  apt-get install -y -qq docker-compose-plugin || true
fi

ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable || true

DEPLOY_USER="${DEPLOY_USER:-deploy}"
if ! id "$DEPLOY_USER" &>/dev/null; then
  useradd -m -s /bin/bash "$DEPLOY_USER"
  usermod -aG docker "$DEPLOY_USER"
fi

mkdir -p /opt/store-platform
chown -R "$DEPLOY_USER:$DEPLOY_USER" /opt/store-platform

echo "Servidor pronto. Próximo passo:"
echo "  1) Adicionar chave SSH do usuário $DEPLOY_USER"
echo "  2) Clonar o repositório em /opt/store-platform"
echo "  3) Criar infra/compose/.env.production a partir do .example"
