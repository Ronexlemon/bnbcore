

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

   -- Identity
    name                TEXT NOT NULL,
    title               TEXT NOT NULL,
    short_description   TEXT,
    description         TEXT,
    slug                TEXT NOT NULL,
    type                TEXT NOT NULL ,

    -- Capacity
    guests              INTEGER NOT NULL DEFAULT 1,
    bedrooms            INTEGER NOT NULL DEFAULT 0,
    beds                INTEGER NOT NULL DEFAULT 0,
    bathrooms           INTEGER NOT NULL DEFAULT 0,

    -- Location
    location            TEXT,
    latitude            DOUBLE PRECISION,
    longitude           DOUBLE PRECISION,
    apartment_name      TEXT,
    house_number        TEXT,
    floor               TEXT,
    access_note         TEXT,

    -- Pricing
    price_weekday       NUMERIC NOT NULL DEFAULT 0,
    price_weekend       NUMERIC NOT NULL DEFAULT 0,

    -- Check-in / Check-out
    checkin_time        TIME,
    checkout_time       TIME,

    -- Flexible content
    amenities           JSONB NOT NULL DEFAULT '[]',
    rules               JSONB NOT NULL DEFAULT '[]',

    -- Contact
    phone_number        TEXT,

    -- Status
    status              unit_status DEFAULT 'active',

    created_at         TIMESTAMPTZ DEFAULT now(),
    updated_at         TIMESTAMPTZ DEFAULT now(),

    UNIQUE(tenant_id, slug)
);



CREATE TABLE unit_images (
    id UUID PRIMARY KEY,
    unit_id UUID REFERENCES units(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    sort_order  INTEGER DEFAULT 0,
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
    source TEXT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,

    total_price NUMERIC NOT NULL,

    status booking_status DEFAULT 'pending',

    created_at TIMESTAMP DEFAULT now()
);