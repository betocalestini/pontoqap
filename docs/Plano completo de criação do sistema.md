# Plano completo de criação do sistema de vendas, estoque, faturamento mensal e pagamento via Pix

## 1. Visão geral do sistema

O sistema será uma plataforma web de venda de produtos com dois ambientes principais:

1. **Loja virtual do cliente**, desenvolvida em React.
2. **Painel administrativo do gerente**, também desenvolvido em React.

O backend será desenvolvido em Go e será responsável por todas as regras de negócio, autenticação, estoque, vendas, faturamento, cobrança Pix, relatórios, auditoria e previsão de demanda.

O banco de dados será PostgreSQL. Todos os componentes serão executados em contêineres Docker e implantados em uma VPS por meio do Portainer.

A estrutura inicial também conterá um projeto React Native com Expo. A aplicação móvel não precisará ser publicada na primeira versão, mas sua base será criada desde o início para reutilizar:

* contratos da API;
* tipos TypeScript;
* schemas de validação;
* cliente HTTP;
* regras de apresentação independentes da plataforma;
* modelos de autenticação;
* tokens de design;
* formatadores;
* tratamento padronizado de erros;
* hooks sem dependência direta do navegador;
* estados e serviços reutilizáveis.

A adoção de React nos dois sistemas web reduzirá a quantidade de tecnologias diferentes no projeto, facilitará a manutenção, permitirá maior compartilhamento de código e aproximará a arquitetura da futura aplicação React Native.

---

# 2. Decisões arquiteturais fundamentais

## 2.1. Arquitetura adotada

A recomendação é utilizar:

* **monólito modular no backend**;
* **duas aplicações React independentes no frontend**;
* **API REST versionada**;
* **Clean Architecture dentro de cada módulo do backend**;
* **organização por funcionalidades nos frontends**;
* **princípios SOLID**;
* **PostgreSQL como fonte principal da verdade**;
* **processamento assíncrono por worker Go**;
* **OpenAPI como contrato da API**;
* **monorepositório para os projetos**;
* **imagens Docker construídas pela integração contínua**;
* **Portainer para implantação e administração operacional**.

Não se recomenda iniciar com microserviços. O sistema não possui, inicialmente, volume ou complexidade que justifiquem:

* comunicação distribuída;
* múltiplos bancos de dados;
* filas externas;
* descoberta de serviços;
* observabilidade distribuída;
* transações entre serviços;
* maior custo de infraestrutura;
* maior complexidade de implantação;
* duplicação de autenticação e autorização.

As divisões internas do monólito permitirão extrair futuramente módulos específicos, como pagamentos, relatórios ou notificações, caso o crescimento do sistema torne isso necessário.

## 2.2. Distribuição das aplicações React

### React: loja do cliente

A aplicação `store-web` será responsável por:

* catálogo público;
* pesquisa e filtro de produtos;
* detalhes dos produtos;
* carrinho;
* confirmação da compra;
* conta corrente do cliente;
* histórico de compras;
* consulta das faturas;
* geração e acompanhamento do Pix;
* consulta de pagamentos;
* atualização cadastral;
* recuperação de senha.

### React: painel do gerente

A aplicação `admin-web` será responsável por:

* cadastro de produtos;
* cadastro de categorias;
* cadastro de SKUs;
* alteração de preços;
* ativação e inativação de produtos;
* controle de estoque;
* acompanhamento de vendas;
* gestão de clientes;
* controle de limite de compras;
* contas a receber;
* dashboard financeiro;
* relatórios;
* previsões de demanda;
* configurações da loja;
* calendário de dias úteis;
* auditoria.

Embora ambas utilizem React, recomenda-se mantê-las como aplicações separadas. Isso permite:

* implantação independente;
* domínios diferentes;
* regras de autenticação específicas;
* bundles menores;
* separação de permissões;
* menor exposição do painel administrativo;
* possibilidade de atualizar a loja sem afetar o painel;
* identidade visual adequada a cada público.

## 2.3. React Native: aplicativo futuro

O projeto móvel estará presente desde o início no monorepositório, mas não será necessário concluí-lo para lançar a versão web.

Sua estrutura inicial conterá:

* navegação;
* autenticação;
* cliente da API;
* armazenamento seguro de credenciais;
* tema visual;
* tratamento de erros;
* componentes móveis fundamentais;
* telas-base;
* integração com os pacotes compartilhados.

Não se deve tentar reutilizar diretamente componentes HTML da loja React no React Native.

O compartilhamento deve ocorrer principalmente nas camadas sem dependência da interface:

* tipos;
* schemas;
* contratos;
* cliente HTTP;
* constantes;
* formatadores;
* validações;
* cálculos;
* regras de apresentação;
* hooks independentes de navegador;
* estados de autenticação;
* estados do carrinho;
* serviços de consulta.

## 2.4. Compartilhamento entre os dois frontends React

A utilização de React nos dois sistemas web permite compartilhar componentes e padrões adicionais, como:

* biblioteca de componentes básicos;
* formulários;
* tabelas;
* modais;
* alertas;
* tratamento de erros;
* hooks de autenticação;
* cliente OpenAPI;
* schemas de validação;
* tokens visuais;
* máscaras;
* formatadores monetários;
* componentes de carregamento;
* componentes de paginação.

Entretanto, o compartilhamento deve ser criterioso.

Componentes muito específicos da loja ou do painel devem permanecer dentro da respectiva aplicação. Não se deve criar uma biblioteca compartilhada genérica apenas para evitar pequenas duplicações.

---

# 3. Escopo funcional

## 3.1. Perfis de usuário

### Gerente

O gerente poderá:

* cadastrar produtos;
* alterar produtos;
* ativar ou inativar produtos;
* cadastrar categorias;
* cadastrar SKUs;
* definir preços;
* controlar o estoque;
* registrar entradas;
* registrar perdas;
* corrigir divergências;
* consultar o histórico de movimentações;
* acompanhar vendas;
* acompanhar recebimentos;
* consultar valores em aberto;
* cadastrar clientes;
* aprovar clientes;
* definir limite de compras por cliente;
* bloquear clientes inadimplentes;
* gerar relatórios;
* gerar previsões de reposição;
* configurar dias não úteis;
* consultar auditoria de operações permitidas.

### Cliente

O cliente poderá:

* criar uma conta ou ser cadastrado pelo gerente;
* aguardar aprovação, quando exigida;
* acessar o catálogo;
* consultar os produtos disponíveis;
* adicionar produtos ao carrinho;
* alterar quantidades;
* concluir uma compra;
* consultar compras realizadas;
* consultar saldo acumulado;
* consultar limite disponível;
* consultar faturas;
* gerar cobrança Pix;
* copiar o código Pix;
* visualizar o QR Code;
* acompanhar a confirmação do pagamento;
* consultar pagamentos anteriores;
* atualizar dados permitidos.

## 3.2. Perfil administrativo adicional

Mesmo que não seja utilizado inicialmente na rotina diária, deve existir o papel de **administrador do sistema**, separado do gerente.

O administrador poderá:

* cadastrar ou remover gerentes;
* alterar configurações sensíveis;
* desbloquear contas administrativas;
* consultar auditoria completa;
* administrar integrações;
* alterar parâmetros de cobrança;
* configurar o provedor Pix;
* alterar permissões;
* acessar configurações de infraestrutura expostas pelo sistema.

Isso evita conceder ao gerente comum permissões técnicas excessivas.

## 3.3. Possível perfil de operador

Em uma fase posterior, poderá ser criado um perfil de operador de estoque ou caixa, com permissões limitadas.

Exemplos:

* registrar entrada de estoque;
* consultar produtos;
* registrar perdas;
* consultar pedidos;
* sem acesso a faturamento;
* sem acesso a relatórios financeiros;
* sem autorização para alterar limite de clientes.

A estrutura de papéis e permissões deverá estar preparada para essa evolução.

---

# 4. Regras de negócio principais

## 4.1. Cadastro de produtos

O gerente deverá cadastrar o produto uma única vez no catálogo.

Cada produto terá, no mínimo:

* nome;
* descrição;
* SKU ou código interno;
* código de barras opcional;
* categoria;
* unidade de medida;
* preço atual;
* custo de aquisição opcional;
* estoque mínimo;
* quantidade disponível;
* imagem;
* status ativo ou inativo;
* indicação de disponibilidade para venda;
* data de criação;
* data da última alteração.

O fluxo de entrada no estoque deverá permitir ao gerente:

1. pesquisar o produto pelo nome;
2. pesquisar pelo SKU;
3. pesquisar pelo código de barras;
4. selecionar o produto cadastrado;
5. informar a quantidade;
6. informar o motivo ou documento de referência;
7. confirmar a entrada.

Dessa forma, não será necessário recadastrar o produto a cada reposição.

## 4.2. Produto inativo

A exclusão física de produtos não deve ser permitida quando houver histórico de vendas.

Ao “excluir” um produto, o sistema deverá:

* marcar o produto como inativo;
* removê-lo do catálogo público;
* impedir novas compras;
* preservar o histórico financeiro;
* preservar o histórico de estoque;
* preservar relatórios anteriores.

O gerente poderá reativar o produto posteriormente, desde que suas informações obrigatórias estejam válidas.

## 4.3. Separação entre produto e SKU

Mesmo que a primeira versão possua apenas produtos simples, recomenda-se separar conceitualmente:

* **produto**: descrição comercial;
* **SKU**: unidade efetivamente vendida e estocada.

Exemplo:

```text
Produto: Café Especial
SKU 1: Café Especial 250 g
SKU 2: Café Especial 500 g
```

Na primeira versão, cada produto poderá possuir apenas um SKU. Mesmo assim, a separação deverá existir no banco e no domínio.

Isso evita uma migração complexa caso surjam:

* tamanhos;
* cores;
* sabores;
* embalagens;
* unidades diferentes;
* códigos de barras diferentes;
* preços diferentes.

## 4.4. Controle de estoque

O estoque não deve ser controlado apenas por um campo editável de quantidade.

Deve existir um **livro imutável de movimentações**, com registros para:

