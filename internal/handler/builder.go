package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/semmidev/atstex-lab/internal/apperrors"
	"github.com/semmidev/atstex-lab/internal/auth"
	"github.com/semmidev/atstex-lab/internal/builder"
	"github.com/semmidev/atstex-lab/internal/compiler"
	"github.com/semmidev/atstex-lab/internal/cvtemplate"
	"github.com/semmidev/atstex-lab/internal/domain"
	"github.com/semmidev/atstex-lab/internal/middleware"
	"github.com/google/uuid"
)

// handleBuilderPage renders the drag-and-drop workspace UI.
func (s *Server) handleBuilderPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)
		if err := s.tmpl.ExecuteTemplate(w, "builder", map[string]interface{}{"User": user}); err != nil {
			s.reqLog(r).Error("template error", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}
}

// handleSaveCustomTemplate API 
func (s *Server) handleSaveCustomTemplate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		var req struct {
			ID     string          `json:"id,omitempty"`
			Name   string          `json:"name"`
			Config json.RawMessage `json:"config"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid json body"))
			return
		}

		if req.Name == "" || len(req.Config) == 0 {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("name and config are required"))
			return
		}

		if req.ID != "" {
			// Update existing template
			// Assuming there's a Parse method or we let uuid.Parse
			parsedID, err := uuid.Parse(req.ID)
			if err != nil {
				middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid template id"))
				return
			}
			
			existing, err := s.repo.GetCustomTemplate(r.Context(), parsedID)
			if err != nil || existing.UserID != user.ID {
				middleware.RespondError(w, r, apperrors.NewNotFound("template not found or access denied", err))
				return
			}
			
			if err := s.repo.UpdateCustomTemplate(r.Context(), parsedID, req.Config); err != nil {
				s.reqLog(r).Error("failed to update custom template", "err", err)
				middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to save template")))
				return
			}
			
			middleware.Respond(w, r, http.StatusOK, map[string]interface{}{"status": "ok", "template_id": parsedID})
			return
		}

		t := &domain.CustomTemplate{
			UserID: user.ID,
			Name:   req.Name,
			Config: req.Config,
		}

		if err := s.repo.CreateCustomTemplate(r.Context(), t); err != nil {
			s.reqLog(r).Error("failed to create custom template", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to save template")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, map[string]interface{}{"status": "ok", "template_id": t.ID})
	}
}

// handleLoadCustomTemplates API
func (s *Server) handleLoadCustomTemplates() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		templates, err := s.repo.GetCustomTemplatesByUserID(r.Context(), user.ID)
		if err != nil {
			s.reqLog(r).Error("failed to load custom templates", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to load templates")))
			return
		}

		middleware.Respond(w, r, http.StatusOK, templates)
	}
}

// handlePreviewCustomTemplate API
func (s *Server) handlePreviewCustomTemplate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := r.Context().Value(auth.UserContextKey).(*domain.User)

		var configRaw json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&configRaw); err != nil {
			middleware.RespondError(w, r, apperrors.NewInvalidInput("invalid json config"))
			return
		}

		// Fetch the user's latest CV profile to populate the preview data
		profiles, err := s.repo.GetCVProfilesByUserID(r.Context(), user.ID)
		if err != nil {
			s.reqLog(r).Error("failed to load cv profiles", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("internal error")))
			return
		}

		var cvData cvtemplate.CVData
		if len(profiles) > 0 {
			// Try to unmarshal the biodata from the latest profile
			_ = json.Unmarshal(profiles[0].Biodata, &cvData)
		} else {
			// Dummy data so preview isn't empty
			name := user.Name
			if name == "" { name = "John Doe" }
			cvData.Personal.Name = name
			cvData.Personal.Title = "Software Engineer Preview"
			cvData.Summary = "Dinamis custom template preview. Silahkan isi biodata Anda di menu Biodata untuk melihat dengan data asli."
			cvData.Experience = []cvtemplate.Experience{
				{Title: "Senior Developer", Company: "Tech Corp", Dates: "Jan 2020 - Present", Bullets: "Building awesome apps."},
			}
		}

		texSource, err := builder.Generate(configRaw, cvData)
		if err != nil {
			s.reqLog(r).Error("failed to generate custom latex", "err", err)
			middleware.RespondError(w, r, apperrors.NewInternal(errors.New("failed to strictly generate layout")))
			return
		}

		opts := compiler.Options{
			Engine:      compiler.EngineTectonic,
			Timeout:     30 * time.Second,
			PhotoBase64: cvData.Personal.Photo,
		}

		result, err := compiler.Compile(r.Context(), []byte(texSource), opts)
		if err != nil {
			var logData string
			if result != nil {
				logData = result.Log
			}
			s.reqLog(r).Error("custom template compilation failed", "err", err, "log", logData)
			http.Error(w, "PDF compilation failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("X-Latex-Elapsed", result.Elapsed.Round(time.Millisecond).String())
		w.Header().Set("Content-Disposition", "inline; filename=\"preview.pdf\"")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(result.PDF)
	}
}
