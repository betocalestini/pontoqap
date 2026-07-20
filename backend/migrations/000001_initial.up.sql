-- Schema inicial consolidado do Store Platform (pré-produção).
-- Substitui as migrations incrementais 000001–000022.

--
-- PostgreSQL database dump
--


-- Dumped from database version 16.14
-- Dumped by pg_dump version 16.14

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.audit_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    actor_user_id uuid,
    action character varying(128) NOT NULL,
    entity_type character varying(64) NOT NULL,
    entity_id uuid,
    request_id uuid,
    old_values jsonb,
    new_values jsonb,
    ip_address inet,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: billing_adjustments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.billing_adjustments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    invoice_id uuid NOT NULL,
    adjustment_type character varying(32) NOT NULL,
    amount_cents bigint NOT NULL,
    reason text NOT NULL,
    created_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: billing_entries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.billing_entries (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    billing_period_id uuid NOT NULL,
    entry_type character varying(32) NOT NULL,
    order_id uuid,
    order_return_id uuid,
    description text NOT NULL,
    amount_cents bigint NOT NULL,
    occurred_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: billing_periods; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.billing_periods (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    customer_id uuid NOT NULL,
    reference_year integer NOT NULL,
    reference_month integer NOT NULL,
    status character varying(32) DEFAULT 'open'::character varying NOT NULL,
    opened_at timestamp with time zone DEFAULT now() NOT NULL,
    closed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT billing_periods_month_check CHECK (((reference_month >= 1) AND (reference_month <= 12)))
);


--
-- Name: business_calendar; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.business_calendar (
    date date NOT NULL,
    name character varying(255) NOT NULL,
    scope character varying(32) DEFAULT 'national'::character varying NOT NULL,
    is_business_day boolean NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by uuid
);


--
-- Name: cart_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cart_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    cart_id uuid NOT NULL,
    sku_id uuid NOT NULL,
    quantity integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT cart_items_quantity_check CHECK ((quantity > 0))
);


--
-- Name: carts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.carts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    customer_id uuid NOT NULL,
    status character varying(32) DEFAULT 'active'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT carts_status_check CHECK (((status)::text = ANY ((ARRAY['active'::character varying, 'checked_out'::character varying, 'abandoned'::character varying])::text[])))
);


--
-- Name: categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.categories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    slug character varying(255) NOT NULL,
    active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: collaborator_categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.collaborator_categories (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(128) NOT NULL,
    slug character varying(128) NOT NULL,
    margin_percent numeric(5,2) NOT NULL,
    active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT collaborator_categories_margin_check CHECK (((margin_percent >= (0)::numeric) AND (margin_percent <= (1000)::numeric)))
);


--
-- Name: customer_limit_history; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.customer_limit_history (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    customer_id uuid NOT NULL,
    previous_limit_cents bigint NOT NULL,
    new_limit_cents bigint NOT NULL,
    reason text,
    changed_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: customers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.customers (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    document character varying(32),
    status character varying(32) DEFAULT 'pending'::character varying NOT NULL,
    credit_limit_cents bigint DEFAULT 0 NOT NULL,
    current_exposure_cents bigint DEFAULT 0 NOT NULL,
    approved_by uuid,
    approved_at timestamp with time zone,
    blocked_reason text,
    status_before_block character varying(32),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    collaborator_category_id uuid,
    CONSTRAINT customers_credit_limit_check CHECK ((credit_limit_cents >= 0)),
    CONSTRAINT customers_exposure_check CHECK ((current_exposure_cents >= 0)),
    CONSTRAINT customers_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'approved'::character varying, 'rejected'::character varying, 'blocked'::character varying])::text[])))
);


--
-- Name: email_verification_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.email_verification_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    token_hash character varying(128) NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    used_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: forecast_snapshots; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.forecast_snapshots (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    sku_id uuid NOT NULL,
    reference_month date NOT NULL,
    forecast_quantity integer NOT NULL,
    safety_stock_quantity integer DEFAULT 0 NOT NULL,
    suggested_purchase_quantity integer DEFAULT 0 NOT NULL,
    confidence_level character varying(32),
    method character varying(64) NOT NULL,
    parameters jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: inventory_balances; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.inventory_balances (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    location_id uuid NOT NULL,
    sku_id uuid NOT NULL,
    available_quantity integer DEFAULT 0 NOT NULL,
    version integer DEFAULT 0 NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT inventory_balances_quantity_check CHECK ((available_quantity >= 0))
);


--
-- Name: inventory_locations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.inventory_locations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    active boolean DEFAULT true NOT NULL
);


--
-- Name: inventory_lots; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.inventory_lots (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    location_id uuid NOT NULL,
    sku_id uuid NOT NULL,
    quantity_remaining integer NOT NULL,
    unit_cost_cents bigint DEFAULT 0 NOT NULL,
    source_movement_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT inventory_lots_cost_check CHECK ((unit_cost_cents >= 0)),
    CONSTRAINT inventory_lots_qty_check CHECK ((quantity_remaining >= 0))
);


