package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents the central domain entity for an authenticated user.
type User struct {
	ID        uuid.UUID `db:"id"`
	GoogleID  string    `db:"google_id"`
	Email     string    `db:"email"`
	Name      string    `db:"name"`
	Picture   string    `db:"picture"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
