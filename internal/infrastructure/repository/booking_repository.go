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
		                      start_date, end_date, total_price,source, status,guest_number, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,$11,$12, NOW())
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
		bk.Source,
		bk.Status,
		bk.GuestNumber,
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
		       start_date, end_date, total_price,source, status,guest_number, created_at
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
		&bk.Source,
		&bk.Status,
		&bk.GuestNumber,
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
		       start_date, end_date, total_price,source, status,guest_number, created_at
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
			&bk.Source,
			&bk.Status,
			&bk.GuestNumber,
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
		       start_date, end_date, total_price,source, status,guest_number, created_at
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
			&bk.Source,
			&bk.Status,
			&bk.GuestNumber,
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
		          start_date, end_date, total_price,source, status,guest_number, created_at
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
		&bk.Source,
		&bk.Status,
		&bk.GuestNumber,
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

func (b *BookingRepository) FindConfirmedBookingsEndingOnDate(ctx context.Context, targetDate time.Time, lastID uuid.UUID, batchSize int) ([]*booking.Booking, error) {
    
    now := time.Now().UTC()
    if targetDate.Before(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)) {
        targetDate = now
    }

    // Base query looking for active checkouts exactly on the targeted day
    query := `
        SELECT id, tenant_id, unit_id, guest_name, guest_email, guest_phone, start_date, end_date, total_price, status, created_at
        FROM bookings
        WHERE status = 'confirmed' 
          AND end_date = $1
    `
    
    var args []interface{}
    args = append(args, targetDate.Format("2006-01-02"))

    // Keyset Pagination anchor using correct SQL positional parameters
    if lastID != uuid.Nil {
        query += " AND id > $2 ORDER BY id ASC LIMIT $3"
        args = append(args, lastID, batchSize)
    } else {
        query += " ORDER BY id ASC LIMIT $2"
        args = append(args, batchSize)
    }

    rows, err := b.DbConnection.Pool.Query(ctx, query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var bookings []*booking.Booking
    for rows.Next() {
        var bkt booking.Booking
        err := rows.Scan(
            &bkt.ID, &bkt.TenantID, &bkt.UnitID, &bkt.GuestName, &bkt.GuestEmail, 
            &bkt.GuestPhone, &bkt.StartDate, &bkt.EndDate, &bkt.TotalPrice, &bkt.Status, &bkt.CreatedAt,
        )
        if err != nil {
            return nil, err
        }
        bookings = append(bookings, &bkt)
    }

    if err = rows.Err(); err != nil {
        return nil, err
    }

    return bookings, nil
}

func (b *BookingRepository) GetPendingBookings(ctx context.Context, tenantID uuid.UUID) ([]*booking.Booking, error) {
	query := `
		SELECT id, tenant_id, unit_id, guest_name, guest_email, guest_phone,
		       start_date, end_date, total_price, source, status, guest_number, created_at
		FROM bookings
		WHERE tenant_id = $1
		  AND status = 'pending'
		ORDER BY created_at ASC
	`
	rows, err := b.DbConnection.Pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending bookings: %w", err)
	}
	defer rows.Close()

	var bookings []*booking.Booking
	for rows.Next() {
		var bk booking.Booking
		if err := rows.Scan(
			&bk.ID, &bk.TenantID, &bk.UnitID, &bk.GuestName, &bk.GuestEmail, &bk.GuestPhone,
			&bk.StartDate, &bk.EndDate, &bk.TotalPrice, &bk.Source, &bk.Status, &bk.GuestNumber, &bk.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, &bk)
	}
	return bookings, rows.Err()
}

func (b *BookingRepository) GetCheckIns(ctx context.Context, tenantID uuid.UUID, date time.Time) ([]*booking.Booking, error) {
	if date.IsZero() {
		date = time.Now().UTC()
	}

	query := `
		SELECT id, tenant_id, unit_id, guest_name, guest_email, guest_phone,
		       start_date, end_date, total_price, source, status, guest_number, created_at
		FROM bookings
		WHERE tenant_id = $1
		  AND status = 'confirmed'
		  AND start_date = $2
		ORDER BY start_date ASC
	`
	rows, err := b.DbConnection.Pool.Query(ctx, query, tenantID, date.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch check-ins: %w", err)
	}
	defer rows.Close()

	var bookings []*booking.Booking
	for rows.Next() {
		var bk booking.Booking
		if err := rows.Scan(
			&bk.ID, &bk.TenantID, &bk.UnitID, &bk.GuestName, &bk.GuestEmail, &bk.GuestPhone,
			&bk.StartDate, &bk.EndDate, &bk.TotalPrice, &bk.Source, &bk.Status, &bk.GuestNumber, &bk.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, &bk)
	}
	return bookings, rows.Err()
}


