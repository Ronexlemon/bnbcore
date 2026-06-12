package subscription

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, sub *Subscription) (*Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Subscription, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Subscription, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateSubscriptionRequest) (*Subscription, error)
	Cancel(ctx context.Context, id, userID uuid.UUID) error
	GetExpired(ctx context.Context) ([]*Subscription, error)
	IsActive(ctx context.Context, userID uuid.UUID) (bool, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error
}