--
-- Name: invoice_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invoice_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    invoice_id uuid NOT NULL,
    billing_entry_id uuid,
    description text NOT NULL,
    quantity integer DEFAULT 1 NOT NULL,
    unit_price_cents bigint NOT NULL,
    total_cents bigint NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: invoices; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invoices (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    invoice_number character varying(32) NOT NULL,
    customer_id uuid NOT NULL,
    billing_period_id uuid NOT NULL,
    status character varying(32) DEFAULT 'open'::character varying NOT NULL,
    subtotal_cents bigint DEFAULT 0 NOT NULL,
    credit_cents bigint DEFAULT 0 NOT NULL,
    adjustment_cents bigint DEFAULT 0 NOT NULL,
    total_cents bigint DEFAULT 0 NOT NULL,
    paid_cents bigint DEFAULT 0 NOT NULL,
    due_at timestamp with time zone,
    closed_at timestamp with time zone,
    paid_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT invoices_paid_check CHECK ((paid_cents >= 0))
);


--
-- Name: jobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.jobs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    type character varying(128) NOT NULL,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    status character varying(32) DEFAULT 'pending'::character varying NOT NULL,
    attempts integer DEFAULT 0 NOT NULL,
    available_at timestamp with time zone DEFAULT now() NOT NULL,
    locked_at timestamp with time zone,
    locked_by character varying(128),
    last_error text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    completed_at timestamp with time zone
);


--
-- Name: order_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.order_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_id uuid NOT NULL,
    sku_id uuid NOT NULL,
    product_name_snapshot character varying(255) NOT NULL,
    sku_code_snapshot character varying(64) NOT NULL,
    unit_price_cents bigint NOT NULL,
    quantity integer NOT NULL,
    total_cents bigint NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT order_items_quantity_check CHECK ((quantity > 0)),
    CONSTRAINT order_items_total_check CHECK ((total_cents >= 0)),
    CONSTRAINT order_items_unit_price_check CHECK ((unit_price_cents >= 0))
);


--
-- Name: order_return_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.order_return_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_return_id uuid NOT NULL,
    order_item_id uuid NOT NULL,
    quantity integer NOT NULL,
    credit_amount_cents bigint NOT NULL,
    return_to_stock boolean DEFAULT true NOT NULL,
    CONSTRAINT order_return_items_quantity_check CHECK ((quantity > 0))
);


--
-- Name: order_returns; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.order_returns (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_id uuid NOT NULL,
    status character varying(32) DEFAULT 'pending'::character varying NOT NULL,
    reason text,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    completed_at timestamp with time zone
);


--
-- Name: orders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.orders (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    order_number character varying(32) NOT NULL,
    customer_id uuid NOT NULL,
    status character varying(32) DEFAULT 'confirmed'::character varying NOT NULL,
    subtotal_cents bigint NOT NULL,
    discount_cents bigint DEFAULT 0 NOT NULL,
    total_cents bigint NOT NULL,
    idempotency_key character varying(128) NOT NULL,
    confirmed_at timestamp with time zone,
    cancelled_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT orders_discount_check CHECK ((discount_cents >= 0)),
    CONSTRAINT orders_subtotal_check CHECK ((subtotal_cents >= 0)),
    CONSTRAINT orders_total_check CHECK ((total_cents >= 0))
);


--
-- Name: outbox_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.outbox_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    event_type character varying(128) NOT NULL,
    aggregate_type character varying(64) NOT NULL,
    aggregate_id uuid NOT NULL,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    status character varying(32) DEFAULT 'pending'::character varying NOT NULL,
    attempts integer DEFAULT 0 NOT NULL,
    available_at timestamp with time zone DEFAULT now() NOT NULL,
    processed_at timestamp with time zone,
    last_error text,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: password_reset_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.password_reset_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    token_hash character varying(128) NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    used_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: payment_charges; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.payment_charges (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    invoice_id uuid NOT NULL,
    provider character varying(64) NOT NULL,
    external_id character varying(255),
    txid character varying(255),
    status character varying(32) DEFAULT 'pending'::character varying NOT NULL,
    amount_cents bigint NOT NULL,
    qr_code_text text,
    qr_code_image_key character varying(512),
    expires_at timestamp with time zone,
    paid_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: payment_events; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.payment_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    provider character varying(64) NOT NULL,
    external_event_id character varying(255) NOT NULL,
    event_type character varying(64) NOT NULL,
    payload_hash character varying(128) NOT NULL,
    payload_encrypted text,
    processed boolean DEFAULT false NOT NULL,
    processed_at timestamp with time zone,
    error_message text,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: payments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.payments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    invoice_id uuid NOT NULL,
    payment_charge_id uuid,
    provider character varying(64) NOT NULL,
    external_payment_id character varying(255),
    amount_cents bigint NOT NULL,
    status character varying(32) NOT NULL,
    settled_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.permissions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code character varying(128) NOT NULL,
    name character varying(255) NOT NULL
);


