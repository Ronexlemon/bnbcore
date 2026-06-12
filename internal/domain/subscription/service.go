package subscription

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) Subscribe(ctx context.Context, userID uuid.UUID, req CreateSubscriptionRequest) (*Subscription, error) {
	if req.Plan == "" {
		return nil, errors.New("plan is required")
	}
	if req.BillingCycle != BillingMonthly && req.BillingCycle != BillingYearly {
		return nil, errors.New("billing_cycle must be 'monthly' or 'yearly'")
	}

	existing, err := s.Repo.GetByUserID(ctx, userID)
	if err == nil && existing != nil && existing.Status == StatusActive {
		return nil, errors.New("user already has an active subscription, upgrade or cancel first")
	}

	price, ok := PlanPricing[req.Plan][req.BillingCycle]
	if !ok {
		return nil, fmt.Errorf("invalid plan or billing cycle")
	}

	now := time.Now()
	var periodEnd time.Time
	switch req.BillingCycle {
	case BillingMonthly:
		periodEnd = now.AddDate(0, 1, 0)
	case BillingYearly:
		periodEnd = now.AddDate(1, 0, 0)
	}

	sub := &Subscription{
		UserID:             userID,
		Plan:               req.Plan,
		BillingCycle:       req.BillingCycle,
		Status:             StatusActive,
		Amount:             price,
		Currency:           "KES",
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   periodEnd,
	}

	return s.Repo.Create(ctx, sub)
}

func (s *Service) GetMySubscription(ctx context.Context, userID uuid.UUID) (*Subscription, error) {
	return s.Repo.GetByUserID(ctx, userID)
}

func (s *Service) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*Subscription, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *Service) Upgrade(ctx context.Context, userID uuid.UUID, req UpdateSubscriptionRequest) (*Subscription, error) {
	existing, err := s.Repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("no active subscription found: %w", err)
	}

	plan := existing.Plan
	cycle := existing.BillingCycle
	if req.Plan != nil {
		plan = *req.Plan
	}
	if req.BillingCycle != nil {
		cycle = *req.BillingCycle
	}

	price, ok := PlanPricing[plan][cycle]
	if !ok {
		return nil, errors.New("invalid plan or billing cycle")
	}

	updated, err := s.Repo.Update(ctx, existing.ID, req)
	if err != nil {
		return nil, err
	}
	updated.Amount = price
	return updated, nil
}

func (s *Service) Cancel(ctx context.Context, userID uuid.UUID) error {
	existing, err := s.Repo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("no subscription found: %w", err)
	}
	return s.Repo.Cancel(ctx, existing.ID, userID)
}

// GetPlans returns available plans and pricing for the frontend
func (s *Service) GetPlans() map[PlanType]map[BillingCycle]float64 {
	return PlanPricing
}