package queue

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
)

type Enqueuer struct {
	client *asynq.Client
}

func NewEnqueuer(cfg RedisConfig) (*Enqueuer, error) {
	client := asynq.NewClient(cfg.AsynqOpt())

	if err := client.Ping(); err != nil {
		return nil, fmt.Errorf("asynq enqueuer ping failed: %w", err)
	}

	return &Enqueuer{client: client}, nil
}

func (e *Enqueuer) Close() error {
	return e.client.Close()
}

func EnqueueTask[T any](ctx context.Context, e *Enqueuer, def TaskDef[T], payload T) (*asynq.TaskInfo, error) {
	task, err := def.NewTask(payload)
	if err != nil {
		return nil, err
	}

	info, err := e.client.EnqueueContext(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("enqueue %s: %w", def.Name, err)
	}
	return info, nil
}