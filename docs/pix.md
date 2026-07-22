# Plano detalhado de integração Pix com Mercado Pago

**Revisão:** 22 de julho de 2026
**Backend:** Go
**Frontend:** React
**Banco:** PostgreSQL
**Infraestrutura:** Docker, VPS e Portainer
**Solução Mercado Pago:** Checkout Transparente por API de Orders

---

# 1. Decisão técnica

Para uma integração nova, recomenda-se utilizar:

```text
Checkout Transparente
+ API de Orders
+ pagamento Pix
+ processing_mode automático
+ Webhooks do tópico Order
```

A API de Orders permite criar uma ordem contendo a transação Pix, receber o QR Code e acompanhar as mudanças de estado. A documentação atual do Mercado Pago apresenta a API de Orders como o novo modelo do Checkout Transparente e orienta selecionar essa API ao criar a aplicação.

Não se recomenda utilizar neste projeto:

* Checkout Pro com redirecionamento;
* links de pagamento criados manualmente;
* API de Pagamentos legada como primeira opção;
* criação da cobrança diretamente pelo React;
* consultas contínuas como mecanismo principal de confirmação;
* baixa baseada apenas no conteúdo recebido pelo webhook.

---

# 2. Fluxo geral

```text
Fatura fechada
      ↓
Cliente acessa a fatura
      ↓
Cliente seleciona “Gerar Pix”
      ↓
Backend valida fatura e cliente
      ↓
Backend cria uma Order Pix no Mercado Pago
      ↓
Mercado Pago retorna QR Code e Pix Copia e Cola
      ↓
React apresenta os dados
      ↓
Cliente paga no banco
      ↓
Mercado Pago envia Webhook de Order
      ↓
Backend valida a assinatura
      ↓
Worker consulta a Order no Mercado Pago
      ↓
Backend confirma estado, referência e valor
      ↓
Fatura é baixada em transação
      ↓
Limite do cliente é liberado
```

---

# 3. Princípio fundamental

A integração deverá separar claramente:

## Fatura interna

Representa o débito do cliente perante a loja.

```text
invoice
```

## Order do Mercado Pago

Representa a solicitação externa de pagamento Pix.

```text
mercado_pago_order
```

## Transação Pix do Mercado Pago

Representa o pagamento contido na Order.

```text
mercado_pago_payment
```

## Pagamento interno

Representa a confirmação financeira registrada no sistema.

```text
payment
```

Uma fatura poderá possuir mais de uma cobrança externa ao longo do tempo, por exemplo, quando um QR Code expirar e o cliente gerar outro.

Entretanto, apenas um pagamento confirmado poderá liquidar o mesmo saldo da fatura.

---

# 4. Evitar conflito de nomenclatura

O sistema já possui uma entidade interna chamada `Order`, que representa a compra do cliente.

O Mercado Pago também utiliza o termo `Order`.

Para evitar confusão, utilize os seguintes nomes no código:

```text
SalesOrder
Invoice
PaymentCharge
MercadoPagoOrder
MercadoPagoPayment
```

No banco:

```text
orders                       → pedido da loja
invoices                     → fatura da loja
payment_charges              → cobranças externas
payments                     → pagamentos confirmados
provider_order_id            → ID da Order do Mercado Pago
provider_payment_id          → ID da transação do Mercado Pago
```

---

# 5. Pré-requisitos no Mercado Pago

## 5.1. Conta

A loja deverá possuir uma conta Mercado Pago válida, preferencialmente empresarial e com os dados cadastrais atualizados.

## 5.2. Chave Pix

A conta deverá possuir uma chave Pix cadastrada. O Mercado Pago informa esse cadastro como pré-requisito para oferecer Pix no Checkout Transparente.

## 5.3. Aplicação

No painel Mercado Pago Developers:

1. acessar **Suas integrações**;
2. criar uma aplicação;
3. selecionar Checkout Transparente;
4. selecionar a API de Orders;
5. obter credenciais de teste;
6. configurar Webhooks de teste;
7. posteriormente ativar as credenciais produtivas.

## 5.4. Credenciais

As principais credenciais são:

```text
Public Key
Access Token
Webhook Secret
```

O `Access Token` é privado e deve ser utilizado exclusivamente pelo backend para gerar pagamentos. A Public Key é normalmente destinada a recursos de frontend, especialmente pagamentos com cartão. Para o fluxo Pix deste sistema, a criação será feita no backend com o Access Token, sem expor credenciais privadas no React.

---

# 6. Variáveis de ambiente

Configurar na VPS e no Portainer:

```text
MERCADO_PAGO_ENVIRONMENT=test
MERCADO_PAGO_BASE_URL=https://api.mercadopago.com
MERCADO_PAGO_ACCESS_TOKEN=...
MERCADO_PAGO_WEBHOOK_SECRET=...
MERCADO_PAGO_APPLICATION_ID=...
MERCADO_PAGO_PIX_EXPIRATION=PT24H
MERCADO_PAGO_REQUEST_TIMEOUT_SECONDS=10
```