func (b *BookingRepository) GetCheckOuts(ctx context.Context, tenantID uuid.UUID, date time.Time) ([]*booking.Booking, error) {
	if date.IsZero() {
		date = time.Now().UTC()
	}

	query := `
		SELECT id, tenant_id, unit_id, guest_name, guest_email, guest_phone,
		       start_date, end_date, total_price, source, status, guest_number, created_at
		FROM bookings
		WHERE tenant_id = $1
		  AND status = 'confirmed'
		  AND end_date = $2
		ORDER BY end_date ASC
	`
	rows, err := b.DbConnection.Pool.Query(ctx, query, tenantID, date.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch check-outs: %w", err)
	}
	defer rows.Close()

	var bookings []*booking.Booking
	for rows.Next() {
		var bk booking.Booking
		if err := rows.Scan(
			&bk.ID, &bk.TenantID, &bk.UnitID, &bk.GuestName, &bk.GuestEmail, &bk.GuestPhone,
			&bk.StartDate, &bk.EndDate, &bk.TotalPrice, &bk.Source, &bk.Status, &bk.GuestNumber, &bk.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, &bk)
	}
	return bookings, rows.Err()
}
// Pass a zero time.Time to get lifetime revenue.
func (b *BookingRepository) GetRevenue(ctx context.Context, tenantID uuid.UUID, date time.Time) (float64, error) {
	var query string
	var args []interface{}

	args = append(args, tenantID)

	if date.IsZero() {
		query = `
			SELECT COALESCE(SUM(total_price), 0)
			FROM bookings
			WHERE tenant_id = $1
			  AND status IN ('confirmed', 'completed')
		`
	} else {
		query = `
			SELECT COALESCE(SUM(total_price), 0)
			FROM bookings
			WHERE tenant_id = $1
			  AND status IN ('confirmed', 'completed')
			  AND created_at::date = $2
		`
		args = append(args, date.Format("2006-01-02"))
	}

	var revenue float64
	err := b.DbConnection.Pool.QueryRow(ctx, query, args...).Scan(&revenue)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate revenue: %w", err)
	}
	return revenue, nil
}

// Pass a zero time.Time for lifetime total, or a date to filter by checkout date.
func (b *BookingRepository) GetTotalGuestsServed(ctx context.Context, tenantID uuid.UUID, date time.Time) (int, error) {
	var query string
	var args []interface{}

	args = append(args, tenantID)

	if date.IsZero() {
		query = `
			SELECT COALESCE(SUM(guest_number), 0)
			FROM bookings
			WHERE tenant_id = $1
			  AND status IN ('confirmed', 'completed')
		`
	} else {
		query = `
			SELECT COALESCE(SUM(guest_number), 0)
			FROM bookings
			WHERE tenant_id = $1
			  AND status IN ('confirmed', 'completed')
			  AND end_date = $2
		`
		args = append(args, date.Format("2006-01-02"))
	}

	var total int
	err := b.DbConnection.Pool.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total guests served: %w", err)
	}
	return total, nil
}

func (b *BookingRepository) GetUnitOccupancy(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*booking.UnitOccupancy, error) {
	totalNights := int(endDate.Sub(startDate).Hours()/24) + 1
	if totalNights <= 0 {
		return nil, fmt.Errorf("invalid date range: endDate must be after startDate")
	}

	// For each unit, sum the overlap (in nights) between its bookings and [startDate, endDate],
	// plus revenue from those overlapping bookings.
	query := `
		SELECT
			u.id AS unit_id,
			COALESCE(SUM(
				GREATEST(0,
					(LEAST(b.end_date, $3::date) - GREATEST(b.start_date, $2::date))
				)
			), 0) AS booked_nights,
			COALESCE(SUM(b.total_price), 0) AS revenue
		FROM units u
		LEFT JOIN bookings b
			ON b.unit_id = u.id
			AND b.tenant_id = u.tenant_id
			AND b.status IN ('confirmed', 'completed')
			AND b.start_date < $3::date
			AND b.end_date   > $2::date
		WHERE u.tenant_id = $1
		GROUP BY u.id
		ORDER BY u.id
	`

	rows, err := b.DbConnection.Pool.Query(ctx, query, tenantID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("failed to calculate unit occupancy: %w", err)
	}
	defer rows.Close()

	var results []*booking.UnitOccupancy
	for rows.Next() {
		var o booking.UnitOccupancy
		if err := rows.Scan(&o.UnitID, &o.BookedNights, &o.Revenue); err != nil {
			return nil, fmt.Errorf("failed to scan unit occupancy: %w", err)
		}
		o.TotalNights = totalNights
		if totalNights > 0 {
			o.OccupancyRate = float64(o.BookedNights) / float64(totalNights)
		}
		results = append(results, &o)
	}
	return results, rows.Err()
}
var _ booking.BookingRepository = (*BookingRepository)(nil)