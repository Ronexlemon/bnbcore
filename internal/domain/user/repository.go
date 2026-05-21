package user

import (
	"context"
	"time"
)


type UserRepository interface{
	Register(ctx context.Context,tenantID string,email string ,password string)error
	Login(ctx context.Context,email,password string)(*User,error)
	GetRefreshToken(ctx context.Context, refreshToken string)(*REFRESHTOKEN,error)
	StoreRefreshToken(ctx context.Context,userID string,refreshTokenHash string,createdAt time.Time,isRevoked bool,expiresAt time.Time)error
	GetUserByID(ctx context.Context,userID string)(*User,error)
}