# Especificação consolidada do sistema

## Documento de requisitos, modelo entidade-relacionamento, backlog técnico e estrutura de arquivos

---

# Parte I — Documento de requisitos

## 1. Identificação do projeto

**Nome provisório:** Store Platform
**Tipo:** plataforma web de vendas pós-pagas, estoque, faturamento mensal e pagamento via Pix
**Arquitetura:** monólito modular com Clean Architecture
**Backend:** Go
**Banco de dados:** PostgreSQL
**Frontend da loja:** React e TypeScript
**Frontend administrativo:** React e TypeScript
**Aplicativo futuro:** React Native com Expo
**Infraestrutura:** Docker, Docker Compose, VPS e Portainer
**Integração financeira:** API Pix disponibilizada por PSP, banco ou instituição de pagamento

---

## 2. Objetivo do documento

Este documento define:

* requisitos de negócio;
* requisitos funcionais;
* requisitos não funcionais;
* regras de negócio;
* atores;
* casos de uso;
* critérios de aceite;
* entidades;
* relacionamentos;
* backlog técnico;
* dependências;
* estrutura completa de arquivos;
* convenções de implementação.

Os identificadores deverão ser preservados durante o desenvolvimento para permitir rastreabilidade entre:

```text
requisito
→ caso de uso
→ história
→ endpoint
→ tabela
→ caso de teste
→ módulo
```

---

## 3. Objetivo do sistema

O sistema deverá permitir que uma loja:

1. cadastre previamente os produtos comercializados;
2. controle entradas, saídas, perdas e ajustes de estoque;
3. disponibilize os produtos em uma loja virtual;
4. permita compras por clientes previamente autorizados;
5. acumule as compras na conta mensal do cliente;
6. feche o período no quinto dia útil do mês subsequente;
7. gere uma fatura;
8. permita o pagamento por Pix;
9. reconheça automaticamente o pagamento;
10. acompanhe vendas, recebimentos e estoque;
11. gere relatórios básicos;
12. estime a necessidade de reposição para o mês seguinte.

---

## 4. Escopo da primeira versão

A primeira versão compreenderá:

* autenticação;
* autorização por papéis e permissões;
* cadastro de clientes;
* aprovação de clientes;
* controle de limite;
* cadastro de produtos;
* cadastro de categorias;
* cadastro de SKUs;
* histórico de preços;
* imagens dos produtos;
* controle transacional de estoque;
* catálogo público;
* carrinho;
* checkout;
* conta mensal;
* fechamento no quinto dia útil;
* geração de fatura;
* cobrança Pix;
* webhook de pagamento;
* reconciliação;
* relatórios básicos;
* previsão simples de reposição;
* auditoria;
* backup;
* monitoramento;
* implantação por Portainer.

---

## 5. Itens fora do escopo inicial

Não farão parte do MVP:

* marketplace com vários lojistas;
* múltiplas moedas;
* múltiplos idiomas;
* programa de fidelidade;
* cupons complexos;
* promoções combinadas;
* cálculo avançado de frete;
* integração com transportadoras;
* emissão fiscal automática, salvo se obrigatória;
* Pix Automático;
* pagamento por cartão;
* aplicativo publicado;
* inteligência artificial avançada;
* microserviços;
* Kubernetes;
* múltiplos bancos de dados;
* data warehouse;
* Redis sem necessidade comprovada;
* integração com ERP;
* integração com fornecedores.

---

# 6. Glossário

| Termo                  | Definição                                                             |
| ---------------------- | --------------------------------------------------------------------- |
| Produto                | Descrição comercial geral de um item                                  |
| SKU                    | Unidade específica comercializada e controlada no estoque             |
| Saldo disponível       | Quantidade que pode ser vendida                                       |
| Movimentação           | Registro imutável de alteração do estoque                             |
| Venda                  | Compra confirmada pelo cliente                                        |
| Período de faturamento | Conjunto mensal de lançamentos de um cliente                          |
| Fatura                 | Documento consolidado após o fechamento do período                    |
| Conta pós-paga         | Modelo em que o cliente compra antes de pagar                         |
| Exposição              | Valor que o cliente já consumiu do limite                             |
| PSP                    | Provedor de serviços de pagamento                                     |
| Cobrança Pix           | Solicitação de pagamento vinculada a uma fatura                       |
| Webhook                | Notificação enviada pelo PSP ao backend                               |
| Idempotência           | Garantia de que uma operação repetida não será duplicada              |
| Quinto dia útil        | Quinta data útil, considerando fins de semana e feriados configurados |
| Snapshot               | Cópia histórica de dados no momento de uma operação                   |
| Outbox                 | Registro transacional de eventos a serem processados posteriormente   |

---

# 7. Atores

## AT-001 — Visitante

Pessoa não autenticada que poderá:

* acessar a página inicial;
* consultar produtos públicos;
* pesquisar produtos;
* consultar categorias;
* realizar cadastro;
* autenticar-se;
* solicitar recuperação de senha.

