package booking


import (
    "context"
    "time"

    "github.com/google/uuid"
)



type BookingRepository interface {
    Create(ctx context.Context, booking *Booking) (*Booking, error)
    GetByID(ctx context.Context, id, tenantID uuid.UUID) (*Booking, error)
    GetAll(ctx context.Context, tenantID uuid.UUID) ([]*Booking, error)
    GetByUnit(ctx context.Context, unitID, tenantID uuid.UUID) ([]*Booking, error)
    UpdateStatus(ctx context.Context, id, tenantID uuid.UUID, status BookingStatus) (*Booking, error)
    Cancel(ctx context.Context, id, tenantID uuid.UUID) error
    CheckAvailability(ctx context.Context, unitID uuid.UUID, startDate, endDate time.Time) (bool, error)
}