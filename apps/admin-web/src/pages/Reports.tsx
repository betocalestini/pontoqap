import { useEffect, useState } from 'react';
import { api } from '../api';
import { useHasPermission } from '../auth/usePermissions';

export function ReportsPage() {
  const canInventoryReport = useHasPermission('inventory.read');
  const [top, setTop] = useState<unknown[]>([]);
  const [inventory, setInventory] = useState<unknown[]>([]);
  const [forecast, setForecast] = useState<unknown[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const tasks: Promise<void>[] = [
      api.adminTopProducts().then((t) => {
        setTop((t as { items?: unknown[] }).items ?? []);
      }),
      api.adminForecast().then((f) => {
        setForecast((f as { items?: unknown[] }).items ?? []);
      }),
    ];
    if (canInventoryReport) {
      tasks.push(
        api.adminInventoryReport().then((inv) => {
          setInventory((inv as { items?: unknown[] }).items ?? []);
        }),
      );
    } else {
      setInventory([]);
    }
    Promise.all(tasks).catch((e: Error) => setError(e.message));
  }, [canInventoryReport]);

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
      {canInventoryReport ? (
        <>
          <h2>Estoque</h2>
          <pre className="code-block">{JSON.stringify(inventory, null, 2)}</pre>
        </>
      ) : (
        <p className="form-hint">Relatório de estoque disponível apenas para papéis com acesso operacional ao estoque.</p>
      )}
      <h2>Previsão</h2>
      <button type="button" onClick={generate}>
        Gerar previsão
      </button>
      <pre className="code-block">{JSON.stringify(forecast, null, 2)}</pre>
    </section>
  );
}
