package utils

import (
	"encoding/json"
	"net/http"
	"strings"
)

func RemoveBadWords(body string) string {
	words := strings.Split(body, " ")
	result := make([]string, 0)

	for _, word := range words {
		if isBadWord(word) {
			result = append(result, "****")
		} else {
			result = append(result, word)
		}
	}

	return strings.Join(result, " ")
}

func isBadWord(word string) bool {
	switch strings.ToLower(word) {
	case "kerfuffle", "sharbert", "fornax":
		return true
	default:
		return false
	}
}

func RespondWithError(w http.ResponseWriter, body map[string]string, status int) {
	data, err := json.Marshal(body)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}
