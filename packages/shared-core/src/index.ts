export {
  THEME_STORAGE_KEY,
  applyDocumentTheme,
  getInitialTheme,
  readThemePreference,
  resolveEffectiveTheme,
  writeThemePreference,
  type ColorTheme,
} from './theme';

/** Códigos de permissão do painel admin (fonte da verdade: backend / migrations). */
export const permissionCodes = {
  productsRead: 'products.read',
  productsWrite: 'products.write',
  inventoryRead: 'inventory.read',
  inventoryAdjust: 'inventory.adjust',
  inventoryEntry: 'inventory.entry',
  inventoryLoss: 'inventory.loss',
  customersRead: 'customers.read',
  customersWrite: 'customers.write',
  customersApprove: 'customers.approve',
  customersChangeLimit: 'customers.change_limit',
  ordersRead: 'orders.read',
  ordersCancel: 'orders.cancel',
  billingRead: 'billing.read',
  billingClose: 'billing.close',
  paymentsRead: 'payments.read',
  reportsDashboardRead: 'reports.dashboard.read',
  reportsSalesRead: 'reports.sales.read',
  reportsInventoryRead: 'reports.inventory.read',
  reportsReceivablesRead: 'reports.receivables.read',
  reportsPaymentsRead: 'reports.payments.read',
  reportsCustomersRead: 'reports.customers.read',
  reportsExceptionsRead: 'reports.exceptions.read',
  reportsForecastingRead: 'reports.forecasting.read',
  reportsRead: 'reports.read',
  settingsWrite: 'settings.write',
  auditRead: 'audit.read',
  usersManage: 'users.manage',
} as const;

export function formatMoney(cents: number): string {
  return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(cents / 100);
}
