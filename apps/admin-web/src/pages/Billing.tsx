import { useCallback, useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import type { AdminBillingSummary, AdminInvoiceListItem } from '@store/api-client';
import { formatMoney, labelInvoiceStatus } from '@store/shared-core';
import { api } from '../api';
import { usePermissions } from '../auth/usePermissions';

const PAGE_SIZE = 50;

function previousMonth(): { year: number; month: number } {
  const now = new Date();
  let y = now.getFullYear();
  let m = now.getMonth() + 1;
  m -= 1;
  if (m < 1) {
    m = 12;
    y -= 1;
  }
  return { year: y, month: m };
}

function formatCompetence(year: number, month: number) {
  return `${String(month).padStart(2, '0')}/${year}`;
}

export function BillingPage() {
  const permissions = usePermissions();
  const canClose = permissions.includes('billing.close');

  const defaultRef = useMemo(() => previousMonth(), []);
  const [summary, setSummary] = useState<AdminBillingSummary | null>(null);
  const [items, setItems] = useState<AdminInvoiceListItem[]>([]);
  const [total, setTotal] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [closeYear, setCloseYear] = useState(defaultRef.year);
  const [closeMonth, setCloseMonth] = useState(defaultRef.month);
  const [closeReason, setCloseReason] = useState('');
  const [closeMsg, setCloseMsg] = useState<string | null>(null);
  const [closing, setClosing] = useState(false);
  const [filters, setFilters] = useState({ status: '', year: '', month: '', search: '' });

  const loadSummary = useCallback(async () => {
    setSummary(await api.adminBillingSummary());
  }, []);

  const loadInvoices = useCallback(async () => {
    const res = await api.adminListInvoices({
      status: filters.status || undefined,
      year: filters.year ? parseInt(filters.year, 10) : undefined,
      month: filters.month ? parseInt(filters.month, 10) : undefined,
      search: filters.search || undefined,
      limit: PAGE_SIZE,
      offset: 0,
    });
    setItems(res.items ?? []);
    setTotal(res.total ?? 0);
  }, [filters]);

  useEffect(() => {
    loadSummary().catch((e: Error) => setError(e.message));
  }, [loadSummary]);

  useEffect(() => {
    loadInvoices().catch((e: Error) => setError(e.message));
  }, [loadInvoices]);

  async function submitClose(e: React.FormEvent) {
    e.preventDefault();
    if (!canClose) return;
    setClosing(true);
    setError(null);
    setCloseMsg(null);
    try {
      const res = await api.adminCloseBilling({
        year: closeYear,
        month: closeMonth,
        reason: closeReason.trim(),
      });
      setCloseMsg(
        `Fechados ${res.closed_periods} período(s) para competência ${formatCompetence(res.year, res.month)}.`,
      );
      setCloseReason('');
      await Promise.all([loadSummary(), loadInvoices()]);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao fechar');
    } finally {
      setClosing(false);
    }
  }

  const yearOptions = useMemo(() => {
    const y = new Date().getFullYear();
    return [y - 1, y, y + 1];
  }, []);

  return (
    <section className="content-section billing-page">
      <h1>Faturamento</h1>
      <p>
        <Link to="/faturamento/parcelas">Configuração de parcelamento</Link>
      </p>
      {error && <p className="error">{error}</p>}

      {summary && (
        <div className="billing-summary">
          <div className="billing-summary__card">
            <span className="billing-summary__label">A receber</span>
            <strong>{formatMoney(summary.open_receivables_cents)}</strong>
          </div>
          <div className="billing-summary__card">
            <span className="billing-summary__label">Faturas vencidas</span>
            <strong>{summary.overdue_invoices_count}</strong>
          </div>
          <div className="billing-summary__card">
            <span className="billing-summary__label">Competências abertas</span>
            <strong>
              {summary.open_periods_count} · {formatMoney(summary.open_periods_total_cents)}
            </strong>
          </div>
          {(summary.scheduled_monthly_close_today ?? summary.scheduled_closing_today) && (
            <p className="billing-summary__hint form-hint">
              Hoje é dia 1: o worker fecha automaticamente a competência do mês anterior (vencimento dia 10).
            </p>
          )}
        </div>
      )}

      <div className="billing-panel">
        <h2 className="billing-panel__title">Fechar competência</h2>
        {!canClose ? (
          <p className="form-hint">
            Sua conta não tem permissão para fechar períodos (<code>billing.close</code>). Peça a um
            gerente ou administrador.
          </p>
        ) : (
          <>
            <p className="form-hint">
              Gera faturas para todos os clientes com período aberto na competência escolhida. O fechamento
              automático ocorre todo dia 1; use este fluxo apenas em exceção. O motivo é obrigatório e fica
              registrado em auditoria.
            </p>
            <form className="billing-close-form" onSubmit={(e) => void submitClose(e)}>
              <label>
                Ano
                <select value={closeYear} onChange={(e) => setCloseYear(parseInt(e.target.value, 10))}>
                  {yearOptions.map((y) => (
                    <option key={y} value={y}>
                      {y}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                Mês
                <select value={closeMonth} onChange={(e) => setCloseMonth(parseInt(e.target.value, 10))}>
                  {Array.from({ length: 12 }, (_, i) => i + 1).map((m) => (
                    <option key={m} value={m}>
                      {String(m).padStart(2, '0')}
                    </option>
                  ))}
                </select>
              </label>
              <label className="billing-close-form__reason">
                Motivo do fechamento
                <textarea
                  value={closeReason}
                  onChange={(e) => setCloseReason(e.target.value)}
                  placeholder="Ex.: homologação — fechamento da competência 03/2026"
                  minLength={10}
                  rows={3}
                  required
                />
                <span className="form-hint">Mínimo 10 caracteres.</span>
              </label>
              <div className="billing-close-form__actions">
                <button type="submit" disabled={closing || closeReason.trim().length < 10}>
                  {closing ? 'Fechando…' : 'Fechar competência'}
                </button>
              </div>
            </form>
            {closeMsg && <p className="billing-close-msg">{closeMsg}</p>}
          </>
        )}
      </div>

      <div className="billing-panel">
        <h2 className="billing-panel__title">Faturas</h2>
        <div className="billing-list-filters">
          <label>
            Buscar
            <input
              type="search"
              placeholder="Cliente, e-mail, número…"
              value={filters.search}
              onChange={(e) => setFilters((f) => ({ ...f, search: e.target.value }))}
            />
          </label>
          <label>
            Status
            <select value={filters.status} onChange={(e) => setFilters((f) => ({ ...f, status: e.target.value }))}>
              <option value="">Todos</option>
              <option value="open">Em aberto</option>
              <option value="overdue">Vencida</option>
              <option value="paid">Paga</option>
            </select>
          </label>
          <label>
            Ano
            <input
              type="number"
              min={2020}
              max={2100}
              value={filters.year}
              onChange={(e) => setFilters((f) => ({ ...f, year: e.target.value }))}
              placeholder="Todos"
            />
          </label>
          <label>
            Mês
            <input
              type="number"
              min={1}
              max={12}
              value={filters.month}
              onChange={(e) => setFilters((f) => ({ ...f, month: e.target.value }))}
              placeholder="Todos"
            />
          </label>
        </div>
        <p className="billing-list-meta">{total} fatura(s)</p>
        <div className="table-scroll">
          <table className="billing-table">
            <thead>
              <tr>
                <th>Cliente</th>
                <th>Competência</th>
                <th>Número</th>
                <th>Status</th>
                <th>Total</th>
                <th>Em aberto</th>
                <th>Vencimento</th>
              </tr>
            </thead>
            <tbody>
              {items.map((inv) => (
                <tr key={inv.id}>
                  <td>
                    <Link to={`/faturamento/${inv.id}`}>{inv.customer_name}</Link>
                    <div className="billing-table__sub">{inv.customer_email}</div>
                  </td>
                  <td>{formatCompetence(inv.reference_year, inv.reference_month)}</td>
                  <td>
                    <Link to={`/faturamento/${inv.id}`}>{inv.invoice_number}</Link>
                  </td>
                  <td>
                    <span className={`billing-status billing-status--${inv.status}`}>
                      {labelInvoiceStatus(inv.status)}
                    </span>
                  </td>
                  <td>{formatMoney(inv.total_cents)}</td>
                  <td>{formatMoney(inv.remaining_cents)}</td>
                  <td>{inv.due_at ? new Date(inv.due_at).toLocaleDateString('pt-BR') : '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {items.length === 0 && <p className="billing-panel__empty">Nenhuma fatura encontrada.</p>}
      </div>
    </section>
  );
}
