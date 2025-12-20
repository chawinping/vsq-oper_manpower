package models

import (
	"time"

	"github.com/google/uuid"
)

type StaffType string

const (
	StaffTypeBranch   StaffType = "branch"
	StaffTypeRotation StaffType = "rotation"
)

type Staff struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	StaffType    StaffType `json:"staff_type" db:"staff_type"`
	PositionID   uuid.UUID `json:"position_id" db:"position_id"`
	Position     *Position `json:"position,omitempty"`
	BranchID     *uuid.UUID `json:"branch_id,omitempty" db:"branch_id"`
	Branch       *Branch    `json:"branch,omitempty"`
	CoverageArea string     `json:"coverage_area" db:"coverage_area"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type Position struct {
	ID                uuid.UUID `json:"id" db:"id"`
	Name              string    `json:"name" db:"name"`
	MinStaffPerBranch int       `json:"min_staff_per_branch" db:"min_staff_per_branch"`
	RevenueMultiplier float64   `json:"revenue_multiplier" db:"revenue_multiplier"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

type EffectiveBranch struct {
	ID            uuid.UUID `json:"id" db:"id"`
	RotationStaffID uuid.UUID `json:"rotation_staff_id" db:"rotation_staff_id"`
	BranchID      uuid.UUID `json:"branch_id" db:"branch_id"`
	Branch        *Branch   `json:"branch,omitempty"`
	Level         int       `json:"level" db:"level"` // 1 = priority, 2 = reserved
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}



