package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileServerHits.Store(0)
	w.Write([]byte("OK"))
}

func main() {
	cfg := apiConfig{
		fileServerHits: atomic.Int32{},
	}
	mux := http.NewServeMux()
	fileServer := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("GET /app/", cfg.middlewareMetricsInc(fileServer))
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		hitsMsg := fmt.Sprintf("Hits: %d", cfg.fileServerHits.Load())
		w.Write([]byte(hitsMsg))
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("POST /reset", cfg.resetMetrics)
	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}
	server.ListenAndServe()

}
