package postgres

import (
	"database/sql"

	"github.com/google/uuid"
	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"
)

type branchConstraintsRepository struct {
	db *sql.DB
}

func NewBranchConstraintsRepository(db *sql.DB) interfaces.BranchConstraintsRepository {
	return &branchConstraintsRepository{db: db}
}

func (r *branchConstraintsRepository) Create(constraint *models.BranchConstraints) error {
	query := `INSERT INTO branch_constraints (id, branch_id, day_of_week, min_front_staff, min_managers, min_doctor_assistant, min_total_staff) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, constraint.ID, constraint.BranchID, constraint.DayOfWeek,
		constraint.MinFrontStaff, constraint.MinManagers, constraint.MinDoctorAssistant, constraint.MinTotalStaff).
		Scan(&constraint.CreatedAt, &constraint.UpdatedAt)
}

func (r *branchConstraintsRepository) Update(constraint *models.BranchConstraints) error {
	query := `UPDATE branch_constraints SET min_front_staff = $1, min_managers = $2, min_doctor_assistant = $3, min_total_staff = $4, updated_at = CURRENT_TIMESTAMP 
	          WHERE id = $5 RETURNING updated_at`
	return r.db.QueryRow(query, constraint.MinFrontStaff, constraint.MinManagers, constraint.MinDoctorAssistant, constraint.MinTotalStaff, constraint.ID).
		Scan(&constraint.UpdatedAt)
}

func (r *branchConstraintsRepository) GetByID(id uuid.UUID) (*models.BranchConstraints, error) {
	constraint := &models.BranchConstraints{}
	query := `SELECT id, branch_id, day_of_week, min_front_staff, min_managers, min_doctor_assistant, min_total_staff, created_at, updated_at 
	          FROM branch_constraints WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&constraint.ID, &constraint.BranchID, &constraint.DayOfWeek,
		&constraint.MinFrontStaff, &constraint.MinManagers, &constraint.MinDoctorAssistant, &constraint.MinTotalStaff,
		&constraint.CreatedAt, &constraint.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return constraint, err
}

func (r *branchConstraintsRepository) GetByBranchID(branchID uuid.UUID) ([]*models.BranchConstraints, error) {
	query := `SELECT id, branch_id, day_of_week, min_front_staff, min_managers, min_doctor_assistant, min_total_staff, created_at, updated_at 
	          FROM branch_constraints WHERE branch_id = $1 ORDER BY day_of_week`
	rows, err := r.db.Query(query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var constraints []*models.BranchConstraints
	for rows.Next() {
		constraint := &models.BranchConstraints{}
		if err := rows.Scan(
			&constraint.ID, &constraint.BranchID, &constraint.DayOfWeek,
			&constraint.MinFrontStaff, &constraint.MinManagers, &constraint.MinDoctorAssistant, &constraint.MinTotalStaff,
			&constraint.CreatedAt, &constraint.UpdatedAt,
		); err != nil {
			return nil, err
		}
		constraints = append(constraints, constraint)
	}
	return constraints, rows.Err()
}

func (r *branchConstraintsRepository) GetByBranchIDAndDayOfWeek(branchID uuid.UUID, dayOfWeek int) (*models.BranchConstraints, error) {
	constraint := &models.BranchConstraints{}
	query := `SELECT id, branch_id, day_of_week, min_front_staff, min_managers, min_doctor_assistant, min_total_staff, created_at, updated_at 
	          FROM branch_constraints WHERE branch_id = $1 AND day_of_week = $2`
	err := r.db.QueryRow(query, branchID, dayOfWeek).Scan(
		&constraint.ID, &constraint.BranchID, &constraint.DayOfWeek,
		&constraint.MinFrontStaff, &constraint.MinManagers, &constraint.MinDoctorAssistant, &constraint.MinTotalStaff,
		&constraint.CreatedAt, &constraint.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return constraint, err
}

func (r *branchConstraintsRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM branch_constraints WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *branchConstraintsRepository) BulkUpsert(constraints []*models.BranchConstraints) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO branch_constraints (id, branch_id, day_of_week, min_front_staff, min_managers, min_doctor_assistant, min_total_staff, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	          ON CONFLICT (branch_id, day_of_week) 
	          DO UPDATE SET min_front_staff = EXCLUDED.min_front_staff, min_managers = EXCLUDED.min_managers, min_doctor_assistant = EXCLUDED.min_doctor_assistant, min_total_staff = EXCLUDED.min_total_staff, updated_at = CURRENT_TIMESTAMP
	          RETURNING created_at, updated_at`

	for _, constraint := range constraints {
		if constraint.ID == uuid.Nil {
			constraint.ID = uuid.New()
		}
		err := tx.QueryRow(query, constraint.ID, constraint.BranchID, constraint.DayOfWeek,
			constraint.MinFrontStaff, constraint.MinManagers, constraint.MinDoctorAssistant, constraint.MinTotalStaff).
			Scan(&constraint.CreatedAt, &constraint.UpdatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
