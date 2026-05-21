package user

import (
	"context"
	"time"
)

type UserService struct{
	Repo UserRepository
}


func NewUserservice(repo UserRepository)(*UserService){
	return &UserService{
		Repo: repo,
	}
}


func (u *UserService)Register(ctx context.Context,tenantID string,email string,password string)error{

	return u.Repo.Register(ctx,tenantID,email,password)

}

func (u *UserService) Login(ctx context.Context,email,password string)(*User,error){
	//todo hash password
	return u.Repo.Login(ctx,email,password)

}
func (u *UserService)GetRefreshToken(ctx context.Context, refreshToken string)(*REFRESHTOKEN,error){
	//todo hash refresh token
	return u.Repo.GetRefreshToken(ctx,refreshToken)
}

func (u *UserService)StoreRefreshToken(ctx context.Context,userID string,refreshTokenHash string,createdAt time.Time,isRevoked bool,expiresAt time.Time)error{
	//todo hash refresh token
	return u.Repo.StoreRefreshToken(ctx,userID,refreshTokenHash,createdAt,isRevoked,expiresAt)
}

func (u *UserService)GetUserByID(ctx context.Context,userID string)(*User,error){
	return u.Repo.GetUserByID(ctx,userID)

}