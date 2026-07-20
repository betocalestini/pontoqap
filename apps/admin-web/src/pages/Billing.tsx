import { useEffect, useState } from 'react';
import { formatMoney } from '@store/shared-core';
import { api } from '../api';

type Invoice = {
  id: string;
  invoice_number: string;
  status: string;
  total_cents: number;
  paid_cents: number;
};

export function BillingPage() {
  const [items, setItems] = useState<Invoice[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [closeMsg, setCloseMsg] = useState<string | null>(null);

  async function load() {
    const res = await api.adminListInvoices();
    setItems((res.items ?? []) as Invoice[]);
  }

  useEffect(() => {
    load().catch((e: Error) => setError(e.message));
  }, []);

  async function closePeriods() {
    setError(null);
    try {
      const res = await api.adminCloseBilling();
      setCloseMsg(JSON.stringify(res));
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Erro');
    }
  }

  return (
    <section className="content-section">
      <h1>Faturamento</h1>
      <button type="button" onClick={closePeriods}>Fechar competência</button>
      {closeMsg && <pre className="code-block">{closeMsg}</pre>}
      {error && <p className="error">{error}</p>}
      <ul className="data-list">
        {items.map((inv) => (
          <li key={inv.id}>
            {inv.invoice_number} — {inv.status} — {formatMoney(inv.total_cents)}
          </li>
        ))}
      </ul>
    </section>
  );
}
