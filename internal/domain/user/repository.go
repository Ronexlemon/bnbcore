package user

import (
	"context"
	"time"

	
	"github.com/satori/go.uuid"
)


type UserRepository interface{
	Register(ctx context.Context,tenantID uuid.UUID,email string ,password string)error
	Login(ctx context.Context,email,password string)(*User,error)
	GetRefreshToken(ctx context.Context, refreshToken string)(*REFRESHTOKEN,error)
	StoreRefreshToken(ctx context.Context,userID uuid.UUID,refreshToken string,createdAt time.Time,isRevoked bool,expiresAt time.Time)error
	GetUserByID(ctx context.Context,userID uuid.UUID)(*User,error)
	GetUserByEmail(ctx context.Context,email string)(*User,error)
	UpdatePasswordHash(ctx context.Context, userID uuid.UUID, newHash string) error
	LoginWithGoogle(ctx context.Context, googleClientID string, req GoogleLoginRequest) (*User, error)
}