package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/semmidev/atstex-lab/internal/domain"
)

// AdminListSubscriptionPlans handles GET /api/admin/subscription-plans.
func (s *Server) handleAdminListSubscriptionPlans() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		plans, err := s.repo.AdminListSubscriptionPlans(r.Context())
		if err != nil {
			s.reqLog(r).Error("admin list subscription plans error", "err", err)
			s.respondErrMsg(w, r, "failed to list subscription plans", http.StatusInternalServerError)
			return
		}
		s.encode(w, r, http.StatusOK, plans)
	}
}

// AdminCreateSubscriptionPlan handles POST /api/admin/subscription-plans.
func (s *Server) handleAdminCreateSubscriptionPlan() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req domain.SubscriptionPlan
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.respondErrMsg(w, r, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			s.respondErrMsg(w, r, "name is required", http.StatusBadRequest)
			return
		}

		if err := s.repo.AdminCreateSubscriptionPlan(r.Context(), &req); err != nil {
			s.reqLog(r).Error("admin create subscription plan error", "err", err)
			s.respondErrMsg(w, r, "failed to create subscription plan", http.StatusInternalServerError)
			return
		}

		s.encode(w, r, http.StatusCreated, req)
	}
}

// AdminUpdateSubscriptionPlan handles PUT /api/admin/subscription-plans/{id}.
func (s *Server) handleAdminUpdateSubscriptionPlan() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			s.respondErrMsg(w, r, "invalid plan ID", http.StatusBadRequest)
			return
		}

		var req domain.SubscriptionPlan
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.respondErrMsg(w, r, "invalid request body", http.StatusBadRequest)
			return
		}
		req.ID = planID

		if req.Name == "" {
			s.respondErrMsg(w, r, "name is required", http.StatusBadRequest)
			return
		}

		if err := s.repo.AdminUpdateSubscriptionPlan(r.Context(), &req); err != nil {
			s.reqLog(r).Error("admin update subscription plan error", "err", err)
			s.respondErrMsg(w, r, "failed to update subscription plan", http.StatusInternalServerError)
			return
		}

		s.encode(w, r, http.StatusOK, req)
	}
}

// AdminToggleSubscriptionPlan handles POST /api/admin/subscription-plans/{id}/toggle.
func (s *Server) handleAdminToggleSubscriptionPlan() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			s.respondErrMsg(w, r, "invalid plan ID", http.StatusBadRequest)
			return
		}

		var req struct {
			IsActive bool `json:"isActive"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.respondErrMsg(w, r, "invalid request body", http.StatusBadRequest)
			return
		}

		if err := s.repo.AdminToggleSubscriptionPlan(r.Context(), planID, req.IsActive); err != nil {
			s.reqLog(r).Error("admin toggle subscription plan error", "err", err)
			s.respondErrMsg(w, r, "failed to toggle subscription plan", http.StatusInternalServerError)
			return
		}

		s.encode(w, r, http.StatusOK, map[string]interface{}{"ok": true})
	}
}

// AdminDeleteSubscriptionPlan handles DELETE /api/admin/subscription-plans/{id}.
func (s *Server) handleAdminDeleteSubscriptionPlan() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			s.respondErrMsg(w, r, "invalid plan ID", http.StatusBadRequest)
			return
		}

		if err := s.repo.AdminDeleteSubscriptionPlan(r.Context(), planID); err != nil {
			s.reqLog(r).Error("admin delete subscription plan error", "err", err)
			s.respondErrMsg(w, r, "failed to delete subscription plan", http.StatusInternalServerError)
			return
		}

		s.encode(w, r, http.StatusOK, map[string]interface{}{"ok": true})
	}
}

// AdminAssignSubscription handles POST /api/admin/users/{id}/subscribe.
func (s *Server) handleAdminAssignSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			s.respondErrMsg(w, r, "invalid user ID", http.StatusBadRequest)
			return
		}

		var req struct {
			PlanID string `json:"planId"`
			Months string `json:"months"` // Can be passed as string by HTML forms
		}
		if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
			s.respondErrMsg(w, r, "invalid request body", http.StatusBadRequest)
			return
		}

		planID, err := uuid.Parse(req.PlanID)
		if err != nil {
			s.respondErrMsg(w, r, "invalid plan ID", http.StatusBadRequest)
			return
		}

		months, err := strconv.Atoi(req.Months)
		if err != nil || months < 1 {
			months = 1
		}

		if err := s.repo.AdminAssignSubscription(r.Context(), userID, planID, months); err != nil {
			s.reqLog(r).Error("admin assign subscription error", "err", err)
			s.respondErrMsg(w, r, "failed to assign subscription", http.StatusInternalServerError)
			return
		}

		s.encode(w, r, http.StatusOK, map[string]interface{}{"ok": true})
	}
}

// AdminGetUserSubscription handles GET /api/admin/users/{id}/subscription.
func (s *Server) handleAdminGetUserSubscription() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			s.respondErrMsg(w, r, "invalid user ID", http.StatusBadRequest)
			return
		}

		sub, err := s.repo.GetUserActiveSubscription(r.Context(), userID)
		if err != nil {
			// If not found, just return empty data with OK so frontend can parse cleanly
			s.encode(w, r, http.StatusOK, map[string]interface{}{"data": nil})
			return
		}

		s.encode(w, r, http.StatusOK, map[string]interface{}{"data": sub})
	}
}
