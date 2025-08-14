package worker

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"log/slog"

	"github.com/szks-repo/usage-based-billing-sample/pkg/rabbitmq"
)

type Worker struct {
	mqConn *rabbitmq.Conn
}

func NewWorker(mqConn *rabbitmq.Conn) *Worker {
	return &Worker{
		mqConn: mqConn,
	}
}

func (w *Worker) Run(ctx context.Context) {
	slog.Info("Worker started")

	// Start worker logic here

	msgs, err := w.mqConn.Channel.Consume(
		"api1_queue", // queue name
		"worker1",    // consumer tag
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		slog.Error("Failed to register consumer", "error", err)
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				slog.Info("Message channel closed, exiting worker")
				return
			}
			dst := make(map[string]any)
			if err := json.Unmarshal(msg.Body, &dst); err != nil {
				slog.Error("Failed to unmarshal message", "error", err)
				msg.Nack(false, false) // nack the message
				continue
			}

			fmt.Println("Received message:", dst)

		}
	}
}
