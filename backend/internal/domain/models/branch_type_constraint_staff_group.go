package models

import (
	"time"

	"github.com/google/uuid"
)

// BranchTypeConstraintStaffGroup represents the relationship between
// a branch type constraint and a staff group with its minimum count requirement
type BranchTypeConstraintStaffGroup struct {
	ID                     uuid.UUID `json:"id" db:"id"`
	BranchTypeConstraintID uuid.UUID `json:"branch_type_constraint_id" db:"branch_type_constraint_id"`
	StaffGroupID           uuid.UUID `json:"staff_group_id" db:"staff_group_id"`
	MinimumCount           int       `json:"minimum_count" db:"minimum_count"`
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
	// Related entities (loaded separately)
	StaffGroup *StaffGroup `json:"staff_group,omitempty"`
}
