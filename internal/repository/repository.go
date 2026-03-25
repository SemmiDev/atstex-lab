package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib" // blank import for pgx driver
	"github.com/jmoiron/sqlx"
	"github.com/semmidev/atstex-lab/internal/apperrors"
	"github.com/semmidev/atstex-lab/internal/domain"
)

var ErrNotFound = errors.New("record not found")

//nolint:interfacebloat // Repository defines the full data-access contract.
type Repository interface {
	UpsertUser(ctx context.Context, googleID, email, name, picture string) (*domain.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	SoftDeleteUser(ctx context.Context, id uuid.UUID) error
	CreateSession(ctx context.Context, userID uuid.UUID, token, ipAddress, userAgent string, expiresAt time.Time) error
	GetSession(ctx context.Context, token string) (*domain.Session, error)
	GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Session, error)
	DeleteSession(ctx context.Context, token string) error
	// CV Profile methods
	CreateCVProfile(ctx context.Context, userID uuid.UUID, title string) (*domain.CVProfile, error)
	GetCVProfile(ctx context.Context, id uuid.UUID) (*domain.CVProfile, error)
	GetCVProfilesByUserID(ctx context.Context, userID uuid.UUID) ([]domain.CVProfile, error)
	UpdateCVProfileBiodata(ctx context.Context, id uuid.UUID, biodata json.RawMessage) error
	UpdateCVProfileTitle(ctx context.Context, id uuid.UUID, title string) error
	DeleteCVProfile(ctx context.Context, id uuid.UUID) error
	// Public profile methods
	SetUsername(ctx context.Context, userID uuid.UUID, username string) error
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	CheckUsernameAvailable(ctx context.Context, username string) (bool, error)
	UpdateCVProfileVisibility(ctx context.Context, profileID uuid.UUID, isPublic bool) error
	GetPublicCVProfilesByUserID(ctx context.Context, userID uuid.UUID) ([]domain.CVProfile, error)
	// Custom Template Builder
	CreateCustomTemplate(ctx context.Context, t *domain.CustomTemplate) error
	GetCustomTemplate(ctx context.Context, id uuid.UUID) (*domain.CustomTemplate, error)
	GetCustomTemplatesByUserID(ctx context.Context, userID uuid.UUID) ([]domain.CustomTemplate, error)
	UpdateCustomTemplate(ctx context.Context, id uuid.UUID, config json.RawMessage) error
	DeleteCustomTemplate(ctx context.Context, id uuid.UUID) error
	// AI usage tracking
	IncrementAITokensUsed(ctx context.Context, userID uuid.UUID, chars int64) error
	// Feedback methods
	CreateFeedback(ctx context.Context, userID uuid.UUID, subject, message string) (*domain.Feedback, error)
	GetFeedbacksByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Feedback, error)
	AdminListFeedbacks(ctx context.Context, params domain.FeedbackListParams) ([]domain.Feedback, int, error)
	AdminReplyFeedback(ctx context.Context, feedbackID uuid.UUID, reply string) error
	AdminDeleteFeedback(ctx context.Context, feedbackID uuid.UUID) error
	// Admin user management
	AdminBlockUser(ctx context.Context, userID uuid.UUID) error
	AdminUnblockUser(ctx context.Context, userID uuid.UUID) error
	AdminDeleteUser(ctx context.Context, userID uuid.UUID) error
	AdminMakeUserAdmin(ctx context.Context, userID uuid.UUID) error
	AdminRevokeUserAdmin(ctx context.Context, userID uuid.UUID) error
	// Admin methods
	AdminGetStats(ctx context.Context) (*domain.AdminStats, error)
	AdminGetAnalytics(ctx context.Context) (*domain.AdminAnalytics, error)
	AdminListUsers(ctx context.Context, params domain.AdminListParams) ([]domain.AdminUserRow, int, error)
	// CV Review
	CreateCVReview(ctx context.Context, review *domain.CVReview) error
	GetCVReviewsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.CVReview, error)
	CountCVReviewsByDate(ctx context.Context, userID uuid.UUID, start, end time.Time) (int, error)
	// Cover Letters
	CreateCoverLetter(ctx context.Context, cl *domain.CoverLetter) error
	GetCoverLettersByUserID(ctx context.Context, userID uuid.UUID) ([]domain.CoverLetter, error)
	CountCoverLettersByDate(ctx context.Context, userID uuid.UUID, start, end time.Time) (int, error)
	// Job Application Tracking
	CreateJobApplication(ctx context.Context, j *domain.JobApplication) (*domain.JobApplication, error)
	UpdateJobApplication(ctx context.Context, app *domain.JobApplication) error
	GetJobApplicationsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.JobApplication, error)
	UpdateJobApplicationStatus(ctx context.Context, id uuid.UUID, status string) error
	DeleteJobApplication(ctx context.Context, id uuid.UUID) error

	// ATS Simmulator methods
	CreateAtsSimulation(ctx context.Context, sim *domain.AtsSimulation) error
	GetAtsSimulationsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.AtsSimulation, error)
	CountAtsSimulationsByDate(ctx context.Context, userID uuid.UUID, start, end time.Time) (int, error)

	// Interview Prep methods
	CreateInterviewPrep(ctx context.Context, prep *domain.InterviewPrep) error
	GetInterviewPrepsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.InterviewPrep, error)
	CountInterviewPrepsByDate(ctx context.Context, userID uuid.UUID, start, end time.Time) (int, error)

	// Mock Interview Session methods
	CreateMockInterviewSession(ctx context.Context, s *domain.MockInterviewSession) error
	UpdateMockInterviewSession(ctx context.Context, s *domain.MockInterviewSession) error
	GetMockInterviewSession(ctx context.Context, id uuid.UUID) (*domain.MockInterviewSession, error)
	GetMockInterviewSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.MockInterviewSession, error)
	EndMockInterviewSession(ctx context.Context, id uuid.UUID, tokensUsed int64, turnCount int, messages json.RawMessage) error
	CountMockInterviewSessionsByDate(ctx context.Context, userID uuid.UUID, start, end time.Time) (int, error)

	// Subscriptions
	AdminListSubscriptionPlans(ctx context.Context) ([]domain.SubscriptionPlan, error)
	AdminCreateSubscriptionPlan(ctx context.Context, plan *domain.SubscriptionPlan) error
	AdminUpdateSubscriptionPlan(ctx context.Context, plan *domain.SubscriptionPlan) error
	AdminToggleSubscriptionPlan(ctx context.Context, id uuid.UUID, isActive bool) error
	AdminDeleteSubscriptionPlan(ctx context.Context, id uuid.UUID) error
	AdminAssignSubscription(ctx context.Context, userID, planID uuid.UUID, months int) error
	GetUserActiveSubscription(ctx context.Context, userID uuid.UUID) (*domain.UserSubscription, error)
	GetUserSubscriptions(ctx context.Context, userID uuid.UUID) ([]domain.UserSubscription, error)
	GetFreeSubscriptionPlan(ctx context.Context) (*domain.SubscriptionPlan, error)

	// Analytics
	CreateProfileAnalyticsEvent(ctx context.Context, e *domain.ProfileAnalyticsEvent) error
	GetDashboardMetrics(ctx context.Context, userID uuid.UUID) (*domain.DashboardMetrics, error)

	Close() error
}