## AT-002 — Cliente

Usuário autenticado associado a uma conta de cliente.

Poderá:

* consultar catálogo;
* manter carrinho;
* concluir compras;
* consultar pedidos;
* consultar limite;
* consultar período atual;
* consultar faturas;
* gerar cobrança Pix;
* consultar pagamentos;
* atualizar dados permitidos.

## AT-003 — Gerente

Usuário responsável pela gestão comercial e operacional.

Poderá:

* administrar catálogo;
* administrar estoque;
* cadastrar e aprovar clientes;
* alterar limites;
* consultar vendas;
* consultar faturas;
* consultar pagamentos;
* gerar relatórios;
* consultar previsões;
* administrar calendário comercial.

## AT-004 — Administrador

Usuário com permissões técnicas e administrativas superiores.

Poderá:

* gerenciar usuários administrativos;
* gerenciar papéis;
* gerenciar permissões;
* configurar integrações;
* configurar parâmetros financeiros;
* consultar auditoria completa;
* executar operações especiais.

## AT-005 — Worker

Processo interno responsável por:

* fechamento mensal;
* marcação de inadimplência;
* reconciliação Pix;
* expiração de cobranças;
* geração de previsões;
* processamento da outbox;
* execução de jobs.

## AT-006 — PSP

Sistema externo que:

* cria cobranças Pix;
* consulta cobranças;
* comunica pagamentos;
* executa estornos, quando suportado.

---

# 8. Requisitos de negócio

| ID     | Requisito                                                                  |
| ------ | -------------------------------------------------------------------------- |
| RB-001 | Centralizar o cadastro dos produtos vendidos pela loja                     |
| RB-002 | Facilitar a entrada e saída de estoque pela seleção de produtos existentes |
| RB-003 | Impedir vendas acima da quantidade disponível                              |
| RB-004 | Permitir vendas pós-pagas apenas para clientes autorizados                 |
| RB-005 | Controlar limite individual por cliente                                    |
| RB-006 | Consolidar as compras por competência mensal                               |
| RB-007 | Fechar as contas no quinto dia útil do mês subsequente                     |
| RB-008 | Permitir pagamento da fatura por Pix                                       |
| RB-009 | Reconhecer automaticamente pagamentos confirmados pelo PSP                 |
| RB-010 | Separar vendas realizadas de valores efetivamente recebidos                |
| RB-011 | Disponibilizar visão atualizada do estoque                                 |
| RB-012 | Disponibilizar relatórios gerenciais básicos                               |
| RB-013 | Estimar reposição para o mês seguinte                                      |
| RB-014 | Preservar histórico de operações financeiras e de estoque                  |
| RB-015 | Permitir evolução futura para aplicativo React Native                      |
| RB-016 | Operar integralmente em contêineres Docker                                 |
| RB-017 | Permitir administração operacional por Portainer                           |

---

# 9. Requisitos funcionais

## 9.1. Identidade e autenticação

| ID         | Requisito                                                        |
| ---------- | ---------------------------------------------------------------- |
| RF-IDN-001 | O sistema deverá permitir autenticação por e-mail e senha        |
| RF-IDN-002 | O sistema deverá permitir encerramento da sessão                 |
| RF-IDN-003 | O sistema deverá permitir recuperação de senha                   |
| RF-IDN-004 | O sistema deverá armazenar somente o hash da senha               |
| RF-IDN-005 | O sistema deverá bloquear temporariamente tentativas excessivas  |
| RF-IDN-006 | O sistema deverá registrar o último acesso do usuário            |
| RF-IDN-007 | O sistema deverá permitir revogação de sessões                   |
| RF-IDN-008 | O sistema deverá exigir MFA para gerente e administrador         |
| RF-IDN-009 | O sistema deverá diferenciar cliente, gerente e administrador    |
| RF-IDN-010 | O sistema deverá verificar permissões no backend                 |
| RF-IDN-011 | O sistema deverá permitir que um usuário possua mais de um papel |
| RF-IDN-012 | O sistema deverá registrar logins administrativos na auditoria   |

**Implementação:** funcionários internos são clientes da loja com papel administrativo adicional; atribuição em `POST /admin/customers/{id}/staff-role`; suspensão administrativa (`users.status = suspended`) não impede compras na loja.

## 9.2. Clientes

| ID         | Requisito                                                       |
| ---------- | --------------------------------------------------------------- |
| RF-CLI-001 | O visitante deverá poder solicitar cadastro como cliente        |
| RF-CLI-002 | O gerente deverá poder cadastrar um cliente                     |
| RF-CLI-003 | O gerente deverá poder aprovar ou rejeitar um cadastro          |
| RF-CLI-004 | O gerente deverá poder bloquear um cliente                      |
| RF-CLI-005 | O gerente deverá poder desbloquear um cliente                   |
| RF-CLI-006 | O gerente deverá poder definir limite individual                |
| RF-CLI-007 | Toda alteração de limite deverá registrar histórico             |
| RF-CLI-008 | O cliente deverá poder consultar o limite aprovado              |
| RF-CLI-009 | O cliente deverá poder consultar o limite disponível            |
| RF-CLI-010 | O gerente deverá poder consultar a exposição do cliente         |
| RF-CLI-011 | O cliente deverá poder atualizar dados não sensíveis            |
| RF-CLI-012 | O sistema deverá impedir compras de clientes bloqueados         |
| RF-CLI-013 | O sistema deverá impedir compras acima do limite                |
| RF-CLI-014 | O sistema deverá permitir bloqueio automático por inadimplência |