--
-- Name: price_history; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.price_history (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    sku_id uuid NOT NULL,
    previous_price_cents bigint NOT NULL,
    new_price_cents bigint NOT NULL,
    changed_by uuid NOT NULL,
    reason text,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: product_images; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.product_images (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    product_id uuid NOT NULL,
    storage_key character varying(512) NOT NULL,
    "position" integer DEFAULT 0 NOT NULL,
    alt_text character varying(255),
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: products; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.products (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    category_id uuid,
    name character varying(255) NOT NULL,
    slug character varying(255) NOT NULL,
    description text,
    active boolean DEFAULT true NOT NULL,
    visible boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    margin_percent numeric(5,2) DEFAULT 30.00 NOT NULL,
    promo_active boolean DEFAULT false NOT NULL,
    promo_margin_percent numeric(5,2),
    promo_quantity_total integer DEFAULT 0 NOT NULL,
    promo_quantity_remaining integer DEFAULT 0 NOT NULL,
    CONSTRAINT products_promo_margin_when_active CHECK (((NOT promo_active) OR ((promo_margin_percent IS NOT NULL) AND (promo_quantity_total > 0)))),
    CONSTRAINT products_promo_remaining_lte_total CHECK ((promo_quantity_remaining <= promo_quantity_total)),
    CONSTRAINT products_promo_remaining_nonneg CHECK ((promo_quantity_remaining >= 0)),
    CONSTRAINT products_promo_total_nonneg CHECK ((promo_quantity_total >= 0))
);


--
-- Name: role_permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.role_permissions (
    role_id uuid NOT NULL,
    permission_id uuid NOT NULL
);


--
-- Name: roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.roles (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    code character varying(64) NOT NULL,
    name character varying(255) NOT NULL
);


--
-- Name: sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    token_hash character varying(128) NOT NULL,
    audience character varying(16) DEFAULT 'store'::character varying NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    revoked_at timestamp with time zone,
    ip_address inet,
    user_agent text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT sessions_audience_check CHECK (((audience)::text = ANY ((ARRAY['store'::character varying, 'admin'::character varying])::text[])))
);


--
-- Name: skus; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.skus (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    product_id uuid NOT NULL,
    code character varying(64) NOT NULL,
    barcode character varying(64),
    unit character varying(32) DEFAULT 'UN'::character varying NOT NULL,
    sale_price_cents bigint DEFAULT 0 NOT NULL,
    cost_price_cents bigint,
    minimum_stock integer DEFAULT 0 NOT NULL,
    active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT skus_cost_price_check CHECK (((cost_price_cents IS NULL) OR (cost_price_cents >= 0))),
    CONSTRAINT skus_minimum_stock_check CHECK ((minimum_stock >= 0)),
    CONSTRAINT skus_sale_price_check CHECK ((sale_price_cents >= 0))
);


--
-- Name: stock_movements; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.stock_movements (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    location_id uuid NOT NULL,
    sku_id uuid NOT NULL,
    movement_type character varying(32) NOT NULL,
    quantity integer NOT NULL,
    previous_balance integer NOT NULL,
    new_balance integer NOT NULL,
    reference_type character varying(64),
    reference_id uuid,
    reason text,
    created_by uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    unit_cost_cents bigint,
    CONSTRAINT stock_movements_quantity_check CHECK ((quantity > 0)),
    CONSTRAINT stock_movements_unit_cost_check CHECK (((unit_cost_cents IS NULL) OR (unit_cost_cents >= 0)))
);


--
-- Name: COLUMN stock_movements.unit_cost_cents; Type: COMMENT; Schema: public; Owner: -
--

COMMENT ON COLUMN public.stock_movements.unit_cost_cents IS 'Custo unitário de aquisição (centavos), preenchido em entradas manuais';


--
-- Name: store_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.store_settings (
    key character varying(128) NOT NULL,
    value text NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: user_roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_roles (
    user_id uuid NOT NULL,
    role_id uuid NOT NULL
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    phone character varying(50),
    password_hash character varying(255) NOT NULL,
    status character varying(32) DEFAULT 'active'::character varying NOT NULL,
    mfa_secret text,
    mfa_enabled boolean DEFAULT false NOT NULL,
    failed_login_attempts integer DEFAULT 0 NOT NULL,
    locked_until timestamp with time zone,
    last_login_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    email_verified_at timestamp with time zone,
    CONSTRAINT users_status_check CHECK (((status)::text = ANY ((ARRAY['active'::character varying, 'inactive'::character varying, 'blocked'::character varying, 'pending_email'::character varying])::text[])))
);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: billing_adjustments billing_adjustments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_adjustments
    ADD CONSTRAINT billing_adjustments_pkey PRIMARY KEY (id);


--
-- Name: billing_entries billing_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_entries
    ADD CONSTRAINT billing_entries_pkey PRIMARY KEY (id);


--
-- Name: billing_periods billing_periods_customer_ref_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_periods
    ADD CONSTRAINT billing_periods_customer_ref_unique UNIQUE (customer_id, reference_year, reference_month);


--
-- Name: billing_periods billing_periods_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_periods
    ADD CONSTRAINT billing_periods_pkey PRIMARY KEY (id);


--
-- Name: business_calendar business_calendar_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_calendar
    ADD CONSTRAINT business_calendar_pkey PRIMARY KEY (date);


--
-- Name: cart_items cart_items_cart_sku_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_cart_sku_unique UNIQUE (cart_id, sku_id);


--
-- Name: cart_items cart_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_pkey PRIMARY KEY (id);


--
-- Name: carts carts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.carts
    ADD CONSTRAINT carts_pkey PRIMARY KEY (id);


--
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (id);


--
-- Name: categories categories_slug_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_slug_unique UNIQUE (slug);


--
-- Name: collaborator_categories collaborator_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collaborator_categories
    ADD CONSTRAINT collaborator_categories_pkey PRIMARY KEY (id);


--
-- Name: collaborator_categories collaborator_categories_slug_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.collaborator_categories
    ADD CONSTRAINT collaborator_categories_slug_unique UNIQUE (slug);


--
-- Name: customer_limit_history customer_limit_history_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customer_limit_history
    ADD CONSTRAINT customer_limit_history_pkey PRIMARY KEY (id);


--
-- Name: customers customers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_pkey PRIMARY KEY (id);


--
-- Name: customers customers_user_id_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_user_id_unique UNIQUE (user_id);


--
-- Name: email_verification_tokens email_verification_tokens_hash_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_verification_tokens
    ADD CONSTRAINT email_verification_tokens_hash_unique UNIQUE (token_hash);


--
-- Name: email_verification_tokens email_verification_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_verification_tokens
    ADD CONSTRAINT email_verification_tokens_pkey PRIMARY KEY (id);


--
-- Name: forecast_snapshots forecast_snapshots_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.forecast_snapshots
    ADD CONSTRAINT forecast_snapshots_pkey PRIMARY KEY (id);


--
-- Name: inventory_balances inventory_balances_location_sku_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.inventory_balances
    ADD CONSTRAINT inventory_balances_location_sku_unique UNIQUE (location_id, sku_id);


--
-- Name: inventory_balances inventory_balances_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.inventory_balances
    ADD CONSTRAINT inventory_balances_pkey PRIMARY KEY (id);


--
-- Name: inventory_locations inventory_locations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.inventory_locations
    ADD CONSTRAINT inventory_locations_pkey PRIMARY KEY (id);


--
-- Name: inventory_lots inventory_lots_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.inventory_lots
    ADD CONSTRAINT inventory_lots_pkey PRIMARY KEY (id);


--
-- Name: invoice_items invoice_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_items
    ADD CONSTRAINT invoice_items_pkey PRIMARY KEY (id);


--
-- Name: invoices invoices_billing_period_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT invoices_billing_period_unique UNIQUE (billing_period_id);


--
-- Name: invoices invoices_number_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT invoices_number_unique UNIQUE (invoice_number);


--
-- Name: invoices invoices_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT invoices_pkey PRIMARY KEY (id);


--
-- Name: jobs jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jobs
    ADD CONSTRAINT jobs_pkey PRIMARY KEY (id);


--
-- Name: order_items order_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);


--
-- Name: order_return_items order_return_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_return_items
    ADD CONSTRAINT order_return_items_pkey PRIMARY KEY (id);


--
-- Name: order_returns order_returns_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_returns
    ADD CONSTRAINT order_returns_pkey PRIMARY KEY (id);


--
-- Name: orders orders_idempotency_key_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_idempotency_key_unique UNIQUE (idempotency_key);


--
-- Name: orders orders_order_number_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_order_number_unique UNIQUE (order_number);


--
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);


