package user

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/google/uuid"
	"github.com/ronexlemon/bnbcore/internal/auth/password"
	"github.com/ronexlemon/bnbcore/internal/auth/token"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/queue"
	"github.com/ronexlemon/bnbcore/internal/worker"
	"github.com/ronexlemon/bnbcore/pkg/helpers"
)

type UserService struct{
	Repo UserRepository
	PasswordEngine *password.PasswordHasher
	TokenEngine    *token.TokenHasher
	GoogleClientID string
	Enqueuer  *queue.Enqueuer
}


func NewUserservice(repo UserRepository,passEngine *password.PasswordHasher,tokenEngine *token.TokenHasher,googleClientID string,enqueuer *queue.Enqueuer )(*UserService){
	return &UserService{
		Repo: repo,
		PasswordEngine: passEngine,
		TokenEngine: tokenEngine,
		GoogleClientID: googleClientID,
		Enqueuer: enqueuer,
	}
}


func (u *UserService) Register(ctx context.Context, email, password string) (*User, error) {

    existingUser, err := u.Repo.GetUserByEmail(ctx, email)
    if err == nil && existingUser != nil {
        if existingUser.IsActive {
            return nil, errors.New("email is already registered")
        }
        
        return existingUser, nil
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
func (u *UserService) GetUserByEmail(ctx context.Context,email string)(*User,error){
	return u.Repo.GetUserByEmail(ctx,email)

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
	userResult.PasswordHash = ""

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

func (s *UserService) CreateMagicLinkToken(ctx context.Context, userID uuid.UUID) (string, error) {
    rawToken, tokenHash, err := helpers.GenerateMagicLink()
    if err != nil {
        return "", err
    }

    expiresAt := time.Now().UTC().Add(30 * time.Minute)

    if err := s.Repo.StoreMagicLinkToken(ctx, userID, tokenHash, expiresAt); err != nil {
        return "", err
    }

    return rawToken, nil 
}

func (s *UserService) ValidateMagicLinkToken(ctx context.Context, rawToken string) (*User, error) {
    record, err := s.Repo.FindMagicLinkToken(ctx, rawToken)
	fmt.Println("The record",record)
    if err != nil {
		fmt.Println("The Error",err)
        return nil, err
    }

    if err := s.Repo.DeleteMagicLinkToken(ctx, rawToken); err != nil {
		fmt.Println("The Delete error",err)
        return nil, err
    }

    // Activate the account on first verified click
    if err := s.Repo.ActivateUser(ctx, record.UserID); err != nil {
		fmt.Println("Activate error",err)
        return nil, err
    }

    return s.Repo.GetUserByID(ctx, record.UserID)
}


func (s *UserService) RequestPasswordReset(ctx context.Context, email, baseURL string) error {
	usr, err := s.Repo.GetUserByEmail(ctx, email)
	if err != nil {
		log.Printf("[user] password reset requested for unknown email: %s", email)
		return nil
	}

	rawToken,err := s.CreateMagicLinkToken(ctx,usr.ID)
	if err !=nil{
		return  fmt.Errorf("failed to store reset token: %w", err)
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", baseURL, rawToken)

	if s.Enqueuer != nil {
		payload := worker.PasswordResetPayload{
			UserID: usr.ID.String(),
			Email:  usr.Email,
			Link:   resetLink,
		}
		if _, err := queue.EnqueueTask(ctx, s.Enqueuer, worker.PasswordResetEmailTask, payload); err != nil {
			log.Printf("[user] failed to enqueue password reset email for %s: %v", usr.Email, err)
		}
	}

	return nil
}


func (s *UserService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	record, err := s.Repo.FindMagicLinkToken(ctx, rawToken)
    if err != nil {
		return fmt.Errorf("invalid or expired token")
    }
	hashedPassword, err := s.PasswordEngine.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.Repo.UpdatePasswordHash(ctx, record.UserID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if err := s.Repo.DeleteMagicLinkToken(ctx, rawToken); err != nil {
        return nil
    }
	return nil
}