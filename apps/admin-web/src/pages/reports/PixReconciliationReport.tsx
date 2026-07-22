import { useCallback, useEffect, useState } from 'react';
import type { ReportPixRow } from '@store/api-client';
import { formatMoney } from '@store/shared-core';
import { api } from '../../api';
import { fetchAllPages } from '../../components/reports/exportReport';
import { exportSubtitle } from '../../components/reports/exportSubtitle';
import { ReportPageHeader } from '../../components/reports/ReportPageHeader';
import { useReportQuery } from '../../components/reports/useReportQuery';

export function PixReconciliationReportPage() {
  const { values, setField, setOffset, queryParams } = useReportQuery({
    divergence_only: '',
    status: '',
  });
  const [items, setItems] = useState<ReportPixRow[]>([]);
  const [total, setTotal] = useState(0);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      const res = await api.adminReportPixReconciliation(queryParams);
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
      api.adminReportPixReconciliation({ ...queryParams, offset, limit }),
    );
    return {
      title: 'Conciliação Pix',
      subtitle: exportSubtitle(),
      filenameBase: 'conciliacao-pix',
      headers: ['Cliente', 'Fatura', 'TxID', 'Cobrança (centavos)', 'Recebido (centavos)', 'Conciliação'],
      rows: rows.map((r) => [
        r.customer_name,
        r.invoice_number,
        r.txid ?? '',
        String(r.charge_amount_cents),
        String(r.received_amount_cents),
        r.reconciliation_status,
      ]),
    };
  }, [queryParams]);

  const offset = Number(queryParams.offset ?? 0);
  const limit = Number(queryParams.limit ?? 50);

  return (
    <section className="content-section">
      <ReportPageHeader
        title="Conciliação Pix"
        description="Cobranças, pagamentos liquidados e situação de conciliação."
        exportTable={buildExportTable}
      >
        <label>
          Situação
          <select value={values.status ?? ''} onChange={(e) => setField('status', e.target.value)}>
            <option value="">Todas</option>
            <option value="CONCILIADO">Conciliado</option>
            <option value="PENDENTE">Pendente</option>
            <option value="VALOR_DIVERGENTE">Valor divergente</option>
            <option value="EXPIRADO">Expirado</option>
          </select>
        </label>
        <label>
          <input
            type="checkbox"
            checked={values.divergence_only === 'true'}
            onChange={(e) => setField('divergence_only', e.target.checked ? 'true' : '')}
          />
          Só divergências
        </label>
      </ReportPageHeader>
      {error && <p className="error">{error}</p>}
      <div className="table-scroll">
        <table>
          <thead>
            <tr>
              <th>Cliente</th>
              <th>Fatura</th>
              <th>TxID</th>
              <th>Cobrança</th>
              <th>Recebido</th>
              <th>Conciliação</th>
            </tr>
          </thead>
          <tbody>
            {items.map((r, i) => (
              <tr key={`${r.invoice_number}-${i}`}>
                <td>{r.customer_name}</td>
                <td>{r.invoice_number}</td>
                <td>
                  <code>{r.txid || '—'}</code>
                </td>
                <td>{formatMoney(r.charge_amount_cents)}</td>
                <td>{formatMoney(r.received_amount_cents)}</td>
                <td>{r.reconciliation_status}</td>
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
