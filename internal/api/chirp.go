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
	Body string `json:"body"`
}

type outputChirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (api *ApiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	userID := parseUserIDFromRequest(r)

	decoder := json.NewDecoder(r.Body)
	var chirp inputChirp
	err := decoder.Decode(&chirp)
	if err != nil {
		BadRequestResponse(w, "Error decoding parameters. Invalid JSON")
		return
	}
	validChirp, err := ValidateChirp(NewChirp(chirp.Body))
	if err != nil {
		BadRequestResponse(w, err.Error())
		return
	}
	validChirp, err = CleanChirp(validChirp, profaneWords, profanityReplacement)
	if err != nil {
		BadRequestResponse(w, err.Error())
		return
	}

	user, err := api.DB.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			BadRequestResponse(w, "User not found")
		} else {
			InternalServerErrorResponse(w, "Unexpected error")
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
		InternalServerErrorResponse(w, "Unexpected error")
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

func (api *ApiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	authorID := r.URL.Query().Get("author_id")
	sort := r.URL.Query().Get("sort")
	var chirps []database.Chirp
	if authorID == "" {
		allChirps, err := api.DB.GetChirps(r.Context())
		if err != nil {
			InternalServerErrorResponse(w, "Unexpected error")
			return
		}
		chirps = allChirps
	} else {
		authorID, err := uuid.Parse(authorID)
		userChirps, err := api.DB.FindChirpsFromUser(r.Context(), uuid.NullUUID{
			UUID:  authorID,
			Valid: true,
		})
		if err != nil {
			InternalServerErrorResponse(w, "Unexpected error")
			return
		}
		chirps = userChirps
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
	if sort == "desc" {
		slices.Reverse(output)
	}
	OkResponse(w, output)
}

func (api *ApiConfig) getChirp(w http.ResponseWriter, r *http.Request) {
	uuid, err := uuid.Parse(r.PathValue("chirpID"))

	if err != nil {
		BadRequestResponse(w, "Invalid chirpID")
		return
	}

	chirp, err := api.DB.FindChirpByID(r.Context(), uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			NotFoundResponse(w, "Chirp not found")
		} else {
			InternalServerErrorResponse(w, "Unexpected error")
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
	OkResponse(w, output)
}

func (api *ApiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {
	userID := parseUserIDFromRequest(r)
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))

	if err != nil {
		BadRequestResponse(w, "Invalid chirpID")
		return
	}
	chirp, err := api.DB.FindChirpByID(r.Context(), chirpID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			NotFoundResponse(w, "Chirp not found")
		} else {
			InternalServerErrorResponse(w, "Unexpected error")
		}
		return
	}
	if chirp.UserID.UUID != userID {
		ForbiddenResponse(w, "Chirp does not belong to your user")
		return
	}
	err = api.DB.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		InternalServerErrorResponse(w, "Could not delete chirp. Try again later")
	}
	RespondWithJSON(w, http.StatusNoContent, struct{}{})
}
