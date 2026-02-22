package main

import (
	"context"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/semmidev/atstex-lab/internal/auth"
	"github.com/semmidev/atstex-lab/internal/config"
	"github.com/semmidev/atstex-lab/internal/extractor"
	"github.com/semmidev/atstex-lab/internal/handler"
	mw "github.com/semmidev/atstex-lab/internal/middleware"
	"github.com/semmidev/atstex-lab/internal/repository"
	"github.com/semmidev/atstex-lab/web"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))

	slog.SetDefault(logger)

	cfg := config.Load()

	// Database
	repo, err := repository.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to db", "err", err)
		os.Exit(1)
	}
	defer repo.Close()

	// Auth Config
	authConfig := auth.NewConfig(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleCallbackURL, repo)

	// Parse embedded templates.
	funcMap := template.FuncMap{
		"deref": func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
		},
		"safeJSON": func(b json.RawMessage) template.JS {
			if len(b) == 0 {
				return template.JS("{}")
			}
			return template.JS(b)
		},
	}
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(web.TemplateFS, "templates/*.html")
	if err != nil {
		logger.Error("parsing templates", "err", err)
		os.Exit(1)
	}

	aiCfg := extractor.AIConfig{
		Provider: cfg.AIProvider,
		Model:    cfg.AIModel,
		APIKey:   cfg.AIAPIKey,
		BaseURL:  cfg.AIBaseURL,
	}

	h := handler.New(tmpl, logger, repo, authConfig, aiCfg)

	r := chi.NewRouter()

	// Middleware stack.
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(mw.RequestLogger(logger))
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(120 * time.Second))
	r.Use(chimw.CleanPath)
	r.Use(chimw.StripSlashes)

	// Public routes
	r.Group(func(r chi.Router) {
		r.Use(auth.OptionalUserMiddleware(repo))
		r.Get("/", h.Home)
		r.Get("/u/{username}", h.PublicProfile)
		r.Get("/u/{username}/{profileID}/pdf", h.PublicProfileDownloadPDF)
	})

	// Auth routes
	r.Get("/auth/google/login", h.GoogleLogin)
	r.Get("/auth/google/callback", h.GoogleCallback)
	r.Post("/auth/logout", h.Logout)
	r.Post("/auth/delete-account", h.DeleteAccount)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(repo))
		r.Get("/profile", h.Profile)
		r.Get("/support", h.Support)
		r.Post("/auth/sessions/{token}/delete", h.DeleteSession)
		r.Get("/input", h.Input)
		r.Get("/input/embed", h.InputEmbed)
		r.Get("/editor", h.Editor)
		r.Get("/publish", h.PublishSettings)
		r.Get("/api/templates", h.ListTemplates)
		r.Get("/api/templates/{name}", h.GetTemplate)
		r.Post("/api/templates/{name}/render", h.RenderTemplate)
		r.Post("/compile", h.Compile)
		r.Post("/api/extract-pdf", h.ExtractPDF)
		r.Post("/api/page-settings/apply", h.ApplyPageSettings)
		// CV Profile API
		r.Get("/api/cv-profiles", h.ListCVProfiles)
		r.Post("/api/cv-profiles", h.CreateCVProfile)
		r.Get("/api/cv-profiles/{id}", h.GetCVProfile)
		r.Put("/api/cv-profiles/{id}", h.SaveCVProfile)
		r.Delete("/api/cv-profiles/{id}", h.DeleteCVProfile)
		r.Put("/api/cv-profiles/{id}/visibility", h.ToggleProfileVisibility)
		// Public profile API
		r.Put("/api/username", h.SetUsername)
		r.Get("/api/username/check", h.CheckUsername)
		// Feedback
		r.Get("/feedback", h.FeedbackPage)
		r.Get("/api/feedback", h.ListMyFeedbacks)
		r.Post("/api/feedback", h.CreateFeedback)
		// CV Review
		r.Get("/cv-review", h.CVReviewPage)
		r.Post("/api/cv-review", h.CreateCVReview)
		r.Get("/api/cv-reviews", h.ListMyCVReviews)
	})

	// Admin routes (requires login + admin role)
	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(repo))
		r.Use(auth.AdminMiddleware())
		r.Get("/admin", h.AdminDashboard)
		r.Get("/api/admin/stats", h.AdminGetStats)
		r.Get("/api/admin/users", h.AdminListUsers)
		r.Get("/api/admin/feedbacks", h.AdminListFeedbacks)
		r.Post("/api/admin/feedbacks/{id}/reply", h.AdminReplyFeedback)
		r.Delete("/api/admin/feedbacks/{id}", h.AdminDeleteFeedback)
		r.Post("/api/admin/users/{id}/block", h.AdminBlockUser)
		r.Post("/api/admin/users/{id}/unblock", h.AdminUnblockUser)
		r.Delete("/api/admin/users/{id}", h.AdminDeleteUser)
	})

	// Forbidden page (accessible without login)
	r.Get("/forbidden", h.ForbiddenPage)

	// Serve static assets (JS/CSS) embedded in the binary.
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(web.StaticFS))))

	addr := envOr("PORT", ":8080")

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("starting atstex-lab", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Info("shutting down", "signal", sig.String())

	// Give in-flight requests up to 15 seconds to finish
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("forced shutdown", "err", err)
		os.Exit(1)
	}

	logger.Info("server stopped cleanly")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
