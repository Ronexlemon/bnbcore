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
    GetBookedDates(ctx context.Context, unitID uuid.UUID) ([]*BookedRange, error)
    GetPendingBookings(ctx context.Context, tenantID uuid.UUID) ([]*Booking, error)
    GetCheckEvents(ctx context.Context, tenantID uuid.UUID, date time.Time, checkType CheckType) ([]*Booking, error)
    GetRevenue(ctx context.Context, tenantID uuid.UUID, date time.Time) (float64, error)
    GetTotalGuestsServed(ctx context.Context, tenantID uuid.UUID, date time.Time) (int, error)
    GetUnitOccupancy(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*UnitOccupancy, error)
    FindConfirmedBookingsEndingOnDate(ctx context.Context, targetDate time.Time, lastID uuid.UUID, batchSize int) ([]*Booking, error)
    Cancel(ctx context.Context, id, tenantID uuid.UUID) error
    CheckAvailability(ctx context.Context, unitID uuid.UUID, startDate, endDate time.Time) (bool, error)
}