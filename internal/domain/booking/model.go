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
    GuestNumber int32
    Source   string
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
    Source     string    `json:"source"`
    GuestNumber int32    `json:"guest_number"`
    GuestPhone string    `json:"guest_phone"`
    StartDate  time.Time `json:"start_date"`
    EndDate    time.Time `json:"end_date"`
}

type UpdateBookingRequest struct {
    Status *BookingStatus `json:"status"`
}
type UnitOccupancy struct {
	UnitID        uuid.UUID `json:"unit_id"`
	BookedNights  int       `json:"booked_nights"`
	TotalNights   int       `json:"total_nights"`
	OccupancyRate float64   `json:"occupancy_rate"` 
	Revenue       float64   `json:"revenue"`
}

type DashboardSummary struct {
	PendingBookings []*Booking          `json:"pending_bookings"`
	CheckIns        []*Booking          `json:"check_ins"`
	CheckOuts       []*Booking          `json:"check_outs"`
	TodayRevenue    float64                     `json:"today_revenue"`
	LifetimeRevenue float64                     `json:"lifetime_revenue"`
	TotalGuests     int                         `json:"total_guests_today"`
	UnitOccupancy   []*UnitOccupancy `json:"unit_occupancy"`
}