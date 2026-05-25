package tenant

import (
	"time"

	"github.com/ronexlemon/bnbcore/internal/domain/user"
	uuid "github.com/satori/go.uuid"
)

type OnboardResult struct {
	Tenant *Tenant
	User   *user.User
}

type TenantStatus string
 
const (
	StatusTrial    TenantStatus = "trial"
	StatusActive   TenantStatus = "active"
	StatusSuspended TenantStatus = "suspended"
)
const TrialDays = 14
type Tenant struct {
	ID          uuid.UUID     `json:"id"`
	Name        string       `json:"name"`
	Subdomain   string       `json:"subdomain"`
	Status      TenantStatus `json:"status"`
	TrialEndsAt time.Time    `json:"trial_ends_at"`
	CreatedAt   time.Time    `json:"created_at"`
}

type OnboardRequest struct {
	TenantName string `json:"tenant_name"` 
	Subdomain  string `json:"subdomain"`   
}
type OnboardResponse struct {
	Tenant      *Tenant      `json:"tenant"`
	User        *user.User`json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	TokenType    string      `json:"token_type"`
	ExpiresIn    int64       `json:"expires_in"`
}