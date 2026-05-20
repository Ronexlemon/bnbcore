package repository

import (
	"context"
	"fmt"

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


func (p *TenantRepository) CreateTenant(ctx context.Context, name, subdomain string) error {
	_, err := p.DbConnection.Pool.Exec(ctx,
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