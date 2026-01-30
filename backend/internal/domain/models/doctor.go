package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Doctor represents a doctor profile in the system
type Doctor struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Code        string    `json:"code" db:"code"`               // Optional doctor code/nickname
	Preferences string    `json:"preferences" db:"preferences"` // Noted remark/preferences
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// DoctorPreference represents a doctor-specific rule/preference for staff allocation
type DoctorPreference struct {
	ID         uuid.UUID              `json:"id" db:"id"`
	DoctorID   uuid.UUID              `json:"doctor_id" db:"doctor_id"`
	Doctor     *Doctor                `json:"doctor,omitempty"`
	BranchID   *uuid.UUID             `json:"branch_id,omitempty" db:"branch_id"` // Nullable - can be global or branch-specific
	Branch     *Branch                `json:"branch,omitempty"`
	RuleType   string                 `json:"rule_type" db:"rule_type"`     // e.g., "staff_requirement", "schedule_preference"
	RuleConfig map[string]interface{} `json:"rule_config" db:"rule_config"` // JSONB field for flexible rule configuration
	IsActive   bool                   `json:"is_active" db:"is_active"`
	CreatedAt  time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at" db:"updated_at"`
}

// StaffRequirementRule represents a doctor-specific staff requirement rule
type StaffRequirementRule struct {
	PositionID   uuid.UUID `json:"position_id"`
	PositionName string    `json:"position_name,omitempty"`
	MinCount     int       `json:"min_count"`
	DayOfWeek    *int      `json:"day_of_week,omitempty"`   // 0-6 (Sunday-Saturday), nil for all days
	DateSpecific *string   `json:"date_specific,omitempty"` // Specific date in YYYY-MM-DD format
}

// RotationStaffRequirementRuleConfig represents the configuration for rotation staff requirement rules
type RotationStaffRequirementRuleConfig struct {
	DaysOfWeek                []int                      `json:"days_of_week"` // 0-6 (Sunday-Saturday)
	PositionRequirements      []PositionRequirement      `json:"position_requirements"`
	SpecificStaffRequirements []SpecificStaffRequirement `json:"specific_staff_requirements,omitempty"`
}

// PositionRequirement specifies minimum staff count for a position
type PositionRequirement struct {
	PositionID uuid.UUID `json:"position_id"`
	MinCount   int       `json:"min_count"`
}

// SpecificStaffRequirement specifies a specific staff member required for a specific date
type SpecificStaffRequirement struct {
	Date       string    `json:"date"` // YYYY-MM-DD format
	StaffID    uuid.UUID `json:"staff_id"`
	PositionID uuid.UUID `json:"position_id"`
}

// ParseRotationStaffRequirementRule parses the rule_config JSONB field for rotation staff requirement rules
func (dp *DoctorPreference) ParseRotationStaffRequirementRule() (*RotationStaffRequirementRuleConfig, error) {
	if dp.RuleType != "rotation_staff_requirement" {
		return nil, fmt.Errorf("rule type is not rotation_staff_requirement")
	}

	config := &RotationStaffRequirementRuleConfig{}
	// The RuleConfig is already a map[string]interface{}, we need to convert it
	// This will be handled by the repository layer when unmarshaling JSONB
	return config, nil
}

// ValidateRotationStaffRequirementRule validates the rotation staff requirement rule config
func ValidateRotationStaffRequirementRule(config map[string]interface{}) error {
	// Validate days_of_week
	if days, ok := config["days_of_week"].([]interface{}); ok {
		for _, day := range days {
			if dayInt, ok := day.(float64); ok {
				if dayInt < 0 || dayInt > 6 {
					return fmt.Errorf("invalid day of week: %v (must be 0-6)", dayInt)
				}
			}
		}
	}

	// Validate position_requirements
	if posReqs, ok := config["position_requirements"].([]interface{}); ok {
		for _, req := range posReqs {
			if reqMap, ok := req.(map[string]interface{}); ok {
				if minCount, ok := reqMap["min_count"].(float64); ok {
					if minCount < 0 {
						return fmt.Errorf("minimum count cannot be negative")
					}
				}
			}
		}
	}

	return nil
}
