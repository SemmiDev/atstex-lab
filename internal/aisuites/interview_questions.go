package aisuites

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/semmidev/atstex-lab/internal/domain"
	"github.com/tmc/langchaingo/llms"
)

func GenerateInterviewQuestions(ctx context.Context, biodataJSON string, jobDesc string, language string, count int, cfg AIConfig) (*domain.InterviewPrepResult, int64, error) {
	llm, err := newLLM(ctx, cfg)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create LLM client (%s): %w", cfg.Provider, err)
	}

	langInstruction := "Generate the questions entirely in English."
	//nolint:gocritic // ifElseChain is fine here
	if strings.EqualFold(language, "id") {
		langInstruction = "Generate the questions entirely in Bahasa Indonesia."
	} else if strings.EqualFold(language, "ja") {
		langInstruction = "Generate the questions entirely in Japanese."
	} else if strings.EqualFold(language, "zh") {
		langInstruction = "Generate the questions entirely in Chinese."
	} else if strings.EqualFold(language, "ko") {
		langInstruction = "Generate the questions entirely in Korean."
	}

	if count < 5 {
		count = 5
	}
	if count > 20 {
		count = 20
	}

	prompt := fmt.Sprintf(`You are an expert HR Recruiter and Technical Interviewer.
Based on the candidate's CV data and the target Job Description, generate exactly %d highly tailored interview questions.
%s

You MUST organize the questions into exactly the following 7 categories:
1. Behavioral Questions
2. Technical Questions
3. Problem-Solving / Analytical Questions
4. Situational Questions
5. Cultural Fit Questions
6. Career / Background Questions
7. Case Study / System Design Questions

Distribute the %d questions across these 7 categories logically based on the role.

Return ONLY valid JSON with this exact structure (no markdown fences, no explanation, no trailing commas):
{
  "categories": [
    {
      "category": "Behavioral Questions",
      "questions": [
        "Question 1 here",
        "Question 2 here"
      ]
    },
    ...
  ]
}

Ensure the questions explicitly reference details from BOTH the Job Description and the CV Data where applicable.

---
Job Description:
%s

---
CV Data:
%s
`, count, langInstruction, count, truncateString(jobDesc, 15000), truncateString(biodataJSON, 20000))

	resp, err := llm.GenerateContent(ctx,
		[]llms.MessageContent{
			{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextContent{Text: prompt}},
			},
		},
		llms.WithTemperature(0.2),
		llms.WithMaxTokens(3000),
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
	var result domain.InterviewPrepResult
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, totalTokens, fmt.Errorf("failed to parse interview questions response as JSON: %w\nraw: %s", err, resp.Choices[0].Content)
	}

	return &result, totalTokens, nil
}
