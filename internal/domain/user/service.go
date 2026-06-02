package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/ronexlemon/bnbcore/internal/auth/password"
	"github.com/ronexlemon/bnbcore/internal/auth/token"
	"github.com/google/uuid"
)

type UserService struct{
	Repo UserRepository
	PasswordEngine *password.PasswordHasher
	TokenEngine    *token.TokenHasher
	GoogleClientID string
}


func NewUserservice(repo UserRepository,passEngine *password.PasswordHasher,tokenEngine *token.TokenHasher,googleClientID string,)(*UserService){
	return &UserService{
		Repo: repo,
		PasswordEngine: passEngine,
		TokenEngine: tokenEngine,
		GoogleClientID: googleClientID,
	}
}


func (u *UserService) Register(ctx context.Context, email, password string) (*User, error) {

    
    exists, err := u.Repo.EmailExists(ctx, email)
    if err != nil {
        return nil, fmt.Errorf("registration failed: %w", err)
    }
    if exists {
        return nil, errors.New("Invalid Credentials")
    }

    hashedPassword, err := u.PasswordEngine.Hash(password)
    if err != nil {
        return nil, err
    }

    return u.Repo.Register(ctx, email, hashedPassword)
}


func (u *UserService) RegisterWithGoogle(ctx context.Context, req GoogleLoginRequest) (*User, error) {

    payload, err := idtoken.Validate(ctx, req.Credential, u.GoogleClientID)
    if err != nil {
        return nil, fmt.Errorf("invalid google token: %w", err)
    }

    email, ok := payload.Claims["email"].(string)
    if !ok {
        return nil, errors.New("google token missing email claim")
    }

    existing, err := u.Repo.GetUserByEmail(ctx, email)
    if err == nil && existing != nil {
        existing.PasswordHash = ""
        return existing, nil
    }

    return u.Repo.GoogleRegister(ctx, email)
}
func (u *UserService) Login(ctx context.Context,email,password string)(*User,error){
	userResult, err := u.Repo.GetUserByEmail(ctx, email)
	if err != nil {
		fmt.Println("Error1",err)
		return nil, errors.New("invalid credentials")
	}
fmt.Println("Pass and email",password,email)

	fmt.Println("stored",userResult.PasswordHash)
	fmt.Println(" results",userResult)
	isValid, needsUpgrade := u.PasswordEngine.Compare(password, userResult.PasswordHash)
	if !isValid {
		fmt.Println("Error2",err)
		return nil, errors.New("invalid credentials")
	}

	//  If the key was rotated, silently update the DB to the newest key version
	if needsUpgrade {
		newHash, err := u.PasswordEngine.Hash(password)
		if err == nil {
			// todo: handle error silently so login doesn't fail if DB write flakes
			_ = u.Repo.UpdatePasswordHash(ctx, userResult.ID, newHash)
		}
	}

	return userResult, nil

}
func (u *UserService)GetRefreshToken(ctx context.Context, refreshToken string)(*REFRESHTOKEN,error){
	lookupHash := u.TokenEngine.Hash(refreshToken)
	
	return u.Repo.GetRefreshToken(ctx, lookupHash)
}

func (u *UserService)StoreRefreshToken(ctx context.Context,userID uuid.UUID,refreshToken string,createdAt time.Time,isRevoked bool,expiresAt time.Time)error{
	lookupHash := u.TokenEngine.Hash(refreshToken)
	
	return u.Repo.StoreRefreshToken(ctx, userID, lookupHash, createdAt, isRevoked, expiresAt)
}

func (u *UserService)GetUserByID(ctx context.Context,userID uuid.UUID)(*User,error){
	return u.Repo.GetUserByID(ctx,userID)

}

func (u *UserService) LoginWithGoogle(ctx context.Context, req GoogleLoginRequest) (*User, error) {
	if u.GoogleClientID == "" {
		return nil, errors.New("google login is not configured")
	}
	return u.Repo.LoginWithGoogle(ctx, u.GoogleClientID, req)
}
