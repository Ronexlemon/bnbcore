CREATE TABLE room_services (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    unit_id     UUID REFERENCES units(id) ON DELETE CASCADE,
    tenant_id   UUID REFERENCES tenants(id) ON DELETE CASCADE,
    agent_name  TEXT NOT NULL,
    mobile      TEXT NOT NULL,
    email       TEXT,
    is_active   BOOLEAN DEFAULT true,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);