Em produção:

```text
MERCADO_PAGO_ENVIRONMENT=production
```

As credenciais de teste e produção deverão ser completamente separadas.

Nunca armazenar essas credenciais:

* no Git;
* em arquivos versionados;
* no frontend;
* em imagens Docker;
* em logs;
* em tabelas comuns do banco;
* em mensagens de erro retornadas ao cliente.

---

# 7. Estratégia de criação da cobrança

## 7.1. Criação sob demanda

A recomendação é não criar automaticamente uma cobrança Mercado Pago quando a fatura for fechada.

O fechamento deverá apenas criar a fatura:

```text
status = OPEN
```

Quando o cliente abrir a fatura, será exibido:

```text
Gerar Pix
```

A cobrança externa será criada quando esse botão for acionado.

## 7.2. Justificativa

A criação sob demanda evita:

* cobranças que nunca serão utilizadas;
* QR Codes expirados sem necessidade;
* volume desnecessário de chamadas externas;
* registros externos excedentes;
* conciliações desnecessárias.

A fatura poderá estar fechada e disponível imediatamente, sem depender da disponibilidade momentânea do Mercado Pago.

---

# 8. Endpoints internos

## 8.1. Gerar cobrança

```text
POST /api/v1/me/invoices/{invoiceId}/pix-charge
```

Permissão:

```text
cliente proprietário da fatura
```

Endpoint administrativo opcional:

```text
POST /api/v1/admin/invoices/{invoiceId}/pix-charge
```

Permissão:

```text
payments.charge.create
```

## 8.2. Consultar cobrança atual

```text
GET /api/v1/me/invoices/{invoiceId}/pix-charge
```

## 8.3. Consultar estado da fatura

```text
GET /api/v1/me/invoices/{invoiceId}
```

## 8.4. Webhook

```text
POST /api/v1/webhooks/mercado-pago/orders
```

Esse endpoint será público, mas protegido pela validação criptográfica da assinatura.

## 8.5. Reconciliação administrativa

```text
POST /api/v1/admin/payments/charges/{chargeId}/reconcile
```

Permissão:

```text
payments.reconcile
```

---

# 9. Validações anteriores à criação

Antes de chamar o Mercado Pago, o backend deverá verificar:

1. a fatura existe;
2. a fatura pertence ao cliente autenticado;
3. a fatura está em estado pagável;
4. o saldo é maior que zero;
5. não existe pagamento integral confirmado;
6. o e-mail do cliente está preenchido;
7. não existe outra cobrança válida;
8. não existe cobrança em processo de criação;
9. o valor continua igual ao saldo esperado;
10. o provedor está habilitado.

Estados pagáveis:

```text
OPEN
OVERDUE
PARTIALLY_PAID
```

No MVP, recomenda-se gerar Pix apenas para pagamento integral do saldo restante.

---

# 10. Reutilização de cobrança válida

Ao receber a solicitação de geração, o sistema deverá procurar uma cobrança:

```text
provider = MERCADO_PAGO
status IN (CREATING, PROCESSING, ACTIVE)
expires_at > agora
amount_cents = saldo atual da fatura
```

Caso exista, ela deverá ser retornada, sem criar outra Order.

Isso protege contra:

* clique duplo;
* atualização da página;
* repetição da requisição;
* falha momentânea da resposta;
* chamadas concorrentes.

---

# 11. Idempotência interna

## 11.1. Chave de idempotência

Cada tentativa de criação deverá possuir um UUID:

```text
idempotency_key
```

Exemplo conceitual:

```text
5cd2ba39-b787-4fc4-b627-69ff30623e62
```

Essa chave será:

* gravada em `payment_charges`;
* enviada no header `X-Idempotency-Key`;
* reutilizada em uma repetição da mesma tentativa.

O Mercado Pago exige `X-Idempotency-Key` na criação de uma Order e recomenda um valor exclusivo para impedir duplicidade.

## 11.2. Restrição de concorrência

Criar índice parcial:

```sql
CREATE UNIQUE INDEX uq_active_charge_per_invoice
ON payment_charges (invoice_id, provider)
WHERE status IN ('CREATING', 'PROCESSING', 'ACTIVE');
```

Assim, dois cliques simultâneos não poderão criar duas cobranças ativas.

---

# 12. Não manter transação aberta durante chamada externa

Não se deve manter uma transação PostgreSQL bloqueada enquanto o sistema aguarda o Mercado Pago.

Utilizar o seguinte fluxo:

## Transação A

1. bloquear a fatura;
2. validar a situação;
3. localizar cobrança ativa;
4. criar `payment_charge` com estado `CREATING`;
5. gerar e salvar a chave de idempotência;
6. confirmar a transação.

## Chamada externa

7. enviar a Order ao Mercado Pago.

## Transação B

8. bloquear a cobrança;
9. salvar IDs externos;
10. salvar QR Code;
11. salvar vencimento;
12. mudar estado para `ACTIVE`;
13. confirmar.

