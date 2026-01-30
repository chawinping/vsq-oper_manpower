package postgres

import (
	"database/sql"
	"time"

	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"

	"github.com/google/uuid"
)

type staffGroupPositionRepository struct {
	db *sql.DB
}

func NewStaffGroupPositionRepository(db *sql.DB) interfaces.StaffGroupPositionRepository {
	return &staffGroupPositionRepository{db: db}
}

func (r *staffGroupPositionRepository) Create(sgp *models.StaffGroupPosition) error {
	sgp.ID = uuid.New()
	sgp.CreatedAt = time.Now()

	query := `INSERT INTO staff_group_positions (id, staff_group_id, position_id, created_at)
	          VALUES ($1, $2, $3, $4) RETURNING created_at`

	return r.db.QueryRow(query, sgp.ID, sgp.StaffGroupID, sgp.PositionID, sgp.CreatedAt).
		Scan(&sgp.CreatedAt)
}

func (r *staffGroupPositionRepository) GetByID(id uuid.UUID) (*models.StaffGroupPosition, error) {
	sgp := &models.StaffGroupPosition{}
	query := `SELECT id, staff_group_id, position_id, created_at
	          FROM staff_group_positions WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&sgp.ID, &sgp.StaffGroupID, &sgp.PositionID, &sgp.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return sgp, nil
}

func (r *staffGroupPositionRepository) GetByStaffGroupID(staffGroupID uuid.UUID) ([]*models.StaffGroupPosition, error) {
	query := `SELECT id, staff_group_id, position_id, created_at
	          FROM staff_group_positions WHERE staff_group_id = $1 ORDER BY position_id`

	rows, err := r.db.Query(query, staffGroupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	positions := []*models.StaffGroupPosition{}
	for rows.Next() {
		sgp := &models.StaffGroupPosition{}
		if err := rows.Scan(
			&sgp.ID, &sgp.StaffGroupID, &sgp.PositionID, &sgp.CreatedAt,
		); err != nil {
			return nil, err
		}
		positions = append(positions, sgp)
	}

	return positions, rows.Err()
}

func (r *staffGroupPositionRepository) GetByPositionID(positionID uuid.UUID) ([]*models.StaffGroupPosition, error) {
	query := `SELECT id, staff_group_id, position_id, created_at
	          FROM staff_group_positions WHERE position_id = $1 ORDER BY staff_group_id`

	rows, err := r.db.Query(query, positionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	positions := []*models.StaffGroupPosition{}
	for rows.Next() {
		sgp := &models.StaffGroupPosition{}
		if err := rows.Scan(
			&sgp.ID, &sgp.StaffGroupID, &sgp.PositionID, &sgp.CreatedAt,
		); err != nil {
			return nil, err
		}
		positions = append(positions, sgp)
	}

	return positions, rows.Err()
}

func (r *staffGroupPositionRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM staff_group_positions WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *staffGroupPositionRepository) DeleteByStaffGroupAndPosition(staffGroupID uuid.UUID, positionID uuid.UUID) error {
	query := `DELETE FROM staff_group_positions WHERE staff_group_id = $1 AND position_id = $2`
	_, err := r.db.Exec(query, staffGroupID, positionID)
	return err
}
