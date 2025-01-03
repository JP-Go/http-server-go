package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/JP-Go/http-server-go/internal/auth"
	"github.com/JP-Go/http-server-go/internal/database"
	"github.com/google/uuid"
)

const polkaUserUpgradedEvent = "user.upgraded"

type PolkaWebhookEvent struct {
	Event string `json:"event"`
	Data  struct {
		UserID string `json:"user_id"`
	} `json:"data"`
}

func (api *ApiConfig) polkaUpgradeToChirpyRed(w http.ResponseWriter, r *http.Request) {
	providedKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		UnauthorizedResponse(w, err.Error())
		return
	}
	if providedKey != api.PolkaApiKey {
		UnauthorizedResponse(w, "Invalid Polka API Key")
	}
	var body PolkaWebhookEvent
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
		BadRequestResponse(w, "Invalid JSON body")
		return
	}
	if body.Event != polkaUserUpgradedEvent {
		RespondWithJSON(w, http.StatusNoContent, struct{}{})
		return
	}
	userID, err := uuid.Parse(body.Data.UserID)
	if err != nil {
		BadRequestResponse(w, "Invalid user ID")
	}
	user, err := api.DB.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			NotFoundResponse(w, "User not found")
		} else {
			InternalServerErrorResponse(w, err.Error())
		}
		return
	}
	user, err = api.DB.UpgradeChirpyRed(r.Context(), database.UpgradeChirpyRedParams{
		ID:          user.ID,
		IsChirpyRed: true,
	})
	if err != nil {
		InternalServerErrorResponse(w, "Could not upgrade user account. Try again later.")
		return
	}
	RespondWithJSON(w, http.StatusNoContent, struct{}{})

}
