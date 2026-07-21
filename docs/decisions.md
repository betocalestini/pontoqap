# Parte III — Backlog técnico

## 17. Convenções do backlog

### Prioridades

| Prioridade | Significado                                            |
| ---------- | ------------------------------------------------------ |
| P0         | Necessário para funcionamento ou segurança do MVP      |
| P1         | Necessário para conclusão comercial da primeira versão |
| P2         | Melhoria importante após estabilização                 |
| P3         | Evolução futura                                        |

### Estimativas

As estimativas são expressas em pontos relativos:

| Pontos | Complexidade aproximada             |
| -----: | ----------------------------------- |
|      1 | alteração pequena                   |
|      2 | tarefa simples                      |
|      3 | tarefa moderada                     |
|      5 | tarefa relevante                    |
|      8 | tarefa complexa                     |
|     13 | deve ser dividida antes da execução |

---

## EP-00 — Descoberta, governança e requisitos

| ID      | Item                                         | Prioridade | Pontos | Dependência       |
| ------- | -------------------------------------------- | ---------: | -----: | ----------------- |
| BK-0001 | Validar o fluxo comercial pós-pago           |         P0 |      3 | —                 |
| BK-0002 | Definir política de aprovação de clientes    |         P0 |      2 | BK-0001           |
| BK-0003 | Definir limite padrão                        |         P0 |      2 | BK-0002           |
| BK-0004 | Definir política de inadimplência            |         P0 |      3 | BK-0002           |
| BK-0005 | Definir política de cancelamento e devolução |         P0 |      3 | BK-0001           |
| BK-0006 | Definir feriados aplicáveis                  |         P0 |      2 | —                 |
| BK-0007 | Selecionar PSP com API Pix                   |         P0 |      5 | —                 |
| BK-0008 | Validar obrigações fiscais e jurídicas       |         P0 |      5 | BK-0001           |
| BK-0009 | Criar glossário de domínio                   |         P1 |      2 | —                 |
| BK-0010 | Criar mapa de permissões                     |         P0 |      3 | —                 |
| BK-0011 | Criar protótipos da loja                     |         P1 |      5 | BK-0001           |
| BK-0012 | Criar protótipos do painel                   |         P1 |      5 | BK-0001           |
| BK-0013 | Aprovar critérios de aceite do MVP           |         P0 |      3 | BK-0001 a BK-0012 |

**Definição de pronto do épico:**

* regras comerciais documentadas;
* PSP selecionado;
* permissões aprovadas;
* fluxos críticos validados;
* critérios de aceite formalizados.

---

## EP-01 — Fundação do monorepositório

| ID      | Item                                     | Prioridade | Pontos | Dependência       |
| ------- | ---------------------------------------- | ---------: | -----: | ----------------- |
| BK-0101 | Criar repositório Git                    |         P0 |      1 | —                 |
| BK-0102 | Configurar `pnpm workspace`              |         P0 |      2 | BK-0101           |
| BK-0103 | Criar módulo Go                          |         P0 |      2 | BK-0101           |
| BK-0104 | Criar `store-web` em React               |         P0 |      2 | BK-0102           |
| BK-0105 | Criar `admin-web` em React               |         P0 |      2 | BK-0102           |
| BK-0106 | Criar projeto React Native/Expo          |         P1 |      2 | BK-0102           |
| BK-0107 | Criar pacote de contratos                |         P0 |      2 | BK-0102           |
| BK-0108 | Criar pacote de cliente da API           |         P0 |      3 | BK-0107           |
| BK-0109 | Criar pacote de validações               |         P1 |      2 | BK-0102           |
| BK-0110 | Criar pacote de tokens de design         |         P1 |      2 | BK-0102           |
| BK-0111 | Criar pacote `web-ui`                    |         P1 |      3 | BK-0110           |
| BK-0112 | Configurar lint e formatação             |         P0 |      3 | BK-0103 a BK-0106 |
| BK-0113 | Configurar testes básicos                |         P0 |      3 | BK-0103 a BK-0106 |
| BK-0114 | Criar Makefile                           |         P1 |      2 | BK-0103           |
| BK-0115 | Criar Docker Compose local               |         P0 |      5 | BK-0103 a BK-0105 |
| BK-0116 | Criar health checks iniciais             |         P0 |      2 | BK-0103           |
| BK-0117 | Configurar OpenAPI inicial               |         P0 |      3 | BK-0103           |
| BK-0118 | Configurar geração do cliente TypeScript |         P0 |      5 | BK-0108 e BK-0117 |
| BK-0119 | Criar pipeline de pull request           |         P0 |      5 | BK-0112 e BK-0113 |
| BK-0120 | Criar ADR da arquitetura                 |         P1 |      2 | —                 |

---

## EP-02 — Identidade, sessões e permissões

| ID      | Item                                      | Prioridade | Pontos | Dependência       |
| ------- | ----------------------------------------- | ---------: | -----: | ----------------- |
| BK-0201 | Criar migrations de identidade            |         P0 |      5 | EP-01             |
| BK-0202 | Implementar entidade `User`               |         P0 |      3 | BK-0201           |
| BK-0203 | Implementar papéis e permissões           |         P0 |      5 | BK-0201           |
| BK-0204 | Implementar hash de senha                 |         P0 |      3 | BK-0202           |
| BK-0205 | Implementar criação de sessão web         |         P0 |      5 | BK-0202           |
| BK-0206 | Implementar login                         |         P0 |      5 | BK-0204 e BK-0205 |
| BK-0207 | Implementar logout                        |         P0 |      2 | BK-0205           |
| BK-0208 | Implementar revogação de sessão           |         P0 |      3 | BK-0205           |
| BK-0209 | Implementar recuperação de senha          |         P1 |      5 | BK-0202           |
| BK-0210 | Implementar rate limiting de autenticação |         P0 |      3 | BK-0206           |
| BK-0211 | Implementar bloqueio progressivo          |         P0 |      3 | BK-0206           |
| BK-0212 | Implementar MFA administrativo            |         P0 |      8 | BK-0206           |
| BK-0213 | Implementar middleware de autenticação    |         P0 |      5 | BK-0205           |
| BK-0214 | Implementar middleware de permissão       |         P0 |      5 | BK-0203           |
| BK-0215 | Criar tela de login da loja               |         P0 |      3 | BK-0206           |
| BK-0216 | Criar tela de login administrativo        |         P0 |      3 | BK-0206           |
| BK-0217 | Criar proteção de rotas React             |         P0 |      3 | BK-0213           |
| BK-0218 | Criar gestão administrativa de usuários   |         P1 |      8 | BK-0203           |
| BK-0219 | Criar testes de autorização               |         P0 |      5 | BK-0214           |
| BK-0220 | Criar auditoria de login                  |         P1 |      3 | BK-0206           |

---

## EP-03 — Catálogo de produtos

| ID      | Item                                  | Prioridade | Pontos | Dependência       |
| ------- | ------------------------------------- | ---------: | -----: | ----------------- |
| BK-0301 | Criar migrations do catálogo          |         P0 |      5 | EP-01             |
| BK-0302 | Implementar entidade `Category`       |         P0 |      2 | BK-0301           |
| BK-0303 | Implementar entidade `Product`        |         P0 |      3 | BK-0301           |
| BK-0304 | Implementar entidade `SKU`            |         P0 |      3 | BK-0301           |
| BK-0305 | Implementar objeto `Money`            |         P0 |      3 | EP-01             |
| BK-0306 | Implementar cadastro de categoria     |         P0 |      3 | BK-0302           |
| BK-0307 | Implementar cadastro de produto       |         P0 |      5 | BK-0303 e BK-0304 |
| BK-0308 | Implementar alteração de produto      |         P0 |      5 | BK-0307           |
| BK-0309 | Implementar inativação e reativação   |         P0 |      3 | BK-0307           |
| BK-0310 | Implementar histórico de preços       |         P0 |      5 | BK-0304           |
| BK-0311 | Implementar armazenamento de imagens  |         P1 |      5 | BK-0303           |
| BK-0312 | Implementar pesquisa e filtros        |         P0 |      5 | BK-0307           |
| BK-0313 | Criar catálogo público na loja        |         P0 |      8 | BK-0312           |
| BK-0314 | Criar página de detalhes              |         P0 |      5 | BK-0313           |
| BK-0315 | Criar listagem administrativa         |         P0 |      5 | BK-0312           |
| BK-0316 | Criar formulário administrativo       |         P0 |      8 | BK-0307           |
| BK-0317 | Criar upload de imagem                |         P1 |      5 | BK-0311           |
| BK-0318 | Criar testes de produto e preço       |         P0 |      5 | BK-0307 a BK-0310 |
| BK-0319 | Criar testes de catálogo público      |         P1 |      3 | BK-0313           |
| BK-0320 | Criar auditoria de alteração de preço |         P0 |      3 | BK-0310           |

---

## EP-04 — Controle de estoque