## 9.3. Categorias e produtos

| ID         | Requisito                                                           |
| ---------- | ------------------------------------------------------------------- |
| RF-CAT-001 | O gerente deverá poder cadastrar categorias                         |
| RF-CAT-002 | O gerente deverá poder editar categorias                            |
| RF-CAT-003 | O gerente deverá poder inativar categorias                          |
| RF-CAT-004 | O gerente deverá poder cadastrar produtos                           |
| RF-CAT-005 | O gerente deverá poder editar produtos                              |
| RF-CAT-006 | O gerente deverá poder inativar produtos                            |
| RF-CAT-007 | O gerente deverá poder reativar produtos                            |
| RF-CAT-008 | O sistema deverá impedir exclusão física de produtos com histórico  |
| RF-CAT-009 | O gerente deverá poder definir a visibilidade do produto            |
| RF-CAT-010 | O gerente deverá poder associar produto a uma categoria             |
| RF-CAT-011 | O gerente deverá poder cadastrar um ou mais SKUs                    |
| RF-CAT-012 | O gerente deverá poder informar código de barras                    |
| RF-CAT-013 | O gerente deverá poder informar unidade de medida                   |
| RF-CAT-014 | O gerente deverá poder definir preço de venda                       |
| RF-CAT-015 | O gerente poderá informar custo de aquisição                        |
| RF-CAT-016 | O sistema deverá registrar histórico de alteração de preço          |
| RF-CAT-017 | O gerente deverá poder anexar imagens                               |
| RF-CAT-018 | O sistema deverá permitir pesquisa por nome, SKU e código de barras |
| RF-CAT-019 | O catálogo público deverá exibir apenas produtos ativos e visíveis  |
| RF-CAT-020 | O catálogo deverá permitir filtros por categoria e disponibilidade  |

## 9.4. Estoque

| ID         | Requisito                                                           |
| ---------- | ------------------------------------------------------------------- |
| RF-EST-001 | O sistema deverá manter saldo por SKU e localização                 |
| RF-EST-002 | O gerente deverá poder registrar entrada de estoque                 |
| RF-EST-003 | O gerente deverá poder registrar perda                              |
| RF-EST-004 | O gerente deverá poder registrar avaria                             |
| RF-EST-005 | O gerente deverá poder registrar ajuste de inventário               |
| RF-EST-006 | Ajustes deverão exigir justificativa                                |
| RF-EST-007 | Toda alteração deverá gerar movimentação                            |
| RF-EST-008 | Movimentações não poderão ser editadas                              |
| RF-EST-009 | O sistema deverá armazenar saldo anterior e posterior               |
| RF-EST-010 | O sistema deverá relacionar saída à venda correspondente            |
| RF-EST-011 | O sistema deverá devolver estoque em cancelamentos aplicáveis       |
| RF-EST-012 | O sistema deverá impedir saldo negativo                             |
| RF-EST-013 | O sistema deverá proteger vendas concorrentes                       |
| RF-EST-014 | O gerente deverá poder consultar histórico por SKU                  |
| RF-EST-015 | O sistema deverá identificar produtos abaixo do estoque mínimo      |
| RF-EST-016 | A entrada deverá permitir seleção de produto previamente cadastrado |
| RF-EST-017 | O sistema deverá suportar uma localização inicial                   |
| RF-EST-018 | O modelo deverá permitir múltiplas localizações futuramente         |

## 9.5. Carrinho

| ID         | Requisito                                                |
| ---------- | -------------------------------------------------------- |
| RF-CAR-001 | O cliente deverá possuir um carrinho ativo               |
| RF-CAR-002 | O cliente deverá poder adicionar item                    |
| RF-CAR-003 | O cliente deverá poder alterar quantidade                |
| RF-CAR-004 | O cliente deverá poder remover item                      |
| RF-CAR-005 | O cliente deverá poder esvaziar o carrinho               |
| RF-CAR-006 | O carrinho deverá permanecer disponível após novo acesso |
| RF-CAR-007 | O sistema deverá revalidar produtos no checkout          |
| RF-CAR-008 | O sistema deverá revalidar preços no checkout            |
| RF-CAR-009 | O sistema deverá informar itens indisponíveis            |
| RF-CAR-010 | Adicionar ao carrinho não deverá reservar estoque no MVP |

## 9.6. Vendas e pedidos

