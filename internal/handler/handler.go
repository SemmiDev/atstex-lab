// Package handler wires HTTP routes to application logic.
package handler

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/semmidev/atstex-lab/internal/compiler"
	"github.com/semmidev/atstex-lab/internal/cvtemplate"
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
	tmpl   *template.Template
	logger *slog.Logger
}

// New constructs a Handler. The provided template must contain a "index" template.
func New(tmpl *template.Template, logger *slog.Logger) *Handler {
	return &Handler{tmpl: tmpl, logger: logger}
}

// Home renders the landing page.
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "home", nil); err != nil {
		h.logger.Error("template error", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// Editor renders the main application UI.
func (h *Handler) Editor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "editor", nil); err != nil {
		h.logger.Error("template error", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// Input renders the data collection UI.
func (h *Handler) Input(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "input", nil); err != nil {
		h.logger.Error("template error", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// ListTemplates returns a JSON list of available CV templates.
func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	tpls, err := cvtemplate.List()
	if err != nil {
		h.logger.Error("listing templates error", "err", err)
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

// RenderTemplate handles POST /api/templates/{name}/render.
func (h *Handler) RenderTemplate(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if name == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return
	}

	var data cvtemplate.CVData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	content, err := cvtemplate.Render(name, data)
	if err != nil {
		h.logger.Error("template render error", "err", err)
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

	h.logger.Info("compiling", "engine", engine, "source_len", len(req.Source))

	result, err := compiler.Compile(r.Context(), []byte(req.Source), opts)
	if err != nil {
		h.logger.Warn("compilation error", "err", err)
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

	h.logger.Info("compilation succeeded", "engine", engine, "elapsed", result.Elapsed, "pdf_bytes", len(result.PDF))

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
