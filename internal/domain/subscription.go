package domain

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionPlan struct {
	ID                uuid.UUID `json:"id" db:"id"`
	Name              string    `json:"name" db:"name"`
	PriceIDR          int64     `json:"priceIdr" db:"price_idr"`
	DurationMonths    int       `json:"durationMonths" db:"duration_months"`
	MaxCVProfiles     int       `json:"maxCvProfiles" db:"max_cv_profiles"`
	MaxCVReviews      int       `json:"maxCvReviews" db:"max_cv_reviews"`
	MaxATSSimulations int       `json:"maxAtsSimulations" db:"max_ats_simulations"`
	MaxCoverLetters   int       `json:"maxCoverLetters" db:"max_cover_letters"`
	IsActive          bool      `json:"isActive" db:"is_active"`
	CreatedAt         time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time `json:"updatedAt" db:"updated_at"`

	// Joined/Computed fields
	ActiveUsersCount int `json:"activeUsersCount" db:"active_users_count"`
}

type UserSubscription struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"userId" db:"user_id"`
	PlanID    uuid.UUID `json:"planId" db:"plan_id"`
	StartDate time.Time `json:"startDate" db:"start_date"`
	EndDate   time.Time `json:"endDate" db:"end_date"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`

	// Joined fields
	Plan *SubscriptionPlan `json:"plan,omitempty" db:"-"`
}
