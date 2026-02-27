package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AtsSimulation stores the history of a CV scored against a Job Description.
type AtsSimulation struct {
	ID              uuid.UUID       `db:"id"               json:"id"`
	UserID          uuid.UUID       `db:"user_id"          json:"userId"`
	ProfileID       uuid.UUID       `db:"profile_id"       json:"profileId"`
	JobDescription  string          `db:"job_description"  json:"jobDescription"`
	Score           int             `db:"score"            json:"score"`
	MissingKeywords json.RawMessage `db:"missing_keywords" json:"missingKeywords"`
	Recommendations string          `db:"recommendations"  json:"recommendations"`
	CreatedAt       time.Time       `db:"created_at"       json:"createdAt"`
}
