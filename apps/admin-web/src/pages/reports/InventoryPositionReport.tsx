import { useCallback, useEffect, useState } from 'react';
import type { ReportInventoryPositionRow } from '@store/api-client';
import { formatMoney } from '@store/shared-core';
import { api } from '../../api';
import { fetchAllPages } from '../../components/reports/exportReport';
import { exportSubtitle } from '../../components/reports/exportSubtitle';
import { ReportPageHeader } from '../../components/reports/ReportPageHeader';
import { useReportQuery } from '../../components/reports/useReportQuery';

export function InventoryPositionReportPage() {
  const { values, setField, setOffset, queryParams } = useReportQuery({ situation: '' });
  const [items, setItems] = useState<ReportInventoryPositionRow[]>([]);
  const [total, setTotal] = useState(0);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      const res = await api.adminReportInventoryPosition(queryParams);
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
      api.adminReportInventoryPosition({ ...queryParams, offset, limit }),
    );
    return {
      title: 'Posição do estoque',
      subtitle: exportSubtitle(),
      filenameBase: 'estoque-posicao',
      headers: ['Produto', 'SKU', 'Disponível', 'Mínimo', 'Situação', 'Valor estoque (centavos)'],
      rows: rows.map((r) => [
        r.product_name,
        r.sku_code,
        String(r.available_quantity),
        String(r.minimum_stock),
        r.situation,
        String(r.stock_value_cents),
      ]),
    };
  }, [queryParams]);

  const offset = Number(queryParams.offset ?? 0);
  const limit = Number(queryParams.limit ?? 50);

  return (
    <section className="content-section">
      <ReportPageHeader
        title="Posição do estoque"
        description="Fotografia atual de disponibilidade por SKU."
        exportTable={buildExportTable}
      >
        <label>
          Situação
          <select value={values.situation ?? ''} onChange={(e) => setField('situation', e.target.value)}>
            <option value="">Todas</option>
            <option value="NORMAL">Normal</option>
            <option value="BAIXO">Baixo</option>
            <option value="ZERADO">Zerado</option>
            <option value="INATIVO">Inativo</option>
          </select>
        </label>
        <label>
          <input
            type="checkbox"
            checked={values.below_minimum === 'true'}
            onChange={(e) => setField('below_minimum', e.target.checked ? 'true' : '')}
          />
          Abaixo do mínimo
        </label>
      </ReportPageHeader>
      {error && <p className="error">{error}</p>}
      <div className="table-scroll">
        <table>
          <thead>
            <tr>
              <th>Produto</th>
              <th>SKU</th>
              <th>Disponível</th>
              <th>Mínimo</th>
              <th>Situação</th>
              <th>Valor est.</th>
            </tr>
          </thead>
          <tbody>
            {items.map((r) => (
              <tr key={r.sku_id}>
                <td>{r.product_name}</td>
                <td>{r.sku_code}</td>
                <td>{r.available_quantity}</td>
                <td>{r.minimum_stock}</td>
                <td>{r.situation}</td>
                <td>{formatMoney(r.stock_value_cents)}</td>
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
          {total === 0 ? 0 : offset + 1}–{Math.min(offset + limit, total)} de {total}
        </span>
        <button type="button" disabled={offset + limit >= total} onClick={() => setOffset(offset + limit)}>
          Próxima
        </button>
      </div>
    </section>
  );
}
