package provider

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/streadway/amqp"
	"github.com/szks-repo/usage-based-billing-sample/pkg/rabbitmq"
)

type ApiHandler struct {
	mqConn *rabbitmq.Conn
}

func NewApiHandler(mqConn *rabbitmq.Conn) *ApiHandler {
	return &ApiHandler{
		mqConn: mqConn,
	}
}

func (h *ApiHandler) HandleHelth(w http.ResponseWriter, r *http.Request) {
	slog.Info("Handling health check request")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}

func (h *ApiHandler) HandleApi1(w http.ResponseWriter, r *http.Request) {
	slog.Info("Handling API 1 request")

	// todo move to interfaces
	if err := h.mqConn.Channel.Publish(
		"",           // exchange
		"api1_queue", // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         []byte(`{"message": "Hello from API 1"}`),
			DeliveryMode: amqp.Persistent, // ensure message is persistent
			Headers: map[string]any{
				"source": "api1",
			},
			Timestamp: time.Now(),
		},
	); err != nil {
		slog.Error("Failed to publish message to RabbitMQ", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello from API 1!"))
}

func (h *ApiHandler) HandleApi2(w http.ResponseWriter, r *http.Request) {
	slog.Info("Handling API 2 request")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello from API 2!"))
}
