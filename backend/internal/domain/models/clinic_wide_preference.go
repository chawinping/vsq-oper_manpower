package models

import (
	"time"

	"github.com/google/uuid"
)

// ClinicPreferenceCriteriaType represents the type of criteria for clinic-wide preferences
type ClinicPreferenceCriteriaType string

const (
	ClinicCriteriaTypeSkinRevenue     ClinicPreferenceCriteriaType = "skin_revenue"
	ClinicCriteriaTypeLaserYagRevenue ClinicPreferenceCriteriaType = "laser_yag_revenue"
	ClinicCriteriaTypeIVCases         ClinicPreferenceCriteriaType = "iv_cases"
	ClinicCriteriaTypeSlimPenCases    ClinicPreferenceCriteriaType = "slim_pen_cases"
	ClinicCriteriaTypeDoctorCount     ClinicPreferenceCriteriaType = "doctor_count"
)

// ClinicWidePreference represents a clinic-wide preference configuration
type ClinicWidePreference struct {
	ID                   uuid.UUID                       `json:"id" db:"id"`
	CriteriaType         ClinicPreferenceCriteriaType    `json:"criteria_type" db:"criteria_type"`
	CriteriaName         string                          `json:"criteria_name" db:"criteria_name"`
	MinValue             float64                         `json:"min_value" db:"min_value"`
	MaxValue             *float64                        `json:"max_value,omitempty" db:"max_value"`
	IsActive             bool                            `json:"is_active" db:"is_active"`
	DisplayOrder         int                             `json:"display_order" db:"display_order"`
	Description          *string                         `json:"description,omitempty" db:"description"`
	PositionRequirements []PreferencePositionRequirement `json:"position_requirements,omitempty"`
	CreatedAt            time.Time                       `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time                       `json:"updated_at" db:"updated_at"`
}

// PreferencePositionRequirement represents position requirements for a clinic-wide preference
type PreferencePositionRequirement struct {
	ID             uuid.UUID             `json:"id" db:"id"`
	PreferenceID   uuid.UUID             `json:"preference_id" db:"preference_id"`
	Preference     *ClinicWidePreference `json:"preference,omitempty"`
	PositionID     uuid.UUID             `json:"position_id" db:"position_id"`
	Position       *Position             `json:"position,omitempty"`
	MinimumStaff   int                   `json:"minimum_staff" db:"minimum_staff"`
	PreferredStaff int                   `json:"preferred_staff" db:"preferred_staff"`
	IsActive       bool                  `json:"is_active" db:"is_active"`
	CreatedAt      time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at" db:"updated_at"`
}

// ClinicWidePreferenceCreate represents data for creating a new clinic-wide preference
type ClinicWidePreferenceCreate struct {
	CriteriaType         ClinicPreferenceCriteriaType          `json:"criteria_type" binding:"required"`
	CriteriaName         string                                `json:"criteria_name" binding:"required,max=100"`
	MinValue             float64                               `json:"min_value" binding:"min=0"` // Note: removed "required" to allow 0 as valid value
	MaxValue             *float64                              `json:"max_value,omitempty"`
	IsActive             bool                                  `json:"is_active"`
	DisplayOrder         int                                   `json:"display_order"`
	Description          *string                               `json:"description,omitempty"`
	PositionRequirements []PreferencePositionRequirementCreate `json:"position_requirements,omitempty"`
}

// PreferencePositionRequirementCreate represents data for creating a position requirement
type PreferencePositionRequirementCreate struct {
	PositionID     uuid.UUID `json:"position_id" binding:"required"`
	MinimumStaff   int       `json:"minimum_staff" binding:"min=0"`
	PreferredStaff int       `json:"preferred_staff" binding:"min=0"`
	IsActive       bool      `json:"is_active"`
}

// ClinicWidePreferenceUpdate represents data for updating a clinic-wide preference
type ClinicWidePreferenceUpdate struct {
	CriteriaName *string  `json:"criteria_name,omitempty"`
	MinValue     *float64 `json:"min_value,omitempty"`
	MaxValue     *float64 `json:"max_value,omitempty"`
	IsActive     *bool    `json:"is_active,omitempty"`
	DisplayOrder *int     `json:"display_order,omitempty"`
	Description  *string  `json:"description,omitempty"`
}

// PreferencePositionRequirementUpdate represents data for updating a position requirement
type PreferencePositionRequirementUpdate struct {
	MinimumStaff   *int  `json:"minimum_staff,omitempty"`
	PreferredStaff *int  `json:"preferred_staff,omitempty"`
	IsActive       *bool `json:"is_active,omitempty"`
}

// ClinicPreferenceFilters represents filters for querying clinic-wide preferences
type ClinicPreferenceFilters struct {
	CriteriaType *ClinicPreferenceCriteriaType
	IsActive     *bool
}
