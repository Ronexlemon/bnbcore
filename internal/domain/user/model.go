package user

import "time"


type User struct {
	ID           string `json:"id"`
	TenantID     string  `json:"tenant_id"`
	Email        string  `json:"email"`
	PasswordHash string  `json:"password"`
	Role         string   `json:"role"`
}


type REFRESHTOKEN struct{
		ID			string `json:"id"` 
		UserID		string	`json:"user_id"` 
		TokenHash	string  `json:"token_hash"` 
		ExpiresAt	time.Time	 `json:"expires_at"` 
		IsRevoked	bool	`json:"is_revoked"` 
		CreateAt	time.Time	`json:"created_at"` 
}