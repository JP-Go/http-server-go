package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type CreateUserRequestBody struct {
	Email string `json:"email"`
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var body CreateUserRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Request body is not valid JSON.")
		return
	}
	if body.Email == "" {
		RespondWithError(w, http.StatusBadRequest, "Email must not be empty.")
		return
	}
	dbUser, err := cfg.DB.CreateUser(r.Context(), body.Email)
	if err != nil {
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
