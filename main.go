package main

import (
	"net/http"
	"sync/atomic"
)

func main() {
	cfg := apiConfig{
		fileServerHits: atomic.Int32{},
	}
	mux := http.NewServeMux()
	fileServer := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("GET /app/", cfg.middlewareMetricsInc(fileServer))

	mux.HandleFunc("GET /api/healthz/", handleReadiness)
	mux.HandleFunc("GET /api/metrics/", cfg.handleMetrics)
	mux.HandleFunc("POST /api/reset/", cfg.resetMetrics)
	mux.HandleFunc("POST /api/reset", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api/reset/", http.StatusPermanentRedirect)
	})

	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	server.ListenAndServe()

}
