package postgres

import (
	"database/sql"

	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"

	"github.com/google/uuid"
)

type scenarioSpecificStaffRequirementRepository struct {
	db *sql.DB
}

func NewScenarioSpecificStaffRequirementRepository(db *sql.DB) interfaces.ScenarioSpecificStaffRequirementRepository {
	return &scenarioSpecificStaffRequirementRepository{db: db}
}

func (r *scenarioSpecificStaffRequirementRepository) Create(requirement *models.ScenarioSpecificStaffRequirement) error {
	query := `INSERT INTO scenario_specific_staff_requirements 
	          (id, scenario_id, staff_id)
	          VALUES ($1, $2, $3) RETURNING created_at, updated_at`
	return r.db.QueryRow(
		query,
		requirement.ID, requirement.ScenarioID, requirement.StaffID,
	).Scan(&requirement.CreatedAt, &requirement.UpdatedAt)
}

func (r *scenarioSpecificStaffRequirementRepository) GetByID(id uuid.UUID) (*models.ScenarioSpecificStaffRequirement, error) {
	requirement := &models.ScenarioSpecificStaffRequirement{}
	query := `SELECT id, scenario_id, staff_id, created_at, updated_at
	          FROM scenario_specific_staff_requirements WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&requirement.ID, &requirement.ScenarioID, &requirement.StaffID,
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

func (r *scenarioSpecificStaffRequirementRepository) GetByScenarioID(scenarioID uuid.UUID) ([]*models.ScenarioSpecificStaffRequirement, error) {
	query := `SELECT id, scenario_id, staff_id, created_at, updated_at
	          FROM scenario_specific_staff_requirements WHERE scenario_id = $1 ORDER BY staff_id`
	rows, err := r.db.Query(query, scenarioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requirements := []*models.ScenarioSpecificStaffRequirement{}
	for rows.Next() {
		requirement := &models.ScenarioSpecificStaffRequirement{}
		if err := rows.Scan(
			&requirement.ID, &requirement.ScenarioID, &requirement.StaffID,
			&requirement.CreatedAt, &requirement.UpdatedAt,
		); err != nil {
			return nil, err
		}
		requirements = append(requirements, requirement)
	}
	return requirements, rows.Err()
}

func (r *scenarioSpecificStaffRequirementRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM scenario_specific_staff_requirements WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *scenarioSpecificStaffRequirementRepository) DeleteByScenarioID(scenarioID uuid.UUID) error {
	query := `DELETE FROM scenario_specific_staff_requirements WHERE scenario_id = $1`
	_, err := r.db.Exec(query, scenarioID)
	return err
}

func (r *scenarioSpecificStaffRequirementRepository) BulkUpsert(requirements []*models.ScenarioSpecificStaffRequirement) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO scenario_specific_staff_requirements 
	          (id, scenario_id, staff_id, created_at, updated_at)
	          VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	          ON CONFLICT (scenario_id, staff_id)
	          DO UPDATE SET updated_at = CURRENT_TIMESTAMP
	          RETURNING created_at, updated_at`

	for _, requirement := range requirements {
		if requirement.ID == uuid.Nil {
			requirement.ID = uuid.New()
		}
		err := tx.QueryRow(
			query,
			requirement.ID, requirement.ScenarioID, requirement.StaffID,
		).Scan(&requirement.CreatedAt, &requirement.UpdatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
