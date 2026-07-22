import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import type { DashboardSeriesPoint, ReportDashboard } from '@store/api-client';
import { formatMoney } from '@store/shared-core';
import { api } from '../api';
import { ReportPageHeader } from '../components/reports/ReportPageHeader';
import { ReportSummaryCards } from '../components/reports/ReportSummaryCards';
import { SalesCashSeriesChart } from '../components/reports/SalesCashSeriesChart';
import { useReportQuery } from '../components/reports/useReportQuery';

export function DashboardContent() {
  const { values, setField, queryParams } = useReportQuery({
    year: String(new Date().getFullYear()),
    month: String(new Date().getMonth() + 1),
  });
  const [data, setData] = useState<ReportDashboard | null>(null);
  const [series, setSeries] = useState<DashboardSeriesPoint[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const y = Number(queryParams.year);
    const m = Number(queryParams.month);
    Promise.all([api.adminDashboard(y, m), api.adminDashboardSeries(6)])
      .then(([dash, ser]) => {
        setData(dash);
        setSeries(ser.items ?? []);
      })
      .catch((e: Error) => setError(e.message));
  }, [queryParams.year, queryParams.month]);

  const cards = data
    ? [
        { label: 'Vendas (competência)', value: formatMoney(data.sales_month_cents) },
        { label: 'Recebimentos (caixa)', value: formatMoney(data.received_month_cents) },
        { label: 'Em aberto', value: formatMoney(data.open_receivables_cents) },
        { label: 'Vencido', value: formatMoney(data.overdue_amount_cents) },
        { label: 'Pedidos', value: String(data.orders_month) },
        { label: 'Ticket médio', value: formatMoney(data.average_ticket_cents) },
      ]
    : [];

  function buildExportTable() {
    const period = `${values.month}/${values.year}`;
    const kpiRows: string[][] = cards.map((c) => [c.label, c.value]);
    const seriesRows = series.map((p) => [
      `${String(p.month).padStart(2, '0')}/${p.year}`,
      String(p.sales_cents),
      String(p.received_cents),
    ]);
    return {
      title: 'Dashboard — resumo operacional',
      subtitle: `KPIs ${period} · gerado em ${new Date().toLocaleString('pt-BR')}`,
      filenameBase: `dashboard-${values.year}-${values.month}`,
      headers: ['Mês', 'Vendas (centavos)', 'Recebimentos (centavos)'],
      rows: [...kpiRows.map((r) => [r[0], r[1], '']), ['', '', ''], ...seriesRows],
    };
  }

  return (
    <section className="content-section">
      <ReportPageHeader
        title="Dashboard"
        description="Visão geral da operação: competência (vendas) e caixa (recebimentos)."
        exportTable={buildExportTable}
      >
        <label>
          Ano
          <input
            type="number"
            value={values.year ?? ''}
            onChange={(e) => setField('year', e.target.value)}
            min={2020}
          />
        </label>
        <label>
          Mês
          <input
            type="number"
            value={values.month ?? ''}
            onChange={(e) => setField('month', e.target.value)}
            min={1}
            max={12}
          />
        </label>
      </ReportPageHeader>
      {error && <p className="error">{error}</p>}
      <SalesCashSeriesChart items={series} />
      <ReportSummaryCards cards={cards} />
      {data && (
        <>
          <p className="form-hint">
            <Link to="/relatorios/vendas">Ver vendas</Link> ·{' '}
            <Link to="/relatorios/recebiveis">Ver recebíveis</Link> ·{' '}
            <Link to="/relatorios/estoque?below_minimum=true">
              Estoque abaixo do mínimo ({data.low_stock_skus})
            </Link>
          </p>
          <div className="report-panels">
            <div>
              <h2>Top produtos</h2>
              <ul>
                {data.top_products.map((p) => (
                  <li key={p.label}>
                    {p.label} — {formatMoney(p.total_cents)} ({p.quantity} un.)
                  </li>
                ))}
              </ul>
            </div>
            <div>
              <h2>Top clientes</h2>
              <ul>
                {data.top_customers.map((c) => (
                  <li key={c.id ?? c.label}>
                    {c.id ? <Link to={`/clientes?highlight=${c.id}`}>{c.label}</Link> : c.label} —{' '}
                    {formatMoney(c.total_cents)}
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </>
      )}
    </section>
  );
}
