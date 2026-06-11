package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/google/uuid"
	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
	"github.com/ronexlemon/bnbcore/internal/domain/user"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/db"
	"github.com/ronexlemon/bnbcore/pkg/helpers"
	"google.golang.org/api/idtoken"
)



type pgxCommon interface {
    Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
    Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
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


func (u *UserRepository) Register(ctx context.Context, email, hashedPassword string) (*user.User, error) {
	var usr user.User

	err := u.DBConnection.Pool.QueryRow(ctx, `
		INSERT INTO users (id, email, password_hash, role, is_active, created_at)
		VALUES (gen_random_uuid(), $1, $2, 'owner', false, NOW())
		RETURNING id, email, role, is_active, created_at
	`, email, hashedPassword).Scan(
		&usr.ID,
		&usr.Email,
		&usr.Role,
		&usr.IsActive,
		&usr.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	return &usr, nil
}

func (u *UserRepository) Login(ctx context.Context, email string,password string) (*user.User, error) {
	var usr user.User

	err := u.DBConnection.Pool.QueryRow(ctx, `
		SELECT id, email, password_hash, role, is_active, created_at
		FROM users
		WHERE email = $1
	`, email).Scan(
		&usr.ID,
		&usr.Email,
		&usr.PasswordHash,
		&usr.Role,
		&usr.IsActive,
		&usr.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("invalid credentials")
		}
		return nil, fmt.Errorf("failed to login: %w", err)
	}

	return &usr, nil
}

func (u *UserRepository) ActivateUser(ctx context.Context, userID uuid.UUID) error {
    result, err := u.DBConnection.Pool.Exec(ctx,
        `UPDATE users SET is_active = true, updated_at = NOW() WHERE id = $1`,
        userID,
    )
    if err != nil {
        return fmt.Errorf("failed to activate user: %w", err)
    }
    if result.RowsAffected() == 0 {
        return fmt.Errorf("user not found: %s", userID)
    }
    return nil
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
	var tn tenant.Tenant

	var tenantID *uuid.UUID
	var tenantTrialEndsAt *time.Time
	var tenantCreatedAt *time.Time

	err := u.DBConnection.Pool.QueryRow(ctx, `
		SELECT 
			u.id, u.email, u.role, u.is_active, u.created_at,
			t.id, t.name, t.subdomain, t.shop_description, t.banner, t.long_description, t.phone_number, t.status, t.trial_ends_at, t.created_at
		FROM users u
		LEFT JOIN tenants t ON u.id = t.user_id
		WHERE u.id = $1
	`, userID).Scan(
		&usr.ID,
		&usr.Email,
		&usr.Role,
		&usr.IsActive,
		&usr.CreatedAt,
		&tenantID,
		&tn.ShopName,
		&tn.Subdomain,
		&tn.ShopDescription,
		&tn.Banner,
		&tn.LongDescription,
		&tn.PhoneNumber,
		&tn.Status,
		&tenantTrialEndsAt,
		&tenantCreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if tenantID != nil {
		tn.ID = tenantID
		tn.UserID = usr.ID
		if tenantTrialEndsAt != nil {
			tn.TrialEndsAt = *tenantTrialEndsAt
		}
		if tenantCreatedAt != nil {
			tn.CreatedAt = *tenantCreatedAt
		}
		usr.Tenant = &tn
	} else {
		usr.Tenant = nil
	}

	return &usr, nil
}

func (u *UserRepository) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	var usr user.User

	var tn tenant.Tenant

	var tenantID *uuid.UUID
	var tenantTrialEndsAt *time.Time
	var tenantCreatedAt *time.Time

	err := u.DBConnection.Pool.QueryRow(ctx, `
		SELECT 
			u.id, u.password_hash, u.email, u.role, u.is_active, u.created_at,
			t.id, t.name, t.subdomain, t.shop_description, t.banner, t.long_description, t.phone_number, t.status, t.trial_ends_at, t.created_at
		FROM users u
		LEFT JOIN tenants t ON u.id = t.user_id
		WHERE u.email = $1
	`, email).Scan(
		&usr.ID,
		&usr.PasswordHash,
		&usr.Email,
		&usr.Role,
		&usr.IsActive,
		&usr.CreatedAt,
		&tenantID,
		&tn.ShopName,
		&tn.Subdomain,
		&tn.ShopDescription,
		&tn.Banner,
		&tn.LongDescription,
		&tn.PhoneNumber,
		&tn.Status,
		&tenantTrialEndsAt,
		&tenantCreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if tenantID != nil {
		tn.ID = tenantID
		tn.UserID = usr.ID
		if tenantTrialEndsAt != nil {
			tn.TrialEndsAt = *tenantTrialEndsAt
		}
		if tenantCreatedAt != nil {
			tn.CreatedAt = *tenantCreatedAt
		}
		usr.Tenant = &tn
	} else {
		usr.Tenant = nil
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
		&refresh.TokenHash,
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

func (u *UserRepository) GoogleRegister(ctx context.Context, email string) (*user.User, error) {
	var usr user.User

	err := u.DBConnection.Pool.QueryRow(ctx, `
		INSERT INTO users (id, email, password_hash, role, is_active, created_at)
		VALUES (gen_random_uuid(), $1, 'OAUTH_EXTERNAL_ACCOUNT', 'owner', true, NOW())
		RETURNING id, email, role, is_active, created_at
	`, email).Scan(
		&usr.ID,
		&usr.Email,
		&usr.Role,
		&usr.IsActive,
		&usr.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register google user: %w", err)
	}

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

	var usr user.User
	
	err = u.DBConnection.Pool.QueryRow(ctx, `
		INSERT INTO users (id, email, password_hash, role, is_active, created_at)
		VALUES (gen_random_uuid(), $1, 'OAUTH_EXTERNAL_ACCOUNT', 'owner', true, NOW())
		RETURNING id, email, role, is_active, created_at
	`, email).Scan(
		&usr.ID,
		&usr.Email,
		&usr.Role,
		&usr.IsActive,
		&usr.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to auto-register google user: %w", err)
	}

	return &usr, nil
}

func (u *UserRepository) StoreMagicLinkToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
    tx, err := u.DBConnection.Pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx) 

    _, err = tx.Exec(ctx, `
        UPDATE magic_link_tokens 
        SET used = true 
        WHERE user_id = $1 AND used = false
    `, userID)
    if err != nil {
        return fmt.Errorf("failed to invalidate outstanding user tokens: %w", err)
    }

    _, err = tx.Exec(ctx, `
        INSERT INTO magic_link_tokens (id, user_id, token_hash, expires_at, used, created_at)
        VALUES (gen_random_uuid(), $1, $2, $3, false, NOW())
    `, userID, tokenHash, expiresAt)
    if err != nil {
        return fmt.Errorf("failed to store magic link token structure: %w", err)
    }

    return tx.Commit(ctx)
}


func (u *UserRepository) FindMagicLinkToken(ctx context.Context, rawToken string) (*user.MagicLinkToken, error) {
    hash := helpers.HashToken(rawToken)

    var t user.MagicLinkToken
    err := u.DBConnection.Pool.QueryRow(ctx, `
        SELECT id, user_id, token_hash, expires_at, used, created_at
        FROM magic_link_tokens
        WHERE token_hash = $1
          AND used       = false
          AND expires_at > NOW() 
    `, hash).Scan(
        &t.ID,
        &t.UserID,
        &t.TokenHash,
        &t.ExpiresAt,
        &t.Used,
        &t.CreatedAt,
    )
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, errors.New("invalid or expired link")
        }
        return nil, fmt.Errorf("failed to find magic link token: %w", err)
    }
    return &t, nil
}
func (u *UserRepository) DeleteMagicLinkToken(ctx context.Context, rawToken string) error {
    hash := helpers.HashToken(rawToken)

    result, err := u.DBConnection.Pool.Exec(ctx, `
        UPDATE magic_link_tokens 
        SET used = true 
        WHERE token_hash = $1 AND used = false -- Strict condition
    `, hash)
    if err != nil {
        return fmt.Errorf("failed to invalidate magic link token: %w", err)
    }

    // If 0 rows were changed, the token was already consumed concurrently
    if result.RowsAffected() == 0 {
        return errors.New("token was already used")
    }
    return nil
}

