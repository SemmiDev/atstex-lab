package aisuites

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

// EnhanceBulletPoint rewrites one or more CV bullet point lines using strong
// action verbs and quantifiable metrics. The number of lines is preserved.
// Returns the enhanced text, tokens consumed, and any error.
func EnhanceBulletPoint(ctx context.Context, rawBullet string, language string, cfg AIConfig) (string, int64, error) {
	llm, err := newLLM(ctx, cfg)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create LLM client (%s): %w", cfg.Provider, err)
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

	prompt := fmt.Sprintf(`You are an expert CV/resume writer specializing in ATS-optimized resumes.
Rewrite the following CV bullet point(s) to:
- Start each line with a strong, varied action verb (e.g. Engineered, Reduced, Spearheaded, Automated)
- Include quantifiable metrics wherever they can be reasonably inferred or improved (e.g. "by 40%%", "for 10k+ users")
- Be concise, impactful, and professional
- Preserve the exact same number of lines as the original
- Do NOT add any lines that were not implied by the original content

%s

Return ONLY the rewritten bullet point(s), one per line — no explanations, no headers, no markdown, no leading dashes.

Original bullet point(s):
%s`, langInstruction, truncateString(rawBullet, 2000))

	resp, err := llm.GenerateContent(ctx,
		[]llms.MessageContent{
			{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextContent{Text: prompt}},
			},
		},
		llms.WithTemperature(0.4),
		llms.WithMaxTokens(600),
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

	enhanced := strings.TrimSpace(resp.Choices[0].Content)
	// Strip any accidental markdown fences the model may have added
	if strings.HasPrefix(enhanced, "```") {
		lines := strings.Split(enhanced, "\n")
		var stripped []string
		for _, l := range lines {
			if strings.HasPrefix(l, "```") {
				continue
			}
			stripped = append(stripped, l)
		}
		enhanced = strings.TrimSpace(strings.Join(stripped, "\n"))
	}

	return enhanced, totalTokens, nil
}
