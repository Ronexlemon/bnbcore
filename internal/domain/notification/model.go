package notification

import (
	"time"

	"github.com/google/uuid"
)

type Type    string
type Channel string
type Status  string

const (
	TypeBookingCreated      Type = "booking_created"
	TypeBookingConfirmed    Type = "booking_confirmed"
	TypeBookingCanceled     Type = "booking_canceled"
	TypeBookingCompleted    Type = "booking_completed"
	TypeSubscriptionExpiring Type = "subscription_expiring"
	TypeSubscriptionExpired  Type = "subscription_expired"
	TypeUnitCreated         Type = "unit_created"
	TypePaymentReceived     Type = "payment_received"
	TypePaymentFailed       Type = "payment_failed"

	ChannelWhatsApp Channel = "whatsapp"
	ChannelEmail    Channel = "email"
	ChannelSMS      Channel = "sms"
	ChannelInApp    Channel = "in_app"

	StatusPending Status = "pending"
	StatusSent    Status = "sent"
	StatusFailed  Status = "failed"
	StatusRead    Status = "read"
)

type Notification struct {
	ID        uuid.UUID
	TenantID  *uuid.UUID
	UserID    *uuid.UUID
	Type      Type
	Channel   Channel
	Status    Status
	Title     string
	Message   string
	Metadata  map[string]any
	ReadAt    *time.Time
	SentAt    *time.Time
	FailedAt  *time.Time
	Error     *string
	CreatedAt time.Time
}

type CreateNotificationRequest struct {
	TenantID *uuid.UUID     `json:"tenant_id"`
	UserID   *uuid.UUID     `json:"user_id"`
	Type     Type           `json:"type"`
	Channel  Channel        `json:"channel"`
	Title    string         `json:"title"`
	Message  string         `json:"message"`
	Metadata map[string]any `json:"metadata"`
}