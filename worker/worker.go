package worker

import (
	"context"
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
	// Start worker logic here
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()
}
