package main

import (
	"encoding/json"
	"net/http"
)

type errorMessage struct {
	Message string `json:"error"`
}

func RespondWithJSON(w http.ResponseWriter, status int, jsonMessage any) {
	message, marshalErr := json.Marshal(jsonMessage)
	if marshalErr != nil {
		w.Header().Add("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("There was an error with our system. Contact the administrator."))
		return
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(status)
	w.Write(message)
}

func RespondWithError(w http.ResponseWriter, status int, msg string) {
	message, marshalErr := json.Marshal(errorMessage{
		Message: msg,
	})
	if marshalErr != nil {
		w.Header().Add("content-type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("There was an error with our system. Contact the administrator."))
		return
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(status)
	w.Write(message)
}
