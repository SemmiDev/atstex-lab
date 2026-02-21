package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/semmidev/atstex-lab/internal/domain"
	"github.com/semmidev/atstex-lab/internal/repository"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	OAuthConfig *oauth2.Config
	Repo        repository.Repository
}

func NewConfig(clientID, clientSecret, redirectURL string, r repository.Repository) *Config {
	return &Config{
		OAuthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
		Repo: r,
	}
}

func GenerateStateOauthCookie(w http.ResponseWriter) string {
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)
	return state
}

func GenerateSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

type contextKey string

const UserContextKey contextKey = "user"

func Middleware(repo repository.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_token")
			if err != nil {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			sess, err := repo.GetSession(r.Context(), cookie.Value)
			if err != nil || sess.ExpiresAt.Before(time.Now()) {
				clearCookie(w)
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			u, err := repo.GetUser(r.Context(), sess.UserID)
			if err != nil {
				clearCookie(w)
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminMiddleware restricts access to users with the "admin" role.
// Must be used after Middleware (which puts User in context).
func AdminMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := r.Context().Value(UserContextKey).(*domain.User)
			if !ok || u == nil || !u.IsAdmin() {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func OptionalUserMiddleware(repo repository.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_token")
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			sess, err := repo.GetSession(r.Context(), cookie.Value)
			if err != nil || sess.ExpiresAt.Before(time.Now()) {
				next.ServeHTTP(w, r)
				return
			}

			u, err := repo.GetUser(r.Context(), sess.UserID)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func clearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
}
