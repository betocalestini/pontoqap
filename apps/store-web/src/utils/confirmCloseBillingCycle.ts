import { formatMoney } from '@store/shared-core';

export function confirmCloseBillingCycle(totalCents: number): boolean {
  const msg =
    `Fechar fatura de ${formatMoney(totalCents)}?\n\n` +
    'Será gerada uma fatura para pagamento (Pix). Um novo ciclo na mesma competência será aberto para suas próximas compras.';
  return window.confirm(msg);
}
