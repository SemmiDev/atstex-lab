package handler

import (
	"net/http"

	"github.com/semmidev/atstex-lab/internal/auth"
	"github.com/semmidev/atstex-lab/internal/domain"
)

// handleAnalyticsPage renders the analytics dashboard page for the user.
func (s *Server) handleAnalyticsPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		metrics, err := s.repo.GetDashboardMetrics(r.Context(), user.ID)
		if err != nil {
			s.reqLog(r).Error("failed to get analytics dashboard metrics", "err", err)
			metrics = &domain.DashboardMetrics{} // avoid nil panic in template
		}

		data := map[string]interface{}{
			"User":    user,
			"Metrics": metrics,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := s.tmpl.ExecuteTemplate(w, "analytics", data); err != nil {
			s.reqLog(r).Error("template error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}
