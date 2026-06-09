package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ronexlemon/bnbcore/internal/domain/unit"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/db"
)

type UnitRepository struct {
	DbConnection *db.PostgresConn
}

func NewUnitRepository(dbconn *db.PostgresConn) (*UnitRepository, error) {
	return &UnitRepository{DbConnection: dbconn}, nil
}



func (u *UnitRepository) Create(ctx context.Context, un *unit.Unit) (*unit.Unit, error) {
	query := `
		INSERT INTO units (
			id, tenant_id,
			title, name, short_description, description, slug, type,
			guests, bedrooms, beds, bathrooms,
			location, latitude, longitude,
			apartment_name, house_number, floor, access_note,
			price_weekday, price_weekend,
			checkin_time, checkout_time,
			amenities, rules,
			phone_number, status,
			created_at, updated_at
		) VALUES (
			$1, $2,
			$3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12,
			$13, $14, $15,
			$16, $17, $18, $19,
			$20, $21,
			$22, $23,
			$24, $25,
			$26, $27,
			NOW(), NOW()
		)
		RETURNING created_at, updated_at
	`
	err := u.DbConnection.Pool.QueryRow(ctx, query,
		un.ID, un.TenantID,
		un.Title, un.Name, un.ShortDescription, un.Description, un.Slug, un.UnitType,
		un.Guests, un.Bedrooms, un.Beds, un.Bathrooms,
		un.Location, un.Latitude, un.Longitude,
		un.ApartmentName, un.HouseNumber, un.Floor, un.AccessNote,
		un.PriceWeekday, un.PriceWeekend,
		un.CheckinTime, un.CheckoutTime,
		un.Amenities, un.Rules,
		un.PhoneNumber, un.Status,
	).Scan(&un.CreatedAt, &un.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create unit: %w", err)
	}
	return un, nil
}


func (u *UnitRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*unit.Unit, error) {
	var un unit.Unit
	query := `
		SELECT
			id, tenant_id,
			title, name, short_description, description, slug, type,
			guests, bedrooms, beds, bathrooms,
			location, latitude, longitude,
			apartment_name, house_number, floor, access_note,
			price_weekday, price_weekend,
			checkin_time, checkout_time,
			amenities, rules,
			phone_number, status,
			created_at, updated_at
		FROM units
		WHERE id = $1 AND tenant_id = $2 AND status != 'deleted'
	`
	err := u.DbConnection.Pool.QueryRow(ctx, query, id, tenantID).Scan(
		&un.ID, &un.TenantID,
		&un.Title, &un.Name, &un.ShortDescription, &un.Description, &un.Slug, &un.UnitType,
		&un.Guests, &un.Bedrooms, &un.Beds, &un.Bathrooms,
		&un.Location, &un.Latitude, &un.Longitude,
		&un.ApartmentName, &un.HouseNumber, &un.Floor, &un.AccessNote,
		&un.PriceWeekday, &un.PriceWeekend,
		&un.CheckinTime, &un.CheckoutTime,
		&un.Amenities, &un.Rules,
		&un.PhoneNumber, &un.Status,
		&un.CreatedAt, &un.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("unit not found")
		}
		return nil, fmt.Errorf("failed to get unit: %w", err)
	}

	images, err := u.getImages(ctx, un.ID)
	if err != nil {
		return nil, err
	}
	un.Images = images
	return &un, nil
}



