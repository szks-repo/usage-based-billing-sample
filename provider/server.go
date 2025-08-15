package provider

import (
	"log/slog"
	"net/http"

	"github.com/szks-repo/usage-based-billing-sample/pkg/rabbitmq"
)

func NewApiServer(
	mqConn *rabbitmq.Conn,
	apiKeyChecker ApiKeyChecker,
	port string,
) *http.Server {
	queue, err := mqConn.Channel.QueueDeclare(
		"api1_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		slog.Error("Failed to declare queue", "error", err)
		panic(err)
	}

	handler := NewApiHandler(
		mqConn,
		queue,
		apiKeyChecker,
	)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", handler.HandleHelth)
	mux.HandleFunc("GET /api/v1/one", handler.HandleApi1)
	mux.HandleFunc("GET /api/v1/two", handler.HandleApi2)

	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}
	return server
}
