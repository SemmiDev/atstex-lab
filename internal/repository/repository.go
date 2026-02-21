package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/semmidev/atstex-lab/internal/domain"
)

var ErrNotFound = errors.New("record not found")

// Repository defines the data-access contract.
type Repository interface {
	UpsertUser(ctx context.Context, googleID, email, name, picture string) (*domain.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	CreateSession(ctx context.Context, userID uuid.UUID, token, ipAddress, userAgent string, expiresAt time.Time) error
	GetSession(ctx context.Context, token string) (*domain.Session, error)
	GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Session, error)
	DeleteSession(ctx context.Context, token string) error
	// CV Profile methods
	CreateCVProfile(ctx context.Context, userID uuid.UUID, title string) (*domain.CVProfile, error)
	GetCVProfile(ctx context.Context, id uuid.UUID) (*domain.CVProfile, error)
	GetCVProfilesByUserID(ctx context.Context, userID uuid.UUID) ([]domain.CVProfile, error)
	UpdateCVProfileBiodata(ctx context.Context, id uuid.UUID, biodata json.RawMessage) error
	DeleteCVProfile(ctx context.Context, id uuid.UUID) error
	// AI usage tracking
	IncrementAITokensUsed(ctx context.Context, userID uuid.UUID, chars int64) error
	// Admin methods
	AdminGetStats(ctx context.Context) (*domain.AdminStats, error)
	AdminListUsers(ctx context.Context, params domain.AdminListParams) ([]domain.AdminUserRow, int, error)
	Close() error
}

type postgresRepo struct {
	db *sqlx.DB
}

func Connect(dsn string) (Repository, error) {
	var db *sqlx.DB
	var err error

	for i := 0; i < 10; i++ {
		db, err = sqlx.Connect("pgx", dsn)
		if err == nil {
			return &postgresRepo{db: db}, nil
		}
		time.Sleep(time.Second)
	}

	return nil, err
}

func (r *postgresRepo) Close() error {
	return r.db.Close()
}

const userColumns = `id, google_id, email, name, picture, role, ai_tokens_used, created_at, updated_at`

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
	return &u, err
}

func (r *postgresRepo) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE id = $1`
	var u domain.User
	err := r.db.GetContext(ctx, &u, query, id)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return &u, err
}

func (r *postgresRepo) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

func (r *postgresRepo) CreateSession(ctx context.Context, userID uuid.UUID, token, ipAddress, userAgent string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (user_id, token, ip_address, user_agent, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, token, ipAddress, userAgent, expiresAt)
	return err
}

func (r *postgresRepo) GetSession(ctx context.Context, token string) (*domain.Session, error) {
	query := `SELECT id, user_id, token, ip_address, user_agent, expires_at, created_at FROM sessions WHERE token = $1`
	var sess domain.Session
	err := r.db.GetContext(ctx, &sess, query, token)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return &sess, err
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
	return err
}

// ── CV Profile CRUD ────────────────────────────────────────────

func (r *postgresRepo) CreateCVProfile(ctx context.Context, userID uuid.UUID, title string) (*domain.CVProfile, error) {
	query := `INSERT INTO cv_profiles (user_id, title) VALUES ($1, $2) RETURNING id, user_id, title, biodata, created_at, updated_at`
	var p domain.CVProfile
	err := r.db.GetContext(ctx, &p, query, userID, title)
	return &p, err
}

func (r *postgresRepo) GetCVProfile(ctx context.Context, id uuid.UUID) (*domain.CVProfile, error) {
	query := `SELECT id, user_id, title, biodata, created_at, updated_at FROM cv_profiles WHERE id = $1`
	var p domain.CVProfile
	err := r.db.GetContext(ctx, &p, query, id)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return &p, err
}

func (r *postgresRepo) GetCVProfilesByUserID(ctx context.Context, userID uuid.UUID) ([]domain.CVProfile, error) {
	query := `SELECT id, user_id, title, biodata, created_at, updated_at FROM cv_profiles WHERE user_id = $1 ORDER BY created_at ASC`
	var profiles []domain.CVProfile
	err := r.db.SelectContext(ctx, &profiles, query, userID)
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

func (r *postgresRepo) UpdateCVProfileBiodata(ctx context.Context, id uuid.UUID, biodata json.RawMessage) error {
	_, err := r.db.ExecContext(ctx, `UPDATE cv_profiles SET biodata = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, biodata, id)
	return err
}

func (r *postgresRepo) DeleteCVProfile(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM cv_profiles WHERE id = $1`, id)
	return err
}

// ── AI Usage Tracking ──────────────────────────────────────────

func (r *postgresRepo) IncrementAITokensUsed(ctx context.Context, userID uuid.UUID, chars int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET ai_tokens_used = ai_tokens_used + $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1`, userID, chars)
	return err
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
			(SELECT COUNT(*) FROM sessions WHERE expires_at > NOW()) AS total_sessions
	`)
	return &stats, err
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
			u.id, u.email, u.name, COALESCE(u.picture, '') AS picture, u.role, u.ai_chars_used, u.created_at,
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
