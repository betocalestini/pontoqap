SET search_path TO public;

ALTER TABLE public.payment_charges
    ADD COLUMN IF NOT EXISTS last_synced_at timestamp with time zone;

CREATE UNIQUE INDEX IF NOT EXISTS payments_provider_external_payment_id_unique
    ON public.payments (provider, external_payment_id)
    WHERE external_payment_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS payment_charges_one_pending_per_installment
    ON public.payment_charges (installment_id)
    WHERE installment_id IS NOT NULL AND status = 'pending';