Se ocorrer falha entre as etapas 7 e 8, a mesma chamada poderá ser repetida com a mesma chave de idempotência.

---

# 13. Conversão de valores

O sistema interno utiliza centavos:

```text
35050
```

O Mercado Pago recebe o valor como string decimal:

```text
"350.50"
```

Não utilizar `float64`.

Implementar um conversor explícito:

```go
func CentsToDecimalString(cents int64) string {
    return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}
```

Exemplos:

```text
100    → "1.00"
1990   → "19.90"
35050  → "350.50"
```

---

# 14. Criação da Order Pix

A requisição será enviada para:

```text
POST https://api.mercadopago.com/v1/orders
```

Headers:

```text
Authorization: Bearer {ACCESS_TOKEN}
Content-Type: application/json
Accept: application/json
X-Idempotency-Key: {UUID}
```

Payload conceitual:

```json
{
  "type": "online",
  "total_amount": "350.50",
  "external_reference": "INV-2026-000123",
  "processing_mode": "automatic",
  "transactions": {
    "payments": [
      {
        "amount": "350.50",
        "payment_method": {
          "id": "pix",
          "type": "bank_transfer"
        },
        "expiration_time": "PT24H"
      }
    ]
  },
  "payer": {
    "email": "cliente@exemplo.com"
  }
}
```

Para Pix, a API exige `payment_method.id = pix`, `payment_method.type = bank_transfer`, e-mail do pagador, referência externa, valor total e modo de processamento. A documentação informa que a expiração padrão é de 24 horas e permite configurá-la entre 30 minutos e 30 dias utilizando duração ISO 8601.

---

# 15. Referência externa

O campo:

```text
external_reference
```

deverá receber uma referência interna estável da fatura.

Recomendação:

```text
INV-2026-000123
```

Não utilizar:

* nome do cliente;
* CPF;
* e-mail;
* descrição contendo dados pessoais;
* ID do pedido de venda;
* número que possa ser reutilizado.

O backend deverá conseguir localizar exatamente uma fatura pela referência externa.

---

# 16. Prazo de expiração

Para o MVP, utilizar:

```text
PT24H
```

Isto significa 24 horas.

A expiração do Pix não deve ser confundida com o vencimento da fatura.

Exemplo:

```text
Fatura vencida em 10/08
Pix gerado em 08/08 às 15h
QR Code expira em 09/08 às 15h
```

Se o QR Code expirar:

* a fatura continuará aberta;
* a cobrança será marcada como expirada;
* o cliente poderá gerar um novo Pix;
* a cobrança antiga será preservada.

Posteriormente, o prazo poderá ser configurável:

```text
PT1H
PT12H
PT24H
P2D
```

---

# 17. Resposta do Mercado Pago

A resposta poderá conter:

```text
Order ID
Payment ID
reference_id
ticket_url
qr_code
qr_code_base64
status
status_detail
```

Enquanto estiver aguardando pagamento, a documentação indica:

```text
status = action_required
status_detail = waiting_transfer
```

A resposta também contém:

* `ticket_url`: página de pagamento;
* `qr_code`: Pix Copia e Cola;
* `qr_code_base64`: imagem do QR Code.

---

# 18. Dados que deverão ser armazenados

## Tabela `payment_charges`

```text
id
invoice_id
provider
idempotency_key
external_reference

provider_order_id
provider_payment_id
provider_reference_id

status
provider_status
provider_status_detail

amount_cents
ticket_url
pix_copy_paste
qr_code_base64

expires_at
paid_at
last_synced_at

creation_attempts
last_error_code
last_error_message

created_at
updated_at
```

## Restrições

```text
UNIQUE (provider, provider_order_id)
UNIQUE (provider, provider_payment_id)
UNIQUE (provider, idempotency_key)
```

Não registrar o Access Token em nenhuma tabela.

---

# 19. Estados internos da cobrança

```text
CREATING
PROCESSING
ACTIVE
PAID
EXPIRED
CANCELED
FAILED
REFUNDED
PARTIALLY_REFUNDED
REQUIRES_REVIEW
```

## Mapeamento inicial

| Mercado Pago                       | Estado interno       |
| ---------------------------------- | -------------------- |
| `created/created`                  | `PROCESSING`         |
| `processing/in_process`            | `PROCESSING`         |
| `action_required/waiting_transfer` | `ACTIVE`             |
| `processed/accredited`             | `PAID`               |
| `expired/expired`                  | `EXPIRED`            |
| `canceled/canceled`                | `CANCELED`           |
| `refunded/refunded`                | `REFUNDED`           |
| `processed/partially_refunded`     | `PARTIALLY_REFUNDED` |
| `failed/*`                         | `FAILED`             |

O estado que deverá autorizar a baixa automática é:

```text
transaction.status = processed
transaction.status_detail = accredited
```

O Mercado Pago define essa combinação como transação processada e valor efetivamente creditado.

---

# 20. Resposta da API interna ao React

