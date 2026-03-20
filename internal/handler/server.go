package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"html"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/semmidev/atstex-lab/internal/aisuites"
	"github.com/semmidev/atstex-lab/internal/apperrors"
	"github.com/semmidev/atstex-lab/internal/auth"
	"github.com/semmidev/atstex-lab/internal/compiler"
	"github.com/semmidev/atstex-lab/internal/cvtemplate"
	"github.com/semmidev/atstex-lab/internal/domain"
	"github.com/semmidev/atstex-lab/internal/middleware"
	"github.com/semmidev/atstex-lab/internal/repository"
	"github.com/semmidev/problem"
)

// compileRequest is the JSON body expected by POST /compile.
type compileRequest struct {
	Source      string `json:"source"`
	Engine      string `json:"engine"`
	PhotoBase64 string `json:"photo_base64"`
}

// Server holds shared dependencies for HTTP handlers.
type Server struct {
	router     chi.Router
	tmpl       *template.Template
	logger     *slog.Logger
	repo       repository.Repository
	authConfig *auth.Config
	aiConfig   aisuites.AIConfig
}

// NewServer constructs a Server and sets up its routes.
func NewServer(tmpl *template.Template, logger *slog.Logger, r repository.Repository, ac *auth.Config, ai aisuites.AIConfig) *Server {
	s := &Server{
		tmpl:       tmpl,
		logger:     logger,
		repo:       r,
		authConfig: ac,
		aiConfig:   ai,
	}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// reqLog returns the per-request logger from context (with request_id and trace_id),
// falling back to the handler's base logger.
func (s *Server) reqLog(r *http.Request) *slog.Logger {
	return middleware.GetLogger(r.Context())
}

// Home renders the landing page.
func (s *Server) handleHome() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		if err := s.tmpl.ExecuteTemplate(w, "home", map[string]interface{}{"User": user}); err != nil {
			s.reqLog(r).Error("template error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

// handleSubscriptionPage renders the subscription cards page.
func (s *Server) handleSubscriptionPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		// Get all active plans to display
		allPlans, err := s.repo.AdminListSubscriptionPlans(r.Context())
		if err != nil {
			s.reqLog(r).Error("error loading plans", "err", err)
		}

		var activePlans []domain.SubscriptionPlan
		for _, p := range allPlans {
			if p.IsActive {
				activePlans = append(activePlans, p)
			}
		}

		// Get subscription history
		historySubs, _ := s.repo.GetUserSubscriptions(r.Context(), user.ID)

		if err := s.tmpl.ExecuteTemplate(w, "subscription", map[string]interface{}{
			"User":        user,
			"ActivePlans": activePlans,
			"HistorySubs": historySubs,
			"Now":         time.Now(),
		}); err != nil {
			s.reqLog(r).Error("template error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

// Editor renders the main application UI.
func (s *Server) handleEditor() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		if err := s.tmpl.ExecuteTemplate(w, "editor", map[string]interface{}{"User": user}); err != nil {
			s.reqLog(r).Error("template error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

// Input renders the data collection UI.
func (s *Server) handleInput() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		if err := s.tmpl.ExecuteTemplate(w, "input", map[string]interface{}{"User": user}); err != nil {
			s.reqLog(r).Error("template error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

// InputEmbed renders a stripped-down biodata form for embedding in the editor.
func (s *Server) handleInputEmbed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		if err := s.tmpl.ExecuteTemplate(w, "input_embed", map[string]interface{}{"User": user}); err != nil {
			s.reqLog(r).Error("template error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

// Profile renders the user profile and session management UI.
func (s *Server) handleProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		sessions, err := s.repo.GetSessionsByUserID(r.Context(), user.ID)
		if err != nil {
			s.reqLog(r).Error("failed to get active sessions", "err", err)
		}

		data := map[string]interface{}{
			"User":     user,
			"Sessions": sessions,
		}

		if err := s.tmpl.ExecuteTemplate(w, "profile", data); err != nil {
			s.reqLog(r).Error("template error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

// DeleteSession handles deleting a specific session (remote logout).
func (s *Server) handleDeleteSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		tokenToDelete := chi.URLParam(r, "token")

		if tokenToDelete == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Make sure the session actually belongs to this user before deleting it
		sess, err := s.repo.GetSession(r.Context(), tokenToDelete)
		if err != nil || sess.UserID != user.ID {
			s.reqLog(r).Warn("unauthorized session deletion attempt", "user", user.ID, "token", tokenToDelete)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if err := s.repo.DeleteSession(r.Context(), tokenToDelete); err != nil {
			s.reqLog(r).Error("failed to delete remote session", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/profile", http.StatusSeeOther)
	}
}

// ListTemplates returns a JSON list of available CV templates.
func (s *Server) handleListTemplates() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpls, err := cvtemplate.List()
		if err != nil {
			s.reqLog(r).Error("listing templates error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to list templates")))
			return
		}
		middleware.Respond(w, r, http.StatusOK, tpls)
	}
}

// GetTemplate returns the raw LaTeX source for a template.
func (s *Server) handleGetTemplate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		if name == "" {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("name is required"))
			return
		}
		content, err := cvtemplate.Get(name)
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewNotFound("template", err))
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(content))
	}
}

// renderTemplateRequest is the request body for both RenderTemplate and ApplyPageSettings.
type renderTemplateRequest struct {
	cvtemplate.CVData
	Settings *cvtemplate.PageSettings `json:"settings,omitempty"`
}

// RenderTemplate handles POST /api/templates/{name}/render.
func (s *Server) handleRenderTemplate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		if name == "" {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("name is required"))
			return
		}

		var req renderTemplateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.reqLog(r).Error("invalid request body", "err", err)
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid request body: "+err.Error()))
			return
		}

		// Sanitize all user-supplied fields before template injection.
		compiler.SanitizeCVData(&req.CVData)

		ps := cvtemplate.DefaultPageSettings()
		if req.Settings != nil {
			ps = *req.Settings
		}

		content, err := cvtemplate.Render(name, req.CVData, ps, s.isFreeTier(r))
		if err != nil {
			s.reqLog(r).Error("template render error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(err))
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(content))
	}
}

// Compile handles POST /compile — accepts JSON, returns JSON + PDF via separate endpoint.
func (s *Server) handleCompile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req compileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid request body"))
			return
		}

		if len(req.Source) == 0 {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("source is empty"))
			return
		}

		engine := compiler.Engine(req.Engine)
		switch engine {
		case compiler.EnginePdfLatex, compiler.EngineXeLatex, compiler.EngineLuaLatex, compiler.EngineTectonic:
			// valid
		default:
			engine = compiler.EnginePdfLatex
		}

		opts := compiler.Options{
			Engine:      engine,
			Timeout:     30 * time.Second,
			PhotoBase64: req.PhotoBase64,
		}

		s.reqLog(r).Info("compiling", "engine", engine, "source_len", len(req.Source))

		result, err := compiler.Compile(r.Context(), []byte(req.Source), opts)
		if err != nil {
			s.reqLog(r).Warn("compilation error", "err", err)

			ext := map[string]any{
				"engine": string(engine),
			}
			if result != nil {
				ext["log"] = result.Log
				ext["elapsed"] = result.Elapsed.Round(time.Millisecond).String()
			}

			prob := problem.New(problem.UnprocessableEntity,
				problem.WithDetail("LaTeX compilation failed: "+err.Error()),
				problem.WithExtensions(ext),
			)
			// Return application/problem+json
			_ = prob.Write(w)
			return
		}

		s.reqLog(r).Info("compilation succeeded", "engine", engine, "elapsed", result.Elapsed, "pdf_bytes", len(result.PDF))

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("X-Latex-Log", truncate(result.Log, 8192))
		w.Header().Set("X-Latex-Elapsed", result.Elapsed.Round(time.Millisecond).String())
		w.Header().Set("X-Latex-Engine", string(result.Engine))
		w.Header().Set("Content-Disposition", "inline; filename=\"document.pdf\"")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(result.PDF)
	}
}

