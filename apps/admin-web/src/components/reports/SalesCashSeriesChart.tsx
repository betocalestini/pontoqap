import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import type { DashboardSeriesPoint } from '@store/api-client';
import { formatMoney } from '@store/shared-core';

function monthLabel(year: number, month: number) {
  return `${String(month).padStart(2, '0')}/${year}`;
}

type Props = {
  items: DashboardSeriesPoint[];
};

export function SalesCashSeriesChart({ items }: Props) {
  const data = items.map((p) => ({
    name: monthLabel(p.year, p.month),
    vendas: p.sales_cents / 100,
    recebimentos: p.received_cents / 100,
    sales_cents: p.sales_cents,
    received_cents: p.received_cents,
  }));

  if (!data.length) {
    return <p className="form-hint">Sem dados para o gráfico.</p>;
  }

  return (
    <div className="report-chart">
      <h2 className="report-chart__title">Vendas (competência) vs recebimentos (caixa) — últimos meses</h2>
      <div className="report-chart__plot" aria-hidden={false}>
        <ResponsiveContainer width="100%" height={280}>
          <BarChart data={data} margin={{ top: 8, right: 16, left: 8, bottom: 8 }}>
            <CartesianGrid strokeDasharray="3 3" opacity={0.3} />
            <XAxis dataKey="name" tick={{ fontSize: 12 }} />
            <YAxis tick={{ fontSize: 11 }} tickFormatter={(v) => `R$${v}`} />
            <Tooltip
              formatter={(value, name) => {
                const n = typeof value === 'number' ? value : Number(value ?? 0);
                const key = String(name ?? '');
                return [
                  formatMoney(Math.round(n * 100)),
                  key === 'vendas' ? 'Vendas' : 'Recebimentos',
                ];
              }}
            />
            <Legend formatter={(v) => (v === 'vendas' ? 'Vendas (competência)' : 'Recebimentos (caixa)')} />
            <Bar dataKey="vendas" fill="var(--color-primary, #2563eb)" radius={[4, 4, 0, 0]} />
            <Bar dataKey="recebimentos" fill="var(--color-accent, #059669)" radius={[4, 4, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </div>
      <table className="report-chart__fallback">
        <caption className="sr-only">Dados do gráfico em tabela</caption>
        <thead>
          <tr>
            <th>Mês</th>
            <th>Vendas</th>
            <th>Recebimentos</th>
          </tr>
        </thead>
        <tbody>
          {items.map((p) => (
            <tr key={`${p.year}-${p.month}`}>
              <td>{monthLabel(p.year, p.month)}</td>
              <td>{formatMoney(p.sales_cents)}</td>
              <td>{formatMoney(p.received_cents)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
