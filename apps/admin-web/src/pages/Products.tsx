import { useEffect, useState } from 'react';

type Product = { id: string; name: string; slug: string };

export function ProductsPage() {
  const [items, setItems] = useState<Product[]>([]);
  const [name, setName] = useState('');
  const [error, setError] = useState<string | null>(null);

  async function load() {
    const res = await fetch('/api/v1/admin/products', { credentials: 'include', headers: { 'X-App-Audience': 'admin' } });
    if (!res.ok) throw new Error('Falha ao listar');
    const data = (await res.json()) as { items: Product[] };
    setItems(data.items ?? []);
  }

  useEffect(() => {
    load().catch((e: Error) => setError(e.message));
  }, []);

  async function create(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    try {
      const res = await fetch('/api/v1/admin/products', {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json', 'X-App-Audience': 'admin' },
        body: JSON.stringify({ name, slug: name.toLowerCase().replace(/\s+/g, '-'), sku_code: name.slice(0, 8).toUpperCase(), sale_price_cents: 1000 }),
      });
      if (!res.ok) throw new Error('Falha ao criar');
      setName('');
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro');
    }
  }

  return (
    <section>
      <h1>Produtos</h1>
      {error && <p className="error">{error}</p>}
      <form onSubmit={create} className="form">
        <label>
          Nome
          <input value={name} onChange={(e) => setName(e.target.value)} required />
        </label>
        <button type="submit">Criar produto (MVP)</button>
      </form>
      <ul>
        {items.map((p) => (
          <li key={p.id}>{p.name}</li>
        ))}
      </ul>
    </section>
  );
}