func (u *UnitRepository) GetBySlug(ctx context.Context, slug string, tenantID uuid.UUID) (*unit.Unit, error) {
	var un unit.Unit
	query := `
		SELECT
			id, tenant_id,
			title, name, short_description, description, slug, type,
			guests, bedrooms, beds, bathrooms,
			location, latitude, longitude,
			apartment_name, house_number, floor, access_note,
			price_weekday, price_weekend,
			checkin_time, checkout_time,
			amenities, rules,
			phone_number, status,
			created_at, updated_at
		FROM units
		WHERE slug = $1 AND tenant_id = $2 AND status != 'deleted'
	`
	err := u.DbConnection.Pool.QueryRow(ctx, query, slug, tenantID).Scan(
		&un.ID, &un.TenantID,
		&un.Title, &un.Name, &un.ShortDescription, &un.Description, &un.Slug, &un.UnitType,
		&un.Guests, &un.Bedrooms, &un.Beds, &un.Bathrooms,
		&un.Location, &un.Latitude, &un.Longitude,
		&un.ApartmentName, &un.HouseNumber, &un.Floor, &un.AccessNote,
		&un.PriceWeekday, &un.PriceWeekend,
		&un.CheckinTime, &un.CheckoutTime,
		&un.Amenities, &un.Rules,
		&un.PhoneNumber, &un.Status,
		&un.CreatedAt, &un.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("unit not found")
		}
		return nil, fmt.Errorf("failed to get unit: %w", err)
	}

	images, err := u.getImages(ctx, un.ID)
	if err != nil {
		return nil, err
	}
	un.Images = images
	return &un, nil
}



