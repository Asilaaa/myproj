package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "userID"

type SessionCredentials struct {
	SessionToken string
	CookieHeader string
}

type Middleware struct {
	authService *Service
}

func NewMiddleware(authService *Service) *Middleware {
	return &Middleware{
		authService: authService,
	}
}

func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		credentials := extractSessionCredentials(r)

		session, err := m.authService.ValidateSession(r.Context(), credentials)
		if err != nil {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		userID, err := m.authService.GetUserID(session)
		if err != nil {
			http.Error(w, "invalid identity", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractSessionCredentials(r *http.Request) SessionCredentials {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return SessionCredentials{
				SessionToken: parts[1],
			}
		}
	}

	return SessionCredentials{
		CookieHeader: r.Header.Get("Cookie"),
	}
}

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