Exemplo:

```json
{
  "data": {
    "charge_id": "f2dbc653-5f48-4a62-a30d-c8752342dc78",
    "invoice_id": "24ac03ed-917e-457c-b05a-ec804c82a421",
    "status": "ACTIVE",
    "amount_cents": 35050,
    "pix_copy_paste": "00020126...",
    "qr_code_base64": "iVBORw0KGgo...",
    "ticket_url": "https://...",
    "expires_at": "2026-08-09T15:00:00-03:00"
  }
}
```

O frontend não deverá receber:

* Access Token;
* segredo do webhook;
* payload interno completo do Mercado Pago;
* identificadores de outro cliente;
* mensagens técnicas da integração.

---

# 21. Interface React

## 21.1. Tela da fatura

Apresentar:

* número da fatura;
* competência;
* valor;
* vencimento;
* situação;
* saldo restante;
* botão “Gerar Pix”.

## 21.2. Após a geração

Apresentar:

* QR Code;
* Pix Copia e Cola;
* botão “Copiar código”;
* botão opcional “Abrir instruções”;
* valor;
* expiração;
* estado “Aguardando pagamento”.

O Mercado Pago permite apresentar diretamente o `qr_code_base64`, o código alfanumérico ou o `ticket_url`.

## 21.3. Atualização do estado

Enquanto a tela estiver aberta, o React poderá consultar:

```text
GET /api/v1/me/invoices/{id}
```

Intervalo sugerido:

```text
30 segundos
```

O polling é apenas uma melhoria de experiência. A baixa real deverá ser realizada pelo backend após webhook ou reconciliação.

## 21.4. Estados visuais

```text
Gerando cobrança
Aguardando pagamento
Pagamento em processamento
Pagamento confirmado
Código expirado
Não foi possível gerar o Pix
Pagamento requer conferência
```

---

# 22. Configuração do webhook

No painel do Mercado Pago:

1. abrir a aplicação;
2. acessar **Webhooks**;
3. configurar a URL HTTPS;
4. selecionar o modo de teste;
5. selecionar o evento **Order (Mercado Pago)**;
6. salvar;
7. copiar o segredo gerado;
8. executar o simulador;
9. repetir a configuração em produção.

Para Checkout Transparente por Orders, o Mercado Pago orienta utilizar o tópico `Order`, que notifica atualizações da Order e de suas transações.

URL:

```text
https://api.exemplo.com/api/v1/webhooks/mercado-pago/orders
```

---

# 23. Estrutura da notificação

A notificação contém, entre outros dados:

```json
{
  "action": "order.action_required",
  "type": "order",
  "live_mode": false,
  "data": {
    "id": "ORD01..."
  }
}
```

O ID importante para consulta é:

```text
data.id
```

A notificação também chega acompanhada dos headers:

```text
x-signature
x-request-id
```

e do query parameter:

```text
data.id
```

Esses valores participam da validação da assinatura.

---

# 24. Validação da assinatura

O backend deverá validar a assinatura antes de aceitar o evento.

A documentação oficial fornece um validador no SDK Go:

```go
err := webhook.ValidateSignature(
    r.Header.Get("x-signature"),
    r.Header.Get("x-request-id"),
    r.URL.Query().Get("data.id"),
    webhookSecret,
)
```

Se inválida:

```text
HTTP 401
```

Se válida:

```text
continuar processamento
```

O Mercado Pago utiliza o header `x-signature` e validação HMAC com o segredo da aplicação.

## Regras adicionais

* verificar se `type = order`;
* verificar se `data.id` não está vazio;
* limitar o tamanho do body;
* aplicar timeout;
* não confiar no `User-Agent`;
* não confiar somente no endereço IP;
* não registrar o segredo;
* comparar assinaturas de maneira segura.

---

# 25. Recepção rápida e processamento assíncrono

O endpoint do webhook não deverá executar toda a baixa financeira antes de responder.

Fluxo recomendado:

1. validar assinatura;
2. verificar estrutura;
3. inserir evento em `payment_events`;
4. inserir job na fila PostgreSQL;
5. responder HTTP 200;
6. worker processar o evento.

O Mercado Pago espera resposta HTTP 200 ou 201. A documentação informa espera de até 22 segundos e novas tentativas quando a confirmação não é recebida.

---

# 26. Idempotência do webhook

A mesma notificação poderá ser enviada mais de uma vez.

Tabela:

```text
payment_events
```

Campos:

```text
id
provider
external_event_id
provider_order_id
action
payload_hash
signature_valid
status
attempts
received_at
processed_at
last_error
```

Restrição:

```sql
UNIQUE (provider, external_event_id);
```

Também é recomendável impedir processamento paralelo do mesmo `provider_order_id`.

Se o evento já estiver processado:

```text
responder 200
não repetir a baixa
```

---

# 27. Não confiar no webhook como estado final

O webhook deverá ser tratado como uma notificação de que algo mudou.

Depois de validar a assinatura, o worker deverá consultar:

