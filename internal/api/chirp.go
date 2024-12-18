package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/JP-Go/http-server-go/internal/database"
	"github.com/google/uuid"
)

const maxChirpLength = 140
const profanityReplacement = "****"

var profaneWords = []string{"kerfuffle", "sharbert", "fornax"}

type ValidChirp struct {
	Chirp
	Valid bool
}

type Chirp struct {
	content string
}

func NewChirp(content string) Chirp {
	return Chirp{content}
}

func ValidateChirp(chirp Chirp) (ValidChirp, error) {
	if len(chirp.content) > 140 {
		return ValidChirp{Valid: false}, errors.New("Chirp too long")
	}
	if len(chirp.content) == 0 {
		return ValidChirp{Valid: false}, errors.New("Empty chirp")
	}
	return ValidChirp{Chirp: chirp, Valid: true}, nil
}

func CleanChirp(chirp ValidChirp, profaneWords []string, replacer string) (ValidChirp, error) {
	if !chirp.Valid {
		return chirp, errors.New("Cannot clean an invalid chirp")
	}
	words := strings.Split(chirp.content, " ")
	for i, word := range words {
		if slices.Contains(profaneWords, strings.ToLower(word)) {
			words[i] = profanityReplacement
		}
	}
	cleanedContent := strings.Join(words, " ")
	return ValidChirp{Valid: true, Chirp: NewChirp(cleanedContent)}, nil
}

type inputChirp struct {
	Body   string `json:"body"`
	UserID string `json:"user_id"`
}

type outputChirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (api *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	var chirp inputChirp
	err := decoder.Decode(&chirp)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Error decoding parameters. Invalid JSON")
		return
	}
	validChirp, err := ValidateChirp(NewChirp(chirp.Body))
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	validChirp, err = CleanChirp(validChirp, profaneWords, profanityReplacement)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	userId, err := uuid.Parse(chirp.UserID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid User ID")
		return
	}

	user, err := api.DB.GetUserByID(r.Context(), userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithError(w, http.StatusBadRequest, "User not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Unexpected error")
		}
		return
	}
	dbChirp, err := api.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		UserID: uuid.NullUUID{
			UUID:  user.ID,
			Valid: true,
		},
		Body: validChirp.content,
	})
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Unexpected error")
		return
	}
	output := outputChirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID.UUID,
	}
	RespondWithJSON(w, http.StatusCreated, output)
}

func (api *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request) {

	chirps, err := api.DB.GetChirps(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Unexpected error")
		return
	}

	output := make([]outputChirp, len(chirps))
	for i, chirp := range chirps {
		output[i] = outputChirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID.UUID,
		}
	}
	RespondWithJSON(w, http.StatusOK, output)
}

func (api *apiConfig) handleGetChirp(w http.ResponseWriter, r *http.Request) {
	uuid, err := uuid.Parse(r.PathValue("chirpID"))

	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid chirpID")
		return
	}

	chirp, err := api.DB.FindChirpByID(r.Context(), uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithError(w, http.StatusNotFound, "Chirp not found")
		} else {
			RespondWithError(w, http.StatusInternalServerError, "Unexpected error")
		}
		return
	}

	output := outputChirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID.UUID,
	}
	RespondWithJSON(w, http.StatusOK, output)
}
