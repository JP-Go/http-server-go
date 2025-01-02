package api

import (
	"database/sql"
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
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
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
		InternalServerErrorResponse(w, "Unexpected error. Contact administrators")
		return
	}
	RespondWithJSON(w, http.StatusCreated, User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	})
}