* entrada por compra;
* venda;
* cancelamento;
* devolução;
* perda;
* avaria;
* correção de inventário;
* ajuste administrativo;
* transferência futura entre locais.

Cada movimentação terá:

* produto ou SKU;
* localização;
* quantidade;
* tipo;
* responsável;
* data;
* motivo;
* referência da venda, quando aplicável;
* saldo anterior;
* saldo posterior.

A quantidade disponível será atualizada em transação com a movimentação correspondente.

## 4.5. Carrinho

Adicionar produtos ao carrinho não reservará estoque no MVP.

O estoque será revalidado no momento da confirmação da compra. Isso evita que carrinhos abandonados bloqueiem produtos.

O carrinho poderá ser persistido no servidor para permitir:

* continuidade entre dispositivos;
* futura utilização pelo aplicativo;
* recuperação após encerramento do navegador;
* sincronização entre sessões;
* recuperação após login.

## 4.6. Confirmação da compra

No momento da confirmação:

1. o cliente deve estar autenticado;
2. o cliente deve estar ativo;
3. sua conta não pode estar bloqueada;
4. não pode haver fatura vencida acima da tolerância configurada;
5. o limite de compras deve ser verificado;
6. o preço deve ser recuperado novamente do banco;
7. o estoque deve ser revalidado;
8. a transação deve bloquear os registros de estoque;
9. a venda deve ser criada;
10. os itens devem registrar o preço praticado naquele momento;
11. o estoque deve ser reduzido;
12. a movimentação de estoque deve ser registrada;
13. a compra deve ser adicionada ao período de faturamento do cliente;
14. um evento de auditoria deve ser criado;
15. o carrinho deve ser encerrado;
16. a resposta deve informar o número da venda.

O frontend enviará apenas:

* identificador do produto ou SKU;
* quantidade;
* observações permitidas;
* chave de idempotência.

O frontend nunca deverá enviar o preço final como fonte confiável. O servidor calculará:

* preço unitário;
* subtotal;
* descontos;
* créditos aplicáveis;
* total final.

## 4.7. Concorrência de estoque

Para impedir que dois clientes comprem simultaneamente a última unidade, o backend deverá executar a confirmação dentro de uma transação PostgreSQL.

A operação deverá:

1. buscar a linha do estoque com bloqueio;
2. verificar a quantidade solicitada;
3. reduzir o saldo;
4. inserir a movimentação;
5. criar a venda;
6. registrar os itens;
7. criar o lançamento de faturamento;
8. confirmar a transação.

Exemplo conceitual:

```sql
SELECT available_quantity
FROM inventory_balances
WHERE sku_id = $1
FOR UPDATE;
```

Não se deve apenas reduzir a quantidade sem validar se o saldo permanecerá não negativo.

Além da validação da aplicação, deve existir uma restrição no banco:

```sql
CHECK (available_quantity >= 0)
```

## 4.8. Conta mensal do cliente

Cada cliente terá uma conta pós-paga.

As compras realizadas durante um mês serão agrupadas em um período de faturamento.

Exemplo:

* compras de 1º a 31 de agosto;
* período de referência: agosto;
* fechamento: quinto dia útil de setembro;
* compras realizadas em setembro pertencem ao período de setembro, mesmo que ocorram antes do fechamento de agosto.

A conta do cliente terá:

* limite aprovado;
* valor consumido no período atual;
* valor disponível;
* faturas abertas;
* faturas vencidas;
* pagamentos realizados;
* créditos;
* ajustes;
* situação da conta.

## 4.9. Limite de compras

Como o cliente recebe os produtos antes de pagar a fatura, existe risco de inadimplência.

O MVP deve possuir:

* aprovação do cliente pelo gerente;
* limite individual;
* limite padrão configurável;
* bloqueio manual;
* bloqueio por atraso;
* impedimento de compra acima do limite;
* registro de quem alterou o limite;
* histórico das alterações;
* possibilidade de redução ou ampliação mediante permissão.

Cálculo básico:

```text
limite disponível =
limite aprovado
- compras ainda não faturadas
- faturas abertas
+ créditos válidos
```

## 4.10. Fechamento no quinto dia útil

O sistema terá uma tabela de calendário com:

* data;
* descrição;
* tipo;
* âmbito nacional, estadual ou municipal;
* indicador de dia útil;
* responsável pela inclusão;
* data da última alteração.

Apenas verificar segunda a sexta-feira não é suficiente, pois existem feriados.

Um worker será executado diariamente no fuso `America/Sao_Paulo`.

Fluxo:

1. calcular os dias úteis do mês atual;
2. identificar o quinto dia útil;
3. verificar se a data atual corresponde ao fechamento;
4. localizar períodos do mês anterior ainda abertos;
5. bloquear os períodos para fechamento;
6. criar uma fatura por cliente;
7. copiar os itens faturáveis para a fatura;
8. calcular total bruto;
9. aplicar créditos;
10. aplicar ajustes;
11. calcular total final;
12. marcar o período como fechado;
13. tornar a fatura disponível;
14. registrar a execução do processo;
15. registrar falhas individualmente;
16. permitir reprocessamento seguro.

O processo deve ser idempotente.

Uma restrição única deve impedir duas faturas para o mesmo cliente e mês:

```text
UNIQUE (customer_id, reference_year, reference_month)
```

## 4.11. Alterações posteriores ao fechamento

Uma venda já faturada não deverá ser simplesmente apagada.

Em caso de cancelamento ou devolução após o fechamento:

* registrar a devolução;
* devolver o produto ao estoque, quando aplicável;
* registrar a movimentação;
* gerar um crédito;
* aplicar o crédito na fatura atual, se permitido;
* ou transferi-lo para o próximo período;
* registrar o motivo;
* registrar o responsável;
* preservar o lançamento original.

A política exata deverá ser configurável e registrada na auditoria.

---

# 5. Integração com Pix

## 5.1. Forma recomendada

O sistema não deve tentar se conectar diretamente à infraestrutura central do Pix como se fosse uma instituição financeira.

A integração deverá ocorrer com:

* banco do lojista;
* instituição de pagamento;
* PSP recebedor;
* gateway que ofereça API Pix.

O backend implementará uma interface genérica:

```go
type PaymentGateway interface {
    CreatePixCharge(
        ctx context.Context,
        input CreatePixChargeInput,
    ) (PixCharge, error)

    GetPixCharge(
        ctx context.Context,
        externalID string,
    ) (PixCharge, error)

    RefundPix(
        ctx context.Context,
        input RefundPixInput,
    ) error

    VerifyWebhook(
        ctx context.Context,
        request WebhookRequest,
    ) (PaymentEvent, error)
}
```

Dessa maneira, será possível trocar de provedor sem alterar os casos de uso de faturamento.

## 5.2. Cobrança Pix da fatura

Quando a fatura for fechada, o cliente poderá solicitar uma cobrança Pix.

O sistema deverá:

1. conferir se a fatura está aberta;
2. verificar se a fatura possui saldo;
3. verificar se já existe cobrança válida;
4. gerar um identificador único;
5. solicitar a cobrança ao PSP;
6. salvar o identificador externo;
7. salvar o `txid`;
8. salvar a data de expiração;
9. salvar o QR Code;
10. salvar o código copia e cola;
11. salvar o status inicial;
12. retornar esses dados ao cliente.

## 5.3. Webhook de pagamento

O endpoint será semelhante a:

```text
POST /api/v1/webhooks/payments/{provider}
```

Esse endpoint deverá:

* validar certificado, assinatura ou token do provedor;
* rejeitar origens inválidas;
* registrar o evento bruto de forma segura;
* impedir processamento duplicado;
* localizar a cobrança pelo identificador externo ou `txid`;
* comparar o valor informado;
* marcar o pagamento como recebido;
* atualizar a fatura;
* liberar o limite do cliente;
* registrar a data de liquidação;
* criar auditoria;
* responder rapidamente ao provedor.

A resposta HTTP não deverá aguardar tarefas secundárias, como:

* envio de e-mail;
* envio de notificações;
* atualização de relatórios;
* geração de documentos.

## 5.4. Idempotência

O mesmo webhook pode ser entregue mais de uma vez.

Deverá existir uma chave única:

```text
UNIQUE (provider, external_event_id)
```

Se o evento já tiver sido processado, o backend deverá responder com sucesso sem duplicar o pagamento.

Também deverão ser usadas chaves de idempotência na criação de cobranças e na confirmação de compras:

```text
Idempotency-Key: UUID
```

## 5.5. Consulta de contingência

Além do webhook, um worker periódico deverá revisar cobranças pendentes:

* cobrança expirada;
* cobrança ainda aguardando;
* cobrança paga sem webhook processado;
* divergência de valor;
* erro temporário do provedor;
* cobrança cancelada;
* fatura já paga com cobrança ainda ativa.

O webhook será o fluxo principal. A consulta será apenas uma contingência.

## 5.6. Pix Automático

O Pix Automático não é necessário no MVP.

A primeira versão deverá utilizar uma cobrança Pix dinâmica por fatura.

Em versão posterior, o sistema poderá oferecer autorização recorrente, conforme:

* disponibilidade do PSP;
* regras comerciais;
* consentimento do cliente;
* requisitos regulatórios;
* capacidade de cancelamento e gestão da autorização.

---

# 6. Estados das principais entidades

## 6.1. Venda

```text
DRAFT
CONFIRMED
CANCELLED
PARTIALLY_RETURNED
RETURNED
```

## 6.2. Período de faturamento

```text
OPEN
CLOSING
CLOSED
FAILED
```

## 6.3. Fatura

```text
OPEN
PARTIALLY_PAID
PAID
OVERDUE
CANCELLED
ADJUSTED
```

## 6.4. Cobrança Pix

```text
CREATED
ACTIVE
PAID
EXPIRED
CANCELLED
ERROR
REFUNDED
PARTIALLY_REFUNDED
```

## 6.5. Conta do cliente

```text
PENDING_APPROVAL
ACTIVE
BLOCKED
OVERDUE
CLOSED
```

Esses estados devem ser definidos no domínio e validados por métodos explícitos.

