package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/szks-repo/usage-based-billing-sample/pkg/rabbitmq"
	"github.com/szks-repo/usage-based-billing-sample/pkg/types"
)

type Worker struct {
	mqConn   *rabbitmq.Conn
	s3Writer *S3Writer
}

func NewWorker(
	mqConn *rabbitmq.Conn,
	s3Client *s3.Client,
) *Worker {
	s3Writer := NewS3Writer(
		s3Client,
		"api-access-log",
		1024<<10*5,
		time.Second*30,
	)

	return &Worker{
		mqConn:   mqConn,
		s3Writer: s3Writer,
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

	w.s3Writer.Start(ctx)
	defer w.s3Writer.Stop()

	slog.Info("Worker is ready to consume messages", "queue", queue.Name)
	for msg := range msgs {
		slog.Info("Received message", "body", string(msg.Body))

		var accessLog types.AccessLog
		if err := json.Unmarshal(msg.Body, &accessLog); err != nil {
			slog.Error("Failed to unmarshal message", "error", err)
			msg.Nack(false, false)
			continue
		}

		w.s3Writer.AddLog(accessLog)
		msg.Ack(false)
	}
}
