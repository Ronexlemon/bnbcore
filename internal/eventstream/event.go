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
	BaseEvent
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

type BaseEvent struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	UserID    uuid.UUID `json:"user_id"`
	UserEmail string    `json:"user_email"`
	ShopName  string    `json:"shop_name"`
	OccuredAt time.Time `json:"occured_at"`
}


type TenantCreatedEvent struct {
	BaseEvent
	Subdomain string `json:"subdomain"`
}

type TenantUpdatedEvent struct {
	BaseEvent
	Changes map[string]any `json:"changes"` 
}

type TenantDeletedEvent struct {
	BaseEvent
}

type UnitUpdatedEvent struct {
	BaseEvent
	UnitID  uuid.UUID      `json:"unit_id"`
	Title   string         `json:"title"`
	Changes map[string]any `json:"changes"`
}

type UnitDeletedEvent struct {
	BaseEvent
	UnitID uuid.UUID `json:"unit_id"`
	Title  string    `json:"title"`
}
type SubscriptionEvent struct {
	BaseEvent
	SubscriptionID uuid.UUID `json:"subscription_id"`
	Plan           string    `json:"plan"`
	BillingCycle   string    `json:"billing_cycle"`
	Amount         float64   `json:"amount"`
	Currency       string    `json:"currency"`
	ExpiresAt      time.Time `json:"expires_at"`
}


//signup
type UserSignedUpEvent struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	SignupLink string `json:"signup_link"`
}