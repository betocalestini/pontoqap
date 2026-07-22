# Pix Mercado Pago — desenvolvimento local

Guia para testar a integração real (credenciais de **teste**) com a API de Orders do Mercado Pago. A baixa automática da fatura via webhook MP ainda não está implementada; eventos são apenas **recebidos e gravados** em `payment_events`.

## Pré-requisitos no Mercado Pago

1. Criar aplicação em [Mercado Pago Developers](https://www.mercadopago.com.br/developers/panel/app): **Checkout Transparente**, **API de Orders** (ver também [mercadopag.md](../mercadopag.md)).
2. Copiar **Access Token de teste** e, se disponível, **Public Key** (não usada no fluxo Pix atual — criação só no backend).
3. Conta com **chave Pix** cadastrada (requisito do MP para Pix no Checkout Transparente).
4. Em **Webhooks** (modo teste): evento **Order (Mercado Pago)**; copiar o **segredo** gerado.

## Variáveis de ambiente

Na raiz do repositório, em `.env` (modelo em `.env.example`):

```bash
PAYMENT_PROVIDER=mercadopago   # só para gerar Pix real; webhook não exige isso
MERCADO_PAGO_ENVIRONMENT=test
MERCADO_PAGO_BASE_URL=https://api.mercadopago.com
MERCADO_PAGO_ACCESS_TOKEN=TEST-...
MERCADO_PAGO_WEBHOOK_SECRET=...   # obrigatório para aceitar o simulador (mesmo valor do painel)
MERCADO_PAGO_PIX_EXPIRATION=PT24H
MERCADO_PAGO_REQUEST_TIMEOUT_SECONDS=10
```

Mantenha `PAYMENT_PROVIDER=sandbox` para testes automatizados e fluxo com `/dev/pix/simulate` sem credenciais MP.

Reinicie a API após alterar o `.env`:

```bash
docker compose up -d --build api
# ou: make api
```

Se você usa Docker e acabou de puxar código novo, **rebuild** do serviço `api` é necessário; imagem antiga não expõe a rota do webhook MP e o simulador retorna **404**.

## Gerar Pix de teste

1. Suba o ambiente (`make dev-up` ou Postgres + `make api` + `pnpm dev:store`).
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
# {"status":"ok"}

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

Registros ficam com `processed = false` até a fase de processamento assíncrono e baixa (ver [pix.md](../pix.md)).

## O que esperar após HTTP 200 do Mercado Pago

O painel MP mostra **200 OK** quando a API aceitou a notificação (assinatura válida e JSON mínimo). Isso **não** significa que a fatura na loja foi paga.

| Efeito | Acontece hoje? |
|--------|----------------|
| Linha nova em `payment_events` (`provider = mercadopago`) | Sim, na primeira vez para cada `x-request-id` |
| Log na API (`docker compose logs -f api`) | Sim: `mercado pago webhook received` com `order_id`, `event_type`, `duplicate` |
| Fatura / parcela paga na UI | Não (worker de baixa ainda não implementado) |
| `payment_charges` atualizada | Não |

O **simulador** do painel envia dados fictícios (`id: "123456"`, etc.) — não estão ligados a uma cobrança Pix gerada na loja.

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
