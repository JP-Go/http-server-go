package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/JP-Go/http-server-go/internal/auth"
	"github.com/JP-Go/http-server-go/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type CreateUserRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *ApiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var body CreateUserRequestBody
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
	dbUser, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email: body.Email,
		HashedPassword: sql.NullString{
			String: auth.HashPassword(body.Password),
			Valid:  true,
		},
	})
	if err != nil {
		if err, ok := err.(*pq.Error); !ok || err.Constraint == "users_email_key" {
			RespondWithError(w, http.StatusConflict, "User already exists.")
			return
		}
		log.Printf("%s\n", err)
		RespondWithError(w, http.StatusInternalServerError, "Unexpected error. Contact administrators")
		return
	}
	RespondWithJSON(w, http.StatusCreated, User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	})
}

type LoginRequestBody struct {
	CreateUserRequestBody
	ExpiresIn int `json:"expires_in_seconds"`
}

type LoginResponseBody struct {
	User
	Token string `json:"token"`
}

func (cfg *ApiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
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
	expiresIn := time.Duration(body.ExpiresIn) * time.Second
	if expiresIn < time.Hour {
		expiresIn = time.Hour
	}
	token, err := auth.MakeJWT(dbUser.ID, cfg.JwtSecret, expiresIn)

	RespondWithJSON(w, http.StatusOK, LoginResponseBody{
		User: User{
			ID:        dbUser.ID,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
			Email:     dbUser.Email,
		},
		Token: token,
	})
}
