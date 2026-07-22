# Testes

Documentação alinhada ao portão da [homologação](../homologation.md) (fase 0) e ao critério MVP em `docs/decisions.md` (§42).

## Baseline (referência)

| Camada | Arquivos | Comando típico |
|--------|----------|----------------|
| Go unitário (`internal/`) | ~18 `*_test.go` | `cd backend && go test ./internal/...` |
| Go integração | 29+ arquivos em `tests/integration/` | `make test-integration` |
| Go E2E HTTP | `tests/e2e/` | ver homologação fase 0 |
| Vitest (TS) | `packages/shared-core`, `packages/api-client`, `packages/ui`, apps | `pnpm test` |
| Playwright | `e2e/playwright/` | `pnpm exec playwright test` (smoke; ver workflow nightly) |

Métrica informativa de cobertura Go (sem meta de % no CI):

```bash
cd backend && go test -coverprofile=/tmp/cover.out ./...
go tool cover -func=/tmp/cover.out | tail -1
```

Integração e E2E exigem PostgreSQL; sem `DATABASE_URL` acessível esses testes fazem `Skip`.

## Backend (Go)

Testes unitários (sem PostgreSQL):

```bash
cd backend && go test ./internal/...
```

Suite completa (integração e E2E com DB quando disponível):

```bash
make test
```

Com PostgreSQL local (`docker compose up postgres -d`):

```bash
make migrate-up
make test-integration
```

E2E explícitos (homologação 0.2–0.3):

```bash
cd backend && go test -p 1 ./tests/e2e/... -run 'TestHTTP|TestMVP'
```

Backup e restore (homologação 0.4):

```bash
make test-backup-restore
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

### Integração (`backend/tests/integration/`)

| Arquivo | Área |
|---------|------|
| `admin_jwt_test.go`, `store_jwt_test.go` | JWT loja/admin |
| `auth_profile_test.go` | PATCH `/auth/me` |
| `identity_admin_test.go`, `identity_admin_ops_test.go` | convites, revoke, role/status/sessions |
| `identity_mfa_http_test.go` | MFA setup/verify HTTP |
| `identity_billing_test.go` | sessão, período aberto |
| `dual_role_test.go`, `rbac_access_test.go` | papéis e permissões |
| `customers_*` | cadastro, e-mail, bloqueio, colaborador, listagem |
| `catalog_*`, `pricing_lots_test.go`, `promo_checkout_test.go` | catálogo, preço, promo |
| `cart_test.go`, `sales_inventory_test.go`, `checkout_concurrency_test.go` | carrinho, checkout, estoque |
| `collaborator_checkout_test.go` | preço colaborador |
| `admin_orders_*` | pedidos admin |
| `billing_*`, `customer_billing_view_test.go` | faturamento e Pix |
| `payments_pix_charge_http_test.go` | cobrança Pix admin HTTP |
| `reports_*`, `reports_endpoints_test.go` | dashboard, séries, relatórios, CSV, forecast |
| `audit_logs_test.go` | `GET /admin/audit/logs` |
| `openapi_drift_test.go` | paths MVP no OpenAPI |

### E2E (`backend/tests/e2e/`)

| Arquivo | Fluxo |
|---------|--------|
| `checkout_flow_test.go` | checkout HTTP ponta a ponta |
| `mvp_flow_test.go` | fechamento, Pix sandbox, dashboard |

### Unitários notáveis (`internal/`)

| Área | Arquivo |
|------|---------|
| Senha (Argon2) | `identity/security/password_test.go` |
| JWT | `identity/security/jwt_test.go` |
| Login / MFA | `identity/service_test.go`, `identity/mfa_test.go` |
| Middleware permissão | `identity/transport/http/middleware_test.go` |
| Limite cliente | `customers/limit_test.go` |
| Preço / promo / arredondamento | `catalog/*_test.go` |
| Dias úteis / vencimento | `billing/*_test.go` |
| Webhook Pix (valor divergente) | `payments/webhook_test.go` |
| Worker jobs | `jobs/runner_test.go` |
| Seed dev | `devseed/*_test.go` |

## Frontend

```bash
pnpm test
```

Pacotes com Vitest: `@store/shared-core`, `@store/api-client`, `@store/ui`, `admin-web`, `store-web`.

## CI

Pull request (`.github/workflows/pull-request.yml`): `go test -p 1 ./...` com Postgres, passo E2E nomeado, `pnpm test`, build dos apps.

Backup-restore: workflow `backup-restore-nightly.yml` (agendado).

## Rastreabilidade MVP (§42) — resumo

| Passo §42 | Automatizado | Manual / lacuna |
|-----------|--------------|-----------------|
| 1 MFA gerente | `identity_mfa_http_test.go` + homologação fase 1 | TOTP em app autenticador |
| 2–4 catálogo/estoque | `catalog_*`, `pricing_lots`, `catalog_inventory` | UI admin |
| 5–9 cliente/checkout | `customers_email`, `cart`, `sales_inventory`, E2E checkout | — |
| 10–12 estoque/exposição/período | `sales_inventory`, `billing_cycles` | — |
| 13 worker fechamento | `jobs/runner_test.go`, `billing_cycles` | processo worker em produção |
| 14–18 fatura/Pix/webhook | `billing_payments`, `mvp_flow`, `payments_*` | PSP real em staging |
| 19–22 dashboard/relatórios/previsão | `reports_*`, `reports_endpoints` | exports na UI |
| 23 backup | `make test-backup-restore` / nightly CI | — |

## Playwright (browser)

Smoke E2E com Playwright em [`e2e/playwright/`](../../e2e/playwright/). Por padrão os testes são **ignorados**; para rodar com stack local (`make dev-up`):

```bash
E2E_RUN=1 E2E_BASE_URL=http://localhost:5174 pnpm test:e2e
```

O script `test:run` instala o Chromium do Playwright automaticamente na primeira execução. `pnpm test` na raiz **não** roda browser E2E (apenas mensagem de skip no pacote `@store/e2e-playwright`).


## Antes de abrir PR

Execute a fase 0 da homologação localmente (ou confira o job verde no CI). Checklist em [`.github/pull_request_template.md`](../../.github/pull_request_template.md).