--
-- Name: outbox_events outbox_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.outbox_events
    ADD CONSTRAINT outbox_events_pkey PRIMARY KEY (id);


--
-- Name: password_reset_tokens password_reset_tokens_hash_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_hash_unique UNIQUE (token_hash);


--
-- Name: password_reset_tokens password_reset_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_pkey PRIMARY KEY (id);


--
-- Name: payment_charges payment_charges_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_charges
    ADD CONSTRAINT payment_charges_pkey PRIMARY KEY (id);


--
-- Name: payment_charges payment_charges_provider_external_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_charges
    ADD CONSTRAINT payment_charges_provider_external_unique UNIQUE (provider, external_id);


--
-- Name: payment_charges payment_charges_provider_txid_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_charges
    ADD CONSTRAINT payment_charges_provider_txid_unique UNIQUE (provider, txid);


--
-- Name: payment_events payment_events_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_events
    ADD CONSTRAINT payment_events_pkey PRIMARY KEY (id);


--
-- Name: payment_events payment_events_provider_event_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_events
    ADD CONSTRAINT payment_events_provider_event_unique UNIQUE (provider, external_event_id);


--
-- Name: payments payments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT payments_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_code_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_code_unique UNIQUE (code);


--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--
-- Name: price_history price_history_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.price_history
    ADD CONSTRAINT price_history_pkey PRIMARY KEY (id);


--
-- Name: product_images product_images_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_images
    ADD CONSTRAINT product_images_pkey PRIMARY KEY (id);


--
-- Name: products products_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);


--
-- Name: products products_slug_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_slug_unique UNIQUE (slug);


--
-- Name: role_permissions role_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (role_id, permission_id);


--
-- Name: roles roles_code_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_code_unique UNIQUE (code);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);


--
-- Name: sessions sessions_token_hash_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_token_hash_unique UNIQUE (token_hash);


--
-- Name: skus skus_barcode_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.skus
    ADD CONSTRAINT skus_barcode_unique UNIQUE (barcode);


--
-- Name: skus skus_code_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.skus
    ADD CONSTRAINT skus_code_unique UNIQUE (code);


--
-- Name: skus skus_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.skus
    ADD CONSTRAINT skus_pkey PRIMARY KEY (id);


--
-- Name: stock_movements stock_movements_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stock_movements
    ADD CONSTRAINT stock_movements_pkey PRIMARY KEY (id);


--
-- Name: store_settings store_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.store_settings
    ADD CONSTRAINT store_settings_pkey PRIMARY KEY (key);


--
-- Name: user_roles user_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_pkey PRIMARY KEY (user_id, role_id);


--
-- Name: users users_email_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_unique UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_audit_entity_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_entity_created ON public.audit_logs USING btree (entity_type, entity_id, created_at DESC);


--
-- Name: idx_billing_entries_period; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_billing_entries_period ON public.billing_entries USING btree (billing_period_id);


--
-- Name: idx_billing_period_customer_reference; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_billing_period_customer_reference ON public.billing_periods USING btree (customer_id, reference_year, reference_month);


--
-- Name: idx_cart_items_cart_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_cart_items_cart_id ON public.cart_items USING btree (cart_id);


--
-- Name: idx_customers_collaborator_category; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_customers_collaborator_category ON public.customers USING btree (collaborator_category_id);


