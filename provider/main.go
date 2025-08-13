package provider

import "net/http"

func NewApiServer(mqUrl string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/hello", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	return server
}