| ID         | Requisito                                                   |
| ---------- | ----------------------------------------------------------- |
| RF-VEN-001 | O cliente deverá poder confirmar uma compra                 |
| RF-VEN-002 | O backend deverá calcular todos os valores                  |
| RF-VEN-003 | O sistema deverá verificar o estado do cliente              |
| RF-VEN-004 | O sistema deverá verificar o limite disponível              |
| RF-VEN-005 | O sistema deverá verificar estoque                          |
| RF-VEN-006 | O sistema deverá registrar snapshot do nome e preço         |
| RF-VEN-007 | O sistema deverá gerar número único do pedido               |
| RF-VEN-008 | O sistema deverá impedir duplicação por idempotência        |
| RF-VEN-009 | Venda e baixa de estoque deverão ocorrer na mesma transação |
| RF-VEN-010 | A venda deverá gerar lançamento no período mensal           |
| RF-VEN-011 | O cliente deverá poder consultar seus pedidos               |
| RF-VEN-012 | O gerente deverá poder consultar pedidos                    |
| RF-VEN-013 | O gerente autorizado deverá poder cancelar pedidos          |
| RF-VEN-014 | Cancelamento deverá exigir justificativa                    |
| RF-VEN-015 | O sistema deverá suportar devolução total                   |
| RF-VEN-016 | O modelo deverá permitir devolução parcial                  |
| RF-VEN-017 | Operações posteriores ao fechamento deverão gerar crédito   |

## 9.7. Faturamento

| ID         | Requisito                                                         |
| ---------- | ----------------------------------------------------------------- |
| RF-FAT-001 | O sistema deverá manter um período por cliente e competência      |
| RF-FAT-002 | Compras deverão ser associadas ao mês da compra                   |
| RF-FAT-003 | O sistema deverá manter calendário de dias úteis                  |
| RF-FAT-004 | O gerente deverá poder cadastrar feriados locais                  |
| RF-FAT-005 | O worker deverá identificar o **dia 1** do mês (America/Sao_Paulo) para fechamento automático |
| RF-FAT-006 | O worker deverá fechar o mês anterior                                             |
| RF-FAT-007 | Cada **ciclo** (período) gera no máximo uma fatura; vários ciclos por competência são permitidos após fechamento parcial |
| RF-FAT-008 | O fechamento deverá ser idempotente                               |
| RF-FAT-009 | O sistema deverá calcular subtotal, créditos e ajustes            |
| RF-FAT-010 | O sistema deverá gerar número único de fatura                     |
| RF-FAT-011 | O cliente deverá poder consultar a fatura                         |
| RF-FAT-012 | O gerente deverá poder consultar todas as faturas                 |
| RF-FAT-013 | O gerente autorizado deverá poder lançar ajuste                   |
| RF-FAT-014 | Ajustes deverão exigir justificativa                              |
| RF-FAT-015 | O sistema deverá identificar faturas vencidas                     |
| RF-FAT-016 | O sistema deverá atualizar o estado financeiro do cliente         |
| RF-FAT-017 | O sistema deverá permitir reprocessamento de fechamento com falha |
| RF-FAT-018 | O sistema deverá registrar auditoria do fechamento                |
| RF-FAT-019 | O sistema deverá permitir fechamento manual autorizado            |
| RF-FAT-020 | O fechamento manual deverá utilizar a mesma regra de geração de fatura do worker (exceto `close_type` administrativo) |

## 9.8. Pagamento Pix

| ID         | Requisito                                                    |
| ---------- | ------------------------------------------------------------ |
| RF-PIX-001 | O cliente deverá poder gerar cobrança para uma fatura aberta |
| RF-PIX-002 | O sistema deverá impedir cobrança para fatura sem saldo      |
| RF-PIX-003 | O sistema deverá reutilizar cobrança ainda válida            |
| RF-PIX-004 | O backend deverá gerar chave de idempotência                 |
| RF-PIX-005 | O sistema deverá armazenar o identificador externo           |
| RF-PIX-006 | O sistema deverá armazenar o `txid`                          |
| RF-PIX-007 | O sistema deverá armazenar o código copia e cola             |
| RF-PIX-008 | O sistema deverá apresentar QR Code                          |
| RF-PIX-009 | O sistema deverá armazenar data de expiração                 |
| RF-PIX-010 | O sistema deverá receber webhook do PSP                      |
| RF-PIX-011 | O webhook deverá ser autenticado                             |
| RF-PIX-012 | O webhook deverá ser idempotente                             |
| RF-PIX-013 | O sistema deverá validar o valor recebido                    |
| RF-PIX-014 | O sistema deverá liquidar a fatura                           |
| RF-PIX-015 | O sistema deverá liberar o limite após pagamento             |
| RF-PIX-016 | O sistema deverá registrar pagamento e evento                |
| RF-PIX-017 | O worker deverá reconciliar cobranças pendentes              |
| RF-PIX-018 | O sistema deverá permitir geração de nova cobrança expirada  |
| RF-PIX-019 | O sistema deverá registrar divergências para análise         |
| RF-PIX-020 | O sistema deverá suportar estorno quando o PSP permitir      |