--
-- Name: idx_customers_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_customers_status ON public.customers USING btree (status);


--
-- Name: idx_email_verification_user; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_email_verification_user ON public.email_verification_tokens USING btree (user_id) WHERE (used_at IS NULL);


--
-- Name: idx_forecast_snapshots_sku_month; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_forecast_snapshots_sku_month ON public.forecast_snapshots USING btree (sku_id, reference_month DESC);


--
-- Name: idx_inventory_lots_sku_fifo; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_inventory_lots_sku_fifo ON public.inventory_lots USING btree (sku_id, created_at);


--
-- Name: idx_invoices_customer_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_invoices_customer_status ON public.invoices USING btree (customer_id, status);


--
-- Name: idx_invoices_due_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_invoices_due_status ON public.invoices USING btree (due_at, status);


--
-- Name: idx_jobs_status_available; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_jobs_status_available ON public.jobs USING btree (status, available_at);


--
-- Name: idx_orders_customer_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_orders_customer_created ON public.orders USING btree (customer_id, created_at DESC);


--
-- Name: idx_orders_status_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_orders_status_created ON public.orders USING btree (status, created_at DESC);


--
-- Name: idx_outbox_status_available; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_outbox_status_available ON public.outbox_events USING btree (status, available_at);


--
-- Name: idx_payment_charges_invoice_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_charges_invoice_status ON public.payment_charges USING btree (invoice_id, status);


--
-- Name: idx_payment_events_processed_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_payment_events_processed_created ON public.payment_events USING btree (processed, created_at);


--
-- Name: idx_products_category_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_products_category_active ON public.products USING btree (category_id, active, visible);


--
-- Name: idx_sessions_expires_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sessions_expires_at ON public.sessions USING btree (expires_at) WHERE (revoked_at IS NULL);


--
-- Name: idx_sessions_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sessions_user_id ON public.sessions USING btree (user_id);


--
-- Name: idx_skus_product_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_skus_product_active ON public.skus USING btree (product_id, active);


--
-- Name: idx_stock_movements_product_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_stock_movements_product_created ON public.stock_movements USING btree (sku_id, created_at DESC) WHERE ((movement_type)::text = ANY ((ARRAY['entry'::character varying, 'initial_stock'::character varying])::text[]));


--
-- Name: idx_stock_movements_sku_created; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_stock_movements_sku_created ON public.stock_movements USING btree (sku_id, created_at DESC);


--
-- Name: audit_logs audit_logs_actor_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES public.users(id);


--
-- Name: billing_adjustments billing_adjustments_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_adjustments
    ADD CONSTRAINT billing_adjustments_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: billing_adjustments billing_adjustments_invoice_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_adjustments
    ADD CONSTRAINT billing_adjustments_invoice_id_fkey FOREIGN KEY (invoice_id) REFERENCES public.invoices(id);


--
-- Name: billing_entries billing_entries_billing_period_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_entries
    ADD CONSTRAINT billing_entries_billing_period_id_fkey FOREIGN KEY (billing_period_id) REFERENCES public.billing_periods(id);


--
-- Name: billing_entries billing_entries_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_entries
    ADD CONSTRAINT billing_entries_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id);


--
-- Name: billing_entries billing_entries_order_return_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_entries
    ADD CONSTRAINT billing_entries_order_return_id_fkey FOREIGN KEY (order_return_id) REFERENCES public.order_returns(id);


--
-- Name: billing_periods billing_periods_customer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.billing_periods
    ADD CONSTRAINT billing_periods_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES public.customers(id);


--
-- Name: business_calendar business_calendar_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.business_calendar
    ADD CONSTRAINT business_calendar_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: cart_items cart_items_cart_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_cart_id_fkey FOREIGN KEY (cart_id) REFERENCES public.carts(id) ON DELETE CASCADE;


--
-- Name: cart_items cart_items_sku_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_sku_id_fkey FOREIGN KEY (sku_id) REFERENCES public.skus(id) ON DELETE RESTRICT;


--
-- Name: carts carts_customer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.carts
    ADD CONSTRAINT carts_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES public.customers(id) ON DELETE CASCADE;


--
-- Name: customer_limit_history customer_limit_history_changed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customer_limit_history
    ADD CONSTRAINT customer_limit_history_changed_by_fkey FOREIGN KEY (changed_by) REFERENCES public.users(id);


--
-- Name: customer_limit_history customer_limit_history_customer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customer_limit_history
    ADD CONSTRAINT customer_limit_history_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES public.customers(id) ON DELETE CASCADE;


--
-- Name: customers customers_approved_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_approved_by_fkey FOREIGN KEY (approved_by) REFERENCES public.users(id);


--
-- Name: customers customers_collaborator_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_collaborator_category_id_fkey FOREIGN KEY (collaborator_category_id) REFERENCES public.collaborator_categories(id) ON DELETE SET NULL;


--
-- Name: customers customers_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE RESTRICT;


--
-- Name: email_verification_tokens email_verification_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_verification_tokens
    ADD CONSTRAINT email_verification_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: forecast_snapshots forecast_snapshots_sku_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.forecast_snapshots
    ADD CONSTRAINT forecast_snapshots_sku_id_fkey FOREIGN KEY (sku_id) REFERENCES public.skus(id) ON DELETE CASCADE;


--
-- Name: inventory_balances inventory_balances_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.inventory_balances
    ADD CONSTRAINT inventory_balances_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.inventory_locations(id);


