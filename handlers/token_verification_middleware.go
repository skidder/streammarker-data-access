package handlers

import (
	"log"
	"net/http"
	"os"
	"strings"
)

type TokenVerificationMiddleware struct {
	apiTokens []string
}

func NewTokenVerificationMiddleware() *TokenVerificationMiddleware {
	return &TokenVerificationMiddleware{}
}

func (t *TokenVerificationMiddleware) Initialize() {
	t.apiTokens = strings.Split(os.Getenv("STREAMMARKER_DATA_ACCESS_API_TOKENS"), ",")
}

func (t *TokenVerificationMiddleware) Run(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	suppliedApiToken := r.Header.Get("X-API-KEY")
	found := false
	for _, token := range t.apiTokens {
		if suppliedApiToken == token {
			found = true
			break
		}
	}
	if !found {
		log.Println("No valid API key was present in request, rejecting at middleware")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	next(w, r)
}
