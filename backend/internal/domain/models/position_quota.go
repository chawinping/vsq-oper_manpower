package models

import (
	"time"

	"github.com/google/uuid"
)

// PositionQuota represents the designated quota for a position at a branch
type PositionQuota struct {
	ID              uuid.UUID `json:"id" db:"id"`
	BranchID        uuid.UUID `json:"branch_id" db:"branch_id"`
	Branch          *Branch   `json:"branch,omitempty"`
	PositionID      uuid.UUID `json:"position_id" db:"position_id"`
	Position        *Position `json:"position,omitempty"`
	DesignatedQuota int       `json:"designated_quota" db:"designated_quota"` // Preferred/target quota
	MinimumRequired int       `json:"minimum_required" db:"minimum_required"` // Minimum required staff
	IsActive        bool      `json:"is_active" db:"is_active"`
	CreatedBy       uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}
