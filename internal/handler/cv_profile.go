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
func (s *Server) handleListCVProfiles() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(auth.UserContextKey).(*domain.User)
		profiles, err := s.repo.GetCVProfilesByUserID(r.Context(), user.ID)
		if err != nil {
			s.respondErrMsg(w, r, "failed to fetch profiles", http.StatusInternalServerError)
			return
		}
		if profiles == nil {
			profiles = []domain.CVProfile{}
		}
		s.encode(w, r, http.StatusOK, profiles)
	}
}

// CreateCVProfile creates a new CV profile with a title.
func (s *Server) handleCreateCVProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(auth.UserContextKey).(*domain.User)

		var body struct {
			Title string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" {
			s.respondErrMsg(w, r, "title is required", http.StatusBadRequest)
			return
		}

		// Check subscription limits before creating a new profile
		if err := s.checkSubscriptionLimits(r.Context(), user.ID, "cv_profile"); err != nil {
			s.respondErrMsg(w, r, err.Error(), http.StatusForbidden)
			return
		}

		profile, err := s.repo.CreateCVProfile(r.Context(), user.ID, body.Title)
		if err != nil {
			s.respondErrMsg(w, r, "failed to create profile", http.StatusInternalServerError)
			return
		}

		s.encode(w, r, http.StatusCreated, profile)
	}
}

// GetCVProfile returns a single CV profile by ID.
func (s *Server) handleGetCVProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(auth.UserContextKey).(*domain.User)

		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			s.respondErrMsg(w, r, "invalid profile id", http.StatusBadRequest)
			return
		}

		profile, err := s.repo.GetCVProfile(r.Context(), id)
		if err != nil {
			s.respondErrMsg(w, r, "profile not found", http.StatusNotFound)
			return
		}

		// Ensure user owns this profile
		if profile.UserID != user.ID {
			s.respondErrMsg(w, r, "forbidden", http.StatusForbidden)
			return
		}

		s.encode(w, r, http.StatusOK, profile)
	}
}

// SaveCVProfile updates the biodata JSON for a CV profile.
func (s *Server) handleSaveCVProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(auth.UserContextKey).(*domain.User)

		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			s.respondErrMsg(w, r, "invalid profile id", http.StatusBadRequest)
			return
		}

		// Verify ownership
		profile, err := s.repo.GetCVProfile(r.Context(), id)
		if err != nil {
			s.respondErrMsg(w, r, "profile not found", http.StatusNotFound)
			return
		}
		if profile.UserID != user.ID {
			s.respondErrMsg(w, r, "forbidden", http.StatusForbidden)
			return
		}

		var body struct {
			Biodata json.RawMessage `json:"biodata"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			s.respondErrMsg(w, r, "invalid JSON body", http.StatusBadRequest)
			return
		}

		if err := s.repo.UpdateCVProfileBiodata(r.Context(), id, body.Biodata); err != nil {
			s.respondErrMsg(w, r, "failed to save biodata", http.StatusInternalServerError)
			return
		}

		s.encode(w, r, http.StatusOK, map[string]string{"status": "saved"})
	}
}

// UpdateCVProfileTitle renames a CV profile.
func (s *Server) handleUpdateCVProfileTitle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(auth.UserContextKey).(*domain.User)

		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			s.respondErrMsg(w, r, "invalid profile id", http.StatusBadRequest)
			return
		}

		// Verify ownership
		profile, err := s.repo.GetCVProfile(r.Context(), id)
		if err != nil {
			s.respondErrMsg(w, r, "profile not found", http.StatusNotFound)
			return
		}
		if profile.UserID != user.ID {
			s.respondErrMsg(w, r, "forbidden", http.StatusForbidden)
			return
		}

		var body struct {
			Title string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" {
			s.respondErrMsg(w, r, "valid title is required", http.StatusBadRequest)
			return
		}

		if err := s.repo.UpdateCVProfileTitle(r.Context(), id, body.Title); err != nil {
			s.respondErrMsg(w, r, "failed to update title", http.StatusInternalServerError)
			return
		}

		s.encode(w, r, http.StatusOK, map[string]string{"status": "title updated"})
	}
}

// DeleteCVProfile removes a CV profile.
func (s *Server) handleDeleteCVProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(auth.UserContextKey).(*domain.User)

		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			s.respondErrMsg(w, r, "invalid profile id", http.StatusBadRequest)
			return
		}

		// Verify ownership
		profile, err := s.repo.GetCVProfile(r.Context(), id)
		if err != nil {
			s.respondErrMsg(w, r, "profile not found", http.StatusNotFound)
			return
		}
		if profile.UserID != user.ID {
			s.respondErrMsg(w, r, "forbidden", http.StatusForbidden)
			return
		}

		if err := s.repo.DeleteCVProfile(r.Context(), id); err != nil {
			s.respondErrMsg(w, r, "failed to delete profile", http.StatusInternalServerError)
			return
		}

		s.encode(w, r, http.StatusOK, map[string]string{"status": "deleted"})
	}
}