// truncate caps a string at n bytes to avoid huge response headers.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// pageSettingsRequest is the JSON body expected by POST /api/page-settings/apply.
type pageSettingsRequest struct {
	Template string                  `json:"template"`
	CVData   cvtemplate.CVData       `json:"cvData"`
	Settings cvtemplate.PageSettings `json:"settings"`
}

// ApplyPageSettings handles POST /api/page-settings/apply.
func (s *Server) handleApplyPageSettings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req pageSettingsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			s.reqLog(r).Error("invalid request body", "err", err)
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid request body: "+err.Error()))
			return
		}

		if req.Template == "" {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("template name is required"))
			return
		}

		// Sanitize all user-supplied fields before template injection.
		compiler.SanitizeCVData(&req.CVData)

		content, err := cvtemplate.Render(req.Template, req.CVData, req.Settings, s.isFreeTier(r))
		if err != nil {
			s.reqLog(r).Error("page settings render error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(err))
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(content))
	}
}

// ExtractPDF handles POST /api/extract-pdf — receives PDF text and returns structured biodata JSON.
func (s *Server) handleExtractPDF() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := s.reqLog(r)

		if s.aiConfig.APIKey == "" && s.aiConfig.Provider != "ollama" {
			log.Error("AI API key not configured")
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("AI extraction not configured")))
			return
		}

		// Enforce 5MB limit on request body entirely to prevent memory exhaustion
		r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

		var req struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Error("invalid request body", "err", err)
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid request body"))
			return
		}

		if len(req.Text) < 50 {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("PDF text is too short to extract meaningful data"))
			return
		}

		const maxChars = 20000
		if len(req.Text) > maxChars {
			log.Warn("PDF text exceeds maximum allowed characters", "text_len", len(req.Text), "max_chars", maxChars)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("PDF text exceeds maximum allowed characters")))
			return
		}

		log.Info("extracting biodata from PDF text", "text_len", len(req.Text), "provider", s.aiConfig.Provider, "model", s.aiConfig.Model)

		result, totalTokens, genInfo, err := aisuites.ExtractBiodata(r.Context(), req.Text, s.aiConfig)
		if err != nil {
			log.Error("extraction failed", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("AI extraction failed: "+err.Error())))
			return
		}

		// Log full AI output for easy tracking
		genInfoJSON, _ := json.Marshal(genInfo)
		log.Info("extraction succeeded",
			"provider", s.aiConfig.Provider,
			"model", s.aiConfig.Model,
			"tokens", totalTokens,
			"generation_info", string(genInfoJSON),
		)

		// Track AI token usage (use Background context since the request context
		// will be cancelled after the HTTP response is sent)
		u, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		if u != nil && totalTokens > 0 {
			uid := u.ID
			tokens := totalTokens
			//nolint:contextcheck // Background context intentionally used for async tracking
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				_ = s.repo.IncrementAITokensUsed(ctx, uid, tokens)
			}()
		}

		middleware.Respond(w, r, http.StatusOK, result)
	}
}

