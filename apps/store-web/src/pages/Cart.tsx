import { useEffect, useState } from 'react';
import { formatMoney } from '@store/shared-core';
import { api } from '../api';

type Cart = { items?: { sku_id: string; quantity: number; line_total_cents?: number }[] };

export function CartPage() {
  const [cart, setCart] = useState<Cart | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [orderId, setOrderId] = useState<string | null>(null);

  useEffect(() => {
    api.getCart().then((c) => setCart(c as Cart)).catch((e: Error) => setError(e.message));
  }, [orderId]);

  async function checkout() {
    setError(null);
    try {
      const res = (await api.checkout()) as { order_id?: string };
      setOrderId(res.order_id ?? 'ok');
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Erro');
    }
  }

  return (
    <section>
      <h1>Carrinho</h1>
      {error && <p className="error">{error}</p>}
      {orderId && <p className="ok">Pedido confirmado: {orderId}</p>}
      <ul>
        {cart?.items?.map((it) => (
          <li key={it.sku_id}>
            SKU {it.sku_id.slice(0, 8)} — qtd {it.quantity}
            {it.line_total_cents != null && ` — ${formatMoney(it.line_total_cents)}`}
          </li>
        ))}
      </ul>
      <button type="button" onClick={checkout} disabled={!cart?.items?.length}>
        Finalizar compra
      </button>
    </section>
  );
}