func (u *UnitRepository) GetAll(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*unit.Unit, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	query := `
		SELECT
			id, tenant_id,
			title, name, short_description, description, slug, type,
			guests, bedrooms, beds, bathrooms,
			location, latitude, longitude,
			apartment_name, house_number, floor, access_note,
			price_weekday, price_weekend,
			checkin_time, checkout_time,
			amenities, rules,
			phone_number, status,
			created_at, updated_at
		FROM units
		WHERE tenant_id = $1 AND status != 'deleted'::unit_status
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := u.DbConnection.Pool.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch units: %w", err)
	}
	defer rows.Close()

	var units []*unit.Unit
	for rows.Next() {
		var un unit.Unit
		if err := rows.Scan(
			&un.ID, &un.TenantID,
			&un.Title, &un.Name, &un.ShortDescription, &un.Description, &un.Slug, &un.UnitType,
			&un.Guests, &un.Bedrooms, &un.Beds, &un.Bathrooms,
			&un.Location, &un.Latitude, &un.Longitude,
			&un.ApartmentName, &un.HouseNumber, &un.Floor, &un.AccessNote,
			&un.PriceWeekday, &un.PriceWeekend,
			&un.CheckinTime, &un.CheckoutTime,
			&un.Amenities, &un.Rules,
			&un.PhoneNumber, &un.Status,
			&un.CreatedAt, &un.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan unit: %w", err)
		}
		units = append(units, &un)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	for _, un := range units {
		images, err := u.getImages(ctx, un.ID)
		if err != nil {
			return nil, err
		}
		un.Images = images
	}
	return units, nil
}



func (u *UnitRepository) GetUnitByIdAndTenant(ctx context.Context, unitID, tenantID uuid.UUID) (*unit.Unit, error) {
	query := `
		SELECT
			id, tenant_id,
			title, name, short_description, description, slug, type,
			guests, bedrooms, beds, bathrooms,
			location, latitude, longitude,
			apartment_name, house_number, floor, access_note,
			price_weekday, price_weekend,
			checkin_time, checkout_time,
			amenities, rules,
			phone_number, status,
			created_at, updated_at
		FROM units
		WHERE id = $1 AND tenant_id = $2
	`
	un := &unit.Unit{}
	err := u.DbConnection.Pool.QueryRow(ctx, query, unitID, tenantID).Scan(
		&un.ID, &un.TenantID,
		&un.Title, &un.Name, &un.ShortDescription, &un.Description, &un.Slug, &un.UnitType,
		&un.Guests, &un.Bedrooms, &un.Beds, &un.Bathrooms,
		&un.Location, &un.Latitude, &un.Longitude,
		&un.ApartmentName, &un.HouseNumber, &un.Floor, &un.AccessNote,
		&un.PriceWeekday, &un.PriceWeekend,
		&un.CheckinTime, &un.CheckoutTime,
		&un.Amenities, &un.Rules,
		&un.PhoneNumber, &un.Status,
		&un.CreatedAt, &un.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("unit not found or access denied")
		}
		return nil, fmt.Errorf("failed to fetch unit: %w", err)
	}
	return un, nil
}


func (u *UnitRepository) GetUnitBySlugAndTenant(ctx context.Context, slug string, tenantID uuid.UUID) (*unit.Unit, error) {
	query := `
		SELECT
			id, tenant_id,
			title, name, short_description, description, slug, type,
			guests, bedrooms, beds, bathrooms,
			location, latitude, longitude,
			apartment_name, house_number, floor, access_note,
			price_weekday, price_weekend,
			checkin_time, checkout_time,
			amenities, rules,
			phone_number, status,
			created_at, updated_at
		FROM units
		WHERE slug = $1 AND tenant_id = $2
	`
	un := &unit.Unit{}
	err := u.DbConnection.Pool.QueryRow(ctx, query, slug, tenantID).Scan(
		&un.ID, &un.TenantID,
		&un.Title, &un.Name, &un.ShortDescription, &un.Description, &un.Slug, &un.UnitType,
		&un.Guests, &un.Bedrooms, &un.Beds, &un.Bathrooms,
		&un.Location, &un.Latitude, &un.Longitude,
		&un.ApartmentName, &un.HouseNumber, &un.Floor, &un.AccessNote,
		&un.PriceWeekday, &un.PriceWeekend,
		&un.CheckinTime, &un.CheckoutTime,
		&un.Amenities, &un.Rules,
		&un.PhoneNumber, &un.Status,
		&un.CreatedAt, &un.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("unit not found or access denied")
		}
		return nil, fmt.Errorf("failed to fetch unit: %w", err)
	}
	return un, nil
}


func (u *UnitRepository) Update(ctx context.Context, id, tenantID uuid.UUID, req unit.UpdateUnitRequest) (*unit.Unit, error) {
	query := `
		UPDATE units SET
			title             = COALESCE($1,  title),
			name              = COALESCE($2,  name),
			short_description = COALESCE($3,  short_description),
			description       = COALESCE($4,  description),
			slug              = COALESCE($5,  slug),
			type              = COALESCE($6,  type),
			guests            = COALESCE($7,  guests),
			bedrooms          = COALESCE($8,  bedrooms),
			beds              = COALESCE($9,  beds),
			bathrooms         = COALESCE($10, bathrooms),
			location          = COALESCE($11, location),
			latitude          = COALESCE($12, latitude),
			longitude         = COALESCE($13, longitude),
			apartment_name    = COALESCE($14, apartment_name),
			house_number      = COALESCE($15, house_number),
			floor             = COALESCE($16, floor),
			access_note       = COALESCE($17, access_note),
			price_weekday     = COALESCE($18, price_weekday),
			price_weekend     = COALESCE($19, price_weekend),
			checkin_time      = COALESCE($20, checkin_time),
			checkout_time     = COALESCE($21, checkout_time),
			amenities         = COALESCE($22, amenities),
			rules             = COALESCE($23, rules),
			phone_number      = COALESCE($24, phone_number),
			status            = COALESCE($25, status),
			updated_at        = NOW()
		WHERE id = $26 AND tenant_id = $27 AND status != 'deleted'
		RETURNING
			id, tenant_id,
			title, name, short_description, description, slug, type,
			guests, bedrooms, beds, bathrooms,
			location, latitude, longitude,
			apartment_name, house_number, floor, access_note,
			price_weekday, price_weekend,
			checkin_time, checkout_time,
			amenities, rules,
			phone_number, status,
			created_at, updated_at
	`
	var un unit.Unit
	err := u.DbConnection.Pool.QueryRow(ctx, query,
		req.Title, req.Name, req.ShortDescription, req.Description, req.Slug, req.UnitType,
		req.Guests, req.Bedrooms, req.Beds, req.Bathrooms,
		req.Location, req.Latitude, req.Longitude,
		req.ApartmentName, req.HouseNumber, req.Floor, req.AccessNote,
		req.PriceWeekday, req.PriceWeekend,
		req.CheckinTime, req.CheckoutTime,
		req.Amenities, req.Rules,
		req.PhoneNumber, req.Status,
		id, tenantID,
	).Scan(
		&un.ID, &un.TenantID,
		&un.Title, &un.Name, &un.ShortDescription, &un.Description, &un.Slug, &un.UnitType,
		&un.Guests, &un.Bedrooms, &un.Beds, &un.Bathrooms,
		&un.Location, &un.Latitude, &un.Longitude,
		&un.ApartmentName, &un.HouseNumber, &un.Floor, &un.AccessNote,
		&un.PriceWeekday, &un.PriceWeekend,
		&un.CheckinTime, &un.CheckoutTime,
		&un.Amenities, &un.Rules,
		&un.PhoneNumber, &un.Status,
		&un.CreatedAt, &un.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("unit not found")
		}
		return nil, fmt.Errorf("failed to update unit: %w", err)
	}
	return &un, nil
}


func (u *UnitRepository) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
	query := `
		UPDATE units SET status = 'deleted', updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
	`
	tag, err := u.DbConnection.Pool.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete unit: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("unit not found")
	}
	return nil
}



func (u *UnitRepository) AddImage(ctx context.Context, image *unit.UnitImage) (*unit.UnitImage, error) {
	query := `
		INSERT INTO unit_images (id, unit_id, url, image_type, sort_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, unit_id, url, image_type, sort_order
	`
	err := u.DbConnection.Pool.QueryRow(ctx, query,
		image.ID, image.UnitID, image.URL, image.ImageType, image.SortOrder,
	).Scan(&image.ID, &image.UnitID, &image.URL, &image.ImageType, &image.SortOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to add image: %w", err)
	}
	return image, nil
}

func (u *UnitRepository) RemoveImage(ctx context.Context, imageID, tenantID uuid.UUID) error {
	query := `
		DELETE FROM unit_images
		WHERE id = $1
		  AND unit_id IN (SELECT id FROM units WHERE tenant_id = $2)
	`
	tag, err := u.DbConnection.Pool.Exec(ctx, query, imageID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to remove image: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("image not found")
	}
	return nil
}

func (u *UnitRepository) getImages(ctx context.Context, unitID uuid.UUID) ([]unit.UnitImage, error) {
	rows, err := u.DbConnection.Pool.Query(ctx,
		`SELECT id, unit_id, url, image_type, sort_order
		 FROM unit_images WHERE unit_id = $1 ORDER BY sort_order ASC`,
		unitID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch images: %w", err)
	}
	defer rows.Close()

	var images []unit.UnitImage
	for rows.Next() {
		var img unit.UnitImage
		if err := rows.Scan(&img.ID, &img.UnitID, &img.URL, &img.ImageType, &img.SortOrder); err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}
		images = append(images, img)
	}
	return images, nil
}

func (u *UnitRepository) GetImagesByUnitID(ctx context.Context, unitID uuid.UUID) ([]*unit.UnitImage, error) {
	query := `
		SELECT id, unit_id, url, image_type, sort_order
		FROM unit_images
		WHERE unit_id = $1
		ORDER BY sort_order ASC
	`
	rows, err := u.DbConnection.Pool.Query(ctx, query, unitID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch unit images: %w", err)
	}
	defer rows.Close()

	var images []*unit.UnitImage
	for rows.Next() {
		img := &unit.UnitImage{}
		if err := rows.Scan(&img.ID, &img.UnitID, &img.URL, &img.ImageType, &img.SortOrder); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, nil
}


func (u *UnitRepository) GenerateUniqueSlug(ctx context.Context, tenantID uuid.UUID, name string) (string, error) {
	baseSlug := slugify(name)
	slug := baseSlug
	count := 1

	for {
		var exists bool
		err := u.DbConnection.Pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM units WHERE tenant_id = $1 AND slug = $2)`,
			tenantID, slug,
		).Scan(&exists)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, count)
		count++
	}
}

func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, " ", "-")
	re := regexp.MustCompile(`[^a-z0-9\-]`)
	return re.ReplaceAllString(s, "")
}

var _ unit.UnitRepository = (*UnitRepository)(nil)