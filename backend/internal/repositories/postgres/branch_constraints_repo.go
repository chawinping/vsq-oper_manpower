package postgres

import (
	"database/sql"

	"vsq-oper-manpower/backend/internal/domain/interfaces"
	"vsq-oper-manpower/backend/internal/domain/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type branchConstraintsRepository struct {
	db             *sql.DB
	staffGroupRepo interfaces.BranchConstraintStaffGroupRepository
}

func NewBranchConstraintsRepository(db *sql.DB) interfaces.BranchConstraintsRepository {
	return &branchConstraintsRepository{
		db:             db,
		staffGroupRepo: NewBranchConstraintStaffGroupRepository(db),
	}
}

func (r *branchConstraintsRepository) Create(constraint *models.BranchConstraints) error {
	query := `INSERT INTO branch_constraints (id, branch_id, day_of_week, min_front_staff, min_managers, min_doctor_assistant, min_total_staff, inherited_from_branch_type_id, is_overridden) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING created_at, updated_at`
	return r.db.QueryRow(query, constraint.ID, constraint.BranchID, constraint.DayOfWeek,
		constraint.MinFrontStaff, constraint.MinManagers, constraint.MinDoctorAssistant, constraint.MinTotalStaff,
		constraint.InheritedFromBranchTypeID, constraint.IsOverridden).
		Scan(&constraint.CreatedAt, &constraint.UpdatedAt)
}

func (r *branchConstraintsRepository) Update(constraint *models.BranchConstraints) error {
	query := `UPDATE branch_constraints SET min_front_staff = $1, min_managers = $2, min_doctor_assistant = $3, min_total_staff = $4, inherited_from_branch_type_id = $5, is_overridden = $6, updated_at = CURRENT_TIMESTAMP 
	          WHERE id = $7 RETURNING updated_at`
	return r.db.QueryRow(query, constraint.MinFrontStaff, constraint.MinManagers, constraint.MinDoctorAssistant, constraint.MinTotalStaff,
		constraint.InheritedFromBranchTypeID, constraint.IsOverridden, constraint.ID).
		Scan(&constraint.UpdatedAt)
}

