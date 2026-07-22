# Plano de implementação do parcelamento de faturas

## 1. Objetivo

O sistema deverá permitir que uma fatura fechada seja paga:

* em uma única parcela;
* ou em várias parcelas, quando atender aos requisitos mínimos.

Cada parcela será registrada separadamente como conta a receber e, posteriormente, poderá gerar sua própria cobrança Pix.

A fatura original continuará sendo a entidade financeira principal e permanecerá aberta enquanto houver saldo pendente.

Fluxo geral:

```text
Fatura fechada
      ↓
Sistema calcula opções de pagamento
      ↓
Cliente seleciona quantidade de parcelas
      ↓
Sistema cria o plano de pagamento
      ↓
Sistema cria as parcelas
      ↓
Cada parcela pode gerar seu próprio Pix
      ↓
Pagamento confirmado baixa a parcela
      ↓
Fatura é atualizada
      ↓
Todas as parcelas pagas
      ↓
Fatura é marcada como paga
```

---

# 2. Regra principal

Os parâmetros iniciais serão:

```text
Valor mínimo da fatura para parcelamento: R$ 300,00
Valor mínimo de cada parcela: R$ 100,00
Quantidade máxima administrativa: 10 parcelas
```

A quantidade máxima permitida será calculada por:

```text
máximo pela regra =
valor total da fatura em centavos
÷ valor mínimo da parcela em centavos
```

Utilizando divisão inteira:

```text
máximo permitido =
mínimo entre:

10
e
inteiro(valor da fatura ÷ R$ 100,00)
```

Em termos de programação:

```text
max_installments =
min(
    administrative_max_installments,
    invoice_total_cents / minimum_installment_cents
)
```

Como a divisão é inteira, o resultado já será arredondado para baixo.

## Regra completa

```text
Se total da fatura < R$ 300,00:
    permitir apenas 1 parcela

Se total da fatura >= R$ 300,00:
    permitir de 1 até o máximo calculado

O máximo nunca poderá ultrapassar 10 parcelas.
```

A opção de pagamento à vista continuará sempre disponível.

---

# 3. Exemplos consolidados

|      Fatura | Opções                  |
| ----------: | ----------------------- |
|   R$ 250,00 | somente 1× de R$ 250,00 |
|   R$ 299,99 | somente 1× de R$ 299,99 |
|   R$ 300,00 | 1×, 2× ou 3×            |
|   R$ 310,00 | 1×, 2× ou 3×            |
|   R$ 350,00 | 1×, 2× ou 3×            |
|   R$ 400,00 | de 1× até 4×            |
|   R$ 450,00 | de 1× até 4×            |
|   R$ 500,00 | de 1× até 5×            |
|   R$ 750,00 | de 1× até 7×            |
| R$ 1.000,00 | de 1× até 10×           |
| R$ 2.000,00 | de 1× até 10×           |

Uma fatura de R$ 2.000,00 não deverá oferecer 20 parcelas de R$ 100,00, porque o limite administrativo será de 10 parcelas.

---

# 4. Tratamento correto dos centavos

A distribuição dos centavos deverá ser determinística e garantir que:

* a soma seja exatamente igual ao total da fatura;
* nenhuma parcela fique abaixo do mínimo;
* nenhum centavo seja perdido;
* o cálculo sempre produza o mesmo resultado.

Não é recomendável calcular parcelas com `float64`.

Todos os valores continuarão sendo manipulados em centavos.

## Algoritmo recomendado

```text
valor_base = total_cents / quantidade_parcelas
resto = total_cents % quantidade_parcelas
```

O valor-base será aplicado a todas as parcelas.

Os centavos restantes serão distribuídos, um por parcela, nas últimas parcelas do plano.

Exemplo em pseudocódigo:

```text
para cada parcela:
    valor = valor_base

para as últimas "resto" parcelas:
    adicionar 1 centavo
```

## Exemplo: R$ 310,00 em 3 parcelas

```text
total_cents = 31000
valor_base = 31000 / 3 = 10333
resto = 1
```

Resultado:

```text
1ª parcela: R$ 103,33
2ª parcela: R$ 103,33
3ª parcela: R$ 103,34
Total:      R$ 310,00
```

## Exemplo: R$ 350,00 em 3 parcelas

```text
total_cents = 35000
valor_base = 35000 / 3 = 11666
resto = 2
```

Resultado padronizado:

```text
1ª parcela: R$ 116,66
2ª parcela: R$ 116,67
3ª parcela: R$ 116,67
Total:      R$ 350,00
```

A ordem das parcelas de R$ 116,66 e R$ 116,67 não altera o resultado financeiro, mas o sistema deverá adotar um único padrão.

Distribuir os centavos adicionais nas últimas parcelas aproxima-se da regra de “ajustar a última parcela”, mas sem concentrar uma diferença excessiva em apenas uma delas.

## Garantia da parcela mínima

Antes de criar o plano, o sistema deverá confirmar:

```text
valor_base >= minimum_installment_cents
```

Como a quantidade máxima já é limitada por:

```text
total_cents / minimum_installment_cents
```

essa condição normalmente será verdadeira, mas a validação deverá permanecer como proteção adicional.

---

# 5. Parâmetros configuráveis

Os parâmetros não deverão ficar fixos diretamente no código.

Criar configuração administrativa:

```text
installment_enabled
minimum_invoice_amount_cents
minimum_installment_amount_cents
maximum_installments
installment_interval_months
allow_installment_after_due_date
allow_early_installment_payment
require_sequential_payment
adjust_due_date_to_business_day
```

Valores iniciais:

```text
installment_enabled = true
minimum_invoice_amount_cents = 30000
minimum_installment_amount_cents = 10000
maximum_installments = 10
installment_interval_months = 1
allow_installment_after_due_date = false
allow_early_installment_payment = false
require_sequential_payment = true
adjust_due_date_to_business_day = true
```

Essas configurações deverão ser alteradas somente por usuário autorizado.

Permissão recomendada:

```text
billing.installment_settings.write
```

Toda alteração deverá gerar auditoria.

---

# 6. Versionamento da política

Uma alteração futura da configuração não poderá modificar faturas já fechadas.

Exemplo:

```text
Janeiro:
mínimo da parcela = R$ 100,00

Março:
mínimo da parcela passa para R$ 150,00
```

As faturas fechadas em janeiro deverão continuar seguindo a regra de R$ 100,00.

Por isso, no fechamento da fatura o sistema deverá copiar um snapshot da política aplicável.

O plano da fatura deverá armazenar:

```text
minimum_invoice_amount_cents_snapshot
minimum_installment_amount_cents_snapshot
maximum_installments_snapshot
installment_interval_months_snapshot
adjust_due_date_to_business_day_snapshot
policy_version
```

Assim, a regra utilizada ficará preservada para auditoria.

---

# 7. Momento da escolha do parcelamento

O parcelamento não deve ser criado automaticamente no fechamento.

No fechamento:

1. a fatura é criada;
2. o sistema registra os parâmetros disponíveis;
3. a fatura fica aguardando a escolha do cliente;
4. o cliente acessa a fatura;
5. o sistema apresenta as opções;
6. o cliente escolhe a quantidade;
7. somente então as parcelas são criadas.

Isso evita criar planos que o cliente não deseja utilizar.

## Situação inicial

```text
invoice.status = OPEN
payment_plan.status = PENDING_SELECTION
```

Mesmo uma fatura que só possa ser paga à vista seguirá a mesma estrutura. Nesse caso, a única opção será:

```text
1 parcela
```

A utilização da mesma estrutura para pagamento à vista e parcelado simplificará posteriormente a integração Pix.

---

# 8. Prazo para escolha

Recomendação para o MVP:

```text
O cliente pode selecionar o plano até o vencimento original da fatura.
```

Depois do vencimento:

* o parcelamento não poderá ser criado automaticamente;
* a fatura permanecerá aberta ou vencida;
* um gerente poderá autorizar excepcionalmente um plano;
* a autorização deverá exigir justificativa;
* a operação deverá ser auditada.

Configuração:

```text
allow_installment_after_due_date = false
```

Permissão de exceção:

```text
billing.installments.override
```

---

# 9. Imutabilidade do plano escolhido

Depois que o cliente escolher a quantidade de parcelas, o plano deverá tornar-se ativo e imutável.

O cliente não poderá trocar livremente:

```text
3 parcelas
→ 5 parcelas
```

A alteração somente poderá ocorrer quando:

* nenhuma parcela tiver sido paga;
* nenhuma cobrança Pix estiver ativa;
* o plano anterior for cancelado;
* um usuário autorizado justificar a alteração;
* a operação for auditada.

Depois do pagamento da primeira parcela, o plano não poderá ser refeito no MVP.