| ID      | Item                                    | Prioridade | Pontos | Dependência       |
| ------- | --------------------------------------- | ---------: | -----: | ----------------- |
| BK-0401 | Criar migrations de estoque             |         P0 |      5 | EP-03             |
| BK-0402 | Implementar saldo por SKU               |         P0 |      5 | BK-0401           |
| BK-0403 | Implementar movimentação imutável       |         P0 |      5 | BK-0401           |
| BK-0404 | Implementar entrada de estoque          |         P0 |      5 | BK-0402           |
| BK-0405 | Implementar perda e avaria              |         P0 |      3 | BK-0403           |
| BK-0406 | Implementar ajuste de inventário        |         P0 |      5 | BK-0403           |
| BK-0407 | Implementar bloqueio de linha           |         P0 |      5 | BK-0402           |
| BK-0408 | Implementar restrição de saldo negativo |         P0 |      2 | BK-0401           |
| BK-0409 | Implementar consulta de movimentações   |         P0 |      3 | BK-0403           |
| BK-0410 | Implementar alerta de estoque mínimo    |         P1 |      3 | BK-0402           |
| BK-0411 | Criar tela de estoque                   |         P0 |      5 | BK-0402           |
| BK-0412 | Criar modal de entrada rápida           |         P0 |      5 | BK-0404           |
| BK-0413 | Criar tela de histórico                 |         P0 |      5 | BK-0409           |
| BK-0414 | Criar fluxo de perda e ajuste           |         P0 |      5 | BK-0405 e BK-0406 |
| BK-0415 | Criar teste de concorrência             |         P0 |      8 | BK-0407           |
| BK-0416 | Criar testes de compensação             |         P0 |      5 | BK-0403           |
| BK-0417 | Criar auditoria de ajustes              |         P0 |      3 | BK-0406           |

---

## EP-05 — Clientes, aprovação e limite

| ID      | Item                                | Prioridade | Pontos | Dependência |
| ------- | ----------------------------------- | ---------: | -----: | ----------- |
| BK-0501 | Criar migrations de clientes        |         P0 |      5 | EP-02       |
| BK-0502 | Implementar entidade `Customer`     |         P0 |      3 | BK-0501     |
| BK-0503 | Implementar solicitação de cadastro |         P0 |      5 | BK-0502     |
| BK-0504 | Implementar cadastro pelo gerente   |         P0 |      5 | BK-0502     |
| BK-0505 | Implementar aprovação               |         P0 |      3 | BK-0502     |
| BK-0506 | Implementar bloqueio e desbloqueio  |         P0 |      3 | BK-0502     |
| BK-0507 | Implementar limite individual       |         P0 |      5 | BK-0502     |
| BK-0508 | Implementar histórico do limite     |         P0 |      3 | BK-0507     |
| BK-0509 | Implementar cálculo de exposição    |         P0 |      5 | BK-0507     |
| BK-0510 | Criar tela de cadastro do cliente   |         P0 |      5 | BK-0503     |
| BK-0511 | Criar lista administrativa          |         P0 |      5 | BK-0504     |
| BK-0512 | Criar tela de detalhes financeiros  |         P0 |      5 | BK-0509     |
| BK-0513 | Criar alteração de limite           |         P0 |      3 | BK-0507     |
| BK-0514 | Criar auditoria do limite           |         P0 |      3 | BK-0508     |
| BK-0515 | Criar testes de bloqueio            |         P0 |      3 | BK-0506     |
| BK-0516 | Criar testes do cálculo de limite   |         P0 |      5 | BK-0509     |

---

## EP-06 — Carrinho, checkout e pedidos

| ID      | Item                                   | Prioridade | Pontos | Dependência   |
| ------- | -------------------------------------- | ---------: | -----: | ------------- |
| BK-0601 | Criar migrations de carrinho e pedidos |         P0 |      5 | EP-03 e EP-05 |
| BK-0602 | Implementar carrinho persistente       |         P0 |      5 | BK-0601       |
| BK-0603 | Implementar inclusão de item           |         P0 |      3 | BK-0602       |
| BK-0604 | Implementar alteração e remoção        |         P0 |      3 | BK-0602       |
| BK-0605 | Implementar entidade `Order`           |         P0 |      5 | BK-0601       |
| BK-0606 | Implementar snapshots dos itens        |         P0 |      3 | BK-0605       |
| BK-0607 | Implementar idempotência de checkout   |         P0 |      5 | BK-0605       |
| BK-0608 | Implementar caso de uso de checkout    |         P0 |      8 | EP-04 e EP-05 |
| BK-0609 | Implementar lançamento no faturamento  |         P0 |      5 | BK-0608       |
| BK-0610 | Implementar cancelamento               |         P0 |      5 | BK-0605       |
| BK-0611 | Implementar devolução total            |         P1 |      5 | BK-0610       |
| BK-0612 | Preparar devolução parcial             |         P2 |      5 | BK-0611       |
| BK-0613 | Criar interface de carrinho            |         P0 |      5 | BK-0602       |
| BK-0614 | Criar tela de revisão                  |         P0 |      3 | BK-0613       |
| BK-0615 | Criar confirmação da compra            |         P0 |      5 | BK-0608       |
| BK-0616 | Criar histórico do cliente             |         P0 |      5 | BK-0605       |
| BK-0617 | Criar listagem administrativa          |         P0 |      5 | BK-0605       |
| BK-0618 | Criar testes end-to-end do checkout    |         P0 |      8 | BK-0608       |
| BK-0619 | Criar teste de repetição da requisição |         P0 |      5 | BK-0607       |

---

## EP-07 — Faturamento mensal

| ID      | Item                                   | Prioridade | Pontos | Dependência       |
| ------- | -------------------------------------- | ---------: | -----: | ----------------- |
| BK-0701 | Criar migrations de faturamento        |         P0 |      8 | EP-06             |
| BK-0702 | Implementar período mensal             |         P0 |      5 | BK-0701           |
| BK-0703 | Implementar lançamentos financeiros    |         P0 |      5 | BK-0701           |
| BK-0704 | Implementar calendário comercial       |         P0 |      5 | BK-0701           |
| BK-0705 | Implementar cálculo do quinto dia útil |         P0 |      5 | BK-0704           |
| BK-0706 | Implementar entidade `Invoice`         |         P0 |      5 | BK-0701           |
| BK-0707 | Implementar fechamento idempotente     |         P0 |      8 | BK-0702 e BK-0706 |
| BK-0708 | Implementar aplicação de créditos      |         P0 |      5 | BK-0707           |
| BK-0709 | Implementar ajustes administrativos    |         P1 |      5 | BK-0706           |
| BK-0710 | Implementar fechamento manual          |         P1 |      3 | BK-0707           |
| BK-0711 | Implementar marcação de vencidas       |         P0 |      3 | BK-0706           |
| BK-0712 | Implementar bloqueio por atraso        |         P0 |      5 | BK-0711           |
| BK-0713 | Criar tela de período atual            |         P0 |      5 | BK-0702           |
| BK-0714 | Criar tela de faturas do cliente       |         P0 |      5 | BK-0706           |
| BK-0715 | Criar gestão administrativa de faturas |         P0 |      5 | BK-0706           |
| BK-0716 | Criar gestão de calendário             |         P0 |      5 | BK-0704           |
| BK-0717 | Criar testes do quinto dia útil        |         P0 |      8 | BK-0705           |
| BK-0718 | Criar testes de fechamento duplicado   |         P0 |      5 | BK-0707           |
| BK-0719 | Criar testes de créditos e ajustes     |         P0 |      5 | BK-0708 e BK-0709 |
| BK-0720 | Criar auditoria do fechamento          |         P0 |      3 | BK-0707           |

---

## EP-08 — Integração Pix

| ID      | Item                                            | Prioridade | Pontos | Dependência       |
| ------- | ----------------------------------------------- | ---------: | -----: | ----------------- |
| BK-0801 | Definir interface `PaymentGateway`              |         P0 |      3 | EP-07             |
| BK-0802 | Implementar adaptador sandbox do PSP            |         P0 |      8 | BK-0801           |
| BK-0803 | Criar migrations de pagamento                   |         P0 |      5 | EP-07             |
| BK-0804 | Implementar criação de cobrança                 |         P0 |      5 | BK-0802 e BK-0803 |
| BK-0805 | Implementar reaproveitamento de cobrança válida |         P0 |      3 | BK-0804           |
| BK-0806 | Implementar endpoint de webhook                 |         P0 |      5 | BK-0802           |
| BK-0807 | Implementar autenticação do webhook             |         P0 |      5 | BK-0806           |
| BK-0808 | Implementar deduplicação de eventos             |         P0 |      5 | BK-0806           |
| BK-0809 | Implementar validação do valor                  |         P0 |      3 | BK-0806           |
| BK-0810 | Implementar liquidação da fatura                |         P0 |      5 | BK-0809           |
| BK-0811 | Implementar liberação do limite                 |         P0 |      3 | BK-0810           |
| BK-0812 | Implementar reconciliação periódica             |         P0 |      5 | BK-0802           |
| BK-0813 | Implementar expiração                           |         P0 |      3 | BK-0812           |
| BK-0814 | Implementar tratamento de divergências          |         P0 |      5 | BK-0809           |
| BK-0815 | Implementar estorno, se disponível              |         P2 |      8 | BK-0802           |
| BK-0816 | Criar tela Pix do cliente                       |         P0 |      5 | BK-0804           |
| BK-0817 | Criar acompanhamento de status                  |         P0 |      3 | BK-0810           |
| BK-0818 | Criar gestão administrativa de pagamentos       |         P1 |      5 | BK-0810           |
| BK-0819 | Criar testes de webhook duplicado               |         P0 |      5 | BK-0808           |
| BK-0820 | Criar testes de valor divergente                |         P0 |      5 | BK-0809           |
| BK-0821 | Criar teste completo em sandbox                 |         P0 |      8 | BK-0802 a BK-0817 |

