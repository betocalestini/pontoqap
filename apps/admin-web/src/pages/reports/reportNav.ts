import type { AdminNavLink } from '../../layout/navLinks';

export const reportNavLinks: (AdminNavLink & { path: string })[] = [
  { path: 'vendas', to: '/relatorios/vendas', label: 'Vendas e pedidos', permission: 'reports.sales.read' },
  { path: 'estoque', to: '/relatorios/estoque', label: 'Posição do estoque', permission: 'reports.inventory.read' },
  { path: 'movimentacoes', to: '/relatorios/movimentacoes', label: 'Movimentações', permission: 'reports.inventory.read' },
  { path: 'recebiveis', to: '/relatorios/recebiveis', label: 'Contas a receber', permission: 'reports.receivables.read' },
  { path: 'pix', to: '/relatorios/pix', label: 'Conciliação Pix', permission: 'reports.payments.read' },
  { path: 'limites', to: '/relatorios/limites', label: 'Limites e exposição', permission: 'reports.customers.read' },
  { path: 'excecoes', to: '/relatorios/excecoes', label: 'Exceções', permission: 'reports.exceptions.read' },
  { path: 'previsao', to: '/relatorios/previsao', label: 'Previsão de reposição', permission: 'reports.forecasting.read' },
  { path: 'auditoria', to: '/relatorios/auditoria', label: 'Auditoria administrativa', permission: 'audit.read' },
];
