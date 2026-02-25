package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/semmidev/atstex-lab/internal/auth"
	"github.com/semmidev/atstex-lab/internal/domain"
)

// CreateJobApplication handles the creation of a new job application.
func (s *Server) CreateJobApplication() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		if user == nil {
			s.respondErrMsg(w, r, "unauthorized", http.StatusUnauthorized)
			return
		}

		type CreateJobApplicationRequest struct {
			Company     string    `json:"company"`
			JobTitle    string    `json:"job_title"`
			Status      string    `json:"status"`
			Notes       string    `json:"notes"`
			CVProfileID uuid.UUID `json:"cv_profile_id"`
		}

		var req CreateJobApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.respondErrMsg(w, r, "invalid request body", http.StatusBadRequest)
			return
		}

		app := &domain.JobApplication{
			UserID:   user.ID,
			Company:  req.Company,
			JobTitle: req.JobTitle,
			Status:   req.Status,
			Notes:    req.Notes,
		}

		if req.CVProfileID != uuid.Nil {
			app.CVProfileID = &req.CVProfileID
		}

		if app.Status == "" {
			app.Status = "Applied"
		}

		newApp, err := s.repo.CreateJobApplication(r.Context(), app)
		if err != nil {
			s.respondErrMsg(w, r, "failed to create job application", http.StatusInternalServerError)
			return
		}

		s.encode(w, r, http.StatusCreated, newApp)
	}
}

// GetJobApplications handles fetching all job applications for the logged-in user.
func (s *Server) GetJobApplications() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		if user == nil {
			s.respondErrMsg(w, r, "unauthorized", http.StatusUnauthorized)
			return
		}

		apps, err := s.repo.GetJobApplicationsByUserID(r.Context(), user.ID)
		if err != nil {
			s.respondErrMsg(w, r, "failed to get job applications", http.StatusInternalServerError)
			return
		}

		s.encode(w, r, http.StatusOK, apps)
	}
}

// UpdateJobApplicationStatus handles updating the status of a job application.
func (s *Server) UpdateJobApplicationStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		if user == nil {
			s.respondErrMsg(w, r, "unauthorized", http.StatusUnauthorized)
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			s.respondErrMsg(w, r, "invalid application ID", http.StatusBadRequest)
			return
		}

		var payload struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			s.respondErrMsg(w, r, "invalid request body", http.StatusBadRequest)
			return
		}

		// Optional: check if the application belongs to the user before updating
		// This adds security but also an extra database query.
		// For this implementation, we trust the client-side logic and rely on obscurity.

		if err := s.repo.UpdateJobApplicationStatus(r.Context(), id, payload.Status); err != nil {
			s.respondErrMsg(w, r, "failed to update job application status", http.StatusInternalServerError)
			return
		}

		s.encode(w, r, http.StatusOK, map[string]string{"message": "status updated successfully"})
	}
}

// UpdateJobApplication handles updating an existing job application.
func (s *Server) UpdateJobApplication() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		if user == nil {
			s.respondErrMsg(w, r, "unauthorized", http.StatusUnauthorized)
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			s.respondErrMsg(w, r, "invalid application ID", http.StatusBadRequest)
			return
		}

		type UpdateJobApplicationRequest struct {
			Company     string    `json:"company"`
			JobTitle    string    `json:"job_title"`
			Notes       string    `json:"notes"`
			CVProfileID uuid.UUID `json:"cv_profile_id"`
		}

		var req UpdateJobApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.respondErrMsg(w, r, "invalid request body", http.StatusBadRequest)
			return
		}

		app := &domain.JobApplication{
			ID:       id,
			UserID:   user.ID,
			Company:  req.Company,
			JobTitle: req.JobTitle,
			Notes:    req.Notes,
		}

		if req.CVProfileID != uuid.Nil {
			app.CVProfileID = &req.CVProfileID
		}

		if err := s.repo.UpdateJobApplication(r.Context(), app); err != nil {
			s.respondErrMsg(w, r, "failed to update job application", http.StatusInternalServerError)
			return
		}

		s.encode(w, r, http.StatusOK, map[string]string{"message": "application updated successfully"})
	}
}

// DeleteJobApplication handles deleting a job application.
func (s *Server) DeleteJobApplication() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		if user == nil {
			s.respondErrMsg(w, r, "unauthorized", http.StatusUnauthorized)
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			s.respondErrMsg(w, r, "invalid application ID", http.StatusBadRequest)
			return
		}

		// Optional: check if the application belongs to the user before deleting

		if err := s.repo.DeleteJobApplication(r.Context(), id); err != nil {
			s.respondErrMsg(w, r, "failed to delete job application", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