---

## EP-09 — Dashboard e relatórios

| ID      | Item                               | Prioridade | Pontos | Dependência       |
| ------- | ---------------------------------- | ---------: | -----: | ----------------- |
| BK-0901 | Definir consultas do dashboard     |         P0 |      3 | EP-06 a EP-08     |
| BK-0902 | Implementar vendas por competência |         P0 |      5 | BK-0901           |
| BK-0903 | Implementar recebimentos por caixa |         P0 |      5 | BK-0901           |
| BK-0904 | Implementar contas em aberto       |         P0 |      3 | BK-0901           |
| BK-0905 | Implementar faturas vencidas       |         P0 |      3 | BK-0901           |
| BK-0906 | Implementar ticket médio           |         P1 |      3 | BK-0901           |
| BK-0907 | Implementar ranking de produtos    |         P0 |      5 | BK-0901           |
| BK-0908 | Implementar ranking de clientes    |         P0 |      5 | BK-0901           |
| BK-0909 | Implementar relatório de estoque   |         P0 |      5 | EP-04             |
| BK-0910 | Implementar relatório mensal       |         P0 |      5 | BK-0902 e BK-0903 |
| BK-0911 | Implementar exportação CSV         |         P1 |      5 | BK-0907 a BK-0910 |
| BK-0912 | Criar dashboard React              |         P0 |      8 | BK-0902 a BK-0908 |
| BK-0913 | Criar páginas de relatórios        |         P0 |      8 | BK-0907 a BK-0910 |
| BK-0914 | Criar filtros de período           |         P0 |      3 | BK-0912           |
| BK-0915 | Criar testes de conciliação        |         P0 |      8 | BK-0902 a BK-0910 |

---

## EP-10 — Previsão de reposição

| ID      | Item                             | Prioridade | Pontos | Dependência       |
| ------- | -------------------------------- | ---------: | -----: | ----------------- |
| BK-1001 | Criar migration de snapshots     |         P1 |      3 | EP-04 e EP-09     |
| BK-1002 | Consolidar vendas por SKU e mês  |         P1 |      5 | BK-1001           |
| BK-1003 | Implementar média ponderada      |         P1 |      3 | BK-1002           |
| BK-1004 | Implementar estoque de segurança |         P1 |      3 | BK-1002           |
| BK-1005 | Implementar sugestão de compra   |         P1 |      3 | BK-1003 e BK-1004 |
| BK-1006 | Implementar nível de confiança   |         P1 |      5 | BK-1003           |
| BK-1007 | Tratar produtos sem histórico    |         P1 |      3 | BK-1003           |
| BK-1008 | Persistir snapshot               |         P1 |      3 | BK-1005           |
| BK-1009 | Criar job mensal de previsão     |         P1 |      5 | BK-1008           |
| BK-1010 | Criar tela de previsão           |         P1 |      5 | BK-1008           |
| BK-1011 | Exibir explicação do cálculo     |         P1 |      3 | BK-1010           |
| BK-1012 | Criar testes dos cálculos        |         P1 |      5 | BK-1003 a BK-1007 |

---

## EP-11 — Segurança, auditoria e operação

| ID      | Item                                | Prioridade | Pontos | Dependência       |
| ------- | ----------------------------------- | ---------: | -----: | ----------------- |
| BK-1101 | Criar tabela de auditoria           |         P0 |      3 | EP-01             |
| BK-1102 | Implementar serviço de auditoria    |         P0 |      5 | BK-1101           |
| BK-1103 | Implementar `request_id`            |         P0 |      3 | EP-01             |
| BK-1104 | Implementar logging estruturado     |         P0 |      3 | EP-01             |
| BK-1105 | Implementar proteção CSRF           |         P0 |      5 | EP-02             |
| BK-1106 | Implementar cabeçalhos de segurança |         P0 |      3 | EP-01             |
| BK-1107 | Implementar limites de upload       |         P0 |      3 | EP-03             |
| BK-1108 | Criar Dockerfiles multi-stage       |         P0 |      5 | EP-01             |
| BK-1109 | Criar Compose de produção           |         P0 |      5 | BK-1108           |
| BK-1110 | Configurar reverse proxy e TLS      |         P0 |      5 | BK-1109           |
| BK-1111 | Configurar redes Docker             |         P0 |      3 | BK-1109           |
| BK-1112 | Configurar Portainer                |         P0 |      3 | BK-1109           |
| BK-1113 | Configurar backup diário            |         P0 |      5 | BK-1109           |
| BK-1114 | Configurar backup externo           |         P0 |      5 | BK-1113           |
| BK-1115 | Criar procedimento de restauração   |         P0 |      5 | BK-1113           |
| BK-1116 | Testar restauração                  |         P0 |      5 | BK-1115           |
| BK-1117 | Implementar métricas                |         P1 |      5 | BK-1104           |
| BK-1118 | Configurar alertas                  |         P1 |      5 | BK-1117           |
| BK-1119 | Configurar scanner de dependências  |         P0 |      3 | EP-01             |
| BK-1120 | Configurar scanner de imagens       |         P0 |      3 | BK-1108           |
| BK-1121 | Criar política de retenção          |         P1 |      3 | BK-1101 e BK-1113 |
| BK-1122 | Criar runbook operacional           |         P1 |      5 | BK-1112 a BK-1118 |

---

## EP-12 — Jobs, outbox e resiliência

| ID      | Item                               | Prioridade | Pontos | Dependência       |
| ------- | ---------------------------------- | ---------: | -----: | ----------------- |
| BK-1201 | Criar migrations de jobs           |         P0 |      3 | EP-01             |
| BK-1202 | Criar migrations de outbox         |         P0 |      3 | EP-01             |
| BK-1203 | Implementar repositório de jobs    |         P0 |      5 | BK-1201           |
| BK-1204 | Implementar aquisição com bloqueio |         P0 |      5 | BK-1203           |
| BK-1205 | Implementar repetição com backoff  |         P0 |      5 | BK-1203           |
| BK-1206 | Implementar dead-letter lógico     |         P1 |      3 | BK-1205           |
| BK-1207 | Implementar publicação na outbox   |         P0 |      5 | BK-1202           |
| BK-1208 | Implementar processador da outbox  |         P0 |      5 | BK-1207           |
| BK-1209 | Integrar pagamento à outbox        |         P0 |      3 | EP-08             |
| BK-1210 | Integrar fechamento à outbox       |         P1 |      3 | EP-07             |
| BK-1211 | Criar métricas de jobs             |         P1 |      3 | BK-1203           |
| BK-1212 | Criar testes de falha e repetição  |         P0 |      5 | BK-1205 e BK-1208 |

---

## EP-13 — Preparação para React Native

| ID      | Item                             | Prioridade | Pontos | Dependência       |
| ------- | -------------------------------- | ---------: | -----: | ----------------- |
| BK-1301 | Validar contratos compartilhados |         P1 |      3 | EP-01             |
| BK-1302 | Validar cliente da API no Expo   |         P1 |      3 | BK-1301           |
| BK-1303 | Criar armazenamento seguro       |         P2 |      3 | EP-02             |
| BK-1304 | Implementar autenticação móvel   |         P2 |      5 | BK-1303           |
| BK-1305 | Criar navegação inicial          |         P2 |      3 | BK-1302           |
| BK-1306 | Criar tela móvel de catálogo     |         P2 |      5 | EP-03             |
| BK-1307 | Criar carrinho móvel             |         P2 |      5 | EP-06             |
| BK-1308 | Criar tela móvel de faturas      |         P2 |      5 | EP-07             |
| BK-1309 | Criar tela móvel de Pix          |         P2 |      5 | EP-08             |
| BK-1310 | Testar Android                   |         P2 |      3 | BK-1304 a BK-1309 |
| BK-1311 | Testar iOS                       |         P2 |      3 | BK-1304 a BK-1309 |

---

# Parte IV — Árvore completa de arquivos

## 18. Raiz do monorepositório

```text
store-platform/
├── .github/
│   ├── workflows/
│   │   ├── pull-request.yml
│   │   ├── build-images.yml
│   │   ├── deploy-staging.yml
│   │   ├── deploy-production.yml
│   │   ├── dependency-scan.yml
│   │   └── backup-verification.yml
│   ├── CODEOWNERS
│   └── pull_request_template.md
│
├── backend/
├── apps/
├── packages/
├── infra/
├── docs/
├── scripts/
│
├── .editorconfig
├── .env.example
├── .gitignore
├── .prettierignore
├── .prettierrc.json
├── compose.yaml
├── eslint.config.js
├── Makefile
├── package.json
├── pnpm-lock.yaml
├── pnpm-workspace.yaml
├── README.md
├── SECURITY.md
└── tsconfig.base.json
```

