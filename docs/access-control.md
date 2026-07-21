# Controle de acesso (MVP)

Documento canônico de **papéis**, **permissões** e **fluxos** do painel administrativo e da loja. O backend é a fonte da verdade (`roles`, `role_permissions`, middleware `RequirePermission`).

## Modelo de usuário

- Uma pessoa = um registro em `users`.
- Clientes da loja têm papel `customer` e linha em `customers`.
- Funcionários internos mantêm o papel `customer` e recebem **um** papel interno adicional (`system_admin`, `manager`, `inventory_operator` ou `finance_operator`). Não há usuário interno sem cadastro na loja no MVP.
- Papéis e permissões são **fixos** (definidos em migrations). Não há UI para montar permissões personalizadas.

## Mapa de nomes (sugestão → código no banco)

| Nome sugerido | Código (`roles.code`) | Nome exibido |
| ------------- | --------------------- | ------------ |
| ADMIN | `system_admin` | Administrador do sistema |
| MANAGER | `manager` | Gerente |
| INVENTORY_OPERATOR | `inventory_operator` | Operador de estoque |
| FINANCE_OPERATOR | `finance_operator` | Financeiro |
| (cliente loja) | `customer` | Cliente |

## Papéis internos — resumo

### `system_admin` (Administrador)

Acesso total ao painel: catálogo, estoque, clientes, faturamento, pagamentos, relatórios, pedidos, **configurações** (`settings.write`), **usuários** (`users.manage`), **auditoria** (`audit.read`).

Regras: número reduzido de administradores; somente administrador pode convidar ou promover outro `system_admin`; não é possível desativar o último administrador ativo.

### `manager` (Gerente)

Operação comercial: produtos, preços, estoque (incl. ajustes), clientes, limites, pedidos, faturamento (consulta e fechamento de competência), pagamentos, relatórios e previsões.

**Não possui:** `settings.write` (calendário comercial, parâmetros críticos, integrações via config), `users.manage`, `audit.read` (após migration `000006`).

### `inventory_operator` (Operador de estoque)

Consultar produtos e estoque; registrar entradas (`inventory.entry`) e perdas (`inventory.loss`). Sem alteração de preços, limites ou dados financeiros.

### `finance_operator` (Financeiro)

Pedidos (leitura), faturamento, fechamento de competência e ajustes de fatura (`billing.close`), pagamentos e relatórios. Sem catálogo, preços ou movimentação de estoque.

## Matriz papel × permissão

Legenda: ● = concedida; — = não concedida.

| Permissão (`code`) | Admin | Gerente | Estoque | Financeiro |
| ------------------ | :---: | :-----: | :-----: | :--------: |
| `products.read` | ● | ● | ● | — |
| `products.write` | ● | ● | — | — |
| `inventory.read` | ● | ● | ● | — |
| `inventory.adjust` | ● | ● | — | — |
| `inventory.entry` | ● | ● | ● | — |
| `inventory.loss` | ● | ● | ● | — |
| `customers.read` | ● | ● | — | — |
| `customers.write` | ● | ● | — | — |
| `customers.approve` | ● | ● | — | — |
| `customers.change_limit` | ● | ● | — | — |
| `orders.read` | ● | ● | — | ● |
| `orders.cancel` | ● | ● | — | — |
| `billing.read` | ● | ● | — | ● |
| `billing.close` | ● | ● | — | ● |
| `payments.read` | ● | ● | — | ● |
| `reports.read` | ● | ● | — | ● |
| `settings.write` | ● | — | — | — |
| `audit.read` | ● | — | — | — |
| `users.manage` | ● | — | — | — |

Alterações futuras na matriz: nova migration em `backend/migrations/`, nunca edição em runtime.

## Estados da conta (`users.status`)

| Conceito | Valor no banco | Painel admin | Loja (cliente) |
| -------- | -------------- | ------------ | -------------- |
| Ativo | `active` | Permitido (com MFA se exigido) | Permitido se cliente não bloqueado |
| Convidado (staff) | `invited` | Após aceite do convite | Conforme cadastro cliente |
| E-mail pendente (loja) | `pending_email` | — | Até confirmar e-mail |
| Bloqueio temporário (login) | `temporarily_blocked` | Negado | Negado |
| Suspenso (staff) | `suspended` | Negado | **Permitido** (compras continuam) |
| Desativado | `disabled` | Negado | Negado |

Bloqueio de **cliente** (`customers.status = blocked`) é independente de `suspended` no usuário.

Funcionário que sai da empresa: `disabled` + revogação de sessões admin; **não** excluir o usuário (preserva auditoria e histórico).

## Fluxos

### Bootstrap

Na subida da API, se não existir `system_admin`, cria-se um usuário via `ADMIN_BOOTSTRAP_*` (ver `README.md`).

### Convite de funcionário

1. Pessoa cadastrada na loja (cliente aprovado ou equivalente).
2. Administrador (ou outro papel com `users.manage` — hoje só admin): `POST /api/v1/admin/users/invitations` ou UI **Usuários**.
3. E-mail com link (token em hash, validade **48h** em `admin_invitations`).
4. Aceite: `POST /api/v1/auth/accept-invitation` (define senha se necessário, atribui papel interno, mantém `customer`).
5. Revogação: `POST /api/v1/admin/users/invitations/{id}/revoke`.

### Atribuição direta

`POST /api/v1/admin/customers/{id}/staff-role` (requer `users.manage` + confirmação de senha ou MFA). Também disponível em **Clientes → Editar → Acesso administrativo**.

### Suspender vs desativar

- **Suspender:** só painel admin; loja inalterada.
- **Desativar:** impede novos acessos; revoga sessões admin.

Operações sensíveis (status, papel, revogar sessões): step-up com senha ou código MFA.

## Segurança (implementado)

- MFA obrigatório no painel quando `ADMIN_MFA_REQUIRED=true` (produção por padrão).
- Bloqueio progressivo por tentativas de login (`temporarily_blocked`).
- Sessão admin via JWT (`SESSION_TTL_ADMIN`); revogação em `sessions`.
- Usuário não pode alterar o próprio status administrativo nem o próprio papel interno.
- Alterações de papel, convites e status geram entradas em `audit_logs`.

## Fora do escopo MVP

- Permissões customizáveis por usuário.
- Convite sem cadastro prévio na loja.
- Renomear códigos de papel (`ADMIN` → `system_admin` permanece no banco).

## Endpoints administrativos (resumo)

| Função | Permissão | Rota |
| ------ | --------- | ---- |
| Listar / detalhar pedidos | `orders.read` | `GET /admin/orders`, `GET /admin/orders/{id}` |
| Cancelar pedido | `orders.cancel` | `POST /admin/orders/{id}/cancel` |
| Consultar auditoria | `audit.read` | `GET /admin/audit/logs` |
| Relatório de estoque | `inventory.read` | `GET /admin/reports/inventory` |

## Referências

- Rotas e permissões: `backend/internal/app/router.go`
- Regras de convite e papéis: `backend/internal/identity/admin_users.go`
- Seeds: `backend/migrations/000001_initial.up.sql`, `000003_*`, `000004_*`, `000005_*`, `000006_*`
