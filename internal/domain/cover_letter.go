package domain

import (
	"time"

	"github.com/google/uuid"
)

type CoverLetter struct {
	ID              uuid.UUID `db:"id"                json:"id"`
	UserID          uuid.UUID `db:"user_id"           json:"userId"`
	ProfileID       uuid.UUID `db:"profile_id"        json:"profileId"`
	ProfileTitle    string    `db:"profile_title"     json:"profileTitle"`
	JobDescription  string    `db:"job_description"   json:"jobDescription"`
	CoverLetterText string    `db:"cover_letter_text" json:"coverLetterText"`
	Language        string    `db:"language"          json:"language"`
	TokensUsed      int64     `db:"tokens_used"       json:"tokensUsed"`
	CreatedAt       time.Time `db:"created_at"        json:"createdAt"`
}
