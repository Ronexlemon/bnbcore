

CREATE TYPE subscription_status AS ENUM (
    'active',
    'trial',
    'expired',
    'canceled'
);

CREATE TYPE plan_type AS ENUM (
    'basic',
    'pro',
    'enterprise'
);

-- =========================
-- SUBSCRIPTIONS
-- =========================

CREATE TABLE subscriptions (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id            UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    plan                 plan_type NOT NULL,
    status               subscription_status DEFAULT 'trial',
    amount               NUMERIC NOT NULL,
    currency             TEXT DEFAULT 'KES',
    billing_cycle TEXT NOT NULL DEFAULT 'monthly',
    current_period_start TIMESTAMP,
    current_period_end   TIMESTAMP,
    created_at           TIMESTAMP DEFAULT now()
);

-- =========================
-- PAYMENTS
-- =========================

CREATE TABLE revenue (
    id UUID PRIMARY KEY,

    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    subscription_id UUID REFERENCES subscriptions(id),

    amount NUMERIC NOT NULL,
    currency TEXT DEFAULT 'USD',

    provider TEXT, -- stripe | mpesa | paypal

    status TEXT, -- pending | success | failed

    created_at TIMESTAMP DEFAULT now()
);