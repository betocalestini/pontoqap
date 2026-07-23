#!/usr/bin/env bash
# Roteiro rápido — teste Pix (MP + sandbox). Ver docs/development/mercadopago-pix.md
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

MP_AUTH_SMOKE=false
MP_ORDERS_SMOKE=false
MP_WEBHOOK_SELF_TEST=false
for arg in "$@"; do
  case "$arg" in
    --mp-auth-smoke) MP_AUTH_SMOKE=true ;;
    --mp-orders-smoke) MP_ORDERS_SMOKE=true ;;
    --webhook-self-test) MP_WEBHOOK_SELF_TEST=true ;;
  esac
done

load_mp_webhook_secret() {
  if [[ -f .env ]]; then
    local line
    line=$(grep -E '^MERCADO_PAGO_WEBHOOK_SECRET=' .env | tail -n1 || true)
    if [[ -n "$line" ]]; then
      MERCADO_PAGO_WEBHOOK_SECRET="${line#MERCADO_PAGO_WEBHOOK_SECRET=}"
      MERCADO_PAGO_WEBHOOK_SECRET="${MERCADO_PAGO_WEBHOOK_SECRET#export }"
      MERCADO_PAGO_WEBHOOK_SECRET="${MERCADO_PAGO_WEBHOOK_SECRET%\"}"
      MERCADO_PAGO_WEBHOOK_SECRET="${MERCADO_PAGO_WEBHOOK_SECRET#\"}"
      MERCADO_PAGO_WEBHOOK_SECRET="${MERCADO_PAGO_WEBHOOK_SECRET%\'}"
      MERCADO_PAGO_WEBHOOK_SECRET="${MERCADO_PAGO_WEBHOOK_SECRET#\'}"
      export MERCADO_PAGO_WEBHOOK_SECRET
    fi
  fi
}

run_webhook_self_test() {
  load_mp_webhook_secret
  if [[ -z "${MERCADO_PAGO_WEBHOOK_SECRET:-}" ]]; then
    echo "ERRO: MERCADO_PAGO_WEBHOOK_SECRET vazio (.env)"
    exit 1
  fi
  echo "== Webhook self-test (HMAC com segredo do .env → API local) =="
  echo "   Se HTTP 200: segredo carregado na API bate com o .env (log esperado: mercado pago webhook received)"
  echo "   Se HTTP 401: rebuild api/worker ou confira env_file; MP real ainda exige o mesmo segredo do painel."
  (cd backend && go run ./cmd/mp-webhook-check/ -post)
}

load_mp_token() {
  if [[ -f .env ]]; then
    local line
    line=$(grep -E '^MERCADO_PAGO_ACCESS_TOKEN=' .env | tail -n1 || true)
    if [[ -n "$line" ]]; then
      MERCADO_PAGO_ACCESS_TOKEN="${line#MERCADO_PAGO_ACCESS_TOKEN=}"
      MERCADO_PAGO_ACCESS_TOKEN="${MERCADO_PAGO_ACCESS_TOKEN#export }"
      MERCADO_PAGO_ACCESS_TOKEN="${MERCADO_PAGO_ACCESS_TOKEN%\"}"
      MERCADO_PAGO_ACCESS_TOKEN="${MERCADO_PAGO_ACCESS_TOKEN#\"}"
      MERCADO_PAGO_ACCESS_TOKEN="${MERCADO_PAGO_ACCESS_TOKEN%\'}"
      MERCADO_PAGO_ACCESS_TOKEN="${MERCADO_PAGO_ACCESS_TOKEN#\'}"
      export MERCADO_PAGO_ACCESS_TOKEN
    fi
  fi
  if [[ -z "${MERCADO_PAGO_ACCESS_TOKEN:-}" ]]; then
    echo "ERRO: MERCADO_PAGO_ACCESS_TOKEN vazio (preencha .env)"
    exit 1
  fi
}

run_mp_auth_smoke() {
  load_mp_token
  echo "== MP auth smoke: GET /v1/payment_methods =="
  code=$(curl -sS -o /tmp/mp_pm.json -w "%{http_code}" \
    -H "Authorization: Bearer ${MERCADO_PAGO_ACCESS_TOKEN}" \
    "https://api.mercadopago.com/v1/payment_methods")
  echo "HTTP $code"
  if [[ "$code" != "200" ]]; then
    head -c 300 /tmp/mp_pm.json 2>/dev/null || true
    echo ""
    exit 1
  fi
  echo "OK (credencial aceita neste endpoint; não comprova POST /v1/orders)"
}

