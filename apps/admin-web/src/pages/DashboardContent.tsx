import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import type { DashboardSeriesPoint, ReportDashboard } from '@store/api-client';
import { formatMoney } from '@store/shared-core';
import { api } from '../api';
import { RankingPanel } from '../components/reports/RankingPanel';
import { ReportPageHeader } from '../components/reports/ReportPageHeader';
import { ReportSummaryCards } from '../components/reports/ReportSummaryCards';
import { SalesCashSeriesChart } from '../components/reports/SalesCashSeriesChart';
import { useReportQuery } from '../components/reports/useReportQuery';

const RANKING_EMPTY = 'Nenhum registro no período selecionado.';

function clampMonthForYear(year: number, month: number, now: Date): number {
  if (year === now.getFullYear() && month > now.getMonth() + 1) {
    return now.getMonth() + 1;
  }
  return month;
}

export function DashboardContent() {
  const now = new Date();
  const { values, setField, queryParams } = useReportQuery({
    year: String(now.getFullYear()),
    month: String(now.getMonth() + 1),
  });
  const [data, setData] = useState<ReportDashboard | null>(null);
  const [series, setSeries] = useState<DashboardSeriesPoint[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const y = Number(queryParams.year);
    const m = Number(queryParams.month);
    setError(null);
    Promise.all([api.adminDashboard(y, m), api.adminDashboardSeries(6)])
      .then(([dash, ser]) => {
        setData(dash);
        setSeries(ser.items ?? []);
      })
      .catch((e: Error) => setError(e.message));
  }, [queryParams.year, queryParams.month]);

  const yearNum = Number(values.year) || now.getFullYear();
  const maxMonth = yearNum === now.getFullYear() ? now.getMonth() + 1 : 12;

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
    <section className="content-section dashboard-page">
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
            onChange={(e) => {
              setField('year', e.target.value);
              const y = Number(e.target.value);
              const m = Number(values.month);
              if (!Number.isNaN(y) && !Number.isNaN(m)) {
                const clamped = clampMonthForYear(y, m, now);
                if (clamped !== m) {
                  setField('month', String(clamped));
                }
              }
            }}
            min={2020}
          />
        </label>
        <label>
          Mês
          <input
            type="number"
            value={values.month ?? ''}
            onChange={(e) => {
              const raw = Number(e.target.value);
              if (!e.target.value || Number.isNaN(raw)) {
                setField('month', e.target.value);
                return;
              }
              const clamped = clampMonthForYear(yearNum, raw, now);
              setField('month', String(clamped));
            }}
            min={1}
            max={maxMonth}
          />
        </label>
      </ReportPageHeader>
      {error && <p className="error">{error}</p>}
      <SalesCashSeriesChart items={series} />
      <ReportSummaryCards cards={cards} />
      {data && (
        <>
          <nav className="dashboard-actions" aria-label="Atalhos do dashboard">
            <Link to="/relatorios/vendas">Ver vendas</Link>
            <Link to="/relatorios/recebiveis">Ver recebíveis</Link>
            <Link to="/relatorios/estoque?below_minimum=true">
              Estoque abaixo do mínimo ({data.low_stock_skus})
            </Link>
          </nav>
          <div className="report-panels">
            <RankingPanel
              title="Top produtos"
              emptyMessage={RANKING_EMPTY}
              items={(data.top_products ?? []).map((p) => ({
                key: p.label,
                primary: p.label,
                secondary: `${formatMoney(p.total_cents)} (${p.quantity} un.)`,
              }))}
            />
            <RankingPanel
              title="Top clientes"
              emptyMessage={RANKING_EMPTY}
              items={(data.top_customers ?? []).map((c) => ({
                key: c.id ?? c.label,
                primary: c.label,
                secondary: formatMoney(c.total_cents),
                href: c.id ? `/clientes?highlight=${c.id}` : undefined,
              }))}
            />
          </div>
        </>
      )}
    </section>
  );
}
