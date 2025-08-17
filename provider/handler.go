package provider

import (
	"log/slog"
	"math/rand/v2"
	"net/http"
	"time"
)

type ApiHandler struct {
}

func NewApiHandler() *ApiHandler {
	return &ApiHandler{}
}

func (h *ApiHandler) HandleHelth(w http.ResponseWriter, r *http.Request) {
	slog.Info("Handling health check request")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}

func (h *ApiHandler) HandleApi1(w http.ResponseWriter, r *http.Request) {
	slog.Info("Handling API 1 request")

	time.Sleep(time.Millisecond * time.Duration(rand.IntN(50)))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello from API 1!"))
}

func (h *ApiHandler) HandleApi2(w http.ResponseWriter, r *http.Request) {
	slog.Info("Handling API 2 request")

	time.Sleep(time.Millisecond * time.Duration(rand.IntN(200)))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello from API 2!"))
}
