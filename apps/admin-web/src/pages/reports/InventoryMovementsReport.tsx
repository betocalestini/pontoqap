import { useCallback, useEffect, useState } from 'react';
import type { ReportInventoryMovementRow } from '@store/api-client';
import { api } from '../../api';
import { fetchAllPages } from '../../components/reports/exportReport';
import { exportSubtitle } from '../../components/reports/exportSubtitle';
import { ReportPageHeader } from '../../components/reports/ReportPageHeader';
import { useReportQuery } from '../../components/reports/useReportQuery';

function defaultMonthRange() {
  const now = new Date();
  const from = new Date(now.getFullYear(), now.getMonth(), 1);
  const to = new Date(now.getFullYear(), now.getMonth() + 1, 0);
  return {
    date_from: from.toISOString().slice(0, 10),
    date_to: to.toISOString().slice(0, 10),
  };
}

export function InventoryMovementsReportPage() {
  const { values, setField, setOffset, queryParams } = useReportQuery(defaultMonthRange());
  const [items, setItems] = useState<ReportInventoryMovementRow[]>([]);
  const [total, setTotal] = useState(0);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      const res = await api.adminReportInventoryMovements(queryParams);
      setItems(res.items ?? []);
      setTotal(res.total ?? 0);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Erro');
    }
  }, [queryParams]);

  useEffect(() => {
    void load();
  }, [load]);

  const buildExportTable = useCallback(async () => {
    const rows = await fetchAllPages((offset, limit) =>
      api.adminReportInventoryMovements({ ...queryParams, offset, limit }),
    );
    return {
      title: 'Movimentações de estoque',
      subtitle: exportSubtitle(),
      filenameBase: 'estoque-movimentacoes',
      headers: ['Quando', 'Produto', 'SKU', 'Tipo', 'Qtd', 'Saldo anterior', 'Saldo posterior', 'Responsável'],
      rows: rows.map((r) => [
        r.created_at,
        r.product_name,
        r.sku_code,
        r.movement_type,
        String(r.quantity),
        String(r.previous_balance),
        String(r.new_balance),
        r.created_by_email ?? '',
      ]),
    };
  }, [queryParams]);

  const offset = Number(queryParams.offset ?? 0);
  const limit = Number(queryParams.limit ?? 50);

  return (
    <section className="content-section">
      <ReportPageHeader
        title="Movimentações de estoque"
        description="Rastreabilidade de entradas, saídas e ajustes."
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
          Tipo
          <input
            value={values.movement_type ?? ''}
            onChange={(e) => setField('movement_type', e.target.value)}
            placeholder="entry, sale, loss…"
          />
        </label>
      </ReportPageHeader>
      {error && <p className="error">{error}</p>}
      <div className="table-scroll">
        <table>
          <thead>
            <tr>
              <th>Quando</th>
              <th>Produto</th>
              <th>Tipo</th>
              <th>Qtd</th>
              <th>Saldo</th>
              <th>Responsável</th>
            </tr>
          </thead>
          <tbody>
            {items.map((r) => (
              <tr key={r.id}>
                <td>{new Date(r.created_at).toLocaleString('pt-BR')}</td>
                <td>
                  {r.product_name} ({r.sku_code})
                </td>
                <td>{r.movement_type}</td>
                <td>{r.quantity}</td>
                <td>
                  {r.previous_balance} → {r.new_balance}
                </td>
                <td>{r.created_by_email ?? '—'}</td>
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