Não se deve permitir qualquer mudança arbitrária de status diretamente pelo handler HTTP ou pelo frontend.

---

# 7. Arquitetura geral

```text
                         INTERNET
                             |
                     Reverse Proxy / TLS
                             |
          +------------------+------------------+
          |                  |                  |
   loja.dominio       admin.dominio       API e webhooks
    React Web           React Web               |
          |                  |                   |
          +------------------+-------------------+
                             |
                           API Go
                             |
          +------------------+-------------------+
          |                  |                   |
      PostgreSQL          Worker Go         Armazenamento
                                             de imagens
                             |
                       PSP / API Pix
```

## 7.1. Componentes Docker

A stack de produção conterá:

| Serviço         | Função                                |
| --------------- | ------------------------------------- |
| `reverse-proxy` | TLS, domínios e encaminhamento        |
| `store-web`     | frontend React do cliente             |
| `admin-web`     | frontend React do gerente             |
| `api`           | API principal em Go                   |
| `worker`        | fechamentos, cobranças e tarefas      |
| `postgres`      | banco de dados                        |
| `backup`        | cópias do PostgreSQL e arquivos       |
| `portainer`     | gestão da infraestrutura              |
| `monitoring`    | monitoramento opcional ou progressivo |

---

# 8. Organização do repositório

```text
store-platform/
├── backend/
│   ├── cmd/
│   │   ├── api/
│   │   │   └── main.go
│   │   ├── worker/
│   │   │   └── main.go
│   │   └── migrate/
│   │       └── main.go
│   │
│   ├── internal/
│   │   ├── identity/
│   │   ├── catalog/
│   │   ├── inventory/
│   │   ├── customers/
│   │   ├── sales/
│   │   ├── billing/
│   │   ├── payments/
│   │   ├── reports/
│   │   ├── forecasting/
│   │   ├── audit/
│   │   └── platform/
│   │
│   ├── migrations/
│   ├── tests/
│   ├── go.mod
│   └── go.sum
│
├── apps/
│   ├── store-web/
│   ├── admin-web/
│   └── mobile/
│
├── packages/
│   ├── api-client/
│   ├── contracts/
│   ├── validation/
│   ├── design-tokens/
│   ├── web-ui/
│   ├── react-hooks/
│   ├── shared-core/
│   └── testing/
│
├── infra/
│   ├── compose/
│   │   ├── compose.yaml
│   │   ├── compose.development.yaml
│   │   └── compose.production.yaml
│   ├── reverse-proxy/
│   ├── postgres/
│   ├── backup/
│   └── monitoring/
│
├── docs/
│   ├── architecture/
│   ├── api/
│   ├── decisions/
│   ├── deployment/
│   └── operations/
│
├── scripts/
├── Makefile
├── package.json
├── pnpm-workspace.yaml
├── .env.example
└── README.md
```

## 8.1. Monorepositório

O monorepositório facilitará:

* sincronização dos contratos;
* execução dos testes;
* versionamento conjunto;
* criação do cliente da API;
* implantação coordenada;
* compartilhamento com React Native;
* compartilhamento entre loja e painel;
* padronização de lint;
* padronização de TypeScript;
* reutilização de componentes web;
* reutilização de hooks;
* centralização de design tokens.

Não é necessário utilizar uma plataforma pesada de monorepositório no início.

Pode-se utilizar:

* `pnpm workspaces`;
* scripts no `package.json`;
* `Makefile`;
* ferramenta de cache de builds apenas se necessária.

## 8.2. Pacote `web-ui`

O pacote `web-ui` poderá conter componentes React reutilizáveis entre a loja e o painel:

```text
packages/web-ui/
├── src/
│   ├── button/
│   ├── input/
│   ├── select/
│   ├── modal/
│   ├── alert/
│   ├── badge/
│   ├── table/
│   ├── pagination/
│   ├── loading/
│   └── index.ts
```

Não devem ser colocados nesse pacote:

* páginas completas;
* regras de negócio;
* consultas à API;
* componentes específicos de faturamento;
* componentes exclusivos do painel;
* componentes exclusivos da loja.

## 8.3. Pacote `react-hooks`

O pacote poderá conter hooks sem dependência de regras específicas de uma única aplicação:

* `useDebounce`;
* `usePagination`;
* `useCurrentUser`;
* `useLogout`;
* `useApiError`;
* `usePermissions`;
* `useOnlineStatus`;
* `useCopyToClipboard`.

Hooks específicos, como `useAdminInventoryAdjustment`, devem permanecer no painel.

---

# 9. Estrutura interna dos módulos Go

Cada módulo seguirá esta organização:

```text
internal/catalog/
├── domain/
│   ├── product.go
│   ├── product_id.go
│   ├── price.go
│   ├── errors.go
│   └── repository.go
│
├── application/
│   ├── create_product.go
│   ├── update_product.go
│   ├── deactivate_product.go
│   ├── get_product.go
│   └── list_products.go
│
├── infrastructure/
│   ├── postgres_product_repository.go
│   ├── product_queries.sql
│   └── product_mapper.go
│
├── transport/
│   └── http/
│       ├── create_product_handler.go
│       ├── update_product_handler.go
│       ├── list_products_handler.go
│       ├── request.go
│       └── response.go
│
└── module.go
```

## 9.1. Responsabilidade das camadas

### Domain

Contém:

* entidades;
* objetos de valor;
* regras invariantes;
* erros de domínio;
* interfaces necessárias pelo domínio;
* transições de estado.

Não deve importar:

* PostgreSQL;
* HTTP;
* JSON;
* frameworks web;
* bibliotecas de pagamento;
* bibliotecas de interface;
* React;
* detalhes de infraestrutura.

### Application

Contém os casos de uso:

* comandos;
* consultas;
* orquestração;
* transações;
* interfaces de serviços externos;
* DTOs internos;
* políticas de autorização de aplicação.

Exemplos:

* `CreateProduct`;
* `ConfirmOrder`;
* `CloseBillingPeriod`;
* `GeneratePixCharge`;
* `ApplyPayment`;
* `RegisterStockEntry`.

### Infrastructure

Contém implementações concretas:

* repositórios PostgreSQL;
* adaptador do PSP;
* armazenamento de imagens;
* envio de notificações;
* relógio do sistema;
* gerador de identificadores;
* acesso a serviços externos;
* implementação de transações.

### Transport

Contém:

* handlers HTTP;
* leitura de parâmetros;
* validação estrutural de entrada;
* autenticação da requisição;
* conversão dos erros para HTTP;
* serialização das respostas;
* middlewares específicos de transporte.

O handler não poderá conter regras de estoque, faturamento, limite ou pagamento.

---

# 10. Aplicação dos princípios SOLID

## 10.1. Single Responsibility Principle

Cada arquivo e estrutura terá uma responsabilidade.

Exemplos:

```text
create_product.go
confirm_order.go
close_billing_period.go
generate_pix_charge.go
process_payment_webhook.go
```

Evitar arquivos genéricos como:

```text
service.go
manager.go
helpers.go
utils.go
common.go
```

Nos frontends React, deve-se aplicar o mesmo princípio:

```text
ProductForm.tsx
ProductTable.tsx
ProductStatusBadge.tsx
useProductForm.ts
productSchema.ts
productApi.ts
```

Evitar componentes que concentrem:

* consulta;
* validação;
* formulário;
* modal;
* tabela;
* regra de permissão;
* tratamento de erro;
* navegação.

## 10.2. Open/Closed Principle

O módulo de pagamentos estará aberto para novos provedores por meio de interfaces.

Adicionar um segundo PSP exigirá um novo adaptador, e não alterações na regra de faturamento.

Nos frontends, componentes reutilizáveis devem aceitar propriedades e composição, sem depender de alterações internas para cada novo uso.

## 10.3. Liskov Substitution Principle

Implementações da interface de pagamentos ou armazenamento deverão respeitar o mesmo contrato.

Um adaptador de teste poderá substituir o adaptador real sem alterar o comportamento esperado pelo caso de uso.

## 10.4. Interface Segregation Principle

Evitar uma interface extensa:

```go
type Repository interface {
    CreateProduct(...)
    UpdateProduct(...)
    DeleteProduct(...)
    CreateOrder(...)
    CreateInvoice(...)
    SavePayment(...)
    GenerateReport(...)
}
```

Preferir interfaces menores:

```go
type ProductWriter interface {
    Save(ctx context.Context, product *Product) error
}

type ProductReader interface {
    FindByID(ctx context.Context, id ProductID) (*Product, error)
}
```

## 10.5. Dependency Inversion Principle

Os casos de uso dependerão de interfaces:

```go
type ConfirmOrder struct {
    orders     OrderRepository
    inventory  InventoryRepository
    accounts   CustomerAccountRepository
    txManager  TransactionManager
    clock      Clock
}
```

O PostgreSQL será conectado a essas interfaces na composição da aplicação, em `cmd/api/main.go`.

Nos frontends, componentes de apresentação não deverão conhecer diretamente os detalhes do cliente HTTP. Eles receberão dados e callbacks ou utilizarão hooks especializados.

---

# 11. Regras para manter arquivos curtos

Serão adotadas regras de qualidade:

* um caso de uso por arquivo;
* um handler por arquivo;
* um conceito de domínio principal por arquivo;
* um componente principal por arquivo;
* hooks separados quando acumularem lógica;
* interfaces declaradas próximas de quem as consome;
* funções pequenas;
* nomes descritivos;
* dependências explícitas;
* ausência de variáveis globais mutáveis;
* ausência de pacotes genéricos;
* divisão baseada em responsabilidade;
* revisão de arquivos extensos;
* extração de componentes apenas quando houver responsabilidade própria.

Um arquivo com 250 linhas e uma responsabilidade clara pode ser melhor que cinco arquivos artificiais. O tamanho deve ser um alerta arquitetural, não uma regra mecânica.

---

# 12. Módulos do backend

## 12.1. Identity

Responsável por:

* autenticação;
* usuários;
* papéis;
* permissões;
* sessões;
* recuperação de senha;
* bloqueio de login;
* autenticação administrativa reforçada;
* revogação de sessões;
* MFA administrativo.

