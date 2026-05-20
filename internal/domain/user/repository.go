package user

import "context"


type UserRepository interface{
	Register(ctx context.Context,tenantID string,email string ,password string)error
}