package models

import (
	"time"

	"github.com/google/uuid"
)

// BranchTypeStaffGroupRequirement represents minimum staff requirements for a staff group within a branch type per day
type BranchTypeStaffGroupRequirement struct {
	ID                uuid.UUID   `json:"id" db:"id"`
	BranchTypeID      uuid.UUID   `json:"branch_type_id" db:"branch_type_id"`
	BranchType        *BranchType `json:"branch_type,omitempty"`
	StaffGroupID      uuid.UUID   `json:"staff_group_id" db:"staff_group_id"`
	StaffGroup        *StaffGroup `json:"staff_group,omitempty"`
	DayOfWeek         int         `json:"day_of_week" db:"day_of_week"` // 0=Sunday, 1=Monday, ..., 6=Saturday
	MinimumStaffCount int         `json:"minimum_staff_count" db:"minimum_staff_count"`
	IsActive          bool        `json:"is_active" db:"is_active"`
	CreatedAt         time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at" db:"updated_at"`
}
