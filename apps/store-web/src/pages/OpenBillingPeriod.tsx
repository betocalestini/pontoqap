import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import type { OpenBillingPeriodDetail } from '@store/api-client';
import { formatMoney } from '@store/shared-core';
import { api } from '../api';
import { BillingEntriesList } from '../components/InvoiceItems';
import { AuthGuestPrompt } from '../components/AuthGuestPrompt';
import { confirmCloseBillingCycle } from '../utils/confirmCloseBillingCycle';
import { guestAuthMessage, isGuestAuthError } from '../utils/authGuest';

function formatCompetence(year: number, month: number) {
  return `${String(month).padStart(2, '0')}/${year}`;
}

export function OpenBillingPeriodPage() {
  const navigate = useNavigate();
  const [data, setData] = useState<OpenBillingPeriodDetail | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [needsAuth, setNeedsAuth] = useState(false);
  const [loading, setLoading] = useState(true);
  const [closing, setClosing] = useState(false);

  useEffect(() => {
    setNeedsAuth(false);
    api
      .getMyOpenBillingPeriod()
      .then(setData)
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

  async function handleCloseCycle() {
    if (!data?.period) return;
    if (!confirmCloseBillingCycle(data.period.total_cents)) return;
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

  const period = data?.period;
  const canClose =
    period != null && period.entry_count > 0 && (period.total_cents ?? 0) > 0;

  return (
    <section className="content-section invoices-page">
      <p>
        <Link to="/faturas">← Minhas faturas</Link>
      </p>
      <h1>Competência atual</h1>
      {loading && <p>Carregando…</p>}
      {needsAuth && <AuthGuestPrompt message={guestAuthMessage('invoices')} />}
      {error && <p className="error">{error}</p>}
      {!loading && !needsAuth && period && (
        <>
          <div className="invoice-card invoice-card--current">
            <p>
              <strong>{formatCompetence(period.reference_year, period.reference_month)}</strong>
              {period.cycle_number != null && period.cycle_number > 1 && (
                <> · ciclo {period.cycle_number}</>
              )}
              {' — '}
              <span className="badge">Em acumulação</span>
            </p>
            <p>Total parcial: {formatMoney(period.total_cents)}</p>
            <p className="invoice-card-meta">
              {period.entry_count === 0
                ? 'Nenhuma compra nesta competência ainda.'
                : `${period.entry_count} lançamento(s).`}
            </p>
          </div>
          <h2>Lançamentos</h2>
          <BillingEntriesList entries={data?.entries ?? []} />
          <div className="invoice-actions">
            <Link to="/faturas" className="invoice-action-btn invoice-action-btn--secondary">
              Voltar
            </Link>
            {canClose && (
              <button
                type="button"
                className="invoice-action-btn invoice-action-btn--primary"
                disabled={closing}
                onClick={() => void handleCloseCycle()}
              >
                {closing ? 'Fechando…' : 'Fechar fatura e pagar'}
              </button>
            )}
          </div>
        </>
      )}
      {!loading && !needsAuth && !period && !error && (
        <p className="invoice-card-meta">Nenhuma competência em aberto no momento.</p>
      )}
    </section>
  );
}