---

# 19. Backend Go

```text
backend/
├── cmd/
│   ├── api/
│   │   ├── main.go
│   │   ├── bootstrap.go
│   │   ├── dependencies.go
│   │   ├── modules.go
│   │   ├── router.go
│   │   └── shutdown.go
│   │
│   ├── worker/
│   │   ├── main.go
│   │   ├── bootstrap.go
│   │   ├── dependencies.go
│   │   ├── handlers.go
│   │   └── shutdown.go
│   │
│   └── migrate/
│       └── main.go
│
├── internal/
│   ├── identity/
│   ├── customers/
│   ├── catalog/
│   ├── inventory/
│   ├── sales/
│   ├── billing/
│   ├── payments/
│   ├── reports/
│   ├── forecasting/
│   ├── audit/
│   ├── jobs/
│   └── platform/
│
├── migrations/
├── openapi/
├── tests/
├── Dockerfile
├── go.mod
└── go.sum
```

---

# 20. Módulo `identity`

**Estado implementado (MVP):** um único `users` por pessoa; clientes da loja têm papel `customer` + linha em `customers`; funcionários internos são **clientes** com papel interno adicional (`POST /admin/customers/{id}/staff-role` ou convite por e-mail já cadastrado na loja); convite com `admin_invitations` e aceite em `POST /auth/accept-invitation` (preserva papel `customer`); `suspended` suspende só o painel admin; papéis fixos `system_admin`, `manager`, `inventory_operator`, `finance_operator`; bootstrap via `ADMIN_BOOTSTRAP_*`; auditoria RF-IDN-012.

```text
backend/internal/identity/
├── domain/
│   ├── user.go
│   ├── user_id.go
│   ├── user_status.go
│   ├── email.go
│   ├── password_hash.go
│   ├── role.go
│   ├── permission.go
│   ├── session.go
│   ├── session_id.go
│   ├── errors.go
│   ├── user_repository.go
│   ├── session_repository.go
│   └── permission_repository.go
│
├── application/
│   ├── login/
│   │   ├── command.go
│   │   ├── handler.go
│   │   ├── result.go
│   │   └── validator.go
│   ├── logout/
│   │   ├── command.go
│   │   └── handler.go
│   ├── refresh_session/
│   │   ├── command.go
│   │   └── handler.go
│   ├── get_current_user/
│   │   ├── query.go
│   │   ├── handler.go
│   │   └── result.go
│   ├── request_password_reset/
│   │   ├── command.go
│   │   └── handler.go
│   ├── reset_password/
│   │   ├── command.go
│   │   ├── handler.go
│   │   └── validator.go
│   ├── create_user/
│   │   ├── command.go
│   │   └── handler.go
│   ├── assign_role/
│   │   ├── command.go
│   │   └── handler.go
│   ├── revoke_sessions/
│   │   ├── command.go
│   │   └── handler.go
│   ├── setup_mfa/
│   │   ├── command.go
│   │   └── handler.go
│   ├── verify_mfa/
│   │   ├── command.go
│   │   └── handler.go
│   ├── ports/
│   │   ├── password_hasher.go
│   │   ├── session_token_generator.go
│   │   ├── mfa_provider.go
│   │   └── notification_sender.go
│   └── errors.go
│
├── infrastructure/
│   ├── postgres/
│   │   ├── user_repository.go
│   │   ├── session_repository.go
│   │   ├── permission_repository.go
│   │   ├── user_mapper.go
│   │   └── queries.sql
│   ├── security/
│   │   ├── argon_password_hasher.go
│   │   ├── random_token_generator.go
│   │   └── totp_mfa_provider.go
│   └── notification/
│       └── password_reset_sender.go
│
├── transport/
│   └── http/
│       ├── login_handler.go
│       ├── logout_handler.go
│       ├── refresh_handler.go
│       ├── me_handler.go
│       ├── forgot_password_handler.go
│       ├── reset_password_handler.go
│       ├── setup_mfa_handler.go
│       ├── verify_mfa_handler.go
│       ├── auth_middleware.go
│       ├── permission_middleware.go
│       ├── csrf_middleware.go
│       ├── request.go
│       ├── response.go
│       └── routes.go
│
├── tests/
│   ├── login_test.go
│   ├── session_test.go
│   ├── permissions_test.go
│   └── password_reset_test.go
│
└── module.go
```

---

# 21. Módulo `customers`

```text
backend/internal/customers/
├── domain/
│   ├── customer.go
│   ├── customer_id.go
│   ├── customer_status.go
│   ├── credit_limit.go
│   ├── exposure.go
│   ├── limit_change.go
│   ├── errors.go
│   ├── customer_repository.go
│   └── limit_history_repository.go
│
├── application/
│   ├── register_customer/
│   │   ├── command.go
│   │   ├── handler.go
│   │   └── validator.go
│   ├── approve_customer/
│   │   ├── command.go
│   │   └── handler.go
│   ├── reject_customer/
│   │   ├── command.go
│   │   └── handler.go
│   ├── block_customer/
│   │   ├── command.go
│   │   └── handler.go
│   ├── unblock_customer/
│   │   ├── command.go
│   │   └── handler.go
│   ├── change_credit_limit/
│   │   ├── command.go
│   │   ├── handler.go
│   │   └── validator.go
│   ├── calculate_available_limit/
│   │   ├── query.go
│   │   ├── handler.go
│   │   └── result.go
│   ├── get_customer/
│   │   ├── query.go
│   │   └── handler.go
│   ├── list_customers/
│   │   ├── query.go
│   │   ├── handler.go
│   │   └── result.go
│   ├── get_limit_history/
│   │   ├── query.go
│   │   └── handler.go
│   └── ports/
│       └── financial_exposure_reader.go
│
├── infrastructure/
│   └── postgres/
│       ├── customer_repository.go
│       ├── limit_history_repository.go
│       ├── exposure_reader.go
│       ├── customer_mapper.go
│       └── queries.sql
│
├── transport/
│   └── http/
│       ├── register_customer_handler.go
│       ├── create_customer_handler.go
│       ├── approve_customer_handler.go
│       ├── reject_customer_handler.go
│       ├── block_customer_handler.go
│       ├── unblock_customer_handler.go
│       ├── change_limit_handler.go
│       ├── get_customer_handler.go
│       ├── list_customers_handler.go
│       ├── get_account_handler.go
│       ├── request.go
│       ├── response.go
│       └── routes.go
│
├── tests/
│   ├── approve_customer_test.go
│   ├── block_customer_test.go
│   ├── credit_limit_test.go
│   └── available_limit_test.go
│
└── module.go
```

---

# 22. Módulo `catalog`

```text
backend/internal/catalog/
├── domain/
│   ├── category.go
│   ├── category_id.go
│   ├── product.go
│   ├── product_id.go
│   ├── sku.go
│   ├── sku_id.go
│   ├── money.go
│   ├── unit.go
│   ├── product_image.go
│   ├── price_change.go
│   ├── errors.go
│   ├── category_repository.go
│   ├── product_repository.go
│   └── price_history_repository.go
│
├── application/
│   ├── create_category/
│   │   ├── command.go
│   │   └── handler.go
│   ├── update_category/
│   │   ├── command.go
│   │   └── handler.go
│   ├── deactivate_category/
│   │   ├── command.go
│   │   └── handler.go
│   ├── create_product/
│   │   ├── command.go
│   │   ├── handler.go
│   │   └── validator.go
│   ├── update_product/
│   │   ├── command.go
│   │   ├── handler.go
│   │   └── validator.go
│   ├── change_product_status/
│   │   ├── command.go
│   │   └── handler.go
│   ├── change_sku_price/
│   │   ├── command.go
│   │   └── handler.go
│   ├── add_product_image/
│   │   ├── command.go
│   │   └── handler.go
│   ├── remove_product_image/
│   │   ├── command.go
│   │   └── handler.go
│   ├── get_product/
│   │   ├── query.go
│   │   └── handler.go
│   ├── list_products/
│   │   ├── query.go
│   │   ├── handler.go
│   │   └── result.go
│   ├── list_categories/
│   │   ├── query.go
│   │   └── handler.go
│   └── ports/
│       └── image_storage.go
│
├── infrastructure/
│   ├── postgres/
│   │   ├── category_repository.go
│   │   ├── product_repository.go
│   │   ├── price_history_repository.go
│   │   ├── product_mapper.go
│   │   └── queries.sql
│   └── storage/
│       └── local_image_storage.go
│
├── transport/
│   └── http/
│       ├── create_category_handler.go
│       ├── update_category_handler.go
│       ├── create_product_handler.go
│       ├── update_product_handler.go
│       ├── change_status_handler.go
│       ├── change_price_handler.go
│       ├── upload_image_handler.go
│       ├── delete_image_handler.go
│       ├── get_product_handler.go
│       ├── list_products_handler.go
│       ├── list_categories_handler.go
│       ├── request.go
│       ├── response.go
│       └── routes.go
│
├── tests/
│   ├── product_test.go
│   ├── sku_test.go
│   ├── money_test.go
│   ├── change_price_test.go
│   └── product_status_test.go
│
└── module.go
```