type postgresRepo struct {
	db *sqlx.DB
}

func Connect(ctx context.Context, dsn string) (Repository, error) {
	// Beri batas waktu maksimal untuk seluruh proses connect
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var (
		db  *sqlx.DB
		err error
	)

	for i := 0; i < 10; i++ {
		// Cek dulu apakah context sudah cancelled/timeout
		if ctx.Err() != nil {
			return nil, fmt.Errorf("connect db cancelled: %w", ctx.Err())
		}

		db, err = sqlx.ConnectContext(ctx, "pgx", dsn)
		if err == nil {
			return &postgresRepo{db: db}, nil
		}

		// Sleep dengan respect terhadap context
		select {
		case <-time.After(time.Second):
			// lanjut retry
		case <-ctx.Done():
			return nil, fmt.Errorf("connect db timeout after %d attempts: %w", i+1, ctx.Err())
		}
	}

	return nil, fmt.Errorf("connect db failed after 10 attempts: %w", err)
}

func (r *postgresRepo) Close() error {
	return r.db.Close()
}

// ── Profile Analytics ──────────────────────────────────────────

func (r *postgresRepo) CreateProfileAnalyticsEvent(ctx context.Context, e *domain.ProfileAnalyticsEvent) error {
	query := `INSERT INTO profile_analytics (user_id, profile_id, action, ip_address, user_agent, referer)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`
	return r.db.QueryRowxContext(ctx, query,
		e.UserID, e.ProfileID, e.Action, e.IPAddress, e.UserAgent, e.Referer,
	).Scan(&e.ID, &e.CreatedAt)
}

func (r *postgresRepo) GetDashboardMetrics(ctx context.Context, userID uuid.UUID) (*domain.DashboardMetrics, error) {
	metrics := &domain.DashboardMetrics{
		ViewsLast30Days: []domain.DailyAnalytic{},
		TopReferers:     []domain.RefererCount{},
	}

	err := r.db.GetContext(ctx, &metrics.TotalViews, `SELECT COUNT(*) FROM profile_analytics WHERE user_id = $1 AND action = 'view'`, userID)
	if err != nil {
		return nil, translatePgError(err, "total views", nil)
	}

	err = r.db.GetContext(ctx, &metrics.TotalDownloads, `SELECT COUNT(*) FROM profile_analytics WHERE user_id = $1 AND action = 'download_pdf'`, userID)
	if err != nil {
		return nil, translatePgError(err, "total downloads", nil)
	}

	queryDaily := `
		SELECT d::date::text AS date, COALESCE(c.cnt, 0) AS count
		FROM generate_series(
			(CURRENT_DATE - INTERVAL '29 days'),
			CURRENT_DATE,
			'1 day'
		) AS d
		LEFT JOIN (
			SELECT DATE_TRUNC('day', created_at)::date AS day, COUNT(*) AS cnt
			FROM profile_analytics
			WHERE user_id = $1 AND action = 'view' AND created_at >= CURRENT_DATE - INTERVAL '29 days'
			GROUP BY day
		) c ON c.day = d::date
		ORDER BY d
	`
	err = r.db.SelectContext(ctx, &metrics.ViewsLast30Days, queryDaily, userID)
	if err != nil {
		return nil, translatePgError(err, "daily views", nil)
	}

	queryReferers := `
		SELECT COALESCE(referer, 'Direct') AS referer, COUNT(*) as count
		FROM profile_analytics
		WHERE user_id = $1 AND action = 'view'
		GROUP BY referer
		ORDER BY count DESC
		LIMIT 5
	`
	err = r.db.SelectContext(ctx, &metrics.TopReferers, queryReferers, userID)
	if err != nil {
		return nil, translatePgError(err, "top referers", nil)
	}

	return metrics, nil
}