## 12.2. Customers

Responsável por:

* cadastro do cliente;
* aprovação;
* dados de contato;
* status da conta;
* limite;
* bloqueios;
* histórico de alterações do limite;
* exposição financeira.

## 12.3. Catalog

Responsável por:

* produtos;
* SKUs;
* categorias;
* preços;
* imagens;
* status comercial;
* pesquisa;
* filtros;
* visibilidade pública.

## 12.4. Inventory

Responsável por:

* saldos;
* movimentações;
* entradas;
* saídas;
* ajustes;
* perdas;
* estoque mínimo;
* alertas;
* localizações.

## 12.5. Sales

Responsável por:

* carrinho;
* venda;
* itens;
* preços praticados;
* confirmação;
* cancelamento;
* devolução;
* idempotência de checkout.

## 12.6. Billing

Responsável por:

* períodos mensais;
* contas correntes;
* faturas;
* itens faturados;
* créditos;
* ajustes;
* fechamento no quinto dia útil;
* inadimplência.

## 12.7. Payments

Responsável por:

* integração com PSP;
* cobranças Pix;
* webhooks;
* liquidação;
* estornos;
* reconciliação;
* eventos externos;
* contingência.

## 12.8. Reports

Responsável por:

* consultas agregadas;
* exportação CSV;
* futura exportação XLSX;
* filtros;
* geração dos dados dos dashboards;
* relatórios financeiros;
* relatórios de estoque.

## 12.9. Forecasting

Responsável por:

* histórico mensal;
* previsão por produto;
* estoque sugerido;
* nível de confiança;
* explicação do cálculo;
* snapshots de previsão.

## 12.10. Audit

Responsável por registrar:

* mudança de preço;
* mudança de limite;
* ajuste de estoque;
* bloqueio de cliente;
* cancelamento;
* fechamento;
* pagamento;
* alteração de configuração;
* mudança de permissão;
* reprocessamento de operações.

---

# 13. Modelo de dados PostgreSQL

## 13.1. Identidade

### `users`

```text
id
name
email
phone
password_hash
status
created_at
updated_at
last_login_at
```

### `roles`

```text
id
code
name
```

### `permissions`

```text
id
code
name
```

### `user_roles`

```text
user_id
role_id
```

### `role_permissions`

```text
role_id
permission_id
```

### `sessions`

```text
id
user_id
token_hash
expires_at
revoked_at
ip_address
user_agent
created_at
```

## 13.2. Clientes

### `customers`

```text
id
user_id
document
status
credit_limit_cents
current_exposure_cents
approved_by
approved_at
blocked_reason
created_at
updated_at
```

Dados documentais só deverão ser coletados quando realmente necessários.

### `customer_limit_history`

```text
id
customer_id
previous_limit_cents
new_limit_cents
reason
changed_by
created_at
```

## 13.3. Catálogo

### `categories`

```text
id
name
slug
active
created_at
updated_at
```

### `products`

```text
id
name
slug
description
category_id
active
visible
created_at
updated_at
```

### `skus`

```text
id
product_id
code
barcode
unit
sale_price_cents
cost_price_cents
minimum_stock
active
created_at
updated_at
```

### `product_images`

```text
id
product_id
storage_key
position
alt_text
created_at
```

### `price_history`

```text
id
sku_id
previous_price_cents
new_price_cents
changed_by
reason
created_at
```

## 13.4. Estoque

### `inventory_locations`

```text
id
name
active
```

A primeira versão poderá ter apenas uma localização, mas o modelo permitirá expansão.

### `inventory_balances`

```text
id
location_id
sku_id
available_quantity
updated_at
version
```

### `stock_movements`

```text
id
location_id
sku_id
movement_type
quantity
previous_balance
new_balance
reference_type
reference_id
reason
created_by
created_at
```

## 13.5. Carrinho e vendas

### `carts`

```text
id
customer_id
status
created_at
updated_at
```

### `cart_items`

```text
id
cart_id
sku_id
quantity
created_at
updated_at
```

### `orders`

```text
id
order_number
customer_id
status
subtotal_cents
discount_cents
total_cents
idempotency_key
confirmed_at
cancelled_at
created_at
updated_at
```

### `order_items`

```text
id
order_id
sku_id
product_name_snapshot
sku_code_snapshot
unit_price_cents
quantity
total_cents
created_at
```

Os campos de snapshot preservam o nome, o código e o preço apresentados no momento da compra.

## 13.6. Faturamento

### `billing_periods`

```text
id
customer_id
reference_year
reference_month
status
opened_at
closed_at
created_at
updated_at
```

### `billing_entries`

```text
id
billing_period_id
entry_type
order_id
description
amount_cents
occurred_at
created_at
```

### `invoices`

```text
id
invoice_number
customer_id
billing_period_id
status
subtotal_cents
credit_cents
adjustment_cents
total_cents
paid_cents
due_at
closed_at
paid_at
created_at
updated_at
```

### `invoice_items`

```text
id
invoice_id
billing_entry_id
description
quantity
unit_price_cents
total_cents
created_at
```

### `billing_adjustments`

```text
id
invoice_id
adjustment_type
amount_cents
reason
created_by
created_at
```

## 13.7. Pagamentos

### `payment_charges`

```text
id
invoice_id
provider
external_id
txid
status
amount_cents
qr_code_text
qr_code_image_key
expires_at
paid_at
created_at
updated_at
```

### `payments`

```text
id
invoice_id
payment_charge_id
provider
external_payment_id
amount_cents
status
settled_at
created_at
```

### `payment_events`

```text
id
provider
external_event_id
event_type
payload_hash
payload_encrypted
processed
processed_at
error_message
created_at
```

## 13.8. Calendário

### `business_calendar`

```text
date
name
scope
is_business_day
created_at
updated_at
created_by
```

## 13.9. Auditoria

### `audit_logs`

```text
id
actor_user_id
action
entity_type
entity_id
request_id
old_values
new_values
ip_address
created_at
```

Não registrar:

* senhas;
* tokens;
* códigos secretos;
* cookies;
* credenciais Pix;
* dados desnecessários.

## 13.10. Previsões

### `forecast_snapshots`

```text
id
sku_id
reference_month
forecast_quantity
suggested_purchase_quantity
confidence_level
method
parameters
created_at
```

---

# 14. Tratamento de valores monetários

Não utilizar `float32` ou `float64` para valores financeiros.

Os valores serão representados em centavos:

```go
type Money struct {
    Cents int64
}
```

Exemplos:

```text
R$ 10,00 = 1000
R$ 168,50 = 16850
```

Benefícios:

* comparação exata;
* soma previsível;
* ausência de erros binários de ponto flutuante;
* maior facilidade de auditoria;
* simplificação das restrições do banco.

A moeda inicial será BRL.

O código deverá evitar abstrações prematuras de múltiplas moedas enquanto elas não forem necessárias.

---

# 15. API REST

## 15.1. Padrão

A API utilizará:

```text
/api/v1
```

Os contratos serão documentados em OpenAPI.

O OpenAPI servirá para:

* documentar os endpoints;
* gerar o cliente TypeScript;
* validar respostas;
* gerar mocks;
* apoiar testes;
* manter React e React Native sincronizados;
* evitar tipos duplicados manualmente.

## 15.2. Autenticação

```text
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
POST   /api/v1/auth/logout
POST   /api/v1/auth/forgot-password
POST   /api/v1/auth/reset-password
GET    /api/v1/auth/me
POST   /api/v1/auth/mfa/setup
POST   /api/v1/auth/mfa/verify
```

## 15.3. Catálogo público

```text
GET    /api/v1/catalog/products
GET    /api/v1/catalog/products/{id}
GET    /api/v1/catalog/categories
```

Filtros:

```text
?search=
?category=
?available=true
?page=
?page_size=
?sort=
```

## 15.4. Carrinho

```text
GET    /api/v1/me/cart
POST   /api/v1/me/cart/items
PATCH  /api/v1/me/cart/items/{item_id}
DELETE /api/v1/me/cart/items/{item_id}
DELETE /api/v1/me/cart
POST   /api/v1/me/cart/checkout
```

## 15.5. Conta do cliente

```text
GET    /api/v1/me/account
GET    /api/v1/me/orders
GET    /api/v1/me/orders/{id}
GET    /api/v1/me/invoices
GET    /api/v1/me/invoices/{id}
POST   /api/v1/me/invoices/{id}/pix-charge
GET    /api/v1/me/invoices/{id}/payments
```

## 15.6. Produtos administrativos

```text
GET    /api/v1/admin/products
POST   /api/v1/admin/products
GET    /api/v1/admin/products/{id}
PUT    /api/v1/admin/products/{id}
PATCH  /api/v1/admin/products/{id}/status
POST   /api/v1/admin/products/{id}/images
DELETE /api/v1/admin/products/{id}/images/{image_id}
```

## 15.7. Estoque

```text
GET    /api/v1/admin/inventory
GET    /api/v1/admin/inventory/{sku_id}
GET    /api/v1/admin/inventory/{sku_id}/movements
POST   /api/v1/admin/inventory/entries
POST   /api/v1/admin/inventory/adjustments
POST   /api/v1/admin/inventory/losses
```

## 15.8. Clientes

```text
GET    /api/v1/admin/customers
POST   /api/v1/admin/customers
GET    /api/v1/admin/customers/{id}
PUT    /api/v1/admin/customers/{id}
PATCH  /api/v1/admin/customers/{id}/status
PATCH  /api/v1/admin/customers/{id}/credit-limit
GET    /api/v1/admin/customers/{id}/account
GET    /api/v1/admin/customers/{id}/orders
GET    /api/v1/admin/customers/{id}/invoices
```

## 15.9. Faturamento

```text
GET    /api/v1/admin/billing/periods
GET    /api/v1/admin/invoices
GET    /api/v1/admin/invoices/{id}
POST   /api/v1/admin/invoices/{id}/adjustments
POST   /api/v1/admin/invoices/{id}/pix-charge
POST   /api/v1/admin/billing/close
```

