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

Smoke do seed (Postgres migrado; cria domínio único `smoke-*.demo.loja.local`):

```bash
make migrate-up
cd backend && go test -run TestRun_smoke ./internal/devseed/ -count=1
```

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
