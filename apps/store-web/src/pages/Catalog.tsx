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
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [msg, setMsg] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    api
      .listProducts()
      .then((res) => setProducts((res.items ?? []) as Product[]))
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false));
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
    <section className="content-section">
      <h1>Catálogo</h1>
      {loading && <p>Carregando produtos…</p>}
      {error && <p className="error">{error}</p>}
      {msg && <p className="ok">{msg}</p>}
      {!loading && !error && products.length === 0 && (
        <p>Nenhum produto disponível. Cadastre itens no painel admin ou rode as migrations (seed de demo).</p>
      )}
      <ul className="product-grid">
        {products.map((p) => {
          const sku = p.skus?.[0];
          return (
            <li key={p.id} className="product-card">
              <strong>{p.name}</strong>
              {sku && (
                <>
                  <div>{formatMoney(sku.sale_price_cents)}</div>
                  <button type="button" onClick={() => add(sku.id)}>Adicionar ao carrinho</button>
                </>
              )}
            </li>
          );
        })}
      </ul>
    </section>
  );
}
