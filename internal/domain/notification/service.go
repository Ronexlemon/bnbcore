package notification

import (
	"context"
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

func (s *Service) Create(ctx context.Context, req CreateNotificationRequest) (*Notification, error) {
	n := &Notification{
		TenantID: req.TenantID,
		UserID:   req.UserID,
		Type:     req.Type,
		Channel:  req.Channel,
		Title:    req.Title,
		Message:  req.Message,
		Metadata: req.Metadata,
		Status:   StatusPending,
	}
	return s.Repo.Create(ctx, n)
}

func (s *Service) GetMyNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Notification, error) {
	return s.Repo.GetByUser(ctx, userID, limit, offset)
}

func (s *Service) GetUnread(ctx context.Context, userID uuid.UUID) ([]*Notification, error) {
	return s.Repo.GetUnread(ctx, userID)
}

func (s *Service) UnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.Repo.UnreadCount(ctx, userID)
}

func (s *Service) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	return s.Repo.MarkAsRead(ctx, id)
}

func (s *Service) MarkAsSent(ctx context.Context, id uuid.UUID) error {
	return s.Repo.MarkAsSent(ctx, id)
}

func (s *Service) MarkAsFailed(ctx context.Context, id uuid.UUID, reason string) error {
	return s.Repo.MarkAsFailed(ctx, id, reason)
}

func (s *Service) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	const batchSize = 20
	const sleepInterval = 50 * time.Millisecond // Space out executions to avoid DB exhaustion

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		rowsUpdated, err := s.Repo.MarkAllAsReadBatched(ctx, userID, batchSize)
		if err != nil {
			return fmt.Errorf("service failed to process batched status change: %w", err)
		}

		if rowsUpdated == 0 {
			break
		}

		time.Sleep(sleepInterval)
	}

	return nil
}