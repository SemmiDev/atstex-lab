package aisuites

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/semmidev/atstex-lab/internal/domain"
	"github.com/tmc/langchaingo/llms"
)

// CritiqueInterviewAnswer takes a single interview question and the candidate's spoken (transcribed)
// answer, then returns structured AI feedback to help them improve their delivery and content.
func CritiqueInterviewAnswer(ctx context.Context, question, answer, language string, cfg AIConfig) (*domain.InterviewAnswerCritique, int64, error) {
	llm, err := newLLM(ctx, cfg)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create LLM client (%s): %w", cfg.Provider, err)
	}

	langInstruction := "Provide your entire critique in English."
	//nolint:gocritic // ifElseChain is fine here
	if strings.EqualFold(language, "id") {
		langInstruction = "Provide your entire critique in Bahasa Indonesia."
	} else if strings.EqualFold(language, "ja") {
		langInstruction = "Provide your entire critique in Japanese."
	} else if strings.EqualFold(language, "zh") {
		langInstruction = "Provide your entire critique in Chinese."
	} else if strings.EqualFold(language, "ko") {
		langInstruction = "Provide your entire critique in Korean."
	}

	prompt := fmt.Sprintf(`You are an expert career coach and interview trainer with 20+ years of experience conducting and evaluating job interviews.

A candidate gave a spoken answer to the following interview question. Evaluate their response and provide structured, actionable feedback to help them improve.

%s

Interview Question:
%s

Candidate's Spoken Answer:
%s

Evaluate the response across these dimensions:
- Relevance and completeness (did they actually answer what was asked?)
- Use of the STAR method (Situation, Task, Action, Result) for behavioral questions
- Specific examples and quantifiable achievements cited
- Clarity, logical structure, and conciseness
- Professional tone and vocabulary

Scoring guide: 9-10 = exceptional, 7-8 = strong, 5-6 = adequate, 3-4 = needs significant work, 1-2 = significantly off-mark or blank.

Return ONLY valid JSON with this exact structure (no markdown fences, no explanation):
{
  "score": <integer from 1 to 10>,
  "strengths": ["strength 1", "strength 2"],
  "improvements": ["specific improvement area 1", "specific improvement area 2"],
  "suggestedAnswer": "A concise outline of the key points a strong answer should cover for this specific question...",
  "deliveryTips": ["concrete delivery tip 1", "concrete delivery tip 2"]
}
`, langInstruction, truncateString(question, 2000), truncateString(answer, 4000))

	resp, err := llm.GenerateContent(ctx,
		[]llms.MessageContent{
			{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextContent{Text: prompt}},
			},
		},
		llms.WithTemperature(0.4),
		llms.WithMaxTokens(1500),
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
	var result domain.InterviewAnswerCritique
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, totalTokens, fmt.Errorf("failed to parse critique response as JSON: %w\nraw: %s", err, resp.Choices[0].Content)
	}

	return &result, totalTokens, nil
}
