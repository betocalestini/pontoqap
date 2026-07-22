import { useCallback, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import type { ReportReceivableRow } from '@store/api-client';
import { formatMoney, labelAgingBucket, labelInvoiceStatus } from '@store/shared-core';
import { api } from '../../api';
import { fetchAllPages } from '../../components/reports/exportReport';
import { exportSubtitle } from '../../components/reports/exportSubtitle';
import { ReportPageHeader } from '../../components/reports/ReportPageHeader';
import { ReportSummaryCards } from '../../components/reports/ReportSummaryCards';
import { useReportQuery } from '../../components/reports/useReportQuery';

export function ReceivablesReportPage() {
  const { values, setField, setOffset, queryParams } = useReportQuery({ status: '', overdue_only: '' });
  const [items, setItems] = useState<ReportReceivableRow[]>([]);
  const [total, setTotal] = useState(0);
  const [summary, setSummary] = useState<Record<string, number>>({});
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      const res = await api.adminReportReceivables(queryParams);
      setItems(res.items ?? []);
      setTotal(res.total ?? 0);
      setSummary((res.summary as Record<string, number>) ?? {});
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Erro');
    }
  }, [queryParams]);

  useEffect(() => {
    void load();
  }, [load]);

  const buildExportTable = useCallback(async () => {
    const rows = await fetchAllPages((offset, limit) =>
      api.adminReportReceivables({ ...queryParams, offset, limit }),
    );
    return {
      title: 'Contas a receber',
      subtitle: exportSubtitle(),
      filenameBase: 'contas-a-receber',
      headers: ['Fatura', 'Cliente', 'Competência', 'Saldo (centavos)', 'Status', 'Faixa atraso'],
      rows: rows.map((r) => [
        r.invoice_number,
        r.customer_name,
        `${String(r.reference_month).padStart(2, '0')}/${r.reference_year}`,
        String(r.remaining_cents),
        labelInvoiceStatus(r.status),
        labelAgingBucket(r.aging_bucket),
      ]),
    };
  }, [queryParams]);

  const offset = Number(queryParams.offset ?? 0);
  const limit = Number(queryParams.limit ?? 50);

  return (
    <section className="content-section">
      <ReportPageHeader
        title="Contas a receber"
        description="Faturas abertas, pagas e vencidas."
        exportTable={buildExportTable}
      >
        <label>
          Status
          <select value={values.status ?? ''} onChange={(e) => setField('status', e.target.value)}>
            <option value="">Todos</option>
            <option value="open">Em aberto</option>
            <option value="overdue">Vencida</option>
            <option value="paid">Paga</option>
          </select>
        </label>
        <label>
          <input
            type="checkbox"
            checked={values.overdue_only === 'true'}
            onChange={(e) => setField('overdue_only', e.target.checked ? 'true' : '')}
          />
          Só vencidas
        </label>
      </ReportPageHeader>
      {error && <p className="error">{error}</p>}
      <ReportSummaryCards
        cards={[
          { label: 'Faturado', value: formatMoney(summary.total_billed_cents ?? 0) },
          { label: 'Recebido', value: formatMoney(summary.total_received_cents ?? 0) },
          { label: 'Em aberto', value: formatMoney(summary.total_open_cents ?? 0) },
          { label: 'Vencido', value: formatMoney(summary.total_overdue_cents ?? 0) },
        ]}
      />
      <div className="table-scroll">
        <table>
          <thead>
            <tr>
              <th>Fatura</th>
              <th>Cliente</th>
              <th>Competência</th>
              <th>Saldo</th>
              <th>Faixa</th>
            </tr>
          </thead>
          <tbody>
            {items.map((r) => (
              <tr key={r.id}>
                <td>
                  <Link to={`/faturamento/${r.id}`}>{r.invoice_number}</Link>
                </td>
                <td>{r.customer_name}</td>
                <td>
                  {String(r.reference_month).padStart(2, '0')}/{r.reference_year}
                </td>
                <td>{formatMoney(r.remaining_cents)}</td>
                <td>{labelAgingBucket(r.aging_bucket)}</td>
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