const userColumns = `id, google_id, email, name, picture, role, ai_tokens_used, username, is_blocked, created_at, updated_at`

func (r *postgresRepo) UpsertUser(ctx context.Context, googleID, email, name, picture string) (*domain.User, error) {
	query := `
		INSERT INTO users (google_id, email, name, picture)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (google_id) DO UPDATE SET
			email = EXCLUDED.email,
			name = EXCLUDED.name,
			picture = EXCLUDED.picture,
			updated_at = CURRENT_TIMESTAMP
		RETURNING ` + userColumns + `
	`
	var u domain.User
	err := r.db.GetContext(ctx, &u, query, googleID, email, name, picture)
	if err != nil {
		return &u, translatePgError(err, "record", nil)
	}
	return &u, nil
}

func (r *postgresRepo) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE id = $1`
	var u domain.User
	err := r.db.GetContext(ctx, &u, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return &u, translatePgError(err, "record", nil)
	}
	return &u, nil
}

func (r *postgresRepo) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	return translatePgError(err, "record", nil)
}

func (r *postgresRepo) SoftDeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET is_blocked = true, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, id)
	return translatePgError(err, "record", nil)
}

func (r *postgresRepo) CreateSession(ctx context.Context, userID uuid.UUID, token, ipAddress, userAgent string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (user_id, token, ip_address, user_agent, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, token, ipAddress, userAgent, expiresAt)
	return translatePgError(err, "record", nil)
}

func (r *postgresRepo) GetSession(ctx context.Context, token string) (*domain.Session, error) {
	query := `SELECT id, user_id, token, ip_address, user_agent, expires_at, created_at FROM sessions WHERE token = $1`
	var sess domain.Session
	err := r.db.GetContext(ctx, &sess, query, token)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return &sess, translatePgError(err, "record", nil)
	}
	return &sess, nil
}

func (r *postgresRepo) GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Session, error) {
	query := `SELECT id, user_id, token, COALESCE(ip_address, '') as ip_address, COALESCE(user_agent, '') as user_agent, expires_at, created_at FROM sessions WHERE user_id = $1 ORDER BY expires_at DESC`
	var sessions []domain.Session
	err := r.db.SelectContext(ctx, &sessions, query, userID)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *postgresRepo) DeleteSession(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = $1`, token)
	return translatePgError(err, "record", nil)
}

// ── CV Profile CRUD ────────────────────────────────────────────

const cvProfileColumns = `id, user_id, title, biodata, is_public, created_at, updated_at`

func (r *postgresRepo) CreateCVProfile(ctx context.Context, userID uuid.UUID, title string) (*domain.CVProfile, error) {
	query := `INSERT INTO cv_profiles (user_id, title) VALUES ($1, $2) RETURNING ` + cvProfileColumns
	var p domain.CVProfile
	err := r.db.GetContext(ctx, &p, query, userID, title)
	if err != nil {
		return &p, translatePgError(err, "record", nil)
	}
	return &p, nil
}

func (r *postgresRepo) GetCVProfile(ctx context.Context, id uuid.UUID) (*domain.CVProfile, error) {
	query := `SELECT ` + cvProfileColumns + ` FROM cv_profiles WHERE id = $1`
	var p domain.CVProfile
	err := r.db.GetContext(ctx, &p, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return &p, translatePgError(err, "record", nil)
	}
	return &p, nil
}

