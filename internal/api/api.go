package api

import (
	"fmt"
	"log/slog"
	"sync/atomic"

	"encoding/json"
	"net/http"

	"github.com/JP-Go/http-server-go/internal/database"
)

type ApiConfig struct {
	serverHits atomic.Int32
	DB         *database.Queries
	JwtSecret  string
}

type Api struct {
	config *ApiConfig
}

type errorMessage struct {
	Message string `json:"error"`
}

func NewApi(apiConfig *ApiConfig) *Api {
	apiConfig.serverHits.Store(0)
	return &Api{config: apiConfig}
}

func (api *Api) RegisterEndpoints(fileServer http.Handler, server *http.ServeMux) {
	server.Handle("GET /app/", api.config.middlewareMetricsInc(fileServer))
	server.HandleFunc("GET /admin/metrics", api.config.metrics)
	server.HandleFunc("POST /admin/reset", api.config.resetMetrics)
	server.HandleFunc("GET /api/healthz", readiness)
	server.HandleFunc("POST /api/chirps", api.config.createChirp)
	server.HandleFunc("GET /api/chirps", api.config.getChirps)
	server.HandleFunc("GET /api/chirps/{chirpID}", api.config.getChirp)
	server.HandleFunc("POST /api/users", api.config.createUser)
	server.HandleFunc("POST /api/login", api.config.login)
	server.HandleFunc("POST /api/refresh", api.config.refreshAccessToken)
	server.HandleFunc("POST /api/revoke", api.config.revokeRefreshToken)
}

func (api *Api) Serve(mux *http.ServeMux, port int) {
	server := http.Server{
		Handler: mux,
		Addr:    fmt.Sprintf(":%d", port),
	}
	slog.Info(fmt.Sprintf("Server running on %d", port))
	server.ListenAndServe()
}

func BadRequestResponse(w http.ResponseWriter, msg string) {
	RespondWithError(w, http.StatusBadRequest, msg)
}

func InternalServerErrorResponse(w http.ResponseWriter, msg string) {
	RespondWithError(w, http.StatusInternalServerError, msg)
}

func NotFoundResponse(w http.ResponseWriter, msg string) {
	RespondWithError(w, http.StatusNotFound, msg)
}

func UnauthorizedResponse(w http.ResponseWriter, msg string) {
	RespondWithError(w, http.StatusUnauthorized, msg)
}

func OkResponse(w http.ResponseWriter, data any) {
	RespondWithJSON(w, http.StatusOK, data)
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
