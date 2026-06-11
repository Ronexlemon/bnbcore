package worker

import (
	"time"

	"github.com/google/uuid"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/queue"
)


type BookingCreatedPayload struct {
	NotifID    uuid.UUID
	BookingID  uuid.UUID
	TenantID   uuid.UUID
	GuestName  string
	GuestPhone string
	UnitTitle  string
	StartDate  time.Time
	EndDate    time.Time
	TotalPrice float64
}

type BookingStatusPayload struct {
	BookingID  uuid.UUID
	GuestName  string
	GuestPhone string
	UnitTitle  string
	Status     string
}

type UserSignupPayload struct {
	UserID  string
	Email  string
	Link string
}

type PasswordResetPayload struct {
	UserID  string
	Email  string
	Link string
}


var WhatsAppCreatedTask = queue.TaskDef[BookingCreatedPayload]{
	Name:    "whatsapp:booking_created",
	Queue:   "created",
	Retry:   3,
	Timeout: 10 * time.Second,
	Unique:  10 * time.Minute,
}



var WhatsAppStatusTask = queue.TaskDef[BookingStatusPayload]{
	Name:    "whatsapp:booking_status",
	Queue:   "status",
	Retry:   3,
	Timeout: 10 * time.Second,
	Unique:  10 * time.Minute,
}

var UserSignUpTask = queue.TaskDef[UserSignupPayload]{
	Name:    "user:registration_signup",
	Queue:   "signup",
	Retry:   3,
	Timeout: 10 * time.Second,
	Unique:  10 * time.Minute,
}

var PasswordResetEmailTask = queue.TaskDef[PasswordResetPayload]{
	Name:    "user:password_reset",
	Queue:   "pass_reset",
	Retry:   3,
	Timeout: 10 * time.Second,
	Unique:  10 * time.Minute,
}