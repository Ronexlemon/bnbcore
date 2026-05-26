package repository

import (
    "context"
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
        INSERT INTO units (id, tenant_id, title, description, price_per_night, location, latitude, longitude, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
        RETURNING created_at, updated_at
    `
    err := u.DbConnection.Pool.QueryRow(ctx, query,
        un.ID,
        un.TenantID,
        un.Title,
        un.Description,
        un.PricePerNight,
        un.Location,
        un.Latitude,
        un.Longitude,
        un.Status,
    ).Scan(&un.CreatedAt, &un.UpdatedAt)
    if err != nil {
        return nil, fmt.Errorf("failed to create unit: %w", err)
    }
    return un, nil
}


func (u *UnitRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*unit.Unit, error) {
    var un unit.Unit

    query := `
        SELECT id, tenant_id, title, description, price_per_night, location,
               latitude, longitude, status, created_at, updated_at
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
        &un.PricePerNight,
        &un.Location,
        &un.Latitude,
        &un.Longitude,
        &un.Status,
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



func (u *UnitRepository) GetAll(ctx context.Context, tenantID uuid.UUID) ([]*unit.Unit, error) {
    query := `
        SELECT id, tenant_id, title, description, price_per_night, location,
               latitude, longitude, status, created_at, updated_at
        FROM units
        WHERE tenant_id = $1
          AND status != 'deleted'
        ORDER BY created_at DESC
    `
    rows, err := u.DbConnection.Pool.Query(ctx, query, tenantID)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch units: %w", err)
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
            &un.PricePerNight,
            &un.Location,
            &un.Latitude,
            &un.Longitude,
            &un.Status,
            &un.CreatedAt,
            &un.UpdatedAt,
        ); err != nil {
            return nil, fmt.Errorf("failed to scan unit: %w", err)
        }
        units = append(units, &un)
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


func (u *UnitRepository) Update(ctx context.Context, id, tenantID uuid.UUID, req unit.UpdateUnitRequest) (*unit.Unit, error) {
    query := `
        UPDATE units SET
            title          = COALESCE($1, title),
            description    = COALESCE($2, description),
            price_per_night = COALESCE($3, price_per_night),
            location       = COALESCE($4, location),
            latitude       = COALESCE($5, latitude),
            longitude      = COALESCE($6, longitude),
            status         = COALESCE($7, status),
            updated_at     = NOW()
        WHERE id = $8
          AND tenant_id = $9
          AND status != 'deleted'
        RETURNING id, tenant_id, title, description, price_per_night, location,
                  latitude, longitude, status, created_at, updated_at
    `
    var un unit.Unit
    err := u.DbConnection.Pool.QueryRow(ctx, query,
        req.Title,
        req.Description,
        req.PricePerNight,
        req.Location,
        req.Latitude,
        req.Longitude,
        req.Status,
        id,
        tenantID,
    ).Scan(
        &un.ID,
        &un.TenantID,
        &un.Title,
        &un.Description,
        &un.PricePerNight,
        &un.Location,
        &un.Latitude,
        &un.Longitude,
        &un.Status,
        &un.CreatedAt,
        &un.UpdatedAt,
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
        INSERT INTO unit_images (id, unit_id, url)
        VALUES ($1, $2, $3)
        RETURNING id, unit_id, url
    `
    err := u.DbConnection.Pool.QueryRow(ctx, query,
        image.ID,
        image.UnitID,
        image.URL,
    ).Scan(&image.ID, &image.UnitID, &image.URL)
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
        `SELECT id, unit_id, url FROM unit_images WHERE unit_id = $1`,
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

var _ unit.UnitRepository = (*UnitRepository)(nil)