package postgres

import (
	"database/sql"

	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"

	"github.com/google/uuid"
)

type branchTypeConstraintStaffGroupRepository struct {
	db *sql.DB
}

func NewBranchTypeConstraintStaffGroupRepository(db *sql.DB) interfaces.BranchTypeConstraintStaffGroupRepository {
	return &branchTypeConstraintStaffGroupRepository{db: db}
}

func (r *branchTypeConstraintStaffGroupRepository) Create(sg *models.BranchTypeConstraintStaffGroup) error {
	query := `INSERT INTO branch_type_constraint_staff_groups (id, branch_type_constraint_id, staff_group_id, minimum_count) 
	          VALUES ($1, $2, $3, $4) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, sg.ID, sg.BranchTypeConstraintID, sg.StaffGroupID, sg.MinimumCount).
		Scan(&sg.CreatedAt, &sg.UpdatedAt)
}

func (r *branchTypeConstraintStaffGroupRepository) GetByConstraintID(constraintID uuid.UUID) ([]*models.BranchTypeConstraintStaffGroup, error) {
	query := `SELECT id, branch_type_constraint_id, staff_group_id, minimum_count, created_at, updated_at 
	          FROM branch_type_constraint_staff_groups 
	          WHERE branch_type_constraint_id = $1 
	          ORDER BY staff_group_id`
	rows, err := r.db.Query(query, constraintID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var staffGroups []*models.BranchTypeConstraintStaffGroup
	for rows.Next() {
		sg := &models.BranchTypeConstraintStaffGroup{}
		if err := rows.Scan(
			&sg.ID, &sg.BranchTypeConstraintID, &sg.StaffGroupID, &sg.MinimumCount,
			&sg.CreatedAt, &sg.UpdatedAt,
		); err != nil {
			return nil, err
		}
		staffGroups = append(staffGroups, sg)
	}
	return staffGroups, rows.Err()
}

func (r *branchTypeConstraintStaffGroupRepository) GetByBranchTypeID(branchTypeID uuid.UUID) ([]*models.BranchTypeConstraintStaffGroup, error) {
	query := `SELECT sg.id, sg.branch_type_constraint_id, sg.staff_group_id, sg.minimum_count, sg.created_at, sg.updated_at 
	          FROM branch_type_constraint_staff_groups sg
	          INNER JOIN branch_type_constraints c ON sg.branch_type_constraint_id = c.id
	          WHERE c.branch_type_id = $1 
	          ORDER BY c.day_of_week, sg.staff_group_id`
	rows, err := r.db.Query(query, branchTypeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var staffGroups []*models.BranchTypeConstraintStaffGroup
	for rows.Next() {
		sg := &models.BranchTypeConstraintStaffGroup{}
		if err := rows.Scan(
			&sg.ID, &sg.BranchTypeConstraintID, &sg.StaffGroupID, &sg.MinimumCount,
			&sg.CreatedAt, &sg.UpdatedAt,
		); err != nil {
			return nil, err
		}
		staffGroups = append(staffGroups, sg)
	}
	return staffGroups, rows.Err()
}

func (r *branchTypeConstraintStaffGroupRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM branch_type_constraint_staff_groups WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *branchTypeConstraintStaffGroupRepository) DeleteByConstraintID(constraintID uuid.UUID) error {
	query := `DELETE FROM branch_type_constraint_staff_groups WHERE branch_type_constraint_id = $1`
	_, err := r.db.Exec(query, constraintID)
	return err
}

func (r *branchTypeConstraintStaffGroupRepository) BulkUpsert(staffGroups []*models.BranchTypeConstraintStaffGroup) error {
	if len(staffGroups) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO branch_type_constraint_staff_groups (id, branch_type_constraint_id, staff_group_id, minimum_count, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	          ON CONFLICT (branch_type_constraint_id, staff_group_id) 
	          DO UPDATE SET minimum_count = EXCLUDED.minimum_count, updated_at = CURRENT_TIMESTAMP
	          RETURNING created_at, updated_at`

	for _, sg := range staffGroups {
		if sg.ID == uuid.Nil {
			sg.ID = uuid.New()
		}
		err := tx.QueryRow(query, sg.ID, sg.BranchTypeConstraintID, sg.StaffGroupID, sg.MinimumCount).
			Scan(&sg.CreatedAt, &sg.UpdatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
