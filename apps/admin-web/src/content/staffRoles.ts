/** Resumo estático dos papéis internos (espelha docs/access-control.md). */
export const STAFF_ROLE_SUMMARIES: { code: string; title: string; bullets: string[] }[] = [
  {
    code: 'system_admin',
    title: 'Administrador do sistema',
    bullets: [
      'Acesso total ao painel, configurações e usuários internos',
      'Único papel que convida ou promove outro administrador',
    ],
  },
  {
    code: 'manager',
    title: 'Gerente',
    bullets: [
      'Catálogo, preços, estoque, clientes, limites, vendas, faturas, pagamentos e relatórios',
      'Sem calendário comercial crítico, usuários internos ou auditoria administrativa',
    ],
  },
  {
    code: 'inventory_operator',
    title: 'Operador de estoque',
    bullets: ['Consultar produtos e estoque', 'Registrar entradas e perdas', 'Sem preços ou financeiro'],
  },
  {
    code: 'finance_operator',
    title: 'Financeiro',
    bullets: [
      'Faturas, fechamento de competência, ajustes, pagamentos e relatórios',
      'Sem catálogo nem movimentação de estoque',
    ],
  },
];
