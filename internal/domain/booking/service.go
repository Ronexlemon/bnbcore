

package booking

import (
    "context"
    "errors"
    "math"
    "github.com/google/uuid"
)

type BookingService struct {
    Repo BookingRepository
}

func NewBookingService(repo BookingRepository) *BookingService {
    return &BookingService{Repo: repo}
}

func (s *BookingService) CreateBooking(ctx context.Context, tenantID uuid.UUID, req CreateBookingRequest) (*Booking, error) {

    if !req.EndDate.After(req.StartDate) {
        return nil, errors.New("end_date must be after start_date")
    }

    available, err := s.Repo.CheckAvailability(ctx, req.UnitID, req.StartDate, req.EndDate)
    if err != nil {
        return nil, err
    }
    if !available {
        return nil, errors.New("unit is not available for the selected dates")
    }

    nights := math.Ceil(req.EndDate.Sub(req.StartDate).Hours() / 24)

    booking := &Booking{
        ID:         uuid.New(),
        TenantID:   tenantID,
        UnitID:     req.UnitID,
        GuestName:  req.GuestName,
        GuestEmail: req.GuestEmail,
        GuestPhone: req.GuestPhone,
        StartDate:  req.StartDate,
        EndDate:    req.EndDate,
        GuestNumber: req.GuestNumber,
        TotalPrice: nights, 
        Status:     BookingStatusPending,
    }

    return s.Repo.Create(ctx, booking)
}

func (s *BookingService) GetBooking(ctx context.Context, id, tenantID uuid.UUID) (*Booking, error) {
    return s.Repo.GetByID(ctx, id, tenantID)
}

func (s *BookingService) GetAllBookings(ctx context.Context, tenantID uuid.UUID) ([]*Booking, error) {
    return s.Repo.GetAll(ctx, tenantID)
}

func (s *BookingService) GetBookingsByUnit(ctx context.Context, unitID, tenantID uuid.UUID) ([]*Booking, error) {
    return s.Repo.GetByUnit(ctx, unitID, tenantID)
}

func (s *BookingService) UpdateStatus(ctx context.Context, id, tenantID uuid.UUID, status BookingStatus) (*Booking, error) {
    return s.Repo.UpdateStatus(ctx, id, tenantID, status)
}

func (s *BookingService) CancelBooking(ctx context.Context, id, tenantID uuid.UUID) error {
    return s.Repo.Cancel(ctx, id, tenantID)
}
func (s *BookingService) GetBookedDates(ctx context.Context, unitID uuid.UUID) ([]*BookedRange, error) {
	return s.Repo.GetBookedDates(ctx, unitID)
}