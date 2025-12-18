package models

import (
	"time"

	"github.com/google/uuid"
)

type Branch struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	Name            string     `json:"name" db:"name"`
	Code            string     `json:"code" db:"code"`
	Address         string     `json:"address" db:"address"`
	AreaManagerID   *uuid.UUID `json:"area_manager_id,omitempty" db:"area_manager_id"`
	AreaManager     *User      `json:"area_manager,omitempty"`
	ExpectedRevenue float64    `json:"expected_revenue" db:"expected_revenue"`
	Priority        int        `json:"priority" db:"priority"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

type RevenueData struct {
	ID              uuid.UUID `json:"id" db:"id"`
	BranchID        uuid.UUID `json:"branch_id" db:"branch_id"`
	Branch          *Branch   `json:"branch,omitempty"`
	Date            time.Time `json:"date" db:"date"`
	ExpectedRevenue float64  `json:"expected_revenue" db:"expected_revenue"`
	ActualRevenue   *float64  `json:"actual_revenue,omitempty" db:"actual_revenue"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}


