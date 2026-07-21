import { useCallback, useEffect, useState } from 'react';
import { formatMoney } from '@store/shared-core';
import { api } from '../api';
import { AuthGuestPrompt } from '../components/AuthGuestPrompt';
import { guestAuthMessage, isGuestAuthError } from '../utils/authGuest';

type CartItem = {
  id: string;
  sku_id: string;
  quantity: number;
  product_name: string;
  unit_price_cents: number;
  line_total_cents: number;
};

type Cart = { items?: CartItem[] };

export function CartPage() {
  const [cart, setCart] = useState<Cart | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [needsAuth, setNeedsAuth] = useState(false);
  const [orderId, setOrderId] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const load = useCallback(() => {
    setNeedsAuth(false);
    api
      .getCart()
      .then((c) => setCart(c as Cart))
      .catch((e: Error) => {
        if (isGuestAuthError(e)) {
          setNeedsAuth(true);
          setCart(null);
          setError(null);
        } else {
          setError(e.message);
        }
      });
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
      if (isGuestAuthError(e)) {
        setNeedsAuth(true);
        setError(null);
      } else {
        setError(e instanceof Error ? e.message : 'Erro');
      }
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
      if (isGuestAuthError(e)) {
        setNeedsAuth(true);
        setError(null);
      } else {
        setError(e instanceof Error ? e.message : 'Erro');
      }
    } finally {
      setBusy(false);
    }
  }

  async function removeItem(skuId: string) {
    await changeQty(skuId, 0);
  }

  const items = cart?.items ?? [];
  const totalCents = items.reduce((sum, it) => sum + it.line_total_cents, 0);

  return (
    <section className="content-section">
      <h1>Carrinho</h1>
      {needsAuth && <AuthGuestPrompt message={guestAuthMessage('cart')} />}
      {error && <p className="error">{error}</p>}
      {orderId && <p className="ok">Pedido confirmado: {orderId}</p>}
      {!needsAuth && items.length === 0 && !orderId && <p>Seu carrinho está vazio.</p>}
      {!needsAuth && (
        <ul className="cart-lines">
          {items.map((it) => (
            <li key={it.id}>
              <div className="cart-line-main">
                <strong>{it.product_name}</strong>
                <span className="cart-line-meta">{formatMoney(it.unit_price_cents)} cada</span>
              </div>
              <div className="cart-line-row">
                <div className="cart-line-qty">
                  <button
                    type="button"
                    aria-label="Diminuir"
                    disabled={busy || it.quantity <= 1}
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
                <button
                  type="button"
                  className="cart-line-remove"
                  aria-label={`Remover ${it.product_name}`}
                  disabled={busy}
                  onClick={() => void removeItem(it.sku_id)}
                >
                  Remover
                </button>
                <div className="cart-line-total">{formatMoney(it.line_total_cents)}</div>
              </div>
            </li>
          ))}
        </ul>
      )}
      {!needsAuth && items.length > 0 && (
        <div className="cart-actions">
          <p className="cart-total">
            <strong>Total: {formatMoney(totalCents)}</strong>
          </p>
          <button
            type="button"
            className="btn-primary"
            onClick={checkout}
            disabled={!items.length || busy || !!orderId}
          >
            Finalizar compra
          </button>
        </div>
      )}
    </section>
  );
}