func (r *postgresRepo) GetCVProfilesByUserID(ctx context.Context, userID uuid.UUID) ([]domain.CVProfile, error) {
	query := `SELECT ` + cvProfileColumns + ` FROM cv_profiles WHERE user_id = $1 ORDER BY created_at ASC`
	var profiles []domain.CVProfile
	err := r.db.SelectContext(ctx, &profiles, query, userID)
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

func (r *postgresRepo) UpdateCVProfileBiodata(ctx context.Context, id uuid.UUID, biodata json.RawMessage) error {
	_, err := r.db.ExecContext(ctx, `UPDATE cv_profiles SET biodata = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, biodata, id)
	return translatePgError(err, "record", nil)
}

func (r *postgresRepo) UpdateCVProfileTitle(ctx context.Context, id uuid.UUID, title string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE cv_profiles SET title = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, title, id)
	return translatePgError(err, "record", nil)
}

func (r *postgresRepo) DeleteCVProfile(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM cv_profiles WHERE id = $1`, id)
	return translatePgError(err, "record", nil)
}

// ── Custom Template CRUD ───────────────────────────────────────

func (r *postgresRepo) CreateCustomTemplate(ctx context.Context, t *domain.CustomTemplate) error {
	query := `INSERT INTO custom_templates (user_id, name, config) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	err := r.db.QueryRowxContext(ctx, query, t.UserID, t.Name, t.Config).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	return translatePgError(err, "record", nil)
}

func (r *postgresRepo) GetCustomTemplate(ctx context.Context, id uuid.UUID) (*domain.CustomTemplate, error) {
	query := `SELECT id, user_id, name, config, created_at, updated_at FROM custom_templates WHERE id = $1`
	var t domain.CustomTemplate
	err := r.db.GetContext(ctx, &t, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, translatePgError(err, "record", nil)
	}
	return &t, nil
}

func (r *postgresRepo) GetCustomTemplatesByUserID(ctx context.Context, userID uuid.UUID) ([]domain.CustomTemplate, error) {
	query := `SELECT id, user_id, name, config, created_at, updated_at FROM custom_templates WHERE user_id = $1 ORDER BY created_at DESC`
	var templates []domain.CustomTemplate
	err := r.db.SelectContext(ctx, &templates, query, userID)
	if err != nil {
		return nil, err
	}
	if templates == nil {
		templates = []domain.CustomTemplate{}
	}
	return templates, nil
}

func (r *postgresRepo) UpdateCustomTemplate(ctx context.Context, id uuid.UUID, config json.RawMessage) error {
	_, err := r.db.ExecContext(ctx, `UPDATE custom_templates SET config = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, config, id)
	return translatePgError(err, "record", nil)
}

func (r *postgresRepo) DeleteCustomTemplate(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM custom_templates WHERE id = $1`, id)
	return translatePgError(err, "record", nil)
}

// ── Public Profile Methods ─────────────────────────────────────

func (r *postgresRepo) SetUsername(ctx context.Context, userID uuid.UUID, username string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET username = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, username, userID)
	return translatePgError(err, "record", nil)
}

func (r *postgresRepo) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE LOWER(username) = LOWER($1)`
	var u domain.User
	err := r.db.GetContext(ctx, &u, query, username)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return &u, translatePgError(err, "record", nil)
	}
	return &u, nil
}

func (r *postgresRepo) CheckUsernameAvailable(ctx context.Context, username string) (bool, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM users WHERE LOWER(username) = LOWER($1)`, username)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (r *postgresRepo) UpdateCVProfileVisibility(ctx context.Context, profileID uuid.UUID, isPublic bool) error {
	_, err := r.db.ExecContext(ctx, `UPDATE cv_profiles SET is_public = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, isPublic, profileID)
	return translatePgError(err, "record", nil)
}

func (r *postgresRepo) GetPublicCVProfilesByUserID(ctx context.Context, userID uuid.UUID) ([]domain.CVProfile, error) {
	query := `SELECT ` + cvProfileColumns + ` FROM cv_profiles WHERE user_id = $1 AND is_public = true ORDER BY created_at ASC`
	var profiles []domain.CVProfile
	err := r.db.SelectContext(ctx, &profiles, query, userID)
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

// ── AI Usage Tracking ──────────────────────────────────────────

func (r *postgresRepo) IncrementAITokensUsed(ctx context.Context, userID uuid.UUID, chars int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET ai_tokens_used = ai_tokens_used + $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, userID, chars)
	return translatePgError(err, "record", nil)
}

// ── Admin Queries ──────────────────────────────────────────────

func (r *postgresRepo) AdminGetStats(ctx context.Context) (*domain.AdminStats, error) {
	var stats domain.AdminStats
	err := r.db.GetContext(ctx, &stats, `
		SELECT
			(SELECT COUNT(*) FROM users) AS total_users,
			(SELECT COUNT(*) FROM users WHERE role = 'admin') AS total_admins,
			(SELECT COALESCE(SUM(ai_tokens_used), 0) FROM users) AS total_ai_tokens,
			(SELECT COUNT(*) FROM cv_profiles) AS total_biodata,
			(SELECT COUNT(*) FROM sessions WHERE expires_at > NOW()) AS total_sessions,
			(SELECT COUNT(*) FROM cv_reviews) AS total_cv_reviews,
			(SELECT COUNT(*) FROM cover_letters) AS total_cover_letters,
			(SELECT COUNT(*) FROM ats_simulations) AS total_ats_simulations,
			(SELECT COUNT(*) FROM job_applications) AS total_job_apps
	`)
	if err != nil {
		return &stats, translatePgError(err, "record", nil)
	}
	return &stats, nil
}

