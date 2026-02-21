// Package handler wires HTTP routes to application logic.
package handler

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/semmidev/atstex-lab/internal/auth"
	"github.com/semmidev/atstex-lab/internal/compiler"
	"github.com/semmidev/atstex-lab/internal/cvtemplate"
	"github.com/semmidev/atstex-lab/internal/domain"
	"github.com/semmidev/atstex-lab/internal/extractor"
	mw "github.com/semmidev/atstex-lab/internal/middleware"
	"github.com/semmidev/atstex-lab/internal/repository"
)

// compileRequest is the JSON body expected by POST /compile.
type compileRequest struct {
	Source string `json:"source"`
	Engine string `json:"engine"`
}

// compileResponse is the JSON body returned by POST /compile.
type compileResponse struct {
	OK      bool   `json:"ok"`
	Log     string `json:"log"`
	Elapsed string `json:"elapsed"`
	Engine  string `json:"engine"`
	Error   string `json:"error,omitempty"`
}

// Handler holds shared dependencies for HTTP handlers.
type Handler struct {
	tmpl       *template.Template
	logger     *slog.Logger
	repo       repository.Repository
	authConfig *auth.Config
	aiConfig   extractor.AIConfig
}

// New constructs a Handler.
func New(tmpl *template.Template, logger *slog.Logger, r repository.Repository, ac *auth.Config, ai extractor.AIConfig) *Handler {
	return &Handler{tmpl: tmpl, logger: logger, repo: r, authConfig: ac, aiConfig: ai}
}

// reqLog returns the per-request logger from context (with request_id and trace_id),
// falling back to the handler's base logger.
func (h *Handler) reqLog(r *http.Request) *slog.Logger {
	return mw.GetLogger(r.Context())
}

