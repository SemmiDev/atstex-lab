package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/semmidev/atstex-lab/internal/auth"
	"github.com/semmidev/atstex-lab/internal/domain"
	"github.com/semmidev/atstex-lab/internal/extractor"
)

// handleInterviewPrepPage renders the AI Interview Prep UI.
func (s *Server) handleInterviewPrepPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		profiles, _ := s.repo.GetCVProfilesByUserID(r.Context(), user.ID)

		if err := s.tmpl.ExecuteTemplate(w, "interview-prep", map[string]interface{}{
			"User":     user,
			"Profiles": profiles,
		}); err != nil {
			s.reqLog(r).Error("template error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

// handleCreateInterviewPrep processes a job description against a CV profile to generate questions.
func (s *Server) handleCreateInterviewPrep() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		var req struct {
			ProfileID      string `json:"profileId"`
			JobDescription string `json:"jobDescription"`
			Language       string `json:"language"`
			Count          int    `json:"count"`
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
		
		count := req.Count
		if count == 0 {
			count = 10
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

		prepResult, tokensUsed, err := extractor.GenerateInterviewQuestions(r.Context(), string(profile.Biodata), req.JobDescription, lang, count, s.aiConfig)
		if err != nil {
			s.reqLog(r).Error("Interview Prep AI generation error", "err", err)
			s.respondErrMsg(w, r, "Interview generation failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if tokensUsed > 0 {
			_ = s.repo.IncrementAITokensUsed(r.Context(), user.ID, tokensUsed)
		}

		questionsJSON, _ := json.Marshal(prepResult.Categories)
		if prepResult.Categories == nil {
			questionsJSON = []byte("[]")
		}

		prep := &domain.InterviewPrep{
			UserID:         user.ID,
			ProfileID:      profileID,
			JobDescription: req.JobDescription,
			Language:       lang,
			Questions:      json.RawMessage(questionsJSON),
			TokensUsed:     tokensUsed,
		}

		if err := s.repo.CreateInterviewPrep(r.Context(), prep); err != nil {
			s.reqLog(r).Error("save interview prep error", "err", err)
			s.respondErrMsg(w, r, "failed to save interview prep", http.StatusInternalServerError)
			return
		}

		s.encode(w, r, http.StatusCreated, prep)
	}
}

// handleListMyInterviewPreps returns the current user's interview prep history.
func (s *Server) handleListMyInterviewPreps() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		preps, err := s.repo.GetInterviewPrepsByUserID(r.Context(), user.ID)
		if err != nil {
			s.reqLog(r).Error("list interview preps error", "err", err)
			s.respondErrMsg(w, r, "failed to list interview preps", http.StatusInternalServerError)
			return
		}

		s.encode(w, r, http.StatusOK, preps)
	}
}
