package models

import (
	"time"

	"github.com/google/uuid"
)

// BranchWeeklyRevenue represents expected revenue for each day of the week for a branch
type BranchWeeklyRevenue struct {
	ID              uuid.UUID `json:"id" db:"id"`
	BranchID        uuid.UUID `json:"branch_id" db:"branch_id"`
	Branch          *Branch   `json:"branch,omitempty"`
	DayOfWeek       int       `json:"day_of_week" db:"day_of_week"` // 0=Sunday, 1=Monday, ..., 6=Saturday
	ExpectedRevenue float64   `json:"expected_revenue" db:"expected_revenue"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}
