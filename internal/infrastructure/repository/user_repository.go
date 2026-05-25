package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
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


func (u *UserRepository) Register(ctx context.Context, email, hashedPassword, shopName, subdomain string) (*user.User, error) {
    tx, err := u.DBConnection.Pool.Begin(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)

    var ten struct {
        ID        uuid.UUID
        ShopName  string
        Subdomain string
    }

    // trial_ends_at = 14 days from now
    trialEndsAt := time.Now().Add(14 * 24 * time.Hour)

    err = tx.QueryRow(ctx, `
        INSERT INTO tenants (id, name, subdomain, status, trial_ends_at, created_at)
        VALUES (gen_random_uuid(), $1, $2, 'trial', $3, NOW())
        RETURNING id, name, subdomain
    `, shopName, subdomain, trialEndsAt).Scan(
        &ten.ID,
        &ten.ShopName,
        &ten.Subdomain,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create tenant: %w", err)
    }

    var usr user.User
    err = tx.QueryRow(ctx, `
        INSERT INTO users (id, tenant_id, email, password_hash, role, is_active, created_at)
        VALUES (gen_random_uuid(), $1, $2, $3, 'owner', true, NOW())
        RETURNING id, tenant_id, email, role, is_active, created_at
    `, ten.ID, email, hashedPassword).Scan(
        &usr.ID,
        &usr.TenantID,
        &usr.Email,
        &usr.Role,
        &usr.IsActive,
        &usr.CreatedAt,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    if err := tx.Commit(ctx); err != nil {
        return nil, fmt.Errorf("failed to commit: %w", err)
    }

    usr.Subdomain = ten.Subdomain
    usr.ShopName  = ten.ShopName

    return &usr, nil
}

func (u *UserRepository) Login(ctx context.Context,email,password string)(*user.User,error){
	var user user.User
	query := `
        SELECT 
            u.id,
            u.tenant_id,
            u.email,
            u.password_hash,
            u.created_at,
            u.is_active,
            u.role,
            t.subdomain,
            t.name
        FROM users u
        INNER JOIN tenants t ON t.id = u.tenant_id
        WHERE u.email = $1
    `
	err:=u.DBConnection.Pool.QueryRow(ctx,query,email).Scan(
		&user.ID,
        &user.TenantID,
        &user.Email,
        &user.PasswordHash,
        &user.CreatedAt,
        &user.IsActive,
        &user.Role,
        &user.Subdomain,  
        &user.ShopName,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}
	 user.PasswordHash = ""

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



func (u *UserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*user.User, error) {
    var usr user.User

    query := `
        SELECT
            u.id,
            u.email,
            u.role,
            u.tenant_id,
            u.created_at,
            u.is_active,
            t.subdomain,
            t.name
        FROM users u
        INNER JOIN tenants t ON t.id = u.tenant_id
        WHERE u.id = $1
    `

    err := u.DBConnection.Pool.QueryRow(ctx, query, userID).Scan(
        &usr.ID,
        &usr.Email,
        &usr.Role,
        &usr.TenantID,
        &usr.CreatedAt,
        &usr.IsActive,
        &usr.Subdomain,
        &usr.ShopName,
    )
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) { 
            return nil, errors.New("user not found")
        }
        return nil, err
    }

    return &usr, nil
}

func (u *UserRepository) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
    var usr user.User

    query := `
        SELECT
            u.id,
            u.email,
            u.password_hash,
            u.role,
            u.tenant_id,
            u.created_at,
            u.is_active,
            t.subdomain,
            t.name
        FROM users u
        INNER JOIN tenants t ON t.id = u.tenant_id
        WHERE u.email = $1
    `

    err := u.DBConnection.Pool.QueryRow(ctx, query, email).Scan(
        &usr.ID,
        &usr.Email,
        &usr.PasswordHash,
        &usr.Role,
        &usr.TenantID,
        &usr.CreatedAt,
        &usr.IsActive,
        &usr.Subdomain,
        &usr.ShopName,
    )
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, errors.New("invalid credentials")
        }
        return nil, fmt.Errorf("failed to get user by email: %w", err)
    }

    return &usr, nil
}

