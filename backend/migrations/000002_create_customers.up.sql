CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    document VARCHAR(32),
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    credit_limit_cents BIGINT NOT NULL DEFAULT 0,
    current_exposure_cents BIGINT NOT NULL DEFAULT 0,
    approved_by UUID REFERENCES users (id),
    approved_at TIMESTAMPTZ,
    blocked_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT customers_user_id_unique UNIQUE (user_id),
    CONSTRAINT customers_credit_limit_check CHECK (credit_limit_cents >= 0),
    CONSTRAINT customers_exposure_check CHECK (current_exposure_cents >= 0),
    CONSTRAINT customers_status_check CHECK (
        status IN ('pending', 'approved', 'rejected', 'blocked')
    )
);

CREATE TABLE customer_limit_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers (id) ON DELETE CASCADE,
    previous_limit_cents BIGINT NOT NULL,
    new_limit_cents BIGINT NOT NULL,
    reason TEXT,
    changed_by UUID NOT NULL REFERENCES users (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_customers_status ON customers (status);