O fechamento manual deverá:

* exigir permissão especial;
* exigir justificativa;
* registrar auditoria;
* utilizar a mesma lógica idempotente do worker.

## 15.10. Relatórios

```text
GET /api/v1/admin/reports/dashboard
GET /api/v1/admin/reports/top-products
GET /api/v1/admin/reports/top-customers
GET /api/v1/admin/reports/monthly-sales
GET /api/v1/admin/reports/receivables
GET /api/v1/admin/reports/inventory
GET /api/v1/admin/reports/forecast
GET /api/v1/admin/reports/export
```

## 15.11. Webhooks

```text
POST /api/v1/webhooks/payments/{provider}
```

---

# 16. Padrão de resposta da API

## 16.1. Sucesso

```json
{
  "data": {
    "id": "uuid",
    "status": "ACTIVE"
  }
}
```

## 16.2. Lista paginada

```json
{
  "data": [],
  "meta": {
    "page": 1,
    "page_size": 20,
    "total_items": 125,
    "total_pages": 7
  }
}
```

## 16.3. Erro

```json
{
  "error": {
    "code": "INSUFFICIENT_STOCK",
    "message": "A quantidade solicitada não está disponível.",
    "fields": {
      "quantity": "Quantidade superior ao estoque disponível."
    },
    "request_id": "uuid"
  }
}
```

Os frontends tomarão decisões com base em `code`, não comparando textos de mensagens.

---

# 17. Frontend React da loja

## 17.1. Estrutura

```text
apps/store-web/src/
├── app/
│   ├── router/
│   ├── providers/
│   ├── layouts/
│   └── bootstrap/
│
├── features/
│   ├── auth/
│   ├── catalog/
│   ├── cart/
│   ├── checkout/
│   ├── account/
│   ├── orders/
│   ├── invoices/
│   └── payments/
│
├── components/
├── hooks/
├── lib/
├── styles/
└── main.tsx
```

## 17.2. Estrutura interna de uma funcionalidade

```text
features/catalog/
├── api/
│   ├── getProducts.ts
│   └── getProduct.ts
├── components/
│   ├── ProductCard.tsx
│   ├── ProductGrid.tsx
│   └── ProductFilters.tsx
├── hooks/
│   ├── useProducts.ts
│   └── useProduct.ts
├── pages/
│   ├── CatalogPage.tsx
│   └── ProductPage.tsx
├── schemas/
├── types/
└── index.ts
```

## 17.3. Telas públicas

* página inicial;
* catálogo;
* detalhes do produto;
* login;
* cadastro;
* recuperação de senha;
* termos de uso;
* aviso de privacidade.

## 17.4. Telas autenticadas

* carrinho;
* confirmação da compra;
* compra concluída;
* minhas compras;
* detalhes da compra;
* minha conta;
* período atual;
* faturas;
* detalhes da fatura;
* pagamento Pix;
* perfil.

## 17.5. Estado

Separar:

* estado de servidor;
* estado local de interface;
* estado de autenticação;
* estado do carrinho;
* preferências locais.

Dados vindos da API não deverão ser duplicados desnecessariamente em stores globais.

## 17.6. Experiência do Pix

A tela deverá mostrar:

* valor;
* número da fatura;
* QR Code;
* código copia e cola;
* data de validade;
* botão para copiar;
* status;
* atualização periódica;
* confirmação visual após liquidação;
* mensagem em caso de expiração;
* opção de gerar nova cobrança, quando permitido.

---

# 18. Painel React do gerente

## 18.1. Estrutura

```text
apps/admin-web/src/
├── app/
│   ├── router/
│   ├── providers/
│   ├── layouts/
│   └── bootstrap/
│
├── features/
│   ├── dashboard/
│   ├── products/
│   ├── categories/
│   ├── inventory/
│   ├── customers/
│   ├── orders/
│   ├── billing/
│   ├── payments/
│   ├── reports/
│   ├── forecasting/
│   ├── audit/
│   └── settings/
│
├── components/
├── hooks/
├── lib/
├── styles/
└── main.tsx
```

## 18.2. Estrutura interna de produtos

```text
features/products/
├── api/
│   ├── createProduct.ts
│   ├── updateProduct.ts
│   ├── listProducts.ts
│   └── changeProductStatus.ts
├── components/
│   ├── ProductForm.tsx
│   ├── ProductTable.tsx
│   ├── ProductFilters.tsx
│   ├── ProductImageUploader.tsx
│   └── ProductStatusBadge.tsx
├── hooks/
│   ├── useProducts.ts
│   ├── useCreateProduct.ts
│   └── useUpdateProduct.ts
├── pages/
│   ├── ProductsPage.tsx
│   ├── NewProductPage.tsx
│   └── EditProductPage.tsx
├── schemas/
│   └── productSchema.ts
└── index.ts
```

## 18.3. Dashboard

O dashboard deve apresentar separadamente:

### Vendas por competência

Valor das vendas confirmadas no mês, independentemente de já terem sido pagas.

### Recebimentos por caixa

Valor efetivamente recebido via Pix durante o mês.

### Outros indicadores

* vendas do mês;
* recebimentos do mês;
* contas em aberto;
* faturas vencidas;
* ticket médio;
* número de clientes compradores;
* produtos com estoque baixo;
* produto mais vendido;
* cliente com maior compra;
* previsão de reposição;
* valor total em estoque, quando houver custo cadastrado.

Essa separação evita confundir faturamento comercial com entrada financeira.

## 18.4. Gestão de produtos

O gerente deverá conseguir:

* listar produtos;
* pesquisar por nome;
* pesquisar por SKU;
* pesquisar por código de barras;
* filtrar por categoria;
* filtrar por status;
* criar produto;
* editar produto;
* alterar preço;
* incluir imagem;
* inativar produto;
* reativar produto;
* visualizar saldo;
* abrir histórico de movimentações.

## 18.5. Gestão de estoque

A entrada de estoque deverá ser simples.

Fluxo recomendado:

1. abrir a tela de estoque;
2. localizar o produto;
3. selecionar o item;
4. clicar em “Adicionar estoque”;
5. informar a quantidade;
6. informar o motivo;
7. confirmar;
8. visualizar o novo saldo.

Para perdas e correções, deverá existir fluxo separado, com justificativa obrigatória.

## 18.6. Gestão de clientes

O painel deverá permitir:

* cadastrar cliente;
* aprovar cadastro;
* bloquear cliente;
* desbloquear cliente;
* alterar limite;
* consultar histórico de limite;
* consultar compras;
* consultar faturas;
* consultar pagamentos;
* consultar saldo devedor;
* consultar limite disponível.

---

# 19. Compartilhamento entre React Web e React Native

## 19.1. Elementos compartilháveis

Poderão ser compartilhados:

* contratos OpenAPI;
* tipos;
* schemas;
* regras de validação;
* cliente HTTP;
* códigos de erro;
* formatadores;
* cálculos;
* lógica de autenticação;
* lógica do carrinho;
* hooks independentes de plataforma;
* design tokens;
* textos padronizados.

## 19.2. Elementos que não devem ser compartilhados diretamente

Não deverão ser compartilhados diretamente:

* elementos HTML;
* componentes com `div`, `button` ou `input`;
* componentes dependentes do DOM;
* armazenamento baseado em `localStorage`;
* APIs específicas do navegador;
* navegação específica do React Router;
* estilos CSS;
* tabelas administrativas;
* componentes muito específicos da loja.

## 19.3. Separação de pacotes

Exemplo:

```text
packages/
├── contracts/
├── api-client/
├── validation/
├── shared-core/
├── design-tokens/
├── react-hooks/
└── web-ui/
```

O `web-ui` será utilizado apenas pelas aplicações web.

O React Native utilizará:

* `contracts`;
* `api-client`;
* `validation`;
* `shared-core`;
* `design-tokens`;
* parte de `react-hooks`, quando compatível.

---

# 20. Relatórios

## 20.1. Produtos mais vendidos

Filtros:

* período;
* categoria;
* produto;
* quantidade mínima;
* incluir ou não devoluções.

Colunas:

* posição;
* produto;
* SKU;
* quantidade vendida;
* valor bruto;
* devoluções;
* valor líquido;
* participação no total.

## 20.2. Clientes com maior gasto

Colunas:

* cliente;
* total comprado;
* total pago;
* saldo aberto;
* número de pedidos;
* ticket médio;
* última compra;
* situação da conta.

## 20.3. Relatório mensal

Apresentará:

* vendas brutas;
* cancelamentos;
* devoluções;
* vendas líquidas;
* recebimentos;
* contas a receber;
* inadimplência;
* produtos vendidos;
* ticket médio;
* clientes ativos.

## 20.4. Estoque

Apresentará:

* quantidade atual;
* estoque mínimo;
* consumo médio;
* dias estimados de cobertura;
* última entrada;
* última venda;
* sugestão de compra;
* itens abaixo do mínimo;
* itens sem movimentação.

## 20.5. Exportação

No MVP:

* CSV;
* impressão pelo navegador;
* filtros refletidos na exportação.

Posteriormente:

* XLSX;
* PDF padronizado;
* envio programado por e-mail.

---

# 21. Previsão de compra para o mês seguinte

## 21.1. Estratégia inicial

Não se recomenda iniciar com inteligência artificial ou modelos complexos.

O sistema terá poucos dados próprios nos primeiros meses.

A primeira previsão deve ser:

* explicável;
* auditável;
* simples;
* estável;
* fácil de corrigir.

Método inicial recomendado: média móvel ponderada.

Exemplo:

```text
previsão =
50% das vendas do último mês
+ 30% das vendas de dois meses atrás
+ 20% das vendas de três meses atrás
```

## 21.2. Sugestão de reposição

```text
sugestão de compra =
demanda prevista
+ estoque de segurança
- estoque disponível
- entradas já previstas
```

O resultado mínimo será zero.

## 21.3. Estoque de segurança

Configuração inicial:

```text
estoque de segurança =
demanda média diária
× dias de segurança
```

O gerente poderá definir:

* prazo médio de reposição;
* dias de segurança;
* estoque mínimo manual;
* multiplicador para produtos sazonais.

## 21.4. Produtos sem histórico

Para produtos novos:

* utilizar estoque mínimo cadastrado;
* permitir previsão manual;
* indicar “dados insuficientes”;
* não apresentar falsa precisão;
* sugerir revisão pelo gerente.

## 21.5. Nível de confiança

Cada previsão terá:

```text
BAIXA
MÉDIA
ALTA
```

Critérios possíveis:

* quantidade de meses disponíveis;
* regularidade das vendas;
* volume de devoluções;
* variação mensal;
* existência de meses sem venda;
* períodos sem estoque.

## 21.6. Evolução futura

Depois de acumular histórico, poderão ser avaliados:

* suavização exponencial;
* sazonalidade;
* tendência;
* demanda intermitente;
* eventos promocionais;
* dias sem estoque;
* prazo do fornecedor;
* comparação entre previsão e realizado.

O sistema deverá registrar qual método produziu cada previsão.

---

# 22. Autenticação e autorização

## 22.1. Aplicações web React

Para `store-web` e `admin-web`:

* sessão por cookie seguro;
* cookie `HttpOnly`;
* cookie `Secure`;
* política `SameSite`;
* proteção CSRF;
* expiração por inatividade;
* revogação no logout;
* renovação controlada;
* sessões armazenadas com token em hash.

O painel administrativo deverá possuir política mais restritiva:

* MFA obrigatório;
* menor duração de sessão;
* reautenticação para operações sensíveis;
* bloqueio progressivo;
* auditoria de login.

## 22.2. Aplicativo futuro

O React Native utilizará:

* access token curto;
* refresh token rotativo;
* armazenamento seguro do dispositivo;
* revogação de sessão;
* identificação do dispositivo;
* renovação controlada.

O backend poderá utilizar transportes diferentes para web e mobile, mantendo o mesmo serviço de autenticação.

## 22.3. Permissões

Exemplos:

```text
products.read
products.write
inventory.read
inventory.adjust
customers.read
customers.approve
customers.change_limit
orders.read
orders.cancel
billing.read
billing.close
payments.read
reports.read
settings.write
audit.read
```

O papel concederá permissões. O código deverá verificar permissões, não apenas nomes de papéis.

## 22.4. Proteção de rotas React

As aplicações React poderão ocultar rotas e componentes conforme as permissões do usuário.

Entretanto, essa proteção será apenas uma camada de interface.

A autorização definitiva sempre será executada no backend.

---

# 23. Segurança

## 23.1. Controles mínimos

* TLS obrigatório;
* senhas com algoritmo de hash forte;
* rate limiting;
* bloqueio progressivo de login;
* MFA para administrador e gerente;
* validação de todas as entradas;
* consultas parametrizadas;
* autorização por recurso;
* proteção contra acesso indevido a identificadores;
* proteção CSRF;
* cabeçalhos de segurança;
* limite de tamanho de uploads;
* validação do tipo real das imagens;
* varredura de dependências;
* imagens Docker sem execução como root;
* segredos fora do Git;
* logs sem dados sensíveis;
* auditoria financeira;
* backups criptografados;
* política de retenção.

## 23.2. Segurança dos webhooks

* assinatura ou autenticação obrigatória;
* limite de tamanho;
* timeout curto;
* deduplicação;
* comparação do valor;
* validação do PSP;
* armazenamento do hash do payload;
* resposta idempotente;
* alertas de falha;
* rejeição de métodos não permitidos.

## 23.3. Auditoria financeira

Operações críticas devem registrar:

* usuário;
* data;
* IP;
* requisição;
* valor anterior;
* valor posterior;
* entidade;
* motivo;
* identificador de correlação.

A auditoria não deverá ser editável pelo gerente.

## 23.4. Segurança dos frontends React

* não armazenar tokens sensíveis no `localStorage`;
* não incorporar segredos no bundle;
* escapar conteúdo dinâmico;
* evitar HTML arbitrário;
* controlar dependências;
* utilizar política de conteúdo;
* validar redirecionamentos;
* tratar erros sem expor detalhes internos;
* limitar informações exibidas pelo painel.

---

# 24. LGPD e privacidade

O sistema tratará dados pessoais de clientes, incluindo:

* nome;
* contato;
* compras;
* pagamentos;
* situação financeira perante a loja;
* histórico de acesso;
* dados cadastrais necessários.

Devem ser previstos:

* aviso de privacidade;
* finalidade de cada dado;
* coleta mínima;
* política de retenção;
* canal para solicitação do titular;
* correção de dados;
* exclusão ou anonimização quando cabível;
* controle de acesso;
* registro das operações de tratamento;
* plano de incidentes;
* contratos com operadores;
* definição do responsável pelo tratamento.

Não se deve armazenar:

* senha em texto;
* credencial completa do PSP;
* token Pix em log;
* payload sensível sem necessidade;
* documento pessoal em mais tabelas que o necessário;
* informações financeiras sem finalidade definida.

A política jurídica, fiscal, consumerista e contábil da venda pós-paga deverá ser validada antes da entrada em produção.

---

# 25. Banco de dados e migrations

## 25.1. Acesso pelo Go

Sugestão:

* driver PostgreSQL nativo para Go;
* pool de conexões;
* consultas parametrizadas;
* geração opcional de código SQL tipado;
* structs de infraestrutura separadas das entidades de domínio;
* transações explícitas.

## 25.2. Migrations

As alterações serão registradas em arquivos versionados:

```text
000001_create_users.up.sql
000001_create_users.down.sql
000002_create_catalog.up.sql
000002_create_catalog.down.sql
```

As migrations serão executadas por um contêiner próprio antes da atualização da API.

## 25.3. Regras

* nunca alterar migration aplicada em produção;
* criar nova migration;
* revisar índices;
* realizar backup antes de alterações destrutivas;
* evitar migration destrutiva junto com remoção imediata de código;
* utilizar estratégia compatível em duas etapas;
* validar migrations em ambiente de homologação;
* registrar versões aplicadas.

---

# 26. Processamento assíncrono

O worker será outro executável Go construído a partir do mesmo código.

```text
backend/cmd/worker/main.go
```

Tarefas:

* fechamento mensal;
* identificação de inadimplência;
* expiração de cobranças;
* reconciliação Pix;
* atualização de previsões;
* alertas de estoque;
* limpeza de sessões;
* processamento de notificações;
* atualização de agregações.

## 26.1. Sem Redis inicialmente

Para manter simplicidade, o MVP poderá utilizar uma tabela PostgreSQL de jobs.

### `jobs`

```text
id
type
payload
status
attempts
available_at
locked_at
locked_by
last_error
created_at
completed_at
```

O worker buscará tarefas utilizando bloqueio compatível com múltiplos consumidores.

Redis ou uma fila dedicada só deverá ser incluído quando houver necessidade comprovada.

## 26.2. Outbox

Operações que precisem gerar eventos deverão inserir o evento na mesma transação da operação principal.

Exemplo:

1. pagamento é salvo;
2. fatura é atualizada;
3. evento `INVOICE_PAID` é inserido na outbox;
4. transação é confirmada;
5. worker processa o evento.

Isso evita atualizar a fatura e perder uma ação secundária devido a uma falha entre operações.

---

# 27. Docker

## 27.1. Imagens

Cada aplicação terá um Dockerfile multi-stage.

Backend:

```text
etapa de build Go
→ geração do binário
→ imagem final mínima
```

Frontends React:

```text
etapa Node para build
→ geração dos arquivos estáticos
→ servidor web mínimo
```

## 27.2. Produção

O Compose de produção não deverá compilar o código dentro da VPS.

Fluxo recomendado:

1. integração contínua executa testes;
2. constrói imagens;
3. envia as imagens para um registry;
4. Portainer atualiza a stack;
5. contêiner de migration é executado;
6. API e worker são atualizados;
7. loja e painel são atualizados;
8. health checks são verificados.

As imagens devem utilizar tags imutáveis:

```text
api:git-a1b2c3d
worker:git-a1b2c3d
store-web:git-a1b2c3d
admin-web:git-a1b2c3d
```

Evitar depender exclusivamente de:

```text
latest
```

## 27.3. Ambiente de desenvolvimento

O Compose de desenvolvimento poderá:

* montar volumes de código;
* executar hot reload;
* disponibilizar PostgreSQL local;
* executar API, worker e frontends;
* utilizar credenciais de teste;
* integrar com sandbox do PSP.

---

# 28. Portainer

A stack deverá ser vinculada a um repositório Git contendo a definição da infraestrutura.

O Portainer poderá:

* puxar o Compose;
* atualizar imagens;
* visualizar logs;
* reiniciar serviços;
* administrar volumes;
* verificar saúde dos contêineres;
* acompanhar utilização básica de recursos.

O Portainer não deverá ser a ferramenta principal de build.

As imagens deverão chegar prontas ao servidor.

## 28.1. Segurança do Portainer

* não expor sem proteção;
* restringir por VPN ou lista de IPs;
* habilitar MFA;
* utilizar senha forte;
* limitar contas administrativas;
* manter atualizado;
* não reutilizar credenciais;
* registrar alterações;
* usar HTTPS.

---

# 29. Estrutura da VPS

## 29.1. Portas externas

Expor apenas:

```text
80
443
```

O SSH deverá ser:

* protegido por chave;
* restrito por firewall;
* sem login direto de root;
* protegido contra tentativas repetidas;
* registrado em logs.

PostgreSQL não será exposto à internet.

## 29.2. Redes Docker

```text
public-network
application-network
database-network
management-network
```

Exemplo:

* reverse proxy acessa frontends e API;
* API e worker acessam PostgreSQL;
* frontend não acessa PostgreSQL;
* PostgreSQL não acessa a rede pública;
* Portainer fica isolado;
* backup acessa apenas os volumes necessários.

## 29.3. Volumes

```text
postgres-data
product-images
portainer-data
backup-data
reverse-proxy-data
```

## 29.4. Dimensionamento inicial

Para uma operação pequena, a VPS deverá priorizar:

* armazenamento SSD;
* memória suficiente para PostgreSQL e contêineres;
* backup externo;
* possibilidade de expansão;
* monitoramento de disco.

O dimensionamento deverá ser revisto conforme:

* quantidade de produtos;
* número de clientes;
* volume de imagens;
* requisições;
* relatórios;
* retenção de auditoria.

---

# 30. Domínios

Sugestão:

```text
loja.exemplo.com
admin.exemplo.com
```

A API poderá ser encaminhada pelo mesmo domínio:

```text
loja.exemplo.com/api
admin.exemplo.com/api
```

Isso simplifica:

* cookies;
* CORS;
* segurança;
* configuração dos frontends.

Webhooks poderão utilizar domínio próprio:

```text
hooks.exemplo.com
```

ou rota dedicada:

```text
loja.exemplo.com/api/v1/webhooks/payments/provider
```

---

# 31. Backups

## 31.1. Conteúdo

Realizar backup de:

* PostgreSQL;
* imagens;
* configurações necessárias;
* arquivos de infraestrutura;
* certificados quando não forem recriáveis;
* documentação operacional;
* segredos por meio de solução segura apropriada.

## 31.2. Estratégia

* backup automatizado diário;
* backup adicional antes de migrations críticas;
* cópia fora da VPS;
* criptografia;
* retenção por períodos;
* verificação de integridade;
* teste periódico de restauração;
* alerta em caso de falha.

Um backup sem teste de restauração não deverá ser considerado confiável.

## 31.3. O que não fazer

* manter backup somente no mesmo disco;
* salvar credenciais em texto;
* copiar volume PostgreSQL em execução sem método consistente;
* confiar apenas em snapshot da VPS;
* nunca testar restauração;
* manter backups indefinidamente sem política.

---

# 32. Observabilidade

## 32.1. Logs

Todos os serviços deverão registrar logs estruturados:

```json
{
  "level": "info",
  "event": "order_confirmed",
  "order_id": "uuid",
  "customer_id": "uuid",
  "request_id": "uuid",
  "timestamp": "..."
}
```

Não registrar:

* senha;
* token;
* cookie;
* QR Code completo;
* credencial do PSP;
* documento sem mascaramento;
* payload financeiro sensível desnecessário.

## 32.2. Métricas

* requisições por minuto;
* taxa de erro;
* latência;
* conexões PostgreSQL;
* jobs pendentes;
* jobs com falha;
* webhooks rejeitados;
* cobranças pendentes;
* falhas de fechamento;
* espaço em disco;
* memória;
* CPU;
* idade do último backup;
* falhas de login;
* pedidos recusados por falta de estoque.

## 32.3. Health checks

```text
GET /health/live
GET /health/ready
```

`live` verifica se o processo está funcionando.

`ready` verifica se está apto a receber tráfego, incluindo conexão com serviços essenciais.

## 32.4. Identificador de correlação

Cada requisição deverá receber um `request_id`.

Esse identificador acompanhará:

* logs;
* erros;
* auditoria;
* chamadas externas;
* respostas da API.

Isso facilitará a investigação de falhas.

---

# 33. Estratégia de testes

## 33.1. Testes unitários do backend

Priorizar:

* cálculo de limite;
* transições de estado;
* fechamento;
* quinto dia útil;
* cálculo da fatura;
* aplicação de créditos;
* estoque;
* previsão;
* cálculo monetário;
* idempotência.

## 33.2. Testes unitários dos frontends React

Priorizar:

* schemas;
* formatadores;
* hooks;
* regras de exibição;
* componentes de formulário;
* componentes de permissão;
* tratamento de erro;
* estados do carrinho.

## 33.3. Testes de integração

Executados com PostgreSQL real em contêiner:

* repositories;
* transações;
* locks;
* constraints;
* migrations;
* concorrência de estoque;
* idempotência;
* fechamento;
* pagamentos.

## 33.4. Testes de contrato

Validar:

* backend contra OpenAPI;
* cliente TypeScript contra OpenAPI;
* respostas de erro;
* compatibilidade dos frontends;
* compatibilidade futura do React Native.

## 33.5. Testes end-to-end

Fluxos essenciais:

1. gerente cadastra produto;
2. gerente adiciona estoque;
3. cliente visualiza o produto;
4. cliente adiciona ao carrinho;
5. cliente compra;
6. estoque é reduzido;
7. compra entra na conta;
8. período é fechado;
9. fatura é gerada;
10. Pix é criado;
11. webhook é recebido;
12. fatura é paga;
13. limite é liberado.

## 33.6. Testes de concorrência

Cenário obrigatório:

* produto com uma unidade;
* dois checkouts simultâneos;
* apenas um deve ser confirmado;
* o saldo nunca pode ficar negativo.

## 33.7. Testes do calendário

Cobrir:

* mês começando em sábado;
* feriado na primeira semana;
* dois feriados consecutivos;
* fevereiro;
* ano bissexto;
* alteração de feriado;
* execução duplicada do worker;
* fechamento parcial com falha.

## 33.8. Testes do Pix

* criação com sucesso;
* timeout do PSP;
* webhook duplicado;
* valor divergente;
* pagamento sem fatura;
* cobrança expirada;
* pagamento após expiração;
* estorno;
* assinatura inválida;
* evento fora de ordem.

## 33.9. Testes de permissões

Cobrir:

* cliente tentando acessar outro cliente;
* gerente sem permissão específica;
* operador tentando alterar limite;
* usuário bloqueado;
* sessão expirada;
* rota administrativa acessada pelo frontend da loja.

---

# 34. Integração e entrega contínuas

## 34.1. Pipeline de pull request

* formatação Go;
* análise estática;
* testes unitários;
* testes de integração;
* lint da loja React;
* lint do painel React;
* verificação TypeScript;
* testes dos frontends;
* validação OpenAPI;
* varredura de dependências;
* varredura de imagens;
* build completo.

## 34.2. Pipeline principal

* gerar versão;
* criar imagens;
* publicar no registry;
* atualizar manifesto de implantação;
* executar migration;
* atualizar stack;
* verificar health checks;
* executar teste básico pós-implantação.

## 34.3. Rollback

O rollback deverá considerar separadamente:

* imagens;
* migrations;
* configurações;
* frontends;
* API;
* worker.

Migrations destrutivas dificultam rollback. Por isso, mudanças de banco deverão ser compatíveis com a versão anterior durante a transição.

## 34.4. Ambientes

Recomenda-se manter:

```text
development
staging
production
```

O ambiente de homologação deverá utilizar:

* banco separado;
* credenciais separadas;
* sandbox do PSP;
* domínio separado;
* dados fictícios.

---

# 35. Etapas de desenvolvimento

## Etapa 0 — Descoberta e definição

### Atividades

* confirmar o funcionamento comercial;
* definir se haverá retirada ou entrega;
* definir política de cancelamento;
* definir limite padrão;
* definir inadimplência;
* definir vencimento da fatura;
* selecionar PSP;
* validar exigências fiscais e jurídicas;
* definir dados obrigatórios do cliente;
* definir relatórios prioritários;
* definir identidade visual;
* mapear permissões;
* definir feriados aplicáveis.

### Entregáveis

* documento de regras;
* fluxos;
* glossário;
* protótipos;
* mapa de permissões;
* modelo de dados inicial;
* critérios de aceite.

## Etapa 1 — Fundação técnica

### Atividades

* criar monorepositório;
* configurar backend Go;
* configurar loja React;
* configurar painel React;
* criar projeto Expo;
* criar pacotes compartilhados;
* configurar PostgreSQL;
* criar Compose;
* configurar migrations;
* configurar OpenAPI;
* configurar pipeline;
* criar health checks;
* criar logging estruturado.

### Critério de conclusão

Todos os projetos compilam, executam em Docker e comunicam-se em ambiente local.

## Etapa 2 — Identidade e acesso

### Atividades

* usuários;
* papéis;
* permissões;
* login;
* logout;
* sessão;
* recuperação de senha;
* MFA administrativo;
* proteção CSRF;
* rate limiting;
* auditoria de login;
* proteção de rotas React.

### Critério de conclusão

Cliente e gerente acessam apenas os recursos permitidos.

## Etapa 3 — Catálogo

### Atividades

* categorias;
* produtos;
* SKU;
* preços;
* imagens;
* ativação;
* inativação;
* catálogo público;
* pesquisa;
* paginação;
* filtros.

### Critério de conclusão

O gerente cadastra um produto e ele aparece corretamente na loja React.

## Etapa 4 — Estoque

### Atividades

* saldo;
* entradas;
* ajustes;
* perdas;
* movimentações;
* estoque mínimo;
* histórico;
* alertas;
* bloqueio concorrente.

### Critério de conclusão

Toda alteração de saldo possui movimentação rastreável.

## Etapa 5 — Clientes e limite

### Atividades

* cadastro;
* aprovação;
* limite;
* bloqueio;
* consulta de exposição;
* histórico de mudanças;
* validação no checkout.

### Critério de conclusão

Cliente bloqueado ou sem limite não consegue confirmar a compra.

## Etapa 6 — Carrinho e vendas

### Atividades

* carrinho;
* itens;
* revalidação de preço;
* confirmação;
* redução do estoque;
* cancelamento;
* histórico;
* detalhes da venda;
* idempotência.

### Critério de conclusão

Uma venda confirmada atualiza estoque e conta do cliente atomicamente.

## Etapa 7 — Faturamento

### Atividades

* períodos;
* lançamentos;
* calendário;
* quinto dia útil;
* fechamento;
* fatura;
* créditos;
* ajustes;
* inadimplência;
* reprocessamento.

### Critério de conclusão

O worker fecha o mês uma única vez e gera faturas corretas.

## Etapa 8 — Pix

### Atividades

* adaptador do PSP;
* criação de cobrança;
* QR Code;
* webhook;
* idempotência;
* reconciliação;
* expiração;
* pagamento;
* estorno quando suportado.

### Critério de conclusão

Pagamento no sandbox atualiza automaticamente a fatura.

## Etapa 9 — Dashboard e relatórios

### Atividades

