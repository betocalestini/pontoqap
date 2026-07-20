import { useEffect, useState } from 'react';
import { api } from '../api';

export function ReportsPage() {
  const [top, setTop] = useState<unknown[]>([]);
  const [inventory, setInventory] = useState<unknown[]>([]);
  const [forecast, setForecast] = useState<unknown[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([api.adminTopProducts(), api.adminInventoryReport(), api.adminForecast()])
      .then(([t, inv, f]) => {
        setTop((t as { items?: unknown[] }).items ?? []);
        setInventory((inv as { items?: unknown[] }).items ?? []);
        setForecast((f as { items?: unknown[] }).items ?? []);
      })
      .catch((e: Error) => setError(e.message));
  }, []);

  async function generate() {
    await api.adminGenerateForecast();
    const f = await api.adminForecast();
    setForecast((f as { items?: unknown[] }).items ?? []);
  }

  return (
    <section className="content-section">
      <h1>Relatórios</h1>
      {error && <p className="error">{error}</p>}
      <h2>Top produtos</h2>
      <pre className="code-block">{JSON.stringify(top, null, 2)}</pre>
      <h2>Estoque</h2>
      <pre className="code-block">{JSON.stringify(inventory, null, 2)}</pre>
      <h2>Previsão</h2>
      <button type="button" onClick={generate}>Gerar previsão</button>
      <pre className="code-block">{JSON.stringify(forecast, null, 2)}</pre>
    </section>
  );
}
