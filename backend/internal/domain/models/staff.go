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

type PositionType string

const (
	PositionTypeBranch   PositionType = "branch"
	PositionTypeRotation PositionType = "rotation"
)

type ManpowerType string

const (
	ManpowerTypeFront      ManpowerType = "พนักงานฟร้อนท์"      // Front/Counter staff
	ManpowerTypeDoctor     ManpowerType = "ผู้ช่วยแพทย์"        // Doctor Assistant
	ManpowerTypeOther      ManpowerType = "อื่นๆ"              // Others
	ManpowerTypeCleaning   ManpowerType = "ทำความสะอาด"        // Cleaning/Housekeeping
)

type Staff struct {
	ID                 uuid.UUID          `json:"id" db:"id"`
	Nickname           string              `json:"nickname" db:"nickname"`
	Name               string              `json:"name" db:"name"` // Full name
	StaffType          StaffType           `json:"staff_type" db:"staff_type"`
	PositionID         uuid.UUID           `json:"position_id" db:"position_id"`
	Position           *Position           `json:"position,omitempty"`
	BranchID           *uuid.UUID          `json:"branch_id,omitempty" db:"branch_id"`
	Branch             *Branch             `json:"branch,omitempty"`
	CoverageArea       string              `json:"coverage_area" db:"coverage_area"` // Legacy field, kept for backward compatibility
	AreaOfOperationID  *uuid.UUID          `json:"area_of_operation_id,omitempty" db:"area_of_operation_id"` // Legacy field, kept for backward compatibility
	AreaOfOperation    *AreaOfOperation    `json:"area_of_operation,omitempty"`
	ZoneID             *uuid.UUID          `json:"zone_id,omitempty" db:"zone_id"` // Zone assignment for rotation staff
	Zone               *Zone               `json:"zone,omitempty"`
	Branches           []*Branch           `json:"branches,omitempty"` // Individual branches for rotation staff (outside zone)
	SkillLevel         int                 `json:"skill_level" db:"skill_level"` // Rating 0-10
	CreatedAt          time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at" db:"updated_at"`
}

type Position struct {
	ID                  uuid.UUID    `json:"id" db:"id"`
	Name                string       `json:"name" db:"name"`
	PositionType        PositionType `json:"position_type" db:"position_type"`
	ManpowerType        ManpowerType `json:"manpower_type" db:"manpower_type"`
	MinStaffPerBranch   int          `json:"min_staff_per_branch,omitempty" db:"min_staff_per_branch"`
	DisplayOrder        int          `json:"display_order" db:"display_order"`
	BranchStaffCount    *int         `json:"branch_staff_count,omitempty" db:"-"`   // Read-only field, count of branch staff allocated to this position
	RotationStaffCount  *int         `json:"rotation_staff_count,omitempty" db:"-"` // Read-only field, count of rotation staff allocated to this position
	CreatedAt           time.Time    `json:"created_at" db:"created_at"`
}

type EffectiveBranch struct {
	ID                    uuid.UUID `json:"id" db:"id"`
	RotationStaffID      uuid.UUID `json:"rotation_staff_id" db:"rotation_staff_id"`
	BranchID             uuid.UUID `json:"branch_id" db:"branch_id"`
	Branch               *Branch   `json:"branch,omitempty"`
	Level                int       `json:"level" db:"level"` // 1 = priority, 2 = reserved
	CommuteDurationMinutes *int     `json:"commute_duration_minutes,omitempty" db:"commute_duration_minutes"` // Travel time in minutes (default: 300)
	TransitCount         *int      `json:"transit_count,omitempty" db:"transit_count"` // Number of transits (default: 10)
	TravelCost           *float64  `json:"travel_cost,omitempty" db:"travel_cost"` // Cost of traveling (default: 1000)
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
}



