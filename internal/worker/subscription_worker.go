package worker

import (
	"context"
	"log"
	"time"

	"github.com/ronexlemon/bnbcore/internal/domain/subscription"
)

type SubscriptionExpiryWorker struct {
	Repo     subscription.Repository
	Interval time.Duration
}

func NewSubscriptionExpiryWorker(repo subscription.Repository, interval time.Duration) *SubscriptionExpiryWorker {
	return &SubscriptionExpiryWorker{
		Repo:     repo,
		Interval: interval,
	}
}

func (w *SubscriptionExpiryWorker) Start(ctx context.Context) {
	log.Println("subscription expiry worker started")

	ticker := time.NewTicker(w.Interval)
	defer ticker.Stop()

	// Run immediately on start then on every tick
	w.run(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("subscription expiry worker stopped")
			return
		case <-ticker.C:
			w.run(ctx)
		}
	}
}

func (w *SubscriptionExpiryWorker) run(ctx context.Context) {
	log.Println("checking for expired subscriptions...")

	expired, err := w.Repo.GetExpired(ctx)
	if err != nil {
		log.Printf("failed to fetch expired subscriptions: %v", err)
		return
	}

	if len(expired) == 0 {
		log.Println("no expired subscriptions found")
		return
	}

	for _, sub := range expired {
		if err := w.Repo.UpdateStatus(ctx, sub.ID, subscription.StatusExpired); err != nil {
			log.Printf("failed to expire subscription %s: %v", sub.ID, err)
			continue
		}
		log.Printf("subscription %s for tenant %s marked as expired", sub.ID, sub.TenantID)
	}
}