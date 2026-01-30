package postgres

import (
	"database/sql"

	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type branchTypeConstraintsRepository struct {
	db             *sql.DB
	staffGroupRepo interfaces.BranchTypeConstraintStaffGroupRepository
}

func NewBranchTypeConstraintsRepository(db *sql.DB) interfaces.BranchTypeConstraintsRepository {
	return &branchTypeConstraintsRepository{
		db:             db,
		staffGroupRepo: NewBranchTypeConstraintStaffGroupRepository(db),
	}
}

func (r *branchTypeConstraintsRepository) GetByID(id uuid.UUID) (*models.BranchTypeConstraints, error) {
	constraint := &models.BranchTypeConstraints{}
	query := `SELECT id, branch_type_id, day_of_week, created_at, updated_at 
	          FROM branch_type_constraints WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&constraint.ID, &constraint.BranchTypeID, &constraint.DayOfWeek,
		&constraint.CreatedAt, &constraint.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return constraint, err
}

func (r *branchTypeConstraintsRepository) GetByBranchTypeID(branchTypeID uuid.UUID) ([]*models.BranchTypeConstraints, error) {
	query := `SELECT id, branch_type_id, day_of_week, created_at, updated_at 
	          FROM branch_type_constraints WHERE branch_type_id = $1 ORDER BY day_of_week`
	rows, err := r.db.Query(query, branchTypeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var constraints []*models.BranchTypeConstraints
	for rows.Next() {
		constraint := &models.BranchTypeConstraints{}
		if err := rows.Scan(
			&constraint.ID, &constraint.BranchTypeID, &constraint.DayOfWeek,
			&constraint.CreatedAt, &constraint.UpdatedAt,
		); err != nil {
			return nil, err
		}
		constraints = append(constraints, constraint)
	}
	return constraints, rows.Err()
}

func (r *branchTypeConstraintsRepository) GetByBranchTypeIDAndDayOfWeek(branchTypeID uuid.UUID, dayOfWeek int) (*models.BranchTypeConstraints, error) {
	constraint := &models.BranchTypeConstraints{}
	query := `SELECT id, branch_type_id, day_of_week, created_at, updated_at 
	          FROM branch_type_constraints WHERE branch_type_id = $1 AND day_of_week = $2`
	err := r.db.QueryRow(query, branchTypeID, dayOfWeek).Scan(
		&constraint.ID, &constraint.BranchTypeID, &constraint.DayOfWeek,
		&constraint.CreatedAt, &constraint.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return constraint, err
}

func (r *branchTypeConstraintsRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM branch_type_constraints WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *branchTypeConstraintsRepository) LoadStaffGroupRequirements(constraints []*models.BranchTypeConstraints) error {
	if len(constraints) == 0 {
		return nil
	}

	// Collect all constraint IDs
	constraintIDs := make([]uuid.UUID, 0, len(constraints))
	constraintMap := make(map[uuid.UUID]*models.BranchTypeConstraints)
	for _, c := range constraints {
		if c.ID != uuid.Nil {
			constraintIDs = append(constraintIDs, c.ID)
			constraintMap[c.ID] = c
		}
	}

	if len(constraintIDs) == 0 {
		return nil
	}

	// Load all staff group requirements for these constraints
	query := `SELECT id, branch_type_constraint_id, staff_group_id, minimum_count, created_at, updated_at 
	          FROM branch_type_constraint_staff_groups 
	          WHERE branch_type_constraint_id = ANY($1) 
	          ORDER BY branch_type_constraint_id, staff_group_id`
	rows, err := r.db.Query(query, pq.Array(constraintIDs))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		sg := &models.BranchTypeConstraintStaffGroup{}
		if err := rows.Scan(
			&sg.ID, &sg.BranchTypeConstraintID, &sg.StaffGroupID, &sg.MinimumCount,
			&sg.CreatedAt, &sg.UpdatedAt,
		); err != nil {
			return err
		}

		if constraint, exists := constraintMap[sg.BranchTypeConstraintID]; exists {
			if constraint.StaffGroupRequirements == nil {
				constraint.StaffGroupRequirements = make([]*models.BranchTypeConstraintStaffGroup, 0)
			}
			constraint.StaffGroupRequirements = append(constraint.StaffGroupRequirements, sg)
		}
	}

	return rows.Err()
}

func (r *branchTypeConstraintsRepository) BulkUpsertWithStaffGroups(constraints []*models.BranchTypeConstraints) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// First, upsert the constraints themselves
	constraintQuery := `INSERT INTO branch_type_constraints (id, branch_type_id, day_of_week, created_at, updated_at)
	          VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	          ON CONFLICT (branch_type_id, day_of_week) 
	          DO UPDATE SET updated_at = CURRENT_TIMESTAMP
	          RETURNING id, created_at, updated_at`

	for _, constraint := range constraints {
		if constraint.ID == uuid.Nil {
			constraint.ID = uuid.New()
		}
		err := tx.QueryRow(constraintQuery, constraint.ID, constraint.BranchTypeID, constraint.DayOfWeek).
			Scan(&constraint.ID, &constraint.CreatedAt, &constraint.UpdatedAt)
		if err != nil {
			return err
		}

		// Delete existing staff group requirements for this constraint
		deleteQuery := `DELETE FROM branch_type_constraint_staff_groups WHERE branch_type_constraint_id = $1`
		if _, err := tx.Exec(deleteQuery, constraint.ID); err != nil {
			return err
		}

		// Insert new staff group requirements
		if len(constraint.StaffGroupRequirements) > 0 {
			staffGroupQuery := `INSERT INTO branch_type_constraint_staff_groups (id, branch_type_constraint_id, staff_group_id, minimum_count, created_at, updated_at)
			          VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			          RETURNING created_at, updated_at`
			for _, sg := range constraint.StaffGroupRequirements {
				if sg.ID == uuid.Nil {
					sg.ID = uuid.New()
				}
				sg.BranchTypeConstraintID = constraint.ID
				err := tx.QueryRow(staffGroupQuery, sg.ID, sg.BranchTypeConstraintID, sg.StaffGroupID, sg.MinimumCount).
					Scan(&sg.CreatedAt, &sg.UpdatedAt)
				if err != nil {
					return err
				}
			}
		}
	}

	return tx.Commit()
}
