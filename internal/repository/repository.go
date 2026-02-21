package repository

import (
	"context"
	"database/sql"
	"errors"
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

func (r *postgresRepo) UpsertUser(ctx context.Context, googleID, email, name, picture string) (*domain.User, error) {
	query := `
		INSERT INTO users (google_id, email, name, picture)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (google_id) DO UPDATE SET
			email = EXCLUDED.email,
			name = EXCLUDED.name,
			picture = EXCLUDED.picture,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, google_id, email, name, picture, created_at, updated_at
	`
	var u domain.User
	err := r.db.GetContext(ctx, &u, query, googleID, email, name, picture)
	return &u, err
}

func (r *postgresRepo) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, google_id, email, name, picture, created_at, updated_at FROM users WHERE id = $1`
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
