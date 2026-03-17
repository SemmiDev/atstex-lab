package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/semmidev/atstex-lab/internal/aisuites"
	"github.com/semmidev/atstex-lab/internal/auth"
	"github.com/semmidev/atstex-lab/internal/domain"
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
			s.respondErrMsg(w, r, "invalid request body", http.StatusBadRequest)
			return
		}

		profileID, err := uuid.Parse(req.ProfileID)
		if err != nil {
			s.respondErrMsg(w, r, "invalid profile ID", http.StatusBadRequest)
			return
		}

		if req.JobDescription == "" {
			s.respondErrMsg(w, r, "job description is required", http.StatusBadRequest)
			return
		}

		lang := req.Language
		if lang == "" {
			lang = "en"
		}

		profile, err := s.repo.GetCVProfile(r.Context(), profileID)
		if err != nil {
			s.respondErrMsg(w, r, "profile not found", http.StatusNotFound)
			return
		}
		if profile.UserID != user.ID {
			s.respondErrMsg(w, r, "forbidden", http.StatusForbidden)
			return
		}
		//nolint:goconst // string 'null' is used here only to check for empty json
		if len(profile.Biodata) == 0 || string(profile.Biodata) == "null" || string(profile.Biodata) == "{}" {
			s.respondErrMsg(w, r, "this CV profile has no biodata — please fill in your biodata first", http.StatusBadRequest)
			return
		}

		// Check subscription limits
		if checkErr := s.checkSubscriptionLimits(r.Context(), user.ID, "ats_simulation"); checkErr != nil {
			s.respondErrMsg(w, r, checkErr.Error(), http.StatusForbidden)
			return
		}

		simResult, tokensUsed, err := aisuites.ScoreATS(r.Context(), string(profile.Biodata), req.JobDescription, lang, s.aiConfig)
		if err != nil {
			s.reqLog(r).Error("ATS scoring error", "err", err)
			s.respondErrMsg(w, r, "ATS scoring failed: "+err.Error(), http.StatusInternalServerError)
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
			s.respondErrMsg(w, r, "failed to save ATS simulation", http.StatusInternalServerError)
			return
		}

		s.encode(w, r, http.StatusCreated, sim)
	}
}

// handleListMyAtsSimulations returns the current user's ATS simulation history.
func (s *Server) handleListMyAtsSimulations() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		simulations, err := s.repo.GetAtsSimulationsByUserID(r.Context(), user.ID)
		if err != nil {
			s.reqLog(r).Error("list ATS simulations error", "err", err)
			s.respondErrMsg(w, r, "failed to list ATS simulations", http.StatusInternalServerError)
			return
		}

		s.encode(w, r, http.StatusOK, simulations)
	}
}
