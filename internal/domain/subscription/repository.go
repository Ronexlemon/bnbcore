package subscription

import (
	"context"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, sub *Subscription) (*Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Subscription, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*Subscription, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateSubscriptionRequest) (*Subscription, error)
	Cancel(ctx context.Context, id, tenantID uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error
}