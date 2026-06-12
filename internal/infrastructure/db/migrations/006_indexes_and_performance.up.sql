-- =========================
-- TENANT ISOLATION INDEXES
-- =========================
CREATE INDEX idx_units_tenant_id     ON units(tenant_id);


CREATE INDEX idx_bookings_tenant_id  ON bookings(tenant_id);

-- =========================
-- SEARCH OPTIMIZATION
-- =========================
CREATE INDEX idx_units_location      ON units(location);
CREATE INDEX idx_units_lat_lng       ON units(latitude, longitude);
CREATE UNIQUE INDEX idx_units_tenant_slug ON units (tenant_id, slug);

-- =========================
-- BOOKING PERFORMANCE
-- =========================
CREATE INDEX idx_bookings_dates      ON bookings(start_date, end_date);
CREATE INDEX idx_bookings_checkout_lookup 
ON bookings (end_date, status, id);

-- =========================
-- SUBSCRIPTION CHECKS
-- =========================
CREATE INDEX idx_subscriptions_user ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);

CREATE INDEX idx_magic_link_tokens_token_hash ON magic_link_tokens(token_hash);
CREATE INDEX idx_magic_link_tokens_user_id    ON magic_link_tokens(user_id);


CREATE INDEX idx_units_amenities_jsonb ON units USING gin(amenities);