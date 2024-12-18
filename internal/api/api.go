package api

import (
	"fmt"
	"os"
	"sync/atomic"

	"encoding/json"
	"net/http"

	"github.com/JP-Go/http-server-go/internal/database"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	DB             *database.Queries
}

type Api struct {
	apiConfig *apiConfig
}

type errorMessage struct {
	Message string `json:"error"`
}

func NewApi(db *database.Queries) *Api {
	return &Api{
		apiConfig: &apiConfig{
			fileServerHits: atomic.Int32{},
			DB:             db,
		},
	}
}

func (api *Api) RegisterEndpoints(fileServer http.Handler, server *http.ServeMux) {
	server.Handle("GET /app/", api.apiConfig.middlewareMetricsInc(fileServer))
	server.HandleFunc("GET /admin/metrics", api.apiConfig.handleMetrics)
	server.HandleFunc("POST /admin/reset", api.apiConfig.resetMetrics)
	server.HandleFunc("GET /api/healthz", handleReadiness)
	server.HandleFunc("POST /api/validate_chirp", handleValidateChirp)
	server.HandleFunc("POST /api/users", api.apiConfig.handleCreateUser)
}

func (api *Api) Serve(mux *http.ServeMux) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := http.Server{
		Handler: mux,
		Addr:    fmt.Sprintf(":%s", port),
	}
	server.ListenAndServe()
}

func RespondWithJSON(w http.ResponseWriter, status int, jsonMessage any) {
	message, marshalErr := json.Marshal(jsonMessage)
	if marshalErr != nil {
		w.Header().Add("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("There was an error with our system. Contact the administrator."))
		return
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(status)
	w.Write(message)
}

func RespondWithError(w http.ResponseWriter, status int, msg string) {
	message, marshalErr := json.Marshal(errorMessage{
		Message: msg,
	})
	if marshalErr != nil {
		w.Header().Add("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("There was an error with our system. Contact the administrator."))
		return
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(status)
	w.Write(message)
}
