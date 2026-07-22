import { useCallback, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import type { ReportCustomerExposureRow } from '@store/api-client';
import { formatMoney } from '@store/shared-core';
import { api } from '../../api';
import { fetchAllPages } from '../../components/reports/exportReport';
import { exportSubtitle } from '../../components/reports/exportSubtitle';
import { ReportPageHeader } from '../../components/reports/ReportPageHeader';
import { useReportQuery } from '../../components/reports/useReportQuery';

export function CustomerExposureReportPage() {
  const { values, setField, setOffset, queryParams } = useReportQuery({ overdue_only: '' });
  const [items, setItems] = useState<ReportCustomerExposureRow[]>([]);
  const [total, setTotal] = useState(0);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      const res = await api.adminReportCustomerExposure(queryParams);
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
      api.adminReportCustomerExposure({ ...queryParams, offset, limit }),
    );
    return {
      title: 'Limites e exposição',
      subtitle: exportSubtitle(),
      filenameBase: 'limites-exposicao',
      headers: [
        'Cliente',
        'E-mail',
        'Limite (centavos)',
        'Exposição (centavos)',
        'Disponível (centavos)',
        'Uso %',
        'Faixa',
      ],
      rows: rows.map((r) => [
        r.customer_name,
        r.customer_email,
        String(r.credit_limit_cents),
        String(r.current_exposure_cents),
        String(r.available_cents),
        String(r.utilization_percent),
        r.utilization_band,
      ]),
    };
  }, [queryParams]);

  const offset = Number(queryParams.offset ?? 0);
  const limit = Number(queryParams.limit ?? 50);

  return (
    <section className="content-section">
      <ReportPageHeader
        title="Limites e exposição"
        description="Uso de crédito interno e risco por cliente."
        exportTable={buildExportTable}
      >
        <label>
          <input
            type="checkbox"
            checked={values.overdue_only === 'true'}
            onChange={(e) => setField('overdue_only', e.target.checked ? 'true' : '')}
          />
          Com fatura vencida
        </label>
        <label>
          <input
            type="checkbox"
            checked={values.limit_exhausted === 'true'}
            onChange={(e) => setField('limit_exhausted', e.target.checked ? 'true' : '')}
          />
          Limite esgotado
        </label>
      </ReportPageHeader>
      {error && <p className="error">{error}</p>}
      <div className="table-scroll">
        <table>
          <thead>
            <tr>
              <th>Cliente</th>
              <th>Limite</th>
              <th>Exposição</th>
              <th>Disponível</th>
              <th>Uso %</th>
              <th>Faixa</th>
            </tr>
          </thead>
          <tbody>
            {items.map((r) => (
              <tr key={r.customer_id}>
                <td>
                  <Link to="/clientes">{r.customer_name}</Link>
                </td>
                <td>{formatMoney(r.credit_limit_cents)}</td>
                <td>{formatMoney(r.current_exposure_cents)}</td>
                <td>{formatMoney(r.available_cents)}</td>
                <td>{r.utilization_percent}%</td>
                <td>{r.utilization_band}</td>
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
