package main

import (
	"context"
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
	tmpl, err := template.New("").ParseFS(web.TemplateFS, "templates/*.html")
	if err != nil {
		logger.Error("parsing templates", "err", err)
		os.Exit(1)
	}

	h := handler.New(tmpl, logger, repo, authConfig, cfg.OpenAIAPIKey)

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
		r.Post("/auth/sessions/{token}/delete", h.DeleteSession)
		r.Get("/input", h.Input)
		r.Get("/input/embed", h.InputEmbed)
		r.Get("/editor", h.Editor)
		r.Get("/api/templates", h.ListTemplates)
		r.Get("/api/templates/{name}", h.GetTemplate)
		r.Post("/api/templates/{name}/render", h.RenderTemplate)
		r.Post("/compile", h.Compile)
		r.Post("/api/extract-pdf", h.ExtractPDF)
		// CV Profile API
		r.Get("/api/cv-profiles", h.ListCVProfiles)
		r.Post("/api/cv-profiles", h.CreateCVProfile)
		r.Get("/api/cv-profiles/{id}", h.GetCVProfile)
		r.Put("/api/cv-profiles/{id}", h.SaveCVProfile)
		r.Delete("/api/cv-profiles/{id}", h.DeleteCVProfile)
	})

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
