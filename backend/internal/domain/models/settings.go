package models

import (
	"time"

	"github.com/google/uuid"
)

type SystemSetting struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Key         string    `json:"key" db:"key"`
	Value       string    `json:"value" db:"value"`
	Description string    `json:"description" db:"description"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type StaffAllocationRule struct {
	ID                uuid.UUID `json:"id" db:"id"`
	PositionID        uuid.UUID `json:"position_id" db:"position_id"`
	Position          *Position  `json:"position,omitempty"`
	MinStaff          int        `json:"min_staff" db:"min_staff"`
	RevenueThreshold  float64   `json:"revenue_threshold" db:"revenue_threshold"`
	StaffCountFormula string    `json:"staff_count_formula" db:"staff_count_formula"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}


