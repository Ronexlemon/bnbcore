package queue

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)


type TaskDef[T any] struct{
	Name    string
	Queue   string
	Retry   int
	Timeout time.Duration
	Unique  time.Duration
}

func (d TaskDef[T]) NewTask(payload T) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal %s payload: %w", d.Name, err)
	}

	opts := []asynq.Option{
		asynq.Queue(d.Queue),
		asynq.MaxRetry(d.Retry),
		asynq.Timeout(d.Timeout),
	}
	if d.Unique > 0 {
		opts = append(opts, asynq.Unique(d.Unique))
	}

	return asynq.NewTask(d.Name, data, opts...), nil
}

// Decode parses a task's raw payload back into T.
func (d TaskDef[T]) Decode(t *asynq.Task) (T, error) {
	var payload T
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return payload, fmt.Errorf("unmarshal %s payload: %w", d.Name, err)
	}
	return payload, nil
}