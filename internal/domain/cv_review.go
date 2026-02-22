package domain

import (
	"time"

	"github.com/google/uuid"
)

// CVReview stores an AI-generated critique/score for a CV profile.
type CVReview struct {
	ID              uuid.UUID `db:"id"              json:"id"`
	UserID          uuid.UUID `db:"user_id"         json:"userId"`
	ProfileID       uuid.UUID `db:"profile_id"      json:"profileId"`
	ProfileTitle    string    `db:"profile_title"    json:"profileTitle"`
	Language        string    `db:"language"         json:"language"`
	Score           int       `db:"score"            json:"score"`
	Strengths       string    `db:"strengths"        json:"strengths"`
	Improvements    string    `db:"improvements"     json:"improvements"`
	Recommendations string    `db:"recommendations"  json:"recommendations"`
	TokensUsed      int64     `db:"tokens_used"      json:"tokensUsed"`
	CreatedAt       time.Time `db:"created_at"       json:"createdAt"`
}
