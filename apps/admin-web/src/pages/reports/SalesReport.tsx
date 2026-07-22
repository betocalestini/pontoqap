import { useCallback, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import type { ReportSalesOrderRow } from '@store/api-client';
import { formatMoney } from '@store/shared-core';
import { api } from '../../api';
import { fetchAllPages } from '../../components/reports/exportReport';
import { exportSubtitle } from '../../components/reports/exportSubtitle';
import { ReportPageHeader } from '../../components/reports/ReportPageHeader';
import { ReportSummaryCards } from '../../components/reports/ReportSummaryCards';
import { useReportQuery } from '../../components/reports/useReportQuery';

export function SalesReportPage() {
  const { values, setField, setOffset, queryParams } = useReportQuery({
    date_from: '',
    date_to: '',
    status: '',
  });
  const [items, setItems] = useState<ReportSalesOrderRow[]>([]);
  const [total, setTotal] = useState(0);
  const [summary, setSummary] = useState<Record<string, number>>({});
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setError(null);
    try {
      const res = await api.adminReportSalesOrders(queryParams);
      setItems(res.items ?? []);
      setTotal(res.total ?? 0);
      setSummary((res.summary as Record<string, number>) ?? {});
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Erro ao carregar');
    }
  }, [queryParams]);

  useEffect(() => {
    void load();
  }, [load]);

  const buildExportTable = useCallback(async () => {
    const rows = await fetchAllPages((offset, limit) =>
      api.adminReportSalesOrders({ ...queryParams, offset, limit }),
    );
    return {
      title: 'Vendas e pedidos',
      subtitle: exportSubtitle(),
      filenameBase: 'vendas-pedidos',
      headers: ['Pedido', 'Cliente', 'Confirmado em', 'Total (centavos)', 'Status'],
      rows: rows.map((r) => [
        r.order_number,
        r.customer_name,
        r.confirmed_at ?? '',
        String(r.total_cents),
        r.status,
      ]),
    };
  }, [queryParams]);

  const offset = Number(queryParams.offset ?? 0);
  const limit = Number(queryParams.limit ?? 50);

  return (
    <section className="content-section">
      <ReportPageHeader
        title="Vendas e pedidos"
        description="Pedidos confirmados no período (competência)."
        exportTable={buildExportTable}
      >
        <label>
          De
          <input type="date" value={values.date_from ?? ''} onChange={(e) => setField('date_from', e.target.value)} />
        </label>
        <label>
          Até
          <input type="date" value={values.date_to ?? ''} onChange={(e) => setField('date_to', e.target.value)} />
        </label>
        <label>
          Status
          <select value={values.status ?? ''} onChange={(e) => setField('status', e.target.value)}>
            <option value="">Confirmados (período)</option>
            <option value="confirmed">Confirmado</option>
            <option value="cancelled">Cancelado</option>
          </select>
        </label>
      </ReportPageHeader>
      {error && <p className="error">{error}</p>}
      <ReportSummaryCards
        cards={[
          { label: 'Total vendido', value: formatMoney(summary.total_sales_cents ?? 0) },
          { label: 'Pedidos', value: String(summary.order_count ?? 0) },
          { label: 'Ticket médio', value: formatMoney(summary.average_ticket_cents ?? 0) },
        ]}
      />
      <div className="table-scroll">
        <table>
          <thead>
            <tr>
              <th>Pedido</th>
              <th>Cliente</th>
              <th>Data</th>
              <th>Total</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {items.map((row) => (
              <tr key={row.id}>
                <td>
                  <Link to="/pedidos">{row.order_number}</Link>
                </td>
                <td>{row.customer_name}</td>
                <td>{row.confirmed_at ? new Date(row.confirmed_at).toLocaleString('pt-BR') : '—'}</td>
                <td>{formatMoney(row.total_cents)}</td>
                <td>{row.status}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="pagination-row">
        <button type="button" disabled={offset <= 0} onClick={() => setOffset(Math.max(0, offset - limit))}>
          Anterior
        </button>
        <span>
          {offset + 1}–{Math.min(offset + limit, total)} de {total}
        </span>
        <button type="button" disabled={offset + limit >= total} onClick={() => setOffset(offset + limit)}>
          Próxima
        </button>
      </div>
    </section>
  );
}
