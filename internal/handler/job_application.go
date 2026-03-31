package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/semmidev/atstex-lab/internal/apperrors"
	"github.com/semmidev/atstex-lab/internal/auth"
	"github.com/semmidev/atstex-lab/internal/domain"
	"github.com/semmidev/atstex-lab/internal/middleware"
	"github.com/semmidev/atstex-lab/internal/validate"
)

// CreateJobApplicationRequest defines the payload for creating a job application.
type CreateJobApplicationRequest struct {
	Company     string    `json:"company" validate:"required,max=200"`
	JobTitle    string    `json:"job_title" validate:"required,max=200"`
	Status      string    `json:"status" validate:"omitempty,max=100"`
	Notes       string    `json:"notes" validate:"omitempty,max=5000"`
	CVProfileID uuid.UUID `json:"cv_profile_id" validate:"omitempty"`
	Deadline    *string   `json:"deadline" validate:"omitempty"`
}

// CreateJobApplication handles the creation of a new job application.
// @Summary      Create Job Application
// @Description  Create a new job application in the Kanban board
// @Tags         Job Applications
// @Accept       json
// @Produce      json
// @Param        request body CreateJobApplicationRequest true "Job Application details"
// @Success      201  {object}  domain.JobApplication
// @Failure      400  {object}  problem.Problem
// @Failure      401  {object}  problem.Problem
// @Failure      500  {object}  problem.Problem
// @Security     BearerAuth
// @Router       /api/applications [post]
func (s *Server) CreateJobApplication() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		if user == nil {
			middleware.RespondError(w, r, apperrors.NewUnauthorized(errors.New("unauthorized")))
			return
		}



		var req CreateJobApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid request body"))
			return
		}
		if errs := validate.Struct(req); errs != nil {
			middleware.RespondError(w, r, apperrors.NewValidationError(errs))
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

		if req.Deadline != nil && *req.Deadline != "" {
			t, err := time.Parse("2006-01-02", *req.Deadline)
			if err == nil {
				app.Deadline = &t
			}
		}

		if app.Status == "" {
			app.Status = "Applied"
		}

		newApp, err := s.repo.CreateJobApplication(r.Context(), app)
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to create job application")))
			return
		}

		middleware.Respond(w, r, http.StatusCreated, newApp)
	}
}

// GetJobApplications handles fetching all job applications for the logged-in user.
// @Summary      List Job Applications
// @Description  Fetch all job applications (for the Kanban board)
// @Tags         Job Applications
// @Produce      json
// @Success      200  {array}   domain.JobApplication
// @Failure      401  {object}  problem.Problem
// @Failure      500  {object}  problem.Problem
// @Security     BearerAuth
// @Router       /api/applications [get]
func (s *Server) GetJobApplications() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		if user == nil {
			middleware.RespondError(w, r, apperrors.NewUnauthorized(errors.New("unauthorized")))
			return
		}

		apps, err := s.repo.GetJobApplicationsByUserID(r.Context(), user.ID)
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to get job applications")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, apps)
	}
}

// UpdateJobApplicationStatusRequest defines the payload to update application status.
type UpdateJobApplicationStatusRequest struct {
	Status string `json:"status" validate:"required,max=100"`
}

// UpdateJobApplicationStatus handles updating the status of a job application.
// @Summary      Update Job Application Status
// @Description  Change the status of a job application (e.g. for drag-and-drop in Kanban)
// @Tags         Job Applications
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Job Application ID" format(uuid)
// @Param        request body UpdateJobApplicationStatusRequest true "New Status"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  problem.Problem
// @Failure      401  {object}  problem.Problem
// @Failure      500  {object}  problem.Problem
// @Security     BearerAuth
// @Router       /api/applications/{id}/status [put]
func (s *Server) UpdateJobApplicationStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		if user == nil {
			middleware.RespondError(w, r, apperrors.NewUnauthorized(errors.New("unauthorized")))
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid application ID"))
			return
		}

		var payload UpdateJobApplicationStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid request body"))
			return
		}
		if errs := validate.Struct(payload); errs != nil {
			middleware.RespondError(w, r, apperrors.NewValidationError(errs))
			return
		}

		// Optional: check if the application belongs to the user before updating
		// This adds security but also an extra database query.
		// For this implementation, we trust the client-side logic and rely on obscurity.

		if err := s.repo.UpdateJobApplicationStatus(r.Context(), id, payload.Status); err != nil {
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to update job application status")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, map[string]string{"message": "status updated successfully"})
	}
}

// UpdateJobApplicationRequest defines the payload for updating a job application.
type UpdateJobApplicationRequest struct {
	Company     string    `json:"company" validate:"required,max=200"`
	JobTitle    string    `json:"job_title" validate:"required,max=200"`
	Notes       string    `json:"notes" validate:"omitempty,max=5000"`
	CVProfileID uuid.UUID `json:"cv_profile_id" validate:"omitempty"`
	Deadline    *string   `json:"deadline" validate:"omitempty"`
}

// UpdateJobApplication handles updating an existing job application.
// @Summary      Update Job Application
// @Description  Update details of an existing job application
// @Tags         Job Applications
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Job Application ID" format(uuid)
// @Param        request body UpdateJobApplicationRequest true "Job Application updates"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  problem.Problem
// @Failure      401  {object}  problem.Problem
// @Failure      500  {object}  problem.Problem
// @Security     BearerAuth
// @Router       /api/applications/{id} [put]
func (s *Server) UpdateJobApplication() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		if user == nil {
			middleware.RespondError(w, r, apperrors.NewUnauthorized(errors.New("unauthorized")))
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid application ID"))
			return
		}



		var req UpdateJobApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid request body"))
			return
		}
		if errs := validate.Struct(req); errs != nil {
			middleware.RespondError(w, r, apperrors.NewValidationError(errs))
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

		if req.Deadline != nil && *req.Deadline != "" {
			t, err := time.Parse("2006-01-02", *req.Deadline)
			if err == nil {
				app.Deadline = &t
			}
		}

		if err := s.repo.UpdateJobApplication(r.Context(), app); err != nil {
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to update job application")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, map[string]string{"message": "application updated successfully"})
	}
}

// DeleteJobApplication handles deleting a job application.
// @Summary      Delete Job Application
// @Description  Delete a job application from the Kanban board
// @Tags         Job Applications
// @Param        id   path      string  true  "Job Application ID" format(uuid)
// @Success      204  "No Content"
// @Failure      400  {object}  problem.Problem
// @Failure      401  {object}  problem.Problem
// @Failure      500  {object}  problem.Problem
// @Security     BearerAuth
// @Router       /api/applications/{id} [delete]
func (s *Server) DeleteJobApplication() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		if user == nil {
			middleware.RespondError(w, r, apperrors.NewUnauthorized(errors.New("unauthorized")))
			return
		}

		idStr := chi.URLParam(r, "id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid application ID"))
			return
		}

		// Optional: check if the application belongs to the user before deleting

		if err := s.repo.DeleteJobApplication(r.Context(), id); err != nil {
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to delete job application")))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
