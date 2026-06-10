package user

import (
	"time"

	"github.com/google/uuid"
	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
)

type GoogleLoginRequest struct {
	Credential  string `json:"credential"` 
}

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string  `json:"email"`
	PasswordHash string  `json:"password_hash"`
	Password  string `json:"password"`
	Role         string   `json:"role"`
	IsActive      bool `json:"is_active"`
	Tenant    *tenant.Tenant   `json:"tenant"`
	CreatedAt    time.Time `json:"created_at"`
	
}


type REFRESHTOKEN struct{
		ID			string `json:"id"` 
		UserID		*uuid.UUID	`json:"user_id"` 
		TokenHash	string  `json:"token_hash"` 
		ExpiresAt	time.Time	 `json:"expires_at"` 
		IsRevoked	bool	`json:"is_revoked"` 
		CreateAt	time.Time	`json:"created_at"` 
}
type MagicLinkToken struct {
    ID        uuid.UUID
    UserID    uuid.UUID
    TokenHash string
    ExpiresAt time.Time
    Used      bool
    CreatedAt time.Time
}