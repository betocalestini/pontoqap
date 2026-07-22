SET search_path TO public;

CREATE TABLE public.invoice_payment_plans (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id uuid NOT NULL REFERENCES public.invoices(id) ON DELETE CASCADE,
    policy_id uuid NOT NULL REFERENCES public.installment_policies(id),
    status character varying(32) NOT NULL DEFAULT 'pending_selection',
    selected_installment_count integer,
    invoice_total_cents bigint NOT NULL,
    paid_cents bigint NOT NULL DEFAULT 0,
    remaining_cents bigint NOT NULL,
    minimum_invoice_amount_cents_snapshot bigint NOT NULL,
    minimum_installment_amount_cents_snapshot bigint NOT NULL,
    maximum_installments_snapshot integer NOT NULL,
    installment_interval_months_snapshot integer NOT NULL,
    installment_enabled_snapshot boolean NOT NULL DEFAULT true,
    allow_early_payment_snapshot boolean NOT NULL DEFAULT false,
    require_sequential_payment_snapshot boolean NOT NULL DEFAULT true,
    adjust_business_day_snapshot boolean NOT NULL DEFAULT true,
    selected_by_user_id uuid REFERENCES public.users(id),
    selected_at timestamptz,
    canceled_by_user_id uuid REFERENCES public.users(id),
    canceled_at timestamptz,
    cancellation_reason text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT invoice_payment_plans_invoice_unique UNIQUE (invoice_id)
);

CREATE INDEX idx_invoice_payment_plans_status ON public.invoice_payment_plans (status);

-- Backfill: faturas abertas com saldo sem plano
INSERT INTO public.invoice_payment_plans (
    invoice_id, policy_id, status,
    invoice_total_cents, remaining_cents,
    minimum_invoice_amount_cents_snapshot, minimum_installment_amount_cents_snapshot,
    maximum_installments_snapshot, installment_interval_months_snapshot,
    installment_enabled_snapshot, allow_early_payment_snapshot,
    require_sequential_payment_snapshot, adjust_business_day_snapshot
)
SELECT
    i.id,
    p.id,
    'pending_selection',
    i.total_cents,
    i.total_cents - i.paid_cents,
    p.minimum_invoice_amount_cents,
    p.minimum_installment_amount_cents,
    p.maximum_installments,
    p.installment_interval_months,
    p.installment_enabled,
    p.allow_early_installment_payment,
    p.require_sequential_payment,
    p.adjust_due_date_to_business_day
FROM public.invoices i
CROSS JOIN public.installment_policies p
WHERE p.active = true
  AND i.status IN ('open', 'overdue')
  AND i.total_cents > i.paid_cents
  AND NOT EXISTS (SELECT 1 FROM public.invoice_payment_plans pp WHERE pp.invoice_id = i.id);
