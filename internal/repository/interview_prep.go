package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/semmidev/atstex-lab/internal/domain"
)

// CreateInterviewPrep saves a new interview prep record.
func (r *postgresRepo) CreateInterviewPrep(ctx context.Context, prep *domain.InterviewPrep) error {
	query := `INSERT INTO interview_preps (user_id, profile_id, job_description, language, questions, tokens_used)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`
	return r.db.QueryRowxContext(ctx, query,
		prep.UserID, prep.ProfileID, prep.JobDescription, prep.Language, prep.Questions, prep.TokensUsed,
	).Scan(&prep.ID, &prep.CreatedAt)
}

// GetInterviewPrepsByUserID retrieves a user's interview prep history.
func (r *postgresRepo) GetInterviewPrepsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.InterviewPrep, error) {
	var preps []domain.InterviewPrep
	err := r.db.SelectContext(ctx, &preps,
		`SELECT id, user_id, profile_id, job_description, language, questions, tokens_used, created_at
		 FROM interview_preps WHERE user_id = $1 ORDER BY created_at DESC LIMIT 50`, userID)
	if err != nil {
		if err != nil {
			return nil, translatePgError(err, "record", nil)
		}
		return nil, nil
	}
	if preps == nil {
		preps = []domain.InterviewPrep{}
	}
	return preps, nil
}

// CountInterviewPrepsByDate counts how many interview preps a user has generated within a time range.
func (r *postgresRepo) CountInterviewPrepsByDate(ctx context.Context, userID uuid.UUID, start, end time.Time) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM interview_preps WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3`,
		userID, start, end)
	if err != nil {
		return count, translatePgError(err, "record", nil)
	}
	return count, nil
}
