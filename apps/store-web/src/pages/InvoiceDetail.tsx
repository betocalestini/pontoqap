import { useCallback, useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import type {
  InvoiceInstallment,
  MyInvoiceDetail,
  PaymentOption,
  PaymentPlan,
} from '@store/api-client';
import { formatMoney, labelInvoiceStatus } from '@store/shared-core';
import { useDialog } from '@store/ui';
import { api } from '../api';
import { InvoiceItemsList } from '../components/InvoiceItems';
import { PaymentPlanConfirmMessage } from '../components/PaymentPlanConfirmMessage';
import { PixPaymentBlock, type PixChargeView } from '../components/PixPaymentBlock';

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

function installmentStatusLabel(status: string) {
  const map: Record<string, string> = {
    scheduled: 'Agendada',
    open: 'Aberta',
    pix_active: 'Pix ativo',
    paid: 'Paga',
    overdue: 'Vencida',
    canceled: 'Cancelada',
  };
  return map[status] ?? status;
}

function optionLabel(opt: PaymentOption) {
  const first = opt.installments[0]?.amount_cents ?? 0;
  if (opt.installment_count === 1) {
    return `À vista — ${formatMoney(first)}`;
  }
  return `${opt.installment_count}× — a partir de ${formatMoney(first)}`;
}

export function InvoiceDetailPage() {
  const { id } = useParams();
  const { confirm } = useDialog();
  const [inv, setInv] = useState<MyInvoiceDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [charge, setCharge] = useState<PixChargeView | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [pixLoading, setPixLoading] = useState(false);
  const [options, setOptions] = useState<PaymentOption[]>([]);
  const [plan, setPlan] = useState<PaymentPlan | null>(null);
  const [installments, setInstallments] = useState<InvoiceInstallment[]>([]);
  const [selectedCount, setSelectedCount] = useState(1);
  const [confirmingPlan, setConfirmingPlan] = useState(false);

  const reload = useCallback(async () => {
    if (!id) return;
    const detail = await api.getMyInvoice(id);
    setInv(detail);
    try {
      const planRes = await api.getMyPaymentPlan(id);
      setPlan(planRes.data);
      if (planRes.data.status === 'active' || planRes.data.status === 'completed') {
        const instRes = await api.listMyInstallments(id);
        setInstallments(instRes.data ?? []);
        setOptions([]);
      } else if (planRes.data.status === 'pending_selection' && detail.remaining_cents > 0) {
        const optRes = await api.getMyPaymentOptions(id);
        setOptions(optRes.data.options ?? []);
        setSelectedCount(optRes.data.options?.[0]?.installment_count ?? 1);
        setInstallments([]);
      } else {
        setInstallments([]);
        setOptions([]);
      }
    } catch {
      setPlan(null);
      setInstallments([]);
      setOptions([]);
    }
  }, [id]);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    reload()
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false));
  }, [id, reload]);

  useEffect(() => {
    if (!id || loading || charge) return;
    const planOk = plan?.status === 'active' || plan?.status === 'completed';
    if (!planOk) return;
    const open = installments.find((i) => i.status === 'open' || i.status === 'pix_active');
    if (!open || open.status !== 'pix_active') return;
    api
      .getInstallmentPixCharge(open.id)
      .then((c) => setCharge(c))
      .catch(() => {});
  }, [id, loading, plan, installments, charge]);

  useEffect(() => {
    if (!id || !charge) return;
    if (inv && inv.remaining_cents <= 0) {
      setCharge(null);
      return;
    }
    const t = window.setInterval(() => {
      reload().catch(() => {});
    }, 30_000);
    return () => window.clearInterval(t);
  }, [id, charge, inv?.remaining_cents, reload]);

  useEffect(() => {
    if (!charge || !installments.length) return;
    const open = installments.find((i) => i.status === 'open' || i.status === 'pix_active');
    if (!open) setCharge(null);
  }, [installments, charge]);

  async function confirmPlan() {
    if (!id || !inv) return;
    const selectedOption = options.find((o) => o.installment_count === selectedCount);
    const ok = await confirm({
      title: 'Fechar fatura',
      message: (
        <PaymentPlanConfirmMessage remainingCents={inv.remaining_cents} option={selectedOption} />
      ),
      confirmLabel: 'Fechar fatura e pagar',
    });
    if (!ok) return;

    setError(null);
    setConfirmingPlan(true);
    try {
      await api.selectMyPaymentPlan(id, selectedCount);
      setCharge(null);
      await reload();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Erro');
    } finally {
      setConfirmingPlan(false);
    }
  }

  async function payInstallmentPix(installmentId: string) {
    setError(null);
    setPixLoading(true);
    try {
      const c = await api.createInstallmentPixCharge(installmentId);
      setCharge(c);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Erro');
    } finally {
      setPixLoading(false);
    }
  }

  async function simulate() {
    if (!charge || !id) return;
    try {
      await api.simulatePixPayment(charge.id);
      setCharge(null);
      await reload();
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Erro');
    }
  }

  const selectedOption = options.find((o) => o.installment_count === selectedCount);
  const showPlanSelection = plan?.status === 'pending_selection' && inv && inv.remaining_cents > 0;
  const planIsActive = plan?.status === 'active' || plan?.status === 'completed';
  const openInstallment = installments.find((i) => i.status === 'open' || i.status === 'pix_active');
  const showPixBlock = charge && planIsActive;

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

      {showPlanSelection && (
        <div className="invoice-card payment-plan-card">
          <h2>Forma de pagamento</h2>
          <p className="invoice-card-meta">
            Escolha em quantas parcelas deseja pagar. Na sequência você confirma o plano para liberar o
            pagamento.
          </p>
          <div className="payment-plan-options" role="radiogroup" aria-label="Forma de pagamento">
            {options.map((opt) => {
              const selected = selectedCount === opt.installment_count;
              return (
                <label
                  key={opt.installment_count}
                  className={`payment-plan-option${selected ? ' payment-plan-option--selected' : ''}`}
                >
                  <input
                    type="radio"
                    name="installment_count"
                    className="payment-plan-option__input"
                    checked={selected}
                    onChange={() => setSelectedCount(opt.installment_count)}
                  />
                  <span className="payment-plan-option__label">{optionLabel(opt)}</span>
                </label>
              );
            })}
          </div>
          {selectedOption && selectedOption.installments.length > 0 && (
            <ul className="payment-plan-preview">
              {selectedOption.installments.map((inst) => (
                <li key={inst.number}>
                  <span>
                    Parcela {inst.number}/{selectedOption.installments.length}
                  </span>
                  <span>{formatMoney(inst.amount_cents)}</span>
                  <span>{new Date(inst.due_date).toLocaleDateString('pt-BR')}</span>
                </li>
              ))}
            </ul>
          )}
          <div className="invoice-actions invoice-actions--stack-mobile">
            <button
              type="button"
              className="invoice-action-btn invoice-action-btn--primary invoice-action-btn--block"
              disabled={confirmingPlan}
              onClick={() => void confirmPlan()}
            >
              {confirmingPlan ? 'Confirmando…' : 'Confirmar forma de pagamento'}
            </button>
          </div>
        </div>
      )}

      {installments.length > 0 && (
        <div className="invoice-card">
          <h2>Parcelas</h2>
          <div className="installment-table-wrap table-scroll">
            <table className="installment-table">
              <thead>
                <tr>
                  <th>Parcela</th>
                  <th>Vencimento</th>
                  <th>Valor</th>
                  <th>Estado</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {installments.map((inst) => (
                  <tr key={inst.id}>
                    <td>
                      {inst.installment_number}/{installments.length}
                    </td>
                    <td>{new Date(inst.due_date).toLocaleDateString('pt-BR')}</td>
                    <td>{formatMoney(inst.amount_cents)}</td>
                    <td>{installmentStatusLabel(inst.status)}</td>
                    <td>
                      {(inst.status === 'open' || inst.status === 'pix_active') &&
                        openInstallment?.id === inst.id && (
                          <button
                            type="button"
                            className="invoice-action-btn invoice-action-btn--primary invoice-action-btn--compact"
                            disabled={pixLoading}
                            onClick={() => void payInstallmentPix(inst.id)}
                          >
                            {pixLoading ? 'Gerando Pix…' : 'Gerar Pix'}
                          </button>
                        )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <ul className="installment-cards">
            {installments.map((inst) => (
              <li key={inst.id} className="installment-card">
                <div className="installment-card__row">
                  <span className="installment-card__title">
                    Parcela {inst.installment_number}/{installments.length}
                  </span>
                  <span className="badge">{installmentStatusLabel(inst.status)}</span>
                </div>
                <div className="installment-card__meta">
                  <span>Venc. {new Date(inst.due_date).toLocaleDateString('pt-BR')}</span>
                  <span>{formatMoney(inst.amount_cents)}</span>
                </div>
                {(inst.status === 'open' || inst.status === 'pix_active') &&
                  openInstallment?.id === inst.id && (
                    <button
                      type="button"
                      className="invoice-action-btn invoice-action-btn--primary invoice-action-btn--block"
                      disabled={pixLoading}
                      onClick={() => void payInstallmentPix(inst.id)}
                    >
                      {pixLoading ? 'Gerando Pix…' : 'Gerar Pix'}
                    </button>
                  )}
              </li>
            ))}
          </ul>
        </div>
      )}

      {showPixBlock && (
        <PixPaymentBlock
          charge={charge}
          onClose={() => setCharge(null)}
          onSimulate={charge.simulatable ? () => void simulate() : undefined}
        />
      )}
    </section>
  );
}
