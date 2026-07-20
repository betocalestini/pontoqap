CREATE TABLE payment_charges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices (id),
    provider VARCHAR(64) NOT NULL,
    external_id VARCHAR(255),
    txid VARCHAR(255),
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    amount_cents BIGINT NOT NULL,
    qr_code_text TEXT,
    qr_code_image_key VARCHAR(512),
    expires_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT payment_charges_provider_external_unique UNIQUE (provider, external_id),
    CONSTRAINT payment_charges_provider_txid_unique UNIQUE (provider, txid)
);

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices (id),
    payment_charge_id UUID REFERENCES payment_charges (id),
    provider VARCHAR(64) NOT NULL,
    external_payment_id VARCHAR(255),
    amount_cents BIGINT NOT NULL,
    status VARCHAR(32) NOT NULL,
    settled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE payment_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(64) NOT NULL,
    external_event_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    payload_hash VARCHAR(128) NOT NULL,
    payload_encrypted TEXT,
    processed BOOLEAN NOT NULL DEFAULT FALSE,
    processed_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT payment_events_provider_event_unique UNIQUE (provider, external_event_id)
);

CREATE INDEX idx_payment_charges_invoice_status ON payment_charges (invoice_id, status);
CREATE INDEX idx_payment_events_processed_created ON payment_events (processed, created_at);
