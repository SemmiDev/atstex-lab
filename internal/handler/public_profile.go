package handler

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/semmidev/atstex-lab/internal/apperrors"
	"github.com/semmidev/atstex-lab/internal/auth"
	"github.com/semmidev/atstex-lab/internal/compiler"
	"github.com/semmidev/atstex-lab/internal/cvtemplate"
	"github.com/semmidev/atstex-lab/internal/domain"
	"github.com/semmidev/atstex-lab/internal/middleware"
	"github.com/semmidev/atstex-lab/internal/repository"
)

var usernameRegexp = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,28}[a-z0-9]$`)

// PublicProfile renders the public-facing profile page for a given username.
func (s *Server) handlePublicProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := chi.URLParam(r, "username")
		if username == "" {
			http.NotFound(w, r)
			return
		}

		profileUser, err := s.repo.GetUserByUsername(r.Context(), username)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				http.NotFound(w, r)
				return
			}
			s.reqLog(r).Error("failed to get user by username", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		profiles, err := s.repo.GetPublicCVProfilesByUserID(r.Context(), profileUser.ID)
		if err != nil {
			s.reqLog(r).Error("failed to get public profiles", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if profiles == nil {
			profiles = []domain.CVProfile{}
		}

		// Parse biodata JSON for each profile into a map for template rendering
		type profileWithBiodata struct {
			domain.CVProfile
			Bio map[string]interface{}
		}
		var enriched []profileWithBiodata
		profileImage := ""
		for _, p := range profiles {
			bio := map[string]interface{}{}
			if len(p.Biodata) > 2 { // not "{}"
				_ = json.Unmarshal(p.Biodata, &bio)
			}
			enriched = append(enriched, profileWithBiodata{CVProfile: p, Bio: bio})

			if profileImage == "" {
				if personal, ok := bio["personal"].(map[string]interface{}); ok {
					if img, ok := personal["photo"].(string); ok && img != "" {
						profileImage = img
					}
				}
			}
		}

		if profileImage == "" {
			profileImage = "https://ui-avatars.com/api/?name=" + url.QueryEscape(profileUser.Name) + "&background=random"
		}

		// The viewer may or may not be logged in
		viewer, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := map[string]interface{}{
			"ProfileUser":  profileUser,
			"Profiles":     enriched,
			"User":         viewer,
			"Username":     username,
			"ProfileImage": template.URL(profileImage),
		}
		if err := s.tmpl.ExecuteTemplate(w, "public_profile", data); err != nil {
			s.reqLog(r).Error("template error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

// PublicProfileDownloadPDF compiles a public profile's biodata to PDF and returns it.
func (s *Server) handlePublicProfileDownloadPDF() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := chi.URLParam(r, "username")
		profileIDStr := chi.URLParam(r, "profileID")

		profileUser, err := s.repo.GetUserByUsername(r.Context(), username)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		profileID, err := uuid.Parse(profileIDStr)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		profile, err := s.repo.GetCVProfile(r.Context(), profileID)
		if err != nil || profile.UserID != profileUser.ID || !profile.IsPublic {
			http.NotFound(w, r)
			return
		}

		// Parse biodata into CVData
		var cvData cvtemplate.CVData
		if parseErr := json.Unmarshal(profile.Biodata, &cvData); parseErr != nil {
			s.reqLog(r).Error("failed to parse biodata for PDF", "err", parseErr)
			http.Error(w, "failed to parse profile data", http.StatusInternalServerError)
			return
		}

		// Use the first available template
		templates, err := cvtemplate.List()
		if err != nil || len(templates) == 0 {
			s.reqLog(r).Error("no templates available for PDF generation", "err", err)
			http.Error(w, "no templates available", http.StatusInternalServerError)
			return
		}

		templateName := templates[0].Name
		ps := cvtemplate.DefaultPageSettings()

		rendered, err := cvtemplate.Render(templateName, cvData, ps, true)
		if err != nil {
			s.reqLog(r).Error("template render error for PDF", "err", err)
			http.Error(w, "failed to render template", http.StatusInternalServerError)
			return
		}

		opts := compiler.Options{
			Engine:  compiler.EnginePdfLatex,
			Timeout: 60 * time.Second,
		}

		result, err := compiler.Compile(r.Context(), []byte(rendered), opts)
		if err != nil {
			s.reqLog(r).Error("PDF compilation failed", "err", err)
			http.Error(w, "PDF compilation failed", http.StatusInternalServerError)
			return
		}

		filename := strings.ReplaceAll(profile.Title, " ", "_") + ".pdf"
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
		_, _ = w.Write(result.PDF)
	}
}

// PublishSettings renders the publish settings dashboard page.
func (s *Server) handlePublishSettings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		profiles, err := s.repo.GetCVProfilesByUserID(r.Context(), user.ID)
		if err != nil {
			s.reqLog(r).Error("failed to get profiles for publish settings", "err", err)
		}
		if profiles == nil {
			profiles = []domain.CVProfile{}
		}

		data := map[string]interface{}{
			"User":     user,
			"Profiles": profiles,
		}
		if err := s.tmpl.ExecuteTemplate(w, "publish", data); err != nil {
			s.reqLog(r).Error("template error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

// SetUsername handles PUT /api/username — validates and saves a username.
func (s *Server) handleSetUsername() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		var body struct {
			Username string `json:"username"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid request body"))
			return
		}

		username := strings.TrimSpace(strings.ToLower(body.Username))

		if !usernameRegexp.MatchString(username) {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("Username must be 3-30 characters, lowercase alphanumeric and hyphens only, cannot start or end with a hyphen"))
			return
		}

		// Check if the user already has this username (no-op)
		if user.Username != nil && *user.Username == username {
			middleware.Respond(w, r, http.StatusOK, map[string]interface{}{"status": "ok", "username": username})
			return
		}

		// Check availability
		available, err := s.repo.CheckUsernameAvailable(r.Context(), username)
		if err != nil {
			s.reqLog(r).Error("failed to check username availability", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("internal error")))
			return
		}
		if !available {
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("Username is already taken")))
			return
		}

		if err := s.repo.SetUsername(r.Context(), user.ID, username); err != nil {
			s.reqLog(r).Error("failed to set username", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to save username")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, map[string]interface{}{"status": "ok", "username": username})
	}
}

