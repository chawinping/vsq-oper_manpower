package postgres

import (
	"database/sql"

	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"

	"github.com/google/uuid"
)

type branchConstraintStaffGroupRepository struct {
	db *sql.DB
}

func NewBranchConstraintStaffGroupRepository(db *sql.DB) interfaces.BranchConstraintStaffGroupRepository {
	return &branchConstraintStaffGroupRepository{db: db}
}

func (r *branchConstraintStaffGroupRepository) Create(sg *models.BranchConstraintStaffGroup) error {
	query := `INSERT INTO branch_constraint_staff_groups (id, branch_constraint_id, staff_group_id, minimum_count) 
	          VALUES ($1, $2, $3, $4) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, sg.ID, sg.BranchConstraintID, sg.StaffGroupID, sg.MinimumCount).
		Scan(&sg.CreatedAt, &sg.UpdatedAt)
}

func (r *branchConstraintStaffGroupRepository) GetByConstraintID(constraintID uuid.UUID) ([]*models.BranchConstraintStaffGroup, error) {
	query := `SELECT id, branch_constraint_id, staff_group_id, minimum_count, created_at, updated_at 
	          FROM branch_constraint_staff_groups 
	          WHERE branch_constraint_id = $1 
	          ORDER BY staff_group_id`
	rows, err := r.db.Query(query, constraintID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var staffGroups []*models.BranchConstraintStaffGroup
	for rows.Next() {
		sg := &models.BranchConstraintStaffGroup{}
		if err := rows.Scan(
			&sg.ID, &sg.BranchConstraintID, &sg.StaffGroupID, &sg.MinimumCount,
			&sg.CreatedAt, &sg.UpdatedAt,
		); err != nil {
			return nil, err
		}
		staffGroups = append(staffGroups, sg)
	}
	return staffGroups, rows.Err()
}

func (r *branchConstraintStaffGroupRepository) GetByBranchID(branchID uuid.UUID) ([]*models.BranchConstraintStaffGroup, error) {
	query := `SELECT sg.id, sg.branch_constraint_id, sg.staff_group_id, sg.minimum_count, sg.created_at, sg.updated_at 
	          FROM branch_constraint_staff_groups sg
	          INNER JOIN branch_constraints c ON sg.branch_constraint_id = c.id
	          WHERE c.branch_id = $1 
	          ORDER BY c.day_of_week, sg.staff_group_id`
	rows, err := r.db.Query(query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var staffGroups []*models.BranchConstraintStaffGroup
	for rows.Next() {
		sg := &models.BranchConstraintStaffGroup{}
		if err := rows.Scan(
			&sg.ID, &sg.BranchConstraintID, &sg.StaffGroupID, &sg.MinimumCount,
			&sg.CreatedAt, &sg.UpdatedAt,
		); err != nil {
			return nil, err
		}
		staffGroups = append(staffGroups, sg)
	}
	return staffGroups, rows.Err()
}

func (r *branchConstraintStaffGroupRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM branch_constraint_staff_groups WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *branchConstraintStaffGroupRepository) DeleteByConstraintID(constraintID uuid.UUID) error {
	query := `DELETE FROM branch_constraint_staff_groups WHERE branch_constraint_id = $1`
	_, err := r.db.Exec(query, constraintID)
	return err
}

func (r *branchConstraintStaffGroupRepository) BulkUpsert(staffGroups []*models.BranchConstraintStaffGroup) error {
	if len(staffGroups) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO branch_constraint_staff_groups (id, branch_constraint_id, staff_group_id, minimum_count, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	          ON CONFLICT (branch_constraint_id, staff_group_id) 
	          DO UPDATE SET minimum_count = EXCLUDED.minimum_count, updated_at = CURRENT_TIMESTAMP
	          RETURNING created_at, updated_at`

	for _, sg := range staffGroups {
		if sg.ID == uuid.Nil {
			sg.ID = uuid.New()
		}
		err := tx.QueryRow(query, sg.ID, sg.BranchConstraintID, sg.StaffGroupID, sg.MinimumCount).
			Scan(&sg.CreatedAt, &sg.UpdatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