run_mp_orders_smoke() {
  load_mp_token
  echo "== MP orders smoke: POST /v1/orders (R\$ 50 APRO) =="
  echo "AVISO: este teste cria uma Order Pix no ambiente de testes do Mercado Pago."
  idem="mp-pix-smoke-$(date +%s)-$$"
  code=$(curl -sS -o /tmp/mp_order.json -w "%{http_code}" \
    --request POST \
    --url "https://api.mercadopago.com/v1/orders" \
    --header "Accept: application/json" \
    --header "Authorization: Bearer ${MERCADO_PAGO_ACCESS_TOKEN}" \
    --header "Content-Type: application/json" \
    --header "X-Idempotency-Key: ${idem}" \
    --data '{
      "type": "online",
      "external_reference": "mp-pix-smoke-001",
      "total_amount": "50.00",
      "processing_mode": "automatic",
      "payer": { "email": "test_user_br@testuser.com", "first_name": "APRO" },
      "transactions": {
        "payments": [{
          "amount": "50.00",
          "payment_method": { "id": "pix", "type": "bank_transfer" }
        }]
      }
    }')
  echo "HTTP $code (esperado 201)"
  if [[ "$code" != "201" ]]; then
    if command -v jq >/dev/null 2>&1; then
      jq -r '.errors[]? | "\(.code): \(.message)"' /tmp/mp_order.json 2>/dev/null | head -5
      jq -r '.message // empty' /tmp/mp_order.json 2>/dev/null | head -1
    else
      head -c 500 /tmp/mp_order.json
      echo ""
    fi
    if grep -q 'Test credentials are not supported' /tmp/mp_order.json 2>/dev/null; then
      echo ""
      echo "Dica: a doc Orders/Pix manda Access Token de TESTE; esta mensagem pode ser inconsistência da API MP."
      echo "  Checklist: docs/development/mercadopago-pix.md#troubleshooting-401-test-credentials"
      echo "  (token ≠ Public Key, app Checkout Transparente + Orders, rebuild api/worker)"
    fi
    exit 1
  fi
  if command -v jq >/dev/null 2>&1; then
    jq -r '.id // empty, .transactions.payments[0].id // empty' /tmp/mp_order.json 2>/dev/null | sed '/^$/d' | while read -r line; do
      echo "id: ${line:0:24}…"
    done
  else
    head -c 400 /tmp/mp_order.json
    echo ""
  fi
  echo "OK"
}

if [[ "$MP_WEBHOOK_SELF_TEST" == true ]]; then
  run_webhook_self_test
  exit 0
fi
if [[ "$MP_AUTH_SMOKE" == true ]]; then
  run_mp_auth_smoke
  exit 0
fi
if [[ "$MP_ORDERS_SMOKE" == true ]]; then
  run_mp_orders_smoke
  exit 0
fi

echo "== Health / Ready API =="
curl -sf "http://localhost:8080/health" | head -c 200 || { echo "API indisponível em :8080"; exit 1; }
echo ""
curl -sf "http://localhost:8080/ready" | head -c 200 || echo "(ready indisponível — rebuild api)"
echo ""

echo "== Webhook MP (esperado 401/400, não 404) =="
code=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
  "http://localhost:8080/api/v1/webhooks/mercado-pago/orders" \
  -H "Content-Type: application/json" \
  -d '{"type":"order","data":{"id":"check-1"}}')
echo "HTTP $code"
if [[ "$code" == "404" ]]; then
  echo "ERRO: rota não encontrada — rebuild: docker compose up -d --build api"
  exit 1
fi

if docker compose ps postgres 2>/dev/null | grep -q running; then
  echo "== payment_events (mercadopago) =="
  docker compose exec -T postgres psql -U store -d store -c \
    "SELECT external_event_id, event_type, processed, created_at FROM payment_events WHERE provider = 'mercadopago' ORDER BY created_at DESC LIMIT 5;" 2>/dev/null || true
  echo "== jobs MP pendentes =="
  docker compose exec -T postgres psql -U store -d store -c \
    "SELECT type, status, created_at FROM jobs WHERE type = 'payments.mercadopago_order' ORDER BY created_at DESC LIMIT 5;" 2>/dev/null || true
fi

echo ""
echo "== Testes automatizados (Pix / MP / parcelas) =="
(cd backend && go test ./internal/payments/... ./tests/integration/... \
  -run 'PixWebhook|BillingPayments|ParseMercadoPago|MercadoPago|MultiInstallment' -count=1)
echo "OK — MP: ./scripts/pix-homologation-check.sh --mp-auth-smoke | --mp-orders-smoke | --webhook-self-test"
echo "    liquidação completa local: PAYMENT_PROVIDER=sandbox + /dev/pix/simulate"