---

# 23. Módulo `inventory`

```text
backend/internal/inventory/
├── domain/
│   ├── inventory_balance.go
│   ├── inventory_location.go
│   ├── stock_movement.go
│   ├── movement_type.go
│   ├── quantity.go
│   ├── errors.go
│   ├── balance_repository.go
│   ├── movement_repository.go
│   └── location_repository.go
│
├── application/
│   ├── register_stock_entry/
│   │   ├── command.go
│   │   ├── handler.go
│   │   └── validator.go
│   ├── register_stock_loss/
│   │   ├── command.go
│   │   └── handler.go
│   ├── adjust_inventory/
│   │   ├── command.go
│   │   ├── handler.go
│   │   └── validator.go
│   ├── decrease_for_order/
│   │   ├── command.go
│   │   └── handler.go
│   ├── restore_from_cancellation/
│   │   ├── command.go
│   │   └── handler.go
│   ├── get_inventory/
│   │   ├── query.go
│   │   └── handler.go
│   ├── list_inventory/
│   │   ├── query.go
│   │   ├── handler.go
│   │   └── result.go
│   ├── list_movements/
│   │   ├── query.go
│   │   └── handler.go
│   └── list_low_stock/
│       ├── query.go
│       └── handler.go
│
├── infrastructure/
│   └── postgres/
│       ├── balance_repository.go
│       ├── movement_repository.go
│       ├── location_repository.go
│       ├── inventory_mapper.go
│       └── queries.sql
│
├── transport/
│   └── http/
│       ├── stock_entry_handler.go
│       ├── stock_loss_handler.go
│       ├── inventory_adjustment_handler.go
│       ├── get_inventory_handler.go
│       ├── list_inventory_handler.go
│       ├── list_movements_handler.go
│       ├── low_stock_handler.go
│       ├── request.go
│       ├── response.go
│       └── routes.go
│
├── tests/
│   ├── stock_entry_test.go
│   ├── stock_loss_test.go
│   ├── inventory_adjustment_test.go
│   ├── negative_stock_test.go
│   └── concurrent_decrease_test.go
│
└── module.go
```

---

# 24. Módulo `sales`

```text
backend/internal/sales/
├── domain/
│   ├── cart.go
│   ├── cart_id.go
│   ├── cart_item.go
│   ├── order.go
│   ├── order_id.go
│   ├── order_item.go
│   ├── order_status.go
│   ├── order_return.go
│   ├── return_item.go
│   ├── idempotency_key.go
│   ├── errors.go
│   ├── cart_repository.go
│   ├── order_repository.go
│   └── return_repository.go
│
├── application/
│   ├── get_cart/
│   │   ├── query.go
│   │   └── handler.go
│   ├── add_cart_item/
│   │   ├── command.go
│   │   └── handler.go
│   ├── update_cart_item/
│   │   ├── command.go
│   │   └── handler.go
│   ├── remove_cart_item/
│   │   ├── command.go
│   │   └── handler.go
│   ├── clear_cart/
│   │   ├── command.go
│   │   └── handler.go
│   ├── checkout/
│   │   ├── command.go
│   │   ├── handler.go
│   │   ├── result.go
│   │   └── validator.go
│   ├── cancel_order/
│   │   ├── command.go
│   │   └── handler.go
│   ├── return_order/
│   │   ├── command.go
│   │   └── handler.go
│   ├── get_order/
│   │   ├── query.go
│   │   └── handler.go
│   ├── list_customer_orders/
│   │   ├── query.go
│   │   └── handler.go
│   ├── list_admin_orders/
│   │   ├── query.go
│   │   └── handler.go
│   └── ports/
│       ├── product_reader.go
│       ├── inventory_service.go
│       ├── customer_account_reader.go
│       └── billing_entry_writer.go
│
├── infrastructure/
│   └── postgres/
│       ├── cart_repository.go
│       ├── order_repository.go
│       ├── return_repository.go
│       ├── sales_mapper.go
│       └── queries.sql
│
├── transport/
│   └── http/
│       ├── get_cart_handler.go
│       ├── add_cart_item_handler.go
│       ├── update_cart_item_handler.go
│       ├── remove_cart_item_handler.go
│       ├── clear_cart_handler.go
│       ├── checkout_handler.go
│       ├── cancel_order_handler.go
│       ├── return_order_handler.go
│       ├── get_order_handler.go
│       ├── list_customer_orders_handler.go
│       ├── list_admin_orders_handler.go
│       ├── request.go
│       ├── response.go
│       └── routes.go
│
├── tests/
│   ├── cart_test.go
│   ├── checkout_test.go
│   ├── checkout_idempotency_test.go
│   ├── checkout_limit_test.go
│   ├── cancel_order_test.go
│   └── return_order_test.go
│
└── module.go
```

---

# 25. Módulo `billing`

```text
backend/internal/billing/
├── domain/
│   ├── billing_period.go
│   ├── billing_period_id.go
│   ├── billing_period_status.go
│   ├── billing_entry.go
│   ├── invoice.go
│   ├── invoice_id.go
│   ├── invoice_item.go
│   ├── invoice_status.go
│   ├── billing_adjustment.go
│   ├── business_day.go
│   ├── errors.go
│   ├── period_repository.go
│   ├── invoice_repository.go
│   ├── entry_repository.go
│   ├── adjustment_repository.go
│   └── calendar_repository.go
│
├── application/
│   ├── add_order_entry/
│   │   ├── command.go
│   │   └── handler.go
│   ├── add_credit_entry/
│   │   ├── command.go
│   │   └── handler.go
│   ├── close_billing_period/
│   │   ├── command.go
│   │   ├── handler.go
│   │   ├── result.go
│   │   └── calculator.go
│   ├── close_due_periods/
│   │   ├── command.go
│   │   └── handler.go
│   ├── add_invoice_adjustment/
│   │   ├── command.go
│   │   └── handler.go
│   ├── mark_overdue_invoices/
│   │   ├── command.go
│   │   └── handler.go
│   ├── apply_payment/
│   │   ├── command.go
│   │   └── handler.go
│   ├── get_current_period/
│   │   ├── query.go
│   │   └── handler.go
│   ├── get_invoice/
│   │   ├── query.go
│   │   └── handler.go
│   ├── list_customer_invoices/
│   │   ├── query.go
│   │   └── handler.go
│   ├── list_admin_invoices/
│   │   ├── query.go
│   │   └── handler.go
│   ├── calculate_fifth_business_day/
│   │   ├── query.go
│   │   └── handler.go
│   ├── upsert_business_day/
│   │   ├── command.go
│   │   └── handler.go
│   └── ports/
│       └── customer_status_writer.go
│
├── infrastructure/
│   └── postgres/
│       ├── period_repository.go
│       ├── entry_repository.go
│       ├── invoice_repository.go
│       ├── adjustment_repository.go
│       ├── calendar_repository.go
│       ├── billing_mapper.go
│       └── queries.sql
│
├── transport/
│   └── http/
│       ├── get_current_period_handler.go
│       ├── get_invoice_handler.go
│       ├── list_customer_invoices_handler.go
│       ├── list_admin_invoices_handler.go
│       ├── add_adjustment_handler.go
│       ├── close_period_handler.go
│       ├── list_calendar_handler.go
│       ├── update_calendar_handler.go
│       ├── request.go
│       ├── response.go
│       └── routes.go
│
├── jobs/
│   ├── close_due_periods_job.go
│   └── mark_overdue_invoices_job.go
│
├── tests/
│   ├── business_day_test.go
│   ├── close_period_test.go
│   ├── duplicate_close_test.go
│   ├── invoice_calculation_test.go
│   ├── adjustment_test.go
│   └── overdue_test.go
│
└── module.go
```

---

# 26. Módulo `payments`

