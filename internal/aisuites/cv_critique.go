package aisuites

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

// CVCritiqueResult holds the structured output from a CV critique.
type CVCritiqueResult struct {
	Score           int    `json:"score"`
	Strengths       string `json:"strengths"`
	Improvements    string `json:"improvements"`
	Recommendations string `json:"recommendations"`
}

// CritiqueCVProfile sends CV biodata JSON to an LLM for critique and scoring.
func CritiqueCVProfile(ctx context.Context, biodataJSON string, language string, cfg AIConfig) (*CVCritiqueResult, int64, error) {
	llm, err := newLLM(ctx, cfg)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create LLM client (%s): %w", cfg.Provider, err)
	}

	langInstruction := "Respond entirely in English."
	//nolint:gocritic // ifElseChain is fine here
	if strings.EqualFold(language, "id") {
		langInstruction = "Respond entirely in Bahasa Indonesia."
	} else if strings.EqualFold(language, "ja") {
		langInstruction = "Respond entirely in Japanese."
	} else if strings.EqualFold(language, "zh") {
		langInstruction = "Respond entirely in Chinese."
	} else if strings.EqualFold(language, "ko") {
		langInstruction = "Respond entirely in Korean."
	}

	critiquePrompt := `You are a senior HR consultant and ATS (Applicant Tracking System) expert with 15+ years of experience reviewing resumes.

Analyze the following CV/resume data and provide a comprehensive critique. ` + langInstruction + `

Return ONLY valid JSON with this exact structure (no markdown fences, no explanation):

{
  "score": <number 0-100>,
  "strengths": "<bullet points of what the CV does well, each on a new line starting with •>",
  "improvements": "<bullet points of areas that need improvement, each on a new line starting with •>",
  "recommendations": "<bullet points of specific actionable recommendations, each on a new line starting with •>"
}

Scoring criteria (total 100 points):
- Professional Summary (15 pts): Clear, targeted, with quantifiable achievements
- Work Experience (25 pts): Action verbs, quantified results, relevance, progressive responsibility
- Education (10 pts): Relevance, GPA/honors if applicable, coursework
- Skills (15 pts): Relevant technical skills, organized by category, ATS-friendly keywords
- Projects (10 pts): Quality, relevance, technical depth
- Formatting & Completeness (10 pts): All sections present, no gaps, consistent formatting
- ATS Compatibility (15 pts): Proper keywords, parseable structure, no graphics/tables issues

Be specific and actionable in your feedback. Reference specific sections and suggest concrete improvements.

CV Data:
` + truncateString(biodataJSON, 30000)

	resp, err := llm.GenerateContent(ctx,
		[]llms.MessageContent{
			{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextContent{Text: critiquePrompt}},
			},
		},
		llms.WithTemperature(0.3),
		llms.WithMaxTokens(4096),
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
	var result CVCritiqueResult
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, totalTokens, fmt.Errorf("failed to parse critique response as JSON: %w\nraw: %s", err, resp.Choices[0].Content)
	}

	// Clamp score
	if result.Score < 0 {
		result.Score = 0
	} else if result.Score > 100 {
		result.Score = 100
	}

	return &result, totalTokens, nil
}