func (r *postgresRepo) AdminGetAnalytics(ctx context.Context) (*domain.AdminAnalytics, error) {
	analytics := &domain.AdminAnalytics{}

	// Helper to query a daily count series for the last 30 days.
	queryDaily := func(table string) ([]domain.AdminDailyCount, error) {
		var rows []domain.AdminDailyCount
		q := fmt.Sprintf(`
			SELECT d::date::text AS date, COALESCE(c.cnt, 0) AS count
			FROM generate_series(
				(CURRENT_DATE - INTERVAL '29 days'),
				CURRENT_DATE,
				'1 day'
			) AS d
			LEFT JOIN (
				SELECT DATE_TRUNC('day', created_at)::date AS day, COUNT(*) AS cnt
				FROM %s
				WHERE created_at >= CURRENT_DATE - INTERVAL '29 days'
				GROUP BY day
			) c ON c.day = d::date
			ORDER BY d
		`, table)
		err := r.db.SelectContext(ctx, &rows, q)
		if rows == nil {
			rows = []domain.AdminDailyCount{}
		}
		return rows, err
	}

	var err error
	analytics.UserRegistrations, err = queryDaily("users")
	if err != nil {
		return nil, fmt.Errorf("user registrations: %w", err)
	}
	analytics.CVReviews, err = queryDaily("cv_reviews")
	if err != nil {
		return nil, fmt.Errorf("cv reviews: %w", err)
	}
	analytics.CoverLetters, err = queryDaily("cover_letters")
	if err != nil {
		return nil, fmt.Errorf("cover letters: %w", err)
	}
	analytics.ATSSimulations, err = queryDaily("ats_simulations")
	if err != nil {
		return nil, fmt.Errorf("ats simulations: %w", err)
	}

	return analytics, nil
}

var allowedSortColumns = map[string]string{
	"name":           "u.name",
	"email":          "u.email",
	"role":           "u.role",
	"ai_tokens_used": "u.ai_tokens_used",
	"biodata_count":  "biodata_count",
	"created_at":     "u.created_at",
}

func (r *postgresRepo) AdminListUsers(ctx context.Context, params domain.AdminListParams) ([]domain.AdminUserRow, int, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 || params.PerPage > 100 {
		params.PerPage = 20
	}

	sortCol, ok := allowedSortColumns[params.Sort]
	if !ok {
		sortCol = "u.created_at"
	}
	orderDir := "DESC"
	if params.Order == "asc" {
		orderDir = "ASC"
	}

	offset := (params.Page - 1) * params.PerPage

	// Count total
	countQuery := `SELECT COUNT(*) FROM users u WHERE ($1 = '' OR u.name ILIKE '%' || $1 || '%' OR u.email ILIKE '%' || $1 || '%')`
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, params.Search); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT
			u.id, u.email, u.name, COALESCE(u.picture, '') AS picture, u.role, u.username, u.is_blocked, u.ai_tokens_used, u.created_at,
			COUNT(cv.id) AS biodata_count
		FROM users u
		LEFT JOIN cv_profiles cv ON cv.user_id = u.id
		WHERE ($1 = '' OR u.name ILIKE '%%' || $1 || '%%' OR u.email ILIKE '%%' || $1 || '%%')
		GROUP BY u.id
		ORDER BY %s %s
		LIMIT $2 OFFSET $3
	`, sortCol, orderDir)

	var rows []domain.AdminUserRow
	if err := r.db.SelectContext(ctx, &rows, query, params.Search, params.PerPage, offset); err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

// ── Feedback CRUD ──────────────────────────────────────────────

func (r *postgresRepo) CreateFeedback(ctx context.Context, userID uuid.UUID, subject, message string) (*domain.Feedback, error) {
	query := `INSERT INTO feedbacks (user_id, subject, message) VALUES ($1, $2, $3) RETURNING id, user_id, subject, message, admin_reply, replied_at, created_at`
	var f domain.Feedback
	err := r.db.GetContext(ctx, &f, query, userID, subject, message)
	if err != nil {
		return &f, translatePgError(err, "record", nil)
	}
	return &f, nil
}

func (r *postgresRepo) GetFeedbacksByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Feedback, error) {
	query := `SELECT id, user_id, subject, message, admin_reply, replied_at, created_at FROM feedbacks WHERE user_id = $1 ORDER BY created_at DESC`
	var feedbacks []domain.Feedback
	err := r.db.SelectContext(ctx, &feedbacks, query, userID)
	if err != nil {
		return nil, err
	}
	return feedbacks, nil
}

func (r *postgresRepo) AdminListFeedbacks(ctx context.Context, params domain.FeedbackListParams) ([]domain.Feedback, int, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 || params.PerPage > 100 {
		params.PerPage = 20
	}

	offset := (params.Page - 1) * params.PerPage

	countQuery := `
		SELECT COUNT(*) FROM feedbacks f
		JOIN users u ON u.id = f.user_id
		WHERE ($1 = '' OR f.subject ILIKE '%' || $1 || '%' OR f.message ILIKE '%' || $1 || '%' OR u.name ILIKE '%' || $1 || '%' OR u.email ILIKE '%' || $1 || '%')
	`
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, params.Search); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT
			f.id, f.user_id, f.subject, f.message, f.admin_reply, f.replied_at, f.created_at,
			u.name AS user_name, u.email AS user_email, COALESCE(u.picture, '') AS user_picture
		FROM feedbacks f
		JOIN users u ON u.id = f.user_id
		WHERE ($1 = '' OR f.subject ILIKE '%' || $1 || '%' OR f.message ILIKE '%' || $1 || '%' OR u.name ILIKE '%' || $1 || '%' OR u.email ILIKE '%' || $1 || '%')
		ORDER BY f.created_at DESC
		LIMIT $2 OFFSET $3
	`

	var rows []domain.Feedback
	if err := r.db.SelectContext(ctx, &rows, query, params.Search, params.PerPage, offset); err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

