// Package extractor uses an LLM to extract structured biodata from raw text.
package extractor

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/mistral"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
)

// toInt64 converts various numeric types to int64.
func toInt64(v interface{}) (int64, bool) {
	switch t := v.(type) {
	case int:
		return int64(t), true
	case int64:
		return t, true
	case float64:
		return int64(t), true
	case int32:
		return int64(t), true
	}
	return 0, false
}

// extractTotalTokens tries to get TotalTokens from GenerationInfo.
// Supports:
//   - Flat map keys: "TotalTokens" (OpenAI, Google)
//   - Nested struct: info["usage"].TotalTokens (Mistral: UsageInfo struct)
func extractTotalTokens(info map[string]any) int64 {
	// 1. Try flat key (OpenAI, Google, etc.)
	if v, ok := info["TotalTokens"]; ok {
		if n, ok := toInt64(v); ok {
			return n
		}
	}

	// 2. Try nested "usage" (Mistral SDK stores UsageInfo struct)
	if usage, ok := info["usage"]; ok && usage != nil {
		// If it's a map (JSON-like)
		if m, ok := usage.(map[string]any); ok {
			if v, ok := m["total_tokens"]; ok {
				if n, ok := toInt64(v); ok {
					return n
				}
			}
			if v, ok := m["TotalTokens"]; ok {
				if n, ok := toInt64(v); ok {
					return n
				}
			}
		}
		// If it's a struct (e.g. sdk.UsageInfo), use reflection
		rv := reflect.ValueOf(usage)
		if rv.Kind() == reflect.Struct {
			f := rv.FieldByName("TotalTokens")
			if f.IsValid() && f.CanInterface() {
				if n, ok := toInt64(f.Interface()); ok {
					return n
				}
			}
		}
	}

	return 0
}

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

// ExtractBiodata sends the raw PDF text to an LLM and returns structured biodata
// along with total tokens consumed (0 if the provider doesn't report usage)
// and the raw GenerationInfo map for logging.
func ExtractBiodata(ctx context.Context, text string, cfg AIConfig) (map[string]any, int64, map[string]any, error) {
	llm, err := newLLM(ctx, cfg)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to create LLM client (%s): %w", cfg.Provider, err)
	}

	userPrompt := systemPrompt + "\n\n---\n\nResume Text:\n" + text

	resp, err := llm.GenerateContent(ctx,
		[]llms.MessageContent{
			{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextContent{Text: userPrompt}},
			},
		},
		llms.WithTemperature(0.1),
		llms.WithMaxTokens(16384),
	)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("LLM call failed (%s/%s): %w", cfg.Provider, cfg.Model, err)
	}

	if len(resp.Choices) == 0 {
		return nil, 0, nil, fmt.Errorf("LLM returned no choices (%s/%s)", cfg.Provider, cfg.Model)
	}

	// Extract token usage from GenerationInfo.
	// Different providers store tokens differently:
	//   - OpenAI/Google: flat keys like "TotalTokens", "PromptTokens", "CompletionTokens"
	//   - Mistral: nested struct under "usage" with fields TotalTokens, PromptTokens, CompletionTokens
	var totalTokens int64
	var genInfo map[string]any
	if info := resp.Choices[0].GenerationInfo; info != nil {
		totalTokens = extractTotalTokens(info)
		genInfo = info
	}

	// Strip potential markdown fences from response
	cleaned := strings.TrimSpace(resp.Choices[0].Content)
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
		return nil, totalTokens, genInfo, fmt.Errorf("failed to parse LLM response as JSON: %w\nraw response: %s", err, resp.Choices[0].Content)
	}

	return result, totalTokens, genInfo, nil
}
