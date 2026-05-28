

CREATE TYPE user_role AS ENUM (
    'owner',
    'admin',
    'staff'
);

-- =========================
-- TENANT USERS (STAFF)
-- =========================

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role          user_role DEFAULT 'owner',
    is_active     BOOLEAN DEFAULT true,
    created_at    TIMESTAMP DEFAULT now(),  
    updated_at    TIMESTAMP DEFAULT now()
);

-- -- =========================
-- -- REFRESH TOKENS (AUTH)
-- -- =========================

-- CREATE TABLE refresh_tokens (
--     id UUID PRIMARY KEY,
--     user_id UUID REFERENCES users(id) ON DELETE CASCADE,
--     token TEXT NOT NULL,
--     expires_at TIMESTAMP NOT NULL,
--     created_at TIMESTAMP DEFAULT now()
-- );