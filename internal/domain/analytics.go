package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProfileAnalyticsEvent struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	ProfileID uuid.UUID `json:"profile_id" db:"profile_id"`
	Action    string    `json:"action" db:"action"` // "view" or "download_pdf"
	IPAddress *string   `json:"ip_address" db:"ip_address"`
	UserAgent *string   `json:"user_agent" db:"user_agent"`
	Referer   *string   `json:"referer" db:"referer"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type DashboardMetrics struct {
	TotalViews     int            `json:"total_views"`
	TotalDownloads int            `json:"total_downloads"`
	ViewsLast30Days []DailyAnalytic `json:"views_last_30_days"`
	TopReferers    []RefererCount `json:"top_referers"`
}

type DailyAnalytic struct {
	Date  string `json:"date" db:"date"`
	Count int    `json:"count" db:"count"`
}

type RefererCount struct {
	Referer string `json:"referer" db:"referer"`
	Count   int    `json:"count" db:"count"`
}
