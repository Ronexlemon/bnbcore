package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	
)


type UserRepository interface{
	Register(ctx context.Context, email, password string) (*User, error) 
	Login(ctx context.Context,email,password string)(*User,error)
	GetRefreshToken(ctx context.Context, refreshToken string)(*REFRESHTOKEN,error)
	RevokeRefreshToken(ctx context.Context, refreshTokenHash string) error
	StoreRefreshToken(ctx context.Context,userID uuid.UUID,refreshToken string,createdAt time.Time,isRevoked bool,expiresAt time.Time)error
	GetUserByID(ctx context.Context,userID uuid.UUID)(*User,error)
	GetUserByEmail(ctx context.Context,email string)(*User,error)
	GoogleRegister(ctx context.Context, email string)(*User, error)
	SubdomainExists(ctx context.Context, subdomain string) (bool, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	UpdatePasswordHash(ctx context.Context, userID uuid.UUID, newHash string) error
	LoginWithGoogle(ctx context.Context, googleClientID string, req GoogleLoginRequest) (*User, error)
	DeleteMagicLinkToken(ctx context.Context, rawToken string) error
	FindMagicLinkToken(ctx context.Context, rawToken string) (*MagicLinkToken, error)
	StoreMagicLinkToken(ctx context.Context, userID uuid.UUID, rawToken string, expiresAt time.Time) error
	ActivateUser(ctx context.Context, userID uuid.UUID) error
	
	// RegisterWithGoogle(ctx context.Context, googleClientID string, req GoogleLoginRequest, shopName, subdomain string) (*User, error)
}
