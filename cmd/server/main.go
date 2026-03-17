package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/semmidev/atstex-lab/internal/auth"
	"github.com/semmidev/atstex-lab/internal/aisuites"
	"github.com/semmidev/atstex-lab/internal/config"
	"github.com/semmidev/atstex-lab/internal/handler"
	"github.com/semmidev/atstex-lab/internal/repository"
	"github.com/semmidev/atstex-lab/web"
)

func main() {
	ctx := context.Background()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))
	slog.SetDefault(logger)

	cfg := config.Load()

	if err := run(ctx, cfg, logger); err != nil {
		logger.Error("server shutdown with error", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg *config.AppConfig, logger *slog.Logger) error {
	// Database
	repo, err := repository.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	defer repo.Close()

	// Auth
	authConfig := auth.NewConfig(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleCallbackURL, repo)

	// Templates
	tmpl, err := buildTemplates()
	if err != nil {
		return fmt.Errorf("build templates: %w", err)
	}

	// AI
	aiCfg := aisuites.AIConfig{
		Provider: cfg.AIProvider,
		Model:    cfg.AIModel,
		APIKey:   cfg.AIAPIKey,
		BaseURL:  cfg.AIBaseURL,
	}

	srv := handler.NewServer(tmpl, logger, repo, authConfig, aiCfg)

	httpSrv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      srv,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errc := make(chan error, 1)

	go func() {
		logger.Info("starting atstex-lab", "addr", cfg.Port)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errc <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errc:
		return err
	case sig := <-quit:
		logger.Info("shutting down", "signal", sig.String())
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	logger.Info("server stopped cleanly")
	return nil
}

func buildTemplates() (*template.Template, error) {
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
			//nolint:gosec // JSON payload is already sanitized
			return template.JS(b)
		},
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(web.TemplateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	return tmpl, nil
}
