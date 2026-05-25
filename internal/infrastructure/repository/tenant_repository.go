package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/db"
	uuid "github.com/satori/go.uuid"
)



type DBQuerier interface {
    Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
    QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
    Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
}



type TenantRepository struct{
	DbConnection *db.PostgresConn
}


var (
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrSubdomainTaken      = errors.New("subdomain is already taken")
	ErrUserAlreadyOnboarded = errors.New("user already belongs to a tenant")
)

func NewTenantRepository(dbconnect *db.PostgresConn)(*TenantRepository,error){
	if dbconnect ==nil{
		return nil,fmt.Errorf("Db Connection required")
	}
	return&TenantRepository{
		DbConnection: dbconnect,
	},nil
}


func(t *TenantRepository)  RegisterTenantWithUser(
	ctx context.Context,
	tenantName string,
	subdomain string,
	email string,
	password string,
) error {

	tx, err := t.DbConnection.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	var tenantID string
	trialEndsAt := time.Now().AddDate(0, 0, 14)

	err = tx.QueryRow(ctx, `
		INSERT INTO tenants (id, name, subdomain,trial_ends_at)
		VALUES (gen_random_uuid(), $1, $2,$3)
		RETURNING id
	`, tenantName, subdomain,trialEndsAt).Scan(&tenantID)

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO users (
			id,
			tenant_id,
			email,
			password_hash,
			role
		)
		VALUES (
			gen_random_uuid(),
			$1,
			$2,
			$3,
			'owner'
		)
	`, tenantID, email, password)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
func (t *TenantRepository) CreateTenant(ctx context.Context, name, subdomain string) error {
	_, err := t.DbConnection.Pool.Exec(ctx,
		`INSERT INTO tenants (id, name, subdomain)
		 VALUES (gen_random_uuid(), $1, $2)`,
		name, subdomain,
	)
	return err
}

func (t *TenantRepository) SubdomainExists(ctx context.Context, subdomain string) (bool, error) {
	var exists bool
	const q = `SELECT EXISTS(SELECT 1 FROM tenants WHERE subdomain = $1)`
	err := t.DbConnection.Pool.QueryRow(ctx, q, subdomain).Scan(&exists)
	return exists, err
}

func (t *TenantRepository) UpdateTenant(ctx context.Context, name, subdomain string) error {
	_, err := t.DbConnection.Pool.Exec(ctx,
		`INSERT INTO tenants (id, name, subdomain)
		 VALUES (gen_random_uuid(), $1, $2)`,
		name, subdomain,
	)
	return err
}

func (t *TenantRepository) FindByID(ctx context.Context, id uuid.UUID) (*tenant.Tenant, error) {
	var tenat_details tenant.Tenant
	const q = `SELECT id, name, subdomain, status, trial_ends_at, created_at
	           FROM tenants WHERE id = $1`
	err := t.DbConnection.Pool.QueryRow(ctx, q, id).Scan(&tenat_details)
	
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTenantNotFound
	}
	return &tenat_details, err
}

func  createTenantTx(ctx context.Context, db DBQuerier, tenant_details *tenant.Tenant) error {
    const q = `
        INSERT INTO tenants (id, name, subdomain, status, trial_ends_at, created_at)
        VALUES ($1, $2, $3, $4, $5, NOW())`
 
    _, err := db.Exec(ctx, q, 
		tenant_details.ID,
        tenant_details.Name, 
        tenant_details.Subdomain, 
        tenant_details.Status, 
        tenant_details.TrialEndsAt,
    )
    
    return err
}

func linkUserToTenantTx(
    ctx context.Context, 
    db DBQuerier, 
    userID uuid.UUID, 
    tenantID uuid.UUID, 
    role string,
) error {
    const q = `UPDATE users SET tenant_id = $1, role = $2 WHERE id = $3`
    
    // In pgx, Exec takes the context directly as the first argument
    cmdTag, err := db.Exec(ctx, q, tenantID, role, userID)
    if err != nil {
        return fmt.Errorf("link user to tenant: %w", err)
    }
    
    if cmdTag.RowsAffected() == 0 {
        return ErrTenantNotFound
    }
    
    return nil
}
func (t *TenantRepository) CreateTenantWithOwner(
    ctx context.Context,
    userID uuid.UUID,
    tenantName string,
    subdomain string,
) (*tenant.OnboardResult, error) {
    
    userRepo := &UserRepository{t.DbConnection}

    user, err := userRepo.GetUserByID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("load user: %w", err)
    }
    if user.TenantID !=nil { 
        return nil, ErrUserAlreadyOnboarded
    }
 
    tx, err := t.DbConnection.Pool.Begin(ctx) 
    if err != nil {
        return nil, fmt.Errorf("begin tx: %w", err)
    }
    
   
    defer func() {
        if err != nil {
            _ = tx.Rollback(ctx)
        }
    }()
 
    // 3. Step 1: Create Tenant (Now bound to the transaction)
    newTenant := &tenant.Tenant{
		ID:  uuid.NewV4(),
        Name:        tenantName,
        Subdomain:   subdomain,
        Status:      tenant.StatusTrial,
        TrialEndsAt: time.Now().AddDate(0, 0, tenant.TrialDays),
    }
 
    // FIXED: Passing `tx` instead of letting it run on the Pool implicitly
    err = createTenantTx(ctx, tx, newTenant) 
    if err != nil {
        return nil, fmt.Errorf("create tenant: %w", err)
    }
 
 
    if err := linkUserToTenantTx(ctx, tx, userID, newTenant.ID, user.Role); err != nil {
        return nil, fmt.Errorf("link user: %w", err)
    }
 
    // 5. Commit Transaction
    if err = tx.Commit(ctx); err != nil {
        return nil, fmt.Errorf("commit: %w", err)
    }
 
    // 6. Reload user
    updatedUser, err := userRepo.GetUserByID(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("reload user after commit: %w", err)
    }
 
    return &tenant.OnboardResult{
        Tenant: newTenant,
        User:   updatedUser,
    }, nil
}