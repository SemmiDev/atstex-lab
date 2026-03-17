package aisuites

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/mistral"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

// newLLM creates an LLM client based on the provider configuration.
//
//nolint:staticcheck // SA1019: llms.LLM is deprecated but still required for older langchaingo versions
func newLLM(ctx context.Context, cfg AIConfig) (llms.LLM, error) {
	provider := strings.ToLower(cfg.Provider)

	switch provider {
	case "openai", "":
		opts := []openai.Option{
			openai.WithToken(cfg.APIKey),
		}
		if cfg.Model != "" {
			opts = append(opts, openai.WithModel(cfg.Model))
		}
		if cfg.BaseURL != "" {
			opts = append(opts, openai.WithBaseURL(cfg.BaseURL))
		}
		return openai.New(opts...)

	case "anthropic":
		opts := []anthropic.Option{
			anthropic.WithToken(cfg.APIKey),
		}
		if cfg.Model != "" {
			opts = append(opts, anthropic.WithModel(cfg.Model))
		}
		return anthropic.New(opts...)

	case "googleai", "gemini", "google":
		opts := []googleai.Option{
			googleai.WithAPIKey(cfg.APIKey),
		}
		if cfg.Model != "" {
			opts = append(opts, googleai.WithDefaultModel(cfg.Model))
		}
		return googleai.New(ctx, opts...)

	case "ollama":
		opts := []ollama.Option{}
		if cfg.Model != "" {
			opts = append(opts, ollama.WithModel(cfg.Model))
		}
		if cfg.BaseURL != "" {
			opts = append(opts, ollama.WithServerURL(cfg.BaseURL))
		}
		return ollama.New(opts...)

	case "mistral":
		opts := []mistral.Option{
			mistral.WithAPIKey(cfg.APIKey),
		}
		if cfg.Model != "" {
			opts = append(opts, mistral.WithModel(cfg.Model))
		}
		if cfg.BaseURL != "" {
			opts = append(opts, mistral.WithEndpoint(cfg.BaseURL))
		}
		return mistral.New(opts...)

	default:
		return nil, fmt.Errorf("unsupported AI provider: %q (supported: openai, anthropic, googleai, mistral, ollama)", provider)
	}
}
