package provider

import (
	"net/http"
)

func NewApiServer(
	port string,
	mw Middleware,
) *http.Server {
	handler := NewApiHandler()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", handler.HandleHelth)
	mux.Handle("GET /api/v1/one", mw.Wrap(http.HandlerFunc(handler.HandleApi1)))
	mux.Handle("GET /api/v1/two", mw.Wrap(http.HandlerFunc(handler.HandleApi2)))

	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}
	return server
}
