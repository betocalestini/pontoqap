SET search_path TO public;

ALTER TABLE public.payments
    ADD COLUMN installment_id uuid REFERENCES public.invoice_installments(id);

CREATE INDEX idx_payments_installment ON public.payments (installment_id)
    WHERE installment_id IS NOT NULL;
