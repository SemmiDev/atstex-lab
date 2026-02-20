package main

import (
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/semmidev/atstex-lab/internal/handler"
	"github.com/semmidev/atstex-lab/web"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Parse embedded templates.
	tmpl, err := template.New("").ParseFS(web.TemplateFS, "templates/*.html")
	if err != nil {
		logger.Error("parsing templates", "err", err)
		os.Exit(1)
	}

	h := handler.New(tmpl, logger)

	r := chi.NewRouter()

	// Middleware stack.
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(90 * time.Second))
	r.Use(middleware.CleanPath)
	r.Use(middleware.StripSlashes)

	// Routes.
	r.Get("/", h.Home)
	r.Get("/input", h.Input)
	r.Get("/editor", h.Editor)
	r.Get("/api/templates", h.ListTemplates)
	r.Get("/api/templates/{name}", h.GetTemplate)
	r.Post("/compile", h.Compile)

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
