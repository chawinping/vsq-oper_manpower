package models

import (
	"time"

	"github.com/google/uuid"
)

// RotationStaffBranchPosition represents a mapping allowing a rotation staff member to work in a branch position
type RotationStaffBranchPosition struct {
	ID                uuid.UUID `json:"id" db:"id"`
	RotationStaffID   uuid.UUID `json:"rotation_staff_id" db:"rotation_staff_id"`
	BranchPositionID  uuid.UUID `json:"branch_position_id" db:"branch_position_id"`
	RotationStaff     *Staff    `json:"rotation_staff,omitempty"`
	BranchPosition    *Position `json:"branch_position,omitempty"`
	SubstitutionLevel int       `json:"substitution_level" db:"substitution_level"` // 1 = preferred, 2 = acceptable, 3 = emergency only
	IsActive          bool      `json:"is_active" db:"is_active"`
	Notes             string    `json:"notes,omitempty" db:"notes"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}