```text
GET /v1/orders/{providerOrderId}
```

A própria documentação orienta consultar a Order após o recebimento da notificação para obter as informações atualizadas do recurso.

O backend deverá confirmar:

1. a Order existe;
2. `external_reference` corresponde à fatura;
3. o valor total corresponde ao saldo esperado;
4. existe uma transação Pix;
5. o ID da transação corresponde à cobrança;
6. o estado é definitivo;
7. a conta e o ambiente correspondem;
8. a fatura ainda não foi liquidada.

---

# 28. Baixa automática da fatura

A baixa deverá ocorrer em uma transação PostgreSQL.

```text
BEGIN
```

## Etapas

1. bloquear `payment_charges`;
2. bloquear `invoices`;
3. verificar se a cobrança já foi processada;
4. verificar se a fatura já foi paga;
5. confirmar valor;
6. criar registro em `payments`;
7. atualizar cobrança para `PAID`;
8. atualizar fatura;
9. registrar `paid_at`;
10. liberar limite do cliente;
11. criar auditoria;
12. criar evento na outbox.

```text
COMMIT
```

## Exemplo de bloqueio

```sql
SELECT id, status, amount_cents
FROM payment_charges
WHERE provider = 'MERCADO_PAGO'
  AND provider_order_id = $1
FOR UPDATE;
```

---

# 29. Condições obrigatórias para baixa

A fatura só deverá ser marcada como paga quando todas as condições forem verdadeiras:

```text
assinatura válida
AND Order consultada diretamente no Mercado Pago
AND external_reference correta
AND provider_order_id correto
AND provider_payment_id correto
AND meio de pagamento = pix
AND valor recebido = saldo esperado
AND status = processed
AND status_detail = accredited
AND pagamento ainda não registrado
```

Se alguma condição falhar:

```text
REQUIRES_REVIEW
```

Nunca realizar baixa automática com valor divergente.

---

# 30. Registro do pagamento

Tabela `payments`:

```text
id
invoice_id
payment_charge_id
provider

provider_order_id
provider_payment_id
provider_reference_id

amount_cents
status
settled_at

created_at
updated_at
```

Restrições:

```text
UNIQUE (provider, provider_payment_id)
```

Isso impede que a mesma transação seja usada para liquidar duas faturas.

---

# 31. Liberação do limite

Após pagamento confirmado:

```text
exposição anterior = compras abertas + faturas abertas
```

O sistema deverá recalcular a exposição, em vez de simplesmente subtrair um valor sem conferência.

Fluxo:

```text
fatura liquidada
→ recalcular exposição
→ recalcular limite disponível
→ atualizar conta do cliente
```

Isso evita erros acumulados em caso de:

* créditos;
* ajustes;
* pagamentos parciais;
* cancelamentos;
* processamentos repetidos.

---

# 32. Cobrança expirada

Quando o Mercado Pago retornar:

```text
expired / expired
```

o sistema deverá:

1. marcar a cobrança como `EXPIRED`;
2. preservar QR Code e IDs para auditoria;
3. manter a fatura aberta;
4. remover a cobrança da condição de “ativa”;
5. permitir nova geração.

A documentação informa que pagamentos pendentes podem ser cancelados e que cobranças não pagas acabam sendo consideradas expiradas ou canceladas.

---

# 33. Cancelamento de cobrança pendente

Quando uma nova cobrança precisar substituir outra ainda pendente, o ideal é:

1. tentar cancelar a cobrança anterior;
2. atualizar seu estado;
3. criar nova cobrança;
4. preservar ambas.

Não apagar registros anteriores.

No MVP, pode-se simplesmente:

* impedir nova geração enquanto a cobrança estiver válida;
* permitir nova geração somente após sua expiração.

Isso reduz complexidade.

---

# 34. Reconciliação periódica

Além do webhook, um worker deverá executar reconciliação.

Periodicidade inicial:

```text
a cada 15 minutos
```

Selecionar cobranças:

```text
CREATING
PROCESSING
ACTIVE
REQUIRES_REVIEW
```

Para cada uma:

1. consultar a Order;
2. atualizar estado;
3. processar pagamento confirmado;
4. marcar expiração;
5. registrar divergência;
6. atualizar `last_synced_at`.

O webhook continua sendo o mecanismo principal. A reconciliação é uma contingência para:

* webhook perdido;
* falha temporária;
* indisponibilidade interna;
* erro de processamento;
* retorno assíncrono da criação.

---

# 35. Tratamento da criação assíncrona

A criação de uma Order pode retornar estado de processamento sem todas as informações imediatamente. O Mercado Pago recomenda utilizar notificações de Order ou consultar posteriormente `/v1/orders/{id}`.

Se a resposta inicial não contiver QR Code:

```text
payment_charge.status = PROCESSING
```

O frontend deverá mostrar:

```text
Estamos preparando o código Pix.
```

O worker ou webhook atualizará a cobrança quando os dados estiverem disponíveis.

