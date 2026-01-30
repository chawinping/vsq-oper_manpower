package models

import (
	"time"

	"github.com/google/uuid"
)

// BranchConstraints represents minimum staff constraints for a branch per day
type BranchConstraints struct {
	ID                        uuid.UUID  `json:"id" db:"id"`
	BranchID                  uuid.UUID  `json:"branch_id" db:"branch_id"`
	Branch                    *Branch    `json:"branch,omitempty"`
	DayOfWeek                 int        `json:"day_of_week" db:"day_of_week"`                                               // 0=Sunday, 1=Monday, ..., 6=Saturday
	MinFrontStaff             int        `json:"min_front_staff,omitempty" db:"min_front_staff"`                             // DEPRECATED: Use StaffGroupRequirements instead
	MinManagers               int        `json:"min_managers,omitempty" db:"min_managers"`                                   // DEPRECATED: Use StaffGroupRequirements instead
	MinDoctorAssistant        int        `json:"min_doctor_assistant,omitempty" db:"min_doctor_assistant"`                   // DEPRECATED: Use StaffGroupRequirements instead
	MinTotalStaff             int        `json:"min_total_staff,omitempty" db:"min_total_staff"`                             // DEPRECATED: Use StaffGroupRequirements instead
	InheritedFromBranchTypeID *uuid.UUID `json:"inherited_from_branch_type_id,omitempty" db:"inherited_from_branch_type_id"` // If set, this constraint is inherited from a branch type
	IsOverridden              bool       `json:"is_overridden" db:"is_overridden"`                                           // If true, this constraint overrides the branch type default
	CreatedAt                 time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at" db:"updated_at"`
	// Staff group requirements (loaded separately, not in DB)
	StaffGroupRequirements []*BranchConstraintStaffGroup `json:"staff_group_requirements,omitempty"`
}
