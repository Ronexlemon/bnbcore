

CREATE TYPE unit_status AS ENUM (
    'active',
    'inactive',
    'deleted',
    'maintenance'
);

CREATE TYPE booking_status AS ENUM (
    'pending',
    'confirmed',
    'canceled',
    'completed'
    
);

-- =========================
-- UNITS (LISTINGS)
-- =========================

CREATE TABLE units (
    id UUID PRIMARY KEY,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,

    title TEXT NOT NULL,
    name TEXT NOT NULL,
    type text NOT NULL,
    description TEXT,
    adults INTEGER NOT NULL DEFAULT 0,
    children INTEGER NOT NULL DEFAULT 0,
    phone_number TEXT,

    price_per_night NUMERIC NOT NULL,

    location TEXT,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,

    status unit_status DEFAULT 'active',
    amenities JSONB NOT NULL DEFAULT '{}',
    rules JSONB NOT NULL DEFAULT '{}',

    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);



CREATE TABLE unit_images (
    id UUID PRIMARY KEY,
    unit_id UUID REFERENCES units(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    image_type TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- =========================
-- BOOKINGS
-- =========================

CREATE TABLE bookings (
    id UUID PRIMARY KEY,

    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    unit_id UUID REFERENCES units(id) ON DELETE CASCADE,

    guest_name TEXT,
    guest_email TEXT,
    guest_phone TEXT,
    guest_number NUMERIC,

    start_date DATE NOT NULL,
    end_date DATE NOT NULL,

    total_price NUMERIC NOT NULL,

    status booking_status DEFAULT 'pending',

    created_at TIMESTAMP DEFAULT now()
);