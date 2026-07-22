#!/usr/bin/env bash
# Roteiro rápido — teste Pix (MP + sandbox). Ver docs/development/mercadopago-pix.md
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "== Health API =="
curl -sf "http://localhost:8080/health" | head -c 200 || { echo "API indisponível em :8080"; exit 1; }
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
fi

echo ""
echo "== Testes automatizados (sandbox Pix) =="
(cd backend && go test ./internal/payments/... ./tests/integration/... -run 'PixWebhook|BillingPayments|ParseMercadoPago' -count=1)
echo "OK — para Pix MP na UI: PAYMENT_PROVIDER=mercadopago + fatura com plano 1× confirmado"