func (r *branchConstraintsRepository) GetByID(id uuid.UUID) (*models.BranchConstraints, error) {
	constraint := &models.BranchConstraints{}
	query := `SELECT id, branch_id, day_of_week, min_front_staff, min_managers, min_doctor_assistant, min_total_staff, inherited_from_branch_type_id, is_overridden, created_at, updated_at 
	          FROM branch_constraints WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&constraint.ID, &constraint.BranchID, &constraint.DayOfWeek,
		&constraint.MinFrontStaff, &constraint.MinManagers, &constraint.MinDoctorAssistant, &constraint.MinTotalStaff,
		&constraint.InheritedFromBranchTypeID, &constraint.IsOverridden,
		&constraint.CreatedAt, &constraint.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return constraint, err
}

func (r *branchConstraintsRepository) GetByBranchID(branchID uuid.UUID) ([]*models.BranchConstraints, error) {
	query := `SELECT id, branch_id, day_of_week, min_front_staff, min_managers, min_doctor_assistant, min_total_staff, inherited_from_branch_type_id, is_overridden, created_at, updated_at 
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
			&constraint.InheritedFromBranchTypeID, &constraint.IsOverridden,
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
	query := `SELECT id, branch_id, day_of_week, min_front_staff, min_managers, min_doctor_assistant, min_total_staff, inherited_from_branch_type_id, is_overridden, created_at, updated_at 
	          FROM branch_constraints WHERE branch_id = $1 AND day_of_week = $2`
	err := r.db.QueryRow(query, branchID, dayOfWeek).Scan(
		&constraint.ID, &constraint.BranchID, &constraint.DayOfWeek,
		&constraint.MinFrontStaff, &constraint.MinManagers, &constraint.MinDoctorAssistant, &constraint.MinTotalStaff,
		&constraint.InheritedFromBranchTypeID, &constraint.IsOverridden,
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

	query := `INSERT INTO branch_constraints (id, branch_id, day_of_week, min_front_staff, min_managers, min_doctor_assistant, min_total_staff, inherited_from_branch_type_id, is_overridden, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	          ON CONFLICT (branch_id, day_of_week) 
	          DO UPDATE SET min_front_staff = EXCLUDED.min_front_staff, min_managers = EXCLUDED.min_managers, min_doctor_assistant = EXCLUDED.min_doctor_assistant, min_total_staff = EXCLUDED.min_total_staff, inherited_from_branch_type_id = EXCLUDED.inherited_from_branch_type_id, is_overridden = EXCLUDED.is_overridden, updated_at = CURRENT_TIMESTAMP
	          RETURNING created_at, updated_at`

	for _, constraint := range constraints {
		if constraint.ID == uuid.Nil {
			constraint.ID = uuid.New()
		}
		err := tx.QueryRow(query, constraint.ID, constraint.BranchID, constraint.DayOfWeek,
			constraint.MinFrontStaff, constraint.MinManagers, constraint.MinDoctorAssistant, constraint.MinTotalStaff,
			constraint.InheritedFromBranchTypeID, constraint.IsOverridden).
			Scan(&constraint.CreatedAt, &constraint.UpdatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *branchConstraintsRepository) LoadStaffGroupRequirements(constraints []*models.BranchConstraints) error {
	if len(constraints) == 0 {
		return nil
	}

	// Collect all constraint IDs
	constraintIDs := make([]uuid.UUID, 0, len(constraints))
	constraintMap := make(map[uuid.UUID]*models.BranchConstraints)
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
	query := `SELECT id, branch_constraint_id, staff_group_id, minimum_count, created_at, updated_at 
	          FROM branch_constraint_staff_groups 
	          WHERE branch_constraint_id = ANY($1) 
	          ORDER BY branch_constraint_id, staff_group_id`
	rows, err := r.db.Query(query, pq.Array(constraintIDs))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		sg := &models.BranchConstraintStaffGroup{}
		if err := rows.Scan(
			&sg.ID, &sg.BranchConstraintID, &sg.StaffGroupID, &sg.MinimumCount,
			&sg.CreatedAt, &sg.UpdatedAt,
		); err != nil {
			return err
		}

		if constraint, exists := constraintMap[sg.BranchConstraintID]; exists {
			if constraint.StaffGroupRequirements == nil {
				constraint.StaffGroupRequirements = make([]*models.BranchConstraintStaffGroup, 0)
			}
			constraint.StaffGroupRequirements = append(constraint.StaffGroupRequirements, sg)
		}
	}

	return rows.Err()
}

func (r *branchConstraintsRepository) BulkUpsertWithStaffGroups(constraints []*models.BranchConstraints) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// First, upsert the constraints themselves (with deprecated fields set to 0)
	constraintQuery := `INSERT INTO branch_constraints (id, branch_id, day_of_week, min_front_staff, min_managers, min_doctor_assistant, min_total_staff, inherited_from_branch_type_id, is_overridden, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	          ON CONFLICT (branch_id, day_of_week) 
	          DO UPDATE SET min_front_staff = EXCLUDED.min_front_staff, min_managers = EXCLUDED.min_managers, min_doctor_assistant = EXCLUDED.min_doctor_assistant, min_total_staff = EXCLUDED.min_total_staff, inherited_from_branch_type_id = EXCLUDED.inherited_from_branch_type_id, is_overridden = EXCLUDED.is_overridden, updated_at = CURRENT_TIMESTAMP
	          RETURNING id, created_at, updated_at`

	for _, constraint := range constraints {
		if constraint.ID == uuid.Nil {
			constraint.ID = uuid.New()
		}
		err := tx.QueryRow(constraintQuery, constraint.ID, constraint.BranchID, constraint.DayOfWeek,
			0, 0, 0, 0, // Deprecated fields set to 0
			constraint.InheritedFromBranchTypeID, constraint.IsOverridden).
			Scan(&constraint.ID, &constraint.CreatedAt, &constraint.UpdatedAt)
		if err != nil {
			return err
		}

		// Delete existing staff group requirements for this constraint
		deleteQuery := `DELETE FROM branch_constraint_staff_groups WHERE branch_constraint_id = $1`
		if _, err := tx.Exec(deleteQuery, constraint.ID); err != nil {
			return err
		}

		// Insert new staff group requirements
		if len(constraint.StaffGroupRequirements) > 0 {
			staffGroupQuery := `INSERT INTO branch_constraint_staff_groups (id, branch_constraint_id, staff_group_id, minimum_count, created_at, updated_at)
			          VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			          RETURNING created_at, updated_at`
			for _, sg := range constraint.StaffGroupRequirements {
				if sg.ID == uuid.Nil {
					sg.ID = uuid.New()
				}
				sg.BranchConstraintID = constraint.ID
				err := tx.QueryRow(staffGroupQuery, sg.ID, sg.BranchConstraintID, sg.StaffGroupID, sg.MinimumCount).
					Scan(&sg.CreatedAt, &sg.UpdatedAt)
				if err != nil {
					return err
				}
			}
		}
	}

	return tx.Commit()
}
