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
	
	// Deprecated: Use SkinRevenue instead. Kept for backward compatibility.
	ExpectedRevenue float64   `json:"expected_revenue,omitempty" db:"expected_revenue"`
	
	// New revenue type fields
	SkinRevenue     float64   `json:"skin_revenue" db:"skin_revenue"`           // Skin revenue (THB)
	LSHMRevenue     float64   `json:"ls_hm_revenue" db:"ls_hm_revenue"`         // LS HM revenue (THB)
	VitaminCases    int       `json:"vitamin_cases" db:"vitamin_cases"`         // Vitamin cases (count)
	SlimPenCases    int       `json:"slim_pen_cases" db:"slim_pen_cases"`       // Slim Pen cases (count)
	
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}
