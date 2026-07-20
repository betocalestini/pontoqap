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
| `docs/` | Requisitos, arquitetura, decisões e backlog |

## Pré-requisitos

- Go 1.25+
- Node 22+ e pnpm 9+
- Docker (opcional, para stack completa)

## Desenvolvimento local

```bash
cp .env.example .env
docker compose up postgres -d
make migrate-up
make api
```

Em outro terminal:

```bash
pnpm install
pnpm dev:store   # http://localhost:5173
pnpm dev:admin   # http://localhost:5174
```

### Credenciais iniciais (bootstrap)

Após a primeira subida da API, é criado um gerente de demonstração:

- **E-mail:** `gerente@loja.local`
- **Senha:** `ChangeMe123!`
- **Audience no login:** enviar `X-App-Audience: admin` ou `"audience": "admin"` no corpo

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

Próximos passos naturais: fechamento mensal, jobs/outbox, Pix, dashboard e relatórios.

## Testes

```bash
make test
```

## Licença

Proprietário — uso interno do projeto.
