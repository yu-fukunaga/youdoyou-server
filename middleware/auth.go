package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

// AuthMiddleware validates Firebase ID tokens from the Authorization header
type AuthMiddleware struct {
	authClient *auth.Client
}

// NewAuthMiddleware creates a new AuthMiddleware instance
func NewAuthMiddleware(app *firebase.App) (*AuthMiddleware, error) {
	ctx := context.Background()
	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}

	return &AuthMiddleware{
		authClient: authClient,
	}, nil
}

// Handler は chi のミドルウェア形式 (func(http.Handler) http.Handler) です
func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		idToken := parts[1]

		token, err := m.authClient.VerifyIDToken(r.Context(), idToken)
		if err != nil {
			log.Printf("Error verifying ID token: %v", err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext retrieves the token information from the request context
func GetUserFromContext(ctx context.Context) (*auth.Token, bool) {
	token, ok := ctx.Value(UserContextKey).(*auth.Token)
	return token, ok
}
