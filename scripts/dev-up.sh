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

# Confere se o seed populou o catálogo a partir do CSV (não binário antigo).
verify_seed_catalog() {
  local legacy_count total_count
  legacy_count="$(docker compose exec -T postgres psql -U store -d store -tAc "SELECT COUNT(*) FROM products WHERE name LIKE 'Produto seed%';" | tr -d '[:space:]')"
  total_count="$(docker compose exec -T postgres psql -U store -d store -tAc "SELECT COUNT(*) FROM products;" | tr -d '[:space:]')"
  if [[ "${legacy_count:-0}" -gt 0 ]]; then
    die "seed parece desatualizado (Produto seed); rode: docker compose --profile seed build --no-cache seed && docker compose --profile seed run --rm seed"
  fi
  if [[ "${total_count:-0}" -eq 0 ]]; then
    die "seed não criou produtos; veja logs do seed e backend/devdata/products.csv (unit_cost_cents obrigatório)"
  fi
  log "Seed OK: ${total_count} produto(s) no catálogo"
}

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
  if [[ "$CLEAN" == true ]]; then
    log "Populando dados de demonstração (seed)"
    docker compose --profile seed build seed
    docker compose --profile seed run --rm --build seed
    verify_seed_catalog
  fi
  log "Ambiente Docker pronto"
  echo ""
  echo "  Loja:   http://localhost:5173"
  echo "  Admin:  http://localhost:5174"
  echo "  API:    http://localhost:8080/health"
  echo "  Mailpit (e-mails dev): http://localhost:8025"
  echo "  Login admin bootstrap: admin@loja.local / ChangeMe123!"
  if [[ "$CLEAN" == true ]]; then
    echo "  Demo (seed): demo-gerente@demo.loja.local / DemoStore123! (e demais @demo.loja.local)"
    echo "  Re-seed manual: make seed-demo"
  fi
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

if [[ "$CLEAN" == true ]]; then
  log "Populando dados de demonstração (seed)"
  (cd backend && SEED_DATA_DIR=./devdata APP_ENV=development SEED_ALLOW=true go run ./cmd/seed)
fi

log "Postgres e schema prontos; suba API/worker/front no host"
echo ""
echo "  Terminal 1: make api"
echo "  Terminal 2: make worker"
echo "  Terminal 3: pnpm dev:store   → http://localhost:5173"
echo "  Terminal 4: pnpm dev:admin   → http://localhost:5174"
echo ""
echo "  Credenciais admin: admin@loja.local / ChangeMe123!"
if [[ "$CLEAN" == true ]]; then
  echo "  Demo (seed): demo-gerente@demo.loja.local / DemoStore123!"
  echo "  Re-seed: make seed-demo"
fi
echo ""
echo "Opcional — API/worker em Docker e front no host:"
echo "  docker compose up --build -d api worker"
