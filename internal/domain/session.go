package domain

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an OAuth token linkage.
type Session struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Token     string    `db:"token"`
	IPAddress string    `db:"ip_address"`
	UserAgent string    `db:"user_agent"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}
