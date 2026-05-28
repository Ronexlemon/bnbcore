package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ronexlemon/bnbcore/internal/domain/subscription"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/db"
)

type SubscriptionRepository struct {
	DbConnection *db.PostgresConn
}

func NewSubscriptionRepository(dbconn *db.PostgresConn) (*SubscriptionRepository, error) {
	if dbconn == nil {
		return nil, fmt.Errorf("db connection required")
	}
	return &SubscriptionRepository{DbConnection: dbconn}, nil
}

func (s *SubscriptionRepository) Create(ctx context.Context, sub *subscription.Subscription) (*subscription.Subscription, error) {
	err := s.DbConnection.Pool.QueryRow(ctx, `
		INSERT INTO subscriptions (
			id, tenant_id, plan, billing_cycle, status, amount, currency,
			current_period_start, current_period_end, created_at
		)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, NOW())
		RETURNING id, tenant_id, plan, billing_cycle, status, amount, currency,
		          current_period_start, current_period_end, created_at
	`,
		sub.TenantID,
		sub.Plan,
		sub.BillingCycle,
		sub.Status,
		sub.Amount,
		sub.Currency,
		sub.CurrentPeriodStart,
		sub.CurrentPeriodEnd,
	).Scan(
		&sub.ID,
		&sub.TenantID,
		&sub.Plan,
		&sub.BillingCycle,
		&sub.Status,
		&sub.Amount,
		&sub.Currency,
		&sub.CurrentPeriodStart,
		&sub.CurrentPeriodEnd,
		&sub.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}
	return sub, nil
}

func (s *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*subscription.Subscription, error) {
	var sub subscription.Subscription

	err := s.DbConnection.Pool.QueryRow(ctx, `
		SELECT id, tenant_id, plan, billing_cycle, status, amount, currency,
		       current_period_start, current_period_end, created_at
		FROM subscriptions
		WHERE id = $1
	`, id).Scan(
		&sub.ID,
		&sub.TenantID,
		&sub.Plan,
		&sub.BillingCycle,
		&sub.Status,
		&sub.Amount,
		&sub.Currency,
		&sub.CurrentPeriodStart,
		&sub.CurrentPeriodEnd,
		&sub.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("subscription not found")
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return &sub, nil
}

func (s *SubscriptionRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*subscription.Subscription, error) {
	var sub subscription.Subscription

	err := s.DbConnection.Pool.QueryRow(ctx, `
		SELECT id, tenant_id, plan, billing_cycle, status, amount, currency,
		       current_period_start, current_period_end, created_at
		FROM subscriptions
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, tenantID).Scan(
		&sub.ID,
		&sub.TenantID,
		&sub.Plan,
		&sub.BillingCycle,
		&sub.Status,
		&sub.Amount,
		&sub.Currency,
		&sub.CurrentPeriodStart,
		&sub.CurrentPeriodEnd,
		&sub.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("no subscription found")
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return &sub, nil
}

func (s *SubscriptionRepository) Update(ctx context.Context, id uuid.UUID, req subscription.UpdateSubscriptionRequest) (*subscription.Subscription, error) {
	var sub subscription.Subscription

	err := s.DbConnection.Pool.QueryRow(ctx, `
		UPDATE subscriptions SET
			plan          = COALESCE($1, plan),
			billing_cycle = COALESCE($2, billing_cycle)
		WHERE id = $3
		RETURNING id, tenant_id, plan, billing_cycle, status, amount, currency,
		          current_period_start, current_period_end, created_at
	`, req.Plan, req.BillingCycle, id).Scan(
		&sub.ID,
		&sub.TenantID,
		&sub.Plan,
		&sub.BillingCycle,
		&sub.Status,
		&sub.Amount,
		&sub.Currency,
		&sub.CurrentPeriodStart,
		&sub.CurrentPeriodEnd,
		&sub.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("subscription not found")
		}
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}
	return &sub, nil
}

func (s *SubscriptionRepository) Cancel(ctx context.Context, id, tenantID uuid.UUID) error {
	tag, err := s.DbConnection.Pool.Exec(ctx, `
		UPDATE subscriptions SET status = 'canceled'
		WHERE id = $1 AND tenant_id = $2
		  AND status NOT IN ('canceled', 'expired')
	`, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("subscription not found or already canceled")
	}
	return nil
}

func (s *SubscriptionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status subscription.Status) error {
	tag, err := s.DbConnection.Pool.Exec(ctx, `
		UPDATE subscriptions SET status = $1 WHERE id = $2
	`, status, id)
	if err != nil {
		return fmt.Errorf("failed to update subscription status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("subscription not found")
	}
	return nil
}

func (s *SubscriptionRepository) GetExpired(ctx context.Context) ([]*subscription.Subscription, error) {
	rows, err := s.DbConnection.Pool.Query(ctx, `
		SELECT id, tenant_id, plan, billing_cycle, status, amount, currency,
		       current_period_start, current_period_end, created_at
		FROM subscriptions
		WHERE status IN ('active', 'trial')
		  AND current_period_end < NOW()
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expired subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []*subscription.Subscription
	for rows.Next() {
		var sub subscription.Subscription
		if err := rows.Scan(
			&sub.ID,
			&sub.TenantID,
			&sub.Plan,
			&sub.BillingCycle,
			&sub.Status,
			&sub.Amount,
			&sub.Currency,
			&sub.CurrentPeriodStart,
			&sub.CurrentPeriodEnd,
			&sub.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subs = append(subs, &sub)
	}
	return subs, nil
}

func (s *SubscriptionRepository) IsActive(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	var active bool
	err := s.DbConnection.Pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM subscriptions
			WHERE tenant_id = $1
			  AND status IN ('active', 'trial')
			  AND current_period_end > NOW()
		)
	`, tenantID).Scan(&active)
	if err != nil {
		return false, fmt.Errorf("failed to check subscription: %w", err)
	}
	return active, nil
}
var _ subscription.Repository = (*SubscriptionRepository)(nil)