Uma renegociação posterior deverá ser tratada como funcionalidade separada, fora da primeira versão.

---

# 10. Vencimentos das parcelas

## Regra inicial recomendada

A primeira parcela vencerá na data de vencimento original da fatura.

As demais vencerão mensalmente.

Exemplo:

```text
Fatura vence em 15/08/2026
Plano escolhido: 3 parcelas
```

Vencimentos:

```text
1ª parcela: 15/08/2026
2ª parcela: 15/09/2026
3ª parcela: 15/10/2026
```

## Meses sem o mesmo dia

Exemplo:

```text
1ª parcela: 31/01
```

O mês seguinte não possui dia 31.

Regra:

```text
utilizar o último dia válido do mês
```

Resultado:

```text
31/01
28/02 ou 29/02
31/03
```

## Dia não útil

Se o vencimento cair em:

* sábado;
* domingo;
* feriado configurado;

a data deverá ser transferida para o próximo dia útil.

Essa regra deverá utilizar o mesmo calendário comercial já previsto para o fechamento das faturas.

---

# 11. Pagamento sequencial

Para manter o MVP simples e controlado, recomenda-se exigir o pagamento sequencial.

Isso significa:

```text
Parcela 1 deve ser paga antes da parcela 2.
Parcela 2 deve ser paga antes da parcela 3.
```

O cliente poderá visualizar todas as parcelas, mas somente a parcela mais antiga ainda não paga poderá gerar um Pix.

Exemplo:

```text
Parcela 1: PAID
Parcela 2: OPEN
Parcela 3: SCHEDULED
```

A parcela 2 poderá gerar Pix.

A parcela 3 permanecerá bloqueada até a confirmação da parcela 2.

## Justificativa

O pagamento sequencial evita:

* parcela futura paga com parcela anterior vencida;
* múltiplas cobranças Pix abertas simultaneamente;
* confusão na conciliação;
* regras complexas de inadimplência;
* dificuldade para calcular qual parcela deve ser cobrada.

Futuramente, o sistema poderá permitir antecipação, mas isso não é necessário no MVP.

---

# 12. Estados do plano

```text
PENDING_SELECTION
ACTIVE
COMPLETED
CANCELED
DEFAULTED
```

## PENDING_SELECTION

A fatura foi fechada, mas o cliente ainda não escolheu a quantidade.

## ACTIVE

As parcelas foram criadas e existe saldo pendente.

## COMPLETED

Todas as parcelas foram pagas.

## CANCELED

O plano foi cancelado antes de qualquer pagamento ou por operação autorizada.

## DEFAULTED

Existe parcela vencida conforme a política de inadimplência.

---

# 13. Estados das parcelas

```text
SCHEDULED
OPEN
PIX_ACTIVE
PAID
OVERDUE
CANCELED
REQUIRES_REVIEW
```

## SCHEDULED

Parcela futura ainda não liberada.

## OPEN

Parcela disponível para pagamento.

## PIX_ACTIVE

Existe uma cobrança Pix válida.

## PAID

Pagamento confirmado.

## OVERDUE

Vencimento ultrapassado sem confirmação do pagamento.

## CANCELED

Parcela cancelada em razão do cancelamento do plano.

## REQUIRES_REVIEW

Existe divergência de valor, estado ou integração.

---

# 14. Estados da fatura

A fatura continuará sendo a entidade principal.

Estados recomendados:

```text
OPEN
PARTIALLY_PAID
PAID
OVERDUE
CANCELED
ADJUSTED
```

## Transições

### Ao fechar a fatura

```text
OPEN
```

### Após pagar a primeira parcela de um plano com várias parcelas

```text
PARTIALLY_PAID
```

### Enquanto houver parcelas pagas e saldo em aberto

```text
PARTIALLY_PAID
```

### Quando uma parcela vencer

```text
OVERDUE
```

### Quando todas forem pagas

```text
PAID
```

A fatura somente será marcada como `PAID` quando:

```text
paid_cents = total_cents
```

e todas as parcelas estiverem com estado:

```text
PAID
```

---

# 15. Modelo de dados

## 15.1. Configuração de parcelamento

### `installment_policies`

```text
id
version
active

minimum_invoice_amount_cents
minimum_installment_amount_cents
maximum_installments
installment_interval_months

allow_installment_after_due_date
allow_early_installment_payment
require_sequential_payment
adjust_due_date_to_business_day

valid_from
valid_until

created_by
created_at
```

