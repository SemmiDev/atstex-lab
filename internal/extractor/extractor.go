// Package extractor uses an LLM to extract structured biodata from raw text.
package extractor

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/semmidev/atstex-lab/internal/domain"
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
//
//nolint:gocognit,nestif // extraction logic handles complex nested map structures output by LLMs
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

// truncateString helper ensures strings do not exceed maxRunes to prevent HTTP 400 Payload Too Large errors.
func truncateString(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) > maxRunes {
		return string(runes[:maxRunes]) + "...\n[Content truncated due to length limits]"
	}
	return s
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
//
//nolint:staticcheck // SA1019: llms.LLM is deprecated but still required for older langchaingo versions
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
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	var result map[string]any
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, totalTokens, genInfo, fmt.Errorf("failed to parse LLM response as JSON: %w\nraw response: %s", err, resp.Choices[0].Content)
	}

	return result, totalTokens, genInfo, nil
}

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

	// Strip potential markdown fences
	cleaned := strings.TrimSpace(resp.Choices[0].Content)
	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
	} else if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
	}
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

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

	// Strip potential markdown fences
	cleaned := strings.TrimSpace(resp.Choices[0].Content)
	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
	} else if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
	}
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

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

// GenerateInterviewQuestions sends CV biodata, job description, language and question count to an LLM to generate targeted interview questions.
// EnhanceBulletPoint rewrites one or more CV bullet point lines using strong
// action verbs and quantifiable metrics. The number of lines is preserved.
// Returns the enhanced text, tokens consumed, and any error.
func EnhanceBulletPoint(ctx context.Context, rawBullet string, language string, cfg AIConfig) (string, int64, error) {
	llm, err := newLLM(ctx, cfg)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create LLM client (%s): %w", cfg.Provider, err)
	}

	langInstruction := "Respond in English."
	//nolint:gocritic // ifElseChain is intentional for language routing
	if strings.EqualFold(language, "id") {
		langInstruction = "Respond entirely in Bahasa Indonesia."
	} else if strings.EqualFold(language, "ja") {
		langInstruction = "Respond entirely in Japanese."
	} else if strings.EqualFold(language, "zh") {
		langInstruction = "Respond entirely in Chinese."
	} else if strings.EqualFold(language, "ko") {
		langInstruction = "Respond entirely in Korean."
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

Return ONLY valid JSON with this exact structure (no markdown fences, no explanation):
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
		llms.WithTemperature(0.5),
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

	cleaned := strings.TrimSpace(resp.Choices[0].Content)
	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
	} else if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
	}
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	var result domain.InterviewPrepResult
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, totalTokens, fmt.Errorf("failed to parse interview questions response as JSON: %w\nraw: %s", err, resp.Choices[0].Content)
	}

	return &result, totalTokens, nil
}
