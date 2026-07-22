import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import type { AdminInvoiceDetail } from '@store/api-client';
import { formatMoney, labelInvoiceStatus } from '@store/shared-core';
import { api } from '../api';
import { usePermissions } from '../auth/usePermissions';

function formatCompetence(year: number, month: number) {
  return `${String(month).padStart(2, '0')}/${year}`;
}

function parseReaisToCents(value: string): number | null {
  const normalized = value.trim().replace(/\./g, '').replace(',', '.');
  if (!normalized) return null;
  const n = Number.parseFloat(normalized);
  if (!Number.isFinite(n) || n <= 0) return null;
  return Math.round(n * 100);
}

export function BillingInvoiceDetailPage() {
  const { id } = useParams();
  const canAdjust = usePermissions().includes('billing.close');
  const [inv, setInv] = useState<AdminInvoiceDetail | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [adjType, setAdjType] = useState<'credit' | 'debit'>('credit');
  const [adjAmount, setAdjAmount] = useState('');
  const [adjReason, setAdjReason] = useState('');
  const [saving, setSaving] = useState(false);

  async function load() {
    if (!id) return;
    setInv(await api.adminGetInvoice(id));
  }

  useEffect(() => {
    load().catch((e: Error) => setError(e.message));
  }, [id]);

  async function submitAdjustment(e: React.FormEvent) {
    e.preventDefault();
    if (!id || !canAdjust) return;
    const cents = parseReaisToCents(adjAmount);
    if (cents == null) {
      setError('Informe um valor válido em R$');
      return;
    }
    if (!adjReason.trim()) {
      setError('Informe a justificativa');
      return;
    }
    setSaving(true);
    setError(null);
    try {
      const updated = await api.adminAddInvoiceAdjustment(id, {
        adjustment_type: adjType,
        amount_cents: cents,
        reason: adjReason.trim(),
      });
      setInv(updated);
      setAdjAmount('');
      setAdjReason('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao ajustar');
    } finally {
      setSaving(false);
    }
  }

  if (!id) {
    return null;
  }

  return (
    <section className="content-section billing-page billing-page--detail">
      <p>
        <Link to="/faturamento">← Faturamento</Link>
      </p>
      <h1>{inv?.invoice_number ?? 'Fatura'}</h1>
      {error && <p className="error">{error}</p>}
      {inv && (
        <>
          <div className="billing-detail-header">
            <p>
              <strong>{inv.customer_name}</strong> ({inv.customer_email})
            </p>
            <p>
              Competência {formatCompetence(inv.reference_year, inv.reference_month)}
              {' · '}
              <span className={`billing-status billing-status--${inv.status}`}>
                {labelInvoiceStatus(inv.status)}
              </span>
            </p>
            {inv.due_at && (
              <p className="form-hint">
                Vencimento: {new Date(inv.due_at).toLocaleDateString('pt-BR')}
              </p>
            )}
          </div>

          <dl className="billing-detail-totals">
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
          <div className="table-scroll">
            <table className="billing-table">
              <thead>
                <tr>
                  <th>Descrição</th>
                  <th>Qtd</th>
                  <th>Unitário</th>
                  <th>Total</th>
                </tr>
              </thead>
              <tbody>
                {inv.items.map((it) => (
                  <tr key={it.id}>
                    <td>{it.description}</td>
                    <td>{it.quantity}</td>
                    <td>{formatMoney(it.unit_price_cents)}</td>
                    <td>{formatMoney(it.total_cents)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <h2>Ajustes lançados</h2>
          {inv.adjustments.length === 0 ? (
            <p className="form-hint">Nenhum ajuste manual nesta fatura.</p>
          ) : (
            <ul className="billing-adjust-list">
              {inv.adjustments.map((a) => (
                <li key={a.id}>
                  <strong>{a.adjustment_type === 'credit' ? 'Crédito' : 'Débito'}</strong>
                  {' — '}
                  {formatMoney(a.amount_cents)}
                  {' — '}
                  {a.reason}
                  <span className="billing-adjust-list__date">
                    {new Date(a.created_at).toLocaleString('pt-BR')}
                  </span>
                </li>
              ))}
            </ul>
          )}

          {canAdjust && inv.status !== 'paid' && (
            <div className="billing-panel">
              <h2 className="billing-panel__title">Novo ajuste</h2>
              <form className="billing-adjust-form" onSubmit={(e) => void submitAdjustment(e)}>
                <label>
                  Tipo
                  <select
                    value={adjType}
                    onChange={(e) => setAdjType(e.target.value as 'credit' | 'debit')}
                  >
                    <option value="credit">Crédito (reduz total)</option>
                    <option value="debit">Débito (aumenta total)</option>
                  </select>
                </label>
                <label>
                  Valor (R$)
                  <input
                    inputMode="decimal"
                    value={adjAmount}
                    onChange={(e) => setAdjAmount(e.target.value)}
                    required
                  />
                </label>
                <label className="billing-adjust-form__reason">
                  Justificativa
                  <input value={adjReason} onChange={(e) => setAdjReason(e.target.value)} required />
                </label>
                <button type="submit" disabled={saving}>
                  {saving ? 'Salvando…' : 'Lançar ajuste'}
                </button>
              </form>
            </div>
          )}
        </>
      )}
    </section>
  );
}
