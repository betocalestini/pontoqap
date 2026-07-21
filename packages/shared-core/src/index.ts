export {
  THEME_STORAGE_KEY,
  applyDocumentTheme,
  getInitialTheme,
  readThemePreference,
  resolveEffectiveTheme,
  writeThemePreference,
  type ColorTheme,
} from './theme';

export const permissionCodes = {
  productsRead: 'products.read',
  productsWrite: 'products.write',
  inventoryRead: 'inventory.read',
  inventoryAdjust: 'inventory.adjust',
  customersRead: 'customers.read',
  customersApprove: 'customers.approve',
  customersChangeLimit: 'customers.change_limit',
} as const;

export function formatMoney(cents: number): string {
  return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(cents / 100);
}
