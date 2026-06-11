package tenant

import (
	"time"

	"github.com/google/uuid"

)

// type OnboardResult struct {
// 	Tenant *Tenant
// 	User   *user.User
// }

type TenantStatus string
 
const (
	StatusTrial    TenantStatus = "trial"
	StatusActive   TenantStatus = "active"
	StatusSuspended TenantStatus = "suspended"
)
const TrialDays = 14
type Tenant struct {
	ID          *uuid.UUID     `json:"id"`
	UserID      uuid.UUID   `json:"user_id"`
	ShopName *string `json:"name"`
	ShopDescription        *string       `json:"shop_description"`
	Subdomain   *string       `json:"subdomain"`
	Banner *string `json:"banner"`
	LongDescription *string `json:"long_description"`
	Owner     *Owner   `json:"user"`
	 PhoneNumber    *string         `json:"phone_number"`
	Status      *TenantStatus `json:"status"`
	TrialEndsAt time.Time    `json:"trial_ends_at"`
	CreatedAt   time.Time    `json:"created_at"`
}

type Owner struct {
	ID           uuid.UUID `json:"id"`
	Email        string  `json:"email"`
	PasswordHash string  `json:"password_hash"`
	Password  string `json:"password"`
	Role         string   `json:"role"`
	Tenant    *Tenant   `json:"tenant"`
	IsActive      bool `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	
}

// type OnboardRequest struct {
// 	TenantName string `json:"tenant_name"` 
// 	Subdomain  string `json:"subdomain"`   
// }
// type OnboardResponse struct {
// 	Tenant      *Tenant      `json:"tenant"`
// 	User        *user.User`json:"user"`
// 	AccessToken  string      `json:"access_token"`
// 	RefreshToken string      `json:"refresh_token"`
// 	TokenType    string      `json:"token_type"`
// 	ExpiresIn    int64       `json:"expires_in"`
// }
type UpdateTenantRequest struct {
	ShopDescription  *string       `json:"shop_description"`
	Subdomain *string       `json:"subdomain"`
	Status    *TenantStatus `json:"status"`
	Banner *string `json:"banner"`
	LongDescription *string `json:"long_description"`
	PhoneNumber    *string         `json:"phone_number"`
	ShopName *string `json:"name"`
}