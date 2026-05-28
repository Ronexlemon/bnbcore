package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/db"
)

type DBQuerier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
}

type TenantRepository struct {
	DbConnection *db.PostgresConn
}

var (
	ErrTenantNotFound       = errors.New("tenant not found")
	ErrSubdomainTaken       = errors.New("subdomain is already taken")
	ErrUserAlreadyOnboarded = errors.New("user already belongs to a tenant")
)

func NewTenantRepository(dbconnect *db.PostgresConn) (*TenantRepository, error) {
	if dbconnect == nil {
		return nil, fmt.Errorf("db connection required")
	}
	return &TenantRepository{DbConnection: dbconnect}, nil
}



func (t *TenantRepository) CreateTenant(ctx context.Context,ShopName ,ShopDescription, subdomain string, userID uuid.UUID) error {
	_, err := t.DbConnection.Pool.Exec(ctx, `
		INSERT INTO tenants (id, user_id,name,shop_description, subdomain, status, trial_ends_at, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3,$4, 'trial', NOW() + INTERVAL '14 days', NOW())
	`, userID,ShopName ,ShopDescription, subdomain)
	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}
	return nil
}


func (t *TenantRepository) FindByID(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
	var tn tenant.Tenant

	err := t.DbConnection.Pool.QueryRow(ctx, `
		SELECT id, user_id,name,shop_description, subdomain, status, trial_ends_at, created_at
		FROM tenants
		WHERE id = $1
	`, id).Scan(
		&tn.ID,
		&tn.UserID,
		&tn.ShopName,
		&tn.ShopDescription,
		&tn.Subdomain,
		&tn.Status,
		&tn.TrialEndsAt,
		&tn.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to find tenant by id: %w", err)
	}
	return &tn, nil
}

func (t *TenantRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*tenant.Tenant, error) {
	var tn tenant.Tenant

	err := t.DbConnection.Pool.QueryRow(ctx, `
		SELECT id, user_id,name, shop_description, subdomain, status, trial_ends_at, created_at
		FROM tenants
		WHERE user_id = $1
	`, userID).Scan(
		&tn.ID,
		&tn.UserID,
		&tn.ShopName,
		&tn.ShopDescription,
		&tn.Subdomain,
		&tn.Status,
		&tn.TrialEndsAt,
		&tn.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to find tenant by user id: %w", err)
	}
	return &tn, nil
}

func (t *TenantRepository) FindBySubdomain(ctx context.Context, sub string) (*tenant.Tenant, error) {
	var tn tenant.Tenant

	err := t.DbConnection.Pool.QueryRow(ctx, `
		SELECT id, user_id,name, shop_description, subdomain, status, trial_ends_at, created_at
		FROM tenants
		WHERE subdomain = $1
	`, sub).Scan(
		&tn.ID,
		&tn.UserID,
		&tn.ShopName,
		&tn.ShopDescription,
		&tn.Subdomain,
		&tn.Status,
		&tn.TrialEndsAt,
		&tn.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to find tenant by subdomain: %w", err)
	}
	return &tn, nil
}

func (t *TenantRepository) GetAll(ctx context.Context) ([]*tenant.Tenant, error) {
	rows, err := t.DbConnection.Pool.Query(ctx, `
		SELECT id, user_id,name, shop_description, subdomain, status, trial_ends_at, created_at
		FROM tenants
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*tenant.Tenant
	for rows.Next() {
		var tn tenant.Tenant
		if err := rows.Scan(
			&tn.ID,
			&tn.UserID,
			&tn.ShopName,
			&tn.ShopDescription,
			&tn.Subdomain,
			&tn.Status,
			&tn.TrialEndsAt,
			&tn.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenants = append(tenants, &tn)
	}
	return tenants, nil
}

// ── Update ────────────────────────────────────────────────────────────────────

func (t *TenantRepository) UpdateTenant(ctx context.Context, id uuid.UUID, req tenant.UpdateTenantRequest) (*tenant.Tenant, error) {
	var tn tenant.Tenant

	err := t.DbConnection.Pool.QueryRow(ctx, `
		UPDATE tenants SET
			shop_description     = COALESCE($1, shop_description),
			subdomain = COALESCE($2, subdomain),
			name = COALESCE($3, name),
			status    = COALESCE($4, status)
		WHERE id = $4
		RETURNING id, user_id,name,shop_description, subdomain, status, trial_ends_at, created_at
	`, req.ShopDescription, req.Subdomain,req.ShopName ,req.Status, id).Scan(
		&tn.ID,
		&tn.UserID,
		&tn.ShopName,
		&tn.ShopDescription,
		&tn.Subdomain,
		&tn.Status,
		&tn.TrialEndsAt,
		&tn.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}
	return &tn, nil
}


func (t *TenantRepository) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	tag, err := t.DbConnection.Pool.Exec(ctx,
		`DELETE FROM tenants WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTenantNotFound
	}
	return nil
}


func (t *TenantRepository) SubdomainExists(ctx context.Context, subdomain string) (bool, error) {
	var exists bool
	err := t.DbConnection.Pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM tenants WHERE subdomain = $1)`,
		subdomain,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check subdomain: %w", err)
	}
	return exists, nil
}

func (t *TenantRepository) TenantExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := t.DbConnection.Pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM tenants WHERE id = $1)`,
		id,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check tenant: %w", err)
	}
	return exists, nil
}

var _ tenant.Repository = (*TenantRepository)(nil)