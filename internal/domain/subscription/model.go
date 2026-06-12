package subscription

import (
	"time"

	"github.com/google/uuid"
)


type Status string
type PlanType string
type BillingCycle string

const (
	StatusActive   Status = "active"
	StatusTrial    Status = "trial"
	StatusExpired  Status = "expired"
	StatusCanceled Status = "canceled"

	PlanBasic      PlanType = "basic"
	PlanPro        PlanType = "pro"
	PlanEnterprise PlanType = "enterprise"

	BillingMonthly BillingCycle = "monthly"
	BillingYearly  BillingCycle = "yearly"
)

var PlanPricing = map[PlanType]map[BillingCycle]float64{
	PlanBasic: {
		BillingMonthly: 1500,
		BillingYearly:  15000, // ~2 months free
	},
	PlanPro: {
		BillingMonthly: 3500,
		BillingYearly:  35000,
	},
	PlanEnterprise: {
		BillingMonthly: 8000,
		BillingYearly:  80000,
	},
}

type Subscription struct {
	ID                 uuid.UUID
	UserID           uuid.UUID
	Plan               PlanType
	BillingCycle       BillingCycle
	Status             Status
	Amount             float64
	Currency           string
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	CreatedAt          time.Time
}

type CreateSubscriptionRequest struct {
	Plan         PlanType     `json:"plan"`
	BillingCycle BillingCycle `json:"billing_cycle"` 
}

type UpdateSubscriptionRequest struct {
	Plan         *PlanType     `json:"plan"`
	BillingCycle *BillingCycle `json:"billing_cycle"`
}
