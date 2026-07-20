import { useEffect, useState } from 'react';
import { api } from '../api';

type Customer = { id: string; status: string; user?: { email: string; name: string } };

export function CustomersPage() {
  const [items, setItems] = useState<Customer[]>([]);
  const [error, setError] = useState<string | null>(null);

  async function load() {
    const res = await api.adminListCustomers();
    setItems((res.items ?? []) as Customer[]);
  }

  useEffect(() => {
    load().catch((e: Error) => setError(e.message));
  }, []);

  async function approve(id: string) {
    setError(null);
    try {
      await api.adminApproveCustomer(id);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Erro');
    }
  }

  return (
    <section>
      <h1>Clientes</h1>
      {error && <p className="error">{error}</p>}
      <ul>
        {items.map((c) => (
          <li key={c.id}>
            {c.user?.email ?? c.id} — {c.status}
            {c.status === 'pending' && (
              <button type="button" onClick={() => approve(c.id)}>Aprovar</button>
            )}
          </li>
        ))}
      </ul>
    </section>
  );
}