```text
backend/internal/payments/
├── domain/
│   ├── payment_charge.go
│   ├── payment_charge_id.go
│   ├── charge_status.go
│   ├── payment.go
│   ├── payment_id.go
│   ├── payment_status.go
│   ├── payment_event.go
│   ├── provider.go
│   ├── txid.go
│   ├── errors.go
│   ├── charge_repository.go
│   ├── payment_repository.go
│   └── event_repository.go
│
├── application/
│   ├── create_pix_charge/
│   │   ├── command.go
│   │   ├── handler.go
│   │   └── result.go
│   ├── get_charge_status/
│   │   ├── query.go
│   │   └── handler.go
│   ├── process_payment_webhook/
│   │   ├── command.go
│   │   ├── handler.go
│   │   └── result.go
│   ├── reconcile_pending_charges/
│   │   ├── command.go
│   │   └── handler.go
│   ├── expire_charges/
│   │   ├── command.go
│   │   └── handler.go
│   ├── refund_payment/
│   │   ├── command.go
│   │   └── handler.go
│   ├── list_payments/
│   │   ├── query.go
│   │   └── handler.go
│   └── ports/
│       ├── payment_gateway.go
│       ├── invoice_reader.go
│       ├── invoice_payment_writer.go
│       └── customer_limit_writer.go
│
├── infrastructure/
│   ├── postgres/
│   │   ├── charge_repository.go
│   │   ├── payment_repository.go
│   │   ├── event_repository.go
│   │   ├── payment_mapper.go
│   │   └── queries.sql
│   └── gateways/
│       ├── sandbox/
│       │   ├── gateway.go
│       │   ├── client.go
│       │   ├── requests.go
│       │   ├── responses.go
│       │   └── webhook_verifier.go
│       └── provider_name/
│           ├── gateway.go
│           ├── client.go
│           ├── requests.go
│           ├── responses.go
│           └── webhook_verifier.go
│
├── transport/
│   └── http/
│       ├── create_pix_charge_handler.go
│       ├── get_charge_handler.go
│       ├── list_payments_handler.go
│       ├── payment_webhook_handler.go
│       ├── request.go
│       ├── response.go
│       └── routes.go
│
├── jobs/
│   ├── reconcile_charges_job.go
│   └── expire_charges_job.go
│
├── tests/
│   ├── create_charge_test.go
│   ├── webhook_test.go
│   ├── duplicate_webhook_test.go
│   ├── divergent_value_test.go
│   ├── reconciliation_test.go
│   └── refund_test.go
│
└── module.go
```

---

# 27. Módulo `reports`

```text
backend/internal/reports/
├── application/
│   ├── get_dashboard/
│   │   ├── query.go
│   │   ├── handler.go
│   │   └── result.go
│   ├── get_top_products/
│   │   ├── query.go
│   │   └── handler.go
│   ├── get_top_customers/
│   │   ├── query.go
│   │   └── handler.go
│   ├── get_monthly_sales/
│   │   ├── query.go
│   │   └── handler.go
│   ├── get_receivables/
│   │   ├── query.go
│   │   └── handler.go
│   ├── get_inventory_report/
│   │   ├── query.go
│   │   └── handler.go
│   ├── export_report/
│   │   ├── command.go
│   │   └── handler.go
│   └── ports/
│       ├── report_reader.go
│       └── csv_writer.go
│
├── infrastructure/
│   ├── postgres/
│   │   ├── dashboard_reader.go
│   │   ├── sales_report_reader.go
│   │   ├── customer_report_reader.go
│   │   ├── inventory_report_reader.go
│   │   └── queries.sql
│   └── export/
│       └── csv_writer.go
│
├── transport/
│   └── http/
│       ├── dashboard_handler.go
│       ├── top_products_handler.go
│       ├── top_customers_handler.go
│       ├── monthly_sales_handler.go
│       ├── receivables_handler.go
│       ├── inventory_report_handler.go
│       ├── export_handler.go
│       ├── response.go
│       └── routes.go
│
├── tests/
│   ├── dashboard_test.go
│   ├── top_products_test.go
│   ├── top_customers_test.go
│   ├── monthly_sales_test.go
│   └── reconciliation_test.go
│
└── module.go
```

---

# 28. Módulo `forecasting`

```text
backend/internal/forecasting/
├── domain/
│   ├── forecast.go
│   ├── forecast_id.go
│   ├── confidence_level.go
│   ├── forecast_method.go
│   ├── demand_history.go
│   ├── errors.go
│   └── forecast_repository.go
│
├── application/
│   ├── calculate_forecast/
│   │   ├── command.go
│   │   ├── handler.go
│   │   ├── weighted_average.go
│   │   ├── safety_stock.go
│   │   └── confidence.go
│   ├── generate_monthly_forecasts/
│   │   ├── command.go
│   │   └── handler.go
│   ├── get_forecast/
│   │   ├── query.go
│   │   └── handler.go
│   ├── list_forecasts/
│   │   ├── query.go
│   │   └── handler.go
│   └── ports/
│       ├── sales_history_reader.go
│       └── inventory_reader.go
│
├── infrastructure/
│   └── postgres/
│       ├── forecast_repository.go
│       ├── sales_history_reader.go
│       ├── inventory_reader.go
│       └── queries.sql
│
├── transport/
│   └── http/
│       ├── get_forecast_handler.go
│       ├── list_forecasts_handler.go
│       ├── response.go
│       └── routes.go
│
├── jobs/
│   └── generate_forecasts_job.go
│
├── tests/
│   ├── weighted_average_test.go
│   ├── safety_stock_test.go
│   ├── confidence_test.go
│   └── new_product_test.go
│
└── module.go
```

---

# 29. Módulo `audit`

```text
backend/internal/audit/
├── domain/
│   ├── audit_log.go
│   ├── audit_action.go
│   ├── audit_repository.go
│   └── errors.go
│
├── application/
│   ├── record_audit/
│   │   ├── command.go
│   │   └── handler.go
│   ├── list_audit_logs/
│   │   ├── query.go
│   │   └── handler.go
│   └── get_entity_history/
│       ├── query.go
│       └── handler.go
│
├── infrastructure/
│   └── postgres/
│       ├── audit_repository.go
│       └── queries.sql
│
├── transport/
│   └── http/
│       ├── list_audit_logs_handler.go
│       ├── entity_history_handler.go
│       ├── response.go
│       └── routes.go
│
├── tests/
│   └── audit_test.go
│
└── module.go
```

---

# 30. Módulo `jobs`

```text
backend/internal/jobs/
├── domain/
│   ├── job.go
│   ├── job_id.go
│   ├── job_status.go
│   ├── outbox_event.go
│   ├── outbox_status.go
│   ├── errors.go
│   ├── job_repository.go
│   └── outbox_repository.go
│
├── application/
│   ├── enqueue_job.go
│   ├── acquire_jobs.go
│   ├── complete_job.go
│   ├── fail_job.go
│   ├── publish_event.go
│   ├── acquire_events.go
│   ├── complete_event.go
│   └── fail_event.go
│
├── infrastructure/
│   └── postgres/
│       ├── job_repository.go
│       ├── outbox_repository.go
│       └── queries.sql
│
├── worker/
│   ├── runner.go
│   ├── registry.go
│   ├── retry_policy.go
│   ├── job_handler.go
│   └── event_handler.go
│
├── tests/
│   ├── acquire_job_test.go
│   ├── retry_policy_test.go
│   ├── concurrent_workers_test.go
│   └── outbox_test.go
│
└── module.go
```

---

# 31. Módulo `platform`

```text
backend/internal/platform/
├── config/
│   ├── config.go
│   ├── database.go
│   ├── http.go
│   ├── security.go
│   ├── payments.go
│   └── validation.go
│
├── database/
│   ├── pool.go
│   ├── transaction.go
│   ├── transaction_manager.go
│   └── health.go
│
├── httpx/
│   ├── error_response.go
│   ├── json.go
│   ├── pagination.go
│   ├── request_id.go
│   ├── recovery.go
│   ├── rate_limit.go
│   ├── security_headers.go
│   └── validation.go
│
├── clock/
│   ├── clock.go
│   └── system_clock.go
│
├── ids/
│   ├── generator.go
│   └── uuid_generator.go
│
├── logging/
│   ├── logger.go
│   ├── context.go
│   └── fields.go
│
├── crypto/
│   ├── encryptor.go
│   └── aes_encryptor.go
│
├── files/
│   ├── validator.go
│   └── mime.go
│
├── telemetry/
│   ├── metrics.go
│   └── health.go
│
└── errors/
    ├── application_error.go
    └── codes.go
```

---

# 32. Migrations

```text
backend/migrations/
├── 000001_initial.up.sql    # schema + seeds (catálogo demo, roles, permissões, etc.)
└── 000001_initial.down.sql  # DROP SCHEMA public CASCADE
```

Antes da produção, migrations incrementais foram consolidadas num único snapshot do schema atual.
Novas alterações voltam a ser migrations numeradas (`000002_…`, `000003_…`).

---

# 33. OpenAPI

```text
backend/openapi/
├── openapi.yaml
├── paths/
│   ├── auth.yaml
│   ├── catalog.yaml
│   ├── cart.yaml
│   ├── orders.yaml
│   ├── customers.yaml
│   ├── inventory.yaml
│   ├── billing.yaml
│   ├── payments.yaml
│   ├── reports.yaml
│   ├── forecasting.yaml
│   └── webhooks.yaml
├── schemas/
│   ├── common.yaml
│   ├── errors.yaml
│   ├── identity.yaml
│   ├── catalog.yaml
│   ├── inventory.yaml
│   ├── customers.yaml
│   ├── sales.yaml
│   ├── billing.yaml
│   ├── payments.yaml
│   ├── reports.yaml
│   └── forecasting.yaml
└── security/
    ├── session-cookie.yaml
    ├── csrf.yaml
    └── webhook-signature.yaml
```

---

# 34. Testes integrados do backend

