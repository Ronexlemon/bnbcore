

CREATE TYPE notification_channel AS ENUM (
    'whatsapp',
    'email',
    'sms',
    'in_app'
);

CREATE TYPE notification_status AS ENUM (
    'pending',
    'sent',
    'failed',
    'read'
);

CREATE TABLE notifications (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID REFERENCES tenants(id) ON DELETE CASCADE,
    user_id      UUID REFERENCES users(id) ON DELETE CASCADE,
    type         TEXT NOT NULL,
    channel      notification_channel NOT NULL,
    status       notification_status DEFAULT 'pending',
    title        TEXT NOT NULL,
    message      TEXT NOT NULL,
    metadata     JSONB,               
    read_at      TIMESTAMPTZ,          
    sent_at      TIMESTAMPTZ,          
    failed_at    TIMESTAMPTZ,          
    error        TEXT,                
    created_at   TIMESTAMPTZ DEFAULT now() 
);

CREATE INDEX idx_notifications_tenant_created ON notifications(tenant_id, created_at DESC);
CREATE INDEX idx_notifications_user_created   ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_status    ON notifications(status);
CREATE INDEX idx_notifications_type      ON notifications(type);