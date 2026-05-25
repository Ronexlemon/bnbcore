package user

import (
	"time"

	pgxsatori "github.com/jackc/pgx/pgtype/ext/satori-uuid"
	//"github.com/ronexlemon/bnbcore/internal/auth"
	"github.com/satori/go.uuid"
)

type GoogleLoginRequest struct {
	Credential  string `json:"credential"` 
}

type User struct {
	ID           uuid.UUID `json:"id"`
	TenantID      *pgxsatori.UUID  `json:"tenant_id"`
	Email        string  `json:"email"`
	PasswordHash string  `json:"password"`
	Role         string   `json:"role"`
	IsActive      bool `json:"is_active"`
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