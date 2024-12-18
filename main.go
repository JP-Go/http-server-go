package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/JP-Go/http-server-go/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		fmt.Println("Misconfigured environment. Missing variable DB_URL")
		os.Exit(1)
	}
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		fmt.Println("Could not connect to database. Exiting")
		os.Exit(1)
	}
	apiCfg := apiConfig{
		fileServerHits: atomic.Int32{},
		db:             database.New(db),
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
