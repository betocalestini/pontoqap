import { useEffect, useState } from 'react';
import { createApiClient } from '@store/api-client';
import { formatMoney } from '@store/shared-core';
import './App.css';

const api = createApiClient('/api/v1');

type Product = {
  id: string;
  name: string;
  skus?: { id: string; sale_price_cents: number; available_quantity?: number }[];
};

export default function App() {
  const [products, setProducts] = useState<Product[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api
      .listProducts()
      .then((res) => setProducts(res.items as Product[]))
      .catch((e: Error) => setError(e.message));
  }, []);

  return (
    <div className="page">
      <header>
        <h1>Store Platform</h1>
        <p>Catálogo público (MVP)</p>
      </header>
      {error && <p className="error">{error}</p>}
      <ul className="grid">
        {products.map((p) => {
          const sku = p.skus?.[0];
          return (
            <li key={p.id}>
              <strong>{p.name}</strong>
              {sku && (
                <span>
                  {' '}
                  — {formatMoney(sku.sale_price_cents)}
                  {sku.available_quantity != null && ` (${sku.available_quantity} em estoque)`}
                </span>
              )}
            </li>
          );
        })}
      </ul>
    </div>
  );
}
