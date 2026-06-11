package queue

import (
	"context"

	"github.com/hibiken/asynq"
)

// HandlerFunc is your business logic for one job type — the payload

type HandlerFunc[T any] func(ctx context.Context, payload T) error

// RegisterHandler is the single reusable entry point for registering ANY
// job type's processing function onto the worker mux.
//
// Decoding happens automatically. Malformed payloads are discarded
// (returns nil, no retry) since retrying won't fix bad JSON.
//

func RegisterHandler[T any](mux *asynq.ServeMux, def TaskDef[T], fn HandlerFunc[T]) {
	mux.HandleFunc(def.Name, func(ctx context.Context, t *asynq.Task) error {
		payload, err := def.Decode(t)
		if err != nil {
			return nil
		}
		return fn(ctx, payload)
	})
}