package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/semmidev/atstex-lab/internal/domain"
)

// CreateAtsSimulation saves an ATS simulation to the database.
func (p *postgresRepo) CreateAtsSimulation(ctx context.Context, sim *domain.AtsSimulation) error {
	query := `
		INSERT INTO ats_simulations (user_id, profile_id, job_description, score, missing_keywords, recommendations)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	err := p.db.QueryRowContext(ctx, query,
		sim.UserID,
		sim.ProfileID,
		sim.JobDescription,
		sim.Score,
		sim.MissingKeywords,
		sim.Recommendations,
	).Scan(&sim.ID, &sim.CreatedAt)

	if err != nil {
		return err
	}
	return nil
}

// GetAtsSimulationsByUserID returns all ATS simulations for a specific user.
func (p *postgresRepo) GetAtsSimulationsByUserID(ctx context.Context, userID uuid.UUID) ([]domain.AtsSimulation, error) {
	query := `
		SELECT id, user_id, profile_id, job_description, score, missing_keywords, recommendations, created_at
		FROM ats_simulations
		WHERE user_id = $1
		ORDER BY created_at DESC`

	var simulations []domain.AtsSimulation
	err := p.db.SelectContext(ctx, &simulations, query, userID)
	if err != nil {
		return nil, err
	}

	return simulations, nil
}
