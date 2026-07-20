import { useEffect, useState } from 'react';
import { formatMoney } from '@store/shared-core';
import { api } from '../api';

export function DashboardPage() {
  const [data, setData] = useState<Record<string, unknown> | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.adminDashboard().then((d) => setData(d as Record<string, unknown>)).catch((e: Error) => setError(e.message));
  }, []);

  return (
    <section>
      <h1>Dashboard</h1>
      {error && <p className="error">{error}</p>}
      {data && (
        <ul>
          {'revenue_cents' in data && <li>Receita: {formatMoney(Number(data.revenue_cents))}</li>}
          {'orders_count' in data && <li>Pedidos: {String(data.orders_count)}</li>}
          {'open_invoices_cents' in data && (
            <li>Faturas em aberto: {formatMoney(Number(data.open_invoices_cents))}</li>
          )}
        </ul>
      )}
    </section>
  );
}