As políticas anteriores não deverão ser apagadas.

Uma nova configuração deverá gerar uma nova versão.

---

## 15.2. Plano de pagamento da fatura

### `invoice_payment_plans`

```text
id
invoice_id
policy_id

status
selected_installment_count

invoice_total_cents
paid_cents
remaining_cents

minimum_invoice_amount_cents_snapshot
minimum_installment_amount_cents_snapshot
maximum_installments_snapshot
installment_interval_months_snapshot

allow_early_payment_snapshot
require_sequential_payment_snapshot
adjust_business_day_snapshot

selected_by_user_id
selected_at

canceled_by_user_id
canceled_at
cancellation_reason

created_at
updated_at
```

Restrições:

```text
UNIQUE(invoice_id)
```

Cada fatura deverá possuir somente um plano vigente.

No estado `PENDING_SELECTION`, o campo:

```text
selected_installment_count
```

permanecerá nulo.

---

## 15.3. Parcelas

### `invoice_installments`

```text
id
payment_plan_id
invoice_id

installment_number
amount_cents
paid_cents
remaining_cents

due_date
status

opened_at
paid_at
overdue_at
canceled_at

created_at
updated_at
```

Restrições:

```text
UNIQUE(payment_plan_id, installment_number)
CHECK(installment_number >= 1)
CHECK(amount_cents > 0)
CHECK(paid_cents >= 0)
CHECK(paid_cents <= amount_cents)
CHECK(remaining_cents >= 0)
```

Índices:

```text
invoice_id
payment_plan_id
status
due_date
```

---

# 16. Relacionamento com o Pix

A integração Pix deverá ser modificada para vincular cada cobrança a uma parcela, e não diretamente apenas à fatura.

## Estrutura futura

```text
invoice
   ↓
invoice_payment_plan
   ↓
invoice_installment
   ↓
payment_charge
   ↓
payment
```

## Alteração em `payment_charges`

Adicionar:

```text
installment_id
```

Estrutura:

```text
payment_charges
├── id
├── invoice_id
├── installment_id
├── provider
├── external_order_id
├── external_payment_id
├── amount_cents
├── status
└── expires_at
```

Restrições:

```text
FOREIGN KEY installment_id
UNIQUE(provider, external_order_id)
```

O valor enviado ao Mercado Pago deverá ser:

```text
invoice_installments.remaining_cents
```

e não o saldo total da fatura.

---

# 17. Registro dos pagamentos

A tabela `payments` também deverá apontar para a parcela:

```text
payments
├── id
├── invoice_id
├── installment_id
├── payment_charge_id
├── provider
├── external_payment_id
├── amount_cents
├── status
├── settled_at
└── created_at
```

Restrições:

```text
UNIQUE(provider, external_payment_id)
```

Uma confirmação de pagamento deverá:

1. localizar a cobrança;
2. localizar a parcela;
3. validar o valor;
4. marcar a parcela como paga;
5. atualizar o plano;
6. atualizar a fatura;
7. recalcular a exposição;
8. liberar a próxima parcela;
9. registrar auditoria.

---

# 18. Atualização após pagamento

## Primeira parcela de três

Situação inicial:

```text
Fatura: R$ 350,00
Parcela 1: R$ 116,66
Parcela 2: R$ 116,67
Parcela 3: R$ 116,67
```

Após a primeira parcela:

```text
Parcela 1: PAID
Parcela 2: OPEN
Parcela 3: SCHEDULED

Fatura:
paid_cents = 11666
remaining_cents = 23334
status = PARTIALLY_PAID
```

Após a segunda:

```text
Parcela 1: PAID
Parcela 2: PAID
Parcela 3: OPEN

Fatura:
paid_cents = 23333
remaining_cents = 11667
status = PARTIALLY_PAID
```

Após a terceira:

```text
Todas as parcelas: PAID

Fatura:
paid_cents = 35000
remaining_cents = 0
status = PAID
```

---

# 19. Limite do cliente

O parcelamento não deverá reduzir artificialmente a exposição do cliente.

Ao criar o plano:

```text
exposição = saldo total ainda não pago
```

A exposição será liberada proporcionalmente conforme cada parcela for confirmada.

Exemplo:

```text
Fatura: R$ 600,00
Limite aprovado: R$ 1.000,00
```

Antes dos pagamentos:

