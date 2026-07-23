# Roteiro de homologação passo a passo (MVP)

Alinhado a `docs/decisions.md` (critério de conclusão, §42). Execute em **staging** antes de promover para produção.

## Ambientes e URLs

| Papel | Local (dev) | Staging (exemplo) |
| ----- | ----------- | ----------------- |
| Loja | http://localhost:5173 | https://loja.seudominio.com |
| Painel | http://localhost:5174 | https://admin.loja.seudominio.com |
| API | http://localhost:8080/api/v1 | https://loja.seudominio.com/api/v1 |
| Health | http://localhost:8080/health | https://loja.seudominio.com/health |

Credencial bootstrap (somente dev): `admin@loja.local` / `ChangeMe123!` — primeiro `system_admin` criado na subida da API se não existir administrador. Demais funcionários entram por convite (menu **Usuários**).

---

## Fase 0 — Portão automatizado (obrigatório)

Execute na máquina de CI ou na sua estação com PostgreSQL acessível:

```bash
cp .env.example .env
docker compose up postgres -d
make migrate-up
make test
cd backend && go test -p 1 ./tests/e2e/... -run 'TestHTTP|TestMVP'
make test-backup-restore
```

| # | Verificação | Comando | OK? | Evidência |
| - | ----------- | ------- | --- | --------- |
| 0.1 | Testes backend | `make test` | ☐ | log CI / screenshot |
| 0.2 | E2E checkout | `go test ./tests/e2e/... -run TestHTTP` | ☐ | |
| 0.3 | E2E MVP (Pix) | `go test ./tests/e2e/... -run TestMVP` | ☐ | |
| 0.4 | Restore backup | `make test-backup-restore` | ☐ | |

**Critério de saída:** todos os itens 0.x marcados.

O mesmo portão roda no CI em pull requests (backend + `pnpm test`). Use o checklist em [`.github/pull_request_template.md`](../.github/pull_request_template.md) antes de pedir revisão. Referência de comandos e cobertura: [`docs/development/testing.md`](development/testing.md).

---

## Fase 1 — Gerente e MFA (RF-IDN-008)

| Passo | Ação | Resultado esperado | OK? | Evidência |
| ----- | ---- | ------------------ | --- | --------- |
| 1.1 | Abrir painel (`ADMIN_DOMAIN`) | Tela de login | ☐ | |
| 1.2 | Login e-mail/senha do gerente | Cookie de sessão; se MFA obrigatório e ainda não configurado, redirecionar para `/mfa` | ☐ | |
| 1.3 | Em `/mfa`, gerar segredo e confirmar código TOTP | MFA ativo (`mfa_enabled: true` em `/api/v1/auth/me` com header `X-App-Audience: admin`) | ☐ | print config app autenticador |
| 1.4 | Logout e novo login | Pedir código MFA de 6 dígitos | ☐ | |
| 1.5 | Login com TOTP inválido | 401 / mensagem de credencial inválida | ☐ | |
| 1.6 | Admin → **Usuários** → convidar gerente; abrir link do e-mail (Mailpit em dev) | Aceitar convite, definir senha, login e MFA se exigido | ☐ | |

---

## Fase 1b — Papéis e permissões (`docs/access-control.md`)

Smoke manual por papel (credenciais de teste criadas via convite ou seeds de integração). Marque após login no painel com `X-App-Audience: admin`.

| Papel (`code`) | Menu / ação | Esperado | OK? |
| -------------- | ----------- | -------- | --- |
| `manager` | **Clientes** | Lista carrega (sem erro 403) | ☐ |
| `manager` | **Usuários** | Menu oculto ou 403 na rota | ☐ |
| `manager` | **Faturamento** → fechar competência | Permitido (`billing.close`) | ☐ |
| `manager` | Calendário comercial (`settings.write`) | 403 ao salvar calendário | ☐ |
| `inventory_operator` | **Produtos** leitura | OK | ☐ |
| `inventory_operator` | Alterar preço de SKU | 403 | ☐ |
| `inventory_operator` | Entrada de estoque | OK | ☐ |
| `finance_operator` | **Faturamento** / ajustes | OK | ☐ |
| `finance_operator` | **Estoque** → entrada | 403 | ☐ |
| `system_admin` | **Usuários** → desativar funcionário | `disabled` + sessão revogada | ☐ |
| `manager` | **Pedidos** (lista) | OK com `orders.read` | ☐ |
| `system_admin` | **Auditoria** | Lista logs | ☐ |

---

## Fase 2 — Catálogo e estoque

