package booking

import (
	"time"

	"github.com/google/uuid"
)

type BookingStatus string

const (
    BookingStatusPending   BookingStatus = "pending"
    BookingStatusConfirmed BookingStatus = "confirmed"
    BookingStatusCanceled  BookingStatus = "canceled"
    BookingStatusCompleted BookingStatus = "completed"
)

type Booking struct {
    ID         uuid.UUID
    TenantID   uuid.UUID
    UnitID     uuid.UUID
    GuestName  string
    GuestEmail string
    GuestPhone string
    StartDate  time.Time
    EndDate    time.Time
    TotalPrice float64
    Status     BookingStatus
    CreatedAt  time.Time
}

type BookedRange struct {
	BookingID uuid.UUID     `json:"booking_id"`
	StartDate time.Time     `json:"start_date"`
	EndDate   time.Time     `json:"end_date"`
	Status    BookingStatus `json:"status"`
}

type CreateBookingRequest struct {
    UnitID     uuid.UUID `json:"unit_id"`
    GuestName  string    `json:"guest_name"`
    GuestEmail string    `json:"guest_email"`
    GuestPhone string    `json:"guest_phone"`
    StartDate  time.Time `json:"start_date"`
    EndDate    time.Time `json:"end_date"`
}

type UpdateBookingRequest struct {
    Status *BookingStatus `json:"status"`
}