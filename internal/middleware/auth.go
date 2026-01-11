package middleware

import (
	"net/http"

	"github.com/example/notes-template/internal/handlers"
	"github.com/example/notes-template/internal/services"
)

const sessionCookieName = "session_token"

type AuthMiddleware struct {
	authService *services.AuthService
	userService *services.UserService
}

func NewAuthMiddleware(authService *services.AuthService, userService *services.UserService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		userService: userService,
	}
}

// Authenticate validates the session cookie and adds user to context if valid.
// Does not reject unauthenticated requests.
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || cookie.Value == "" {
			next.ServeHTTP(w, r)
			return
		}

		user, err := m.authService.ValidateSession(r.Context(), cookie.Value)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := handlers.SetUserInContext(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth rejects unauthenticated requests with 401.
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := handlers.GetUserFromContext(r.Context())
		if user == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"Authentication required"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireSession mirrors RequireAuth for endpoints that must use a session.
func (m *AuthMiddleware) RequireSession(next http.Handler) http.Handler {
	return m.RequireAuth(next)
}
