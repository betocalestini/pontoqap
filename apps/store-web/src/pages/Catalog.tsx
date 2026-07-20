import { useEffect, useState } from 'react';
import { formatMoney } from '@store/shared-core';
import { api } from '../api';

type Product = {
  id: string;
  name: string;
  skus?: { id: string; sale_price_cents: number; available_quantity?: number }[];
};

export function CatalogPage() {
  const [products, setProducts] = useState<Product[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [msg, setMsg] = useState<string | null>(null);

  useEffect(() => {
    api.listProducts().then((res) => setProducts(res.items as Product[])).catch((e: Error) => setError(e.message));
  }, []);

  async function add(skuId: string) {
    setMsg(null);
    try {
      await api.addToCart(skuId, 1);
      setMsg('Item adicionado ao carrinho.');
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Erro');
    }
  }

  return (
    <section>
      <h1>Catálogo</h1>
      {error && <p className="error">{error}</p>}
      {msg && <p className="ok">{msg}</p>}
      <ul className="grid">
        {products.map((p) => {
          const sku = p.skus?.[0];
          return (
            <li key={p.id}>
              <strong>{p.name}</strong>
              {sku && (
                <>
                  <div>{formatMoney(sku.sale_price_cents)}</div>
                  <button type="button" onClick={() => add(sku.id)}>Comprar</button>
                </>
              )}
            </li>
          );
        })}
      </ul>
    </section>
  );
}