--
-- Name: inventory_balances inventory_balances_sku_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.inventory_balances
    ADD CONSTRAINT inventory_balances_sku_id_fkey FOREIGN KEY (sku_id) REFERENCES public.skus(id) ON DELETE RESTRICT;


--
-- Name: inventory_lots inventory_lots_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.inventory_lots
    ADD CONSTRAINT inventory_lots_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.inventory_locations(id);


--
-- Name: inventory_lots inventory_lots_sku_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.inventory_lots
    ADD CONSTRAINT inventory_lots_sku_id_fkey FOREIGN KEY (sku_id) REFERENCES public.skus(id) ON DELETE RESTRICT;


--
-- Name: inventory_lots inventory_lots_source_movement_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.inventory_lots
    ADD CONSTRAINT inventory_lots_source_movement_id_fkey FOREIGN KEY (source_movement_id) REFERENCES public.stock_movements(id) ON DELETE SET NULL;


--
-- Name: invoice_items invoice_items_billing_entry_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_items
    ADD CONSTRAINT invoice_items_billing_entry_id_fkey FOREIGN KEY (billing_entry_id) REFERENCES public.billing_entries(id);


--
-- Name: invoice_items invoice_items_invoice_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoice_items
    ADD CONSTRAINT invoice_items_invoice_id_fkey FOREIGN KEY (invoice_id) REFERENCES public.invoices(id) ON DELETE CASCADE;


--
-- Name: invoices invoices_billing_period_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT invoices_billing_period_id_fkey FOREIGN KEY (billing_period_id) REFERENCES public.billing_periods(id);


--
-- Name: invoices invoices_customer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.invoices
    ADD CONSTRAINT invoices_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES public.customers(id);


--
-- Name: order_items order_items_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id) ON DELETE CASCADE;


--
-- Name: order_items order_items_sku_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_sku_id_fkey FOREIGN KEY (sku_id) REFERENCES public.skus(id);


--
-- Name: order_return_items order_return_items_order_item_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_return_items
    ADD CONSTRAINT order_return_items_order_item_id_fkey FOREIGN KEY (order_item_id) REFERENCES public.order_items(id);


--
-- Name: order_return_items order_return_items_order_return_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_return_items
    ADD CONSTRAINT order_return_items_order_return_id_fkey FOREIGN KEY (order_return_id) REFERENCES public.order_returns(id) ON DELETE CASCADE;


--
-- Name: order_returns order_returns_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_returns
    ADD CONSTRAINT order_returns_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: order_returns order_returns_order_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_returns
    ADD CONSTRAINT order_returns_order_id_fkey FOREIGN KEY (order_id) REFERENCES public.orders(id);


--
-- Name: orders orders_customer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES public.customers(id);


--
-- Name: password_reset_tokens password_reset_tokens_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: payment_charges payment_charges_invoice_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payment_charges
    ADD CONSTRAINT payment_charges_invoice_id_fkey FOREIGN KEY (invoice_id) REFERENCES public.invoices(id);


--
-- Name: payments payments_invoice_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT payments_invoice_id_fkey FOREIGN KEY (invoice_id) REFERENCES public.invoices(id);


--
-- Name: payments payments_payment_charge_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT payments_payment_charge_id_fkey FOREIGN KEY (payment_charge_id) REFERENCES public.payment_charges(id);


--
-- Name: price_history price_history_changed_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.price_history
    ADD CONSTRAINT price_history_changed_by_fkey FOREIGN KEY (changed_by) REFERENCES public.users(id);


--
-- Name: price_history price_history_sku_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.price_history
    ADD CONSTRAINT price_history_sku_id_fkey FOREIGN KEY (sku_id) REFERENCES public.skus(id) ON DELETE CASCADE;


--
-- Name: product_images product_images_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.product_images
    ADD CONSTRAINT product_images_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: products products_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.categories(id);


--
-- Name: role_permissions role_permissions_permission_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_permission_id_fkey FOREIGN KEY (permission_id) REFERENCES public.permissions(id) ON DELETE CASCADE;


--
-- Name: role_permissions role_permissions_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;


--
-- Name: sessions sessions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: skus skus_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.skus
    ADD CONSTRAINT skus_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: stock_movements stock_movements_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stock_movements
    ADD CONSTRAINT stock_movements_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: stock_movements stock_movements_location_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stock_movements
    ADD CONSTRAINT stock_movements_location_id_fkey FOREIGN KEY (location_id) REFERENCES public.inventory_locations(id);


--
-- Name: stock_movements stock_movements_sku_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.stock_movements
    ADD CONSTRAINT stock_movements_sku_id_fkey FOREIGN KEY (sku_id) REFERENCES public.skus(id) ON DELETE RESTRICT;


--
-- Name: user_roles user_roles_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;


--
-- Name: user_roles user_roles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--



-- Dados iniciais (roles, permissões, catálogo demo, configurações)
--
-- PostgreSQL database dump
--


-- Dumped from database version 16.14
-- Dumped by pg_dump version 16.14

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: audit_logs; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: collaborator_categories; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.collaborator_categories VALUES ('d0000000-0000-4000-8000-000000000001', 'Funcionário', 'funcionario', 15.00, true, '2026-07-20 20:15:56.749728+00', '2026-07-20 20:15:56.749728+00');


--
-- Data for Name: customers; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: billing_periods; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: invoices; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: billing_adjustments; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: orders; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: order_returns; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: billing_entries; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: business_calendar; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: carts; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: categories; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.categories VALUES ('d0000000-0000-4000-8000-000000000001', 'Mercearia', 'mercearia', true, '2026-07-20 20:15:56.696864+00', '2026-07-20 20:15:56.696864+00');


