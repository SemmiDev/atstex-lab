package handler

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/semmidev/atstex-lab/internal/auth"
	"github.com/semmidev/atstex-lab/internal/compiler"
	"github.com/semmidev/atstex-lab/internal/cvtemplate"
	"github.com/semmidev/atstex-lab/internal/domain"
	"github.com/semmidev/atstex-lab/internal/repository"
)

var usernameRegexp = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,28}[a-z0-9]$`)

// PublicProfile renders the public-facing profile page for a given username.
func (h *Handler) PublicProfile(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	if username == "" {
		http.NotFound(w, r)
		return
	}

	profileUser, err := h.repo.GetUserByUsername(r.Context(), username)
	if err != nil {
		if err == repository.ErrNotFound {
			http.NotFound(w, r)
			return
		}
		h.reqLog(r).Error("failed to get user by username", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	profiles, err := h.repo.GetPublicCVProfilesByUserID(r.Context(), profileUser.ID)
	if err != nil {
		h.reqLog(r).Error("failed to get public profiles", "err", err)
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
	for _, p := range profiles {
		bio := map[string]interface{}{}
		if len(p.Biodata) > 2 { // not "{}"
			_ = json.Unmarshal(p.Biodata, &bio)
		}
		enriched = append(enriched, profileWithBiodata{CVProfile: p, Bio: bio})
	}

	// The viewer may or may not be logged in
	viewer, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]interface{}{
		"ProfileUser": profileUser,
		"Profiles":    enriched,
		"User":        viewer,
		"Username":    username,
	}
	if err := h.tmpl.ExecuteTemplate(w, "public_profile", data); err != nil {
		h.reqLog(r).Error("template error", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// PublicProfileDownloadPDF compiles a public profile's biodata to PDF and returns it.
func (h *Handler) PublicProfileDownloadPDF(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")
	profileIDStr := chi.URLParam(r, "profileID")

	profileUser, err := h.repo.GetUserByUsername(r.Context(), username)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	profileID, err := uuid.Parse(profileIDStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	profile, err := h.repo.GetCVProfile(r.Context(), profileID)
	if err != nil || profile.UserID != profileUser.ID || !profile.IsPublic {
		http.NotFound(w, r)
		return
	}

	// Parse biodata into CVData
	var cvData cvtemplate.CVData
	if err := json.Unmarshal(profile.Biodata, &cvData); err != nil {
		h.reqLog(r).Error("failed to parse biodata for PDF", "err", err)
		http.Error(w, "failed to parse profile data", http.StatusInternalServerError)
		return
	}

	// Use the first available template
	templates, err := cvtemplate.List()
	if err != nil || len(templates) == 0 {
		h.reqLog(r).Error("no templates available for PDF generation", "err", err)
		http.Error(w, "no templates available", http.StatusInternalServerError)
		return
	}

	templateName := templates[0].Name
	ps := cvtemplate.DefaultPageSettings()

	rendered, err := cvtemplate.Render(templateName, cvData, ps)
	if err != nil {
		h.reqLog(r).Error("template render error for PDF", "err", err)
		http.Error(w, "failed to render template", http.StatusInternalServerError)
		return
	}

	opts := compiler.Options{
		Engine:  compiler.EnginePdfLatex,
		Timeout: 60 * time.Second,
	}

	result, err := compiler.Compile(r.Context(), []byte(rendered), opts)
	if err != nil {
		h.reqLog(r).Error("PDF compilation failed", "err", err)
		http.Error(w, "PDF compilation failed", http.StatusInternalServerError)
		return
	}

	filename := strings.ReplaceAll(profile.Title, " ", "_") + ".pdf"
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Write(result.PDF)
}

// PublishSettings renders the publish settings dashboard page.
func (h *Handler) PublishSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

	profiles, err := h.repo.GetCVProfilesByUserID(r.Context(), user.ID)
	if err != nil {
		h.reqLog(r).Error("failed to get profiles for publish settings", "err", err)
	}
	if profiles == nil {
		profiles = []domain.CVProfile{}
	}

	data := map[string]interface{}{
		"User":     user,
		"Profiles": profiles,
	}
	if err := h.tmpl.ExecuteTemplate(w, "publish", data); err != nil {
		h.reqLog(r).Error("template error", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// SetUsername handles PUT /api/username — validates and saves a username.
func (h *Handler) SetUsername(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*domain.User)

	var body struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	username := strings.TrimSpace(strings.ToLower(body.Username))

	if !usernameRegexp.MatchString(username) {
		jsonError(w, "Username must be 3-30 characters, lowercase alphanumeric and hyphens only, cannot start or end with a hyphen", http.StatusBadRequest)
		return
	}

	// Check if the user already has this username (no-op)
	if user.Username != nil && *user.Username == username {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "username": username})
		return
	}

	// Check availability
	available, err := h.repo.CheckUsernameAvailable(r.Context(), username)
	if err != nil {
		h.reqLog(r).Error("failed to check username availability", "err", err)
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !available {
		jsonError(w, "Username is already taken", http.StatusConflict)
		return
	}

	if err := h.repo.SetUsername(r.Context(), user.ID, username); err != nil {
		h.reqLog(r).Error("failed to set username", "err", err)
		jsonError(w, "failed to save username", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "username": username})
}

// CheckUsername handles GET /api/username/check?q=... — returns availability.
func (h *Handler) CheckUsername(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*domain.User)
	q := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("q")))

	if !usernameRegexp.MatchString(q) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"available": false,
			"reason":    "Invalid format. Use 3-30 lowercase letters, numbers, and hyphens.",
		})
		return
	}

	// If the user already owns this username, it's available for them
	if user.Username != nil && strings.EqualFold(*user.Username, q) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"available": true})
		return
	}

	available, err := h.repo.CheckUsernameAvailable(r.Context(), q)
	if err != nil {
		h.reqLog(r).Error("failed to check username", "err", err)
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	reason := ""
	if !available {
		reason = "Username is already taken"
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"available": available,
		"reason":    reason,
	})
}

// ToggleProfileVisibility handles PUT /api/cv-profiles/{id}/visibility.
func (h *Handler) ToggleProfileVisibility(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*domain.User)

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid profile id", http.StatusBadRequest)
		return
	}

	profile, err := h.repo.GetCVProfile(r.Context(), id)
	if err != nil {
		jsonError(w, "profile not found", http.StatusNotFound)
		return
	}
	if profile.UserID != user.ID {
		jsonError(w, "forbidden", http.StatusForbidden)
		return
	}

	var body struct {
		IsPublic bool `json:"is_public"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateCVProfileVisibility(r.Context(), id, body.IsPublic); err != nil {
		h.reqLog(r).Error("failed to update profile visibility", "err", err)
		jsonError(w, "failed to update visibility", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"is_public": body.IsPublic,
	})
}
