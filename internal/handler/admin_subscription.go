package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/semmidev/atstex-lab/internal/apperrors"
	"github.com/semmidev/atstex-lab/internal/domain"
	"github.com/semmidev/atstex-lab/internal/middleware"
)

// AdminListSubscriptionPlans handles GET /api/admin/subscription-plans.
// @Summary      List Subscription Plans
// @Description  Get a list of all subscription plans (Admin only)
// @Tags         Admin Subscriptions
// @Produce      json
// @Success      200  {array}   domain.SubscriptionPlan
// @Failure      500  {object}  problem.Problem
// @Security     BearerAuth
// @Router       /api/admin/subscription-plans [get]
func (s *Server) handleAdminListSubscriptionPlans() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		plans, err := s.repo.AdminListSubscriptionPlans(r.Context())
		if err != nil {
			s.reqLog(r).Error("admin list subscription plans error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to list subscription plans")))
			return
		}
		middleware.Respond(w, r, http.StatusOK, plans)
	}
}

// AdminCreateSubscriptionPlan handles POST /api/admin/subscription-plans.
// @Summary      Create Subscription Plan
// @Description  Create a new subscription plan (Admin only)
// @Tags         Admin Subscriptions
// @Accept       json
// @Produce      json
// @Param        request body domain.SubscriptionPlan true "Plan data"
// @Success      201  {object}  domain.SubscriptionPlan
// @Failure      400  {object}  problem.Problem
// @Failure      500  {object}  problem.Problem
// @Security     BearerAuth
// @Router       /api/admin/subscription-plans [post]
func (s *Server) handleAdminCreateSubscriptionPlan() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req domain.SubscriptionPlan
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid request body"))
			return
		}

		if req.Name == "" {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("name is required"))
			return
		}

		if err := s.repo.AdminCreateSubscriptionPlan(r.Context(), &req); err != nil {
			s.reqLog(r).Error("admin create subscription plan error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to create subscription plan")))
			return
		}

		middleware.Respond(w, r, http.StatusCreated, req)
	}
}

// AdminUpdateSubscriptionPlan handles PUT /api/admin/subscription-plans/{id}.
// @Summary      Update Subscription Plan
// @Description  Update an existing subscription plan (Admin only)
// @Tags         Admin Subscriptions
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Plan ID" format(uuid)
// @Param        request body domain.SubscriptionPlan true "Plan data"
// @Success      200  {object}  domain.SubscriptionPlan
// @Failure      400  {object}  problem.Problem
// @Failure      500  {object}  problem.Problem
// @Security     BearerAuth
// @Router       /api/admin/subscription-plans/{id} [put]
func (s *Server) handleAdminUpdateSubscriptionPlan() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid plan ID"))
			return
		}

		var req domain.SubscriptionPlan
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid request body"))
			return
		}
		req.ID = planID

		if req.Name == "" {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("name is required"))
			return
		}

		if err := s.repo.AdminUpdateSubscriptionPlan(r.Context(), &req); err != nil {
			s.reqLog(r).Error("admin update subscription plan error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to update subscription plan")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, req)
	}
}

// AdminToggleSubscriptionPlanRequest defines the payload for toggling a plan.
type AdminToggleSubscriptionPlanRequest struct {
	IsActive bool `json:"isActive"`
}

// AdminToggleSubscriptionPlan handles POST /api/admin/subscription-plans/{id}/toggle.
// @Summary      Toggle Subscription Plan Active State
// @Description  Activate or deactivate a subscription plan
// @Tags         Admin Subscriptions
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Plan ID" format(uuid)
// @Param        request body AdminToggleSubscriptionPlanRequest true "Toggle status"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  problem.Problem
// @Failure      500  {object}  problem.Problem
// @Security     BearerAuth
// @Router       /api/admin/subscription-plans/{id}/toggle [post]
func (s *Server) handleAdminToggleSubscriptionPlan() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid plan ID"))
			return
		}

		var req AdminToggleSubscriptionPlanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid request body"))
			return
		}

		if err := s.repo.AdminToggleSubscriptionPlan(r.Context(), planID, req.IsActive); err != nil {
			s.reqLog(r).Error("admin toggle subscription plan error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to toggle subscription plan")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, map[string]interface{}{"ok": true})
	}
}

// AdminDeleteSubscriptionPlan handles DELETE /api/admin/subscription-plans/{id}.
// @Summary      Delete Subscription Plan
// @Description  Permanently delete a subscription plan
// @Tags         Admin Subscriptions
// @Produce      json
// @Param        id   path      string  true  "Plan ID" format(uuid)
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  problem.Problem
// @Failure      500  {object}  problem.Problem
// @Security     BearerAuth
// @Router       /api/admin/subscription-plans/{id} [delete]
func (s *Server) handleAdminDeleteSubscriptionPlan() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid plan ID"))
			return
		}

		if err := s.repo.AdminDeleteSubscriptionPlan(r.Context(), planID); err != nil {
			s.reqLog(r).Error("admin delete subscription plan error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to delete subscription plan")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, map[string]interface{}{"ok": true})
	}
}

// AdminAssignSubscriptionRequest defines the API payload to assign a plan.
type AdminAssignSubscriptionRequest struct {
	PlanID string `json:"planId"`
	Months string `json:"months"` // Can be passed as string by HTML forms
}

// AdminAssignSubscription handles POST /api/admin/users/{id}/subscribe.
// @Summary      Assign Subscription to User
// @Description  Grant a specific subscription plan to a user
// @Tags         Admin Subscriptions
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "User ID" format(uuid)
// @Param        request body AdminAssignSubscriptionRequest true "Subscription Details"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  problem.Problem
// @Failure      500  {object}  problem.Problem
// @Security     BearerAuth
// @Router       /api/admin/users/{id}/subscribe [post]
func (s *Server) handleAdminAssignSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid user ID"))
			return
		}

		var req AdminAssignSubscriptionRequest
		if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid request body"))
			return
		}

		planID, err := uuid.Parse(req.PlanID)
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid plan ID"))
			return
		}

		months, err := strconv.Atoi(req.Months)
		if err != nil || months < 1 {
			months = 1
		}

		if err := s.repo.AdminAssignSubscription(r.Context(), userID, planID, months); err != nil {
			s.reqLog(r).Error("admin assign subscription error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to assign subscription")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, map[string]interface{}{"ok": true})
	}
}

// AdminGetUserSubscription handles GET /api/admin/users/{id}/subscription.
// @Summary      Get User's Subscription
// @Description  Get details of a user's active subscription
// @Tags         Admin Subscriptions
// @Produce      json
// @Param        id   path      string  true  "User ID" format(uuid)
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  problem.Problem
// @Security     BearerAuth
// @Router       /api/admin/users/{id}/subscription [get]
func (s *Server) handleAdminGetUserSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid user ID"))
			return
		}

		sub, err := s.repo.GetUserActiveSubscription(r.Context(), userID)
		if err != nil {
			// If not found, just return empty data with OK so frontend can parse cleanly
			middleware.Respond(w, r, http.StatusOK, map[string]interface{}{"data": nil})
			return
		}

		middleware.Respond(w, r, http.StatusOK, map[string]interface{}{"data": sub})
	}
}
