package handlers

import (
	"log"
	"net/http"
	"os"
	"strings"
)

// TokenVerificationMiddleware with set of allowed tokens
type TokenVerificationMiddleware struct {
	apiTokens []string
}

// NewTokenVerificationMiddleware constructs a new TokenVerificationMiddleware instance
func NewTokenVerificationMiddleware() *TokenVerificationMiddleware {
	return &TokenVerificationMiddleware{}
}

// Initialize will prepare the instance for use
func (t *TokenVerificationMiddleware) Initialize() {
	t.apiTokens = strings.Split(os.Getenv("STREAMMARKER_DATA_ACCESS_API_TOKENS"), ",")
}

// Run the middleware to verify the request includes a valid token
func (t *TokenVerificationMiddleware) Run(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	suppliedAPIToken := r.Header.Get("X-API-KEY")
	found := false
	for _, token := range t.apiTokens {
		if suppliedAPIToken == token {
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