## 9.9. Dashboard

| ID         | Requisito                                        |
| ---------- | ------------------------------------------------ |
| RF-DSH-001 | O painel deverá exibir vendas do mês             |
| RF-DSH-002 | O painel deverá exibir recebimentos do mês       |
| RF-DSH-003 | O painel deverá separar competência e caixa      |
| RF-DSH-004 | O painel deverá exibir contas em aberto          |
| RF-DSH-005 | O painel deverá exibir faturas vencidas          |
| RF-DSH-006 | O painel deverá exibir ticket médio              |
| RF-DSH-007 | O painel deverá exibir clientes compradores      |
| RF-DSH-008 | O painel deverá exibir produtos abaixo do mínimo |
| RF-DSH-009 | O painel deverá exibir produto mais vendido      |
| RF-DSH-010 | O painel deverá exibir cliente com maior gasto   |
| RF-DSH-011 | O painel deverá exibir previsão de reposição     |
| RF-DSH-012 | O painel deverá permitir filtro por período      |

## 9.10. Relatórios

| ID         | Requisito                                                    |
| ---------- | ------------------------------------------------------------ |
| RF-REL-001 | O sistema deverá gerar relatório de produtos mais vendidos   |
| RF-REL-002 | O sistema deverá gerar relatório de clientes com maior gasto |
| RF-REL-003 | O sistema deverá gerar relatório mensal                      |
| RF-REL-004 | O sistema deverá gerar relatório de contas a receber         |
| RF-REL-005 | O sistema deverá gerar relatório de estoque                  |
| RF-REL-006 | O sistema deverá gerar relatório de movimentações            |
| RF-REL-007 | O sistema deverá gerar relatório de previsão                 |
| RF-REL-008 | Os relatórios deverão permitir filtros                       |
| RF-REL-009 | Os relatórios deverão permitir exportação CSV                |
| RF-REL-010 | Os totais deverão ser conciliáveis com registros detalhados  |

## 9.11. Previsões

| ID         | Requisito                                                    |
| ---------- | ------------------------------------------------------------ |
| RF-PRV-001 | O sistema deverá calcular previsão mensal por SKU            |
| RF-PRV-002 | O método inicial será média móvel ponderada                  |
| RF-PRV-003 | O sistema deverá calcular estoque de segurança               |
| RF-PRV-004 | O sistema deverá calcular sugestão de compra                 |
| RF-PRV-005 | A sugestão não poderá ser negativa                           |
| RF-PRV-006 | O sistema deverá informar o método utilizado                 |
| RF-PRV-007 | O sistema deverá informar nível de confiança                 |
| RF-PRV-008 | Produtos novos deverão ser marcados como dados insuficientes |
| RF-PRV-009 | O sistema deverá preservar snapshots das previsões           |
| RF-PRV-010 | O gerente deverá poder consultar os dados utilizados         |

## 9.12. Auditoria

| ID         | Requisito                                                |
| ---------- | -------------------------------------------------------- |
| RF-AUD-001 | O sistema deverá auditar mudança de preço                |
| RF-AUD-002 | O sistema deverá auditar mudança de limite               |
| RF-AUD-003 | O sistema deverá auditar ajuste de estoque               |
| RF-AUD-004 | O sistema deverá auditar bloqueio de cliente             |
| RF-AUD-005 | O sistema deverá auditar cancelamento                    |
| RF-AUD-006 | O sistema deverá auditar fechamento                      |
| RF-AUD-007 | O sistema deverá auditar pagamento                       |
| RF-AUD-008 | O sistema deverá auditar configuração sensível           |
| RF-AUD-009 | O log deverá conter usuário, data, ação e entidade       |
| RF-AUD-010 | Logs de auditoria não poderão ser alterados pelo gerente |
| RF-AUD-011 | O sistema deverá relacionar auditoria ao `request_id`    |
| RF-AUD-012 | Dados secretos não poderão ser gravados na auditoria     |

## 9.13. Jobs e processamento assíncrono

| ID         | Requisito                                                      |
| ---------- | -------------------------------------------------------------- |
| RF-JOB-001 | O sistema deverá manter jobs persistentes                      |
| RF-JOB-002 | Jobs deverão possuir status e número de tentativas             |
| RF-JOB-003 | Jobs com falha deverão poder ser reprocessados                 |
| RF-JOB-004 | O worker deverá impedir processamento concorrente do mesmo job |
| RF-JOB-005 | O sistema deverá manter uma outbox transacional                |
| RF-JOB-006 | Eventos processados deverão ser marcados                       |
| RF-JOB-007 | O sistema deverá preservar o erro da última tentativa          |
| RF-JOB-008 | Jobs críticos deverão gerar alerta após falhas sucessivas      |

---

# 10. Regras de negócio

