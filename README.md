# Store Platform

Monorepositório da **Store Platform**: vendas pós-pagas, estoque, faturamento mensal e pagamento via Pix, conforme `docs/`.

## Estrutura

| Caminho | Descrição |
| -------- | ----------- |
| `backend/` | API e worker em Go (monólito modular) |
| `apps/store-web/` | Loja React (cliente) |
| `apps/admin-web/` | Painel React (gerente) |
| `apps/mobile/` | Base Expo (futuro) |
| `packages/` | Contratos, cliente HTTP, tokens, validações |
| `docs/` | Requisitos, arquitetura, decisões, [controle de acesso](docs/access-control.md) e backlog |

## Pré-requisitos

- Go 1.25+ (recomendado: [mise](https://mise.jdx.dev) — na raiz do repo rode `mise install`)
- Node 22+ e pnpm 9+
- Docker (Postgres e stack completa)

**Importante:** execute `make` e `docker compose` sempre na **raiz** do repositório (`store/`), não dentro de `backend/`.

## Desenvolvimento local

### Subida completa (recomendado)

Zera o banco, aplica migrations e sobe Postgres, API, worker, loja e admin via Docker:

```bash
cd /caminho/para/store
make dev-up
# equivalente: ./scripts/dev-up.sh
```

Para manter os dados do Postgres: `./scripts/dev-up.sh --no-clean`

Para só resetar o banco e rodar API/front no host (Vite com hot reload):

```bash
make dev-up-local
```

### Passo a passo manual

```bash
cd /caminho/para/store   # raiz do monorepo
mise install             # se usar mise (instala Go/Node do .mise.toml)
cp .env.example .env
docker compose up postgres -d
make migrate-up
```

Em **três terminais** (todos na raiz `store/`):

```bash
make api          # :8080 — obrigatório antes do pnpm dev:store/admin
make worker
pnpm dev:store    # :5173 (proxy /api → localhost:8080)
```

Outro terminal para o painel:

```bash
pnpm dev:admin    # :5174
```

```bash
pnpm install
make openapi-gen   # tipos OpenAPI (só na raiz do repo)
```

### Credenciais iniciais (bootstrap)

Após a primeira subida da API, é criado um **administrador** bootstrap (se ainda não existir `system_admin` no banco), configurável via `ADMIN_BOOTSTRAP_*` no `.env`:

- **E-mail (padrão):** `admin@loja.local`
- **Senha (padrão):** `ChangeMe123!`
- Demais funcionários: cadastro na **loja** como cliente, depois convite pelo menu **Usuários** ou papel em **Clientes** (papel fixo; ver [docs/access-control.md](docs/access-control.md)).
- **Audience no login:** enviar `X-App-Audience: admin` ou `"audience": "admin"` no corpo
- **Painel admin:** após o login, a API devolve um JWT (`access_token`) com validade fixa definida por `SESSION_TTL_ADMIN` (padrão 8h). O navegador envia `Authorization: Bearer …` nas requisições; sem token válido o painel redireciona para `/login`.

### Clientes (cadastro e e-mail)

1. O cliente se cadastra na loja (`/cadastro`).
2. Recebe um e-mail de confirmação (em dev: [Mailpit](http://localhost:8025) com `docker compose up mailpit -d` ou stack `make dev-up`).
3. Após confirmar o link, a conta é liberada com limite padrão (`DEFAULT_CUSTOMER_CREDIT_LIMIT_CENTS` no `.env`).
4. Ao fechar a fatura, o worker envia um e-mail com resumo e link para `/faturas/{id}`.

O painel admin ainda pode **aprovar manualmente** ou ajustar limite para casos excepcionais.

## Ordem de implementação

O backlog em `docs/decisions.md` (seção 41) define a sequência. Neste repositório já estão materializados:

1. Fundação do monorepositório (EP-01)
2. Migrations PostgreSQL (EP-02 base)
3. Identidade, sessões e permissões (EP-02)
4. Auditoria e `request_id` (parcial, EP-11)
5. Catálogo (EP-03)
6. Estoque (EP-04)
7. Clientes e limite (EP-05)
8. Carrinho e checkout transacional (EP-06)
9. Lançamento em período de faturamento (EP-07 parcial)

## Homologação e deploy

- Roteiro passo a passo (MVP): [docs/homologation.md](docs/homologation.md)
- Servidor, secrets GitHub e pipelines: [docs/deployment.md](docs/deployment.md)

## Testes

Na raiz do repositório:

```bash
make test
make test-backup-restore
```

### Problemas comuns

| Sintoma | Causa | Solução |
| -------- | ----- | ------- |
| `mise ERROR ... shim: pnpm` | pnpm não no mise | Na raiz: `mise trust && mise install` (ver `.mise.toml`) |
| `No rule to make target 'migrate-up'` | `make` rodou em `backend/` | `cd` para a raiz do monorepo |
| Vite `ECONNREFUSED` em `/api/...` | API não está no ar | Terminal separado: `make api` (porta 8080) |
| `Internal Server Error` na loja | Idem | Confirme `curl http://localhost:8080/health` → `{"status":"ok"}` |

Alternativa sem Go local: `docker compose up --build -d` (API em :8080, loja :5173, painel :5174).

## Licença

Proprietário — uso interno do projeto.
