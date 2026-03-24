package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CustomTemplate represents a user-created resume template built via the drag-and-drop tool.
type CustomTemplate struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	UserID    uuid.UUID       `json:"user_id" db:"user_id"`
	Name      string          `json:"name" db:"name"`
	Config    json.RawMessage `json:"config" db:"config"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}