--
-- Data for Name: products; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.products VALUES ('d0000000-0000-4000-8000-000000000011', 'd0000000-0000-4000-8000-000000000001', 'Arroz 5kg', 'arroz-5kg', 'Arroz branco tipo 1', true, true, '2026-07-20 20:15:56.696864+00', '2026-07-20 20:15:56.696864+00', 30.00, false, NULL, 0, 0);
INSERT INTO public.products VALUES ('d0000000-0000-4000-8000-000000000012', 'd0000000-0000-4000-8000-000000000001', 'Feijão Carioca 1kg', 'feijao-carioca-1kg', 'Feijão carioca', true, true, '2026-07-20 20:15:56.696864+00', '2026-07-20 20:15:56.696864+00', 30.00, false, NULL, 0, 0);
INSERT INTO public.products VALUES ('d0000000-0000-4000-8000-000000000013', 'd0000000-0000-4000-8000-000000000001', 'Óleo de Soja 900ml', 'oleo-soja-900ml', 'Óleo de soja refinado', true, true, '2026-07-20 20:15:56.696864+00', '2026-07-20 20:15:56.696864+00', 30.00, false, NULL, 0, 0);
INSERT INTO public.products VALUES ('d0000000-0000-4000-8000-000000000014', 'd0000000-0000-4000-8000-000000000001', 'Macarrão Espaguete 500g', 'macarrao-espaguete-500g', 'Massa seca', true, true, '2026-07-20 20:15:56.696864+00', '2026-07-20 20:15:56.696864+00', 30.00, false, NULL, 0, 0);


--
-- Data for Name: skus; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.skus VALUES ('d0000000-0000-4000-8000-000000000021', 'd0000000-0000-4000-8000-000000000011', 'ARR-5KG', NULL, 'UN', 2490, NULL, 0, true, '2026-07-20 20:15:56.696864+00', '2026-07-20 20:15:56.696864+00');
INSERT INTO public.skus VALUES ('d0000000-0000-4000-8000-000000000022', 'd0000000-0000-4000-8000-000000000012', 'FEJ-1KG', NULL, 'UN', 899, NULL, 0, true, '2026-07-20 20:15:56.696864+00', '2026-07-20 20:15:56.696864+00');
INSERT INTO public.skus VALUES ('d0000000-0000-4000-8000-000000000023', 'd0000000-0000-4000-8000-000000000013', 'OLE-900', NULL, 'UN', 699, NULL, 0, true, '2026-07-20 20:15:56.696864+00', '2026-07-20 20:15:56.696864+00');
INSERT INTO public.skus VALUES ('d0000000-0000-4000-8000-000000000024', 'd0000000-0000-4000-8000-000000000014', 'MAC-500', NULL, 'UN', 459, NULL, 0, true, '2026-07-20 20:15:56.696864+00', '2026-07-20 20:15:56.696864+00');


--
-- Data for Name: cart_items; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: customer_limit_history; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: email_verification_tokens; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: forecast_snapshots; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: inventory_locations; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.inventory_locations VALUES ('c0000000-0000-4000-8000-000000000001', 'Principal', true);


--
-- Data for Name: inventory_balances; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.inventory_balances VALUES ('d0000000-0000-4000-8000-000000000031', 'c0000000-0000-4000-8000-000000000001', 'd0000000-0000-4000-8000-000000000021', 120, 0, '2026-07-20 20:15:56.696864+00');
INSERT INTO public.inventory_balances VALUES ('d0000000-0000-4000-8000-000000000032', 'c0000000-0000-4000-8000-000000000001', 'd0000000-0000-4000-8000-000000000022', 80, 0, '2026-07-20 20:15:56.696864+00');
INSERT INTO public.inventory_balances VALUES ('d0000000-0000-4000-8000-000000000033', 'c0000000-0000-4000-8000-000000000001', 'd0000000-0000-4000-8000-000000000023', 200, 0, '2026-07-20 20:15:56.696864+00');
INSERT INTO public.inventory_balances VALUES ('d0000000-0000-4000-8000-000000000034', 'c0000000-0000-4000-8000-000000000001', 'd0000000-0000-4000-8000-000000000024', 150, 0, '2026-07-20 20:15:56.696864+00');


--
-- Data for Name: stock_movements; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: inventory_lots; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.inventory_lots VALUES ('82dea06d-325c-4575-8e41-653df2698e4e', 'c0000000-0000-4000-8000-000000000001', 'd0000000-0000-4000-8000-000000000021', 120, 0, NULL, '2026-07-20 20:15:56.733158+00');
INSERT INTO public.inventory_lots VALUES ('24518f53-d701-4211-ad26-ec1c67a352ec', 'c0000000-0000-4000-8000-000000000001', 'd0000000-0000-4000-8000-000000000022', 80, 0, NULL, '2026-07-20 20:15:56.733158+00');
INSERT INTO public.inventory_lots VALUES ('cbffcba9-d1f3-4a5d-8ba9-c142db37de74', 'c0000000-0000-4000-8000-000000000001', 'd0000000-0000-4000-8000-000000000023', 200, 0, NULL, '2026-07-20 20:15:56.733158+00');
INSERT INTO public.inventory_lots VALUES ('1b144340-9e20-408e-9fc6-bf6fa324ec9b', 'c0000000-0000-4000-8000-000000000001', 'd0000000-0000-4000-8000-000000000024', 150, 0, NULL, '2026-07-20 20:15:56.733158+00');