```text
Exposição: R$ 600,00
Limite disponível: R$ 400,00
```

Após pagamento de R$ 100,00:

```text
Exposição: R$ 500,00
Limite disponível: R$ 500,00
```

Não se deve considerar apenas a parcela atual como exposição.

Isso impediria que o cliente acumulasse novas compras sem considerar o saldo total das parcelas futuras.

---

# 20. Inadimplência

Uma fatura deverá ser considerada vencida quando houver ao menos uma parcela:

```text
OVERDUE
```

Regra recomendada:

```text
Se existir parcela vencida:
    invoice.status = OVERDUE
    customer.status = OVERDUE
```

A política de bloqueio poderá considerar:

* quantidade de dias em atraso;
* valor em atraso;
* tolerância administrativa;
* bloqueio imediato ou após período de carência.

Para o MVP:

```text
qualquer parcela vencida bloqueia novas compras
```

Essa regra poderá ser configurada posteriormente.

---

# 21. Cancelamentos, devoluções e créditos

## Antes da escolha do plano

Créditos ou ajustes poderão alterar normalmente o valor da fatura.

## Depois da escolha, mas antes de qualquer pagamento

O plano poderá ser cancelado e recriado com o novo saldo.

## Depois de uma parcela paga

Não se deve modificar retroativamente as parcelas já pagas.

A recomendação é:

1. registrar o crédito na fatura;
2. aplicar o crédito às últimas parcelas ainda abertas;
3. reduzir ou cancelar parcelas futuras;
4. preservar parcelas já pagas;
5. registrar auditoria.

Esse fluxo adiciona alguma complexidade.

Para o MVP, pode-se estabelecer uma regra mais simples:

```text
Após o primeiro pagamento:
créditos serão aplicados da última parcela para a primeira.
```

Exemplo:

```text
Parcelas futuras:
R$ 120,00
R$ 120,00

Crédito:
R$ 50,00
```

Resultado:

```text
Penúltima parcela: R$ 120,00
Última parcela: R$ 70,00
```

Entretanto, como a última ficaria abaixo do mínimo de R$ 100,00, será necessário redistribuir ou cancelar uma parcela.

Por isso, recomenda-se inicialmente:

```text
não permitir alterações automáticas no plano após a primeira parcela paga;
encaminhar créditos e devoluções para revisão administrativa.
```

Uma funcionalidade específica de recálculo poderá ser criada posteriormente.

---

# 22. Endpoints anteriores ao Pix

## Consultar opções

```text
GET /api/v1/me/invoices/{invoiceId}/payment-options
```

Resposta:

```json
{
  "data": {
    "invoice_id": "uuid",
    "total_cents": 35000,
    "installment_eligible": true,
    "maximum_installments": 3,
    "options": [
      {
        "installment_count": 1,
        "installments": [
          {
            "number": 1,
            "amount_cents": 35000,
            "due_date": "2026-08-15"
          }
        ]
      },
      {
        "installment_count": 2,
        "installments": [
          {
            "number": 1,
            "amount_cents": 17500,
            "due_date": "2026-08-15"
          },
          {
            "number": 2,
            "amount_cents": 17500,
            "due_date": "2026-09-15"
          }
        ]
      },
      {
        "installment_count": 3,
        "installments": [
          {
            "number": 1,
            "amount_cents": 11666,
            "due_date": "2026-08-15"
          },
          {
            "number": 2,
            "amount_cents": 11667,
            "due_date": "2026-09-15"
          },
          {
            "number": 3,
            "amount_cents": 11667,
            "due_date": "2026-10-15"
          }
        ]
      }
    ]
  }
}
```

## Selecionar plano

```text
POST /api/v1/me/invoices/{invoiceId}/payment-plan
```

Requisição:

```json
{
  "installment_count": 3
}
```

O backend deverá recalcular tudo.

Não deverá confiar nos valores enviados pelo React.

## Consultar plano

```text
GET /api/v1/me/invoices/{invoiceId}/payment-plan
```

## Consultar parcelas

```text
GET /api/v1/me/invoices/{invoiceId}/installments
```

## Alteração administrativa excepcional

```text
POST /api/v1/admin/invoices/{invoiceId}/payment-plan/reset
```

Permissão:

```text
billing.installments.override
```

---

# 23. Endpoint futuro do Pix

Após a implementação do Mercado Pago, a geração deixará de ocorrer pela fatura inteira:

