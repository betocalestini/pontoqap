# Logging operacional (API e worker)

Logs estruturados em JSON via `log/slog`, nível configurável com `LOG_LEVEL` (`debug`, `info`, `warn`, `error`).

## Onde ver

```bash
docker compose logs -f api
docker compose logs -f worker
```

Em desenvolvimento local, `make api` e `make worker` escrevem no stdout.

## Correlação

- **HTTP:** campo `request_id` (middleware Chi + `httpx.RequestIDMiddleware`).
- **Worker:** `job_id`, `worker_id` (poll do worker), `type` do job.

## Eventos críticos (resumo)

| Área | Exemplos de mensagem |
|------|----------------------|
| Pix sandbox | `sandbox pix webhook received/rejected`, `pix charge created/reused`, `sandbox pix payment settled` |
| Mercado Pago | `mercado pago webhook received/rejected`, `mercado pago api call` |
| Worker | `job completed/failed`, `billing monthly close job completed`, `outbox email sent/failed` |
| Checkout | `checkout rejected` (`reason`: `credit_limit`, `stock`, `customer_blocked`, …) |
| Admin auth | `admin login failed` (`reason` genérico, sem e-mail) |
| Access log restrito | `http access` em webhooks e fechamento/ajuste de faturamento admin |

## O que não logar

Conforme `docs/pix.md`: access token, segredo de webhook, QR/copia e cola, corpo integral de PSP, e-mail completo (usar máscara `u***@dominio` em envios de outbox).

## Auditoria vs log

Ações de usuário continuam em `audit_logs` (`audit.Log`). Log operacional cobre falhas, integrações e jobs — evitar duplicar o mesmo evento com o mesmo detalhe nos dois canais.
