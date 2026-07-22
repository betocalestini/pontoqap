import { useCallback, useEffect, useState } from 'react';
import type { ReportExceptionRow } from '@store/api-client';
import { formatMoney, labelExceptionEventType } from '@store/shared-core';
import { api } from '../../api';
import { fetchAllPages } from '../../components/reports/exportReport';
import { exportSubtitle } from '../../components/reports/exportSubtitle';
import { ReportPageHeader } from '../../components/reports/ReportPageHeader';
import { useReportQuery } from '../../components/reports/useReportQuery';

export function ExceptionsReportPage() {
  const now = new Date();
  const from = new Date(now.getFullYear(), now.getMonth(), 1).toISOString().slice(0, 10);
  const to = now.toISOString().slice(0, 10);
  const { values, setField, setOffset, queryParams } = useReportQuery({ date_from: from, date_to: to });
  const [items, setItems] = useState<ReportExceptionRow[]>([]);
  const [total, setTotal] = useState(0);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      const res = await api.adminReportExceptions(queryParams);
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
      api.adminReportExceptions({ ...queryParams, offset, limit }),
    );
    return {
      title: 'Exceções e ajustes',
      subtitle: exportSubtitle(),
      filenameBase: 'excecoes',
      headers: ['Quando', 'Tipo', 'Referência', 'Valor (centavos)', 'Quantidade', 'Responsável'],
      rows: rows.map((r) => [
        r.occurred_at,
        labelExceptionEventType(r.event_type),
        r.label,
        r.amount_cents != null ? String(r.amount_cents) : '',
        r.quantity != null ? String(r.quantity) : '',
        r.actor_email ?? '',
      ]),
    };
  }, [queryParams]);

  const offset = Number(queryParams.offset ?? 0);
  const limit = Number(queryParams.limit ?? 50);

  return (
    <section className="content-section">
      <ReportPageHeader title="Exceções e ajustes" description="Cancelamentos, perdas, ajustes de fatura e similares." exportTable={buildExportTable}>
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
          <input value={values.event_type ?? ''} onChange={(e) => setField('event_type', e.target.value)} />
        </label>
      </ReportPageHeader>
      {error && <p className="error">{error}</p>}
      <div className="table-scroll">
        <table>
          <thead>
            <tr>
              <th>Quando</th>
              <th>Tipo</th>
              <th>Referência</th>
              <th>Valor / Qtd</th>
              <th>Responsável</th>
            </tr>
          </thead>
          <tbody>
            {items.map((r, i) => (
              <tr key={`${r.entity_id}-${i}`}>
                <td>{new Date(r.occurred_at).toLocaleString('pt-BR')}</td>
                <td>{labelExceptionEventType(r.event_type)}</td>
                <td>{r.label}</td>
                <td>
                  {r.amount_cents != null ? formatMoney(r.amount_cents) : r.quantity != null ? r.quantity : '—'}
                </td>
                <td>{r.actor_email || '—'}</td>
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
