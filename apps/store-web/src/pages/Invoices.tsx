import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { formatMoney } from '@store/shared-core';
import { api } from '../api';
import { AuthGuestPrompt } from '../components/AuthGuestPrompt';
import { guestAuthMessage, isGuestAuthError } from '../utils/authGuest';

type OpenPeriod = {
  billing_period_id: string;
  reference_year: number;
  reference_month: number;
  status: string;
  total_cents: number;
  entry_count: number;
};

type Invoice = {
  id: string;
  invoice_number: string;
  status: string;
  total_cents: number;
  paid_cents: number;
  due_at?: string;
};

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

export function InvoicesPage() {
  const [current, setCurrent] = useState<OpenPeriod | null>(null);
  const [items, setItems] = useState<Invoice[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [needsAuth, setNeedsAuth] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    setNeedsAuth(false);
    api
      .listMyInvoices()
      .then((res) => {
        const data = res as { current_period?: OpenPeriod | null; items?: Invoice[] };
        setCurrent(data.current_period ?? null);
        setItems(data.items ?? []);
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
  }, []);

  const hasCurrent = current != null;
  const hasHistory = items.length > 0;

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
                {' — '}
                <span className="badge">Em acumulação</span>
              </p>
              <p>Total parcial: {formatMoney(current.total_cents)}</p>
              <p className="invoice-card-meta">
                {current.entry_count === 0
                  ? 'Nenhuma compra nesta competência ainda.'
                  : `${current.entry_count} lançamento(s) — a fatura será gerada no fechamento mensal.`}
              </p>
            </div>
          ) : (
            <p className="invoice-card-meta">Nenhuma competência em aberto no momento.</p>
          )}

          <h2>Últimas faturas</h2>
          {hasHistory ? (
            <ul className="invoice-list">
              {items.map((inv) => {
                const remaining = Math.max(0, inv.total_cents - inv.paid_cents);
                return (
                  <li key={inv.id} className="invoice-card">
                    <Link to={`/faturas/${inv.id}`}>
                      <strong>{inv.invoice_number}</strong>
                    </Link>
                    <span className="badge">{invoiceStatusLabel(inv.status)}</span>
                    <p>
                      Total {formatMoney(inv.total_cents)}
                      {remaining > 0 && ` · ${formatMoney(remaining)} em aberto`}
                      {remaining === 0 && inv.total_cents > 0 && ' · quitada'}
                    </p>
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
            <p className="invoice-card-meta">Ainda não há faturas fechadas. Após o fechamento do mês, elas aparecem aqui.</p>
          )}

          {!hasCurrent && !hasHistory && (
            <p>Faça compras na loja para ver valores na competência atual.</p>
          )}
        </>
      )}
    </section>
  );
}
