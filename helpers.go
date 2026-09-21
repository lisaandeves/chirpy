package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

type responseError struct {
	Error string `json:"error"`
}

func chirpCensor(text string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	censorText := "****"

	words := strings.Split(text, " ")
	for i, word := range words {
		wordLower := strings.ToLower(word)
		if slices.Contains(badWords, wordLower) {
			words[i] = censorText
		}
	}
	textCensored := strings.Join(words, " ")
	return textCensored
}

func writeErrorJson(writer http.ResponseWriter, err error, msg string, code int) {
	writer.Header().Set("Content-Type", "application/json")
	resp := responseError{Error: fmt.Sprintf("%s: %s", msg, err.Error())}
	dat, err := json.Marshal(resp)
	if err != nil {
		//Unreachable
	}
	writer.WriteHeader(code)
	writer.Write(dat)
}

func writeResponseJson(writer http.ResponseWriter, resp any, code int) {
	writer.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(resp)
	if err != nil {
		writeErrorJson(writer, err, "Couldn't marshall JSON", http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(code)
	writer.Write(dat)
}
