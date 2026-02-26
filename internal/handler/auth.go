package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/semmidev/atstex-lab/internal/auth"
)

type googleUser struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (s *Server) handleGoogleLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := auth.GenerateStateOauthCookie(w)
		url := s.authConfig.OAuthConfig.AuthCodeURL(state)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func (s *Server) handleGoogleCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("oauthstate")
		if err != nil {
			s.reqLog(r).Error("oauth cookie missing", "err", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		if r.FormValue("state") != cookie.Value {
			s.reqLog(r).Error("invalid oauth state")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		code := r.FormValue("code")
		token, err := s.authConfig.OAuthConfig.Exchange(context.Background(), code)
		if err != nil {
			s.reqLog(r).Error("oauth exchange failed", "err", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
		if err != nil {
			s.reqLog(r).Error("failed to get user info", "err", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			s.reqLog(r).Error("failed to read body", "err", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		var gUser googleUser
		if err := json.Unmarshal(body, &gUser); err != nil {
			s.reqLog(r).Error("failed to parse user info", "err", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		u, err := s.repo.UpsertUser(r.Context(), gUser.ID, gUser.Email, gUser.Name, gUser.Picture)
		if err != nil {
			s.reqLog(r).Error("failed to upsert user", "err", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		sessionToken, err := auth.GenerateSessionToken()
		if err != nil {
			s.reqLog(r).Error("failed to generate session", "err", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		ipAddress := getClientIP(r)
		userAgent := r.UserAgent()

		// 30 days session
		expiresAt := time.Now().Add(30 * 24 * time.Hour)
		if err := s.repo.CreateSession(r.Context(), u.ID, sessionToken, ipAddress, userAgent, expiresAt); err != nil {
			s.reqLog(r).Error("failed to store session", "err", err)
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    sessionToken,
			Expires:  expiresAt,
			HttpOnly: true,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})

		// Redirect based on role
		if u.IsAdmin() {
			http.Redirect(w, r, "/admin", http.StatusTemporaryRedirect)
		} else {
			http.Redirect(w, r, "/input", http.StatusTemporaryRedirect)
		}
	}
}

func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

func (s *Server) handleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err == nil {
			s.repo.DeleteSession(r.Context(), cookie.Value)
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    "",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (s *Server) handleDeleteAccount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.reqLog(r).Info("Delete account requested")
		if err := r.ParseForm(); err != nil {
			s.reqLog(r).Error("delete account form parse failed", "err", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		sess, err := s.repo.GetSession(r.Context(), cookie.Value)
		if err == nil {
			s.repo.SoftDeleteUser(r.Context(), sess.UserID)
			s.repo.DeleteSession(r.Context(), cookie.Value)
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    "",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
