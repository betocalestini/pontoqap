# Pix Mercado Pago — desenvolvimento local

Guia para testar a integração real com a API de Orders do Mercado Pago. O webhook valida assinatura, grava `payment_events`, enfileira o job `payments.mercadopago_order` e o **worker** consulta `GET /v1/orders/{id}` antes de qualquer baixa (mesmo fluxo de `ApplyInstallmentPaymentTx` do sandbox).

## Pré-requisitos no Mercado Pago

1. Criar aplicação em [Mercado Pago Developers](https://www.mercadopago.com.br/developers/panel/app): **Pagamentos online**, loja própria, **Checkout Transparente**, **API de Orders** (ver também [mercadopag.md](../mercadopag.md)).
2. **Access Token para o backend (`POST /v1/orders`):** em homologação use **Testes → Credenciais de teste → Access Token** (não a Public Key). A [doc Pix Orders em teste](https://www.mercadopago.com.br/developers/pt/docs/checkout-api-orders/integration-test/pix) e a [notícia de credenciais automáticas (nov/2025)](https://www.mercadopago.com.br/developers/pt/news/2025/11/19/Streamlined-integration-testing-with-automatic-credentials) orientam desenvolvimento com credenciais de teste no endpoint `https://api.mercadopago.com/v1/orders`.
3. Homologação: `MERCADO_PAGO_ENVIRONMENT=test`, `MERCADO_PAGO_TEST_AUTO_APPROVE=true`, pagador de teste (e-mail `test_user_br@testuser.com`, `first_name` **APRO** — aplicado no servidor em [`payer.go`](../../backend/internal/payments/mercadopago/payer.go)).
4. Conta com **chave Pix** cadastrada (requisito do MP para Pix no Checkout Transparente).
5. Em **Webhooks** (modo teste): evento **Order (Mercado Pago)**; copiar o **segredo** gerado.

### Qual credencial em cada fase

| Situação | Credencial no `.env` |
| --- | --- |
| Desenvolvimento / homologação local | **Testes → Access Token de teste** |
| Order Pix com payer **APRO** | Mesmo Access Token de teste |
| `GET /v1/orders/{id}` da Order simulada | Mesmo Access Token de teste |
| Medição de qualidade (painel MP, modo teste) | Order criada com credenciais de teste |
| Pagamento real (go-live) | **Produção → Access Token** + `MERCADO_PAGO_ENVIRONMENT=production` + `MERCADO_PAGO_TEST_AUTO_APPROVE=false` |

A **Public Key de teste** é para integrações com SDK no frontend; este projeto usa Pix **só no backend** — não é necessário `MERCADO_PAGO_PUBLIC_KEY` na loja atual.

## Variáveis de ambiente

Na raiz do repositório, em `.env` (modelo em `.env.example`):

```bash
PAYMENT_PROVIDER=mercadopago   # alias aceito: mercado_pago
MERCADO_PAGO_ENVIRONMENT=test
MERCADO_PAGO_BASE_URL=https://api.mercadopago.com
MERCADO_PAGO_ACCESS_TOKEN=...   # Testes → Credenciais de teste → Access Token (homologação)
MERCADO_PAGO_WEBHOOK_SECRET=...   # obrigatório para aceitar o simulador (mesmo valor do painel)
MERCADO_PAGO_PIX_EXPIRATION=PT24H
MERCADO_PAGO_REQUEST_TIMEOUT_SECONDS=10
# Aprovação automática Pix em teste (documentação MP — payer APRO). Obrigatório false em produção.
MERCADO_PAGO_TEST_AUTO_APPROVE=true
```

No painel MP, copie o **Access Token** da aba **Testes → Credenciais de teste** para homologação. O prefixo pode ser `APP_USR-` ou `TEST-` conforme a aplicação; o backend **não** valida pelo prefixo. `MERCADO_PAGO_ENVIRONMENT=test` controla payer APRO e flags internas — **não** substitui o uso do token de teste no painel. Confirme com `--mp-orders-smoke` (**HTTP 201**); `GET /v1/payment_methods` sozinho não basta.

**Escopo:** integração **Checkout Transparente + API de Orders** + Pix. Não use a coleção Postman de **Payments legacy** (`/v1/payments`) como referência deste fluxo.

### Diagnóstico em dois níveis

| Teste | Endpoint | O que comprova |
| --- | --- | --- |
| Auth smoke | `GET /v1/payment_methods` | Bearer aceito **nesse** endpoint |
| Orders smoke | `POST /v1/orders` | Integração Pix Orders (prova definitiva) |

`GET /v1/payment_methods` com **200** não garante `POST /v1/orders`. Sucesso na criação da Order = **HTTP 201** ([referência](https://www.mercadopago.com.br/developers/pt/reference/online-payments/checkout-api/create-order/post)).

Scripts locais:

```bash
./scripts/pix-homologation-check.sh --mp-auth-smoke
./scripts/pix-homologation-check.sh --mp-orders-smoke   # cria Order de teste no MP
```

#### Tabela de interpretação

| payment_methods | orders | Interpretação |
| --- | --- | --- |
| 401 | — | Token ausente, incorreto ou header inválido |
| 200 | 401 | Ver [Troubleshooting: 401 Test credentials](#troubleshooting-401-test-credentials) e corpo JSON (`errors`, `message`) |
| 200 | 403 | Acesso recusado; analisar corpo, produto, conta e permissões |
| 200 | 400 | Payload, e-mail, valor ou idempotência |
| 200 | 409 | `X-Idempotency-Key` reutilizada de forma incompatível |
| 200 | 429 | Rate limit |
| 200 | 201 | Order criada |
| 200 | 402 | Order criada com falha em transação |

Na loja, falha ao **Gerar Pix** ocorre antes de existir Order no MP (`ORD…`). O registro em `payment_charges` só é criado após sucesso do gateway ([`service.go`](../../backend/internal/payments/service.go)). Logs da API incluem `mp_http_status`, `mp_error_code`, `mp_error_message`, `mp_request_id` (sem token nem QR completo).

### Validar credencial (curl)

Auth smoke (`GET`):

```bash
curl -i -H "Authorization: Bearer $MERCADO_PAGO_ACCESS_TOKEN" \
  "https://api.mercadopago.com/v1/payment_methods"
```

- **HTTP 200** — token reconhecido neste endpoint (não comprova Orders).
- **HTTP 401** — token inválido, incompleto ou copiado errado (confira Access Token vs Public Key, aspas/espaços no `.env`, rebuild `api`/`worker`).
- **HTTP 403** — token reconhecido, sem permissão para aquele recurso.

Orders smoke (`POST` — esperado **201**):

```bash
curl -i --request POST \
  --url "https://api.mercadopago.com/v1/orders" \
  --header "Accept: application/json" \
  --header "Authorization: Bearer $MERCADO_PAGO_ACCESS_TOKEN" \
  --header "Content-Type: application/json" \
  --header "X-Idempotency-Key: $(uuidgen)" \
  --data '{
    "type": "online",
    "external_reference": "pix-local-test-001",
    "total_amount": "50.00",
    "processing_mode": "automatic",
    "payer": { "email": "test_user_br@testuser.com", "first_name": "APRO" },
    "transactions": {
      "payments": [{
        "amount": "50.00",
        "payment_method": { "id": "pix", "type": "bank_transfer" }
      }]
    }
  }'
```

Se `payment_methods` for 200 mas `orders` não for **201**, confira o corpo JSON, se a aplicação é **Checkout Transparente + API de Orders** ([criar aplicação](https://www.mercadopago.com.br/developers/pt/docs/checkout-api-orders/create-application)) e a seção [Troubleshooting](#troubleshooting-401-test-credentials). Homologação deve incluir reconciliação sem webhook (`POST /api/v1/admin/payment-charges/{id}/sync` ou worker com `GET /v1/orders/{id}`).

### Troubleshooting: 401 Test credentials {#troubleshooting-401-test-credentials}

A referência genérica da API ainda pode devolver `invalid_credentials` com a mensagem *Test credentials are not supported…* ([exemplo na referência](https://www.mercadopago.com.br/developers/pt/reference/online-payments/checkout-api/get-order/get)). Isso **conflita** com a documentação específica de [Checkout Transparente + Orders + Pix em teste](https://www.mercadopago.com.br/developers/pt/docs/checkout-api-orders/integration-test/pix), que manda usar **Access Token de teste**. Para este projeto, a política é: **homologação com credenciais de teste**; produção só no go-live.

Quando `--mp-orders-smoke` retornar **401** com essa mensagem:

1. Confirmar **Access Token** (não Public Key), sem espaços ou aspas quebradas no `.env` (ex.: `SMTP_FROM` com `<` precisa de aspas).
2. Token da **mesma aplicação** Checkout Transparente + API de Orders (não app legada só `/v1/payments`).
3. Aplicação criada ou atualizada após nov/2025 (credenciais de teste automáticas) — em contas antigas, considerar nova aplicação no painel.
4. Headers na requisição: `Authorization`, `Accept: application/json`, `Content-Type`, `X-Idempotency-Key` (o backend e o script de smoke já enviam).
5. `docker compose up -d --build api worker` após alterar `.env`.
6. Anotar `x-request-id` da resposta e o corpo JSON para suporte MP se persistir.
7. **Último recurso** (não é o padrão do projeto): testar Access Token de **produção** apenas para destravar homologação em conta que ainda rejeita teste na API — não use para simular Pix real no banco; para go-live use produção com `MERCADO_PAGO_ENVIRONMENT=production` e `MERCADO_PAGO_TEST_AUTO_APPROVE=false`.

Conferir variável no container (sem expor o token):

```bash
docker compose exec api sh -c 'case "$MERCADO_PAGO_ACCESS_TOKEN" in TEST-*) echo "Token TEST carregado";; APP_USR-*) echo "Token APP_USR carregado";; "") echo "Token vazio";; *) echo "Outro prefixo";; esac'
```

Mantenha `PAYMENT_PROVIDER=sandbox` para testes automatizados e fluxo com **Simular pagamento (dev)** na loja (`POST /dev/pix/simulate`). Esse botão **não funciona** com `PAYMENT_PROVIDER=mercadopago` — a baixa passa por webhook + worker (ou sync admin).

Reinicie **api e worker** após alterar o `.env`:

```bash
docker compose up -d --build api worker
```

Se você usa Docker e acabou de puxar código novo, **rebuild** do serviço `api` é necessário; imagem antiga não expõe a rota do webhook MP e o simulador retorna **404**.

## Teste de aprovação automática (APRO)

Com `MERCADO_PAGO_TEST_AUTO_APPROVE=true` e `MERCADO_PAGO_ENVIRONMENT=test`, o backend substitui o pagador **somente no servidor** ([`payer.go`](../../backend/internal/payments/mercadopago/payer.go)) — o frontend **nunca** envia `APRO`. Em `APP_ENV=production`, `MERCADO_PAGO_TEST_AUTO_APPROVE=true` impede a subida da API/worker.

```json
"payer": {
  "email": "test_user_br@testuser.com",
  "first_name": "APRO"
}
```

Referência MP: [integração Pix Orders em teste](https://www.mercadopago.com.br/developers/pt/docs/checkout-api-orders/integration-test/pix).

**Importante:** neste teste você **não** paga o QR em app bancário, **não** usa conta compradora e **não** simula Pix manual. O MP aprova sozinho após criar a Order com `APRO`.

Com `MERCADO_PAGO_TEST_AUTO_APPROVE=true`, após **Gerar Pix** a API agenda um job (~12s) para `GET /v1/orders/{id}` e baixar a parcela **sem depender só do webhook** (útil quando o túnel ou a assinatura falham). O webhook continua sendo o caminho preferido em homologação completa.

### Identificadores: fatura `INV-…` vs Mercado Pago

O número da fatura na loja (ex. `INV-202607-c1-f9245c83`) é o `invoice_number` — **só para exibição e relatórios**. O Mercado Pago **não** recebe esse valor no fluxo com **plano de parcelas**.

| Camada | Campo | Exemplo |
| --- | --- | --- |
| Loja | `invoices.invoice_number` | `INV-202607-c1-f9245c83` |
| MP `external_reference` (Pix na parcela) | `INSTALLMENT-{uuid da parcela}` | `INSTALLMENT-5deb16ee-b973-45f4-8840-04c48561f4a1` |
| MP Order / webhook `data.id` | id da Order | `ORDTST01KY80J7JH50E17Q2K2N300S6V` |
| Banco | `payment_charges.external_id` | mesmo `ORDTST…` |

Consulta útil no Postgres:

```sql
SELECT pc.id AS charge_id, pc.external_id AS mp_order_id, pc.status,
       ii.id AS installment_id,
       'INSTALLMENT-' || ii.id::text AS mp_external_reference
FROM payment_charges pc
JOIN invoices i ON i.id = pc.invoice_id
LEFT JOIN invoice_installments ii ON ii.id = pc.installment_id
WHERE i.invoice_number = 'INV-202607-c1-f9245c83'
ORDER BY pc.created_at DESC LIMIT 1;
```

O webhook e a baixa usam o **Order ID** (`ORD…`), não o `INV-…`.

### Webhook: assinatura (`invalid_signature`)

Se os logs da API mostram `mercado pago webhook rejected` / `invalid_signature` ou `mercado pago webhook signature failed` com `mp_signature_reason`:

1. **`MERCADO_PAGO_WEBHOOK_SECRET`** no `.env` deve ser **idêntico** ao segredo do painel MP (modo teste, evento **Order**, mesma URL do túnel). Não use `PAYMENT_WEBHOOK_SECRET` (é do sandbox interno). **Não confunda com o Access Token** — o token pode estar certo (`POST /v1/orders` 201) e o webhook ainda falhar com `SignatureMismatch`.
2. Se o log de subida mostra `webhook_secret_configured: true` mas continua `SignatureMismatch`, o valor no `.env` **não é** o segredo que o MP usa para assinar notificações dessa URL — copie de novo no painel após mudar o túnel (sem aspas nem espaços no fim; o backend faz `TrimSpace`).
3. URL cadastrada: `https://<túnel>/api/v1/webhooks/mercado-pago/orders` — o túnel deve **preservar** o query parameter `data.id` que o MP envia (a assinatura HMAC inclui esse valor).
4. Logs incluem `data_id_query` (URL) e `data_id_used_for_signature` (variante tentada no HMAC). O backend tenta **minúsculas** (doc Orders) e, se falhar, o **casing original** (comportamento reportado no sdk-go). Se ambos falham com `SignatureMismatch`, o problema é o **segredo do painel** (não o túnel).
5. Reinicie `api` e `worker` após alterar o `.env`: `docker compose up -d --build api worker`.

**Docker Compose:** `MERCADO_PAGO_ACCESS_TOKEN` e `MERCADO_PAGO_WEBHOOK_SECRET` entram pelos containers via `env_file: .env` (não são redefinidos com `${VAR:-}` no `compose.yaml`, para um `export` vazio no shell não apagar o segredo). Confira sem expor o valor:

```bash
docker compose ps worker
docker compose exec api sh -c 'test -n "$MERCADO_PAGO_WEBHOOK_SECRET" && echo WEBHOOK_SECRET=set || echo WEBHOOK_SECRET=empty'
docker compose exec worker sh -c 'test -n "$MERCADO_PAGO_ACCESS_TOKEN" && echo ACCESS_TOKEN=set || echo ACCESS_TOKEN=empty'
```

**Self-test local (evidência parcial 6.2b):** valida que o segredo do `.env` é o mesmo que a API usa para HMAC (não substitui o POST real do MP):

```bash
./scripts/pix-homologation-check.sh --webhook-self-test
# Esperado: HTTP 200 e log mercado pago webhook received
```

Se o self-test passa mas o MP continua com `SignatureMismatch`, o túnel **não é o problema** (os POSTs já chegam na API). O self-test só prova `.env` ↔ container; o POST real prova `.env` ↔ **secret que o MP usou para assinar**.

Checklist Cloudflare + painel:

1. No painel MP, aba **Webhooks**, modo **teste** (não produção), tópico **Order (Mercado Pago)**.
2. URL exatamente: `https://<seu-túnel>/api/v1/webhooks/mercado-pago/orders` (sem barra extra no fim).
3. Clique **revelar/copiar** o secret **dessa** configuração (após salvar a URL do túnel atual — trocar o hostname do cloudflared regenera ou exige secret novo).
4. Cole em `MERCADO_PAGO_WEBHOOK_SECRET` (64 hex típico), sem aspas; `docker compose up -d --build api worker`.
5. No painel, use **Simular** nessa URL: se a API logar `mercado pago webhook received` (HTTP 200), o secret está certo; se ainda 401, o valor no `.env` não é o daquela linha do painel (URL duplicada / app errada / modo produção vs teste).
6. Remova URLs antigas de outros túneis ou tópicos `payment` legados na mesma aplicação.

O Access Token correto (`POST /v1/orders` 201) **não** implica webhook secret correto — são credenciais diferentes.

#### Simulador manual 200 vs notificações automáticas 401

São evidências diferentes:

| Fonte | O que comprova |
| --- | --- |
| `--webhook-self-test` | O `.env` é o segredo que a API usa no HMAC |
| Simulador do painel (200 + `mercado pago webhook received`) | A URL **Order** atual e o secret **dessa** URL assinam um POST aceito naquele momento |
| POSTs automáticos do MP (`SignatureMismatch`, HTTP 401) | O HMAC do POST espontâneo não bate com o manifest / secret usados na validação |

**Importante:** simulador **200** com `data.id` numérico (`"123456"`) **não** prova que POSTs reais com `ORDTST…` vão passar. O `application_id` do JSON do simulador (ex. `1671…`) é payload de exemplo — use o **N.º da aplicação** das credenciais de teste e o `webhook_body_application_id` do POST automático.

O backend já tenta `data.id` em minúsculas e no casing original, e também omite `request-id` se necessário. Cloudflared (quick tunnel) **não** corrompe headers nos logs observados; o hostname volátil só exige recopiar URL/secret quando muda.

##### Forense (determinístico)

Quando simulador = 200, app do body = `MERCADO_PAGO_APPLICATION_ID`, e automático ainda dá 401:

1. No `.env`: `MERCADO_PAGO_WEBHOOK_DEBUG=true` (proibido em `APP_ENV=production`).
2. `docker compose up -d --build api` e gere um Pix (POST automático).
3. Log `mercado pago webhook debug capture` + arquivo `/tmp/mp-webhook-last.json` **dentro do container**:
   ```bash
   docker compose exec api cat /tmp/mp-webhook-last.json > /tmp/mp-webhook-last.json
   ```
4. Verificar offline contra o secret do `.env`:
   ```bash
   cd backend && set -a && source ../.env && set +a
   go run ./cmd/mp-webhook-check -verify-capture=/tmp/mp-webhook-last.json
   ```
5. Interpretação:
   - `verdict: NONE` → o secret do `.env` **não** assina esse POST (ops: na app do `webhook_body_application_id`, modo **teste**, **Reset** da assinatura secreta, recopiar, remover URL duplicada em modo produção, rebuild).
   - `verdict: HIT` → anotar a variante `data_id`/`request_id` e ajustar `validateMercadoPagoWebhookSignature`.

Checklist quando manual = 200 e automático = 401:

1. Rodar forense acima.
2. Uma única URL modo teste: `https://<túnel>/api/v1/webhooks/mercado-pago/orders`, tópico **Order** — remova a mesma URL em **produção** e túneis antigos.
3. Após trocar o túnel: reset/copiar secret → `.env` → `docker compose up -d --build api worker`.
4. Liquidação em staging com `MERCADO_PAGO_TEST_AUTO_APPROVE=true` pode ocorrer via job + `GET /v1/orders/{id}` **mesmo com webhooks em 401**; o 401 impede só webhook → `payment_events`.

### Validade do QR Pix vs vencimento da fatura

- `MERCADO_PAGO_PIX_EXPIRATION` (padrão `PT24H`) define a expiração do Pix na Order do MP, gravada em `payment_charges.expires_at`. **Não** é o vencimento da fatura (`invoices.due_at` / `due_date` da parcela).
- Enquanto existir cobrança `pending` com `expires_at > now()`, `POST .../pix-charge` **reutiliza** o mesmo QR (incluindo parcela em `pix_active`).
- Após expirar o Pix no MP, cobranças `pending` vencidas são marcadas `expired` e uma nova chamada gera nova Order (nova idempotency).
- Na loja, ao reabrir a fatura com parcela `pix_active`, a UI tenta recuperar o Pix pendente automaticamente (mesmo endpoint `POST .../pix-charge`).

### Worker: job `payments.mercadopago_order` falha na baixa

Com `MERCADO_PAGO_TEST_AUTO_APPROVE=true`, a API agenda reconciliação ~12s após gerar Pix. O **worker** deve estar `Up` (`docker compose ps worker`).

Se o log do worker mostra `GET /v1/orders/ORD…` **200** seguido de `job failed` com `no unique or exclusion constraint matching the ON CONFLICT specification`, a liquidação SQL falhou — atualize o backend (correção no `INSERT` em `payments` alinhado ao índice parcial da migration `000013`) e faça rebuild de `api` + `worker`. Atalho: `POST /api/v1/admin/payment-charges/{charge_id}/sync` após o fix.

O **simulador** do painel com `data.id` fictício (ex. `123456`) não liquida cobranças reais — use o Order ID (`ORD…`) da parcela ou `POST /api/v1/admin/payment-charges/{charge_id}/sync`.

### Credenciais na tela do painel

| Campo no painel | Uso neste projeto |
| --- | --- |
| Public Key | Frontend MP (não usada no Pix backend atual); pode aparecer como `TEST-…` |
| Access Token | `MERCADO_PAGO_ACCESS_TOKEN` — **Testes → Credenciais de teste** (`TEST-…` ou `APP_USR-…`) |
| Webhook secret | `MERCADO_PAGO_WEBHOOK_SECRET` |

Use `MERCADO_PAGO_ENVIRONMENT=test` em homologação; produção usa `production` e `MERCADO_PAGO_TEST_AUTO_APPROVE=false`.

### Cenário oficial recomendado: fatura de R$ 50,00 à vista

A documentação MP descreve o fluxo com Order de **R$ 50,00**. No sistema, faturas abaixo do mínimo de parcelamento (ex. R$ 300) continuam com **plano 1×** — uma parcela de R$ 50,00.

1. Cliente de teste → compras/pedidos totalizando **R$ 50,00** no ciclo.
2. Admin: fechar competência → fatura em `open`.
3. Loja: confirmar plano **1×** → parcela `1/1` em `open`, valor `5000` centavos.
4. **Gerar Pix** na parcela → `POST /api/v1/me/installments/{installmentId}/pix-charge`.
5. Backend: `POST /v1/orders` com `Authorization`, `X-Idempotency-Key`, `processing_mode: automatic`, Pix `bank_transfer`, payer APRO (se flags ativas), `external_reference`: `INSTALLMENT-{uuid}`.
6. UI: QR / copia-e-cola (sem prefixo `SBX`). No MP, estado inicial típico: `action_required` / `waiting_transfer`.
7. Aguardar aprovação automática → webhook Order (túnel HTTPS) → job `payments.mercadopago_order` → worker **`GET /v1/orders/{id}`** → só então baixa (não liquida só pelo webhook).

Estados no **Mercado Pago** para considerar pago: `status = processed`, `status_detail = accredited` ([status da Order](https://www.mercadopago.com.br/developers/pt/docs/checkout-api-orders/payment-management/status/order-status)).

Estados no **banco** após baixa bem-sucedida (R$ 50 à vista):

| Entidade | Valores esperados |
| --- | --- |
| `payment_charges` | `status = paid`, `external_id` = Order MP, `amount_cents = 5000`, `paid_at` preenchido |
| `payments` | `provider = mercadopago`, `status = settled`, `amount_cents = 5000`, uma linha |
| `invoice_installments` | `status = paid`, `paid_cents = 5000` |
| `invoice_payment_plans` | `status = completed` |
| `invoices` | `status = paid`, `paid_cents = 5000` |
| Cliente | `current_exposure_cents` reduz R$ 50 |

Antes da aprovação: cobrança `pending`, parcela `pix_active`, fatura ainda `open` — sem linha `payments` liquidada.

### Parcelamento maior (ex. R$ 350 em 3×)

A doc MP **não garante** `APRO` para valores arbitrários como critério principal. Depois do cenário de R$ 50:

- **MP:** repetir Pix + APRO parcela a parcela (experimental).
- **Regras de parcelas do sistema:** validar com `PAYMENT_PROVIDER=sandbox` + simular ou com MP real após o teste oficial.

Reconciliação se o túnel cair: `POST /api/v1/admin/payment-charges/{chargeId}/sync` (`payments.read`).

### Idempotência

Reenviar o mesmo webhook ou chamar sync de novo não deve duplicar `payments` nem somar `paid_cents` duas vezes. Log esperado: `duplicate=true` no webhook quando o `x-request-id` já foi visto.

### Checklist homologação MP (APRO)

```text
[ ] Access Token da aba Testes (não confundir com Public Key)
[ ] MERCADO_PAGO_ENVIRONMENT=test e MERCADO_PAGO_TEST_AUTO_APPROVE=true (só dev)
[ ] api + worker com mesmo .env
[ ] Fatura de teste R$ 50,00, plano 1×, parcela open
[ ] Gerar Pix → Order 2xx, QR exibido
[ ] Não pagar QR no banco
[ ] Webhook Order no túnel, x-signature OK, resposta 200
[ ] GET Order → processed / accredited
[ ] external_reference = INSTALLMENT-{uuid} da parcela
[ ] Uma cobrança, um payment settled, fatura paid
[ ] Reenvio webhook → sem duplicar liquidação
```

Fluxo resumido (APRO):

1. **Gerar Pix** na parcela em aberto → QR exibido; estado inicial `action_required` / `waiting_transfer` no MP.
2. **Não** pagar o QR em banco real — o MP aprova automaticamente após alguns segundos.
3. Webhook Order → worker → `GET /v1/orders/{id}` → pagamento `processed` / `accredited` → baixa da parcela.
4. Para múltiplas parcelas, repetir após cada quitação (priorize validar R$ 50 antes).

### Checklist Docker (MP APRO)

| Serviço | Obrigatório | Observação |
| --- | --- | --- |
| `postgres` + `migrate` | Sim | |
| `api` | Sim | `env_file: .env`, `PAYMENT_PROVIDER=mercadopago` |
| `worker` | Sim | Mesmas variáveis MP que a API; sem worker o job não baixa a parcela |
| `store-web` | Sim | QR **sem** `SBX` no copia-e-cola = Order real MP |
| Túnel + webhook Order | Sim em localhost | Botão **Simular pagamento** na loja não substitui este fluxo |

### Mapeamento de status (checklist → banco)

| Checklist | Sistema |
| --- | --- |
| `provider_order_id` | `payment_charges.external_id` |
| `provider_payment_id` | `payments.external_payment_id` |
| Cobrança ativa | `payment_charges.status = pending` |
| Cobrança paga | `payment_charges.status = paid` |
| Parcela com Pix | `invoice_installments.status = pix_active` |
| Pagamento confirmado | `payments.status = settled` |
| Provider | `mercadopago` |

## Gerar Pix de teste

1. Suba o ambiente (`make dev-up` ou Postgres + **api + worker** + `pnpm dev:store`).
2. Entre como cliente demo (ex.: `demo-cliente-001@demo.loja.local` / `DemoStore123!`).
3. Abra uma fatura em aberto → **confirme o plano de pagamento (ex. 1×)** se a UI solicitar — faturas fechadas criam plano em `pending_selection` e o Pix da fatura inteira só é permitido após confirmação (parcela única: Pix na parcela em aberto).
4. **Gerar Pix** (UI ou `POST /api/v1/me/invoices/{id}/pix-charge` / `…/installments/{id}/pix-charge`).
5. Confira o QR Code / copia e cola na tela.

Verificação automatizada (sandbox, liquidação): `./scripts/pix-homologation-check.sh` na raiz do repo.

O backend chama `POST https://api.mercadopago.com/v1/orders` com e-mail do cliente e referência externa (`invoice_number` ou `INSTALLMENT-{uuid}` para parcelas).

## Webhook local (HTTPS)

O Mercado Pago não envia webhooks para `http://localhost`. Aponte o túnel para a **API na porta 8080** (não use só a loja em 5173).

### Cloudflare Tunnel (rápido)

```bash
cloudflared tunnel --url http://localhost:8080
```

Anote a URL `https://….trycloudflare.com` exibida no terminal. Ela **muda** cada vez que você reinicia o `cloudflared`.

### ngrok

```bash
ngrok http 8080
```

### URL no painel Mercado Pago (modo teste, evento Order)

Cadastre o **caminho completo**, não apenas o host do túnel:

```text
https://<seu-túnel>/api/v1/webhooks/mercado-pago/orders
```

Exemplo:

```text
https://node-you-chief-elementary.trycloudflare.com/api/v1/webhooks/mercado-pago/orders
```

Erro comum: cadastrar só `https://….trycloudflare.com` → o MP faz `POST` na raiz `/` e a API responde **404 Not Found**.

### Verificar antes do simulador

```bash
curl -s http://localhost:8080/health
curl -s http://localhost:8080/ready
# {"status":"ok"} / {"status":"ready","db":"ok"}

curl -s -o /dev/null -w "%{http_code}\n" -X POST \
  http://localhost:8080/api/v1/webhooks/mercado-pago/orders \
  -H "Content-Type: application/json" \
  -d '{"type":"order","data":{"id":"123456"}}'
# esperado: 401 ou 400 — NÃO 404 (401 = rota ok, falta assinatura válida)
```

Substitua `localhost:8080` pelo host HTTPS do túnel para testar de fora.

Use o **simulador de webhooks** do painel para o tópico Order. Com segredo correto no `.env` e assinatura do MP, resposta esperada: HTTP **200** com `{"status":"ok"}`.

Verifique no banco:

```sql
SELECT provider, external_event_id, event_type, processed, created_at
FROM payment_events
WHERE provider = 'mercadopago'
ORDER BY created_at DESC
LIMIT 10;
```

Registros entram com `processed = false` até o worker processar o job. Após processamento sem liquidação (Order pendente ou simulador), `processed = true` sem linha em `payments`.

### Mapeamento checklist → banco

| Checklist | Coluna / valor no sistema |
| --- | --- |
| `provider_order_id` | `payment_charges.external_id` |
| `MERCADO_PAGO` | `payment_charges.provider = mercadopago` |
| Status cobrança ativa | `pending` (expirada: nova cobrança permitida) |
| `last_synced_at` | `payment_charges.last_synced_at` (após GET Order) |

Sincronização manual (admin): `POST /api/v1/admin/payment-charges/{id}/sync` (`payments.read`).

## O que esperar após HTTP 200 do Mercado Pago

O painel MP mostra **200 OK** quando a API aceitou a notificação (assinatura válida e JSON mínimo). A baixa na loja só ocorre se o worker confirmar a Order como **processed/accredited** com valor e referência corretos.

| Efeito | Acontece hoje? |
|--------|----------------|
| Linha nova em `payment_events` (`provider = mercadopago`) | Sim, na primeira vez para cada `x-request-id` |
| Job `payments.mercadopago_order` | Sim (requer **worker** ativo) |
| Log na API | `mercado pago webhook received` |
| GET Order + `last_synced_at` | Sim (worker ou sync admin) |
| Fatura / parcela paga | Somente Order **accredited** e valor/ref OK |
| Simulador MP com `id` fictício | Evento processado; **sem** baixa (Order não casa com cobrança ou fica pendente) |

O **simulador** do painel envia `data.id` fictício — use para validar túnel e assinatura; liquidação completa em dev: `PAYMENT_PROVIDER=sandbox` + `POST /dev/pix/simulate/{chargeId}`.

### Ver no banco (Postgres do Docker)

```bash
docker compose exec postgres psql -U store -d store -c "
SELECT external_event_id, event_type, processed, created_at
FROM payment_events
WHERE provider = 'mercadopago'
ORDER BY created_at DESC
LIMIT 5;"
```

### Ver nos logs da API

Após rebuild com código recente, cada webhook aceito gera algo como:

```text
level=INFO msg="mercado pago webhook received" order_id=... event_type=order.processed duplicate=false
```

Reenvio com o mesmo `x-request-id` retorna **200** de novo, com `duplicate=true` e sem nova linha no banco.

## Voltar ao sandbox interno

```bash
PAYMENT_PROVIDER=sandbox
```

Confirmação de pagamento em dev: `POST /api/v1/dev/pix/simulate/{chargeId}` (somente `APP_ENV=development`).

## Referência

- Plano completo: [pix.md](../pix.md)
- Criação de aplicação e credenciais: [mercadopag.md](../mercadopag.md)
