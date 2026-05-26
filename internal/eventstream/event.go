package eventstream


import (
    "time"
    "github.com/google/uuid"
)

type BookingCreatedEvent struct {
    BookingID  uuid.UUID `json:"booking_id"`
    TenantID   uuid.UUID `json:"tenant_id"`
    UnitID     uuid.UUID `json:"unit_id"`
    UnitTitle  string    `json:"unit_title"`
    GuestName  string    `json:"guest_name"`
    GuestEmail string    `json:"guest_email"`
    GuestPhone string    `json:"guest_phone"`
    StartDate  time.Time `json:"start_date"`
    EndDate    time.Time `json:"end_date"`
    TotalPrice float64   `json:"total_price"`
    ShopName   string    `json:"shop_name"`
    CreatedAt  time.Time `json:"created_at"`
}

type UnitCreatedEvent struct {
    UnitID    uuid.UUID `json:"unit_id"`
    TenantID  uuid.UUID `json:"tenant_id"`
    Title     string    `json:"title"`
    Location  string    `json:"location"`
    ShopName  string    `json:"shop_name"`
    CreatedAt time.Time `json:"created_at"`
}

type BookingStatusEvent struct {
    BookingID  uuid.UUID `json:"booking_id"`
    TenantID   uuid.UUID `json:"tenant_id"`
    GuestName  string    `json:"guest_name"`
    GuestPhone string    `json:"guest_phone"`
    Status     string    `json:"status"`
    UnitTitle  string    `json:"unit_title"`
}