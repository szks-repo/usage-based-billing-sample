package provider

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/streadway/amqp"

	"github.com/szks-repo/usage-based-billing-sample/pkg/rabbitmq"
	"github.com/szks-repo/usage-based-billing-sample/pkg/types"
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

	// todo move to api before middleware
	apiKey := r.Header.Get("x-api-key")
	if err := h.apiKeyChecker.Check(r.Context(), apiKey); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// todo move to after middleware
	payload, err := json.Marshal(&types.AccessLog{
		Timestamp:  time.Now(),
		ClientIP:   r.RemoteAddr,
		Path:       r.URL.Path,
		Method:     r.Method,
		Protocol:   r.URL.Scheme,
		StatusCode: 200, //todo
		Latency:    int64(time.Millisecond * 10),
		UserAgent:  r.UserAgent(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.mqConn.Channel.Publish(
		"",           // exchange
		h.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         payload,
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
