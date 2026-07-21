SET search_path TO public;

ALTER TABLE public.billing_periods
    ADD COLUMN IF NOT EXISTS cycle_number integer NOT NULL DEFAULT 1;

ALTER TABLE public.billing_periods
    DROP CONSTRAINT IF EXISTS billing_periods_customer_ref_unique;

ALTER TABLE public.billing_periods
    ADD CONSTRAINT billing_periods_customer_ref_cycle_unique
    UNIQUE (customer_id, reference_year, reference_month, cycle_number);

CREATE UNIQUE INDEX IF NOT EXISTS billing_periods_one_open_per_customer
    ON public.billing_periods (customer_id)
    WHERE ((status)::text = 'open'::text);

ALTER TABLE public.invoices
    ADD COLUMN IF NOT EXISTS close_type character varying(32) NOT NULL DEFAULT 'legacy';

ALTER TABLE public.invoices
    ADD COLUMN IF NOT EXISTS reminder_sent_at timestamp with time zone;

ALTER TABLE public.invoices
    ADD COLUMN IF NOT EXISTS escalation_sent_at timestamp with time zone;
