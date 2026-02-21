package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/semmidev/atstex-lab/internal/auth"
	"github.com/semmidev/atstex-lab/internal/config"
	"github.com/semmidev/atstex-lab/internal/handler"
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

	h := handler.New(tmpl, logger, repo, authConfig)

	r := chi.NewRouter()

	// Middleware stack.
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(120 * time.Second))
	r.Use(middleware.CleanPath)
	r.Use(middleware.StripSlashes)

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
		r.Get("/editor", h.Editor)
		r.Get("/api/templates", h.ListTemplates)
		r.Get("/api/templates/{name}", h.GetTemplate)
		r.Post("/api/templates/{name}/render", h.RenderTemplate)
		r.Post("/compile", h.Compile)
	})

	// Serve static assets (JS/CSS) embedded in the binary.
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(web.StaticFS))))

	addr := envOr("PORT", ":8080")
	logger.Info("starting atstex-lab", "addr", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
