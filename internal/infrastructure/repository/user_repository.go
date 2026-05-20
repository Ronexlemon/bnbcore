package repository

import (
	"context"
	"fmt"

	"github.com/ronexlemon/bnbcore/internal/infrastructure/db"
)

type UserRepository struct{
	DBConnection *db.PostgresConn
}


func NewUserRepository(dbconnect *db.PostgresConn)(*UserRepository,error){
	if dbconnect ==nil{
		return nil,fmt.Errorf("Db Connection required")
	}

	return &UserRepository{
		DBConnection: dbconnect,
	},nil
}


func (u *UserRepository) Register(ctx context.Context,tenantID string,email string,password string,
) error {
	_, err := u.DBConnection.Pool.Exec(
		ctx,`INSERT INTO users (id,tenant_id,email,password_hash,role)
		VALUES (gen_random_uuid(),$1,$2,$3,'owner')`,tenantID,email,password,
	)

	if err != nil {
		return err
	}

	return nil
}