import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import type { InstallmentPolicy, UpdateInstallmentPolicyBody } from '@store/api-client';
import { formatMoney } from '@store/shared-core';
import { api } from '../api';
import { usePermissions } from '../auth/usePermissions';

function centsToReais(cents: number) {
  return (cents / 100).toFixed(2).replace('.', ',');
}

function reaisToCents(value: string): number {
  const normalized = value.trim().replace(/\./g, '').replace(',', '.');
  const n = Number.parseFloat(normalized);
  if (!Number.isFinite(n) || n < 0) return 0;
  return Math.round(n * 100);
}

export function InstallmentPolicyPage() {
  const canEdit = usePermissions().includes('billing.installment_settings.write');
  const [policy, setPolicy] = useState<InstallmentPolicy | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState<UpdateInstallmentPolicyBody | null>(null);

  useEffect(() => {
    api
      .adminGetInstallmentPolicy()
      .then((p) => {
        setPolicy(p);
        setForm({
          installment_enabled: p.installment_enabled,
          minimum_invoice_amount_cents: p.minimum_invoice_amount_cents,
          minimum_installment_amount_cents: p.minimum_installment_amount_cents,
          maximum_installments: p.maximum_installments,
          installment_interval_months: p.installment_interval_months,
          allow_installment_after_due_date: p.allow_installment_after_due_date,
          allow_early_installment_payment: p.allow_early_installment_payment,
          require_sequential_payment: p.require_sequential_payment,
          adjust_due_date_to_business_day: p.adjust_due_date_to_business_day,
        });
      })
      .catch((e: Error) => setError(e.message));
  }, []);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!form || !canEdit) return;
    setSaving(true);
    setError(null);
    try {
      const updated = await api.adminUpdateInstallmentPolicy(form);
      setPolicy(updated);
      setForm({
        installment_enabled: updated.installment_enabled,
        minimum_invoice_amount_cents: updated.minimum_invoice_amount_cents,
        minimum_installment_amount_cents: updated.minimum_installment_amount_cents,
        maximum_installments: updated.maximum_installments,
        installment_interval_months: updated.installment_interval_months,
        allow_installment_after_due_date: updated.allow_installment_after_due_date,
        allow_early_installment_payment: updated.allow_early_installment_payment,
        require_sequential_payment: updated.require_sequential_payment,
        adjust_due_date_to_business_day: updated.adjust_due_date_to_business_day,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Erro ao salvar');
    } finally {
      setSaving(false);
    }
  }

  if (!form) {
    return (
      <section className="content-section">
        <p>Carregando política de parcelamento…</p>
      </section>
    );
  }

  return (
    <section className="content-section billing-page">
      <p>
        <Link to="/faturamento">← Faturamento</Link>
      </p>
      <h1>Parcelamento de faturas</h1>
      {policy && (
        <p className="invoice-card-meta">
          Versão ativa: {policy.version}
          {policy.installment_enabled ? '' : ' — parcelamento desabilitado para novos planos'}
        </p>
      )}
      {error && <p className="error">{error}</p>}
      <form className="form form--wide" onSubmit={(e) => void onSubmit(e)}>
        <label className="form__full">
          <input
            type="checkbox"
            checked={form.installment_enabled}
            disabled={!canEdit}
            onChange={(e) => setForm({ ...form, installment_enabled: e.target.checked })}
          />{' '}
          Parcelamento habilitado para clientes (planos ativos em andamento não são alterados)
        </label>
        <label>
          Mínimo da fatura (R$)
          <input
            value={centsToReais(form.minimum_invoice_amount_cents)}
            disabled={!canEdit}
            onChange={(e) =>
              setForm({ ...form, minimum_invoice_amount_cents: reaisToCents(e.target.value) })
            }
          />
        </label>
        <label>
          Mínimo por parcela (R$)
          <input
            value={centsToReais(form.minimum_installment_amount_cents)}
            disabled={!canEdit}
            onChange={(e) =>
              setForm({ ...form, minimum_installment_amount_cents: reaisToCents(e.target.value) })
            }
          />
        </label>
        <label>
          Máximo de parcelas
          <input
            type="number"
            min={1}
            max={24}
            value={form.maximum_installments}
            disabled={!canEdit}
            onChange={(e) =>
              setForm({ ...form, maximum_installments: Number.parseInt(e.target.value, 10) || 1 })
            }
          />
        </label>
        <label>
          Intervalo (meses)
          <input
            type="number"
            min={1}
            value={form.installment_interval_months}
            disabled={!canEdit}
            onChange={(e) =>
              setForm({
                ...form,
                installment_interval_months: Number.parseInt(e.target.value, 10) || 1,
              })
            }
          />
        </label>
        <label className="form__full">
          <input
            type="checkbox"
            checked={form.allow_installment_after_due_date}
            disabled={!canEdit}
            onChange={(e) => setForm({ ...form, allow_installment_after_due_date: e.target.checked })}
          />{' '}
          Permitir escolha de parcelas após o vencimento da fatura
        </label>
        <label className="form__full">
          <input
            type="checkbox"
            checked={form.require_sequential_payment}
            disabled={!canEdit}
            onChange={(e) => setForm({ ...form, require_sequential_payment: e.target.checked })}
          />{' '}
          Exigir pagamento sequencial das parcelas
        </label>
        <label className="form__full">
          <input
            type="checkbox"
            checked={form.adjust_due_date_to_business_day}
            disabled={!canEdit}
            onChange={(e) =>
              setForm({ ...form, adjust_due_date_to_business_day: e.target.checked })
            }
          />{' '}
          Ajustar vencimentos para o próximo dia útil
        </label>
        {canEdit && (
          <p className="form__full">
            Referência: fatura mínima {formatMoney(form.minimum_invoice_amount_cents)} · parcela mínima{' '}
            {formatMoney(form.minimum_installment_amount_cents)}
          </p>
        )}
        {canEdit && (
          <button type="submit" disabled={saving}>
            {saving ? 'Salvando…' : 'Salvar nova versão da política'}
          </button>
        )}
      </form>
    </section>
  );
}
