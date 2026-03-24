package aisuites

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

// AutoTailorCV sends CV biodata JSON and a Job Description to the LLM to rewrite the Summary and Experience sections.
// Note: It assumes the structure of biodataJSON matches what extractor generates.
func AutoTailorCV(ctx context.Context, biodataJSON string, jobDesc string, language string, cfg AIConfig) ([]byte, int64, error) {
	llm, err := newLLM(ctx, cfg)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create LLM client (%s): %w", cfg.Provider, err)
	}

	langInstruction := langEnglish
	//nolint:gocritic // ifElseChain is intentional for language routing
	if strings.EqualFold(language, "id") {
		langInstruction = langBahasaIndonesia
	} else if strings.EqualFold(language, "ja") {
		langInstruction = langJapanese
	} else if strings.EqualFold(language, "zh") {
		langInstruction = langChinese
	} else if strings.EqualFold(language, "ko") {
		langInstruction = langKorean
	}

	prompt := fmt.Sprintf(`You are an expert CV/resume writer and career coach.
I will provide you with a candidate's existing CV Data (in JSON format) and a target Job Description.

Your task is to duplicate the provided JSON CV Data exactly, but rewrite the following key sections to better match the Job Description:
1. **summary**: Rewrite the summary to highlight the candidate's most relevant qualifications and align with the keywords and requirements of the Job Description. It should be an impactful professional summary.
2. **experience**: For each job in the experience array, rewrite the "bullets" to emphasize the skills, keywords, and responsibilities that are relevant to the Job Description. Use strong action verbs and maintain any quantifiable metrics from the original text while blending in the keywords naturally.

DO NOT invent new experiences, jobs, or degrees. ONLY enhance and reframe the existing experiences.
DO NOT modify the structure of the JSON. Return the exact same schema.
DO NOT wrap the output in markdown blocks (e.g. no "'''json"). JUST the raw JSON string.
%s

---
Job Description:
%s

---
CV Data:
%s
`, langInstruction, truncateString(jobDesc, 15000), truncateString(biodataJSON, 20000))

	resp, err := llm.GenerateContent(ctx,
		[]llms.MessageContent{
			{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextContent{Text: prompt}},
			},
		},
		llms.WithTemperature(0.4),
		llms.WithMaxTokens(16384),
		llms.WithJSONMode(),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("LLM call failed (%s/%s): %w", cfg.Provider, cfg.Model, err)
	}

	if len(resp.Choices) == 0 {
		return nil, 0, fmt.Errorf("LLM returned no choices (%s/%s)", cfg.Provider, cfg.Model)
	}

	var totalTokens int64
	if info := resp.Choices[0].GenerationInfo; info != nil {
		totalTokens = extractTotalTokens(info)
	}

	cleaned := stripMarkdownFences(resp.Choices[0].Content)
	var result map[string]any
	if errUnmarshal := json.Unmarshal([]byte(cleaned), &result); errUnmarshal != nil {
		return nil, totalTokens, fmt.Errorf("failed to parse AutoTailorCV response as JSON: %w\nraw: %s", errUnmarshal, resp.Choices[0].Content)
	}

	finalResp, err := json.Marshal(result)
	if err != nil {
		return nil, totalTokens, fmt.Errorf("failed to marshal AutoTailorCV response back to JSON: %w", err)
	}

	return finalResp, totalTokens, nil
}
