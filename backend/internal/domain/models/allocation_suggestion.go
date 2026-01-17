package models

import (
	"time"

	"github.com/google/uuid"
)

// SuggestionStatus represents the status of an allocation suggestion
type SuggestionStatus string

const (
	SuggestionStatusPending  SuggestionStatus = "pending"
	SuggestionStatusApproved SuggestionStatus = "approved"
	SuggestionStatusRejected SuggestionStatus = "rejected"
	SuggestionStatusModified SuggestionStatus = "modified"
)

// AllocationSuggestion represents a system-generated allocation suggestion
type AllocationSuggestion struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	RotationStaffID uuid.UUID       `json:"rotation_staff_id" db:"rotation_staff_id"`
	RotationStaff   *Staff          `json:"rotation_staff,omitempty"`
	BranchID        uuid.UUID       `json:"branch_id" db:"branch_id"`
	Branch          *Branch          `json:"branch,omitempty"`
	Date            time.Time       `json:"date" db:"date"`
	PositionID      uuid.UUID       `json:"position_id" db:"position_id"`
	Position        *Position       `json:"position,omitempty"`
	Status          SuggestionStatus `json:"status" db:"status"`
	Confidence      float64         `json:"confidence" db:"confidence"` // Confidence score (0.0 - 1.0)
	Reason          string          `json:"reason" db:"reason"`          // Reason for suggestion
	CriteriaUsed   string          `json:"criteria_used" db:"criteria_used"` // JSON array of criteria IDs used
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	ReviewedBy     *uuid.UUID      `json:"reviewed_by,omitempty" db:"reviewed_by"`
	ReviewedAt      *time.Time      `json:"reviewed_at,omitempty" db:"reviewed_at"`
}
