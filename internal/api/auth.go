package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/JP-Go/http-server-go/internal/auth"
	"github.com/JP-Go/http-server-go/internal/database"
)

const defaultAccessTokenTTL = time.Hour
const defaultRefreshTokenTTL = time.Hour * 24 * 60

type LoginRequestBody = CreateUserRequestBody

type LoginResponseBody struct {
	User
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (cfg *ApiConfig) login(w http.ResponseWriter, r *http.Request) {
	var body LoginRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Request body is not valid JSON.")
		return
	}
	if body.Email == "" {
		RespondWithError(w, http.StatusBadRequest, "Email must not be empty.")
		return
	}

	if body.Password == "" {
		RespondWithError(w, http.StatusBadRequest, "Password must not be empty.")
		return
	}
	dbUser, err := cfg.DB.GetUserByEmail(r.Context(), body.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithError(w, http.StatusUnauthorized, "Invalid email or password.")
			return
		}
	}
	err = auth.VerifyPassword(body.Password, dbUser.HashedPassword.String)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Invalid email or password.")
		return
	}
	token, err := auth.MakeJWT(dbUser.ID, cfg.JwtSecret, defaultAccessTokenTTL)

	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Invalid email or password.")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error on login. Try again later.")
		return
	}
	_, err = cfg.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    dbUser.ID,
		ExpiresAt: time.Now().UTC().Add(defaultRefreshTokenTTL),
	})
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Error on login. Try again later.")
		return
	}

	RespondWithJSON(w, http.StatusOK, LoginResponseBody{
		User: User{
			ID:        dbUser.ID,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
			Email:     dbUser.Email,
		},
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (api *ApiConfig) refreshAccessToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		BadRequestResponse(w, "Missing authentication or authentication type invalid")
		return
	}
	userWithToken, err := api.DB.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			UnauthorizedResponse(w, "User not registered.")
		} else {
			InternalServerErrorResponse(w, "Unexpected error. Try again later.")
		}
		return
	}
	if userWithToken.ExpiresAt.Before(time.Now()) {
		UnauthorizedResponse(w, "Session expired")
		return
	}
	if userWithToken.RevokedAt.Valid {
		UnauthorizedResponse(w, "Session expired")
		return
	}

	token, err := auth.MakeJWT(userWithToken.ID, api.JwtSecret, defaultAccessTokenTTL)
	if err != nil {
		InternalServerErrorResponse(w, "Unexpected error. Try again later.")
		return
	}
	OkResponse(w, struct {
		Token string `json:"token"`
	}{
		Token: token,
	})

}

func (api *ApiConfig) revokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		BadRequestResponse(w, "Missing authentication or authentication type invalid")
		return
	}
	userWithToken, err := api.DB.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			NotFoundResponse(w, "Session not found")
		} else {
			InternalServerErrorResponse(w, "Unexpected error. Try again later.")
		}
		return
	}
	if userWithToken.ExpiresAt.Before(time.Now()) {
		UnauthorizedResponse(w, "Expired token")
		return
	}

	if userWithToken.RevokedAt.Valid {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	err = api.DB.RevokeRefreshToken(r.Context(), refreshToken)
	w.WriteHeader(http.StatusNoContent)
}
