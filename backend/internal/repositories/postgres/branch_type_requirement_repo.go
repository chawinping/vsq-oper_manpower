package postgres

import (
	"database/sql"
	"time"

	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"

	"github.com/google/uuid"
)

type branchTypeStaffGroupRequirementRepository struct {
	db *sql.DB
}

func NewBranchTypeStaffGroupRequirementRepository(db *sql.DB) interfaces.BranchTypeStaffGroupRequirementRepository {
	return &branchTypeStaffGroupRequirementRepository{db: db}
}

func (r *branchTypeStaffGroupRequirementRepository) Create(requirement *models.BranchTypeStaffGroupRequirement) error {
	requirement.ID = uuid.New()
	requirement.CreatedAt = time.Now()
	requirement.UpdatedAt = time.Now()

	query := `INSERT INTO branch_type_staff_group_requirements 
	          (id, branch_type_id, staff_group_id, day_of_week, minimum_staff_count, is_active, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING created_at, updated_at`

	return r.db.QueryRow(query,
		requirement.ID, requirement.BranchTypeID, requirement.StaffGroupID, requirement.DayOfWeek,
		requirement.MinimumStaffCount, requirement.IsActive, requirement.CreatedAt, requirement.UpdatedAt,
	).Scan(&requirement.CreatedAt, &requirement.UpdatedAt)
}

func (r *branchTypeStaffGroupRequirementRepository) GetByID(id uuid.UUID) (*models.BranchTypeStaffGroupRequirement, error) {
	requirement := &models.BranchTypeStaffGroupRequirement{}
	query := `SELECT id, branch_type_id, staff_group_id, day_of_week, minimum_staff_count, is_active, created_at, updated_at
	          FROM branch_type_staff_group_requirements WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&requirement.ID, &requirement.BranchTypeID, &requirement.StaffGroupID, &requirement.DayOfWeek,
		&requirement.MinimumStaffCount, &requirement.IsActive, &requirement.CreatedAt, &requirement.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return requirement, nil
}

func (r *branchTypeStaffGroupRequirementRepository) GetByBranchTypeID(branchTypeID uuid.UUID) ([]*models.BranchTypeStaffGroupRequirement, error) {
	query := `SELECT id, branch_type_id, staff_group_id, day_of_week, minimum_staff_count, is_active, created_at, updated_at
	          FROM branch_type_staff_group_requirements WHERE branch_type_id = $1 ORDER BY day_of_week, staff_group_id`

	rows, err := r.db.Query(query, branchTypeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requirements := []*models.BranchTypeStaffGroupRequirement{}
	for rows.Next() {
		requirement := &models.BranchTypeStaffGroupRequirement{}
		if err := rows.Scan(
			&requirement.ID, &requirement.BranchTypeID, &requirement.StaffGroupID, &requirement.DayOfWeek,
			&requirement.MinimumStaffCount, &requirement.IsActive, &requirement.CreatedAt, &requirement.UpdatedAt,
		); err != nil {
			return nil, err
		}
		requirements = append(requirements, requirement)
	}

	return requirements, rows.Err()
}

func (r *branchTypeStaffGroupRequirementRepository) GetByStaffGroupID(staffGroupID uuid.UUID) ([]*models.BranchTypeStaffGroupRequirement, error) {
	query := `SELECT id, branch_type_id, staff_group_id, day_of_week, minimum_staff_count, is_active, created_at, updated_at
	          FROM branch_type_staff_group_requirements WHERE staff_group_id = $1 ORDER BY branch_type_id, day_of_week`

	rows, err := r.db.Query(query, staffGroupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requirements := []*models.BranchTypeStaffGroupRequirement{}
	for rows.Next() {
		requirement := &models.BranchTypeStaffGroupRequirement{}
		if err := rows.Scan(
			&requirement.ID, &requirement.BranchTypeID, &requirement.StaffGroupID, &requirement.DayOfWeek,
			&requirement.MinimumStaffCount, &requirement.IsActive, &requirement.CreatedAt, &requirement.UpdatedAt,
		); err != nil {
			return nil, err
		}
		requirements = append(requirements, requirement)
	}

	return requirements, rows.Err()
}

func (r *branchTypeStaffGroupRequirementRepository) Update(requirement *models.BranchTypeStaffGroupRequirement) error {
	requirement.UpdatedAt = time.Now()

	query := `UPDATE branch_type_staff_group_requirements
	          SET minimum_staff_count = $1, is_active = $2, updated_at = $3
	          WHERE id = $4 RETURNING updated_at`

	return r.db.QueryRow(query,
		requirement.MinimumStaffCount, requirement.IsActive, requirement.UpdatedAt, requirement.ID,
	).Scan(&requirement.UpdatedAt)
}

func (r *branchTypeStaffGroupRequirementRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM branch_type_staff_group_requirements WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *branchTypeStaffGroupRequirementRepository) BulkUpsert(requirements []*models.BranchTypeStaffGroupRequirement) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO branch_type_staff_group_requirements (id, branch_type_id, staff_group_id, day_of_week, minimum_staff_count, is_active, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	          ON CONFLICT (branch_type_id, staff_group_id, day_of_week) 
	          DO UPDATE SET minimum_staff_count = EXCLUDED.minimum_staff_count, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP
	          RETURNING created_at, updated_at`

	for _, requirement := range requirements {
		if requirement.ID == uuid.Nil {
			requirement.ID = uuid.New()
		}
		err := tx.QueryRow(query, requirement.ID, requirement.BranchTypeID, requirement.StaffGroupID, requirement.DayOfWeek,
			requirement.MinimumStaffCount, requirement.IsActive).
			Scan(&requirement.CreatedAt, &requirement.UpdatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