// ── Admin Handlers ─────────────────────────────────────────────

// AdminDashboard renders the admin page.
func (s *Server) handleAdminDashboard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		if err := s.tmpl.ExecuteTemplate(w, "admin", map[string]interface{}{
			"User": u,
		}); err != nil {
			s.reqLog(r).Error("admin template error", "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// AdminGetStats returns JSON aggregate stats for the admin dashboard.
func (s *Server) handleAdminGetStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := s.repo.AdminGetStats(r.Context())
		if err != nil {
			s.reqLog(r).Error("admin stats error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to load stats")))
			return
		}
		middleware.Respond(w, r, http.StatusOK, stats)
	}
}

// AdminListUsers returns paginated, searchable user list JSON.
func (s *Server) handleAdminListUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		page, _ := strconv.Atoi(q.Get("page"))
		if page < 1 {
			page = 1
		}
		perPage, _ := strconv.Atoi(q.Get("per_page"))
		if perPage < 1 {
			perPage = 20
		}

		params := domain.AdminListParams{
			Page:    page,
			PerPage: perPage,
			Search:  q.Get("search"),
			Sort:    q.Get("sort"),
			Order:   q.Get("order"),
		}

		rows, total, err := s.repo.AdminListUsers(r.Context(), params)
		if err != nil {
			s.reqLog(r).Error("admin list users error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to list users")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, map[string]interface{}{
			"users":   rows,
			"total":   total,
			"page":    page,
			"perPage": perPage,
		})
	}
}