--
-- Data for Name: invoice_items; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: jobs; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: order_items; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: order_return_items; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: outbox_events; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: password_reset_tokens; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: payment_charges; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: payment_events; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: payments; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: permissions; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.permissions VALUES ('b0000000-0000-4000-8000-000000000001', 'products.read', 'Consultar produtos');
INSERT INTO public.permissions VALUES ('b0000000-0000-4000-8000-000000000002', 'products.write', 'Gerenciar produtos');
INSERT INTO public.permissions VALUES ('b0000000-0000-4000-8000-000000000003', 'inventory.read', 'Consultar estoque');
INSERT INTO public.permissions VALUES ('b0000000-0000-4000-8000-000000000004', 'inventory.adjust', 'Ajustar estoque');
INSERT INTO public.permissions VALUES ('b0000000-0000-4000-8000-000000000005', 'customers.read', 'Consultar clientes');
INSERT INTO public.permissions VALUES ('b0000000-0000-4000-8000-000000000006', 'customers.approve', 'Aprovar clientes');
INSERT INTO public.permissions VALUES ('b0000000-0000-4000-8000-000000000007', 'customers.change_limit', 'Alterar limite');
INSERT INTO public.permissions VALUES ('b0000000-0000-4000-8000-000000000008', 'orders.read', 'Consultar pedidos');
INSERT INTO public.permissions VALUES ('b0000000-0000-4000-8000-000000000009', 'orders.cancel', 'Cancelar pedidos');
INSERT INTO public.permissions VALUES ('b0000000-0000-4000-8000-00000000000a', 'billing.read', 'Consultar faturamento');
INSERT INTO public.permissions VALUES ('b0000000-0000-4000-8000-00000000000b', 'billing.close', 'Fechar período');
INSERT INTO public.permissions VALUES ('b0000000-0000-4000-8000-00000000000c', 'payments.read', 'Consultar pagamentos');
INSERT INTO public.permissions VALUES ('b0000000-0000-4000-8000-00000000000d', 'reports.read', 'Consultar relatórios');
INSERT INTO public.permissions VALUES ('b0000000-0000-4000-8000-00000000000e', 'settings.write', 'Alterar configurações');
INSERT INTO public.permissions VALUES ('b0000000-0000-4000-8000-00000000000f', 'audit.read', 'Consultar auditoria');
INSERT INTO public.permissions VALUES ('b0000000-0000-4000-8000-000000000010', 'users.manage', 'Gerenciar usuários');
INSERT INTO public.permissions VALUES ('b0000000-0000-4000-8000-000000000011', 'inventory.entry', 'Registrar entrada de estoque');
INSERT INTO public.permissions VALUES ('b0000000-0000-4000-8000-000000000012', 'inventory.loss', 'Registrar perda e avaria');
INSERT INTO public.permissions VALUES ('b0000000-0000-4000-8000-000000000014', 'customers.write', 'Gerenciar clientes');


--
-- Data for Name: price_history; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: product_images; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.product_images VALUES ('d0000000-0000-4000-8000-000000000041', 'd0000000-0000-4000-8000-000000000011', '/product-images/arroz-5kg', 0, 'Arroz 5kg', '2026-07-20 20:15:56.710736+00');
INSERT INTO public.product_images VALUES ('d0000000-0000-4000-8000-000000000042', 'd0000000-0000-4000-8000-000000000012', '/product-images/feijao-carioca-1kg', 0, 'Feijão Carioca 1kg', '2026-07-20 20:15:56.710736+00');
INSERT INTO public.product_images VALUES ('d0000000-0000-4000-8000-000000000043', 'd0000000-0000-4000-8000-000000000013', '/product-images/oleo-soja-900ml', 0, 'Óleo de Soja 900ml', '2026-07-20 20:15:56.710736+00');
INSERT INTO public.product_images VALUES ('d0000000-0000-4000-8000-000000000044', 'd0000000-0000-4000-8000-000000000014', '/product-images/macarrao-espaguete-500g', 0, 'Macarrão Espaguete 500g', '2026-07-20 20:15:56.710736+00');


--
-- Data for Name: roles; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.roles VALUES ('a0000000-0000-4000-8000-000000000001', 'system_admin', 'Administrador do sistema');
INSERT INTO public.roles VALUES ('a0000000-0000-4000-8000-000000000002', 'manager', 'Gerente');
INSERT INTO public.roles VALUES ('a0000000-0000-4000-8000-000000000003', 'customer', 'Cliente');


--
-- Data for Name: role_permissions; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-000000000001');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-000000000002');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-000000000003');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-000000000004');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-000000000005');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-000000000006');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-000000000007');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-000000000008');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-000000000009');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-00000000000a');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-00000000000b');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-00000000000c');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-00000000000d');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-00000000000e');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-00000000000f');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-000000000010');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-000000000001');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-000000000002');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-000000000003');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-000000000004');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-000000000005');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-000000000006');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-000000000007');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-000000000008');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-000000000009');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-00000000000a');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-00000000000b');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-00000000000c');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-00000000000d');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-00000000000e');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-00000000000f');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-000000000011');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-000000000012');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-000000000011');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-000000000012');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-000000000014');
INSERT INTO public.role_permissions VALUES ('a0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-000000000014');


--
-- Data for Name: sessions; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: store_settings; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.store_settings VALUES ('default_margin_percent', '30', '2026-07-20 20:15:56.733158+00');


--
-- Data for Name: user_roles; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- PostgreSQL database dump complete
--