```text
backend/tests/
├── integration/
│   ├── setup_test.go
│   ├── database_test.go
│   ├── identity_repository_test.go
│   ├── catalog_repository_test.go
│   ├── inventory_repository_test.go
│   ├── sales_repository_test.go
│   ├── billing_repository_test.go
│   ├── payments_repository_test.go
│   └── jobs_repository_test.go
│
├── e2e/
│   ├── authentication_flow_test.go
│   ├── product_to_stock_flow_test.go
│   ├── checkout_flow_test.go
│   ├── monthly_closing_flow_test.go
│   ├── pix_payment_flow_test.go
│   └── cancellation_credit_flow_test.go
│
├── contract/
│   ├── openapi_response_test.go
│   └── error_contract_test.go
│
└── fixtures/
    ├── users.go
    ├── products.go
    ├── customers.go
    ├── orders.go
    └── payments.go
```

---

# 35. Aplicação React da loja

```text
apps/store-web/
├── public/
│   ├── favicon.svg
│   └── manifest.webmanifest
│
├── src/
│   ├── app/
│   │   ├── bootstrap/
│   │   │   ├── createApp.tsx
│   │   │   └── environment.ts
│   │   ├── providers/
│   │   │   ├── AppProviders.tsx
│   │   │   ├── ApiProvider.tsx
│   │   │   ├── AuthProvider.tsx
│   │   │   └── QueryProvider.tsx
│   │   ├── router/
│   │   │   ├── router.tsx
│   │   │   ├── PublicRoute.tsx
│   │   │   ├── ProtectedRoute.tsx
│   │   │   └── routes.ts
│   │   └── layouts/
│   │       ├── StoreLayout.tsx
│   │       ├── AccountLayout.tsx
│   │       └── AuthLayout.tsx
│   │
│   ├── features/
│   │   ├── auth/
│   │   ├── registration/
│   │   ├── catalog/
│   │   ├── cart/
│   │   ├── checkout/
│   │   ├── account/
│   │   ├── orders/
│   │   ├── invoices/
│   │   ├── payments/
│   │   └── profile/
│   │
│   ├── components/
│   │   ├── AppErrorBoundary.tsx
│   │   ├── PageLoading.tsx
│   │   ├── EmptyState.tsx
│   │   └── AppLogo.tsx
│   │
│   ├── hooks/
│   │   ├── useDocumentTitle.ts
│   │   └── useStoreNavigation.ts
│   │
│   ├── lib/
│   │   ├── api.ts
│   │   ├── errors.ts
│   │   ├── routes.ts
│   │   └── environment.ts
│   │
│   ├── styles/
│   │   ├── globals.css
│   │   └── tokens.css
│   │
│   ├── main.tsx
│   └── vite-env.d.ts
│
├── tests/
│   ├── setup.ts
│   ├── fixtures/
│   └── e2e/
│       ├── catalog.spec.ts
│       ├── cart.spec.ts
│       ├── checkout.spec.ts
│       └── pix-payment.spec.ts
│
├── Dockerfile
├── index.html
├── package.json
├── tsconfig.json
└── vite.config.ts
```

## Estrutura de uma funcionalidade da loja

```text
apps/store-web/src/features/cart/
├── api/
│   ├── getCart.ts
│   ├── addCartItem.ts
│   ├── updateCartItem.ts
│   ├── removeCartItem.ts
│   └── clearCart.ts
├── components/
│   ├── CartItem.tsx
│   ├── CartItems.tsx
│   ├── CartSummary.tsx
│   └── EmptyCart.tsx
├── hooks/
│   ├── useCart.ts
│   ├── useAddCartItem.ts
│   ├── useUpdateCartItem.ts
│   └── useRemoveCartItem.ts
├── pages/
│   └── CartPage.tsx
├── schemas/
│   └── cartItemSchema.ts
├── types/
│   └── cartView.ts
└── index.ts
```

---

# 36. Aplicação React administrativa

```text
apps/admin-web/
├── public/
│   └── favicon.svg
│
├── src/
│   ├── app/
│   │   ├── bootstrap/
│   │   │   ├── createApp.tsx
│   │   │   └── environment.ts
│   │   ├── providers/
│   │   │   ├── AppProviders.tsx
│   │   │   ├── ApiProvider.tsx
│   │   │   ├── AuthProvider.tsx
│   │   │   └── QueryProvider.tsx
│   │   ├── router/
│   │   │   ├── router.tsx
│   │   │   ├── ProtectedRoute.tsx
│   │   │   ├── PermissionRoute.tsx
│   │   │   └── routes.ts
│   │   └── layouts/
│   │       ├── AdminLayout.tsx
│   │       ├── AuthLayout.tsx
│   │       └── SettingsLayout.tsx
│   │
│   ├── features/
│   │   ├── auth/
│   │   ├── dashboard/
│   │   ├── products/
│   │   ├── categories/
│   │   ├── inventory/
│   │   ├── customers/
│   │   ├── orders/
│   │   ├── billing/
│   │   ├── payments/
│   │   ├── reports/
│   │   ├── forecasting/
│   │   ├── audit/
│   │   ├── calendar/
│   │   ├── users/
│   │   └── settings/
│   │
│   ├── components/
│   │   ├── AdminNavigation.tsx
│   │   ├── AdminHeader.tsx
│   │   ├── PermissionGuard.tsx
│   │   ├── ConfirmActionDialog.tsx
│   │   ├── PageHeader.tsx
│   │   ├── PageLoading.tsx
│   │   └── AppErrorBoundary.tsx
│   │
│   ├── hooks/
│   │   ├── useAdminNavigation.ts
│   │   └── useRequiredPermission.ts
│   │
│   ├── lib/
│   │   ├── api.ts
│   │   ├── permissions.ts
│   │   ├── routes.ts
│   │   ├── errors.ts
│   │   └── environment.ts
│   │
│   ├── styles/
│   │   ├── globals.css
│   │   └── tokens.css
│   │
│   ├── main.tsx
│   └── vite-env.d.ts
│
├── tests/
│   ├── setup.ts
│   ├── fixtures/
│   └── e2e/
│       ├── product-management.spec.ts
│       ├── inventory-entry.spec.ts
│       ├── customer-limit.spec.ts
│       ├── billing-close.spec.ts
│       └── reports.spec.ts
│
├── Dockerfile
├── index.html
├── package.json
├── tsconfig.json
└── vite.config.ts
```

## Estrutura da funcionalidade de estoque

```text
apps/admin-web/src/features/inventory/
├── api/
│   ├── listInventory.ts
│   ├── getInventory.ts
│   ├── registerStockEntry.ts
│   ├── registerStockLoss.ts
│   ├── adjustInventory.ts
│   └── listMovements.ts
├── components/
│   ├── InventoryTable.tsx
│   ├── InventoryFilters.tsx
│   ├── InventoryStatusBadge.tsx
│   ├── StockEntryDialog.tsx
│   ├── StockLossDialog.tsx
│   ├── InventoryAdjustmentDialog.tsx
│   └── MovementTable.tsx
├── hooks/
│   ├── useInventory.ts
│   ├── useInventoryItem.ts
│   ├── useStockEntry.ts
│   ├── useStockLoss.ts
│   ├── useInventoryAdjustment.ts
│   └── useStockMovements.ts
├── pages/
│   ├── InventoryPage.tsx
│   └── InventoryDetailsPage.tsx
├── schemas/
│   ├── stockEntrySchema.ts
│   ├── stockLossSchema.ts
│   └── inventoryAdjustmentSchema.ts
├── types/
│   └── inventoryView.ts
└── index.ts
```

---

# 37. Aplicação React Native futura

```text
apps/mobile/
├── app/
│   ├── _layout.tsx
│   ├── index.tsx
│   ├── auth/
│   │   ├── login.tsx
│   │   └── forgot-password.tsx
│   ├── catalog/
│   │   ├── index.tsx
│   │   └── [productId].tsx
│   ├── cart/
│   │   └── index.tsx
│   ├── account/
│   │   ├── index.tsx
│   │   ├── orders.tsx
│   │   └── invoices.tsx
│   └── invoices/
│       └── [invoiceId].tsx
│
├── src/
│   ├── providers/
│   │   ├── AppProviders.tsx
│   │   ├── AuthProvider.tsx
│   │   └── QueryProvider.tsx
│   ├── components/
│   │   ├── Screen.tsx
│   │   ├── Loading.tsx
│   │   └── ErrorMessage.tsx
│   ├── features/
│   │   ├── auth/
│   │   ├── catalog/
│   │   ├── cart/
│   │   ├── account/
│   │   ├── orders/
│   │   ├── invoices/
│   │   └── payments/
│   ├── lib/
│   │   ├── api.ts
│   │   ├── secureStorage.ts
│   │   └── environment.ts
│   └── theme/
│       ├── colors.ts
│       ├── spacing.ts
│       └── typography.ts
│
├── app.json
├── eas.json
├── package.json
└── tsconfig.json
```

---

# 38. Pacotes compartilhados

## `contracts`

```text
packages/contracts/
├── src/
│   ├── generated/
│   │   ├── models.ts
│   │   ├── operations.ts
│   │   └── paths.ts
│   ├── errors.ts
│   └── index.ts
├── package.json
└── tsconfig.json
```

