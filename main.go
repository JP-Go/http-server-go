package main

import (
	"net/http"
	"sync/atomic"
)

func main() {
	apiCfg := apiConfig{
		fileServerHits: atomic.Int32{},
	}
	mux := http.NewServeMux()
	fileServer := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("GET /app", apiCfg.middlewareMetricsInc(fileServer))

	mux.HandleFunc("GET /admin/metrics", apiCfg.handleMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetrics)

	mux.HandleFunc("GET /api/healthz", handleReadiness)
	mux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)

	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	server.ListenAndServe()
}