## RN-001 — Exclusão lógica de produtos

Produtos, SKUs e categorias com histórico não poderão ser removidos fisicamente.

## RN-002 — Cálculo pelo backend

Preço, subtotal, desconto, crédito e total serão calculados exclusivamente pelo backend.

## RN-003 — Estoque não negativo

Nenhuma operação poderá deixar `available_quantity` abaixo de zero.

## RN-004 — Carrinho não reserva estoque

O estoque será reservado ou reduzido apenas durante a confirmação da compra.

## RN-005 — Validação no checkout

Estado do cliente, limite, preço e estoque deverão ser verificados novamente no checkout.

## RN-006 — Atomicidade da venda

Pedido, itens, estoque, movimentação e lançamento financeiro serão confirmados na mesma transação.

## RN-007 — Limite disponível

```text
limite disponível =
limite aprovado
- compras não faturadas
- faturas abertas
+ créditos aplicáveis
```

## RN-008 — Competência mensal

A compra pertence ao mês em que foi confirmada.

## RN-009 — Fechamento mensal e vencimento

O worker fecha automaticamente no **dia 1** (fuso America/Sao_Paulo) todos os períodos abertos da **competência do mês anterior**. O vencimento da fatura é o **dia 10** do mês seguinte à competência. O cliente pode solicitar **fechamento parcial** no mês (nova fatura + novo ciclo). Faturas fechadas e não pagas recebem lembrete no 2º dia após o fechamento e escalada (status `overdue`) no 3º dia.

## RN-010 — Fatura por ciclo

Cada período de faturamento (`billing_period` / ciclo) gera no máximo uma fatura; pode haver vários ciclos na mesma competência civil após fechamentos parciais.

## RN-011 — Idempotência

Checkout, fechamento, geração de cobrança e webhook deverão ser idempotentes.

## RN-012 — Pagamento confirmado

Uma fatura só poderá ser liquidada automaticamente após:

* autenticação do evento;
* identificação da cobrança;
* confirmação do valor;
* confirmação do estado do pagamento.

## RN-013 — Divergência financeira

Pagamento com valor divergente não liquidará automaticamente a fatura.

## RN-014 — Liberação de limite

O limite será liberado após a confirmação da liquidação.

## RN-015 — Venda faturada

Venda faturada não poderá ser apagada. Cancelamentos posteriores gerarão crédito ou ajuste.

## RN-016 — Movimentações imutáveis

Uma movimentação de estoque incorreta deverá ser compensada por nova movimentação.

## RN-017 — Histórico de preço

Alterar preço não modificará os valores registrados em pedidos anteriores.

## RN-018 — Dias úteis

Sábados, domingos e datas marcadas como não úteis não contarão no cálculo.

## RN-019 — Previsão inicial

A previsão inicial utilizará:

```text
50% do último mês
+ 30% do penúltimo mês
+ 20% do antepenúltimo mês
```

## RN-020 — Sugestão de compra

```text
máximo de zero entre:

demanda prevista
+ estoque de segurança
- estoque disponível
- entradas previstas
```

---

# 11. Requisitos não funcionais

## 11.1. Arquitetura

| ID          | Requisito                                                |
| ----------- | -------------------------------------------------------- |
| RNF-ARQ-001 | O backend deverá ser um monólito modular                 |
| RNF-ARQ-002 | Os módulos deverão possuir fronteiras explícitas         |
| RNF-ARQ-003 | O domínio não deverá depender de HTTP ou PostgreSQL      |
| RNF-ARQ-004 | Os casos de uso dependerão de interfaces                 |
| RNF-ARQ-005 | Handlers não deverão conter regras de negócio            |
| RNF-ARQ-006 | O sistema deverá aplicar princípios SOLID                |
| RNF-ARQ-007 | Arquivos deverão possuir responsabilidades bem definidas |
| RNF-ARQ-008 | A API será versionada                                    |
| RNF-ARQ-009 | O contrato da API será descrito em OpenAPI               |
| RNF-ARQ-010 | Loja e painel serão aplicações React separadas           |

## 11.2. Segurança

| ID          | Requisito                                            |
| ----------- | ---------------------------------------------------- |
| RNF-SEG-001 | Todo acesso externo deverá utilizar HTTPS            |
| RNF-SEG-002 | PostgreSQL não será exposto à internet               |
| RNF-SEG-003 | Cookies web deverão ser `HttpOnly` e `Secure`        |
| RNF-SEG-004 | Operações web deverão possuir proteção CSRF          |
| RNF-SEG-005 | Senhas deverão utilizar hash forte                   |
| RNF-SEG-006 | Segredos não poderão ser armazenados no Git          |
| RNF-SEG-007 | Gerentes e administradores deverão utilizar MFA      |
| RNF-SEG-008 | A API deverá aplicar rate limiting                   |
| RNF-SEG-009 | Uploads deverão ser validados                        |
| RNF-SEG-010 | Imagens Docker não deverão executar como root        |
| RNF-SEG-011 | Tokens não poderão ser registrados em logs           |
| RNF-SEG-012 | O backend deverá validar autorização em cada recurso |

