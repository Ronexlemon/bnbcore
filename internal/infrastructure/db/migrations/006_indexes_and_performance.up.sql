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

-- =========================
-- BOOKING PERFORMANCE
-- =========================
CREATE INDEX idx_bookings_dates      ON bookings(start_date, end_date);

-- =========================
-- SUBSCRIPTION CHECKS
-- =========================
CREATE INDEX idx_subscriptions_tenant ON subscriptions(tenant_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);

CREATE INDEX idx_magic_link_tokens_token_hash ON magic_link_tokens(token_hash);
CREATE INDEX idx_magic_link_tokens_user_id    ON magic_link_tokens(user_id);