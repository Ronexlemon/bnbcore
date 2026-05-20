package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ronexlemon/bnbcore/internal/infrastructure/db"
)


type TenantRepository struct{
	DbConnection *db.PostgresConn
}


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
	passwordHash string,
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
	`, tenantID, email, passwordHash)

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

func (p *TenantRepository) UpdateTenant(ctx context.Context, name, subdomain string) error {
	_, err := p.DbConnection.Pool.Exec(ctx,
		`INSERT INTO tenants (id, name, subdomain)
		 VALUES (gen_random_uuid(), $1, $2)`,
		name, subdomain,
	)
	return err
}