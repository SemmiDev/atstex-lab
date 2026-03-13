package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MockInterviewSession stores a complete live mock interview session.
type MockInterviewSession struct {
	ID             uuid.UUID       `db:"id"              json:"id"`
	UserID         uuid.UUID       `db:"user_id"         json:"userId"`
	ProfileID      uuid.UUID       `db:"profile_id"      json:"profileId"`
	JobDescription string          `db:"job_description" json:"jobDescription"`
	Language       string          `db:"language"        json:"language"`
	Messages       json.RawMessage `db:"messages"        json:"messages"`
	TokensUsed     int64           `db:"tokens_used"     json:"tokensUsed"`
	TurnCount      int             `db:"turn_count"      json:"turnCount"`
	CreatedAt      time.Time       `db:"created_at"      json:"createdAt"`
	EndedAt        *time.Time      `db:"ended_at"        json:"endedAt"`
}

// MockInterviewMessage is a single message in the interview conversation.
type MockInterviewMessage struct {
	Role      string    `json:"role"` // "ai" | "user"
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
}

// MockInterviewWSMessage is the WebSocket protocol message exchanged between
// the browser client and the server.
type MockInterviewWSMessage struct {
	Type      string `json:"type"`
	Text      string `json:"text,omitempty"`
	SessionID string `json:"sessionId,omitempty"`
	TurnID    int    `json:"turnId,omitempty"`
	Error     string `json:"error,omitempty"`
}

// WSMessageType constants for the WebSocket protocol.
const (
	WSTypePing       = "ping"
	WSTypePong       = "pong"
	WSTypeAIMessage  = "ai_message"
	WSTypeUserMsg    = "user_message"
	WSTypeEndSession = "end_session"
	WSTypeEnded      = "session_ended"
	WSTypeError      = "error"
	WSTypeThinking   = "ai_thinking"
)
