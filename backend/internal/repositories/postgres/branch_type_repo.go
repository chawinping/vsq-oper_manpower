package postgres

import (
	"database/sql"
	"time"

	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"

	"github.com/google/uuid"
)

type branchTypeRepository struct {
	db *sql.DB
}

func NewBranchTypeRepository(db *sql.DB) interfaces.BranchTypeRepository {
	return &branchTypeRepository{db: db}
}

func (r *branchTypeRepository) Create(branchType *models.BranchType) error {
	branchType.ID = uuid.New()
	branchType.CreatedAt = time.Now()
	branchType.UpdatedAt = time.Now()

	query := `INSERT INTO branch_types (id, name, description, is_active, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at, updated_at`

	var description sql.NullString
	if branchType.Description != "" {
		description = sql.NullString{String: branchType.Description, Valid: true}
	}

	return r.db.QueryRow(query, branchType.ID, branchType.Name, description, branchType.IsActive, branchType.CreatedAt, branchType.UpdatedAt).
		Scan(&branchType.CreatedAt, &branchType.UpdatedAt)
}

func (r *branchTypeRepository) GetByID(id uuid.UUID) (*models.BranchType, error) {
	branchType := &models.BranchType{}
	query := `SELECT id, name, description, is_active, created_at, updated_at
	          FROM branch_types WHERE id = $1`

	var description sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&branchType.ID, &branchType.Name, &description, &branchType.IsActive, &branchType.CreatedAt, &branchType.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if description.Valid {
		branchType.Description = description.String
	}

	return branchType, nil
}

func (r *branchTypeRepository) List() ([]*models.BranchType, error) {
	query := `SELECT id, name, description, is_active, created_at, updated_at
	          FROM branch_types ORDER BY name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	branchTypes := []*models.BranchType{}
	for rows.Next() {
		branchType := &models.BranchType{}
		var description sql.NullString

		if err := rows.Scan(
			&branchType.ID, &branchType.Name, &description, &branchType.IsActive, &branchType.CreatedAt, &branchType.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if description.Valid {
			branchType.Description = description.String
		}

		branchTypes = append(branchTypes, branchType)
	}

	return branchTypes, rows.Err()
}

func (r *branchTypeRepository) Update(branchType *models.BranchType) error {
	branchType.UpdatedAt = time.Now()

	query := `UPDATE branch_types SET name = $1, description = $2, is_active = $3, updated_at = $4 WHERE id = $5`

	var description sql.NullString
	if branchType.Description != "" {
		description = sql.NullString{String: branchType.Description, Valid: true}
	}

	_, err := r.db.Exec(query, branchType.Name, description, branchType.IsActive, branchType.UpdatedAt, branchType.ID)
	return err
}

func (r *branchTypeRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM branch_types WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
