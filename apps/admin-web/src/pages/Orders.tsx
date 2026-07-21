import { Fragment, useCallback, useEffect, useState } from 'react';
import type { AdminOrderDetail, AdminOrderListItem } from '@store/api-client';
import { formatMoney } from '@store/shared-core';
import { useDialog } from '@store/ui';
import { api } from '../api';
import { useHasPermission } from '../auth/usePermissions';

const ORDER_TABLE_COLS = 6;

export function OrdersPage() {
  const { confirm, prompt } = useDialog();
  const canCancel = useHasPermission('orders.cancel');
  const [items, setItems] = useState<AdminOrderListItem[]>([]);
  const [total, setTotal] = useState(0);
  const [status, setStatus] = useState('');
  const [search, setSearch] = useState('');
  const [selected, setSelected] = useState<AdminOrderDetail | null>(null);
  const [detailLoadingId, setDetailLoadingId] = useState<string | null>(null);
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
    if (selected?.id === id && detailLoadingId === null) {
      setSelected(null);
      return;
    }
    setError(null);
    setDetailLoadingId(id);
    try {
      const order = await api.adminGetOrder(id);
      setSelected(order);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao carregar pedido');
    } finally {
      setDetailLoadingId(null);
    }
  }

  async function cancelOrder(id: string) {
    const ok = await confirm({
      title: 'Cancelar pedido',
      message:
        'Cancelar este pedido? O estoque será devolvido e o valor será estornado na competência aberta ou creditado na fatura, conforme o caso.',
      confirmLabel: 'Continuar',
      variant: 'danger',
    });
    if (!ok) return;
    const password = await prompt({
      title: 'Confirmar identidade',
      message: 'Informe sua senha para cancelar o pedido.',
      label: 'Senha',
      inputType: 'password',
      confirmLabel: 'Cancelar pedido',
    });
    if (!password) return;
    setError(null);
    try {
      const order = await api.adminCancelOrder(id, { password });
      setSelected(order);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha ao cancelar');
    }
  }

  return (
    <section className="content-section">
      <h1>Pedidos</h1>
      <p className="form-hint">
        Consulta de vendas confirmadas na loja (RF-VEN-012). Cancelamento exige confirmação com senha e estorna o valor na
        competência aberta ou credita fatura fechada em aberto.
      </p>
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
                {items.map((o) => {
                  const isExpanded = selected?.id === o.id;
                  const isLoadingRow = detailLoadingId === o.id;
                  return (
                    <Fragment key={o.id}>
                      <tr className={isExpanded ? 'orders-row--expanded' : undefined}>
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
                          <button
                            type="button"
                            className="btn-link"
                            disabled={detailLoadingId !== null && detailLoadingId !== o.id}
                            onClick={() => void openDetail(o.id)}
                          >
                            {isLoadingRow ? 'Abrindo…' : isExpanded ? 'Fechar' : 'Detalhes'}
                          </button>
                        </td>
                      </tr>
                      {(isLoadingRow || isExpanded) && (
                        <tr className="orders-detail-row">
                          <td colSpan={ORDER_TABLE_COLS}>
                            {isLoadingRow ? (
                              <p className="orders-detail-row__loading">Carregando detalhes…</p>
                            ) : selected ? (
                              <div className="orders-detail-panel">
                                <h3 className="orders-detail-panel__title">Itens do pedido</h3>
                                <p className="orders-detail-panel__meta">
                                  Status: <strong>{selected.status}</strong> — Total:{' '}
                                  {formatMoney(selected.total_cents)}
                                </p>
                                <ul className="orders-detail-panel__items">
                                  {selected.items?.map((it) => (
                                    <li key={it.id}>
                                      {it.product_name} ({it.sku_code}) × {it.quantity} —{' '}
                                      {formatMoney(it.total_cents)}
                                    </li>
                                  ))}
                                </ul>
                                <div className="orders-detail-panel__actions">
                                  {canCancel && selected.status === 'confirmed' && (
                                    <button
                                      type="button"
                                      className="button--danger"
                                      onClick={() => void cancelOrder(selected.id)}
                                    >
                                      Cancelar pedido
                                    </button>
                                  )}
                                </div>
                              </div>
                            ) : null}
                          </td>
                        </tr>
                      )}
                    </Fragment>
                  );
                })}
              </tbody>
            </table>
          </div>
        </>
      )}
    </section>
  );
}