// ── Feedback Handlers ──────────────────────────────────────────

// FeedbackPage renders the user feedback page.
func (s *Server) handleFeedbackPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		if err := s.tmpl.ExecuteTemplate(w, "feedback", map[string]interface{}{"User": user}); err != nil {
			s.reqLog(r).Error("template error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

// CreateFeedback handles POST /api/feedback — user submits new feedback.
func (s *Server) handleCreateFeedback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		// Limit the payload to 10KB to prevent DoS via massive JSON payloads
		r.Body = http.MaxBytesReader(w, r.Body, 10<<10)

		var req struct {
			Subject string `json:"subject"`
			Message string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid request body or payload too large"))
			return
		}

		req.Subject = strings.TrimSpace(req.Subject)
		req.Message = strings.TrimSpace(req.Message)

		if req.Subject == "" || req.Message == "" {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("subject and message are required"))
			return
		}

		if len(req.Subject) > 200 {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("subject must not exceed 200 characters"))
			return
		}

		if len(req.Message) > 2000 {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("message must not exceed 2000 characters"))
			return
		}

		// Escape the input strings so they are safe to render without execution
		safeSubject := html.EscapeString(req.Subject)
		safeMessage := html.EscapeString(req.Message)

		fb, err := s.repo.CreateFeedback(r.Context(), user.ID, safeSubject, safeMessage)
		if err != nil {
			s.reqLog(r).Error("create feedback error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to create feedback")))
			return
		}

		middleware.Respond(w, r, http.StatusCreated, fb)
	}
}

// ListMyFeedbacks handles GET /api/feedback — returns the current user's feedbacks.
func (s *Server) handleListMyFeedbacks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		feedbacks, err := s.repo.GetFeedbacksByUserID(r.Context(), user.ID)
		if err != nil {
			s.reqLog(r).Error("list feedbacks error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to list feedbacks")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, feedbacks)
	}
}

// AdminListFeedbacks handles GET /api/admin/feedbacks — paginated feedback list.
func (s *Server) handleAdminListFeedbacks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		page, _ := strconv.Atoi(q.Get("page"))
		if page < 1 {
			page = 1
		}
		perPage, _ := strconv.Atoi(q.Get("per_page"))
		if perPage < 1 {
			perPage = 20
		}

		params := domain.FeedbackListParams{
			Page:    page,
			PerPage: perPage,
			Search:  q.Get("search"),
		}

		rows, total, err := s.repo.AdminListFeedbacks(r.Context(), params)
		if err != nil {
			s.reqLog(r).Error("admin list feedbacks error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to list feedbacks")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, map[string]interface{}{
			"feedbacks": rows,
			"total":     total,
			"page":      page,
			"perPage":   perPage,
		})
	}
}

