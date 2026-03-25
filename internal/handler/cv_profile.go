package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/semmidev/atstex-lab/internal/aisuites"
	"github.com/semmidev/atstex-lab/internal/apperrors"
	"github.com/semmidev/atstex-lab/internal/auth"
	"github.com/semmidev/atstex-lab/internal/domain"
	"github.com/semmidev/atstex-lab/internal/middleware"
	"github.com/semmidev/atstex-lab/internal/validate"
)

// ListCVProfiles returns all CV profiles for the authenticated user.
func (s *Server) handleListCVProfiles() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		profiles, err := s.repo.GetCVProfilesByUserID(r.Context(), user.ID)
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to fetch profiles")))
			return
		}
		if profiles == nil {
			profiles = []domain.CVProfile{}
		}
		middleware.Respond(w, r, http.StatusOK, profiles)
	}
}

// CreateCVProfile creates a new CV profile with a title.
func (s *Server) handleCreateCVProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		var body struct {
			Title string `json:"title" validate:"required,safe_text,max=100"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid JSON body"))
			return
		}
		if errs := validate.Struct(body); errs != nil {
			middleware.RespondError(w, r, apperrors.NewValidationError(errs))
			return
		}

		// Check subscription limits before creating a new profile
		if errLimit := s.checkSubscriptionLimits(r.Context(), user.ID, "cv_profile"); errLimit != nil {
			middleware.RespondError(w, r, apperrors.NewSubscriptionLimit(errLimit.Error()))
			return
		}

		profile, err := s.repo.CreateCVProfile(r.Context(), user.ID, body.Title)
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to create profile")))
			return
		}

		middleware.Respond(w, r, http.StatusCreated, profile)
	}
}

// GetCVProfile returns a single CV profile by ID.
func (s *Server) handleGetCVProfile() http.HandlerFunc {
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

		// Ensure user owns this profile
		if profile.UserID != user.ID {
			middleware.RespondError(w, r, apperrors.NewForbidden())
			return
		}

		middleware.Respond(w, r, http.StatusOK, profile)
	}
}

// SaveCVProfile updates the biodata JSON for a CV profile.
func (s *Server) handleSaveCVProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid profile id"))
			return
		}

		// Verify ownership
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
			Biodata json.RawMessage `json:"biodata"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid JSON body"))
			return
		}

		if err := s.repo.UpdateCVProfileBiodata(r.Context(), id, body.Biodata); err != nil {
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to save biodata")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, map[string]string{"status": "saved"})
	}
}

// UpdateCVProfileTitle renames a CV profile.
func (s *Server) handleUpdateCVProfileTitle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid profile id"))
			return
		}

		// Verify ownership
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
			Title string `json:"title" validate:"required,safe_text,max=100"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid JSON body"))
			return
		}
		if errs := validate.Struct(body); errs != nil {
			middleware.RespondError(w, r, apperrors.NewValidationError(errs))
			return
		}

		if err := s.repo.UpdateCVProfileTitle(r.Context(), id, body.Title); err != nil {
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to update title")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, map[string]string{"status": "title updated"})
	}
}

// DeleteCVProfile removes a CV profile.
func (s *Server) handleDeleteCVProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid profile id"))
			return
		}

		// Verify ownership
		profile, err := s.repo.GetCVProfile(r.Context(), id)
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("profile not found")))
			return
		}
		if profile.UserID != user.ID {
			middleware.RespondError(w, r, apperrors.NewForbidden())
			return
		}

		if err := s.repo.DeleteCVProfile(r.Context(), id); err != nil {
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to delete profile")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, map[string]string{"status": "deleted"})
	}
}

// AutoTailorCVProfile duplicates an existing profile and uses AI to rewrite parts of it based on a job description.
func (s *Server) handleAutoTailorCVProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		id, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid profile id"))
			return
		}

		var req struct {
			JobDescription string `json:"jobDescription" validate:"required,min=10,max=5000"`
			Language       string `json:"language" validate:"omitempty,oneof=en id"`
		}
		if errDecode := json.NewDecoder(r.Body).Decode(&req); errDecode != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid request body"))
			return
		}
		if errs := validate.Struct(req); errs != nil {
			middleware.RespondError(w, r, apperrors.NewValidationError(errs))
			return
		}

		lang := req.Language
		if lang == "" {
			lang = "en"
		}

		// 1. Verify ownership of the base profile
		baseProfile, err := s.repo.GetCVProfile(r.Context(), id)
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("profile not found")))
			return
		}
		if baseProfile.UserID != user.ID {
			middleware.RespondError(w, r, apperrors.NewForbidden())
			return
		}
		if len(baseProfile.Biodata) == 0 || string(baseProfile.Biodata) == "null" || string(baseProfile.Biodata) == "{}" {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("base profile has no biodata — please fill in your biodata first"))
			return
		}

		// 2. Check subscription limits (using ats_simulation limit type or cv_profile)
		// Auto-tailoring creates a profile AND uses AI. We'll check cv_profile limit.
		if errLimit := s.checkSubscriptionLimits(r.Context(), user.ID, "cv_profile"); errLimit != nil {
			middleware.RespondError(w, r, apperrors.NewSubscriptionLimit(errLimit.Error()))
			return
		}

		// 3. Call LLM to rewrite
		rewrittenJSON, tokensUsed, err := aisuites.AutoTailorCV(r.Context(), string(baseProfile.Biodata), req.JobDescription, lang, s.aiConfig)
		if err != nil {
			s.reqLog(r).Error("Auto-tailor CV rewriting failed", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(fmt.Errorf("AI rewriting failed: %w", err)))
			return
		}

		// 4. Atomically persist changes
		var newProfile *domain.CVProfile
		errTx := s.txm.WithTransaction(r.Context(), func(txCtx context.Context) error {
			newTitle := baseProfile.Title + " - Auto-Tailored"
			// Create profile
			p, err := s.repo.CreateCVProfile(txCtx, user.ID, newTitle)
			if err != nil {
				return err
			}
			newProfile = p

			// Deduct AI tokens
			if tokensUsed > 0 {
				_ = s.repo.IncrementAITokensUsed(txCtx, user.ID, tokensUsed)
			}

			// Save rewritten biodata
			return s.repo.UpdateCVProfileBiodata(txCtx, p.ID, rewrittenJSON)
		})

		if errTx != nil {
			s.reqLog(r).Error("Failed to save rewritten CV transaction", "err", errTx)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to save generated CV automatically")))
			return
		}

		// Fetch the final newly generated profile to return
		finalProfile, err := s.repo.GetCVProfile(r.Context(), newProfile.ID)
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to load generated profile")))
			return
		}

		middleware.Respond(w, r, http.StatusCreated, finalProfile)
	}
}
