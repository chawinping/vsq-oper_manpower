package postgres

import (
	"database/sql"

	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
)

type scenarioPositionRequirementRepository struct {
	db *sql.DB
}

func NewScenarioPositionRequirementRepository(db *sql.DB) interfaces.ScenarioPositionRequirementRepository {
	return &scenarioPositionRequirementRepository{db: db}
}

func (r *scenarioPositionRequirementRepository) Create(requirement *models.ScenarioPositionRequirement) error {
	query := `INSERT INTO scenario_position_requirements 
	          (id, scenario_id, position_id, preferred_staff, minimum_staff, override_base)
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at, updated_at`
	return r.db.QueryRow(
		query,
		requirement.ID, requirement.ScenarioID, requirement.PositionID,
		requirement.PreferredStaff, requirement.MinimumStaff, requirement.OverrideBase,
	).Scan(&requirement.CreatedAt, &requirement.UpdatedAt)
}

func (r *scenarioPositionRequirementRepository) GetByID(id uuid.UUID) (*models.ScenarioPositionRequirement, error) {
	requirement := &models.ScenarioPositionRequirement{}
	query := `SELECT id, scenario_id, position_id, preferred_staff, minimum_staff, override_base, created_at, updated_at
	          FROM scenario_position_requirements WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&requirement.ID, &requirement.ScenarioID, &requirement.PositionID,
		&requirement.PreferredStaff, &requirement.MinimumStaff, &requirement.OverrideBase,
		&requirement.CreatedAt, &requirement.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return requirement, nil
}

func (r *scenarioPositionRequirementRepository) GetByScenarioID(scenarioID uuid.UUID) ([]*models.ScenarioPositionRequirement, error) {
	query := `SELECT id, scenario_id, position_id, preferred_staff, minimum_staff, override_base, created_at, updated_at
	          FROM scenario_position_requirements WHERE scenario_id = $1 ORDER BY position_id`
	rows, err := r.db.Query(query, scenarioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requirements := []*models.ScenarioPositionRequirement{}
	for rows.Next() {
		requirement := &models.ScenarioPositionRequirement{}
		if err := rows.Scan(
			&requirement.ID, &requirement.ScenarioID, &requirement.PositionID,
			&requirement.PreferredStaff, &requirement.MinimumStaff, &requirement.OverrideBase,
			&requirement.CreatedAt, &requirement.UpdatedAt,
		); err != nil {
			return nil, err
		}
		requirements = append(requirements, requirement)
	}
	return requirements, rows.Err()
}

func (r *scenarioPositionRequirementRepository) GetByScenarioAndPosition(scenarioID, positionID uuid.UUID) (*models.ScenarioPositionRequirement, error) {
	requirement := &models.ScenarioPositionRequirement{}
	query := `SELECT id, scenario_id, position_id, preferred_staff, minimum_staff, override_base, created_at, updated_at
	          FROM scenario_position_requirements WHERE scenario_id = $1 AND position_id = $2`
	err := r.db.QueryRow(query, scenarioID, positionID).Scan(
		&requirement.ID, &requirement.ScenarioID, &requirement.PositionID,
		&requirement.PreferredStaff, &requirement.MinimumStaff, &requirement.OverrideBase,
		&requirement.CreatedAt, &requirement.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return requirement, nil
}

func (r *scenarioPositionRequirementRepository) Update(requirement *models.ScenarioPositionRequirement) error {
	query := `UPDATE scenario_position_requirements
	          SET preferred_staff = $1, minimum_staff = $2, override_base = $3, updated_at = CURRENT_TIMESTAMP
	          WHERE id = $4 RETURNING updated_at`
	return r.db.QueryRow(
		query,
		requirement.PreferredStaff, requirement.MinimumStaff, requirement.OverrideBase, requirement.ID,
	).Scan(&requirement.UpdatedAt)
}

func (r *scenarioPositionRequirementRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM scenario_position_requirements WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *scenarioPositionRequirementRepository) DeleteByScenarioID(scenarioID uuid.UUID) error {
	query := `DELETE FROM scenario_position_requirements WHERE scenario_id = $1`
	_, err := r.db.Exec(query, scenarioID)
	return err
}

func (r *scenarioPositionRequirementRepository) BulkUpsert(requirements []*models.ScenarioPositionRequirement) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO scenario_position_requirements 
	          (id, scenario_id, position_id, preferred_staff, minimum_staff, override_base, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	          ON CONFLICT (scenario_id, position_id)
	          DO UPDATE SET preferred_staff = EXCLUDED.preferred_staff,
	                        minimum_staff = EXCLUDED.minimum_staff,
	                        override_base = EXCLUDED.override_base,
	                        updated_at = CURRENT_TIMESTAMP
	          RETURNING created_at, updated_at`

	for _, requirement := range requirements {
		if requirement.ID == uuid.Nil {
			requirement.ID = uuid.New()
		}
		err := tx.QueryRow(
			query,
			requirement.ID, requirement.ScenarioID, requirement.PositionID,
			requirement.PreferredStaff, requirement.MinimumStaff, requirement.OverrideBase,
		).Scan(&requirement.CreatedAt, &requirement.UpdatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
