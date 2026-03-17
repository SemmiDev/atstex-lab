package aisuites

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tmc/langchaingo/llms"
)

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

	var totalTokens int64
	var genInfo map[string]any
	if info := resp.Choices[0].GenerationInfo; info != nil {
		totalTokens = extractTotalTokens(info)
		genInfo = info
	}

	cleaned := stripMarkdownFences(resp.Choices[0].Content)
	var result map[string]any
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, totalTokens, genInfo, fmt.Errorf("failed to parse LLM response as JSON: %w\nraw response: %s", err, resp.Choices[0].Content)
	}

	return result, totalTokens, genInfo, nil
}
