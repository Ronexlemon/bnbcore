package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ronexlemon/bnbcore/internal/domain/unit"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/db"
)

type UnitRepository struct {
    DbConnection *db.PostgresConn
}

func NewUnitRepository(dbconn *db.PostgresConn) (*UnitRepository,error) {
    return &UnitRepository{DbConnection: dbconn},nil
}


func (u *UnitRepository) Create(ctx context.Context, un *unit.Unit) (*unit.Unit, error) {
   
    query := `
        INSERT INTO units (id, tenant_id, title, description,name,type, price_per_night, location, latitude, longitude, status,amenities, rules,adults,children, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9,$10,$11,$12,$13,$14,$15, NOW(), NOW())
        RETURNING created_at, updated_at
    `
    err := u.DbConnection.Pool.QueryRow(ctx, query,
        un.ID,
        un.TenantID,
        un.Title,
        un.Description,
		un.Name,
		un.UnitType,
        un.PricePerNight,
        un.Location,
        un.Latitude,
        un.Longitude,
        un.Status,
        un.Amenities,
        un.Rules,
        un.Adults,
        un.Children,
    ).Scan(&un.CreatedAt, &un.UpdatedAt)
    if err != nil {
        return nil, fmt.Errorf("failed to create unit: %w", err)
    }
    return un, nil
}


