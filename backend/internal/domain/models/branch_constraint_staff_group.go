package models

import (
	"time"

	"github.com/google/uuid"
)

// BranchConstraintStaffGroup represents the relationship between
// a branch constraint and a staff group with its minimum count requirement
type BranchConstraintStaffGroup struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	BranchConstraintID uuid.UUID `json:"branch_constraint_id" db:"branch_constraint_id"`
	StaffGroupID       uuid.UUID `json:"staff_group_id" db:"staff_group_id"`
	MinimumCount       int       `json:"minimum_count" db:"minimum_count"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
	// Related entities (loaded separately)
	StaffGroup *StaffGroup `json:"staff_group,omitempty"`
}
