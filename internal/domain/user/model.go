package user

import (
	"time"
"github.com/google/uuid"
)

type GoogleLoginRequest struct {
	Credential  string `json:"credential"` 
}

type User struct {
	ID           uuid.UUID `json:"id"`
	TenantID      *uuid.UUID  `json:"tenant_id"`
	Email        string  `json:"email"`
	PasswordHash string  `json:"password_hash"`
	Password  string `json:"password"`
	Role         string   `json:"role"`
	IsActive      bool `json:"is_active"`
	Subdomain    string  `json:"subdomain"`
    ShopName     string `json:"shop_name"`
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