```text
POST /api/v1/me/invoices/{invoiceId}/pix-charge
```

e passará a ocorrer pela parcela:

```text
POST /api/v1/me/installments/{installmentId}/pix-charge
```

O backend deverá verificar:

* a parcela pertence ao cliente;
* a parcela é a próxima disponível;
* a parcela está `OPEN`;
* não existe cobrança Pix ativa;
* o valor é o saldo da parcela;
* a fatura ainda está aberta;
* não existe parcela anterior pendente.

---

# 24. Interface React

## Tela da fatura sem plano

Apresentar:

```text
Valor total: R$ 350,00

Escolha uma forma de pagamento:

○ 1× de R$ 350,00
○ 2× de R$ 175,00
○ 3× de aproximadamente R$ 116,67
```

Ao selecionar uma opção, mostrar:

* valor de cada parcela;
* datas de vencimento;
* total;
* aviso sobre bloqueio em caso de atraso;
* confirmação da escolha.

## Tela com plano ativo

Exemplo:

| Parcela | Vencimento |     Valor | Estado   | Ação      |
| ------: | ---------: | --------: | -------- | --------- |
|     1/3 | 15/08/2026 | R$ 116,66 | Aberta   | Gerar Pix |
|     2/3 | 15/09/2026 | R$ 116,67 | Agendada | Aguardar  |
|     3/3 | 15/10/2026 | R$ 116,67 | Agendada | Aguardar  |

Após o primeiro pagamento:

| Parcela | Vencimento |     Valor | Estado   | Ação          |
| ------: | ---------: | --------: | -------- | ------------- |
|     1/3 | 15/08/2026 | R$ 116,66 | Paga     | Ver pagamento |
|     2/3 | 15/09/2026 | R$ 116,67 | Aberta   | Gerar Pix     |
|     3/3 | 15/10/2026 | R$ 116,67 | Agendada | Aguardar      |

---

# 25. Casos de uso do backend

Criar os seguintes casos de uso:

```text
CalculateInvoicePaymentOptions
SelectInvoicePaymentPlan
GenerateInstallmentSchedule
GetInvoicePaymentPlan
ListInvoiceInstallments
OpenNextInstallment
MarkInstallmentOverdue
ApplyInstallmentPayment
RecalculatePaymentPlan
RecalculateInvoicePaymentStatus
CancelPaymentPlan
OverridePaymentPlan
```

## Responsabilidades

### `CalculateInvoicePaymentOptions`

* carregar a fatura;
* carregar o snapshot da política;
* calcular máximo;
* gerar simulações;
* validar mínimo.

### `SelectInvoicePaymentPlan`

* verificar propriedade;
* verificar estado da fatura;
* verificar prazo;
* recalcular opção;
* criar parcelas;
* ativar a primeira;
* registrar auditoria.

### `ApplyInstallmentPayment`

Será usado posteriormente pela integração Pix para:

* validar o valor;
* marcar parcela paga;
* abrir a próxima;
* atualizar plano;
* atualizar fatura;
* atualizar limite.

---

# 26. Estrutura de arquivos

```text
backend/internal/billing/
├── domain/
│   ├── installment_policy.go
│   ├── payment_plan.go
│   ├── payment_plan_status.go
│   ├── invoice_installment.go
│   ├── installment_status.go
│   ├── installment_schedule.go
│   ├── installment_calculator.go
│   ├── due_date_calculator.go
│   ├── errors.go
│   ├── installment_policy_repository.go
│   ├── payment_plan_repository.go
│   └── installment_repository.go
│
├── application/
│   ├── calculate_payment_options/
│   │   ├── query.go
│   │   ├── handler.go
│   │   └── result.go
│   ├── select_payment_plan/
│   │   ├── command.go
│   │   ├── handler.go
│   │   └── validator.go
│   ├── get_payment_plan/
│   │   ├── query.go
│   │   └── handler.go
│   ├── list_installments/
│   │   ├── query.go
│   │   └── handler.go
│   ├── open_next_installment/
│   │   ├── command.go
│   │   └── handler.go
│   ├── apply_installment_payment/
│   │   ├── command.go
│   │   └── handler.go
│   ├── mark_overdue_installments/
│   │   ├── command.go
│   │   └── handler.go
│   ├── cancel_payment_plan/
│   │   ├── command.go
│   │   └── handler.go
│   └── override_payment_plan/
│       ├── command.go
│       └── handler.go
│
├── infrastructure/
│   └── postgres/
│       ├── installment_policy_repository.go
│       ├── payment_plan_repository.go
│       ├── installment_repository.go
│       └── installment_queries.sql
│
├── transport/
│   └── http/
│       ├── get_payment_options_handler.go
│       ├── select_payment_plan_handler.go
│       ├── get_payment_plan_handler.go
│       ├── list_installments_handler.go
│       ├── reset_payment_plan_handler.go
│       ├── request.go
│       ├── response.go
│       └── routes.go
│
├── jobs/
│   └── mark_overdue_installments_job.go
│
└── tests/
    ├── installment_eligibility_test.go
    ├── maximum_installments_test.go
    ├── cent_distribution_test.go
    ├── due_date_test.go
    ├── select_plan_test.go
    ├── duplicate_selection_test.go
    ├── sequential_payment_test.go
    ├── partial_invoice_payment_test.go
    └── overdue_installment_test.go
```

