package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents the central domain entity for an authenticated user.
type User struct {
	ID           uuid.UUID `db:"id"`
	GoogleID     string    `db:"google_id"`
	Email        string    `db:"email"`
	Name         string    `db:"name"`
	Picture      string    `db:"picture"`
	Role         string    `db:"role"`
	AITokensUsed int64     `db:"ai_tokens_used"`
	Username     *string   `db:"username"`
	IsBlocked    bool      `db:"is_blocked"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// IsAdmin returns true if the user has the admin role.
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// AdminStats holds aggregate dashboard statistics.
type AdminStats struct {
	TotalUsers    int   `json:"totalUsers"      db:"total_users"`
	TotalAdmins   int   `json:"totalAdmins"     db:"total_admins"`
	TotalAITokens int64 `json:"totalAITokens"    db:"total_ai_tokens"`
	TotalBiodata  int   `json:"totalBiodata"    db:"total_biodata"`
	TotalSessions int   `json:"totalSessions"   db:"total_sessions"`
}

// AdminUserRow is returned by the admin user listing endpoint.
type AdminUserRow struct {
	ID           uuid.UUID `json:"id"            db:"id"`
	Email        string    `json:"email"         db:"email"`
	Name         string    `json:"name"          db:"name"`
	Picture      string    `json:"picture"       db:"picture"`
	Role         string    `json:"role"          db:"role"`
	Username     *string   `json:"username"      db:"username"`
	IsBlocked    bool      `json:"isBlocked"     db:"is_blocked"`
	AITokensUsed int64     `json:"aiTokensUsed"   db:"ai_tokens_used"`
	BiodataCount int       `json:"biodataCount"  db:"biodata_count"`
	CreatedAt    time.Time `json:"createdAt"     db:"created_at"`
}

// AdminListParams controls pagination, search, and ordering for user listing.
type AdminListParams struct {
	Page    int
	PerPage int
	Search  string
	Sort    string // column name
	Order   string // asc|desc
}
