package aisuites

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

// GenerateCoverLetter sends CV biodata, job description, tone and max paragraphs to an LLM to generate a cover letter.
func GenerateCoverLetter(ctx context.Context, biodataJSON string, jobDesc string, tone string, maxParagraphs int, language string, cfg AIConfig) (string, int64, error) {
	llm, err := newLLM(ctx, cfg)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create LLM client (%s): %w", cfg.Provider, err)
	}

	langInstruction := "Write the cover letter entirely in English."
	//nolint:gocritic // ifElseChain is fine here
	if strings.EqualFold(language, "id") {
		langInstruction = "Write the cover letter entirely in Bahasa Indonesia."
	} else if strings.EqualFold(language, "ja") {
		langInstruction = "Write the cover letter entirely in Japanese."
	} else if strings.EqualFold(language, "zh") {
		langInstruction = "Write the cover letter entirely in Chinese."
	} else if strings.EqualFold(language, "ko") {
		langInstruction = "Write the cover letter entirely in Korean."
	}

	if maxParagraphs < 1 {
		maxParagraphs = 3
	}
	if maxParagraphs > 10 {
		maxParagraphs = 10
	}

	if tone == "" {
		tone = "Professional"
	}

	prompt := fmt.Sprintf(`You are an expert career consultant and copywriter.
Generate a highly tailored cover letter based on the provided CV Data and Job Description.

Tone/Style: %s
Maximum Paragraphs: %d
Language Constraint: %s

Guidelines:
- Do not include placeholders like "[Your Name]" if the data is available in the CV; use the actual data. If the specific hiring manager name is missing from the Job Description, you can use a general professional greeting ("Dear Hiring Manager,").
- Make sure it flows naturally and hits the key requirements in the Job Description, bridging them with the candidate's actual experience from the CV Data.
- Keep the response strictly to the cover letter content (no conversational filler from the AI).
- Do not use markdown code blocks for the letter. Just return the text.

---
Job Description:
%s

---
CV Data:
%s
`, tone, maxParagraphs, langInstruction, truncateString(jobDesc, 15000), truncateString(biodataJSON, 20000))

	resp, err := llm.GenerateContent(ctx,
		[]llms.MessageContent{
			{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextContent{Text: prompt}},
			},
		},
		llms.WithTemperature(0.5),
		llms.WithMaxTokens(2048),
	)
	if err != nil {
		return "", 0, fmt.Errorf("LLM call failed (%s/%s): %w", cfg.Provider, cfg.Model, err)
	}

	if len(resp.Choices) == 0 {
		return "", 0, fmt.Errorf("LLM returned no choices (%s/%s)", cfg.Provider, cfg.Model)
	}

	var totalTokens int64
	if info := resp.Choices[0].GenerationInfo; info != nil {
		totalTokens = extractTotalTokens(info)
	}

	resultText := strings.TrimSpace(resp.Choices[0].Content)
	return resultText, totalTokens, nil
}