| Passo | Ação | Resultado esperado | OK? | Evidência |
| ----- | ---- | ------------------ | --- | --------- |
| 2.1 | Painel → **Produtos** → criar produto (nome, SKU, preço) | Produto listado no admin | ☐ | ID produto: |
| 2.2 | (Opcional) API `POST /admin/categories` | Categoria criada | ☐ | |
| 2.3 | Entrada de estoque (API `POST /admin/inventory/entries` ou fluxo futuro na UI) | Saldo disponível no catálogo público | ☐ | SKU: ___ qtd: ___ |

---

## Fase 3 — Cliente e confirmação de e-mail

| Passo | Ação | Resultado esperado | OK? | Evidência |
| ----- | ---- | ------------------ | --- | --------- |
| 3.1 | Loja → **Criar conta** (menu ou `/cadastro`) | Cliente `pending`; usuário `pending_email`; e-mail na outbox/Mailpit | ☐ | e-mail: |
| 3.2 | Abrir link do e-mail (`/verificar-email?token=`) ou Mailpit (http://localhost:8025) | Conta `approved` com limite padrão (`DEFAULT_CUSTOMER_CREDIT_LIMIT_CENTS`) | ☐ | |
| 3.3 | Loja → **Entrar** com o cliente | Sessão store ativa | ☐ | |
| 3.4 | (Opcional) Painel → **Clientes** → Aprovar pendente manual | Útil para exceções; define limite customizado | ☐ | customer_id: |

---

## Fase 4 — Compra e estoque

| Passo | Ação | Resultado esperado | OK? | Evidência |
| ----- | ---- | ------------------ | --- | --------- |
| 4.1 | Catálogo → adicionar ao carrinho | Item no carrinho | ☐ | |
| 4.2 | **Carrinho** → finalizar compra | HTTP 201, pedido criado | ☐ | order_id: |
| 4.3 | Conferir estoque do SKU | Quantidade reduzida | ☐ | antes/depois: |
| 4.4 | Conferir exposição do cliente | `current_exposure_cents` aumentou | ☐ | valor: |
| 4.5 | Conferir período de faturamento | Entrada em `billing_entries` / período aberto | ☐ | |

---

## Fase 5 — Faturamento e worker

| Passo | Ação | Resultado esperado | OK? | Evidência |
| ----- | ---- | ------------------ | --- | --------- |
| 5.1 | Painel → **Faturamento** → Fechar competência (mês da compra) | Fatura gerada | ☐ | invoice_id: |
| 5.2 | Worker processa outbox | E-mail da fatura no Mailpit / SMTP | ☐ | assunto: Fatura INV-… |
| 5.3 | Loja → **Faturas** | Fatura listada com total correto; vencimento dia 10 | ☐ | |
| 5.4 | Loja → **Fechar fatura e pagar** (com compras no ciclo) | Nova fatura + novo ciclo aberto; Pix disponível | ☐ | |
| 5.5 | (Opcional) Dia 1 com worker ativo | Fechamento automático do mês anterior | ☐ | log worker |
| 5.6 | (Opcional) Fatura fechada sem pagar após 48h/72h | E-mail lembrete / escalada `overdue` | ☐ | Mailpit |

---

## Fase 6 — Pix e webhook

Roteiro detalhado (MP + sandbox): plano em `.cursor/plans/` e [development/mercadopago-pix.md](development/mercadopago-pix.md). Script local: `./scripts/pix-homologation-check.sh` (health, rota webhook, testes Go sandbox).

**Pré-requisito loja (MP e sandbox):** após fechar a fatura, confirmar **plano 1×** na UI antes de **Gerar Pix** (fatura fechada cria plano `pending_selection`).

**Credenciais MP (API de Orders):** no `.env`, use **Testes → Credenciais de teste → Access Token** em `MERCADO_PAGO_ACCESS_TOKEN`, com `MERCADO_PAGO_ENVIRONMENT=test` e `MERCADO_PAGO_TEST_AUTO_APPROVE=true`. Produção só no go-live. Validar antes da loja:

```bash
./scripts/pix-homologation-check.sh --mp-auth-smoke    # opcional (200 não garante Orders)
./scripts/pix-homologation-check.sh --mp-orders-smoke  # obrigatório: HTTP 201
docker compose up -d --build api worker
```

Se o orders smoke ≠ 201, ver [mercadopago-pix.md — Troubleshooting 401](development/mercadopago-pix.md#troubleshooting-401-test-credentials).

| Passo | Ação | Resultado esperado | OK? | Evidência |
| ----- | ---- | ------------------ | --- | --------- |
| 6.1 | Plano 1× confirmado → **Gerar Pix** (`PAYMENT_PROVIDER=mercadopago`) | QR/código retornado; log `pix charge created` / `mercado pago api call` | ☐ | charge_id: |
| 6.2 | Sandbox: `PAYMENT_PROVIDER=sandbox` + `POST /dev/pix/simulate/{chargeId}` ou webhook | Pagamento confirmado uma vez | ☑ | `go test … -run PixWebhook` / `TestPixWebhookDuplicateIsIgnored` |
| 6.2b | MP teste + túnel HTTPS → webhook Order | `payment_events` + job `payments.mercadopago_order`; worker consulta Order; simulador **não** baixa sem accredited | ☐ | `./scripts/pix-homologation-check.sh --webhook-self-test` (200 local) + log `mercado pago webhook received` (POST real MP) |
| 6.2c | MP: fatura **R$ 50** à vista → Pix + APRO + worker + túnel | Order `processed/accredited`; fatura `paid`; ver checklist em [mercadopago-pix.md](development/mercadopago-pix.md) | ☐ | cenário documentado MP |
| 6.2d | MP parcelado (ex. R$ 350, 3×) após 6.2c | Parcelas sequenciais; opcional — doc MP não garante APRO fora do payload R$ 50 | ☐ | |
| 6.3 | Repetir mesmo webhook (teste técnico) | Sem duplicar liquidação (sandbox); MP: `duplicate=true` sem nova linha | ☑ | `TestPixWebhookDuplicateIsIgnored`; log MP `duplicate=true` |
| 6.4 | Fatura | Status pago ou `paid_cents` = total (sandbox) | ☑ | integração `billing_payments_test` |
| 6.5 | Exposição do cliente | Reduzida após pagamento (sandbox) | ☐ | conferir UI após 6.2 manual |

### Roteiro 6.2b–6.2c (Mercado Pago real)

1. `--mp-orders-smoke` → **HTTP 201** (Access Token de **teste** no `.env`; ver pré-requisito acima).
2. Túnel HTTPS (ex. Cloudflare) apontando para `localhost:8080`; URL no painel MP: `…/api/v1/webhooks/mercado-pago/orders` + `MERCADO_PAGO_WEBHOOK_SECRET`.
3. `docker compose up -d api worker` com `PAYMENT_PROVIDER=mercadopago`.
4. Loja: fechar fatura, plano **1×**, parcela em aberto → **Gerar Pix** (ideal total **R$ 50** para APRO).
5. UI: QR/copiar/vencimento; aguardar ~1 min ou admin `POST /api/v1/admin/payment-charges/{chargeId}/sync`.
6. Conferir: `payment_events`, job `payments.mercadopago_order`, logs worker `GET /v1/orders`, parcela/fatura **paga**.

**Evidência 6.2c (R$ 50,00)** — após Pix + baixa:

```sql
SELECT pc.external_id, pc.status, pc.amount_cents, i.invoice_number, i.status AS invoice_status, i.paid_cents
FROM payment_charges pc
JOIN invoices i ON i.id = pc.invoice_id
WHERE pc.provider = 'mercadopago' AND pc.amount_cents = 5000
ORDER BY pc.created_at DESC LIMIT 1;
```

Logs esperados: API `pix charge created` + `amount` coerente; worker `mercado pago payment settled` com `amount_cents":5000` e `invoice_paid":true`.

**Webhook antes do go MP:** `./scripts/pix-homologation-check.sh --webhook-self-test` (valida `.env` ↔ API). Notificações reais do painel MP exigem o **mesmo** `MERCADO_PAGO_WEBHOOK_SECRET` cadastrado na URL do túnel. Se o **simulador** do painel retorna 200 e os POSTs **automáticos** continuam em 401, ver [mercadopago-pix.md § Simulador manual vs automático](development/mercadopago-pix.md#simulador-manual-200-vs-notificações-automáticas-401) (secret pós-túnel, URLs duplicadas, `data.id` na query).

---

## Fase 7 — Relatórios e operação

| Passo | Ação | Resultado esperado | OK? | Evidência |
| ----- | ---- | ------------------ | --- | --------- |
| 7.1 | **Dashboard** (`/`) | Gráfico 6 meses (vendas vs recebimentos) e KPIs coerentes | ☐ | |
| 7.2 | **Relatórios** → cada subseção | Exportar CSV, Excel e PDF com filtros aplicados | ☐ | |
| 7.3 | **Relatórios → Estoque / Movimentações** | Posição e histórico consistentes com tela Estoque | ☐ | |
| 7.4 | **Relatórios → Contas a receber** | Faixas de atraso e saldos batem com Faturamento | ☐ | |
| 7.5 | **Relatórios → Pix / Limites / Exceções** | Conforme papel (financeiro vs estoque) | ☐ | |
| 7.6 | **Relatórios → Previsão** → Gerar snapshots | Linhas com vendas 3m e estoque atual | ☐ | |
| 7.7 | **Auditoria** (admin) | Filtros por período e listagem | ☐ | |
| 7.8 | `make backup` no servidor | Arquivo `.sql.gz` em `backups/` ou `/var/backups/store-platform` | ☐ | caminho: |
| 7.9 | `make test-backup-restore` ou restore em DB temporário | Restore OK | ☐ | |

---

## Fase 8 — Segurança e infra (staging)

| Passo | Ação | Resultado esperado | OK? | Evidência |
| ----- | ---- | ------------------ | --- | --------- |
| 8.1 | HTTPS no domínio da loja e do admin | Certificado válido (Caddy) | ☐ | |
| 8.2 | Cookies com `Secure` em produção | `COOKIE_SECURE=true` | ☐ | |
| 8.3 | `ADMIN_MFA_REQUIRED=true` | Painel bloqueado sem MFA | ☐ | |
| 8.4 | Portas DB/API não expostas publicamente | Apenas 80/443 no host | ☐ | |

---

## Registro da homologação

Preencha ao final:

- **Versão / commit:** `git rev-parse HEAD`
- **Ambiente:** staging / produção
- **Responsável:** 
- **Data:** 
- **Resultado:** Aprovado ☐ / Reprovado ☐
- **Observações:**

### Política de variáveis — staging vs produção (Pix MP)

| Variável | Dev local (Docker) | Staging UAT | Produção |
| -------- | ------------------ | ----------- | -------- |
| `PAYMENT_PROVIDER` | `mercadopago` para teste real | `mercadopago` | `mercadopago` |
| `MERCADO_PAGO_ENVIRONMENT` | `test` | `test` até go-live | `production` |
| `MERCADO_PAGO_TEST_AUTO_APPROVE` | `true` permitido (job ~12s sem webhook) | `true` só como **ponte** enquanto 6.2b não passa; depois `false` + webhook 200 | **`false`** (obrigatório; API recusa `true`) |
| `MERCADO_PAGO_WEBHOOK_SECRET` | Painel MP modo teste, URL do túnel | URL HTTPS estável do staging | Painel produção |

**Evidências mínimas (anexar ao registro):**

- Fase 0: saída de `make test`, `go test ./tests/e2e/... -run 'TestHTTP|TestMVP'`, `./scripts/pix-homologation-check.sh --mp-orders-smoke`
- Fase 6: log `mercado pago webhook received` (MP real) **ou** `--webhook-self-test` HTTP 200 + worker `mercado pago payment settled`
- Commit deployado em staging = mesmo que passou Fase 0

Itens fora do MVP atual (documentar como pendência, não reprovar se já conhecido):

- Pix com PSP real (produção financeira).
- MFA em app mobile (EP-13 em evolução).

---

---

## Fase — Parcelamento de faturas

| # | Passo | OK? |
| - | ----- | --- |
| P.1 | Admin: **Faturamento → Configuração de parcelamento** — conferir defaults (mín. R$ 300, parcela R$ 100, máx. 10×) | ☐ |
| P.2 | Fechar ciclo na competência: modal **só “Fechar fatura”** (sem Pix); na fatura, ver opções 1×…N× em cards | ☐ |
| P.3 | **Confirmar forma de pagamento** → modal **“Fechar fatura e pagar”** com resumo das parcelas; após OK, tabela/cards de parcelas; **Gerar Pix** só na parcela aberta (sem reabrir esse modal) | ☐ |
| P.4 | Simular pagamento (dev); fatura `partially_paid`; segunda parcela liberada | ☐ |
| P.5 | Desabilitar parcelamento no admin; nova fatura só oferece 1×; plano 3× em andamento continua | ☐ |

---

## Referência rápida — comandos

```bash
# Stack local completa
docker compose up -d --build

# Homologação automatizada (Fase 0)
make test && make test-backup-restore
cd backend && go test -p 1 ./tests/e2e/... -run 'TestHTTP|TestMVP'

# Pix / Mercado Pago
./scripts/pix-homologation-check.sh
./scripts/pix-homologation-check.sh --mp-orders-smoke
./scripts/pix-homologation-check.sh --webhook-self-test
```

Deploy no servidor: ver `docs/deployment.md`.