func (r *postgresRepo) AdminReplyFeedback(ctx context.Context, feedbackID uuid.UUID, reply string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE feedbacks SET admin_reply = $1, replied_at = CURRENT_TIMESTAMP WHERE id = $2`, reply, feedbackID)
	return translatePgError(err, "record", nil)
}

func (r *postgresRepo) AdminDeleteFeedback(ctx context.Context, feedbackID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM feedbacks WHERE id = $1`, feedbackID)
	return translatePgError(err, "record", nil)
}

// ── Admin User Management ──────────────────────────────────────

func (r *postgresRepo) AdminBlockUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET is_blocked = true, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, userID)
	return translatePgError(err, "record", nil)
}

func (r *postgresRepo) AdminUnblockUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET is_blocked = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, userID)
	return translatePgError(err, "record", nil)
}

func (r *postgresRepo) AdminMakeUserAdmin(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET role = 'admin', updated_at = CURRENT_TIMESTAMP WHERE id = $1`, userID)
	return translatePgError(err, "record", nil)
}

func (r *postgresRepo) AdminRevokeUserAdmin(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET role = 'user', updated_at = CURRENT_TIMESTAMP WHERE id = $1`, userID)
	return translatePgError(err, "record", nil)
}

func (r *postgresRepo) AdminDeleteUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, userID)
	return translatePgError(err, "record", nil)
}

// ── CV Reviews ────────────────────────────────────────────────

func (r *postgresRepo) CreateCVReview(ctx context.Context, review *domain.CVReview) error {
	query := `INSERT INTO cv_reviews (user_id, profile_id, profile_title, language, score, strengths, improvements, recommendations, tokens_used)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at`
	return r.db.QueryRowxContext(ctx, query,
		review.UserID, review.ProfileID, review.ProfileTitle, review.Language,
		review.Score, review.Strengths, review.Improvements, review.Recommendations, review.TokensUsed,
	).Scan(&review.ID, &review.CreatedAt)
}

func (r *postgresRepo) GetCVReviewsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.CVReview, error) {
	var reviews []domain.CVReview
	err := r.db.SelectContext(ctx, &reviews,
		`SELECT id, user_id, profile_id, profile_title, language, score, strengths, improvements, recommendations, tokens_used, created_at
		 FROM cv_reviews WHERE user_id = $1 ORDER BY created_at DESC LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	if reviews == nil {
		reviews = []domain.CVReview{}
	}
	return reviews, nil
}

func (r *postgresRepo) CountCVReviewsByDate(ctx context.Context, userID uuid.UUID, start, end time.Time) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM cv_reviews WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3`,
		userID, start, end)
	if err != nil {
		return count, translatePgError(err, "record", nil)
	}
	return count, nil
}

// ── Cover Letters ──────────────────────────────────────────────

func (r *postgresRepo) CreateCoverLetter(ctx context.Context, cl *domain.CoverLetter) error {
	query := `INSERT INTO cover_letters (user_id, profile_id, profile_title, job_description, cover_letter_text, language, tokens_used)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`
	return r.db.QueryRowxContext(ctx, query,
		cl.UserID, cl.ProfileID, cl.ProfileTitle, cl.JobDescription, cl.CoverLetterText, cl.Language, cl.TokensUsed,
	).Scan(&cl.ID, &cl.CreatedAt)
}

func (r *postgresRepo) GetCoverLettersByUserID(ctx context.Context, userID uuid.UUID) ([]domain.CoverLetter, error) {
	var letters []domain.CoverLetter
	err := r.db.SelectContext(ctx, &letters,
		`SELECT id, user_id, profile_id, profile_title, job_description, cover_letter_text, language, tokens_used, created_at
		 FROM cover_letters WHERE user_id = $1 ORDER BY created_at DESC LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	if letters == nil {
		letters = []domain.CoverLetter{}
	}
	return letters, nil
}

func (r *postgresRepo) CountCoverLettersByDate(ctx context.Context, userID uuid.UUID, start, end time.Time) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM cover_letters WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3`,
		userID, start, end)
	if err != nil {
		return count, translatePgError(err, "record", nil)
	}
	return count, nil
}

// ── Job Application Tracking ──────────────────────────────────

