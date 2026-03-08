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
	TotalUsers    int   `db:"total_users"     json:"totalUsers"`
	TotalAdmins   int   `db:"total_admins"    json:"totalAdmins"`
	TotalAITokens int64 `db:"total_ai_tokens" json:"totalAITokens"`
	TotalBiodata  int   `db:"total_biodata"   json:"totalBiodata"`
	TotalSessions int   `db:"total_sessions"  json:"totalSessions"`
}

// AdminUserRow is returned by the admin user listing endpoint.
type AdminUserRow struct {
	ID           uuid.UUID `db:"id"             json:"id"`
	Email        string    `db:"email"          json:"email"`
	Name         string    `db:"name"           json:"name"`
	Picture      string    `db:"picture"        json:"picture"`
	Role         string    `db:"role"           json:"role"`
	Username     *string   `db:"username"       json:"username"`
	IsBlocked    bool      `db:"is_blocked"     json:"isBlocked"`
	AITokensUsed int64     `db:"ai_tokens_used" json:"aiTokensUsed"`
	BiodataCount int       `db:"biodata_count"  json:"biodataCount"`
	CreatedAt    time.Time `db:"created_at"     json:"createdAt"`
}

// AdminListParams controls pagination, search, and ordering for user listing.
type AdminListParams struct {
	Page    int
	PerPage int
	Search  string
	Sort    string // column name
	Order   string // asc|desc
}
