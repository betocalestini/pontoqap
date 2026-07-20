# Roteiro de homologação (MVP)

Checklist alinhado ao critério de conclusão em `docs/decisions.md` (seção 42). Use em staging antes de produção.

## Pré-requisitos

- Stack local ou staging: `docker compose up -d` (postgres, migrate, api, worker, frontends).
- Variáveis de `.env` preenchidas (secrets fortes em staging/prod).
- `make migrate-up` aplicado.

## Automatizado

| Verificação | Comando |
| ----------- | ------- |
| Testes unitários e integração | `make test` (PostgreSQL em `localhost:5432`) |
| Fluxo HTTP checkout | `go test -p 1 ./tests/e2e/...` |
| Fluxo MVP (checkout → fechamento → Pix) | `go test -p 1 ./tests/e2e/... -run TestMVP` |

## Manual (painel e loja)

1. **Gerente** — login em `admin-web` (`gerente@loja.local` / senha do bootstrap).
2. **Catálogo** — cadastrar categoria e produto/SKU (API ou tela Produtos).
3. **Estoque** — entrada de quantidade para o SKU.
4. **Cliente** — cadastro na loja; gerente aprova e define limite.
5. **Compra** — cliente loga, adiciona ao carrinho, finaliza compra.
6. **Faturamento** — conferir período aberto; após competência, worker ou botão “Fechar competência” gera fatura.
7. **Pix** — cliente gera cobrança na fatura; em dev, “Simular pagamento”.
8. **Dashboard** — conferir receita/pedidos/faturas em aberto.
9. **Relatórios** — top produtos, estoque, gerar previsão.
10. **Backup** — `make backup`; validar arquivo em `backups/`.
11. **Restauração** — em banco descartável: `make restore BACKUP=backups/<arquivo>.sql.gz`.

## Pendências conhecidas do MVP completo

- MFA administrativo (BK-0212).
- TLS/reverse proxy em produção (BK-1110).
- PSP Pix real (hoje: sandbox).

Registrar evidências (prints ou IDs de pedido/fatura) na issue de release.
