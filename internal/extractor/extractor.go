// Package extractor uses an LLM to extract structured biodata from raw text.
package extractor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

const systemPrompt = `You are a resume parser. Given raw text extracted from a PDF resume/CV,
extract the information into the following JSON structure. Return ONLY valid JSON, no markdown fences, no explanation.

{
  "personal": {
    "name": "",
    "title": "",
    "email": "",
    "phone": "",
    "location": "",
    "linkedin": { "display": "", "url": "" },
    "github": { "display": "", "url": "" },
    "website": { "display": "", "url": "" }
  },
  "summary": "",
  "experience": [
    {
      "company": "",
      "title": "",
      "location": "",
      "dates": "",
      "bullets": "line1\nline2\nline3"
    }
  ],
  "education": [
    {
      "institution": "",
      "degree": "",
      "location": "",
      "dates": "",
      "gpa": "",
      "activities": ""
    }
  ],
  "projects": [
    {
      "name": "",
      "role": "",
      "link": "",
      "bullets": "line1\nline2"
    }
  ],
  "skills": {
    "languages": "",
    "frameworks": "",
    "tools": "",
    "other": ""
  },
  "certifications": [
    { "name": "", "issuer": "" }
  ],
  "volunteer": [
    {
      "organization": "",
      "role": "",
      "location": "",
      "dates": "",
      "bullets": "line1\nline2"
    }
  ],
  "awards": [
    {
      "title": "",
      "issuer": "",
      "date": "",
      "description": ""
    }
  ],
  "talks": [
    {
      "title": "",
      "event": "",
      "location": "",
      "date": "",
      "description": ""
    }
  ]
}

Rules:
- Fill in as many fields as possible from the provided text.
- If a section has no data, use an empty array [] or empty string "".
- For bullet points, put each bullet on a new line separated by \n, WITHOUT leading dashes or bullet characters.
- For dates, use the format found in the resume (e.g. "Jan 2020 — Present").
- For skills, group them into languages, frameworks, tools, and other as best as you can.
- Return ONLY the JSON object, nothing else.`

// ExtractBiodata sends the raw PDF text to an LLM and returns structured biodata.
func ExtractBiodata(ctx context.Context, text, apiKey string) (map[string]any, error) {
	llm, err := openai.New(
		openai.WithModel("gpt-4o-mini"),
		openai.WithToken(apiKey),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	fullPrompt := systemPrompt + "\n\n---\n\nResume Text:\n" + text

	resp, err := llm.Call(ctx, fullPrompt,
		llms.WithTemperature(0.1),
		llms.WithMaxTokens(16384),
	)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response as JSON: %w\nraw response: %s", err, resp)
	}

	return result, nil
}