---

# 36. Tratamento de erros HTTP

## 400 — requisição inválida

* não repetir automaticamente;
* registrar código;
* marcar como `FAILED`;
* revisar payload.

## 401 ou 403 — credencial inválida

* não repetir continuamente;
* gerar alerta crítico;
* não expor o erro ao cliente;
* verificar credencial e ambiente.

## 404 — Order não encontrada

* confirmar ambiente;
* confirmar ID;
* marcar para revisão;
* não liquidar fatura.

## 409 — conflito

* consultar estado existente;
* verificar idempotência;
* não criar segunda cobrança.

## 429 — limite excedido

O Mercado Pago orienta respeitar o header `Retry-After`. A repetição deverá utilizar a mesma chave de idempotência.

## 500 a 599

* repetir com backoff;
* manter a mesma chave;
* limitar tentativas;
* gerar alerta após esgotamento.

---

# 37. Política de repetição

Para operações de consulta:

```text
tentativa 1: imediata
tentativa 2: após 5 segundos
tentativa 3: após 30 segundos
tentativa 4: após 2 minutos
tentativa 5: após 15 minutos
```

Para criação:

* repetir apenas com a mesma chave de idempotência;
* nunca gerar nova chave após timeout sem antes verificar o estado;
* consultar a Order quando houver ID disponível.

Não repetir automaticamente erros de validação.

---

# 38. Cliente HTTP Go

Configurar:

```text
timeout total: 10 segundos
timeout de conexão: 3 segundos
limite de body: definido
pool de conexões: habilitado
TLS: obrigatório
```

Headers comuns:

```text
Authorization
Content-Type
Accept
X-Idempotency-Key
User-Agent
```

Registrar apenas:

* método;
* endpoint sem credenciais;
* status HTTP;
* duração;
* `x-request-id`;
* código de erro;
* ID interno da cobrança.

Nunca registrar:

* Access Token;
* QR Code;
* Pix Copia e Cola;
* e-mail completo, quando desnecessário;
* body integral com dados pessoais.

---

# 39. SDK oficial ou cliente HTTP

O Mercado Pago disponibiliza SDK oficial para Go e recomenda o uso dos SDKs server-side para reduzir o esforço de integração. A documentação também fornece diretamente o validador Go para assinatura de webhook.

Estratégia recomendada:

* utilizar o SDK oficial para validação de webhook;
* utilizar o cliente de Orders do SDK, caso esteja estável na versão adotada;
* manter toda a dependência dentro do adaptador Mercado Pago;
* usar `net/http` em um cliente tipado se algum recurso de Orders ainda não estiver exposto adequadamente pelo SDK.

O domínio nunca deverá depender diretamente do SDK.

---

# 40. Interface de domínio

```go
type PaymentGateway interface {
    CreatePixCharge(
        ctx context.Context,
        input CreatePixChargeInput,
    ) (PixCharge, error)

    GetCharge(
        ctx context.Context,
        externalOrderID string,
    ) (PixCharge, error)

    CancelCharge(
        ctx context.Context,
        externalOrderID string,
    ) error

    RefundPayment(
        ctx context.Context,
        input RefundPaymentInput,
    ) error

    ValidateWebhook(
        ctx context.Context,
        input WebhookValidationInput,
    ) error
}
```

O caso de uso não deverá conhecer:

* endpoint do Mercado Pago;
* Access Token;
* payload externo;
* SDK;
* nomes de status externos.

---

# 41. Estrutura de arquivos

```text
backend/internal/payments/
├── domain/
│   ├── payment_charge.go
│   ├── payment.go
│   ├── charge_status.go
│   ├── provider.go
│   └── errors.go
│
├── application/
│   ├── create_pix_charge/
│   │   ├── command.go
│   │   ├── handler.go
│   │   └── result.go
│   ├── process_webhook/
│   │   ├── command.go
│   │   └── handler.go
│   ├── synchronize_charge/
│   │   ├── command.go
│   │   └── handler.go
│   ├── reconcile_charges/
│   │   ├── command.go
│   │   └── handler.go
│   └── ports/
│       └── payment_gateway.go
│
├── infrastructure/
│   ├── postgres/
│   │   ├── charge_repository.go
│   │   ├── payment_repository.go
│   │   ├── event_repository.go
│   │   └── queries.sql
│   │
│   └── gateways/
│       └── mercadopago/
│           ├── gateway.go
│           ├── client.go
│           ├── config.go
│           ├── authentication.go
│           ├── create_order.go
│           ├── get_order.go
│           ├── cancel_order.go
│           ├── refund_payment.go
│           ├── webhook_validator.go
│           ├── status_mapper.go
│           ├── money_mapper.go
│           ├── requests.go
│           ├── responses.go
│           └── errors.go
│
├── transport/
│   └── http/
│       ├── create_pix_charge_handler.go
│       ├── get_pix_charge_handler.go
│       ├── webhook_handler.go
│       ├── reconcile_handler.go
│       ├── request.go
│       ├── response.go
│       └── routes.go
│
├── jobs/
│   ├── process_webhook_job.go
│   └── reconcile_charges_job.go
│
└── tests/
    ├── create_pix_charge_test.go
    ├── duplicate_request_test.go
    ├── webhook_signature_test.go
    ├── duplicate_webhook_test.go
    ├── reconciliation_test.go
    ├── amount_mismatch_test.go
    └── expiration_test.go
```