* vendas;
* recebimentos;
* contas abertas;
* estoque;
* produtos mais vendidos;
* melhores clientes;
* exportação CSV;
* indicadores no painel React.

### Critério de conclusão

Todos os valores do dashboard podem ser reconciliados com consultas detalhadas.

## Etapa 10 — Previsões

### Atividades

* consolidação mensal;
* média móvel ponderada;
* estoque de segurança;
* sugestão de compra;
* confiança;
* explicação.

### Critério de conclusão

O gerente visualiza a fórmula e os dados utilizados em cada previsão.

## Etapa 11 — Produção

### Atividades

* preparação da VPS;
* firewall;
* domínio;
* TLS;
* Portainer;
* registry;
* stack;
* secrets;
* backups;
* monitoramento;
* restauração;
* teste completo.

### Critério de conclusão

A plataforma opera exclusivamente por HTTPS, com backup e restauração testados.

## Etapa 12 — Preparação efetiva do aplicativo

### Atividades

* validar cliente da API compartilhado;
* implementar login móvel;
* implementar catálogo;
* implementar carrinho;
* implementar conta;
* implementar faturas;
* testar Android e iOS.

A aplicação poderá ser desenvolvida sem mudanças estruturais no backend.

---

# 36. Backlog por épicos

| Épico       | Resultado esperado                       |
| ----------- | ---------------------------------------- |
| Fundação    | Projetos, Docker, banco e CI funcionando |
| Identidade  | Login, sessão, papéis e permissões       |
| Clientes    | Cadastro, aprovação e limite             |
| Catálogo    | Produtos pesquisáveis e administráveis   |
| Estoque     | Saldos e movimentações rastreáveis       |
| Vendas      | Carrinho e confirmação transacional      |
| Faturamento | Períodos e fechamento mensal             |
| Pix         | Cobrança, webhook e liquidação           |
| Relatórios  | Indicadores e exportações                |
| Forecast    | Sugestão de reposição explicável         |
| Operações   | VPS, Portainer, backup e monitoramento   |
| Mobile      | Base React Native pronta para evolução   |

---

# 37. Critérios de aceite essenciais

## Produto

* produto inativo não aparece na loja;
* produto vendido não pode ser apagado fisicamente;
* preço anterior permanece no histórico;
* produto pode ser selecionado rapidamente em uma nova entrada de estoque.

## Estoque

* estoque nunca fica negativo;
* toda alteração gera movimentação;
* ajustes exigem motivo;
* vendas simultâneas são protegidas;
* saldo e movimentações permanecem consistentes.

## Venda

* valor é calculado pelo servidor;
* compra respeita limite;
* compra de cliente bloqueado é rejeitada;
* repetição da requisição não duplica a venda;
* estoque é reduzido na mesma transação.

## Faturamento

* cada cliente recebe no máximo uma fatura por mês;
* compras são incluídas no mês correto;
* fechamento duplicado não duplica valores;
* devoluções geram crédito rastreável;
* falha parcial pode ser reprocessada.

## Pix

* cobrança possui identificador único;
* webhook duplicado não duplica pagamento;
* valor divergente não liquida a fatura automaticamente;
* fatura paga libera o limite;
* cobrança expirada pode ser substituída com segurança.

## Segurança

* gerente não acessa função de administrador;
* cliente não acessa dados de outro cliente;
* senha nunca aparece no banco em texto;
* segredos não aparecem nos logs;
* painel exige autenticação reforçada.

## Frontends React

* loja e painel utilizam o mesmo cliente de API;
* contratos são gerados a partir do OpenAPI;
* componentes compartilhados não contêm regras de negócio;
* falha na aplicação administrativa não impede a loja de ser implantada;
* aplicação móvel pode reutilizar os pacotes independentes de plataforma.

---

# 38. Riscos e mitigação

| Risco                                   | Mitigação                                          |
| --------------------------------------- | -------------------------------------------------- |
| Venda acima do estoque                  | transação e bloqueio de linha                      |
| Webhook duplicado                       | chave única e idempotência                         |
| Cliente inadimplente                    | aprovação, limite e bloqueio                       |
| Fechamento duplicado                    | restrição única por cliente e mês                  |
| Mudança de PSP                          | interface e adaptadores                            |
| Feriado calculado incorretamente        | calendário administrável                           |
| Valor manipulado no frontend            | cálculo integral no backend                        |
| Perda da VPS                            | backup externo e restauração testada               |
| Duplicação entre loja e painel          | pacotes React compartilhados                       |
| Acoplamento excessivo entre frontends   | aplicações separadas por finalidade                |
| Duplicação para mobile                  | pacotes TypeScript compartilhados                  |
| Componente compartilhado muito genérico | compartilhar apenas responsabilidades estáveis     |
| Relatório pesado                        | consultas específicas e índices                    |
| Previsão enganosa                       | método simples e nível de confiança                |
| Vazamento de dados                      | minimização, controle e logs seguros               |
| Falha durante fechamento                | status intermediário e reprocessamento idempotente |
| Bundle administrativo exposto           | domínio e aplicação separados                      |
| Estado inconsistente no frontend        | dados de servidor tratados por camada própria      |

---

# 39. Itens que devem permanecer fora do MVP

Para preservar simplicidade:

* microserviços;
* Kubernetes;
* múltiplos bancos;
* marketplace com vários lojistas;
* inteligência artificial complexa;
* promoções extremamente flexíveis;
* programa de pontos;
* múltiplas moedas;
* logística avançada;
* emissão fiscal automática, salvo se obrigatória;
* Pix Automático;
* aplicativo publicado;
* arquitetura baseada em eventos externos;
* Redis sem necessidade comprovada;
* data warehouse;
* dashboards de BI externos;
* compartilhamento universal de componentes entre web e mobile;
* arquitetura excessivamente genérica.

---

# 40. Stack técnica recomendada

## 40.1. Backend

* Go;
* biblioteca HTTP leve;
* roteador minimalista;
* PostgreSQL;
* driver PostgreSQL para Go;
* migrations SQL;
* OpenAPI;
* logging estruturado;
* validação explícita;
* worker Go;
* testes com PostgreSQL em contêiner.

## 40.2. Loja web

* React;
* TypeScript;
* React Router;
* biblioteca de consulta e cache de dados do servidor;
* schemas de validação;
* formulários tipados;
* testes de componentes;
* testes end-to-end;
* componentes compartilhados por pacote.

## 40.3. Painel administrativo

* React;
* TypeScript;
* React Router;
* biblioteca de consulta e cache de dados do servidor;
* schemas compartilhados;
* formulários tipados;
* biblioteca de componentes administrativos;
* tabelas;
* gráficos;
* testes de componentes;
* testes end-to-end;
* componentes compartilhados com a loja quando apropriado.

## 40.4. Aplicativo futuro

* React Native;
* Expo;
* roteamento compatível com React Native;
* TypeScript;
* cliente OpenAPI compartilhado;
* armazenamento seguro;
* design tokens compartilhados;
* validações compartilhadas.

## 40.5. Infraestrutura

* Docker;
* Docker Compose;
* Portainer;
* reverse proxy com TLS;
* registry de imagens;
* backups externos;
* monitoramento;
* VPS Linux.

---

# 41. Convenções para os projetos React

## 41.1. Componentes

* componentes em `PascalCase`;
* hooks iniciados por `use`;
* arquivos curtos;
* propriedades tipadas;
* ausência de lógica financeira no componente;
* componentes de apresentação separados dos hooks de dados;
* evitar componentes excessivamente genéricos.

## 41.2. Consultas à API

Toda consulta deverá ocorrer por:

* cliente gerado ou tipado;
* função de API específica;
* hook de consulta ou mutação;
* tratamento padronizado de erros.

Exemplo:

```text
features/products/
├── api/
│   └── createProduct.ts
├── hooks/
│   └── useCreateProduct.ts
└── components/
    └── ProductForm.tsx
```

## 41.3. Formulários

Cada formulário deverá possuir:

* schema de validação;
* tipo de entrada;
* mensagens de erro;
* estado de envio;
* prevenção de múltiplos envios;
* tratamento de erro de API;
* confirmação para operações críticas.

## 41.4. Permissões

A interface poderá utilizar componentes como:

```tsx
<Can permission="inventory.adjust">
    <AdjustInventoryButton />
</Can>
```

Entretanto, o backend continuará sendo a autoridade definitiva.

## 41.5. Padrão visual

O pacote de design deverá definir:

* tipografia;
* espaçamento;
* bordas;
* sombras;
* tamanhos;
* estados de foco;
* breakpoints;
* tokens de cor;
* componentes básicos.

Os tokens poderão ser parcialmente compartilhados com o React Native, mas os estilos serão implementados separadamente.

---

# 42. Resultado arquitetural esperado

Ao final da primeira versão, o sistema terá:

* loja React acessível pelo navegador;
* painel React acessível pelo navegador;
* biblioteca React compartilhada entre as aplicações web;
* API Go independente dos frontends;
* PostgreSQL consistente;
* estoque transacional;
* conta mensal por cliente;
* fechamento no quinto dia útil;
* cobrança Pix;
* confirmação por webhook;
* controle de inadimplência;
* relatórios básicos;
* previsão simples de reposição;
* implantação Docker pelo Portainer;
* backups;
* auditoria;
* segurança mínima adequada;
* base React Native integrada ao monorepositório.

A estrutura permitirá evoluir para aplicativo sem alterar as regras centrais.

O React Native consumirá a mesma API e compartilhará pacotes independentes de interface. A loja React e o painel React permanecerão como aplicações separadas, mas compartilharão componentes, hooks, contratos, validações e padrões sempre que isso reduzir manutenção sem criar acoplamento desnecessário.

A decisão de utilizar React em todo o frontend proporcionará:

* menor diversidade tecnológica;
* curva de aprendizagem mais simples;
* reutilização mais ampla;
* padronização de testes;
* padronização de lint;
* maior proximidade com React Native;
* facilidade de movimentação de desenvolvedores entre aplicações;
* manutenção mais previsível;
* redução da duplicação de contratos e componentes.
