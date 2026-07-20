import { useCallback, useEffect, useState } from 'react';
import { formatMoney } from '@store/shared-core';
import { api } from '../api';

type CartItem = {
  id: string;
  sku_id: string;
  quantity: number;
  product_name: string;
  sku_code?: string;
  unit_price_cents: number;
  line_total_cents: number;
};

type Cart = { items?: CartItem[] };

export function CartPage() {
  const [cart, setCart] = useState<Cart | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [orderId, setOrderId] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const load = useCallback(() => {
    api.getCart().then((c) => setCart(c as Cart)).catch((e: Error) => setError(e.message));
  }, []);

  useEffect(() => {
    load();
  }, [load, orderId]);

  async function checkout() {
    setError(null);
    setBusy(true);
    try {
      const res = (await api.checkout()) as { id?: string; order_number?: string };
      setOrderId(res.order_number ?? res.id ?? 'ok');
      setCart({ items: [] });
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Erro');
    } finally {
      setBusy(false);
    }
  }

  async function changeQty(skuId: string, quantity: number) {
    setError(null);
    setBusy(true);
    try {
      const c = (await api.setCartItemQuantity(skuId, quantity)) as Cart;
      setCart(c);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Erro');
    } finally {
      setBusy(false);
    }
  }

  const items = cart?.items ?? [];
  const totalCents = items.reduce((sum, it) => sum + it.line_total_cents, 0);

  return (
    <section>
      <h1>Carrinho</h1>
      {error && <p className="error">{error}</p>}
      {orderId && <p className="ok">Pedido confirmado: {orderId}</p>}
      {items.length === 0 && !orderId && <p>Seu carrinho está vazio.</p>}
      <ul className="cart-lines">
        {items.map((it) => (
          <li key={it.id}>
            <div className="cart-line-main">
              <strong>{it.product_name}</strong>
              <span className="cart-line-meta">
                {formatMoney(it.unit_price_cents)} cada
                {it.sku_code ? ` · ${it.sku_code}` : ''}
              </span>
            </div>
            <div className="cart-line-qty">
              <button
                type="button"
                aria-label="Diminuir"
                disabled={busy}
                onClick={() => changeQty(it.sku_id, it.quantity - 1)}
              >
                −
              </button>
              <span>{it.quantity}</span>
              <button
                type="button"
                aria-label="Aumentar"
                disabled={busy}
                onClick={() => changeQty(it.sku_id, it.quantity + 1)}
              >
                +
              </button>
            </div>
            <div className="cart-line-total">{formatMoney(it.line_total_cents)}</div>
          </li>
        ))}
      </ul>
      {items.length > 0 && (
        <p className="cart-total">
          <strong>Total: {formatMoney(totalCents)}</strong>
        </p>
      )}
      <button type="button" onClick={checkout} disabled={!items.length || busy || !!orderId}>
        Finalizar compra
      </button>
    </section>
  );
}
