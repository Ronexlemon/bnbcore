package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ronexlemon/bnbcore/internal/domain/booking"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/db"
)

type BookingRepository struct {
	DbConnection *db.PostgresConn
}

func NewBookingRepository(dbconn *db.PostgresConn) (*BookingRepository,error) {
	return &BookingRepository{DbConnection: dbconn},nil
}


func (b *BookingRepository) Create(ctx context.Context, bk *booking.Booking) (*booking.Booking, error) {
	query := `
		INSERT INTO bookings (id, tenant_id, unit_id, guest_name, guest_email, guest_phone,
		                      start_date, end_date, total_price, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		RETURNING created_at
	`
	err := b.DbConnection.Pool.QueryRow(ctx, query,
		bk.ID,
		bk.TenantID,
		bk.UnitID,
		bk.GuestName,
		bk.GuestEmail,
		bk.GuestPhone,
		bk.StartDate,
		bk.EndDate,
		bk.TotalPrice,
		bk.Status,
	).Scan(&bk.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}
	return bk, nil
}

func (b *BookingRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*booking.Booking, error) {
	var bk booking.Booking

	query := `
		SELECT id, tenant_id, unit_id, guest_name, guest_email, guest_phone,
		       start_date, end_date, total_price, status, created_at
		FROM bookings
		WHERE id = $1
		  AND tenant_id = $2
	`
	err := b.DbConnection.Pool.QueryRow(ctx, query, id, tenantID).Scan(
		&bk.ID,
		&bk.TenantID,
		&bk.UnitID,
		&bk.GuestName,
		&bk.GuestEmail,
		&bk.GuestPhone,
		&bk.StartDate,
		&bk.EndDate,
		&bk.TotalPrice,
		&bk.Status,
		&bk.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("booking not found")
		}
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	return &bk, nil
}


func (b *BookingRepository) GetAll(ctx context.Context, tenantID uuid.UUID) ([]*booking.Booking, error) {
	query := `
		SELECT id, tenant_id, unit_id, guest_name, guest_email, guest_phone,
		       start_date, end_date, total_price, status, created_at
		FROM bookings
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := b.DbConnection.Pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bookings: %w", err)
	}
	defer rows.Close()

	var bookings []*booking.Booking
	for rows.Next() {
		var bk booking.Booking
		if err := rows.Scan(
			&bk.ID,
			&bk.TenantID,
			&bk.UnitID,
			&bk.GuestName,
			&bk.GuestEmail,
			&bk.GuestPhone,
			&bk.StartDate,
			&bk.EndDate,
			&bk.TotalPrice,
			&bk.Status,
			&bk.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, &bk)
	}
	return bookings, nil
}



func (b *BookingRepository) GetByUnit(ctx context.Context, unitID, tenantID uuid.UUID) ([]*booking.Booking, error) {
	query := `
		SELECT id, tenant_id, unit_id, guest_name, guest_email, guest_phone,
		       start_date, end_date, total_price, status, created_at
		FROM bookings
		WHERE unit_id = $1
		  AND tenant_id = $2
		ORDER BY start_date ASC
	`
	rows, err := b.DbConnection.Pool.Query(ctx, query, unitID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bookings by unit: %w", err)
	}
	defer rows.Close()

	var bookings []*booking.Booking
	for rows.Next() {
		var bk booking.Booking
		if err := rows.Scan(
			&bk.ID,
			&bk.TenantID,
			&bk.UnitID,
			&bk.GuestName,
			&bk.GuestEmail,
			&bk.GuestPhone,
			&bk.StartDate,
			&bk.EndDate,
			&bk.TotalPrice,
			&bk.Status,
			&bk.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, &bk)
	}
	return bookings, nil
}


func (b *BookingRepository) UpdateStatus(ctx context.Context, id, tenantID uuid.UUID, status booking.BookingStatus) (*booking.Booking, error) {
	query := `
		UPDATE bookings
		SET status = $1
		WHERE id = $2
		  AND tenant_id = $3
		RETURNING id, tenant_id, unit_id, guest_name, guest_email, guest_phone,
		          start_date, end_date, total_price, status, created_at
	`
	var bk booking.Booking
	err := b.DbConnection.Pool.QueryRow(ctx, query, status, id, tenantID).Scan(
		&bk.ID,
		&bk.TenantID,
		&bk.UnitID,
		&bk.GuestName,
		&bk.GuestEmail,
		&bk.GuestPhone,
		&bk.StartDate,
		&bk.EndDate,
		&bk.TotalPrice,
		&bk.Status,
		&bk.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("booking not found")
		}
		return nil, fmt.Errorf("failed to update booking status: %w", err)
	}
	return &bk, nil
}


func (b *BookingRepository) Cancel(ctx context.Context, id, tenantID uuid.UUID) error {
	query := `
		UPDATE bookings
		SET status = 'canceled'
		WHERE id = $1
		  AND tenant_id = $2
		  AND status NOT IN ('canceled', 'completed')
	`
	tag, err := b.DbConnection.Pool.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to cancel booking: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("booking not found or cannot be canceled")
	}
	return nil
}


func (b *BookingRepository) CheckAvailability(ctx context.Context, unitID uuid.UUID, startDate, endDate time.Time) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM bookings
			WHERE unit_id = $1
			  AND status IN ('pending', 'confirmed')
			  AND start_date < $3
			  AND end_date   > $2
		)
	`
	var overlaps bool
	err := b.DbConnection.Pool.QueryRow(ctx, query, unitID, startDate, endDate).Scan(&overlaps)
	if err != nil {
		return false, fmt.Errorf("failed to check availability: %w", err)
	}
	return !overlaps, nil
}
func (b *BookingRepository) GetBookedDates(ctx context.Context, unitID uuid.UUID) ([]*booking.BookedRange, error) {
	query := `
		SELECT id, start_date, end_date, status
		FROM bookings
		WHERE unit_id = $1
		  AND status IN ('pending', 'confirmed')
		  AND end_date >= NOW()
		ORDER BY start_date ASC
	`
	rows, err := b.DbConnection.Pool.Query(ctx, query, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch booked dates: %w", err)
	}
	defer rows.Close()

	var ranges []*booking.BookedRange
	for rows.Next() {
		var br booking.BookedRange
		if err := rows.Scan(
			&br.BookingID,
			&br.StartDate,
			&br.EndDate,
			&br.Status,
		); err != nil {
			return nil, fmt.Errorf("failed to scan booked range: %w", err)
		}
		ranges = append(ranges, &br)
	}
	return ranges, nil
}
var _ booking.BookingRepository = (*BookingRepository)(nil)