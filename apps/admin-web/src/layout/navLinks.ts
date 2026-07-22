export type AdminNavLink = {
  to: string;
  label: string;
  permission?: string | string[];
};

export const adminNavLinks: AdminNavLink[] = [
  { to: '/', label: 'Dashboard', permission: 'reports.dashboard.read' },
  { to: '/clientes', label: 'Clientes', permission: 'customers.read' },
  { to: '/pedidos', label: 'Pedidos', permission: 'orders.read' },
  { to: '/faturamento', label: 'Faturamento', permission: 'billing.read' },
  { to: '/produtos', label: 'Produtos', permission: 'products.read' },
  { to: '/estoque', label: 'Estoque', permission: 'inventory.read' },
  { to: '/relatorios', label: 'Relatórios', permission: [
    'reports.dashboard.read',
    'reports.sales.read',
    'reports.inventory.read',
    'reports.receivables.read',
    'reports.payments.read',
    'reports.customers.read',
    'reports.exceptions.read',
    'reports.forecasting.read',
  ] },
  { to: '/usuarios', label: 'Usuários', permission: 'users.manage' },
  { to: '/auditoria', label: 'Auditoria', permission: 'audit.read' },
];
