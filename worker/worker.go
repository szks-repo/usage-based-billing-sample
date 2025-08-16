package worker

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/szks-repo/usage-based-billing-sample/pkg/rabbitmq"
	"github.com/szks-repo/usage-based-billing-sample/pkg/types"
)

type Worker struct {
	mqConn   *rabbitmq.Conn
	recorder *AccessLogRecorder
}

func NewWorker(
	mqConn *rabbitmq.Conn,
	recorder *AccessLogRecorder,
) *Worker {
	return &Worker{
		mqConn:   mqConn,
		recorder: recorder,
	}
}

func (w *Worker) Run(ctx context.Context) {
	slog.Info("Worker started")

	queue, err := w.mqConn.Channel.QueueDeclare(
		"api1_queue", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		slog.Error("Failed to declare queue", "error", err)
		return

	}
	slog.Info("Queue declared", "name", queue.Name)

	msgs, err := w.mqConn.Channel.Consume(
		queue.Name,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		slog.Error("Failed to register consumer", "error", err)
		return
	}

	w.recorder.Observe(ctx)
	defer w.recorder.Stop()

	slog.Info("Worker is ready to consume messages", "queue", queue.Name)
	for msg := range msgs {
		slog.Info("Received message", "body", string(msg.Body))

		var accessLog types.ApiAccessLog
		if err := json.Unmarshal(msg.Body, &accessLog); err != nil {
			slog.Error("Failed to unmarshal message", "error", err)
			msg.Nack(false, false)
			continue
		}

		w.recorder.Push(accessLog)
		msg.Ack(false)
	}
}
