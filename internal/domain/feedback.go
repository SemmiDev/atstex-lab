package domain

import (
	"time"

	"github.com/google/uuid"
)

// Feedback represents a user-submitted feedback or recommendation.
type Feedback struct {
	ID         uuid.UUID  `db:"id"          json:"id"`
	UserID     uuid.UUID  `db:"user_id"     json:"userId"`
	Subject    string     `db:"subject"     json:"subject"`
	Message    string     `db:"message"     json:"message"`
	AdminReply *string    `db:"admin_reply" json:"adminReply"`
	RepliedAt  *time.Time `db:"replied_at"  json:"repliedAt"`
	CreatedAt  time.Time  `db:"created_at"  json:"createdAt"`
	// Joined fields (only populated in admin listing)
	UserName    string `db:"user_name"    json:"userName,omitempty"`
	UserEmail   string `db:"user_email"   json:"userEmail,omitempty"`
	UserPicture string `db:"user_picture" json:"userPicture,omitempty"`
}

// FeedbackListParams controls pagination and search for feedback listing.
type FeedbackListParams struct {
	Page    int
	PerPage int
	Search  string
}
