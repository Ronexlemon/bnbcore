
CREATE TYPE billing_cycle_type AS ENUM ('monthly', 'yearly', 'weekly', 'quarterly', 'lifetime');
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
    user_id              UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    plan                 plan_type NOT NULL,
    status               subscription_status DEFAULT 'trial',
    amount               NUMERIC(10,2) NOT NULL,
    currency             TEXT DEFAULT 'KES',
    billing_cycle        billing_cycle_type NOT NULL DEFAULT 'monthly',
    current_period_start TIMESTAMPTZ,
    current_period_end   TIMESTAMPTZ,
    created_at           TIMESTAMPTZ DEFAULT now(),
    updated_at           TIMESTAMPTZ DEFAULT now()
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

    created_at TIMESTAMPTZ DEFAULT now()
);