## `api-client`

```text
packages/api-client/
├── src/
│   ├── createApiClient.ts
│   ├── request.ts
│   ├── response.ts
│   ├── ApiError.ts
│   ├── authentication.ts
│   ├── csrf.ts
│   ├── idempotency.ts
│   └── index.ts
├── tests/
│   ├── request.test.ts
│   ├── error.test.ts
│   └── idempotency.test.ts
├── package.json
└── tsconfig.json
```

## `validation`

```text
packages/validation/
├── src/
│   ├── auth/
│   ├── customers/
│   ├── products/
│   ├── inventory/
│   ├── checkout/
│   ├── billing/
│   ├── common/
│   └── index.ts
├── tests/
├── package.json
└── tsconfig.json
```

## `shared-core`

```text
packages/shared-core/
├── src/
│   ├── money/
│   │   ├── formatMoney.ts
│   │   └── parseMoney.ts
│   ├── dates/
│   │   ├── formatDate.ts
│   │   └── formatMonth.ts
│   ├── errors/
│   │   ├── errorCodes.ts
│   │   └── errorMessages.ts
│   ├── permissions/
│   │   └── permissionCodes.ts
│   ├── status/
│   │   ├── customerStatus.ts
│   │   ├── orderStatus.ts
│   │   ├── invoiceStatus.ts
│   │   └── paymentStatus.ts
│   └── index.ts
├── tests/
├── package.json
└── tsconfig.json
```

## `design-tokens`

```text
packages/design-tokens/
├── src/
│   ├── colors.ts
│   ├── spacing.ts
│   ├── typography.ts
│   ├── radii.ts
│   ├── shadows.ts
│   ├── breakpoints.ts
│   ├── cssVariables.ts
│   └── index.ts
├── package.json
└── tsconfig.json
```

## `web-ui`

```text
packages/web-ui/
├── src/
│   ├── button/
│   │   ├── Button.tsx
│   │   ├── Button.test.tsx
│   │   └── index.ts
│   ├── form/
│   │   ├── FormField.tsx
│   │   ├── TextInput.tsx
│   │   ├── NumberInput.tsx
│   │   ├── Select.tsx
│   │   ├── Checkbox.tsx
│   │   └── index.ts
│   ├── feedback/
│   │   ├── Alert.tsx
│   │   ├── Toast.tsx
│   │   ├── ErrorMessage.tsx
│   │   ├── LoadingIndicator.tsx
│   │   └── index.ts
│   ├── overlay/
│   │   ├── Modal.tsx
│   │   ├── Dialog.tsx
│   │   └── index.ts
│   ├── data/
│   │   ├── DataTable.tsx
│   │   ├── Pagination.tsx
│   │   ├── EmptyState.tsx
│   │   └── index.ts
│   ├── display/
│   │   ├── Badge.tsx
│   │   ├── MoneyText.tsx
│   │   ├── DateText.tsx
│   │   └── index.ts
│   └── index.ts
├── package.json
└── tsconfig.json
```

## `react-hooks`

```text
packages/react-hooks/
├── src/
│   ├── useApiError.ts
│   ├── useCurrentUser.ts
│   ├── useDebounce.ts
│   ├── usePagination.ts
│   ├── usePermissions.ts
│   ├── useCopyToClipboard.ts
│   ├── useOnlineStatus.ts
│   └── index.ts
├── tests/
├── package.json
└── tsconfig.json
```

## `testing`

```text
packages/testing/
├── src/
│   ├── renderWithProviders.tsx
│   ├── createMockApi.ts
│   ├── factories/
│   │   ├── userFactory.ts
│   │   ├── customerFactory.ts
│   │   ├── productFactory.ts
│   │   ├── orderFactory.ts
│   │   └── invoiceFactory.ts
│   └── index.ts
├── package.json
└── tsconfig.json
```

---

# 39. Infraestrutura

```text
infra/
├── compose/
│   ├── compose.development.yaml
│   ├── compose.staging.yaml
│   └── compose.production.yaml
│
├── reverse-proxy/
│   ├── Caddyfile
│   ├── Dockerfile
│   └── snippets/
│       ├── security-headers.conf
│       └── rate-limit.conf
│
├── postgres/
│   ├── postgresql.conf
│   ├── pg_hba.conf
│   └── init/
│       └── 001-create-extensions.sql
│
├── backup/
│   ├── Dockerfile
│   ├── backup.sh
│   ├── restore.sh
│   ├── verify.sh
│   └── retention.sh
│
├── monitoring/
│   ├── prometheus/
│   │   └── prometheus.yml
│   ├── grafana/
│   │   ├── provisioning/
│   │   └── dashboards/
│   └── alerts/
│       ├── application.yml
│       ├── database.yml
│       ├── backup.yml
│       └── payments.yml
│
├── portainer/
│   ├── stack.env.example
│   └── README.md
│
└── scripts/
    ├── deploy.sh
    ├── rollback.sh
    ├── migrate.sh
    └── smoke-test.sh
```

---

# 40. Documentação

```text
docs/
├── requirements/
│   ├── requirements.md
│   ├── business-rules.md
│   ├── use-cases.md
│   └── traceability.md
│
├── architecture/
│   ├── overview.md
│   ├── modules.md
│   ├── data-model.md
│   ├── security.md
│   ├── deployment.md
│   └── diagrams/
│       ├── context.mmd
│       ├── containers.mmd
│       ├── components.mmd
│       ├── erd.mmd
│       ├── checkout-sequence.mmd
│       ├── billing-sequence.mmd
│       └── payment-sequence.mmd
│
├── decisions/
│   ├── ADR-0001-monolith-modular.md
│   ├── ADR-0002-react-frontends.md
│   ├── ADR-0003-rest-openapi.md
│   ├── ADR-0004-postgresql-jobs.md
│   ├── ADR-0005-payment-gateway.md
│   ├── ADR-0006-cookie-sessions.md
│   └── ADR-0007-money-in-cents.md
│
├── api/
│   ├── conventions.md
│   ├── errors.md
│   ├── pagination.md
│   ├── idempotency.md
│   └── webhooks.md
│
├── development/
│   ├── getting-started.md
│   ├── testing.md
│   ├── migrations.md
│   ├── frontend-conventions.md
│   └── backend-conventions.md
│
├── operations/
│   ├── deployment.md
│   ├── rollback.md
│   ├── backup.md
│   ├── restore.md
│   ├── monitoring.md
│   ├── incident-response.md
│   └── pix-reconciliation.md
│
└── privacy/
    ├── data-inventory.md
    ├── retention-policy.md
    ├── privacy-notice.md
    └── incident-plan.md
```

---

# 41. Ordem recomendada de implementação

A sequência técnica recomendada é:

```text
1. Fundação do repositório
2. PostgreSQL e migrations
3. Identidade e autorização
4. Auditoria e request_id
5. Catálogo
6. Estoque
7. Clientes e limite
8. Carrinho
9. Checkout transacional
10. Faturamento mensal
11. Jobs e outbox
12. Integração Pix
13. Dashboard
14. Relatórios
15. Previsão
16. Hardening de segurança
17. Backup e restauração
18. Homologação
19. Implantação
20. Preparação efetiva do React Native
```

---

# 42. Critério de conclusão do MVP

O MVP será considerado concluído quando for possível executar, em ambiente de produção, o seguinte fluxo completo:

1. gerente autentica-se com MFA;
2. gerente cadastra uma categoria;
3. gerente cadastra um produto e SKU;
4. gerente adiciona estoque;
5. cliente cadastra-se;
6. gerente aprova o cliente e define limite;
7. cliente autentica-se;
8. cliente adiciona produto ao carrinho;
9. cliente confirma a compra;
10. sistema reduz o estoque;
11. sistema atualiza a exposição;
12. sistema inclui a compra no período;
13. worker fecha o período no quinto dia útil;
14. sistema cria uma fatura;
15. cliente gera cobrança Pix;
16. PSP confirma o pagamento;
17. webhook é processado uma única vez;
18. fatura é liquidada;
19. limite do cliente é liberado;
20. gerente consulta o resultado no dashboard;
21. gerente extrai relatório;
22. sistema apresenta previsão de reposição;
23. backup é executado e pode ser restaurado.

O fluxo deverá estar coberto por testes automatizados e por um roteiro de homologação manual.

## Nota — painel Faturamento (admin)

Implementado: listagem/detalhe de faturas (`GET /admin/billing/invoices*`), resumo, fechamento manual com motivo e auditoria (`billing.close_manual`), ajustes em fatura (`billing_adjustments`), UI de calendário e detalhe formatado na loja. **Ciclos de faturamento (2026):** fechamento automático no dia 1, vencimento dia 10, fechamento parcial pelo cliente (`POST /me/billing/close-cycle`), `cycle_number` / `close_type`, lembretes D+2/D+3 via outbox. Calendário admin não governa mais o fechamento automático. Lacunas: créditos pós-fechamento (RF-VEN-017) sem fluxo dedicado; bloqueio automático por inadimplência no D+3 permanece evolução (RF-CLI-014).