---

# 42. Permissões

## Cliente

Pode:

* gerar cobrança para sua própria fatura;
* consultar sua própria cobrança;
* consultar seu próprio pagamento.

## Gerente

```text
payments.read
payments.charge.create
```

## Financeiro

```text
payments.read
payments.charge.create
payments.reconcile
```

## Administrador

```text
payments.read
payments.charge.create
payments.reconcile
payments.refund
payments.settings.write
```

Estorno e reprocessamento deverão exigir:

* permissão específica;
* justificativa;
* step-up authentication;
* auditoria.

---

# 43. Auditoria

Registrar:

## Criação

```text
PIX_CHARGE_CREATED
```

## Reutilização

```text
PIX_CHARGE_REUSED
```

## Expiração

```text
PIX_CHARGE_EXPIRED
```

## Webhook

```text
MERCADO_PAGO_WEBHOOK_RECEIVED
MERCADO_PAGO_WEBHOOK_REJECTED
MERCADO_PAGO_WEBHOOK_PROCESSED
```

## Pagamento

```text
PAYMENT_CONFIRMED
PAYMENT_AMOUNT_MISMATCH
PAYMENT_REQUIRES_REVIEW
```

## Fatura

```text
INVOICE_PAID
```

A auditoria deverá registrar:

* usuário, quando houver;
* origem automática;
* fatura;
* cobrança;
* `request_id`;
* Order ID;
* Payment ID;
* valor;
* estado anterior;
* estado posterior.

---

# 44. Métricas

Coletar:

```text
mercadopago_order_creation_total
mercadopago_order_creation_errors_total
mercadopago_webhooks_received_total
mercadopago_webhooks_invalid_total
mercadopago_webhooks_duplicate_total
mercadopago_reconciliation_total
mercadopago_amount_mismatch_total
mercadopago_request_duration_seconds
mercadopago_pending_charges
```

Alertas:

* credencial inválida;
* muitos erros 5xx;
* aumento de assinaturas inválidas;
* cobranças ativas sem sincronização;
* pagamento divergente;
* webhook acumulado;
* reconciliação sem sucesso.

---

# 45. Tela administrativa de conciliação

Criar relatório simples com:

* fatura;
* cliente;
* Order ID;
* Payment ID;
* valor;
* estado interno;
* estado Mercado Pago;
* última sincronização;
* data do pagamento;
* divergência;
* ação disponível.

Filtros:

```text
Pendente
Pago
Expirado
Falha
Divergente
Sem sincronização
```

Ações:

* consultar novamente;
* visualizar fatura;
* visualizar auditoria;
* enviar para revisão.

Não permitir “marcar como pago” livremente.

Uma baixa manual excepcional deverá possuir caso de uso específico e permissão própria.

---

# 46. Testes em ambiente de desenvolvimento

O Mercado Pago fornece credenciais de teste e fluxo específico para criar Orders Pix de teste. A documentação atual utiliza `/v1/orders` e valores ou dados predefinidos para simular o processamento.

## Testes obrigatórios

### Criação

* fatura válida;
* fatura de outro cliente;
* fatura paga;
* fatura sem saldo;
* e-mail ausente;
* clique duplo;
* criação concorrente;
* timeout após envio;
* resposta sem QR Code;
* erro de autenticação;
* erro 429.

### QR Code

* exibição da imagem;
* cópia do código;
* expiração;
* regeneração;
* atualização de estado.

### Webhook

* assinatura válida;
* assinatura inválida;
* ausência de assinatura;
* evento duplicado;
* evento fora de ordem;
* evento de outro ambiente;
* evento sem `data.id`;
* body excessivo;
* Order inexistente.

### Baixa

* pagamento correto;
* valor divergente;
* referência divergente;
* Payment ID duplicado;
* fatura já paga;
* dois workers simultâneos;
* falha no meio da transação;
* repetição após falha.

### Reconciliação

* webhook perdido;
* cobrança expirada;
* cobrança ainda pendente;
* pagamento confirmado sem evento local;
* indisponibilidade temporária do Mercado Pago.

---

# 47. Teste end-to-end

Cenário completo:

1. criar cliente;
2. criar pedidos;
3. fechar competência;
4. gerar fatura;
5. cliente solicitar Pix;
6. backend criar Order;
7. React exibir QR Code;
8. executar pagamento de teste;
9. receber webhook;
10. validar assinatura;
11. consultar Order;
12. confirmar `processed/accredited`;
13. criar pagamento;
14. liquidar fatura;
15. liberar limite;
16. atualizar React;
17. conferir dashboard;
18. conferir auditoria;
19. repetir webhook;
20. confirmar que nenhum dado foi duplicado.

