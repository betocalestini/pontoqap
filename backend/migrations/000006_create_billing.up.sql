CREATE TABLE business_calendar (
    date DATE PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    scope VARCHAR(32) NOT NULL DEFAULT 'national',
    is_business_day BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users (id)
);

CREATE TABLE billing_periods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers (id),
    reference_year INTEGER NOT NULL,
    reference_month INTEGER NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'open',
    opened_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT billing_periods_customer_ref_unique UNIQUE (customer_id, reference_year, reference_month),
    CONSTRAINT billing_periods_month_check CHECK (reference_month BETWEEN 1 AND 12)
);

CREATE TABLE billing_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    billing_period_id UUID NOT NULL REFERENCES billing_periods (id),
    entry_type VARCHAR(32) NOT NULL,
    order_id UUID REFERENCES orders (id),
    order_return_id UUID REFERENCES order_returns (id),
    description TEXT NOT NULL,
    amount_cents BIGINT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_number VARCHAR(32) NOT NULL,
    customer_id UUID NOT NULL REFERENCES customers (id),
    billing_period_id UUID NOT NULL REFERENCES billing_periods (id),
    status VARCHAR(32) NOT NULL DEFAULT 'open',
    subtotal_cents BIGINT NOT NULL DEFAULT 0,
    credit_cents BIGINT NOT NULL DEFAULT 0,
    adjustment_cents BIGINT NOT NULL DEFAULT 0,
    total_cents BIGINT NOT NULL DEFAULT 0,
    paid_cents BIGINT NOT NULL DEFAULT 0,
    due_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT invoices_number_unique UNIQUE (invoice_number),
    CONSTRAINT invoices_billing_period_unique UNIQUE (billing_period_id),
    CONSTRAINT invoices_paid_check CHECK (paid_cents >= 0)
);

CREATE TABLE invoice_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices (id) ON DELETE CASCADE,
    billing_entry_id UUID REFERENCES billing_entries (id),
    description TEXT NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price_cents BIGINT NOT NULL,
    total_cents BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE billing_adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices (id),
    adjustment_type VARCHAR(32) NOT NULL,
    amount_cents BIGINT NOT NULL,
    reason TEXT NOT NULL,
    created_by UUID NOT NULL REFERENCES users (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_billing_period_customer_reference ON billing_periods (customer_id, reference_year, reference_month);
CREATE INDEX idx_invoices_customer_status ON invoices (customer_id, status);
CREATE INDEX idx_invoices_due_status ON invoices (due_at, status);
