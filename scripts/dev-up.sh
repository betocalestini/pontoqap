#!/usr/bin/env bash
# Reset e subida do ambiente de desenvolvimento (raiz do monorepo).
#
# Uso:
#   ./scripts/dev-up.sh              # limpa volumes, sobe tudo via Docker Compose
#   ./scripts/dev-up.sh --no-clean   # mantém dados do Postgres
#   ./scripts/dev-up.sh --local      # só Postgres + migrations + deps; API/front no host
#   ./scripts/dev-up.sh --help
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

CLEAN=true
MODE=docker

usage() {
  sed -n '2,8p' "$0" | sed 's/^# \{0,1\}//'
}

log() { printf '==> %s\n' "$*"; }

die() { printf 'erro: %s\n' "$*" >&2; exit 1; }

run_make() {
  if command -v mise >/dev/null 2>&1 && [[ -f .mise.toml ]]; then
    mise exec -- make "$@"
  else
    make "$@"
  fi
}

run_pnpm() {
  if command -v mise >/dev/null 2>&1 && [[ -f .mise.toml ]]; then
    mise exec -- pnpm "$@"
  else
    pnpm "$@"
  fi
}

wait_postgres() {
  log "Aguardando Postgres ficar pronto..."
  local i=0
  until docker compose exec -T postgres pg_isready -U store -d store >/dev/null 2>&1; do
    i=$((i + 1))
    [[ $i -gt 60 ]] && die "Postgres não respondeu a tempo (60s)"
    sleep 1
  done
}

wait_api() {
  log "Aguardando API em http://localhost:8080/health ..."
  local i=0
  until curl -sf http://localhost:8080/health >/dev/null 2>&1; do
    i=$((i + 1))
    [[ $i -gt 90 ]] && die "API não respondeu a tempo; veja: docker compose logs api"
    sleep 2
  done
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --no-clean) CLEAN=false; shift ;;
    --local) MODE=local; shift ;;
    -h | --help)
      usage
      exit 0
      ;;
    *)
      die "opção desconhecida: $1 (use --help)"
      ;;
  esac
done

command -v docker >/dev/null || die "Docker não encontrado no PATH"
docker compose version >/dev/null 2>&1 || die "Docker Compose (plugin) não encontrado"

if [[ ! -f .env ]]; then
  log "Criando .env a partir de .env.example"
  cp .env.example .env
fi

if command -v mise >/dev/null 2>&1 && [[ -f .mise.toml ]]; then
  log "Ferramentas mise (.mise.toml)"
  mise trust -q 2>/dev/null || true
  mise install -q
fi

if [[ "$CLEAN" == true ]]; then
  log "Parando stack e removendo volumes (banco zerado)"
  docker compose down --remove-orphans -v
else
  log "Parando stack (dados do Postgres preservados)"
  docker compose down --remove-orphans
fi

if [[ "$MODE" == docker ]]; then
  log "Build e subida completa (postgres → migrate → api, worker, loja, admin)"
  docker compose up --build -d
  wait_api
  log "Ambiente Docker pronto"
  echo ""
  echo "  Loja:   http://localhost:5173"
  echo "  Admin:  http://localhost:5174"
  echo "  API:    http://localhost:8080/health"
  echo "  Mailpit (e-mails dev): http://localhost:8025"
  echo "  Login:  gerente@loja.local / ChangeMe123! (audience admin)"
  echo ""
  echo "Logs: docker compose logs -f api"
  exit 0
fi

# Modo local: Postgres no Docker; Go/pnpm no host (hot reload nos frontends)
log "Subindo apenas Postgres"
docker compose up -d postgres
wait_postgres

log "Dependências (go mod + pnpm)"
run_make deps

log "Migrations"
run_make migrate-up

log "Postgres e schema prontos; suba API/worker/front no host"
echo ""
echo "  Terminal 1: make api"
echo "  Terminal 2: make worker"
echo "  Terminal 3: pnpm dev:store   → http://localhost:5173"
echo "  Terminal 4: pnpm dev:admin   → http://localhost:5174"
echo ""
echo "  Credenciais: gerente@loja.local / ChangeMe123!"
echo ""
echo "Opcional — API/worker em Docker e front no host:"
echo "  docker compose up --build -d api worker"
