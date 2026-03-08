package domain

import (
	"time"

	"github.com/google/uuid"
)

type JobApplication struct {
	ID          uuid.UUID  `db:"id"            json:"id"`
	UserID      uuid.UUID  `db:"user_id"       json:"user_id"`
	CVProfileID *uuid.UUID `db:"cv_profile_id" json:"cv_profile_id,omitempty"`
	Company     string     `db:"company"       json:"company"`
	JobTitle    string     `db:"job_title"     json:"job_title"`
	Status      string     `db:"status"        json:"status"`
	Notes       string     `db:"notes"         json:"notes"`
	CreatedAt   time.Time  `db:"created_at"    json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"    json:"updated_at"`
}
