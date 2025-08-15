package provider

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/streadway/amqp"
	"github.com/szks-repo/usage-based-billing-sample/pkg/rabbitmq"
)

type ApiHandler struct {
	mqConn        *rabbitmq.Conn
	queue         amqp.Queue
	apiKeyChecker ApiKeyChecker
}

func NewApiHandler(
	mqConn *rabbitmq.Conn,
	queue amqp.Queue,
	apiKeyChecker ApiKeyChecker,
) *ApiHandler {
	return &ApiHandler{
		mqConn:        mqConn,
		queue:         queue,
		apiKeyChecker: apiKeyChecker,
	}
}

func (h *ApiHandler) HandleHelth(w http.ResponseWriter, r *http.Request) {
	slog.Info("Handling health check request")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}

func (h *ApiHandler) HandleApi1(w http.ResponseWriter, r *http.Request) {
	slog.Info("Handling API 1 request", "queueName", h.queue.Name)

	apiKey := r.Header.Get("x-api-key")
	if err := h.apiKeyChecker.Check(r.Context(), apiKey); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// todo move to interfaces
	if err := h.mqConn.Channel.Publish(
		"",           // exchange 
		h.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         []byte(fmt.Sprintf(`{"message": "Hello from API 1","messageId":"%s"}`, rand.Text())),
			DeliveryMode: amqp.Persistent,
			Headers: map[string]any{
				"x-api-key": apiKey,
				"timestamp": time.Now().Format(time.RFC3339Nano),
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
