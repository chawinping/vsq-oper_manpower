package models

import (
	"time"

	"github.com/google/uuid"
)

// Doctor represents a doctor profile in the system
type Doctor struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Code        string    `json:"code" db:"code"` // Optional doctor code/nickname
	Preferences string    `json:"preferences" db:"preferences"` // Noted remark/preferences
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// DoctorPreference represents a doctor-specific rule/preference for staff allocation
type DoctorPreference struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	DoctorID    uuid.UUID              `json:"doctor_id" db:"doctor_id"`
	Doctor      *Doctor                `json:"doctor,omitempty"`
	BranchID    *uuid.UUID             `json:"branch_id,omitempty" db:"branch_id"` // Nullable - can be global or branch-specific
	Branch      *Branch                 `json:"branch,omitempty"`
	RuleType    string                 `json:"rule_type" db:"rule_type"` // e.g., "staff_requirement", "schedule_preference"
	RuleConfig  map[string]interface{} `json:"rule_config" db:"rule_config"` // JSONB field for flexible rule configuration
	IsActive    bool                   `json:"is_active" db:"is_active"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// StaffRequirementRule represents a doctor-specific staff requirement rule
type StaffRequirementRule struct {
	PositionID    uuid.UUID `json:"position_id"`
	PositionName  string    `json:"position_name,omitempty"`
	MinCount      int       `json:"min_count"`
	DayOfWeek     *int      `json:"day_of_week,omitempty"` // 0-6 (Sunday-Saturday), nil for all days
	DateSpecific   *string   `json:"date_specific,omitempty"` // Specific date in YYYY-MM-DD format
}
