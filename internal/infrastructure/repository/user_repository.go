package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/ronexlemon/bnbcore/internal/domain/user"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/db"
	"github.com/satori/go.uuid"
	"google.golang.org/api/idtoken"
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


func (u *UserRepository) Register(ctx context.Context,tenantID uuid.UUID,email string,password string,
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

func (u *UserRepository) Login(ctx context.Context,email,password string)(*user.User,error){
	var user user.User
	query := `SELECT id, tenant_id, email, password_hash, role FROM users WHERE email=$1`
	err:=u.DBConnection.Pool.QueryRow(ctx,query,email).Scan(
		&user.ID,
		&user.TenantID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	//bycrypt compare passwrds hash

	return &user,nil
	
}


func (u *UserRepository)StoreRefreshToken(ctx context.Context,userID uuid.UUID,refreshTokenHash string,createdAt time.Time,isRevoked bool,expiresAt time.Time)error{
	_, err := u.DBConnection.Pool.Exec(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, is_revoked, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)
	`, 
		userID,          
		refreshTokenHash, 
		expiresAt,        
		isRevoked,        
		createdAt,        
	)
	
	if err != nil {
		return err
	}

	return nil
}



func (u *UserRepository)GetUserByID(ctx context.Context,userID uuid.UUID)(*user.User,error){
	var user user.User
	query:=`SELECT id,email,role,tenant_id FROM users WHERE id=$1`

	err:=u.DBConnection.Pool.QueryRow(ctx,query,userID).Scan(
		&user.ID,
		&user.Email,
		&user.Role,
		&user.TenantID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}


	return &user,nil

}

func (u *UserRepository) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	var user user.User
	
	query := `SELECT id, email, role, tenant_id, password_hash FROM users WHERE email = $1`

	err := u.DBConnection.Pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Role,
		&user.TenantID,
		&user.PasswordHash, 
	)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	return &user, nil
}

func (u *UserRepository) UpdatePasswordHash(ctx context.Context, userID uuid.UUID, newHash string) error {
	query := `UPDATE users SET password_hash = $1 WHERE id = $2`

	
	result, err := u.DBConnection.Pool.Exec(ctx, query, newHash, userID)
	if err != nil {
		return fmt.Errorf("failed to update password hash for rotation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no user found with ID %s to update password hash", userID)
	}

	return nil
}

func (u *UserRepository) GetRefreshToken(ctx context.Context, refreshToken string)(*user.REFRESHTOKEN,error){
	var refresh user.REFRESHTOKEN

	query:=`SELECT id,user_id,token_hash,expires_at,is_revoked,created_at FROM refresh_tokens WHERE token_hash=$1`

	err:=u.DBConnection.Pool.QueryRow(ctx,query,refreshToken).Scan(
		&refresh.ID,
		&refresh.UserID,
		&refresh.ExpiresAt,
		&refresh.IsRevoked,
		&refresh.CreateAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invalid refresh token")
		}
		return nil, err
	}


	return  &refresh,nil

}

func (u *UserRepository) LoginWithGoogle(ctx context.Context, googleClientID string, req user.GoogleLoginRequest) (*user.User, error) {
	
	payload, err := idtoken.Validate(ctx, req.Credential, googleClientID)
	if err != nil {
		return nil, fmt.Errorf("invalid google token: %w", err)
	}

	email, ok := payload.Claims["email"].(string)
	if !ok {
		return nil, errors.New("google token missing email claim")
	}

	// Step 2: Check if user already exists for this tenant + email pair
	var existing user.User
	selectQuery := `
		SELECT id, email, password_hash, role, is_active, created_at
		FROM users
		WHERE  email = $2`

	err = u.DBConnection.Pool.QueryRow(ctx, selectQuery, email).Scan(
		&existing.ID,
		&existing.TenantID,
		&existing.Email,
		&existing.PasswordHash,
		&existing.Role,
		&existing.IsActive,
		&existing.CreatedAt,
	)

	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	
	newUser := &user.User{
		ID:           uuid.NewV4(),
		Email:        email,
		PasswordHash: "OAUTH_EXTERNAL_ACCOUNT", 
		Role:         "owner",                  
		IsActive:     true,
	}

	insertQuery := `
		INSERT INTO users (id, tenant_id, email, password_hash, role, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING created_at`

	err = u.DBConnection.Pool.QueryRow(ctx, insertQuery,
		newUser.ID,
		newUser.TenantID,
		newUser.Email,
		newUser.PasswordHash,
		newUser.Role,
		newUser.IsActive,
	).Scan(&newUser.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to auto-register google user: %w", err)
	}

	return newUser, nil
}