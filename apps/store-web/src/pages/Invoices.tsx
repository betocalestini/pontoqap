import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { formatMoney } from '@store/shared-core';
import { api } from '../api';

type Invoice = {
  id: string;
  invoice_number: string;
  status: string;
  total_cents: number;
  paid_cents: number;
};

export function InvoicesPage() {
  const [items, setItems] = useState<Invoice[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.listMyInvoices()
      .then((res) => setItems((res.items ?? []) as Invoice[]))
      .catch((e: Error) => setError(e.message));
  }, []);

  return (
    <section>
      <h1>Minhas faturas</h1>
      {error && <p className="error">{error}</p>}
      <ul>
        {items.map((inv) => (
          <li key={inv.id}>
            <Link to={`/faturas/${inv.id}`}>{inv.invoice_number}</Link> — {inv.status} —{' '}
            {formatMoney(inv.total_cents - inv.paid_cents)} em aberto
          </li>
        ))}
      </ul>
    </section>
  );
}
