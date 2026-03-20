package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/semmidev/atstex-lab/internal/domain"
)

// CreateMockInterviewSession inserts a new mock interview session record and returns the saved entity.
func (r *postgresRepo) CreateMockInterviewSession(ctx context.Context, s *domain.MockInterviewSession) error {
	msgs, err := json.Marshal(s.Messages)
	if err != nil {
		msgs = []byte("[]")
	}

	query := `
		INSERT INTO mock_interview_sessions
			(user_id, profile_id, job_description, language, interviewer_style, messages, tokens_used, turn_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`

	return r.db.QueryRowxContext(ctx, query,
		s.UserID,
		s.ProfileID,
		s.JobDescription,
		s.Language,
		s.InterviewerStyle,
		msgs,
		s.TokensUsed,
		s.TurnCount,
	).Scan(&s.ID, &s.CreatedAt)
}

// UpdateMockInterviewSession persists changes to an existing session (messages, tokens, turn count, ended_at).
func (r *postgresRepo) UpdateMockInterviewSession(ctx context.Context, s *domain.MockInterviewSession) error {
	msgs, err := json.Marshal(s.Messages)
	if err != nil {
		msgs = []byte("[]")
	}

	query := `
		UPDATE mock_interview_sessions
		SET messages    = $1,
		    tokens_used = $2,
		    turn_count  = $3,
		    ended_at    = $4
		WHERE id = $5`

	_, err = r.db.ExecContext(ctx, query,
		msgs,
		s.TokensUsed,
		s.TurnCount,
		s.EndedAt,
		s.ID,
	)
	return translatePgError(err, "record", nil)
}

// GetMockInterviewSession fetches a single session by ID.
func (r *postgresRepo) GetMockInterviewSession(ctx context.Context, id uuid.UUID) (*domain.MockInterviewSession, error) {
	var s domain.MockInterviewSession
	err := r.db.GetContext(ctx, &s,
		`SELECT id, user_id, profile_id, job_description, language, interviewer_style, messages, tokens_used, turn_count, created_at, ended_at
		 FROM mock_interview_sessions
		 WHERE id = $1`, id)
	if err != nil {
		return nil, translatePgError(err, "record", nil)
	}
	return &s, nil
}

// GetMockInterviewSessionsByUserID returns all sessions belonging to a user, most recent first.
func (r *postgresRepo) GetMockInterviewSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.MockInterviewSession, error) {
	var sessions []domain.MockInterviewSession
	err := r.db.SelectContext(ctx, &sessions,
		`SELECT id, user_id, profile_id, job_description, language, interviewer_style, messages, tokens_used, turn_count, created_at, ended_at
		 FROM mock_interview_sessions
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT 20`, userID)
	if err != nil {
		return nil, translatePgError(err, "record", nil)
	}
	if sessions == nil {
		sessions = []domain.MockInterviewSession{}
	}
	return sessions, nil
}

// EndMockInterviewSession marks a session as ended by setting ended_at to now.
func (r *postgresRepo) EndMockInterviewSession(ctx context.Context, id uuid.UUID, tokensUsed int64, turnCount int, messages json.RawMessage) error {
	now := time.Now()
	msgs := messages
	if msgs == nil {
		msgs = []byte("[]")
	}

	_, err := r.db.ExecContext(ctx,
		`UPDATE mock_interview_sessions
		 SET ended_at    = $1,
		     tokens_used = $2,
		     turn_count  = $3,
		     messages    = $4
		 WHERE id = $5`,
		now, tokensUsed, turnCount, msgs, id,
	)
	return translatePgError(err, "record", nil)
}

// CountMockInterviewSessionsByDate counts sessions for rate-limiting purposes.
func (r *postgresRepo) CountMockInterviewSessionsByDate(ctx context.Context, userID uuid.UUID, start, end time.Time) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM mock_interview_sessions
		 WHERE user_id = $1 AND created_at >= $2 AND created_at <= $3`,
		userID, start, end)
	if err != nil {
		return count, translatePgError(err, "record", nil)
	}
	return count, nil
}
