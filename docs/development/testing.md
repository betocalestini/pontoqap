# Testes

## Backend (Go)

Testes unitários (sem PostgreSQL):

```bash
cd backend && go test ./internal/...
```

Suite completa (integração e E2E pulam automaticamente se `DATABASE_URL` não estiver acessível):

```bash
make test
```

Com PostgreSQL local (`docker compose up postgres -d`):

```bash
make migrate-up
make test-integration
```

Smoke do seed (Postgres migrado; usa `internal/devseed/testdata/`; domínio único `smoke-*.demo.loja.local`):

```bash
make migrate-up
cd backend && go test -run TestRun_smoke ./internal/devseed/ -count=1
```

### Dados CSV do seed (`backend/devdata/`)

| Arquivo | Colunas |
|---------|---------|
| `products.csv` | `name`, `slug`, `sku_code`, `category`, `unit`, `unit_cost_cents`, `margin_percent` (opcional), `stock_qty`, `image_slug` (opcional) |
| `customers.csv` | `name`, `email`, `credit_limit_cents`, `collaborator` (`true`/`false`) |

Preço de venda **não** vai no CSV: após entrada de estoque o seed chama `RecalculateSKU` (margem + arredondamento R$ 0,50). Variável `SEED_DATA_DIR` ou flag `-data-dir` apontam para outra pasta. `stock_qty` 0 ou vazio → estoque demo padrão (50). `./scripts/dev-up.sh --no-clean` não roda o seed; use `make seed-demo` para reler o CSV.

### Cobertura atual

| Área | Tipo | Arquivo |
|------|------|---------|
| Senha (Argon2) | unitário | `internal/identity/security/password_test.go` |
| Login / papéis | unitário (mock) | `internal/identity/service_test.go` |
| Middleware de permissão | HTTP | `internal/identity/transport/http/middleware_test.go` |
| Limite disponível | unitário | `internal/customers/limit_test.go` |
| Catálogo, clientes, estoque, checkout | integração | `tests/integration/*` |
| Seed de desenvolvimento (`Run`) | integração (skip sem DB) | `internal/devseed/devseed_test.go` |
| Fluxo HTTP checkout | E2E | `tests/e2e/checkout_flow_test.go` |

## Frontend

```bash
pnpm --filter @store/shared-core test
```
