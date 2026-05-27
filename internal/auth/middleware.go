package auth

import (
	"net/http"
	"strings"
)

func Middleware(tokens *TokenManager, audience string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeAuthError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}

		scheme, token, ok := strings.Cut(authHeader, " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
			writeAuthError(w, http.StatusUnauthorized, "invalid authorization header")
			return
		}

		principal, err := tokens.Verify(strings.TrimSpace(token), audience)
		if err != nil {
			writeAuthError(w, http.StatusUnauthorized, "invalid bearer token")
			return
		}

		next(w, r.WithContext(ContextWithPrincipal(r.Context(), principal)))
	}
}

func writeAuthError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(`{"error":"` + message + `"}` + "\n"))
}
