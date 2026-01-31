package models

import (
	"time"

	"github.com/google/uuid"
)

// BranchQuotaSummary represents pre-computed quota status for a branch on a specific date
type BranchQuotaSummary struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	BranchID           uuid.UUID `json:"branch_id" db:"branch_id"`
	Date               time.Time `json:"date" db:"date"`
	TotalDesignated    int       `json:"total_designated" db:"total_designated"`
	TotalAvailable     int       `json:"total_available" db:"total_available"`
	TotalAssigned      int       `json:"total_assigned" db:"total_assigned"`
	TotalRequired      int       `json:"total_required" db:"total_required"`
	Group1Score        int       `json:"group1_score" db:"group1_score"`
	Group2Score        int       `json:"group2_score" db:"group2_score"`
	Group3Score        int       `json:"group3_score" db:"group3_score"`
	Group1MissingStaff []string  `json:"group1_missing_staff" db:"group1_missing_staff"`
	Group2MissingStaff []string  `json:"group2_missing_staff" db:"group2_missing_staff"`
	Group3MissingStaff []string  `json:"group3_missing_staff" db:"group3_missing_staff"`
	CalculatedAt       time.Time `json:"calculated_at" db:"calculated_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// PositionQuotaSummary represents pre-computed quota status for a position on a specific date
type PositionQuotaSummary struct {
	ID                uuid.UUID `json:"id" db:"id"`
	BranchID          uuid.UUID `json:"branch_id" db:"branch_id"`
	PositionID        uuid.UUID `json:"position_id" db:"position_id"`
	Date              time.Time `json:"date" db:"date"`
	DesignatedQuota   int       `json:"designated_quota" db:"designated_quota"`
	MinimumRequired   int       `json:"minimum_required" db:"minimum_required"`
	AvailableLocal    int       `json:"available_local" db:"available_local"`
	AssignedRotation  int       `json:"assigned_rotation" db:"assigned_rotation"`
	TotalAssigned     int       `json:"total_assigned" db:"total_assigned"`
	StillRequired     int       `json:"still_required" db:"still_required"`
	CalculatedAt      time.Time `json:"calculated_at" db:"calculated_at"`
}