## 11.3. Desempenho

| ID          | Requisito                                                                                        |
| ----------- | ------------------------------------------------------------------------------------------------ |
| RNF-DES-001 | Consultas comuns deverão responder em até 500 ms no percentil 95, excluindo integrações externas |
| RNF-DES-002 | Operações de escrita comuns deverão responder em até 1 segundo no percentil 95                   |
| RNF-DES-003 | Relatórios pesados poderão ser processados de forma assíncrona                                   |
| RNF-DES-004 | Listagens deverão utilizar paginação                                                             |
| RNF-DES-005 | Consultas deverão utilizar índices adequados                                                     |
| RNF-DES-006 | Imagens deverão possuir limite de tamanho                                                        |
| RNF-DES-007 | Frontends deverão utilizar build otimizado de produção                                           |

## 11.4. Disponibilidade e recuperação

| ID          | Requisito                                       |
| ----------- | ----------------------------------------------- |
| RNF-DIS-001 | O sistema deverá possuir health checks          |
| RNF-DIS-002 | O sistema deverá possuir backup diário          |
| RNF-DIS-003 | O backup deverá ser copiado para fora da VPS    |
| RNF-DIS-004 | A restauração deverá ser testada periodicamente |
| RNF-DIS-005 | Falhas do PSP não deverão corromper a fatura    |
| RNF-DIS-006 | Jobs deverão permitir repetição segura          |
| RNF-DIS-007 | O sistema deverá permitir rollback de imagens   |

## 11.5. Observabilidade

| ID          | Requisito                                           |
| ----------- | --------------------------------------------------- |
| RNF-OBS-001 | Logs deverão ser estruturados                       |
| RNF-OBS-002 | Cada requisição deverá possuir `request_id`         |
| RNF-OBS-003 | O sistema deverá coletar métricas básicas           |
| RNF-OBS-004 | Falhas críticas deverão gerar alerta                |
| RNF-OBS-005 | Logs não deverão conter dados sensíveis             |
| RNF-OBS-006 | O sistema deverá monitorar a idade do último backup |

## 11.6. Usabilidade

| ID          | Requisito                                                         |
| ----------- | ----------------------------------------------------------------- |
| RNF-USA-001 | Interfaces deverão ser responsivas                                |
| RNF-USA-002 | A loja deverá funcionar em navegador móvel                        |
| RNF-USA-003 | Formulários deverão apresentar erros por campo                    |
| RNF-USA-004 | Operações críticas deverão exigir confirmação                     |
| RNF-USA-005 | Estados de carregamento deverão ser visíveis                      |
| RNF-USA-006 | Falhas deverão apresentar mensagens compreensíveis                |
| RNF-USA-007 | A entrada de estoque deverá exigir poucos passos                  |
| RNF-USA-008 | A interface deverá respeitar requisitos básicos de acessibilidade |

## 11.7. Manutenibilidade

| ID          | Requisito                                                |
| ----------- | -------------------------------------------------------- |
| RNF-MAN-001 | O código deverá possuir lint e formatação automática     |
| RNF-MAN-002 | Casos de uso deverão possuir testes unitários            |
| RNF-MAN-003 | Repositórios deverão possuir testes de integração        |
| RNF-MAN-004 | Fluxos críticos deverão possuir testes end-to-end        |
| RNF-MAN-005 | Migrations aplicadas não poderão ser modificadas         |
| RNF-MAN-006 | Contratos compartilhados deverão ser gerados do OpenAPI  |
| RNF-MAN-007 | Dependências deverão ser atualizadas de forma controlada |
| RNF-MAN-008 | Decisões arquiteturais deverão ser registradas em ADRs   |

---

# 12. Casos de uso principais

## UC-001 — Autenticar usuário

**Atores:** cliente, gerente ou administrador.

**Pré-condições:**

* usuário ativo;
* credenciais cadastradas.

**Fluxo principal:**

1. usuário informa e-mail e senha;
2. frontend envia as credenciais;
3. backend localiza o usuário;
4. backend verifica o hash;
5. backend verifica o estado;
6. backend cria a sessão;
7. backend registra o acesso;
8. frontend redireciona conforme o perfil.

**Fluxos alternativos:**

* credenciais inválidas;
* usuário bloqueado;
* MFA pendente;
* excesso de tentativas.

**Pós-condição:** sessão válida criada.

---

## UC-002 — Cadastrar produto

**Ator:** gerente.

**Pré-condições:**

* gerente autenticado;
* permissão `products.write`.

**Fluxo principal:**

1. gerente abre o cadastro;
2. informa os dados;
3. frontend valida o formulário;
4. backend valida os dados;
5. backend verifica unicidade do SKU;
6. backend cria produto e SKU;
7. backend registra preço inicial;
8. backend registra auditoria;
9. produto fica disponível para entrada de estoque.

