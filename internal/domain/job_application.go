package domain

import (
	"time"

	"github.com/google/uuid"
)

type JobApplication struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	CVProfileID *uuid.UUID `json:"cv_profile_id,omitempty" db:"cv_profile_id"`
	Company     string     `json:"company" db:"company"`
	JobTitle    string     `json:"job_title" db:"job_title"`
	Status      string     `json:"status" db:"status"`
	Notes       string     `json:"notes" db:"notes"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}
