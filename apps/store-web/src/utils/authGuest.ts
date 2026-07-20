import { ApiError } from '@store/api-client';

export function isGuestAuthError(err: unknown): boolean {
  return err instanceof ApiError && (err.code === 'UNAUTHORIZED' || err.code === 'FORBIDDEN');
}

export type GuestAuthContext = 'cart' | 'invoices' | 'invoice' | 'catalog';

export function guestAuthMessage(context: GuestAuthContext): string {
  switch (context) {
    case 'cart':
      return 'Entre na sua conta para usar o carrinho e finalizar compras.';
    case 'invoices':
      return 'Entre na sua conta para ver suas faturas.';
    case 'invoice':
      return 'Entre na sua conta para acessar esta fatura.';
    case 'catalog':
      return 'Entre na sua conta para adicionar itens ao carrinho.';
    default:
      return 'Entre na sua conta para continuar.';
  }
}