**Pós-condição:** produto cadastrado.

---

## UC-003 — Registrar entrada de estoque

**Ator:** gerente.

**Pré-condições:**

* produto ativo;
* SKU existente;
* permissão `inventory.adjust`.

**Fluxo principal:**

1. gerente pesquisa o produto;
2. seleciona o SKU;
3. informa a quantidade;
4. informa o motivo;
5. backend bloqueia o saldo;
6. backend calcula o novo saldo;
7. backend atualiza o saldo;
8. backend cria a movimentação;
9. backend registra auditoria;
10. frontend apresenta o novo saldo.

**Pós-condição:** estoque aumentado e movimentação registrada.

---

## UC-004 — Confirmar compra

**Ator:** cliente.

**Pré-condições:**

* cliente autenticado;
* cliente ativo;
* carrinho com itens.

**Fluxo principal:**

1. cliente revisa o carrinho;
2. frontend gera chave de idempotência;
3. backend verifica o cliente;
4. backend verifica o limite;
5. backend recupera os preços;
6. backend bloqueia os saldos;
7. backend valida as quantidades;
8. backend cria o pedido;
9. backend cria os itens;
10. backend reduz o estoque;
11. backend cria movimentações;
12. backend cria lançamentos financeiros;
13. backend encerra o carrinho;
14. backend confirma a transação;
15. frontend apresenta a confirmação.

**Fluxos alternativos:**

* cliente bloqueado;
* limite insuficiente;
* preço alterado;
* produto inativo;
* estoque insuficiente;
* requisição repetida.

**Pós-condição:** pedido confirmado e conta atualizada.

---

## UC-005 — Fechar período mensal

**Ator:** worker ou administrador autorizado.

**Pré-condições:**

* data correspondente ao fechamento ou execução manual autorizada;
* período anterior aberto.

**Fluxo principal:**

1. sistema identifica a competência anterior;
2. bloqueia o período;
3. reúne os lançamentos;
4. calcula subtotal;
5. aplica créditos;
6. aplica ajustes;
7. cria a fatura;
8. cria os itens da fatura;
9. fecha o período;
10. atualiza a exposição;
11. registra auditoria;
12. confirma a transação.

**Pós-condição:** fatura única criada.

---

## UC-006 — Gerar cobrança Pix

**Ator:** cliente ou gerente autorizado.

**Pré-condições:**

* fatura aberta;
* saldo maior que zero.

**Fluxo principal:**

1. sistema procura cobrança válida;
2. se não houver, cria idempotency key;
3. solicita cobrança ao PSP;
4. grava dados externos;
5. devolve QR Code e copia e cola;
6. frontend apresenta a cobrança.

**Pós-condição:** cobrança ativa associada à fatura.

---

## UC-007 — Processar webhook Pix

**Ator:** PSP.

**Pré-condições:**

* endpoint disponível;
* evento autenticável.

**Fluxo principal:**

1. backend valida a origem;
2. verifica duplicidade;
3. registra o evento;
4. localiza a cobrança;
5. verifica o valor;
6. cria o pagamento;
7. atualiza a cobrança;
8. atualiza a fatura;
9. libera o limite;
10. registra auditoria;
11. cria evento na outbox;
12. responde sucesso.

**Pós-condição:** pagamento processado uma única vez.

---

## UC-008 — Gerar previsão de reposição

**Ator:** worker.

**Pré-condições:**

* histórico mensal consolidado.

**Fluxo principal:**

1. worker carrega os últimos meses;
2. calcula a média ponderada;
3. calcula o estoque de segurança;
4. considera o estoque atual;
5. calcula a sugestão;
6. determina confiança;
7. grava o snapshot.

**Pós-condição:** previsão disponível no painel.

---

# 13. Matriz resumida de rastreabilidade

| Necessidade                 | Requisitos              | Módulos                         | Épico         |
| --------------------------- | ----------------------- | ------------------------------- | ------------- |
| Cadastrar produtos          | RF-CAT-001 a RF-CAT-020 | `catalog`                       | EP-03         |
| Controlar estoque           | RF-EST-001 a RF-EST-018 | `inventory`                     | EP-04         |
| Controlar clientes e limite | RF-CLI-001 a RF-CLI-014 | `customers`                     | EP-05         |
| Realizar venda              | RF-CAR e RF-VEN         | `sales`, `inventory`, `billing` | EP-06         |
| Fechar contas               | RF-FAT-001 a RF-FAT-020 | `billing`                       | EP-07         |
| Receber por Pix             | RF-PIX-001 a RF-PIX-020 | `payments`                      | EP-08         |
| Gerar relatórios            | RF-DSH e RF-REL         | `reports`                       | EP-09         |
| Prever reposição            | RF-PRV-001 a RF-PRV-010 | `forecasting`                   | EP-10         |
| Administrar segurança       | RF-IDN e RNF-SEG        | `identity`, `audit`             | EP-02 e EP-11 |

---
