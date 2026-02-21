// Package extractor uses an LLM to extract structured biodata from raw text.
package extractor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

// AIConfig holds the provider-agnostic AI settings.
type AIConfig struct {
	Provider string // "openai", "anthropic", "ollama"
	Model    string // e.g. "gpt-4o-mini", "claude-3-haiku-20240307", "llama3"
	APIKey   string
	BaseURL  string // optional: for OpenAI-compatible endpoints (Groq, Together, etc.)
}

var systemPrompt = "You are a resume parser. Given raw text extracted from a PDF resume/CV,\n" +
	"extract the information into the following JSON structure. Return ONLY valid JSON, no markdown fences, no explanation.\n\n" +
	"{\n" +
	"  \"personal\": {\n" +
	"    \"name\": \"\",\n" +
	"    \"title\": \"\",\n" +
	"    \"email\": \"\",\n" +
	"    \"phone\": \"\",\n" +
	"    \"location\": \"\",\n" +
	"    \"linkedin\": { \"display\": \"\", \"url\": \"\" },\n" +
	"    \"github\": { \"display\": \"\", \"url\": \"\" },\n" +
	"    \"website\": { \"display\": \"\", \"url\": \"\" }\n" +
	"  },\n" +
	"  \"summary\": \"\",\n" +
	"  \"experience\": [\n" +
	"    {\n" +
	"      \"company\": \"\",\n" +
	"      \"title\": \"\",\n" +
	"      \"location\": \"\",\n" +
	"      \"dates\": \"\",\n" +
	"      \"bullets\": \"line1\\nline2\\nline3\"\n" +
	"    }\n" +
	"  ],\n" +
	"  \"education\": [\n" +
	"    {\n" +
	"      \"institution\": \"\",\n" +
	"      \"degree\": \"\",\n" +
	"      \"location\": \"\",\n" +
	"      \"dates\": \"\",\n" +
	"      \"gpa\": \"\",\n" +
	"      \"activities\": \"\"\n" +
	"    }\n" +
	"  ],\n" +
	"  \"projects\": [\n" +
	"    {\n" +
	"      \"name\": \"\",\n" +
	"      \"role\": \"\",\n" +
	"      \"link\": \"\",\n" +
	"      \"bullets\": \"line1\\nline2\"\n" +
	"    }\n" +
	"  ],\n" +
	"  \"skills\": {\n" +
	"    \"languages\": \"\",\n" +
	"    \"frameworks\": \"\",\n" +
	"    \"tools\": \"\",\n" +
	"    \"other\": \"\"\n" +
	"  },\n" +
	"  \"certifications\": [\n" +
	"    { \"name\": \"\", \"issuer\": \"\" }\n" +
	"  ],\n" +
	"  \"volunteer\": [\n" +
	"    {\n" +
	"      \"organization\": \"\",\n" +
	"      \"role\": \"\",\n" +
	"      \"location\": \"\",\n" +
	"      \"dates\": \"\",\n" +
	"      \"bullets\": \"line1\\nline2\"\n" +
	"    }\n" +
	"  ],\n" +
	"  \"awards\": [\n" +
	"    {\n" +
	"      \"title\": \"\",\n" +
	"      \"issuer\": \"\",\n" +
	"      \"date\": \"\",\n" +
	"      \"description\": \"\"\n" +
	"    }\n" +
	"  ],\n" +
	"  \"talks\": [\n" +
	"    {\n" +
	"      \"title\": \"\",\n" +
	"      \"event\": \"\",\n" +
	"      \"location\": \"\",\n" +
	"      \"date\": \"\",\n" +
	"      \"description\": \"\"\n" +
	"    }\n" +
	"  ]\n" +
	"}\n\n" +
	"Rules:\n" +
	"- Fill in as many fields as possible from the provided text.\n" +
	"- If a section has no data, use an empty array [] or empty string \"\".\n" +
	"- For bullet points, put each bullet on a new line separated by \\n, WITHOUT leading dashes or bullet characters.\n" +
	"- For dates, use the format found in the resume (e.g. \"Jan 2020 - Present\").\n" +
	"- For skills, group them into languages, frameworks, tools, and other as best as you can.\n" +
	"- Return ONLY the JSON object, nothing else."

// newLLM creates an LLM client based on the provider configuration.
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
	default:
		return nil, fmt.Errorf("unsupported AI provider: %q (supported: openai, anthropic, googleai, ollama)", provider)
	}
}

// ExtractBiodata sends the raw PDF text to an LLM and returns structured biodata.
func ExtractBiodata(ctx context.Context, text string, cfg AIConfig) (map[string]any, error) {
	llm, err := newLLM(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client (%s): %w", cfg.Provider, err)
	}

	fullPrompt := systemPrompt + "\n\n---\n\nResume Text:\n" + text

	resp, err := llm.Call(ctx, fullPrompt,
		llms.WithTemperature(0.1),
		llms.WithMaxTokens(16384),
	)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed (%s/%s): %w", cfg.Provider, cfg.Model, err)
	}

	// Strip potential markdown fences from response
	cleaned := strings.TrimSpace(resp)
	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
	} else if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
	}
	if strings.HasSuffix(cleaned, "```") {
		cleaned = strings.TrimSuffix(cleaned, "```")
	}
	cleaned = strings.TrimSpace(cleaned)

	var result map[string]any
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response as JSON: %w\nraw response: %s", err, resp)
	}

	return result, nil
}
