package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	rs "github.com/ronexlemon/bnbcore/internal/domain/services"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/db"
)

type UnitServiceRepository struct {
    DbConnection *db.PostgresConn
}

func NewUnitServiceRepository(dbconn *db.PostgresConn) (*UnitServiceRepository,error) {
	if dbconn ==nil{
		return nil,fmt.Errorf("Db connection failed")
	}
    return &UnitServiceRepository{DbConnection: dbconn},nil
}

func (r *UnitServiceRepository) Create(ctx context.Context, s *rs.UnitService) (*rs.UnitService, error) {
    query := `
        INSERT INTO room_services (id, unit_id, tenant_id, agent_name, mobile, email, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
        RETURNING created_at, updated_at
    `
    err := r.DbConnection.Pool.QueryRow(ctx, query,
        s.ID,
        s.UnitID,
        s.TenantID,
        s.AgentName,
        s.Mobile,
        s.Email,
        s.IsActive,
    ).Scan(&s.CreatedAt, &s.UpdatedAt)
    if err != nil {
        return nil, fmt.Errorf("failed to create room service: %w", err)
    }
    return s, nil
}

func (r *UnitServiceRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*rs.UnitService, error) {
    var s rs.UnitService
    query := `
        SELECT id, unit_id, tenant_id, agent_name, mobile, email, is_active, created_at, updated_at
        FROM room_services
        WHERE id = $1
          AND tenant_id = $2
    `
    err := r.DbConnection.Pool.QueryRow(ctx, query, id, tenantID).Scan(
        &s.ID,
        &s.UnitID,
        &s.TenantID,
        &s.AgentName,
        &s.Mobile,
        &s.Email,
        &s.IsActive,
        &s.CreatedAt,
        &s.UpdatedAt,
    )
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, errors.New("room service not found")
        }
        return nil, fmt.Errorf("failed to get room service: %w", err)
    }
    return &s, nil
}

func (r *UnitServiceRepository) GetByUnit(ctx context.Context, unitID, tenantID uuid.UUID) ([]*rs.UnitService, error) {
    query := `
        SELECT id, unit_id, tenant_id, agent_name, mobile, email, is_active, created_at, updated_at
        FROM room_services
        WHERE unit_id = $1
          AND tenant_id = $2
        ORDER BY created_at DESC
    `
    rows, err := r.DbConnection.Pool.Query(ctx, query, unitID, tenantID)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch room services: %w", err)
    }
    defer rows.Close()

    var services []*rs.UnitService
    for rows.Next() {
        var s rs.UnitService
        if err := rows.Scan(
            &s.ID,
            &s.UnitID,
            &s.TenantID,
            &s.AgentName,
            &s.Mobile,
            &s.Email,
            &s.IsActive,
            &s.CreatedAt,
            &s.UpdatedAt,
        ); err != nil {
            return nil, fmt.Errorf("failed to scan room service: %w", err)
        }
        services = append(services, &s)
    }
    return services, nil
}

func (r *UnitServiceRepository) Update(ctx context.Context, id, tenantID uuid.UUID, req rs.UpdateUnitServiceRequest) (*rs.UnitService, error) {
    query := `
        UPDATE room_services SET
            agent_name = COALESCE($1, agent_name),
            mobile     = COALESCE($2, mobile),
            email      = COALESCE($3, email),
            is_active  = COALESCE($4, is_active),
            updated_at = NOW()
        WHERE id = $5
          AND tenant_id = $6
        RETURNING id, unit_id, tenant_id, agent_name, mobile, email, is_active, created_at, updated_at
    `
    var s rs.UnitService
    err := r.DbConnection.Pool.QueryRow(ctx, query,
        req.AgentName,
        req.Mobile,
        req.Email,
        req.IsActive,
        id,
        tenantID,
    ).Scan(
        &s.ID,
        &s.UnitID,
        &s.TenantID,
        &s.AgentName,
        &s.Mobile,
        &s.Email,
        &s.IsActive,
        &s.CreatedAt,
        &s.UpdatedAt,
    )
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, errors.New("room service not found")
        }
        return nil, fmt.Errorf("failed to update room service: %w", err)
    }
    return &s, nil
}

func (r *UnitServiceRepository) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
    tag, err := r.DbConnection.Pool.Exec(ctx,
        `DELETE FROM room_services WHERE id = $1 AND tenant_id = $2`,
        id, tenantID,
    )
    if err != nil {
        return fmt.Errorf("failed to delete room service: %w", err)
    }
    if tag.RowsAffected() == 0 {
        return errors.New("room service not found")
    }
    return nil
}

var _ rs.RoomServiceRepository = (*UnitServiceRepository)(nil)