// AdminReplyFeedback handles POST /api/admin/feedbacks/{id}/reply — admin answers feedback.
func (s *Server) handleAdminReplyFeedback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		feedbackID, err := uuid.Parse(idStr)
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid feedback ID"))
			return
		}

		var req struct {
			Reply string `json:"reply"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid request body"))
			return
		}
		if req.Reply == "" {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("reply is required"))
			return
		}

		if err := s.repo.AdminReplyFeedback(r.Context(), feedbackID, req.Reply); err != nil {
			s.reqLog(r).Error("admin reply feedback error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to reply to feedback")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, map[string]interface{}{"ok": true})
	}
}

// AdminMakeUserAdmin handles POST /api/admin/users/{id}/make-admin.
func (s *Server) handleAdminMakeUserAdmin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		userID, err := uuid.Parse(idStr)
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid user id"))
			return
		}

		if err := s.repo.AdminMakeUserAdmin(r.Context(), userID); err != nil {
			s.reqLog(r).Error("admin make user admin error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to make user admin")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, map[string]string{"message": "user updated to admin"})
	}
}

// AdminRevokeUserAdmin handles POST /api/admin/users/{id}/revoke-admin.
func (s *Server) handleAdminRevokeUserAdmin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		targetUserID, err := uuid.Parse(idStr)
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid user id"))
			return
		}

		// Prevent admins from revoking their own access
		currentUser, ok := r.Context().Value(auth.UserContextKey).(*domain.User)
		if ok && currentUser.ID == targetUserID {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("cannot revoke your own admin access"))
			return
		}

		if err := s.repo.AdminRevokeUserAdmin(r.Context(), targetUserID); err != nil {
			s.reqLog(r).Error("admin revoke user admin error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to revoke user admin")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, map[string]string{"message": "admin access revoked"})
	}
}

// ForbiddenPage renders a styled 403 page.
func (s *Server) handleForbiddenPage() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		if err := s.tmpl.ExecuteTemplate(w, "forbidden", nil); err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	}
}

// AdminBlockUser blocks a user.
func (s *Server) handleAdminBlockUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid user ID"))
			return
		}
		if err := s.repo.AdminBlockUser(r.Context(), userID); err != nil {
			s.reqLog(r).Error("admin block user error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to block user")))
			return
		}
		middleware.Respond(w, r, http.StatusOK, map[string]interface{}{"ok": true})
	}
}

// AdminUnblockUser unblocks a user.
func (s *Server) handleAdminUnblockUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid user ID"))
			return
		}
		if err := s.repo.AdminUnblockUser(r.Context(), userID); err != nil {
			s.reqLog(r).Error("admin unblock user error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to unblock user")))
			return
		}
		middleware.Respond(w, r, http.StatusOK, map[string]interface{}{"ok": true})
	}
}

// AdminDeleteUser deletes a user and all their data.
func (s *Server) handleAdminDeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid user ID"))
			return
		}
		if err := s.repo.AdminDeleteUser(r.Context(), userID); err != nil {
			s.reqLog(r).Error("admin delete user error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to delete user")))
			return
		}
		middleware.Respond(w, r, http.StatusOK, map[string]interface{}{"ok": true})
	}
}

// AdminDeleteFeedback deletes a feedback entry.
func (s *Server) handleAdminDeleteFeedback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		feedbackID, err := uuid.Parse(chi.URLParam(r, "id"))
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid feedback ID"))
			return
		}
		if err := s.repo.AdminDeleteFeedback(r.Context(), feedbackID); err != nil {
			s.reqLog(r).Error("admin delete feedback error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to delete feedback")))
			return
		}
		middleware.Respond(w, r, http.StatusOK, map[string]interface{}{"ok": true})
	}
}

// isFreeTier returns true if the current user has no paid subscription.
func (s *Server) isFreeTier(r *http.Request) bool {
	user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
	if user == nil {
		return true
	}
	sub, err := s.repo.GetUserActiveSubscription(r.Context(), user.ID)
	if err != nil || sub == nil || sub.Plan == nil {
		return true
	}
	return sub.Plan.PriceIDR == 0
}

// checkSubscriptionLimits checks if a user has exceeded their AI feature limits for the active plan.
func (s *Server) checkSubscriptionLimits(ctx context.Context, userID uuid.UUID, feature string) error {
	sub, err := s.repo.GetUserActiveSubscription(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil // No active subscription means they are likely on a default gratis plan, or should be handled separately. For now, assume they fall back to gratis limits if not explicitly saved, or we can assume Free plan limit.
			// Ideally the system assigns the free plan on signup. Let's assume there is at least a Free subscription row, or if not, no limits can be checked properly.
		}
		return err
	}

	if sub.Plan == nil {
		return nil // Shouldn't happen if query is correct, but safe fallback
	}

	switch feature {
	case "cv_review":
		if sub.Plan.MaxCVReviews == -1 {
			return nil // Unlimited
		}
		count, err := s.repo.CountCVReviewsByDate(ctx, userID, sub.StartDate, sub.EndDate)
		if err != nil {
			return err
		}
		if count >= sub.Plan.MaxCVReviews {
			return errors.New("you have reached the limit for AI CV Reviews on your current plan, please upgrade to continue")
		}
	case "cover_letter":
		if sub.Plan.MaxCoverLetters == -1 {
			return nil
		}
		count, err := s.repo.CountCoverLettersByDate(ctx, userID, sub.StartDate, sub.EndDate)
		if err != nil {
			return err
		}
		if count >= sub.Plan.MaxCoverLetters {
			return errors.New("you have reached the limit for Cover Letter Generation on your current plan, please upgrade to continue")
		}
	case "ats_simulation":
		if sub.Plan.MaxATSSimulations == -1 {
			return nil
		}
		count, err := s.repo.CountAtsSimulationsByDate(ctx, userID, sub.StartDate, sub.EndDate)
		if err != nil {
			return err
		}
		if count >= sub.Plan.MaxATSSimulations {
			return errors.New("you have reached the limit for ATS Simulations on your current plan, please upgrade to continue")
		}
	case "cv_profile":
		if sub.Plan.MaxCVProfiles == -1 {
			return nil
		}
		// Count total CV profiles currently owned by the user
		profiles, err := s.repo.GetCVProfilesByUserID(ctx, userID)
		if err != nil {
			return err
		}
		if len(profiles) >= sub.Plan.MaxCVProfiles {
			return errors.New("you have reached the maximum number of CV Profiles allowed on your current plan, please upgrade to continue")
		}
	}

	return nil
}

// ── CV Review ─────────────────────────────────────────────────

// CVReviewPage renders the AI CV critique page.
func (s *Server) handleCVReviewPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		profiles, _ := s.repo.GetCVProfilesByUserID(r.Context(), user.ID)

		if err := s.tmpl.ExecuteTemplate(w, "cv-review", map[string]interface{}{
			"User":     user,
			"Profiles": profiles,
		}); err != nil {
			s.reqLog(r).Error("template error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

// CreateCVReview generates an AI critique for a selected CV profile.
func (s *Server) handleCreateCVReview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		var req struct {
			ProfileID string `json:"profileId"`
			Language  string `json:"language"`
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
		if len(profile.Biodata) == 0 || string(profile.Biodata) == "null" || string(profile.Biodata) == "{}" {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("this CV profile has no biodata — please fill in your biodata first"))
			return
		}

		// Check subscription limits
		if checkErr := s.checkSubscriptionLimits(r.Context(), user.ID, "cv_review"); checkErr != nil {
			middleware.RespondError(w, r, apperrors.NewForbidden())
			return
		}

		critiqueResult, tokensUsed, err := aisuites.CritiqueCVProfile(r.Context(), string(profile.Biodata), lang, s.aiConfig)
		if err != nil {
			s.reqLog(r).Error("AI critique error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("AI critique failed: "+err.Error())))
			return
		}

		if tokensUsed > 0 {
			_ = s.repo.IncrementAITokensUsed(r.Context(), user.ID, tokensUsed)
		}

		review := &domain.CVReview{
			UserID:          user.ID,
			ProfileID:       profileID,
			ProfileTitle:    profile.Title,
			Language:        lang,
			Score:           critiqueResult.Score,
			Strengths:       critiqueResult.Strengths,
			Improvements:    critiqueResult.Improvements,
			Recommendations: critiqueResult.Recommendations,
			TokensUsed:      tokensUsed,
		}
		if err := s.repo.CreateCVReview(r.Context(), review); err != nil {
			s.reqLog(r).Error("save cv review error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to save review")))
			return
		}

		middleware.Respond(w, r, http.StatusCreated, review)
	}
}

// ListMyCVReviews returns the current user's CV review history.
func (s *Server) handleListMyCVReviews() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		reviews, err := s.repo.GetCVReviewsByUserID(r.Context(), user.ID)
		if err != nil {
			s.reqLog(r).Error("list cv reviews error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to list reviews")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, reviews)
	}
}

// ── Cover Letter ──────────────────────────────────────────────

// CoverLetterPage renders the AI cover letter generator page.
func (s *Server) handleCoverLetterPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		profiles, _ := s.repo.GetCVProfilesByUserID(r.Context(), user.ID)

		if err := s.tmpl.ExecuteTemplate(w, "cover-letter", map[string]interface{}{
			"User":     user,
			"Profiles": profiles,
		}); err != nil {
			s.reqLog(r).Error("template error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

// GenerateCoverLetter handles the API request to generate a custom cover letter.
func (s *Server) handleGenerateCoverLetter() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		var req struct {
			ProfileID     string `json:"profileId"`
			JobDesc       string `json:"jobDesc"`
			Tone          string `json:"tone"`
			MaxParagraphs int    `json:"maxParagraphs"`
			Language      string `json:"language"`
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

		if req.JobDesc == "" {
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
		if len(profile.Biodata) == 0 || string(profile.Biodata) == "null" || string(profile.Biodata) == "{}" {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("this CV profile has no biodata — please fill in your biodata first"))
			return
		}

		// Check subscription limits
		if checkErr := s.checkSubscriptionLimits(r.Context(), user.ID, "cover_letter"); checkErr != nil {
			middleware.RespondError(w, r, apperrors.NewForbidden())
			return
		}

		coverLetter, tokensUsed, err := aisuites.GenerateCoverLetter(
			r.Context(),
			string(profile.Biodata),
			req.JobDesc,
			req.Tone,
			req.MaxParagraphs,
			lang,
			s.aiConfig,
		)
		if err != nil {
			s.reqLog(r).Error("AI cover letter generator error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("AI generation failed: "+err.Error())))
			return
		}

		if tokensUsed > 0 {
			_ = s.repo.IncrementAITokensUsed(r.Context(), user.ID, tokensUsed)
		}

		cl := &domain.CoverLetter{
			UserID:          user.ID,
			ProfileID:       profileID,
			ProfileTitle:    profile.Title,
			JobDescription:  req.JobDesc,
			CoverLetterText: coverLetter,
			Language:        lang,
			TokensUsed:      tokensUsed,
		}

		if err := s.repo.CreateCoverLetter(r.Context(), cl); err != nil {
			s.reqLog(r).Error("save cover letter error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to save cover letter")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, map[string]interface{}{
			"coverLetter": coverLetter,
			"tokensUsed":  tokensUsed,
			"id":          cl.ID,
		})
	}
}

// ListMyCoverLetters returns the current user's cover letter history.
func (s *Server) handleListMyCoverLetters() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		letters, err := s.repo.GetCoverLettersByUserID(r.Context(), user.ID)
		if err != nil {
			s.reqLog(r).Error("list cover letters error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to list cover letters")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, letters)
	}
}

// ── Job Application Tracking ──────────────────────────────────

// handleKanbanPage renders the job application tracking board.
func (s *Server) handleKanbanPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		if err := s.tmpl.ExecuteTemplate(w, "kanban", map[string]interface{}{"User": user}); err != nil {
			s.reqLog(r).Error("template error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

// handleAdminAnalytics returns time-series analytics data for charts.
func (s *Server) handleAdminAnalytics() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		analytics, err := s.repo.AdminGetAnalytics(r.Context())
		if err != nil {
			s.reqLog(r).Error("admin analytics error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to load analytics")))
			return
		}
		middleware.Respond(w, r, http.StatusOK, analytics)
	}
}

// ── Gallery (Multi-Template Preview) ──────────────────────────

// handleGalleryPage renders the template gallery page.
func (s *Server) handleGalleryPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		if err := s.tmpl.ExecuteTemplate(w, "gallery", map[string]interface{}{"User": user}); err != nil {
			s.reqLog(r).Error("template error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

// galleryCompileRequest is the JSON body for POST /api/gallery/compile.
type galleryCompileRequest struct {
	cvtemplate.CVData
	Settings *cvtemplate.PageSettings `json:"settings,omitempty"`
}

// galleryCompileResult holds the result for one template compilation.
type galleryCompileResult struct {
	Name      string `json:"name"`
	PDFBase64 string `json:"pdf_base64,omitempty"`
	Error     string `json:"error,omitempty"`
}

// handleGalleryCompile compiles the user's biodata against all templates in parallel.
func (s *Server) handleGalleryCompile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req galleryCompileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid request body: "+err.Error()))
			return
		}

		compiler.SanitizeCVData(&req.CVData)

		ps := cvtemplate.DefaultPageSettings()
		if req.Settings != nil {
			ps = *req.Settings
		}

		showWatermark := s.isFreeTier(r)

		tpls, err := cvtemplate.List()
		if err != nil {
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to list templates")))
			return
		}

		type indexedResult struct {
			idx    int
			result galleryCompileResult
		}

		ch := make(chan indexedResult, len(tpls))

		for i, tpl := range tpls {
			go func(ctx context.Context, idx int, name string) {
				res := galleryCompileResult{Name: name}

				source, renderErr := cvtemplate.Render(name, req.CVData, ps, showWatermark)
				if renderErr != nil {
					res.Error = "render: " + renderErr.Error()
					ch <- indexedResult{idx, res}
					return
				}

				opts := compiler.Options{
					Engine:  compiler.EngineTectonic,
					Timeout: 30 * time.Second,
				}

				// Extract photo for compilation if present.
				if req.CVData.Personal.Photo != "" {
					opts.PhotoBase64 = req.CVData.Personal.Photo
				}

				result, compileErr := compiler.Compile(ctx, []byte(source), opts)
				if compileErr != nil {
					errMsg := compileErr.Error()
					if result != nil && result.Log != "" {
						errMsg += "\n" + result.Log
					}
					res.Error = errMsg
					ch <- indexedResult{idx, res}
					return
				}

				res.PDFBase64 = "data:application/pdf;base64," + base64Encode(result.PDF)
				ch <- indexedResult{idx, res}
			}(r.Context(), i, tpl.Name)
		}

		results := make([]galleryCompileResult, len(tpls))
		for range tpls {
			ir := <-ch
			results[ir.idx] = ir.result
		}

		middleware.Respond(w, r, http.StatusOK, results)
	}
}

// base64Encode encodes bytes to base64 string.
func base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// handleEnhanceBullet rewrites a single set of CV bullet points using AI.
// POST /api/ai/enhance-bullet.
// Body: { "bullet": "raw text", "language": "en" }.
// Response: { "enhanced": "rewritten text", "tokensUsed": 123 }.
func (s *Server) handleEnhanceBullet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		var req struct {
			Bullet   string `json:"bullet"`
			Language string `json:"language"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid request body"))
			return
		}

		req.Bullet = strings.TrimSpace(req.Bullet)
		if req.Bullet == "" {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("bullet text is required"))
			return
		}
		if len(req.Bullet) > 4000 {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("bullet text is too long (max 4000 characters)"))
			return
		}

		lang := req.Language
		if lang == "" {
			lang = "en"
		}

		enhanced, tokensUsed, err := aisuites.EnhanceBulletPoint(r.Context(), req.Bullet, lang, s.aiConfig)
		if err != nil {
			s.reqLog(r).Error("bullet enhancement error", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("AI enhancement failed: "+err.Error())))
			return
		}

		if tokensUsed > 0 {
			_ = s.repo.IncrementAITokensUsed(r.Context(), user.ID, tokensUsed)
		}

		middleware.Respond(w, r, http.StatusOK, map[string]interface{}{
			"enhanced":   enhanced,
			"tokensUsed": tokensUsed,
		})
	}
}
