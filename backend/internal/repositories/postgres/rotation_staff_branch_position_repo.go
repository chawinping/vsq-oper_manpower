package postgres

import (
	"database/sql"
	"time"

	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"

	"github.com/google/uuid"
)

type rotationStaffBranchPositionRepository struct {
	db *sql.DB
}

func NewRotationStaffBranchPositionRepository(db *sql.DB) interfaces.RotationStaffBranchPositionRepository {
	return &rotationStaffBranchPositionRepository{db: db}
}

func (r *rotationStaffBranchPositionRepository) Create(mapping *models.RotationStaffBranchPosition) error {
	mapping.ID = uuid.New()
	mapping.CreatedAt = time.Now()
	mapping.UpdatedAt = time.Now()

	query := `INSERT INTO rotation_staff_branch_positions 
	          (id, rotation_staff_id, branch_position_id, substitution_level, is_active, notes, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING created_at, updated_at`

	var notes sql.NullString
	if mapping.Notes != "" {
		notes = sql.NullString{String: mapping.Notes, Valid: true}
	}

	return r.db.QueryRow(query,
		mapping.ID,
		mapping.RotationStaffID,
		mapping.BranchPositionID,
		mapping.SubstitutionLevel,
		mapping.IsActive,
		notes,
		mapping.CreatedAt,
		mapping.UpdatedAt,
	).Scan(&mapping.CreatedAt, &mapping.UpdatedAt)
}

func (r *rotationStaffBranchPositionRepository) GetByID(id uuid.UUID) (*models.RotationStaffBranchPosition, error) {
	mapping := &models.RotationStaffBranchPosition{}
	query := `SELECT id, rotation_staff_id, branch_position_id, substitution_level, is_active, notes, created_at, updated_at
	          FROM rotation_staff_branch_positions WHERE id = $1`

	var notes sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&mapping.ID,
		&mapping.RotationStaffID,
		&mapping.BranchPositionID,
		&mapping.SubstitutionLevel,
		&mapping.IsActive,
		&notes,
		&mapping.CreatedAt,
		&mapping.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if notes.Valid {
		mapping.Notes = notes.String
	}

	return mapping, nil
}

func (r *rotationStaffBranchPositionRepository) List() ([]*models.RotationStaffBranchPosition, error) {
	query := `SELECT id, rotation_staff_id, branch_position_id, substitution_level, is_active, notes, created_at, updated_at
	          FROM rotation_staff_branch_positions 
	          ORDER BY rotation_staff_id, substitution_level, branch_position_id`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mappings := []*models.RotationStaffBranchPosition{}
	for rows.Next() {
		mapping := &models.RotationStaffBranchPosition{}
		var notes sql.NullString

		if err := rows.Scan(
			&mapping.ID,
			&mapping.RotationStaffID,
			&mapping.BranchPositionID,
			&mapping.SubstitutionLevel,
			&mapping.IsActive,
			&notes,
			&mapping.CreatedAt,
			&mapping.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if notes.Valid {
			mapping.Notes = notes.String
		}

		mappings = append(mappings, mapping)
	}

	return mappings, rows.Err()
}

func (r *rotationStaffBranchPositionRepository) GetByStaffID(rotationStaffID uuid.UUID) ([]*models.RotationStaffBranchPosition, error) {
	query := `SELECT id, rotation_staff_id, branch_position_id, substitution_level, is_active, notes, created_at, updated_at
	          FROM rotation_staff_branch_positions 
	          WHERE rotation_staff_id = $1 
	          ORDER BY substitution_level, branch_position_id`

	rows, err := r.db.Query(query, rotationStaffID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mappings := []*models.RotationStaffBranchPosition{}
	for rows.Next() {
		mapping := &models.RotationStaffBranchPosition{}
		var notes sql.NullString

		if err := rows.Scan(
			&mapping.ID,
			&mapping.RotationStaffID,
			&mapping.BranchPositionID,
			&mapping.SubstitutionLevel,
			&mapping.IsActive,
			&notes,
			&mapping.CreatedAt,
			&mapping.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if notes.Valid {
			mapping.Notes = notes.String
		}

		mappings = append(mappings, mapping)
	}

	return mappings, rows.Err()
}

func (r *rotationStaffBranchPositionRepository) GetByPositionID(branchPositionID uuid.UUID) ([]*models.RotationStaffBranchPosition, error) {
	query := `SELECT id, rotation_staff_id, branch_position_id, substitution_level, is_active, notes, created_at, updated_at
	          FROM rotation_staff_branch_positions 
	          WHERE branch_position_id = $1 AND is_active = true
	          ORDER BY substitution_level, rotation_staff_id`

	rows, err := r.db.Query(query, branchPositionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mappings := []*models.RotationStaffBranchPosition{}
	for rows.Next() {
		mapping := &models.RotationStaffBranchPosition{}
		var notes sql.NullString

		if err := rows.Scan(
			&mapping.ID,
			&mapping.RotationStaffID,
			&mapping.BranchPositionID,
			&mapping.SubstitutionLevel,
			&mapping.IsActive,
			&notes,
			&mapping.CreatedAt,
			&mapping.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if notes.Valid {
			mapping.Notes = notes.String
		}

		mappings = append(mappings, mapping)
	}

	return mappings, rows.Err()
}

func (r *rotationStaffBranchPositionRepository) GetByStaffAndPosition(rotationStaffID uuid.UUID, branchPositionID uuid.UUID) (*models.RotationStaffBranchPosition, error) {
	mapping := &models.RotationStaffBranchPosition{}
	query := `SELECT id, rotation_staff_id, branch_position_id, substitution_level, is_active, notes, created_at, updated_at
	          FROM rotation_staff_branch_positions 
	          WHERE rotation_staff_id = $1 AND branch_position_id = $2 AND is_active = true`

	var notes sql.NullString

	err := r.db.QueryRow(query, rotationStaffID, branchPositionID).Scan(
		&mapping.ID,
		&mapping.RotationStaffID,
		&mapping.BranchPositionID,
		&mapping.SubstitutionLevel,
		&mapping.IsActive,
		&notes,
		&mapping.CreatedAt,
		&mapping.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if notes.Valid {
		mapping.Notes = notes.String
	}

	return mapping, nil
}

func (r *rotationStaffBranchPositionRepository) Update(mapping *models.RotationStaffBranchPosition) error {
	mapping.UpdatedAt = time.Now()

	query := `UPDATE rotation_staff_branch_positions 
	          SET substitution_level = $1, is_active = $2, notes = $3, updated_at = $4 
	          WHERE id = $5`

	var notes sql.NullString
	if mapping.Notes != "" {
		notes = sql.NullString{String: mapping.Notes, Valid: true}
	}

	_, err := r.db.Exec(query,
		mapping.SubstitutionLevel,
		mapping.IsActive,
		notes,
		mapping.UpdatedAt,
		mapping.ID,
	)
	return err
}

func (r *rotationStaffBranchPositionRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM rotation_staff_branch_positions WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *rotationStaffBranchPositionRepository) DeleteByStaffAndPosition(rotationStaffID uuid.UUID, branchPositionID uuid.UUID) error {
	query := `DELETE FROM rotation_staff_branch_positions 
	          WHERE rotation_staff_id = $1 AND branch_position_id = $2`
	_, err := r.db.Exec(query, rotationStaffID, branchPositionID)
	return err
}
