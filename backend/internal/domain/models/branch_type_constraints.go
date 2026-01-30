package models

import (
	"time"

	"github.com/google/uuid"
)

// BranchTypeConstraints represents minimum staff constraints for a branch type per day
// These constraints serve as templates that branches can inherit
type BranchTypeConstraints struct {
	ID           uuid.UUID   `json:"id" db:"id"`
	BranchTypeID uuid.UUID   `json:"branch_type_id" db:"branch_type_id"`
	BranchType   *BranchType `json:"branch_type,omitempty"`
	DayOfWeek    int         `json:"day_of_week" db:"day_of_week"` // 0=Sunday, 1=Monday, ..., 6=Saturday
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
	// Staff group requirements (loaded separately, not in DB)
	StaffGroupRequirements []*BranchTypeConstraintStaffGroup `json:"staff_group_requirements,omitempty"`
}
