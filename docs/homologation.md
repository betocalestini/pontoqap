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

| Passo | Ação | Resultado esperado | OK? | Evidência |
| ----- | ---- | ------------------ | --- | --------- |
| 6.1 | Detalhe da fatura → **Gerar Pix** | QR/código retornado | ☐ | charge_id: |
| 6.2 | Staging sandbox: simular pagamento (dev) ou webhook PSP | Pagamento confirmado uma vez | ☐ | |
| 6.3 | Repetir mesmo webhook (teste técnico) | Sem duplicar liquidação | ☐ | log / paid_cents |
| 6.4 | Fatura | Status pago ou `paid_cents` = total | ☐ | |
| 6.5 | Exposição do cliente | Reduzida após pagamento | ☐ | |

---

## Fase 7 — Relatórios e operação

| Passo | Ação | Resultado esperado | OK? | Evidência |
| ----- | ---- | ------------------ | --- | --------- |
| 7.1 | Painel → **Dashboard** | KPIs coerentes com a compra | ☐ | |
| 7.2 | **Relatórios** → top produtos / estoque | Dados consistentes | ☐ | |
| 7.3 | Gerar previsão | Snapshots criados | ☐ | |
| 7.4 | `make backup` no servidor | Arquivo `.sql.gz` em `backups/` ou `/var/backups/store-platform` | ☐ | caminho: |
| 7.5 | `make test-backup-restore` ou restore em DB temporário | Restore OK | ☐ | |

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

Itens fora do MVP atual (documentar como pendência, não reprovar se já conhecido):

- Pix com PSP real (produção financeira).
- MFA em app mobile (EP-13 em evolução).

---

## Referência rápida — comandos

```bash
# Stack local completa
docker compose up -d --build

# Homologação automatizada
make test && make test-backup-restore
cd backend && go test -p 1 ./tests/e2e/...
```

Deploy no servidor: ver `docs/deployment.md`.
