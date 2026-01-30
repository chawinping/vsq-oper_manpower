package postgres

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
)

type preferencePositionRequirementRepository struct {
	db *sql.DB
}

func NewPreferencePositionRequirementRepository(db *sql.DB) interfaces.PreferencePositionRequirementRepository {
	return &preferencePositionRequirementRepository{db: db}
}

func (r *preferencePositionRequirementRepository) Create(requirement *models.PreferencePositionRequirement) error {
	query := `INSERT INTO clinic_preference_position_requirements 
	          (id, preference_id, position_id, minimum_staff, preferred_staff, is_active)
	          VALUES ($1, $2, $3, $4, $5, $6)
	          RETURNING created_at, updated_at`
	return r.db.QueryRow(
		query,
		requirement.ID, requirement.PreferenceID, requirement.PositionID,
		requirement.MinimumStaff, requirement.PreferredStaff, requirement.IsActive,
	).Scan(&requirement.CreatedAt, &requirement.UpdatedAt)
}

func (r *preferencePositionRequirementRepository) GetByID(id uuid.UUID) (*models.PreferencePositionRequirement, error) {
	requirement := &models.PreferencePositionRequirement{}
	query := `SELECT id, preference_id, position_id, minimum_staff, preferred_staff, is_active, 
	          created_at, updated_at
	          FROM clinic_preference_position_requirements WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&requirement.ID, &requirement.PreferenceID, &requirement.PositionID,
		&requirement.MinimumStaff, &requirement.PreferredStaff, &requirement.IsActive,
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

func (r *preferencePositionRequirementRepository) GetByPreferenceID(preferenceID uuid.UUID) ([]*models.PreferencePositionRequirement, error) {
	query := `SELECT id, preference_id, position_id, minimum_staff, preferred_staff, is_active, 
	          created_at, updated_at
	          FROM clinic_preference_position_requirements 
	          WHERE preference_id = $1
	          ORDER BY position_id`
	
	rows, err := r.db.Query(query, preferenceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requirements := []*models.PreferencePositionRequirement{}
	for rows.Next() {
		requirement := &models.PreferencePositionRequirement{}
		if err := rows.Scan(
			&requirement.ID, &requirement.PreferenceID, &requirement.PositionID,
			&requirement.MinimumStaff, &requirement.PreferredStaff, &requirement.IsActive,
			&requirement.CreatedAt, &requirement.UpdatedAt,
		); err != nil {
			return nil, err
		}
		requirements = append(requirements, requirement)
	}

	return requirements, rows.Err()
}

func (r *preferencePositionRequirementRepository) GetByPreferenceAndPosition(preferenceID, positionID uuid.UUID) (*models.PreferencePositionRequirement, error) {
	requirement := &models.PreferencePositionRequirement{}
	query := `SELECT id, preference_id, position_id, minimum_staff, preferred_staff, is_active, 
	          created_at, updated_at
	          FROM clinic_preference_position_requirements 
	          WHERE preference_id = $1 AND position_id = $2`
	err := r.db.QueryRow(query, preferenceID, positionID).Scan(
		&requirement.ID, &requirement.PreferenceID, &requirement.PositionID,
		&requirement.MinimumStaff, &requirement.PreferredStaff, &requirement.IsActive,
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

func (r *preferencePositionRequirementRepository) Update(requirement *models.PreferencePositionRequirement) error {
	query := `UPDATE clinic_preference_position_requirements
	          SET minimum_staff = $1, preferred_staff = $2, is_active = $3, updated_at = CURRENT_TIMESTAMP
	          WHERE id = $4 RETURNING updated_at`
	return r.db.QueryRow(
		query,
		requirement.MinimumStaff, requirement.PreferredStaff, requirement.IsActive,
		requirement.ID,
	).Scan(&requirement.UpdatedAt)
}

func (r *preferencePositionRequirementRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM clinic_preference_position_requirements WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *preferencePositionRequirementRepository) DeleteByPreferenceID(preferenceID uuid.UUID) error {
	query := `DELETE FROM clinic_preference_position_requirements WHERE preference_id = $1`
	_, err := r.db.Exec(query, preferenceID)
	return err
}

func (r *preferencePositionRequirementRepository) BulkUpsert(requirements []*models.PreferencePositionRequirement) error {
	if len(requirements) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO clinic_preference_position_requirements 
		(id, preference_id, position_id, minimum_staff, preferred_staff, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (preference_id, position_id) 
		DO UPDATE SET 
			minimum_staff = EXCLUDED.minimum_staff,
			preferred_staff = EXCLUDED.preferred_staff,
			is_active = EXCLUDED.is_active,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, req := range requirements {
		_, err := stmt.Exec(
			req.ID, req.PreferenceID, req.PositionID,
			req.MinimumStaff, req.PreferredStaff, req.IsActive,
		)
		if err != nil {
			return fmt.Errorf("failed to upsert requirement: %w", err)
		}
	}

	return tx.Commit()
}
