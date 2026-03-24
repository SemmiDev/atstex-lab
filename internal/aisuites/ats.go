package aisuites

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

// ATSCritiqueResult holds the structured output from evaluating a CV against a Job Description.
type ATSCritiqueResult struct {
	Score           int      `json:"score"`
	MissingKeywords []string `json:"missingKeywords"`
	Recommendations string   `json:"recommendations"`
}

// ScoreATS sends CV biodata JSON and a Job Description to the LLM to act as an ATS simulator.
func ScoreATS(ctx context.Context, biodataJSON string, jobDesc string, language string, cfg AIConfig) (*ATSCritiqueResult, int64, error) {
	llm, err := newLLM(ctx, cfg)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create LLM client (%s): %w", cfg.Provider, err)
	}

	langInstruction := langEnglish
	//nolint:gocritic // ifElseChain is fine here
	if strings.EqualFold(language, "id") {
		langInstruction = langBahasaIndonesia
	} else if strings.EqualFold(language, "ja") {
		langInstruction = langJapanese
	} else if strings.EqualFold(language, "zh") {
		langInstruction = langChinese
	} else if strings.EqualFold(language, "ko") {
		langInstruction = langKorean
	}

	atsPrompt := `You are an advanced Applicant Tracking System (ATS) and a Senior Technical Recruiter.
Analyze the provided CV/resume data against the given Job Description. ` + langInstruction + `

Return ONLY valid JSON with this exact structure (no markdown fences, no explanation):

{
  "score": <number 0-100 indicating the match percentage>,
  "missingKeywords": ["<keyword 1>", "<keyword 2>"],
  "recommendations": "<specific actionable advice on how to improve the CV for this specific job>"
}

Scoring criteria (total 100 points):
- Exact Match (40 pts): Presence of exact required skills/technologies.
- Experience Relevance (30 pts): How well past roles map to the job responsibilities.
- Education/Certifications (10 pts): Match with educational requirements.
- Soft Skills & Culture (20 pts): Evidence of required soft skills.

Be highly critical. Highlight the most crucial missing keywords (limit to top 15).

Job Description:
` + truncateString(jobDesc, 15000) + `

CV Data:
` + truncateString(biodataJSON, 20000)

	resp, err := llm.GenerateContent(ctx,
		[]llms.MessageContent{
			{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextContent{Text: atsPrompt}},
			},
		},
		llms.WithTemperature(0.2), // keep it analytical
		llms.WithMaxTokens(1024),
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
	var result ATSCritiqueResult
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, totalTokens, fmt.Errorf("failed to parse ATS response as JSON: %w\nraw: %s", err, resp.Choices[0].Content)
	}

	if result.Score < 0 {
		result.Score = 0
	} else if result.Score > 100 {
		result.Score = 100
	}

	return &result, totalTokens, nil
}
