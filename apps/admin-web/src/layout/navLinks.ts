export type AdminNavLink = {
  to: string;
  label: string;
  permission?: string | string[];
};

export const adminNavLinks: AdminNavLink[] = [
  { to: '/', label: 'Dashboard', permission: 'reports.read' },
  { to: '/clientes', label: 'Clientes', permission: 'customers.read' },
  { to: '/pedidos', label: 'Pedidos', permission: 'orders.read' },
  { to: '/faturamento', label: 'Faturamento', permission: 'billing.read' },
  { to: '/produtos', label: 'Produtos', permission: 'products.read' },
  { to: '/estoque', label: 'Estoque', permission: 'inventory.read' },
  { to: '/relatorios', label: 'Relatórios', permission: 'reports.read' },
  { to: '/usuarios', label: 'Usuários', permission: 'users.manage' },
  { to: '/auditoria', label: 'Auditoria', permission: 'audit.read' },
];
