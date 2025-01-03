package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/JP-Go/http-server-go/internal/auth"
	"github.com/JP-Go/http-server-go/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

type CreateUserRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (cfg *ApiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	var body CreateUserRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		InternalServerErrorResponse(w, "Request body is not valid JSON.")
		return
	}
	if body.Email == "" {
		BadRequestResponse(w, "Email must not be empty.")
		return
	}

	if body.Password == "" {
		BadRequestResponse(w, "Password must not be empty.")
		return
	}
	dbUser, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:          body.Email,
		HashedPassword: auth.HashPassword(body.Password),
	})
	if err != nil {
		if err, ok := err.(*pq.Error); !ok || err.Constraint == "users_email_key" {
			RespondWithError(w, http.StatusConflict, "User already exists.")
			return
		}
		log.Printf("%s\n", err)
		InternalServerErrorResponse(w, "Unexpected error. Contact administrators")
		return
	}
	RespondWithJSON(w, http.StatusCreated, User{
		ID:          dbUser.ID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
		Email:       dbUser.Email,
		IsChirpyRed: dbUser.IsChirpyRed,
	})
}

func (api *ApiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	userID := parseUserIDFromRequest(r)
	type input = CreateUserRequestBody
	var body input
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		InternalServerErrorResponse(w, "Request body is not valid JSON.")
		return
	}
	if body.Email == "" {
		BadRequestResponse(w, "Email must not be empty.")
		return
	}
	if body.Password == "" {
		BadRequestResponse(w, "Password must not be empty.")
		return
	}
	hashedPassword := auth.HashPassword(body.Password)

	user, err := api.DB.UpdateUserCredentials(r.Context(), database.UpdateUserCredentialsParams{
		ID:             userID,
		Email:          body.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code.Name() == "unique_violation" {
			BadRequestResponse(w, "Email already taken")
		}
		return
	}
	OkResponse(w, User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}
