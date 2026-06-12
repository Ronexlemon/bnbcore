

CREATE TYPE tenant_status AS ENUM (
    'trial',
    'active',
    'expired',
    'suspended'
);

-- =========================
-- PLATFORM ADMINS 
-- =========================

CREATE TABLE platform_users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL, -- super_admin | billing_admin | support
    created_at TIMESTAMP DEFAULT now()
);

-- =========================
-- PLATFORM SETTINGS
-- =========================

CREATE TABLE platform_settings (
    id UUID PRIMARY KEY,
    key TEXT UNIQUE NOT NULL,
    value TEXT NOT NULL,
    updated_by UUID REFERENCES platform_users(id),
    updated_at TIMESTAMP DEFAULT now()
);

-- =========================
-- TENANTS (SHOPS)
-- =========================

CREATE TABLE tenants (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    subdomain     TEXT UNIQUE NOT NULL,
    shop_description TEXT NOT NULL,
    banner  TEXT ,
    profile_photo TEXT,
    long_description TEXT,
    socials           JSONB NOT NULL DEFAULT '[]',
    phone_number  TEXT,
    status        tenant_status DEFAULT 'trial',
    trial_ends_at TIMESTAMPTZ NOT NULL DEFAULT now() + INTERVAL '14 days',
    created_at    TIMESTAMPTZ DEFAULT now()
);