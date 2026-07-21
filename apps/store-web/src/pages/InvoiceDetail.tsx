import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import type { AdminInvoiceDetail } from '@store/api-client';
import { formatMoney } from '@store/shared-core';
import { api } from '../api';
import { AuthGuestPrompt } from '../components/AuthGuestPrompt';
import { guestAuthMessage, isGuestAuthError } from '../utils/authGuest';

type Charge = { id: string; qr_code_text: string; amount_cents: number };

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

export function InvoiceDetailPage() {
  const { id } = useParams();
  const [inv, setInv] = useState<AdminInvoiceDetail | null>(null);
  const [charge, setCharge] = useState<Charge | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [needsAuth, setNeedsAuth] = useState(false);

  useEffect(() => {
    if (!id) return;
    setNeedsAuth(false);
    api
      .getMyInvoice(id)
      .then((v) => setInv(v as AdminInvoiceDetail))
      .catch((e: Error) => {
        if (isGuestAuthError(e)) {
          setNeedsAuth(true);
          setError(null);
        } else {
          setError(e.message);
        }
      });
  }, [id]);

  async function payPix() {
    if (!id) return;
    setError(null);
    try {
      const c = (await api.createPixCharge(id)) as Charge;
      setCharge(c);
    } catch (e) {
      if (isGuestAuthError(e)) {
        setNeedsAuth(true);
        setError(null);
      } else {
        setError(e instanceof Error ? e.message : 'Erro');
      }
    }
  }

  async function simulate() {
    if (!charge || !id) return;
    try {
      await api.simulatePixPayment(charge.id);
      setInv((await api.getMyInvoice(id)) as AdminInvoiceDetail);
      setCharge(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Erro');
    }
  }

  return (
    <section className="content-section invoices-page">
      <p>
        <Link to="/faturas">← Minhas faturas</Link>
      </p>
      <h1>{inv?.invoice_number ?? 'Fatura'}</h1>
      {needsAuth && <AuthGuestPrompt message={guestAuthMessage('invoice')} />}
      {error && <p className="error">{error}</p>}
      {!needsAuth && inv && (
        <div className="invoice-card">
          <p>
            <span className="badge">{invoiceStatusLabel(inv.status)}</span>
            {closeTypeLabel(inv.close_type) && (
              <>
                {' · '}
                <span className="invoice-card-meta">{closeTypeLabel(inv.close_type)}</span>
              </>
            )}
            {' · '}
            Competência {formatCompetence(inv.reference_year, inv.reference_month)}
            {inv.cycle_number != null && inv.cycle_number > 1 && ` · ciclo ${inv.cycle_number}`}
          </p>
          {inv.due_at && (
            <p className="invoice-card-meta">
              Vencimento: {new Date(inv.due_at).toLocaleDateString('pt-BR')}
            </p>
          )}
          <dl className="invoice-detail-dl">
            <div>
              <dt>Subtotal</dt>
              <dd>{formatMoney(inv.subtotal_cents)}</dd>
            </div>
            <div>
              <dt>Ajustes</dt>
              <dd>{formatMoney(inv.adjustment_cents)}</dd>
            </div>
            <div>
              <dt>Total</dt>
              <dd>{formatMoney(inv.total_cents)}</dd>
            </div>
            <div>
              <dt>Pago</dt>
              <dd>{formatMoney(inv.paid_cents)}</dd>
            </div>
            <div>
              <dt>Em aberto</dt>
              <dd>{formatMoney(inv.remaining_cents)}</dd>
            </div>
          </dl>
          <h2>Itens</h2>
          <ul className="invoice-item-list">
            {inv.items.map((it) => (
              <li key={it.id}>
                {it.description} — {it.quantity} × {formatMoney(it.unit_price_cents)} ={' '}
                {formatMoney(it.total_cents)}
              </li>
            ))}
          </ul>
        </div>
      )}
      {!needsAuth && inv && inv.remaining_cents > 0 && inv.status !== 'paid' && (
        <div className="stack-sm stack-sm--row">
          <button type="button" onClick={() => void payPix()}>
            Gerar Pix
          </button>
        </div>
      )}
      {!needsAuth && charge && (
        <div className="pix">
          <p>Valor: {formatMoney(charge.amount_cents)}</p>
          <code>{charge.qr_code_text}</code>
          <button type="button" onClick={() => void simulate()}>
            Simular pagamento (dev)
          </button>
        </div>
      )}
    </section>
  );
}
