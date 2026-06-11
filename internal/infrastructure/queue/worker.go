package queue

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

type WorkerConfig struct {
	Concurrency     int
	Queues          map[string]int
	StrictPriority  bool
	ShutdownTimeout time.Duration
}

func DefaultWorkerConfig() WorkerConfig {
	return WorkerConfig{
		Concurrency: 50,
		Queues: map[string]int{
			"created": 2, // ~14 workers — Twilio template sends
			"status":  2, // ~14 workers — Twilio status sends
			"signup":  3, // ~22 workers — SMTP signup emails
			"pass_reset":2,
		},
		StrictPriority:  false,
		ShutdownTimeout: 30 * time.Second,
	}
}

type Worker struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

func NewWorker(cfg RedisConfig, workerCfg WorkerConfig) (*Worker, error) {
	server := asynq.NewServer(
		cfg.AsynqOpt(),
		asynq.Config{
			Concurrency:     workerCfg.Concurrency,
			Queues:          workerCfg.Queues,
			StrictPriority:  workerCfg.StrictPriority,
			ShutdownTimeout: workerCfg.ShutdownTimeout,
			RetryDelayFunc: func(n int, err error, t *asynq.Task) time.Duration {
				return time.Duration(n*n) * 5 * time.Second
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Printf("[worker] task %s permanently failed: %v", task.Type(), err)
			}),
			Logger:              newAsynqLogger(),
			HealthCheckInterval: 15 * time.Second,
		},
	)

	return &Worker{
		server: server,
		mux:    asynq.NewServeMux(),
	}, nil
}

// Mux exposes the ServeMux so any processor can call queue.RegisterHandler on it.
func (w *Worker) Mux() *asynq.ServeMux {
	return w.mux
}

func (w *Worker) Run() error {
	if err := w.server.Run(w.mux); err != nil {
		return fmt.Errorf("asynq worker run: %w", err)
	}
	return nil
}

func (w *Worker) Shutdown() {
	w.server.Shutdown()
}

type asynqLogger struct{}

func newAsynqLogger() *asynqLogger { return &asynqLogger{} }

func (l *asynqLogger) Debug(args ...any) {}
func (l *asynqLogger) Info(args ...any)  {}
func (l *asynqLogger) Warn(args ...any)  { log.Println(append([]any{"[worker:warn]"}, args...)...) }
func (l *asynqLogger) Error(args ...any) { log.Println(append([]any{"[worker:error]"}, args...)...) }
func (l *asynqLogger) Fatal(args ...any) { log.Fatal(append([]any{"[worker:fatal]"}, args...)...) }