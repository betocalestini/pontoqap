SET search_path TO public;

DROP INDEX IF EXISTS billing_periods_one_open_per_customer;

ALTER TABLE public.invoices
    DROP COLUMN IF EXISTS escalation_sent_at,
    DROP COLUMN IF EXISTS reminder_sent_at,
    DROP COLUMN IF EXISTS close_type;

ALTER TABLE public.billing_periods
    DROP CONSTRAINT IF EXISTS billing_periods_customer_ref_cycle_unique;

ALTER TABLE public.billing_periods
    DROP COLUMN IF EXISTS cycle_number;

ALTER TABLE public.billing_periods
    ADD CONSTRAINT billing_periods_customer_ref_unique UNIQUE (customer_id, reference_year, reference_month);