func (r *postgresRepo) CreateJobApplication(ctx context.Context, j *domain.JobApplication) (*domain.JobApplication, error) {
	query := `INSERT INTO job_applications (user_id, cv_profile_id, company, job_title, status, notes, deadline)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, cv_profile_id, company, job_title, status, notes, deadline, created_at, updated_at`

	var cvProfileID interface{}
	if j.CVProfileID != nil {
		cvProfileID = *j.CVProfileID
	}
	err := r.db.GetContext(ctx, j, query, j.UserID, cvProfileID, j.Company, j.JobTitle, j.Status, j.Notes, j.Deadline)
	if err != nil {
		return j, translatePgError(err, "record", nil)
	}
	return j, nil
}

func (r *postgresRepo) GetJobApplicationsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.JobApplication, error) {
	var apps []domain.JobApplication
	err := r.db.SelectContext(ctx, &apps,
		`SELECT id, user_id, cv_profile_id, company, job_title, status, notes, deadline, created_at, updated_at
		 FROM job_applications WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	if apps == nil {
		apps = []domain.JobApplication{}
	}
	return apps, nil
}

func (r *postgresRepo) CountAtsSimulationsByDate(ctx context.Context, userID uuid.UUID, start, end time.Time) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM ats_simulations WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3`,
		userID, start, end)
	if err != nil {
		return count, translatePgError(err, "record", nil)
	}
	return count, nil
}

func (r *postgresRepo) UpdateJobApplicationStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE job_applications SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, status, id)
	return translatePgError(err, "record", nil)
}

func (r *postgresRepo) UpdateJobApplication(ctx context.Context, app *domain.JobApplication) error {
	var cvProfileID interface{}
	if app.CVProfileID != nil {
		cvProfileID = *app.CVProfileID
	}

	query := `UPDATE job_applications
			  SET company = $1, job_title = $2, cv_profile_id = $3, notes = $4, deadline = $5, updated_at = CURRENT_TIMESTAMP
			  WHERE id = $6 AND user_id = $7`

	_, err := r.db.ExecContext(ctx, query, app.Company, app.JobTitle, cvProfileID, app.Notes, app.Deadline, app.ID, app.UserID)
	return translatePgError(err, "record", nil)
}

func (r *postgresRepo) DeleteJobApplication(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM job_applications WHERE id = $1`, id)
	return translatePgError(err, "record", nil)
}

// ── Subscriptions ──────────────────────────────────────────────

func (r *postgresRepo) AdminListSubscriptionPlans(ctx context.Context) ([]domain.SubscriptionPlan, error) {
	var plans []domain.SubscriptionPlan
	query := `
		SELECT
			sp.id, sp.name, sp.price_idr, sp.duration_months,
			sp.max_cv_profiles, sp.max_cv_reviews, sp.max_ats_simulations, sp.max_cover_letters,
			sp.is_active, sp.created_at, sp.updated_at,
			COUNT(us.id) AS active_users_count
		FROM subscription_plans sp
		LEFT JOIN user_subscriptions us ON sp.id = us.plan_id AND us.end_date > CURRENT_TIMESTAMP
		GROUP BY sp.id
		ORDER BY sp.price_idr ASC
	`
	if err := r.db.SelectContext(ctx, &plans, query); err != nil {
		return nil, err
	}
	if plans == nil {
		plans = []domain.SubscriptionPlan{}
	}
	return plans, nil
}

func (r *postgresRepo) GetFreeSubscriptionPlan(ctx context.Context) (*domain.SubscriptionPlan, error) {
	query := `SELECT id, name, price_idr, duration_months, max_cv_profiles, max_cv_reviews, max_ats_simulations, max_cover_letters, is_active, created_at, updated_at
		FROM subscription_plans WHERE price_idr = 0 AND is_active = true LIMIT 1`
	var plan domain.SubscriptionPlan
	err := r.db.GetContext(ctx, &plan, query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return &plan, translatePgError(err, "record", nil)
	}
	return &plan, nil
}

func (r *postgresRepo) AdminCreateSubscriptionPlan(ctx context.Context, plan *domain.SubscriptionPlan) error {
	query := `INSERT INTO subscription_plans (name, price_idr, duration_months, max_cv_profiles, max_cv_reviews, max_ats_simulations, max_cover_letters, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`
	return r.db.QueryRowxContext(ctx, query,
		plan.Name, plan.PriceIDR, plan.DurationMonths, plan.MaxCVProfiles, plan.MaxCVReviews, plan.MaxATSSimulations, plan.MaxCoverLetters, plan.IsActive,
	).Scan(&plan.ID, &plan.CreatedAt, &plan.UpdatedAt)
}

func (r *postgresRepo) AdminUpdateSubscriptionPlan(ctx context.Context, plan *domain.SubscriptionPlan) error {
	query := `UPDATE subscription_plans
		SET name = $1, price_idr = $2, duration_months = $3, max_cv_profiles = $4, max_cv_reviews = $5, max_ats_simulations = $6, max_cover_letters = $7, updated_at = CURRENT_TIMESTAMP
		WHERE id = $8`
	_, err := r.db.ExecContext(ctx, query,
		plan.Name, plan.PriceIDR, plan.DurationMonths, plan.MaxCVProfiles, plan.MaxCVReviews, plan.MaxATSSimulations, plan.MaxCoverLetters, plan.ID,
	)
	return translatePgError(err, "record", nil)
}

func (r *postgresRepo) AdminToggleSubscriptionPlan(ctx context.Context, id uuid.UUID, isActive bool) error {
	_, err := r.db.ExecContext(ctx, `UPDATE subscription_plans SET is_active = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, isActive, id)
	return translatePgError(err, "record", nil)
}

