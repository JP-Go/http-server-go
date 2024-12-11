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
	mux.Handle("GET /app", cfg.middlewareMetricsInc(fileServer))

	mux.HandleFunc("GET /admin/metrics", cfg.handleMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.resetMetrics)

	mux.HandleFunc("GET /api/healthz", handleReadiness)

	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	server.ListenAndServe()
}
