
package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ronexlemon/bnbcore/internal/domain/notification"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/db"
)

type NotificationRepository struct {
	DbConnection *db.PostgresConn
}

func NewNotificationRepository(dbconn *db.PostgresConn) (*NotificationRepository, error) {
	if dbconn == nil {
		return nil, fmt.Errorf("db connection required")
	}
	return &NotificationRepository{DbConnection: dbconn}, nil
}

func (r *NotificationRepository) Create(ctx context.Context, n *notification.Notification) (*notification.Notification, error) {
	metadata, err := json.Marshal(n.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	err = r.DbConnection.Pool.QueryRow(ctx, `
		INSERT INTO notifications (
			id, tenant_id, user_id, type, channel, status,
			title, message, metadata, created_at
		)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, 'pending', $5, $6, $7, NOW())
		RETURNING id, created_at
	`,
		n.TenantID,
		n.UserID,
		n.Type,
		n.Channel,
		n.Title,
		n.Message,
		metadata,
	).Scan(&n.ID, &n.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}
	return n, nil
}

func (r *NotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*notification.Notification, error) {
	var n notification.Notification
	var metadata []byte

	err := r.DbConnection.Pool.QueryRow(ctx, `
		SELECT id, tenant_id, user_id, type, channel, status,
		       title, message, metadata, read_at, sent_at, failed_at, error, created_at
		FROM notifications
		WHERE id = $1
	`, id).Scan(
		&n.ID,
		&n.TenantID,
		&n.UserID,
		&n.Type,
		&n.Channel,
		&n.Status,
		&n.Title,
		&n.Message,
		&metadata,
		&n.ReadAt,
		&n.SentAt,
		&n.FailedAt,
		&n.Error,
		&n.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("notification not found")
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	if metadata != nil {
		_ = json.Unmarshal(metadata, &n.Metadata)
	}
	return &n, nil
}

func (r *NotificationRepository) GetByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*notification.Notification, error) {
	rows, err := r.DbConnection.Pool.Query(ctx, `
		SELECT id, tenant_id, user_id, type, channel, status,
		       title, message, metadata, read_at, sent_at, failed_at, error, created_at
		FROM notifications
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch notifications: %w", err)
	}
	defer rows.Close()

	return scanNotifications(rows)
}

func (r *NotificationRepository) GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*notification.Notification, error) {
	rows, err := r.DbConnection.Pool.Query(ctx, `
		SELECT id, tenant_id, user_id, type, channel, status,
		       title, message, metadata, read_at, sent_at, failed_at, error, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch notifications: %w", err)
	}
	defer rows.Close()

	return scanNotifications(rows)
}

func (r *NotificationRepository) GetUnread(ctx context.Context, userID uuid.UUID) ([]*notification.Notification, error) {
	rows, err := r.DbConnection.Pool.Query(ctx, `
		SELECT id, tenant_id, user_id, type, channel, status,
		       title, message, metadata, read_at, sent_at, failed_at, error, created_at
		FROM notifications
		WHERE user_id = $1
		  AND status != 'read'
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch unread notifications: %w", err)
	}
	defer rows.Close()

	return scanNotifications(rows)
}

func (r *NotificationRepository) GetPending(ctx context.Context) ([]*notification.Notification, error) {
	rows, err := r.DbConnection.Pool.Query(ctx, `
		SELECT id, tenant_id, user_id, type, channel, status,
		       title, message, metadata, read_at, sent_at, failed_at, error, created_at
		FROM notifications
		WHERE status = 'pending'
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending notifications: %w", err)
	}
	defer rows.Close()

	return scanNotifications(rows)
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	tag, err := r.DbConnection.Pool.Exec(ctx, `
		UPDATE notifications SET status = 'read', read_at = $1
		WHERE id = $2
	`, now, id)
	if err != nil {
		return fmt.Errorf("failed to mark as read: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("notification not found")
	}
	return nil
}

func (r *NotificationRepository) MarkAsSent(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	tag, err := r.DbConnection.Pool.Exec(ctx, `
		UPDATE notifications SET status = 'sent', sent_at = $1
		WHERE id = $2
	`, now, id)
	if err != nil {
		return fmt.Errorf("failed to mark as sent: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("notification not found")
	}
	return nil
}

func (r *NotificationRepository) MarkAsFailed(ctx context.Context, id uuid.UUID, reason string) error {
	now := time.Now()
	tag, err := r.DbConnection.Pool.Exec(ctx, `
		UPDATE notifications SET status = 'failed', failed_at = $1, error = $2
		WHERE id = $3
	`, now, reason, id)
	if err != nil {
		return fmt.Errorf("failed to mark as failed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("notification not found")
	}
	return nil
}

func (r *NotificationRepository) UnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.DbConnection.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM notifications
		WHERE user_id = $1 AND status NOT IN ('read', 'failed')
	`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count unread: %w", err)
	}
	return count, nil
}


func scanNotifications(rows pgx.Rows) ([]*notification.Notification, error) {
	var notifications []*notification.Notification
	for rows.Next() {
		var n notification.Notification
		var metadata []byte
		if err := rows.Scan(
			&n.ID,
			&n.TenantID,
			&n.UserID,
			&n.Type,
			&n.Channel,
			&n.Status,
			&n.Title,
			&n.Message,
			&metadata,
			&n.ReadAt,
			&n.SentAt,
			&n.FailedAt,
			&n.Error,
			&n.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		if metadata != nil {
			_ = json.Unmarshal(metadata, &n.Metadata)
		}
		notifications = append(notifications, &n)
	}
	return notifications, nil
}

func (r *NotificationRepository) MarkAllAsReadBatched(ctx context.Context, userID uuid.UUID, batchSize int) (int64, error) {
	if batchSize <= 0 {
		batchSize = 20 
	}

	query := `
		WITH target_notifications AS (
			SELECT id 
			FROM notifications
			WHERE user_id = $1 
			  AND status != 'read'::notification_status
			LIMIT $2
			FOR UPDATE SKIP LOCKED
		)
		UPDATE notifications SET 
			status = 'read'::notification_status, 
			read_at = NOW()
		WHERE id IN (SELECT id FROM target_notifications)
	`

	tag, err := r.DbConnection.Pool.Exec(ctx, query, userID, batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to batch mark notifications as read: %w", err)
	}

	return tag.RowsAffected(), nil
}
var _ notification.Repository = (*NotificationRepository)(nil)