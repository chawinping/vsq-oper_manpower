package models

import (
	"time"

	"github.com/google/uuid"
)

type Branch struct {
	ID            uuid.UUID   `json:"id" db:"id"`
	Name          string      `json:"name" db:"name"`
	Code          string      `json:"code" db:"code"`
	AreaManagerID *uuid.UUID  `json:"area_manager_id,omitempty" db:"area_manager_id"`
	AreaManager   *User       `json:"area_manager,omitempty"`
	BranchTypeID  *uuid.UUID  `json:"branch_type_id,omitempty" db:"branch_type_id"`
	BranchType    *BranchType `json:"branch_type,omitempty"`
	Priority      int         `json:"priority" db:"priority"`
	CreatedAt     time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at" db:"updated_at"`
}

type RevenueData struct {
	ID       uuid.UUID `json:"id" db:"id"`
	BranchID uuid.UUID `json:"branch_id" db:"branch_id"`
	Branch   *Branch   `json:"branch,omitempty"`
	Date     time.Time `json:"date" db:"date"`

	// Deprecated: Use SkinRevenue instead. Kept for backward compatibility.
	ExpectedRevenue float64 `json:"expected_revenue,omitempty" db:"expected_revenue"`

	// New revenue type fields
	SkinRevenue  float64 `json:"skin_revenue" db:"skin_revenue"`     // Skin revenue (THB)
	LSHMRevenue  float64 `json:"ls_hm_revenue" db:"ls_hm_revenue"`   // LS HM revenue (THB)
	VitaminCases int     `json:"vitamin_cases" db:"vitamin_cases"`   // Vitamin cases (count)
	SlimPenCases int     `json:"slim_pen_cases" db:"slim_pen_cases"` // Slim Pen cases (count)

	ActualRevenue *float64  `json:"actual_revenue,omitempty" db:"actual_revenue"`
	RevenueSource string    `json:"revenue_source" db:"revenue_source"` // 'branch', 'doctor', or 'excel'
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}
