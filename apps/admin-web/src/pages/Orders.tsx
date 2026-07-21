import { useCallback, useEffect, useState } from 'react';
import type { AdminOrderDetail, AdminOrderListItem } from '@store/api-client';
import { formatMoney } from '@store/shared-core';
import { api } from '../api';
import { useHasPermission } from '../auth/usePermissions';

export function OrdersPage() {
  const canCancel = useHasPermission('orders.cancel');
  const [items, setItems] = useState<AdminOrderListItem[]>([]);
  const [total, setTotal] = useState(0);
  const [status, setStatus] = useState('');
  const [search, setSearch] = useState('');
  const [selected, setSelected] = useState<AdminOrderDetail | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await api.adminListOrders({
        status: status || undefined,
        search: search || undefined,
        limit: 50,
      });
      setItems(res.items ?? []);
      setTotal(res.total ?? 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao carregar pedidos');
    } finally {
      setLoading(false);
    }
  }, [status, search]);

  useEffect(() => {
    void load();
  }, [load]);

  async function openDetail(id: string) {
    setError(null);
    try {
      setSelected(await api.adminGetOrder(id));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao carregar pedido');
    }
  }

  async function cancelOrder(id: string) {
    if (!window.confirm('Cancelar este pedido? Estoque e faturamento em aberto serão estornados.')) return;
    setError(null);
    try {
      const order = await api.adminCancelOrder(id);
      setSelected(order);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao cancelar');
    }
  }

  return (
    <section className="content-section">
      <h1>Pedidos</h1>
      <p className="form-hint">Consulta de vendas confirmadas na loja (RF-VEN-012). Cancelamento exige competência em aberto.</p>
      {error && <p className="error">{error}</p>}
      <div className="form form--wide customers-list-filters">
        <label>
          Status
          <select value={status} onChange={(e) => setStatus(e.target.value)}>
            <option value="">Todos</option>
            <option value="confirmed">Confirmados</option>
            <option value="cancelled">Cancelados</option>
          </select>
        </label>
        <label>
          Buscar
          <input
            type="search"
            placeholder="Número, cliente, e-mail…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </label>
        <button type="button" onClick={() => void load()}>
          Filtrar
        </button>
      </div>
      {loading ? (
        <p>Carregando…</p>
      ) : (
        <>
          <p className="form-hint">{total} pedido(s)</p>
          <div className="table-scroll">
            <table>
              <thead>
                <tr>
                  <th>Número</th>
                  <th>Cliente</th>
                  <th>Status</th>
                  <th>Total</th>
                  <th>Data</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {items.map((o) => (
                  <tr key={o.id}>
                    <td>{o.order_number}</td>
                    <td>
                      {o.customer_name}
                      <br />
                      <span className="customer-email-hint">{o.customer_email}</span>
                    </td>
                    <td>{o.status}</td>
                    <td>{formatMoney(o.total_cents)}</td>
                    <td>{new Date(o.created_at).toLocaleString('pt-BR')}</td>
                    <td>
                      <button type="button" className="btn-link" onClick={() => void openDetail(o.id)}>
                        Detalhes
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}
      {selected && (
        <div className="customers-panel">
          <h2>Pedido {selected.order_number}</h2>
          <p>
            Status: <strong>{selected.status}</strong> — Total: {formatMoney(selected.total_cents)}
          </p>
          <ul>
            {selected.items?.map((it) => (
              <li key={it.id}>
                {it.product_name} ({it.sku_code}) × {it.quantity} — {formatMoney(it.total_cents)}
              </li>
            ))}
          </ul>
          {canCancel && selected.status === 'confirmed' && (
            <button type="button" className="button--danger" onClick={() => void cancelOrder(selected.id)}>
              Cancelar pedido
            </button>
          )}
          <button type="button" onClick={() => setSelected(null)}>
            Fechar
          </button>
        </div>
      )}
    </section>
  );
}
