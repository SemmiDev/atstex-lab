package domain

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionPlan struct {
	ID                uuid.UUID `db:"id"                  json:"id"`
	Name              string    `db:"name"                json:"name"`
	PriceIDR          int64     `db:"price_idr"           json:"priceIdr"`
	DurationMonths    int       `db:"duration_months"     json:"durationMonths"`
	MaxCVProfiles     int       `db:"max_cv_profiles"     json:"maxCvProfiles"`
	MaxCVReviews      int       `db:"max_cv_reviews"      json:"maxCvReviews"`
	MaxATSSimulations int       `db:"max_ats_simulations" json:"maxAtsSimulations"`
	MaxCoverLetters   int       `db:"max_cover_letters"   json:"maxCoverLetters"`
	IsActive          bool      `db:"is_active"           json:"isActive"`
	CreatedAt         time.Time `db:"created_at"          json:"createdAt"`
	UpdatedAt         time.Time `db:"updated_at"          json:"updatedAt"`

	// Joined/Computed fields
	ActiveUsersCount int `db:"active_users_count" json:"activeUsersCount"`
}

type UserSubscription struct {
	ID        uuid.UUID `db:"id"         json:"id"`
	UserID    uuid.UUID `db:"user_id"    json:"userId"`
	PlanID    uuid.UUID `db:"plan_id"    json:"planId"`
	StartDate time.Time `db:"start_date" json:"startDate"`
	EndDate   time.Time `db:"end_date"   json:"endDate"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`

	// Joined fields
	Plan *SubscriptionPlan `db:"-" json:"plan,omitempty"`
}
