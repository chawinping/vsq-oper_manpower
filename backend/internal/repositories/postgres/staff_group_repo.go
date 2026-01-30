package postgres

import (
	"database/sql"
	"time"

	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"

	"github.com/google/uuid"
)

type staffGroupRepository struct {
	db *sql.DB
}

func NewStaffGroupRepository(db *sql.DB) interfaces.StaffGroupRepository {
	return &staffGroupRepository{db: db}
}

func (r *staffGroupRepository) Create(staffGroup *models.StaffGroup) error {
	staffGroup.ID = uuid.New()
	staffGroup.CreatedAt = time.Now()
	staffGroup.UpdatedAt = time.Now()

	query := `INSERT INTO staff_groups (id, name, description, is_active, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING created_at, updated_at`

	var description sql.NullString
	if staffGroup.Description != "" {
		description = sql.NullString{String: staffGroup.Description, Valid: true}
	}

	return r.db.QueryRow(query, staffGroup.ID, staffGroup.Name, description, staffGroup.IsActive, staffGroup.CreatedAt, staffGroup.UpdatedAt).
		Scan(&staffGroup.CreatedAt, &staffGroup.UpdatedAt)
}

func (r *staffGroupRepository) GetByID(id uuid.UUID) (*models.StaffGroup, error) {
	staffGroup := &models.StaffGroup{}
	query := `SELECT id, name, description, is_active, created_at, updated_at
	          FROM staff_groups WHERE id = $1`

	var description sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&staffGroup.ID, &staffGroup.Name, &description, &staffGroup.IsActive, &staffGroup.CreatedAt, &staffGroup.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if description.Valid {
		staffGroup.Description = description.String
	}

	return staffGroup, nil
}

func (r *staffGroupRepository) List() ([]*models.StaffGroup, error) {
	query := `SELECT id, name, description, is_active, created_at, updated_at
	          FROM staff_groups ORDER BY name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	staffGroups := []*models.StaffGroup{}
	for rows.Next() {
		staffGroup := &models.StaffGroup{}
		var description sql.NullString

		if err := rows.Scan(
			&staffGroup.ID, &staffGroup.Name, &description, &staffGroup.IsActive, &staffGroup.CreatedAt, &staffGroup.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if description.Valid {
			staffGroup.Description = description.String
		}

		staffGroups = append(staffGroups, staffGroup)
	}

	return staffGroups, rows.Err()
}

func (r *staffGroupRepository) GetByPositionID(positionID uuid.UUID) ([]*models.StaffGroup, error) {
	query := `SELECT DISTINCT sg.id, sg.name, sg.description, sg.is_active, sg.created_at, sg.updated_at
	          FROM staff_groups sg
	          INNER JOIN staff_group_positions sgp ON sg.id = sgp.staff_group_id
	          WHERE sgp.position_id = $1 AND sg.is_active = true
	          ORDER BY sg.name`

	rows, err := r.db.Query(query, positionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	staffGroups := []*models.StaffGroup{}
	for rows.Next() {
		staffGroup := &models.StaffGroup{}
		var description sql.NullString

		if err := rows.Scan(
			&staffGroup.ID, &staffGroup.Name, &description, &staffGroup.IsActive, &staffGroup.CreatedAt, &staffGroup.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if description.Valid {
			staffGroup.Description = description.String
		}

		staffGroups = append(staffGroups, staffGroup)
	}

	return staffGroups, rows.Err()
}

func (r *staffGroupRepository) Update(staffGroup *models.StaffGroup) error {
	staffGroup.UpdatedAt = time.Now()

	query := `UPDATE staff_groups SET name = $1, description = $2, is_active = $3, updated_at = $4 WHERE id = $5`

	var description sql.NullString
	if staffGroup.Description != "" {
		description = sql.NullString{String: staffGroup.Description, Valid: true}
	}

	_, err := r.db.Exec(query, staffGroup.Name, description, staffGroup.IsActive, staffGroup.UpdatedAt, staffGroup.ID)
	return err
}

func (r *staffGroupRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM staff_groups WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
