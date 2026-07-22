SET search_path TO public;

CREATE TABLE public.invoice_installments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_plan_id uuid NOT NULL REFERENCES public.invoice_payment_plans(id) ON DELETE CASCADE,
    invoice_id uuid NOT NULL REFERENCES public.invoices(id) ON DELETE CASCADE,
    installment_number integer NOT NULL,
    amount_cents bigint NOT NULL,
    paid_cents bigint NOT NULL DEFAULT 0,
    remaining_cents bigint NOT NULL,
    due_date date NOT NULL,
    status character varying(32) NOT NULL DEFAULT 'scheduled',
    opened_at timestamptz,
    paid_at timestamptz,
    overdue_at timestamptz,
    canceled_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT invoice_installments_plan_number_unique UNIQUE (payment_plan_id, installment_number),
    CONSTRAINT invoice_installments_number_check CHECK (installment_number >= 1),
    CONSTRAINT invoice_installments_amount_check CHECK (amount_cents > 0),
    CONSTRAINT invoice_installments_paid_check CHECK (paid_cents >= 0 AND paid_cents <= amount_cents),
    CONSTRAINT invoice_installments_remaining_check CHECK (remaining_cents >= 0)
);

CREATE INDEX idx_invoice_installments_invoice ON public.invoice_installments (invoice_id);
CREATE INDEX idx_invoice_installments_plan ON public.invoice_installments (payment_plan_id);
CREATE INDEX idx_invoice_installments_status ON public.invoice_installments (status);
CREATE INDEX idx_invoice_installments_due ON public.invoice_installments (due_date);
