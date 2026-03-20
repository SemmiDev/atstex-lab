package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/semmidev/atstex-lab/internal/aisuites"
	"github.com/semmidev/atstex-lab/internal/apperrors"
	"github.com/semmidev/atstex-lab/internal/auth"
	"github.com/semmidev/atstex-lab/internal/domain"
	"github.com/semmidev/atstex-lab/internal/middleware"
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

		count := req.Count
		if count == 0 {
			count = 10
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

		prepResult, tokensUsed, err := aisuites.GenerateInterviewQuestions(r.Context(), string(profile.Biodata), req.JobDescription, lang, count, s.aiConfig)
		if err != nil {
			s.reqLog(r).Error("Interview Prep AI generation error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("Interview generation failed: "+err.Error())))
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
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to save interview prep")))
			return
		}

		middleware.Respond(w, r, http.StatusCreated, prep)
	}
}

// handleListMyInterviewPreps returns the current user's interview prep history.
func (s *Server) handleListMyInterviewPreps() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		preps, err := s.repo.GetInterviewPrepsByUserID(r.Context(), user.ID)
		if err != nil {
			s.reqLog(r).Error("list interview preps error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to list interview preps")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, preps)
	}
}

// handleCritiqueInterviewAnswer receives a single interview question + the candidate's spoken
// (transcribed) answer and returns structured AI feedback in real time.
func (s *Server) handleCritiqueInterviewAnswer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		var req struct {
			Question string `json:"question"`
			Answer   string `json:"answer"`
			Language string `json:"language"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid request body"))
			return
		}

		req.Question = strings.TrimSpace(req.Question)
		req.Answer = strings.TrimSpace(req.Answer)

		if req.Question == "" {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("question is required"))
			return
		}
		if req.Answer == "" {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("answer is required — please record your spoken response first"))
			return
		}

		lang := req.Language
		if lang == "" {
			lang = "en"
		}

		critique, tokensUsed, err := aisuites.CritiqueInterviewAnswer(r.Context(), req.Question, req.Answer, lang, s.aiConfig)
		if err != nil {
			s.reqLog(r).Error("interview answer critique AI error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("critique generation failed: "+err.Error())))
			return
		}

		if tokensUsed > 0 {
			_ = s.repo.IncrementAITokensUsed(r.Context(), user.ID, tokensUsed)
		}

		middleware.Respond(w, r, http.StatusOK, critique)
	}
}
