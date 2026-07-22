SET search_path TO public;

CREATE TABLE public.installment_policies (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    version integer NOT NULL,
    active boolean NOT NULL DEFAULT false,
    installment_enabled boolean NOT NULL DEFAULT true,
    minimum_invoice_amount_cents bigint NOT NULL,
    minimum_installment_amount_cents bigint NOT NULL,
    maximum_installments integer NOT NULL,
    installment_interval_months integer NOT NULL DEFAULT 1,
    allow_installment_after_due_date boolean NOT NULL DEFAULT false,
    allow_early_installment_payment boolean NOT NULL DEFAULT false,
    require_sequential_payment boolean NOT NULL DEFAULT true,
    adjust_due_date_to_business_day boolean NOT NULL DEFAULT true,
    valid_from timestamptz NOT NULL DEFAULT now(),
    valid_until timestamptz,
    created_by uuid REFERENCES public.users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT installment_policies_version_unique UNIQUE (version)
);

CREATE UNIQUE INDEX installment_policies_one_active ON public.installment_policies (active) WHERE active = true;

INSERT INTO public.installment_policies (
    id, version, active, installment_enabled,
    minimum_invoice_amount_cents, minimum_installment_amount_cents, maximum_installments,
    installment_interval_months, allow_installment_after_due_date, allow_early_installment_payment,
    require_sequential_payment, adjust_due_date_to_business_day, valid_from
) VALUES (
    'c0000000-0000-4000-8000-000000000001', 1, true, true,
    30000, 10000, 10,
    1, false, false,
    true, true, now()
);

INSERT INTO public.permissions VALUES
  ('b0000000-0000-4000-8000-00000000001d', 'billing.installment_settings.write', 'Configurar parcelamento de faturas'),
  ('b0000000-0000-4000-8000-00000000001e', 'billing.installments.override', 'Exceção: reset de plano de parcelas')
ON CONFLICT (id) DO NOTHING;

INSERT INTO public.role_permissions (role_id, permission_id) VALUES
  ('a0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-00000000001d'),
  ('a0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-00000000001e'),
  ('a0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-00000000001d'),
  ('a0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-00000000001e'),
  ('a0000000-0000-4000-8000-000000000005', 'b0000000-0000-4000-8000-00000000001d'),
  ('a0000000-0000-4000-8000-000000000005', 'b0000000-0000-4000-8000-00000000001e')
ON CONFLICT DO NOTHING;