func (u *UnitRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*unit.Unit, error) {
    var un unit.Unit

    query := `
        SELECT id, tenant_id, title, description,name,type, price_per_night, location,
               latitude, longitude, status,amenities, rules,adults,children, created_at, updated_at
        FROM units
        WHERE id = $1
          AND tenant_id = $2
          AND status != 'deleted'
    `
    err := u.DbConnection.Pool.QueryRow(ctx, query, id, tenantID).Scan(
        &un.ID,
        &un.TenantID,
        &un.Title,
        &un.Description,
		&un.Name,
		&un.UnitType,
        &un.PricePerNight,
        &un.Location,
        &un.Latitude,
        &un.Longitude,
        &un.Status,
        &un.Amenities,
        &un.Rules,
        &un.Adults,
        &un.Children,
        &un.CreatedAt,
        &un.UpdatedAt,
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
        SELECT id, tenant_id, title, description, name, type, price_per_night, location,
               latitude, longitude, status, amenities, rules, adults, children, created_at, updated_at
        FROM units
        WHERE tenant_id = $1
       AND status != 'deleted'::unit_status
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `
    
    rows, err := u.DbConnection.Pool.Query(ctx, query, tenantID, limit, offset)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch paginated units: %w", err)
    }
    defer rows.Close()

    var units []*unit.Unit
    for rows.Next() {
        var un unit.Unit
        if err := rows.Scan(
            &un.ID,
            &un.TenantID,
            &un.Title,
            &un.Description,
            &un.Name,
            &un.UnitType,
            &un.PricePerNight,
            &un.Location,
            &un.Latitude,
            &un.Longitude,
            &un.Status,
            &un.Amenities,
            &un.Rules,
            &un.Adults,
            &un.Children,
            &un.CreatedAt,
            &un.UpdatedAt,
        ); err != nil {
            return nil, fmt.Errorf("failed to scan unit: %w", err)
        }
        units = append(units, &un)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error during rows iteration: %w", err)
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

func (u *UnitRepository) GetUnitByIdAndTenant(ctx context.Context, unitID uuid.UUID, tenantID uuid.UUID) (*unit.Unit, error) {
	query := `
		SELECT id, tenant_id, title, description,price_per_night,amenities, rules,adults,children, created_at, updated_at
		FROM units
		WHERE id = $1 AND tenant_id = $2
	`

	un := &unit.Unit{}

	err := u.DbConnection.Pool.QueryRow(ctx, query, unitID, tenantID).Scan(
		&un.ID,
		&un.TenantID,
		&un.Title,
		&un.Description,
		&un.PricePerNight,
        &un.Amenities,
        &un.Rules,
        &un.Adults,
        &un.Children,
		&un.CreatedAt,
		&un.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("unit not found or access denied")
		}
		return nil, fmt.Errorf("failed to fetch unit by id and tenant: %w", err)
	}

	return un, nil
}


func (u *UnitRepository) Update(ctx context.Context, id, tenantID uuid.UUID, req unit.UpdateUnitRequest) (*unit.Unit, error) {
    query := `
        UPDATE units SET
            title           = COALESCE($1, title),
            description     = COALESCE($2, description),
            name            = COALESCE($3, name),
            type            = COALESCE($4, type),
            price_per_night  = COALESCE($5, price_per_night),
            location        = COALESCE($6, location),
            latitude        = COALESCE($7, latitude),
            longitude       = COALESCE($8, longitude),
            status          = COALESCE($9, status),
            amenities       = COALESCE($10, amenities), 
            rules           = COALESCE($11, rules),
             adults           = COALESCE($12, adults),
              children           = COALESCE($13, children),     
            updated_at      = NOW()
        WHERE id = $14
          AND tenant_id = $15
          AND status != 'deleted'
        RETURNING id, tenant_id, title, description, name, type, price_per_night, location,
                  latitude, longitude, status, amenities, rules,adults,children, created_at, updated_at
    `
    var un unit.Unit
    err := u.DbConnection.Pool.QueryRow(ctx, query,
        req.Title,
        req.Description,
        req.Name,
        req.UnitType,
        req.PricePerNight,
        req.Location,
        req.Latitude,
        req.Longitude,
        req.Status,
        req.Amenities, 
        req.Rules,    
        req.Adults,
        req.Children, 
        id,            
        tenantID,     
    ).Scan(
        &un.ID,
        &un.TenantID,
        &un.Title,
        &un.Description,
        &un.Name,
        &un.UnitType,
        &un.PricePerNight,
        &un.Location,
        &un.Latitude,
        &un.Longitude,
        &un.Status,
        &un.Amenities, 
        &un.Rules,   
        &un.Adults,
        &un.Children,  
        &un.CreatedAt,
        &un.UpdatedAt,
    )
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, errors.New("unit not found")
        }
        return nil, fmt.Errorf("failed to update unit record: %w", err)
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
        INSERT INTO unit_images (id, unit_id, url,image_type)
        VALUES ($1, $2, $3, $4)
        RETURNING id, unit_id, url,image_type
    `
    err := u.DbConnection.Pool.QueryRow(ctx, query,
        image.ID,
        image.UnitID,
        image.URL,
		image.ImageType,
    ).Scan(&image.ID, &image.UnitID, &image.URL, &image.ImageType)
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
        `SELECT id, unit_id, url,image_type FROM unit_images WHERE unit_id = $1`,
        unitID,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to fetch images: %w", err)
    }
    defer rows.Close()

    var images []unit.UnitImage
    for rows.Next() {
        var img unit.UnitImage
        if err := rows.Scan(&img.ID, &img.UnitID, &img.URL); err != nil {
            return nil, fmt.Errorf("failed to scan image: %w", err)
        }
        images = append(images, img)
    }
    return images, nil
}

func (u *UnitRepository) GetImagesByUnitID(ctx context.Context, unitID uuid.UUID) ([]*unit.UnitImage, error) {
    query := `
        SELECT id, unit_id, url, image_type 
        FROM unit_images 
        WHERE unit_id = $1
        ORDER BY created_at DESC
    `
    rows, err := u.DbConnection.Pool.Query(ctx, query, unitID)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch unit images: %w", err)
    }
    defer rows.Close()

    images := []*unit.UnitImage{}
    for rows.Next() {
        img := &unit.UnitImage{}
        err := rows.Scan(&img.ID, &img.UnitID, &img.URL, &img.ImageType)
        if err != nil {
            return nil, err
        }
        images = append(images, img)
    }

    return images, nil
}
var _ unit.UnitRepository = (*UnitRepository)(nil)