// CheckUsername handles GET /api/username/check?q=... — returns availability.
func (s *Server) handleCheckUsername() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		q := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("q")))

		if !usernameRegexp.MatchString(q) {
			middleware.Respond(w, r, http.StatusOK, map[string]interface{}{
				"available": false,
				"reason":    "Invalid format. Use 3-30 lowercase letters, numbers, and hyphens.",
			})
			return
		}

		// If the user already owns this username, it's available for them
		if user.Username != nil && strings.EqualFold(*user.Username, q) {
			middleware.Respond(w, r, http.StatusOK, map[string]interface{}{"available": true})
			return
		}

		available, err := s.repo.CheckUsernameAvailable(r.Context(), q)
		if err != nil {
			s.reqLog(r).Error("failed to check username", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("internal error")))
			return
		}

		reason := ""
		if !available {
			reason = "Username is already taken"
		}
		middleware.Respond(w, r, http.StatusOK, map[string]interface{}{
			"available": available,
			"reason":    reason,
		})
	}
}

// ToggleProfileVisibility handles PUT /api/cv-profiles/{id}/visibility.
func (s *Server) handleToggleProfileVisibility() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid profile id"))
			return
		}

		profile, err := s.repo.GetCVProfile(r.Context(), id)
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("profile not found")))
			return
		}
		if profile.UserID != user.ID {
			middleware.RespondError(w, r, apperrors.NewForbidden())
			return
		}

		var body struct {
			IsPublic bool `json:"is_public"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid request body"))
			return
		}

		if err := s.repo.UpdateCVProfileVisibility(r.Context(), id, body.IsPublic); err != nil {
			s.reqLog(r).Error("failed to update profile visibility", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to update visibility")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, map[string]interface{}{
			"status":    "ok",
			"is_public": body.IsPublic,
		})
	}
}