---

# 48. Implantação na VPS

## Stack Docker

```text
api
worker
postgres
store-web
admin-web
reverse-proxy
portainer
```

## Exposição

Somente o reverse proxy deverá expor:

```text
80
443
```

O endpoint do webhook deverá estar disponível por HTTPS.

## Segredos

Configurar no Portainer:

```text
MERCADO_PAGO_ACCESS_TOKEN
MERCADO_PAGO_WEBHOOK_SECRET
```

Recomenda-se:

* não utilizar essas variáveis em build args;
* não incorporá-las à imagem;
* restringir quem pode visualizar a stack;
* rotacionar a credencial em procedimento controlado.

---

# 49. Estratégia de entrada em produção

## Etapa 1 — Ambiente de teste

* aplicação de teste;
* credenciais de teste;
* webhook de teste;
* pagamentos simulados;
* logs e métricas ativos.

## Etapa 2 — Homologação

* testes end-to-end;
* testes de concorrência;
* testes de repetição;
* teste de webhook pelo simulador;
* teste de recuperação;
* validação da conciliação.

O Mercado Pago disponibiliza simulador de Webhooks para testar a URL e a recepção das notificações.

## Etapa 3 — Produção controlada

* ativar credenciais produtivas;
* cadastrar URL produtiva;
* configurar segredo produtivo;
* realizar pagamento real de valor baixo;
* conferir baixa;
* conferir saldo no Mercado Pago;
* conferir auditoria;
* monitorar primeiras cobranças.

## Etapa 4 — Estabilização

* acompanhar erros;
* acompanhar tempo de confirmação;
* revisar eventos duplicados;
* revisar cobranças expiradas;
* comparar relatórios internos com Mercado Pago.

---

# 50. Backlog de implementação

## Fase 1 — Fundação

1. criar aplicação Mercado Pago;
2. configurar credenciais de teste;
3. configurar chave Pix;
4. criar variáveis de ambiente;
5. criar interface `PaymentGateway`;
6. criar adaptador Mercado Pago.

## Fase 2 — Banco

1. ampliar `payment_charges`;
2. criar `payment_events`;
3. ampliar `payments`;
4. adicionar índices únicos;
5. adicionar estados;
6. criar migrations.

## Fase 3 — Geração

1. implementar validações;
2. implementar idempotência local;
3. implementar criação da Order;
4. mapear resposta;
5. salvar QR Code;
6. criar endpoint;
7. criar tela React.

## Fase 4 — Webhook

1. cadastrar URL;
2. implementar assinatura;
3. persistir evento;
4. responder rapidamente;
5. criar job;
6. consultar Order;
7. mapear estados.

## Fase 5 — Baixa

1. confirmar referência;
2. confirmar valor;
3. confirmar estado;
4. criar pagamento;
5. liquidar fatura;
6. liberar limite;
7. registrar auditoria;
8. publicar evento.

## Fase 6 — Contingência

1. criar worker de reconciliação;
2. tratar expiração;
3. tratar divergências;
4. criar tela administrativa;
5. configurar alertas.

## Fase 7 — Produção

1. ativar credenciais;
2. configurar segredo;
3. executar pagamento real controlado;
4. revisar logs;
5. homologar baixa;
6. liberar gradualmente.

---

# 51. Critérios de aceite

A integração estará concluída quando:

* uma fatura fechada permitir gerar Pix;
* dois cliques não gerarem duas cobranças;
* o QR Code puder ser exibido;
* o Pix Copia e Cola puder ser utilizado;
* cobrança válida for reutilizada;
* cobrança expirada puder ser substituída;
* webhook inválido for rejeitado;
* webhook duplicado não duplicar pagamento;
* pagamento for confirmado pela consulta da Order;
* somente `processed/accredited` baixar a fatura;
* valor divergente não baixar a fatura;
* fatura paga liberar o limite;
* falha de webhook ser recuperada pela reconciliação;
* credenciais não aparecerem nos logs;
* todos os eventos críticos possuírem auditoria;
* o fluxo funcionar em teste e produção.

---

# 52. Decisão final recomendada

A implementação deverá utilizar:

```text
API de Orders
processing_mode = automatic
payment_method.id = pix
payment_method.type = bank_transfer
X-Idempotency-Key obrigatório
external_reference vinculada à fatura
Webhook do tópico Order
Validação x-signature
Consulta GET da Order antes da baixa
Worker de reconciliação
```

O fluxo mais simples e seguro é:

```text
fatura
→ geração sob demanda
→ Order Pix
→ QR Code
→ webhook assinado
→ consulta da Order
→ confirmação processed/accredited
→ baixa transacional
→ liberação do limite
```

Essa abordagem evita acoplamento do domínio ao Mercado Pago, protege contra eventos falsos ou duplicados e permite substituir o provedor futuramente sem alterar o módulo de faturamento.
