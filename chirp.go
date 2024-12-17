package main

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
)

const maxChirpLength = 140
const profanityReplacement = "****"

var profaneWords = []string{"kerfuffle", "sharbert", "fornax"}

func handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	type inputChirp struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	var chirp inputChirp
	err := decoder.Decode(&chirp)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Error decoding parameters. Invalid JSON")
		return
	}
	if len(chirp.Body) > 140 {
		RespondWithJSON(w, http.StatusBadRequest, "Chirp is too long")
		return
	}
	if len(chirp.Body) == 0 {
		RespondWithJSON(w, http.StatusBadRequest, "Chirp is too short")
		return
	}
	words := strings.Split(chirp.Body, " ")
	for i, word := range words {
		if slices.Contains(profaneWords, strings.ToLower(word)) {
			words[i] = profanityReplacement
		}
	}
	cleanedBody := strings.Join(words, " ")

	type validResponse struct {
		CleanedBody string `json:"cleaned_body"`
	}

	RespondWithJSON(w, http.StatusOK, validResponse{
		CleanedBody: cleanedBody,
	})

}
