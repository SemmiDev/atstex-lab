package domain

import (
	"time"

	"github.com/google/uuid"
)

// Feedback represents a user-submitted feedback or recommendation.
type Feedback struct {
	ID          uuid.UUID  `json:"id"          db:"id"`
	UserID      uuid.UUID  `json:"userId"      db:"user_id"`
	Subject     string     `json:"subject"     db:"subject"`
	Message     string     `json:"message"     db:"message"`
	AdminReply  *string    `json:"adminReply"  db:"admin_reply"`
	RepliedAt   *time.Time `json:"repliedAt"   db:"replied_at"`
	CreatedAt   time.Time  `json:"createdAt"   db:"created_at"`
	// Joined fields (only populated in admin listing)
	UserName    string `json:"userName,omitempty"    db:"user_name"`
	UserEmail   string `json:"userEmail,omitempty"   db:"user_email"`
	UserPicture string `json:"userPicture,omitempty" db:"user_picture"`
}

// FeedbackListParams controls pagination and search for feedback listing.
type FeedbackListParams struct {
	Page    int
	PerPage int
	Search  string
}
