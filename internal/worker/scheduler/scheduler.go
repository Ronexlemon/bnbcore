package scheduler

import (
	"context"
	"github.com/go-co-op/gocron/v2"
	"github.com/ronexlemon/bnbcore/internal/worker"
)

type Scheduler struct {
	s *gocron.Scheduler
}

func New(worker *worker.CheckoutNotificationWorker, ctx context.Context) (*Scheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	_, err = s.NewJob(
		gocron.CronJob("0 0 * * *", false),
		gocron.NewTask(func() {
			 worker.ProcessTodayCheckouts(ctx);
		}),
	)
	if err != nil {
		return nil, err
	}

	_, err = s.NewJob(
		gocron.CronJob("0 6 * * *", false),
		gocron.NewTask(func() {
			 worker.ProcessTodayCheckouts(ctx)
		}),
	)
	if err != nil {
		return nil, err
	}

	return &Scheduler{s:&s}, nil
}
func (s *Scheduler) Start() {
	s.Start()
}