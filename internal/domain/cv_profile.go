package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CVProfile represents a single CV/resume profile belonging to a user.
type CVProfile struct {
	ID        uuid.UUID       `db:"id"         json:"id"`
	UserID    uuid.UUID       `db:"user_id"    json:"user_id"`
	Title     string          `db:"title"      json:"title"`
	Biodata   json.RawMessage `db:"biodata"    json:"biodata"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt time.Time       `db:"updated_at" json:"updated_at"`
}
