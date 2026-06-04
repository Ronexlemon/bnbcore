package notification
import (
	"context"
	"github.com/google/uuid"
)



type Repository interface {
	Create(ctx context.Context, n *Notification) (*Notification, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Notification, error)
	GetByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*Notification, error)
	GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Notification, error)
	GetUnread(ctx context.Context, userID uuid.UUID) ([]*Notification, error)
	 MarkAllAsReadBatched(ctx context.Context, userID uuid.UUID, batchSize int) (int64, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAsSent(ctx context.Context, id uuid.UUID) error
	MarkAsFailed(ctx context.Context, id uuid.UUID, reason string) error
	GetPending(ctx context.Context) ([]*Notification, error)
	UnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
}