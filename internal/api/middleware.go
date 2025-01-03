package api

import (
	"context"
	"net/http"

	"github.com/JP-Go/http-server-go/internal/auth"
)

const userIDKey = "user.id"

func (api *ApiConfig) loggedInMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			UnauthorizedResponse(w, "No credentials provided")
			return
		}
		userID, err := auth.ValidateJWT(token, api.JwtSecret)
		if err != nil {
			UnauthorizedResponse(w, "Unauthorized")
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, userID.String())
		req := r.WithContext(ctx)
		next.ServeHTTP(w, req)
	})
}
