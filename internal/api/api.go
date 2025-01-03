package api

import (
	"fmt"
	"log"
	"log/slog"
	"sync/atomic"

	"encoding/json"
	"net/http"

	"github.com/JP-Go/http-server-go/internal/database"
	"github.com/google/uuid"
)

type ApiConfig struct {
	serverHits  atomic.Int32
	DB          *database.Queries
	JwtSecret   string
	PolkaApiKey string
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
func parseUserIDFromRequest(r *http.Request) uuid.UUID {
	return uuid.MustParse(r.Context().Value(userIDKey).(string))
}

func (api *Api) RegisterEndpoints(fileServer http.Handler, server *http.ServeMux) {
	server.Handle("GET /app/", api.config.middlewareMetricsInc(fileServer))

	adminRoutes := http.NewServeMux()
	adminRoutes.HandleFunc("GET /metrics", api.config.metrics)
	adminRoutes.HandleFunc("POST /reset", api.config.resetMetrics)
	adminRoutes.HandleFunc("GET /api/healthz", readiness)

	apiRoutes := http.NewServeMux()
	apiRoutes.HandleFunc("GET /chirps", api.config.getChirps)
	apiRoutes.HandleFunc("POST /users", api.config.createUser)
	apiRoutes.HandleFunc("GET /chirps/{chirpID}", api.config.getChirp)

	apiRoutes.HandleFunc("POST /login", api.config.login)
	apiRoutes.HandleFunc("POST /refresh", api.config.refreshAccessToken)
	apiRoutes.HandleFunc("POST /revoke", api.config.revokeRefreshToken)
	apiRoutes.HandleFunc("POST /polka/webhooks", api.config.polkaUpgradeToChirpyRed)

	loggedInRoutes := http.NewServeMux()

	loggedInRoutes.Handle("DELETE /chirps/{chirpID}", http.HandlerFunc(api.config.deleteChirp))
	loggedInRoutes.Handle("POST /chirps", http.HandlerFunc(api.config.createChirp))
	loggedInRoutes.Handle("PUT /users", http.HandlerFunc(api.config.updateUser))
	apiRoutes.Handle("/", api.config.loggedInMiddleware(loggedInRoutes))

	server.Handle("/admin/", http.StripPrefix("/admin", adminRoutes))
	server.Handle("/api/", http.StripPrefix("/api", apiRoutes))

}

func (api *Api) Serve(mux *http.ServeMux, port int) {
	server := http.Server{
		Handler: mux,
		Addr:    fmt.Sprintf(":%d", port),
	}
	slog.Info(fmt.Sprintf("Server running on port :%d", port))
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Error initializing API. %s. Can not recover. Exiting", err)
	}
}

func OkResponse(w http.ResponseWriter, data any) {
	RespondWithJSON(w, http.StatusOK, data)
}

func BadRequestResponse(w http.ResponseWriter, msg string) {
	RespondWithError(w, http.StatusBadRequest, msg)
}

func NotFoundResponse(w http.ResponseWriter, msg string) {
	RespondWithError(w, http.StatusNotFound, msg)
}

func InternalServerErrorResponse(w http.ResponseWriter, msg string) {
	RespondWithError(w, http.StatusInternalServerError, msg)
}

func UnauthorizedResponse(w http.ResponseWriter, msg string) {
	RespondWithError(w, http.StatusUnauthorized, msg)
}

func ForbiddenResponse(w http.ResponseWriter, msg string) {
	RespondWithError(w, http.StatusForbidden, msg)
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
