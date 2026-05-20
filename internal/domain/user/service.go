package user

import (
	"context"
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