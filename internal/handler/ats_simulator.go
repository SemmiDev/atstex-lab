package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/semmidev/atstex-lab/internal/aisuites"
	"github.com/semmidev/atstex-lab/internal/apperrors"
	"github.com/semmidev/atstex-lab/internal/auth"
	"github.com/semmidev/atstex-lab/internal/domain"
	"github.com/semmidev/atstex-lab/internal/middleware"
)

// handleAtsSimulatorPage renders the ATS Simulator UI.
func (s *Server) handleAtsSimulatorPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		profiles, _ := s.repo.GetCVProfilesByUserID(r.Context(), user.ID)

		if err := s.tmpl.ExecuteTemplate(w, "ats-simulator", map[string]interface{}{
			"User":     user,
			"Profiles": profiles,
		}); err != nil {
			s.reqLog(r).Error("template error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

// handleCreateAtsSimulation processes a job description against a CV profile using the ATS AI.
func (s *Server) handleCreateAtsSimulation() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		var req struct {
			ProfileID      string `json:"profileId"`
			JobDescription string `json:"jobDescription"`
			Language       string `json:"language"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid request body"))
			return
		}

		profileID, err := uuid.Parse(req.ProfileID)
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid profile ID"))
			return
		}

		if req.JobDescription == "" {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("job description is required"))
			return
		}

		lang := req.Language
		if lang == "" {
			lang = "en"
		}

		profile, err := s.repo.GetCVProfile(r.Context(), profileID)
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("profile not found")))
			return
		}
		if profile.UserID != user.ID {
			middleware.RespondError(w, r, apperrors.NewForbidden())
			return
		}
		//nolint:goconst // string 'null' is used here only to check for empty json
		if len(profile.Biodata) == 0 || string(profile.Biodata) == "null" || string(profile.Biodata) == "{}" {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("this CV profile has no biodata — please fill in your biodata first"))
			return
		}

		// Check subscription limits
		if checkErr := s.checkSubscriptionLimits(r.Context(), user.ID, "ats_simulation"); checkErr != nil {
			middleware.RespondError(w, r, apperrors.NewForbidden())
			return
		}

		simResult, tokensUsed, err := aisuites.ScoreATS(r.Context(), string(profile.Biodata), req.JobDescription, lang, s.aiConfig)
		if err != nil {
			s.reqLog(r).Error("ATS scoring error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("ATS scoring failed: "+err.Error())))
			return
		}

		if tokensUsed > 0 {
			_ = s.repo.IncrementAITokensUsed(r.Context(), user.ID, tokensUsed)
		}

		missingKeywordsJSON, _ := json.Marshal(simResult.MissingKeywords)
		if simResult.MissingKeywords == nil {
			missingKeywordsJSON = []byte("[]")
		}

		sim := &domain.AtsSimulation{
			UserID:          user.ID,
			ProfileID:       profileID,
			JobDescription:  req.JobDescription,
			Score:           simResult.Score,
			MissingKeywords: json.RawMessage(missingKeywordsJSON),
			Recommendations: simResult.Recommendations,
		}

		if err := s.repo.CreateAtsSimulation(r.Context(), sim); err != nil {
			s.reqLog(r).Error("save ATS simulation error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to save ATS simulation")))
			return
		}

		middleware.Respond(w, r, http.StatusCreated, sim)
	}
}

// handleListMyAtsSimulations returns the current user's ATS simulation history.
func (s *Server) handleListMyAtsSimulations() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		simulations, err := s.repo.GetAtsSimulationsByUserID(r.Context(), user.ID)
		if err != nil {
			s.reqLog(r).Error("list ATS simulations error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to list ATS simulations")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, simulations)
	}
}