---

# 27. Migrations

Criar:

```text
000013_create_installment_policies.up.sql
000013_create_installment_policies.down.sql

000014_create_invoice_payment_plans.up.sql
000014_create_invoice_payment_plans.down.sql

000015_create_invoice_installments.up.sql
000015_create_invoice_installments.down.sql

000016_add_installment_to_payment_charges.up.sql
000016_add_installment_to_payment_charges.down.sql

000017_add_installment_to_payments.up.sql
000017_add_installment_to_payments.down.sql
```

As migrations `000016` e `000017` poderão ser preparadas antes do Mercado Pago, mas somente ativadas junto ao módulo Pix.

---

# 28. Requisitos funcionais adicionais

```text
RF-PAR-001
O sistema deverá permitir configurar os parâmetros do parcelamento.

RF-PAR-002
O sistema deverá preservar a política utilizada em cada fatura.

RF-PAR-003
Faturas abaixo do mínimo deverão permitir somente pagamento à vista.

RF-PAR-004
O sistema deverá calcular o máximo de parcelas.

RF-PAR-005
O máximo deverá respeitar o limite administrativo.

RF-PAR-006
Nenhuma parcela poderá ser inferior ao mínimo configurado.

RF-PAR-007
A soma das parcelas deverá ser exatamente igual à fatura.

RF-PAR-008
O cliente deverá poder consultar as opções antes da escolha.

RF-PAR-009
O cliente deverá selecionar uma única opção.

RF-PAR-010
O plano deverá tornar-se imutável após a ativação.

RF-PAR-011
Cada parcela deverá possuir valor e vencimento próprios.

RF-PAR-012
Cada parcela deverá ser registrada como conta a receber.

RF-PAR-013
A fatura deverá permanecer aberta enquanto houver saldo.

RF-PAR-014
O pagamento de uma parcela deverá reduzir o saldo da fatura.

RF-PAR-015
A fatura somente deverá ser paga após todas as parcelas.

RF-PAR-016
A exposição deverá considerar todo o saldo não pago.

RF-PAR-017
Uma parcela vencida deverá tornar a fatura vencida.

RF-PAR-018
O sistema deverá permitir geração Pix por parcela.

RF-PAR-019
O pagamento Pix deverá ser vinculado à parcela correta.

RF-PAR-020
A mesma transação não poderá pagar duas parcelas.
```

---

# 29. Auditoria

Registrar:

```text
INSTALLMENT_POLICY_CREATED
INSTALLMENT_POLICY_CHANGED
PAYMENT_PLAN_SELECTED
PAYMENT_PLAN_CANCELED
PAYMENT_PLAN_OVERRIDDEN
INSTALLMENT_OPENED
INSTALLMENT_MARKED_OVERDUE
INSTALLMENT_PAID
INVOICE_PARTIALLY_PAID
INVOICE_FULLY_PAID
```

O log deverá conter:

* fatura;
* plano;
* parcela;
* usuário;
* valores;
* número de parcelas;
* política aplicada;
* estado anterior;
* estado posterior;
* justificativa, quando necessária.

---

# 30. Relatórios afetados

## Contas a receber

Passará a apresentar:

* fatura;
* parcela;
* número da parcela;
* vencimento;
* valor;
* saldo;
* estado.

## Exposição do cliente

Continuará apresentando o saldo total não pago da fatura.

## Inadimplência

Será baseada nas parcelas vencidas.

## Conciliação Pix

Cada pagamento deverá apontar:

