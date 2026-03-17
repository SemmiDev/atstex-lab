package aisuites

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

// MockInterviewMessage represents a single message in the interview conversation
type MockInterviewMessage struct {
	Role    string `json:"role"` // "system", "human", "ai"
	Content string `json:"content"`
}

// MockInterviewTurn takes the conversation history, the candidate's last answer,
// and background context (CV, Job Description) to generate the next interview question or response.
func HandleMockInterviewTurn(
	ctx context.Context,
	history []MockInterviewMessage,
	biodataJSON string,
	jobDesc string,
	language string,
	interviewerStyle string,
	cfg AIConfig,
) (string, []MockInterviewMessage, int64, error) {
	llm, err := newLLM(ctx, cfg)
	if err != nil {
		return "", nil, 0, fmt.Errorf("failed to create LLM client (%s): %w", cfg.Provider, err)
	}

	langInstruction := "Conduct the interview entirely in English."
	if strings.EqualFold(language, "id") {
		langInstruction = "Conduct the interview entirely in Bahasa Indonesia."
	} else if strings.EqualFold(language, "ja") {
		langInstruction = "Conduct the interview entirely in Japanese."
	} else if strings.EqualFold(language, "zh") {
		langInstruction = "Conduct the interview entirely in Chinese."
	} else if strings.EqualFold(language, "ko") {
		langInstruction = "Conduct the interview entirely in Korean."
	}

	styleInstruction := "Use a balanced interview style: professional, warm, and appropriately challenging."
	switch strings.ToLower(strings.TrimSpace(interviewerStyle)) {
	case "friendly":
		styleInstruction = "Use a friendly, encouraging interview style. Keep pressure low, ask clear questions, and help the candidate open up with gentle follow-ups."
	case "strict":
		styleInstruction = "Use a strict, high-bar interview style. Be direct, skeptical, and concise. Ask tougher follow-ups when answers are vague, and press for specifics, metrics, and trade-offs."
	case "technical":
		styleInstruction = "Use a technical interview style. Emphasize systems, problem-solving, and deep reasoning. Ask about architecture, edge cases, trade-offs, and how they would debug or design solutions."
	case "behavioral":
		styleInstruction = "Use a behavioral interview style. Emphasize STAR-format questions, collaboration, conflict, leadership, and decision-making. Ask for specific examples and outcomes."
	default:
		// balanced
	}

	// Build the system prompt if we are at the start of the conversation
	if len(history) == 0 {
		systemPrompt := fmt.Sprintf(`You are an expert HR Recruiter and Technical Interviewer.
You are conducting a live voice mock interview.

Guidelines for your responses:
1. Keep your responses conversational, concise, and natural (as if spoken aloud).
2. Avoid using markdown lists, code blocks, or special formatting since your response will be read by Text-To-Speech.
3. Ask ONE specific question at a time. Do not overwhelm the candidate.
4. If the candidate answers well, briefly acknowledge it before moving to the next question.
5. If the candidate's answer is lacking, you may ask a follow-up probing question.
6. Base your questions on the candidate's CV and the Job Description.
7. Interviewer style: %s
8. %s

Candidate CV Data:
%s

Job Description:
%s

Start the interview by greeting the candidate by name (if known from the CV) and asking the first question.`,
			styleInstruction, langInstruction, truncateString(biodataJSON, 5000), truncateString(jobDesc, 5000))

		history = append(history, MockInterviewMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	// Convert our history to LangChain message types
	var messages []llms.MessageContent
	for _, msg := range history {
		var role llms.ChatMessageType
		switch msg.Role {
		case "system":
			role = llms.ChatMessageTypeSystem
		case "human":
			role = llms.ChatMessageTypeHuman
		case "ai":
			role = llms.ChatMessageTypeAI
		default:
			role = llms.ChatMessageTypeHuman
		}

		messages = append(messages, llms.MessageContent{
			Role:  role,
			Parts: []llms.ContentPart{llms.TextContent{Text: msg.Content}},
		})
	}

	resp, err := llm.GenerateContent(ctx,
		messages,
		llms.WithTemperature(0.6), // slightly more creative for conversation
		llms.WithMaxTokens(1024),
	)
	if err != nil {
		return "", history, 0, fmt.Errorf("LLM call failed (%s/%s): %w", cfg.Provider, cfg.Model, err)
	}

	if len(resp.Choices) == 0 {
		return "", history, 0, fmt.Errorf("LLM returned no choices (%s/%s)", cfg.Provider, cfg.Model)
	}

	var totalTokens int64
	if info := resp.Choices[0].GenerationInfo; info != nil {
		totalTokens = extractTotalTokens(info)
	}

	aiResponse := strings.TrimSpace(resp.Choices[0].Content)

	// Add the AI's response to the history before returning
	newHistory := append(history, MockInterviewMessage{
		Role:    "ai",
		Content: aiResponse,
	})

	return aiResponse, newHistory, totalTokens, nil
}
