import { useCallback, useEffect, useState } from 'react';
import type { ReportForecastRow } from '@store/api-client';
import { labelForecastConfidence, labelForecastMethod } from '@store/shared-core';
import { api } from '../../api';
import { exportSubtitle } from '../../components/reports/exportSubtitle';
import { ReportPageHeader } from '../../components/reports/ReportPageHeader';
import { useReportQuery } from '../../components/reports/useReportQuery';

export function ForecastReportPage() {
  const { queryParams } = useReportQuery({});
  const [items, setItems] = useState<ReportForecastRow[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [generating, setGenerating] = useState(false);

  const load = useCallback(async () => {
    try {
      const res = await api.adminForecast(Number(queryParams.limit) || 100);
      setItems(res.items ?? []);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Erro');
    }
  }, [queryParams.limit]);

  useEffect(() => {
    void load();
  }, [load]);

  async function generate() {
    setGenerating(true);
    setError(null);
    try {
      await api.adminGenerateForecast();
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Erro ao gerar');
    } finally {
      setGenerating(false);
    }
  }

  const buildExportTable = useCallback(async () => {
    const res = await api.adminForecast(10_000);
    const rows = res.items ?? [];
    return {
      title: 'Previsão de reposição',
      subtitle: exportSubtitle(),
      filenameBase: 'previsao-reposicao',
      headers: [
        'SKU',
        'Produto',
        'Vendas 3m',
        'Estoque',
        'Previsão',
        'Sugestão compra',
        'Confiança',
        'Método',
      ],
      rows: rows.map((r) => [
        r.sku_code,
        r.product_name ?? '',
        String(r.sales_last_3_months),
        String(r.current_stock),
        String(r.forecast_quantity),
        String(r.suggested_purchase_quantity),
        labelForecastConfidence(r.confidence_level),
        labelForecastMethod(r.method),
      ]),
    };
  }, []);

  return (
    <section className="content-section">
      <ReportPageHeader
        title="Previsão de reposição"
        description="Sugestão de compra com base na média dos últimos 3 meses (apoio à decisão)."
        exportTable={buildExportTable}
      >
        <button type="button" onClick={() => void generate()} disabled={generating}>
          {generating ? 'Gerando…' : 'Gerar snapshots'}
        </button>
      </ReportPageHeader>
      {error && <p className="error">{error}</p>}
      <div className="table-scroll">
        <table>
          <thead>
            <tr>
              <th>SKU</th>
              <th>Produto</th>
              <th>Vendas 3m</th>
              <th>Estoque</th>
              <th>Previsão</th>
              <th>Sugestão compra</th>
              <th>Confiança</th>
            </tr>
          </thead>
          <tbody>
            {items.map((r) => (
              <tr key={`${r.sku_id}-${r.reference_month}`}>
                <td>{r.sku_code}</td>
                <td>{r.product_name}</td>
                <td>{r.sales_last_3_months}</td>
                <td>{r.current_stock}</td>
                <td>{r.forecast_quantity}</td>
                <td>{r.suggested_purchase_quantity}</td>
                <td>{labelForecastConfidence(r.confidence_level)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