```text
fatura
→ parcela
→ cobrança
→ pagamento
```

## Dashboard

Adicionar:

* saldo parcelado em aberto;
* parcelas a vencer;
* parcelas vencidas;
* faturas parcialmente pagas.

Não é necessário criar novos relatórios complexos. Esses dados podem integrar os relatórios já planejados.

---

# 31. Testes obrigatórios

## Limites

```text
R$ 299,99 → somente 1 parcela
R$ 300,00 → máximo 3
R$ 399,99 → máximo 3
R$ 400,00 → máximo 4
R$ 999,99 → máximo 9
R$ 1.000,00 → máximo 10
R$ 2.000,00 → máximo 10
```

## Centavos

Testar:

```text
R$ 310,00 em 3
R$ 350,00 em 3
R$ 620,00 em 6
R$ 1.000,01 em 10
```

Validar sempre:

```text
soma das parcelas = total
nenhuma parcela < mínimo
```

## Concorrência

* duas escolhas simultâneas;
* dois pagamentos simultâneos;
* geração duplicada de plano;
* abertura duplicada da próxima parcela.

## Estados

* pagamento da primeira parcela;
* pagamento intermediário;
* pagamento da última;
* parcela vencida;
* plano cancelado;
* fatura parcialmente paga;
* fatura totalmente paga.

## Configuração

* mudança da política;
* faturas antigas preservam snapshot;
* nova fatura utiliza nova política.

---

# 32. Backlog técnico

## Etapa 1 — Política

1. criar `installment_policies`;
2. cadastrar política inicial;
3. criar tela de configuração;
4. adicionar permissão;
5. registrar auditoria.

## Etapa 2 — Cálculos

1. implementar elegibilidade;
2. implementar quantidade máxima;
3. implementar distribuição de centavos;
4. implementar vencimentos;
5. criar testes unitários.

## Etapa 3 — Plano e parcelas

1. criar migrations;
2. criar entidades;
3. criar repositories;
4. criar caso de uso de seleção;
5. criar parcelas;
6. ativar primeira parcela.

## Etapa 4 — API e React

1. criar endpoint de opções;
2. criar endpoint de escolha;
3. criar endpoint de consulta;
4. criar seleção no React;
5. criar tabela de parcelas.

## Etapa 5 — Estados financeiros

1. atualizar estado da fatura;
2. atualizar exposição;
3. marcar atrasos;
4. bloquear cliente inadimplente;
5. atualizar relatórios.

## Etapa 6 — Preparação para Pix

1. adicionar `installment_id` ao modelo da cobrança;
2. adicionar `installment_id` ao pagamento;
3. criar interface `ApplyInstallmentPayment`;
4. substituir geração Pix por fatura pela geração por parcela;
5. preparar reconciliação por parcela.

---

# 33. Decisões recomendadas para o MVP

Para limitar a complexidade:

```text
Parcelamento sem juros
Máximo de 10 parcelas
Mínimo de R$ 300,00 por fatura
Mínimo de R$ 100,00 por parcela
Escolha única do plano
Pagamento sequencial
Uma cobrança Pix ativa por vez
Primeira parcela no vencimento original
Parcelas seguintes mensais
Vencimento ajustado para próximo dia útil
Sem pagamento parcial de uma parcela
Sem renegociação automática
Sem alteração após primeiro pagamento
```

Essas decisões cobrem o parcelamento necessário sem transformar o sistema em uma solução completa de financiamento ou renegociação de dívidas.

---

# 34. Critérios de aceite

O parcelamento estará pronto para receber a integração Pix quando:

* faturas abaixo de R$ 300,00 aceitarem apenas 1 parcela;
* faturas elegíveis apresentarem opções corretas;
* o máximo respeitar R$ 100,00 por parcela;
* nenhuma opção ultrapassar 10 parcelas;
* a soma das parcelas for exatamente igual à fatura;
* os vencimentos forem gerados corretamente;
* a política ficar registrada na fatura;
* o cliente puder escolher uma única opção;
* cada parcela possuir estado próprio;
* a primeira parcela ficar disponível;
* as demais permanecerem agendadas;
* uma parcela paga liberar a seguinte;
* a fatura permanecer parcialmente paga;
* o saldo total continuar na exposição;
* a última parcela paga liquidar a fatura;
* as tabelas de Pix puderem apontar para `installment_id`;
* todos os fluxos críticos possuírem testes automatizados.