// Home renders the landing page.
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
	if err := h.tmpl.ExecuteTemplate(w, "home", map[string]interface{}{"User": user}); err != nil {
		h.reqLog(r).Error("template error", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// Editor renders the main application UI.
func (h *Handler) Editor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
	if err := h.tmpl.ExecuteTemplate(w, "editor", map[string]interface{}{"User": user}); err != nil {
		h.reqLog(r).Error("template error", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// Input renders the data collection UI.
func (h *Handler) Input(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
	if err := h.tmpl.ExecuteTemplate(w, "input", map[string]interface{}{"User": user}); err != nil {
		h.reqLog(r).Error("template error", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// InputEmbed renders a stripped-down biodata form for embedding in the editor.
func (h *Handler) InputEmbed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
	if err := h.tmpl.ExecuteTemplate(w, "input_embed", map[string]interface{}{"User": user}); err != nil {
		h.reqLog(r).Error("template error", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// Profile renders the user profile and session management UI.
func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

	sessions, err := h.repo.GetSessionsByUserID(r.Context(), user.ID)
	if err != nil {
		h.reqLog(r).Error("failed to get active sessions", "err", err)
	}

	data := map[string]interface{}{
		"User":     user,
		"Sessions": sessions,
	}

	if err := h.tmpl.ExecuteTemplate(w, "profile", data); err != nil {
		h.reqLog(r).Error("template error", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// DeleteSession handles deleting a specific session (remote logout).
func (h *Handler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
	tokenToDelete := chi.URLParam(r, "token")

	if tokenToDelete == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Make sure the session actually belongs to this user before deleting it
	sess, err := h.repo.GetSession(r.Context(), tokenToDelete)
	if err != nil || sess.UserID != user.ID {
		h.reqLog(r).Warn("unauthorized session deletion attempt", "user", user.ID, "token", tokenToDelete)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := h.repo.DeleteSession(r.Context(), tokenToDelete); err != nil {
		h.reqLog(r).Error("failed to delete remote session", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

// ListTemplates returns a JSON list of available CV templates.
func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	tpls, err := cvtemplate.List()
	if err != nil {
		h.reqLog(r).Error("listing templates error", "err", err)
		jsonError(w, "failed to list templates", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tpls)
}

// GetTemplate returns the raw LaTeX source for a template.
func (h *Handler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return
	}
	content, err := cvtemplate.Get(name)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(content))
}

// renderTemplateRequest is the request body for both RenderTemplate and ApplyPageSettings.
type renderTemplateRequest struct {
	cvtemplate.CVData
	Settings *cvtemplate.PageSettings `json:"settings,omitempty"`
}

// RenderTemplate handles POST /api/templates/{name}/render.
func (h *Handler) RenderTemplate(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return
	}

	var req renderTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.reqLog(r).Error("invalid request body", "err", err)
		jsonError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	ps := cvtemplate.DefaultPageSettings()
	if req.Settings != nil {
		ps = *req.Settings
	}

	content, err := cvtemplate.Render(name, req.CVData, ps)
	if err != nil {
		h.reqLog(r).Error("template render error", "err", err)
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(content))
}

// Compile handles POST /compile — accepts JSON, returns JSON + PDF via separate endpoint.
// To keep things simple and avoid base64 bloat, the PDF is returned as raw bytes
// with Content-Type application/pdf when the compilation succeeds, and JSON on error.
func (h *Handler) Compile(w http.ResponseWriter, r *http.Request) {
	var req compileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Source) == 0 {
		jsonError(w, "source is empty", http.StatusBadRequest)
		return
	}

	engine := compiler.Engine(req.Engine)
	switch engine {
	case compiler.EnginePdfLatex, compiler.EngineXeLatex, compiler.EngineLuaLatex:
		// valid
	default:
		engine = compiler.EnginePdfLatex
	}

	opts := compiler.Options{
		Engine:  engine,
		Timeout: 60 * time.Second,
	}

	h.reqLog(r).Info("compiling", "engine", engine, "source_len", len(req.Source))

	result, err := compiler.Compile(r.Context(), []byte(req.Source), opts)
	if err != nil {
		h.reqLog(r).Warn("compilation error", "err", err)
		resp := compileResponse{
			OK:     false,
			Error:  err.Error(),
			Engine: string(engine),
		}
		if result != nil {
			resp.Log = result.Log
			resp.Elapsed = result.Elapsed.Round(time.Millisecond).String()
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(resp)
		return
	}

	h.reqLog(r).Info("compilation succeeded", "engine", engine, "elapsed", result.Elapsed, "pdf_bytes", len(result.PDF))

	// Return metadata as response headers so the browser JS can read them,
	// while the body is the raw PDF binary.
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("X-Latex-Log", truncate(result.Log, 8192))
	w.Header().Set("X-Latex-Elapsed", result.Elapsed.Round(time.Millisecond).String())
	w.Header().Set("X-Latex-Engine", string(result.Engine))
	w.Header().Set("Content-Disposition", "inline; filename=\"document.pdf\"")
	w.WriteHeader(http.StatusOK)
	w.Write(result.PDF)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(compileResponse{OK: false, Error: msg})
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
// It re-renders the template from the .tex file with the provided settings and biodata.
func (h *Handler) ApplyPageSettings(w http.ResponseWriter, r *http.Request) {
	var req pageSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.reqLog(r).Error("invalid request body", "err", err)
		jsonError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Template == "" {
		jsonError(w, "template name is required", http.StatusBadRequest)
		return
	}

	content, err := cvtemplate.Render(req.Template, req.CVData, req.Settings)
	if err != nil {
		h.reqLog(r).Error("page settings render error", "err", err)
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(content))
}

// ExtractPDF handles POST /api/extract-pdf — receives PDF text and returns structured biodata JSON.
func (h *Handler) ExtractPDF(w http.ResponseWriter, r *http.Request) {
	log := h.reqLog(r)

	if h.aiConfig.APIKey == "" && h.aiConfig.Provider != "ollama" {
		log.Error("AI API key not configured")
		jsonError(w, "AI extraction not configured", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("invalid request body", "err", err)
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Text) < 50 {
		jsonError(w, "PDF text is too short to extract meaningful data", http.StatusBadRequest)
		return
	}

	log.Info("extracting biodata from PDF text", "text_len", len(req.Text), "provider", h.aiConfig.Provider, "model", h.aiConfig.Model)

	result, totalTokens, err := extractor.ExtractBiodata(r.Context(), req.Text, h.aiConfig)
	if err != nil {
		log.Error("extraction failed", "err", err)
		jsonError(w, "AI extraction failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info("extraction succeeded", "provider", h.aiConfig.Provider, "tokens", totalTokens)

	// Track AI token usage
	u, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
	if u != nil && totalTokens > 0 {
		go func() {
			_ = h.repo.IncrementAITokensUsed(r.Context(), u.ID, totalTokens)
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ── Admin Handlers ─────────────────────────────────────────────

// AdminDashboard renders the admin page.
func (h *Handler) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	u, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
	if err := h.tmpl.ExecuteTemplate(w, "admin", map[string]interface{}{
		"User": u,
	}); err != nil {
		h.reqLog(r).Error("admin template error", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// AdminGetStats returns JSON aggregate stats for the admin dashboard.
func (h *Handler) AdminGetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.repo.AdminGetStats(r.Context())
	if err != nil {
		h.reqLog(r).Error("admin stats error", "err", err)
		jsonError(w, "failed to load stats", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// AdminListUsers returns paginated, searchable user list JSON.
func (h *Handler) AdminListUsers(w http.ResponseWriter, r *http.Request) {
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

	rows, total, err := h.repo.AdminListUsers(r.Context(), params)
	if err != nil {
		h.reqLog(r).Error("admin list users error", "err", err)
		jsonError(w, "failed to list users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users":   rows,
		"total":   total,
		"page":    page,
		"perPage": perPage,
	})
}
