import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import type { MyInvoiceDetail } from '@store/api-client';
import { formatMoney, labelInvoiceStatus } from '@store/shared-core';
import { api } from '../api';
import { InvoiceItemsList } from '../components/InvoiceItems';

type Charge = { id: string; qr_code_text: string; amount_cents: number };

function formatCompetence(year: number, month: number) {
  return `${String(month).padStart(2, '0')}/${year}`;
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
  const [inv, setInv] = useState<MyInvoiceDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [charge, setCharge] = useState<Charge | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    api
      .getMyInvoice(id)
      .then(setInv)
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false));
  }, [id]);

  async function payPix() {
    if (!id) return;
    setError(null);
    try {
      const c = (await api.createPixCharge(id)) as Charge;
      setCharge(c);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Erro');
    }
  }

  async function simulate() {
    if (!charge || !id) return;
    try {
      await api.simulatePixPayment(charge.id);
      setInv(await api.getMyInvoice(id));
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
      {loading && <p>Carregando…</p>}
      {error && <p className="error">{error}</p>}
      {!loading && inv && (
        <div className="invoice-card">
          <p>
            <span className="badge">{labelInvoiceStatus(inv.status)}</span>
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
          <InvoiceItemsList items={inv.items} />
        </div>
      )}
      {!loading && inv && inv.remaining_cents > 0 && inv.status !== 'paid' && (
        <div className="stack-sm stack-sm--row">
          <button type="button" onClick={() => void payPix()}>
            Gerar Pix
          </button>
        </div>
      )}
      {charge && (
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
