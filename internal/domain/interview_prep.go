package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// InterviewPrep stores the history of an AI Interview Prep generation.
type InterviewPrep struct {
	ID             uuid.UUID       `db:"id"               json:"id"`
	UserID         uuid.UUID       `db:"user_id"          json:"userId"`
	ProfileID      uuid.UUID       `db:"profile_id"       json:"profileId"`
	JobDescription string          `db:"job_description"  json:"jobDescription"`
	Language       string          `db:"language"         json:"language"`
	Questions      json.RawMessage `db:"questions"        json:"questions"`
	TokensUsed     int64           `db:"tokens_used"      json:"tokensUsed"`
	CreatedAt      time.Time       `db:"created_at"       json:"createdAt"`
}

// InterviewQuestionCategory groups questions by their type.
type InterviewQuestionCategory struct {
	Category  string   `json:"category"`
	Questions []string `json:"questions"`
}

// InterviewPrepResult holds the structured output from the LLM for interview prep.
type InterviewPrepResult struct {
	Categories []InterviewQuestionCategory `json:"categories"`
}

// InterviewAnswerCritique holds AI feedback on a candidate's spoken interview answer.
type InterviewAnswerCritique struct {
	Score           int      `json:"score"`           // 1–10 overall quality score
	Strengths       []string `json:"strengths"`       // what the candidate did well
	Improvements    []string `json:"improvements"`    // concrete areas to improve
	SuggestedAnswer string   `json:"suggestedAnswer"` // key points a strong answer should cover
	DeliveryTips    []string `json:"deliveryTips"`    // tips on structure, confidence, STAR method, etc.
}