func (u *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
    var exists bool
    query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
    err := u.DBConnection.Pool.QueryRow(ctx, query, email).Scan(&exists)
    if err != nil {
        return false, fmt.Errorf("failed to check email: %w", err)
    }
    return exists, nil
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


func (u *UserRepository) SubdomainExists(ctx context.Context, subdomain string) (bool, error) {
    var exists bool
    query := `SELECT EXISTS(SELECT 1 FROM tenants WHERE subdomain = $1)`
    err := u.DBConnection.Pool.QueryRow(ctx, query, subdomain).Scan(&exists)
    if err != nil {
        return false, fmt.Errorf("failed to check subdomain: %w", err)
    }
    return exists, nil
}

func (u *UserRepository) GoogleRegister(ctx context.Context, email, shopName, subdomain string) (*user.User, error) {
    tx, err := u.DBConnection.Pool.Begin(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)

    var ten struct {
        ID        uuid.UUID
        ShopName  string
        Subdomain string
    }

    trialEndsAt := time.Now().Add(14 * 24 * time.Hour)

    err = tx.QueryRow(ctx, `
        INSERT INTO tenants (id, name, subdomain, status, trial_ends_at, created_at)
        VALUES (gen_random_uuid(), $1, $2, 'trial', $3, NOW())
        RETURNING id, name, subdomain
    `, shopName, subdomain, trialEndsAt).Scan(
        &ten.ID,
        &ten.ShopName,
        &ten.Subdomain,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create tenant: %w", err)
    }

    var usr user.User
    err = tx.QueryRow(ctx, `
        INSERT INTO users (id, tenant_id, email, password_hash, role, is_active, created_at)
        VALUES (gen_random_uuid(), $1, $2, 'OAUTH_EXTERNAL_ACCOUNT', 'owner', true, NOW())
        RETURNING id, tenant_id, email, role, is_active, created_at
    `, ten.ID, email).Scan(
        &usr.ID,
        &usr.TenantID,
        &usr.Email,
        &usr.Role,
        &usr.IsActive,
        &usr.CreatedAt,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    if err := tx.Commit(ctx); err != nil {
        return nil, fmt.Errorf("failed to commit: %w", err)
    }

    usr.Subdomain = ten.Subdomain
    usr.ShopName  = ten.ShopName

    return &usr, nil
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

    // Pull tenant from context — set by SubdomainResolver upstream
    t := tenant.FromContext(ctx)
    if t == nil {
        return nil, errors.New("tenant not found in context")
    }

    var existing user.User
    selectQuery := `
        SELECT
            u.id,
            u.tenant_id,
            u.email,
            u.password_hash,
            u.role,
            u.is_active,
            u.created_at,
            t.subdomain,
            t.name
        FROM users u
        INNER JOIN tenants t ON t.id = u.tenant_id
        WHERE u.email = $1
          AND u.tenant_id = $2
    `
    err = u.DBConnection.Pool.QueryRow(ctx, selectQuery, email, t.ID).Scan(
        &existing.ID,
        &existing.TenantID,
        &existing.Email,
        &existing.PasswordHash,
        &existing.Role,
        &existing.IsActive,
        &existing.CreatedAt,
        &existing.Subdomain,
        &existing.ShopName,
    )
    if err == nil {
        return &existing, nil
    }
    if !errors.Is(err, pgx.ErrNoRows) {
        return nil, fmt.Errorf("failed to query user: %w", err)
    }

    newUser := &user.User{
        ID:           uuid.NewV4(),
        TenantID:     &t.ID,
        Email:        email,
        PasswordHash: "OAUTH_EXTERNAL_ACCOUNT",
        Role:         "owner",
        IsActive:     true,
        Subdomain:    t.Subdomain,
        ShopName:     t.Name,
    }

    insertQuery := `
        INSERT INTO users (id, tenant_id, email, password_hash, role, is_active, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW())
        RETURNING created_at
    `
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