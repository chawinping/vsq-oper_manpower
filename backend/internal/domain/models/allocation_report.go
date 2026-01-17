package models

import (
	"time"

	"github.com/google/uuid"
)

// AllocationReport represents a report generated for each automatic allocation iteration
type AllocationReport struct {
	ID                    uuid.UUID                `json:"id" db:"id"`
	IterationID           uuid.UUID                `json:"iteration_id" db:"iteration_id"` // Unique ID for this allocation iteration
	StartDate             time.Time                 `json:"start_date" db:"start_date"`
	EndDate               time.Time                 `json:"end_date" db:"end_date"`
	BranchesCovered       int                       `json:"branches_covered" db:"branches_covered"`
	TotalAssignments      int                       `json:"total_assignments" db:"total_assignments"`
	TotalPositionsFilled  int                       `json:"total_positions_filled" db:"total_positions_filled"`
	TotalPositionsNeeded  int                       `json:"total_positions_needed" db:"total_positions_needed"`
	OverallFulfillmentRate float64                  `json:"overall_fulfillment_rate" db:"overall_fulfillment_rate"`
	AverageConfidenceScore float64                  `json:"average_confidence_score" db:"average_confidence_score"`
	CreatedAt             time.Time                 `json:"created_at" db:"created_at"`
	CreatedBy             uuid.UUID                 `json:"created_by" db:"created_by"`
	// Embedded assignment details (stored as JSON or retrieved via joins)
	AssignmentDetails     []*AllocationReportAssignment `json:"assignment_details,omitempty"`
	// Gap analysis (stored as JSON or retrieved via joins)
	GapAnalysis           []*AllocationReportGap        `json:"gap_analysis,omitempty"`
}

// AllocationReportAssignment represents a single assignment detail in the report
type AllocationReportAssignment struct {
	ID                    uuid.UUID       `json:"id"`
	RotationStaffID       uuid.UUID       `json:"rotation_staff_id"`
	RotationStaffName     string          `json:"rotation_staff_name"`
	BranchID              uuid.UUID       `json:"branch_id"`
	BranchName            string          `json:"branch_name"`
	BranchCode            string          `json:"branch_code"`
	Date                  time.Time       `json:"date"`
	PositionID            uuid.UUID       `json:"position_id"`
	PositionName          string          `json:"position_name"`
	Reason                string          `json:"reason"`                // Detailed reason for assignment
	CriteriaUsed          string          `json:"criteria_used"`         // JSON array of criteria IDs used
	ConfidenceScore       float64         `json:"confidence_score"`
	IsOverridden          bool            `json:"is_overridden"`
	OverrideReason        string          `json:"override_reason,omitempty"`
	OverrideBy            *uuid.UUID       `json:"override_by,omitempty"`
	OverrideAt            *time.Time       `json:"override_at,omitempty"`
	Status                string          `json:"status"`                // approved, rejected, overridden
}

// AllocationReportGap represents gap analysis for positions that still need staff
type AllocationReportGap struct {
	BranchID              uuid.UUID       `json:"branch_id"`
	BranchName            string          `json:"branch_name"`
	BranchCode            string          `json:"branch_code"`
	Date                  time.Time       `json:"date"`
	PositionID            uuid.UUID       `json:"position_id"`
	PositionName          string          `json:"position_name"`
	RequiredStaffCount    int             `json:"required_staff_count"`    // From quota/criteria
	AvailableLocalStaff   int             `json:"available_local_staff"`  // Local branch staff available
	AssignedRotationStaff int             `json:"assigned_rotation_staff"` // Rotation staff assigned
	StillRequiredStaff    int             `json:"still_required_staff"`   // Still needed to satisfy criteria
}
