package worker

import (
	"context"
	"log/slog"
	"time"
)

type Worker struct {
	mqUrl string
}

func NewWorker(mqUrl string) *Worker {
	return &Worker{
		mqUrl: mqUrl,
	}
}

func (w *Worker) Run(ctx context.Context) {
	slog.Info("Worker started")

	// Start worker logic here
	for {
		time.Sleep(3 * time.Second) // Simulate work
	}
}