func (r *postgresRepo) AdminDeleteSubscriptionPlan(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM subscription_plans WHERE id = $1`, id)
	return translatePgError(err, "record", nil)
}

func (r *postgresRepo) AdminAssignSubscription(ctx context.Context, userID, planID uuid.UUID, months int) error {
	query := `INSERT INTO user_subscriptions (user_id, plan_id, start_date, end_date)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + ($3 * interval '1 month'))`
	_, err := r.db.ExecContext(ctx, query, userID, planID, months)
	return translatePgError(err, "record", nil)
}

func (r *postgresRepo) GetUserActiveSubscription(ctx context.Context, userID uuid.UUID) (*domain.UserSubscription, error) {
	query := `
		SELECT
			us.id, us.user_id, us.plan_id, us.start_date, us.end_date, us.created_at,
			sp.id AS "plan.id", sp.name AS "plan.name", sp.price_idr AS "plan.price_idr", sp.duration_months AS "plan.duration_months",
			sp.max_cv_profiles AS "plan.max_cv_profiles", sp.max_cv_reviews AS "plan.max_cv_reviews",
			sp.max_ats_simulations AS "plan.max_ats_simulations", sp.max_cover_letters AS "plan.max_cover_letters",
			sp.is_active AS "plan.is_active"
		FROM user_subscriptions us
		JOIN subscription_plans sp ON sp.id = us.plan_id
		WHERE us.user_id = $1 AND us.end_date > CURRENT_TIMESTAMP
		ORDER BY us.end_date DESC
		LIMIT 1
	`
	var us domain.UserSubscription
	var sp domain.SubscriptionPlan
	row := r.db.QueryRowxContext(ctx, query, userID)
	err := row.Scan(
		&us.ID, &us.UserID, &us.PlanID, &us.StartDate, &us.EndDate, &us.CreatedAt,
		&sp.ID, &sp.Name, &sp.PriceIDR, &sp.DurationMonths,
		&sp.MaxCVProfiles, &sp.MaxCVReviews, &sp.MaxATSSimulations, &sp.MaxCoverLetters, &sp.IsActive,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	us.Plan = &sp
	return &us, nil
}

func (r *postgresRepo) GetUserSubscriptions(ctx context.Context, userID uuid.UUID) ([]domain.UserSubscription, error) {
	query := `
		SELECT
			us.id, us.user_id, us.plan_id, us.start_date, us.end_date, us.created_at,
			sp.id AS "plan.id", sp.name AS "plan.name", sp.price_idr AS "plan.price_idr", sp.duration_months AS "plan.duration_months",
			sp.max_cv_profiles AS "plan.max_cv_profiles", sp.max_cv_reviews AS "plan.max_cv_reviews",
			sp.max_ats_simulations AS "plan.max_ats_simulations", sp.max_cover_letters AS "plan.max_cover_letters",
			sp.is_active AS "plan.is_active"
		FROM user_subscriptions us
		JOIN subscription_plans sp ON sp.id = us.plan_id
		WHERE us.user_id = $1
		ORDER BY us.created_at DESC
	`
	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []domain.UserSubscription
	for rows.Next() {
		var us domain.UserSubscription
		var sp domain.SubscriptionPlan
		err := rows.Scan(
			&us.ID, &us.UserID, &us.PlanID, &us.StartDate, &us.EndDate, &us.CreatedAt,
			&sp.ID, &sp.Name, &sp.PriceIDR, &sp.DurationMonths,
			&sp.MaxCVProfiles, &sp.MaxCVReviews, &sp.MaxATSSimulations, &sp.MaxCoverLetters, &sp.IsActive,
		)
		if err != nil {
			return nil, err
		}
		us.Plan = &sp
		subscriptions = append(subscriptions, us)
	}
	return subscriptions, nil
}

// translatePgError is the subsystem boundary firewall that maps Postgres errors and sql.ErrNoRows to safe domain-level SafeError.
//
//nolint:unparam // resource string is intentionally identical for generic repository wrappers
func translatePgError(err error, resource string, meta map[string]string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return apperrors.NewNotFound(resource, err)
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return apperrors.NewConflict(resource, err, meta)
		case "23503": // foreign_key_violation
			return apperrors.NewInvalidInput("referenced " + resource + " does not exist")
		case "23502": // not_null_violation
			return apperrors.NewInvalidInput("missing required field")
		}
	}
	return apperrors.NewInternal(fmt.Errorf("%s repo: %w", resource, err))
}
