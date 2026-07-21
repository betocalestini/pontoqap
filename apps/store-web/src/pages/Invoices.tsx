import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import type { MyInvoiceListItem, OpenBillingPeriod } from '@store/api-client';
import { formatMoney } from '@store/shared-core';
import { api } from '../api';
import { AuthGuestPrompt } from '../components/AuthGuestPrompt';
import { guestAuthMessage, isGuestAuthError } from '../utils/authGuest';

function formatCompetence(year: number, month: number) {
  return `${String(month).padStart(2, '0')}/${year}`;
}

function invoiceStatusLabel(status: string) {
  switch (status) {
    case 'open':
      return 'Em aberto';
    case 'paid':
      return 'Paga';
    case 'overdue':
      return 'Vencida';
    default:
      return status;
  }
}

function closeTypeLabel(closeType?: string) {
  switch (closeType) {
    case 'customer_request':
      return 'Fechamento parcial';
    case 'monthly_auto':
      return 'Fechamento mensal';
    case 'admin_manual':
      return 'Fechamento manual';
    default:
      return null;
  }
}

export function InvoicesPage() {
  const navigate = useNavigate();
  const [current, setCurrent] = useState<OpenBillingPeriod | null>(null);
  const [items, setItems] = useState<MyInvoiceListItem[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [needsAuth, setNeedsAuth] = useState(false);
  const [loading, setLoading] = useState(true);
  const [closing, setClosing] = useState(false);

  function reload() {
    setLoading(true);
    setNeedsAuth(false);
    api
      .listMyInvoices()
      .then((res) => {
        setCurrent(res.current_period ?? null);
        setItems(res.items ?? []);
      })
      .catch((e: Error) => {
        if (isGuestAuthError(e)) {
          setNeedsAuth(true);
          setError(null);
        } else {
          setError(e.message);
        }
      })
      .finally(() => setLoading(false));
  }

  useEffect(() => {
    reload();
  }, []);

  async function handleCloseCycle() {
    setClosing(true);
    setError(null);
    try {
      const inv = await api.closeMyBillingCycle();
      navigate(`/faturas/${inv.id}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Não foi possível fechar a fatura');
    } finally {
      setClosing(false);
    }
  }

  const hasCurrent = current != null;
  const hasHistory = items.length > 0;
  const canClosePartial =
    hasCurrent && current.entry_count > 0 && (current.total_cents ?? 0) > 0;

  return (
    <section className="content-section invoices-page">
      <h1>Minhas faturas</h1>
      {loading && <p>Carregando…</p>}
      {needsAuth && <AuthGuestPrompt message={guestAuthMessage('invoices')} />}
      {error && <p className="error">{error}</p>}
      {!loading && !error && !needsAuth && (
        <>
          <h2>Competência atual</h2>
          {hasCurrent ? (
            <div className="invoice-card invoice-card--current">
              <p>
                <strong>{formatCompetence(current.reference_year, current.reference_month)}</strong>
                {current.cycle_number != null && current.cycle_number > 1 && (
                  <> · ciclo {current.cycle_number}</>
                )}
                {' — '}
                <span className="badge">Em acumulação</span>
              </p>
              <p>Total parcial: {formatMoney(current.total_cents)}</p>
              <p className="invoice-card-meta">
                {current.entry_count === 0
                  ? 'Nenhuma compra nesta competência ainda.'
                  : `${current.entry_count} lançamento(s). Fechamento automático no dia 1; você também pode fechar e pagar agora.`}
              </p>
              {canClosePartial && (
                <button type="button" className="btn-primary" disabled={closing} onClick={handleCloseCycle}>
                  {closing ? 'Fechando…' : 'Fechar fatura e pagar'}
                </button>
              )}
            </div>
          ) : (
            <p className="invoice-card-meta">Nenhuma competência em aberto no momento.</p>
          )}

          <h2>Histórico de faturas</h2>
          {hasHistory ? (
            <ul className="invoice-list">
              {items.map((inv) => {
                const remaining = Math.max(0, inv.total_cents - inv.paid_cents);
                const typeLabel = closeTypeLabel(inv.close_type);
                return (
                  <li key={inv.id} className="invoice-card">
                    <Link to={`/faturas/${inv.id}`}>
                      <strong>{inv.invoice_number}</strong>
                    </Link>
                    <span className="badge">{invoiceStatusLabel(inv.status)}</span>
                    {typeLabel && <span className="invoice-card-meta"> · {typeLabel}</span>}
                    <p>
                      Total {formatMoney(inv.total_cents)}
                      {remaining > 0 && ` · ${formatMoney(remaining)} em aberto`}
                      {remaining === 0 && inv.total_cents > 0 && ' · quitada'}
                    </p>
                    {inv.reference_year != null && inv.reference_month != null && (
                      <p className="invoice-card-meta">
                        Competência {formatCompetence(inv.reference_year, inv.reference_month)}
                        {inv.cycle_number != null && inv.cycle_number > 1 && ` · ciclo ${inv.cycle_number}`}
                      </p>
                    )}
                    {inv.due_at && (
                      <p className="invoice-card-meta">
                        Vencimento: {new Date(inv.due_at).toLocaleDateString('pt-BR')}
                      </p>
                    )}
                  </li>
                );
              })}
            </ul>
          ) : (
            <p className="invoice-card-meta">Ainda não há faturas fechadas.</p>
          )}

          {!hasCurrent && !hasHistory && (
            <p>Faça compras na loja para ver valores na competência atual.</p>
          )}
        </>
      )}
    </section>
  );
}
