package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/semmidev/atstex-lab/internal/auth"
	"github.com/semmidev/atstex-lab/internal/domain"
)

// ListCVProfiles returns all CV profiles for the authenticated user.
func (h *Handler) ListCVProfiles(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*domain.User)
	profiles, err := h.repo.GetCVProfilesByUserID(r.Context(), user.ID)
	if err != nil {
		jsonError(w, "failed to fetch profiles", http.StatusInternalServerError)
		return
	}
	if profiles == nil {
		profiles = []domain.CVProfile{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profiles)
}

// CreateCVProfile creates a new CV profile with a title.
func (h *Handler) CreateCVProfile(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*domain.User)

	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" {
		jsonError(w, "title is required", http.StatusBadRequest)
		return
	}

	profile, err := h.repo.CreateCVProfile(r.Context(), user.ID, body.Title)
	if err != nil {
		jsonError(w, "failed to create profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(profile)
}

// GetCVProfile returns a single CV profile by ID.
func (h *Handler) GetCVProfile(w http.ResponseWriter, r *http.Request) {
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

	// Ensure user owns this profile
	if profile.UserID != user.ID {
		jsonError(w, "forbidden", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// SaveCVProfile updates the biodata JSON for a CV profile.
func (h *Handler) SaveCVProfile(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*domain.User)

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid profile id", http.StatusBadRequest)
		return
	}

	// Verify ownership
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
		Biodata json.RawMessage `json:"biodata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateCVProfileBiodata(r.Context(), id, body.Biodata); err != nil {
		jsonError(w, "failed to save biodata", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
}

// DeleteCVProfile removes a CV profile.
func (h *Handler) DeleteCVProfile(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(auth.UserContextKey).(*domain.User)

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		jsonError(w, "invalid profile id", http.StatusBadRequest)
		return
	}

	// Verify ownership
	profile, err := h.repo.GetCVProfile(r.Context(), id)
	if err != nil {
		jsonError(w, "profile not found", http.StatusNotFound)
		return
	}
	if profile.UserID != user.ID {
		jsonError(w, "forbidden", http.StatusForbidden)
		return
	}

	if err := h.repo.DeleteCVProfile(r.Context(), id); err != nil {
		jsonError(w, "failed to delete profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}
