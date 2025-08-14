package provider

import (
	"net/http"

	"github.com/szks-repo/usage-based-billing-sample/pkg/rabbitmq"
)

func NewApiServer(
	mqConn *rabbitmq.Conn,
	port string,
) *http.Server {

	handler := NewApiHandler(mqConn) // Assuming mqConn is not used in this example

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
