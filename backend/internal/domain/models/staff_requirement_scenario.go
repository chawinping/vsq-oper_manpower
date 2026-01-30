package models

import (
	"time"

	"github.com/google/uuid"
)

// StaffRequirementScenario represents a staff requirement scenario
type StaffRequirementScenario struct {
	ID                        uuid.UUID                          `json:"id" db:"id"`
	ScenarioName              string                             `json:"scenario_name" db:"scenario_name"`
	Description               *string                            `json:"description,omitempty" db:"description"`
	DoctorID                  *uuid.UUID                         `json:"doctor_id,omitempty" db:"doctor_id"`
	Doctor                    *Doctor                            `json:"doctor,omitempty"`
	BranchID                  *uuid.UUID                         `json:"branch_id,omitempty" db:"branch_id"`
	Branch                    *Branch                            `json:"branch,omitempty"`
	RevenueLevelTierID        *uuid.UUID                         `json:"revenue_level_tier_id,omitempty" db:"revenue_level_tier_id"`
	RevenueLevelTier          *RevenueLevelTier                  `json:"revenue_level_tier,omitempty"`
	MinRevenue                *float64                           `json:"min_revenue,omitempty" db:"min_revenue"`
	MaxRevenue                *float64                           `json:"max_revenue,omitempty" db:"max_revenue"`
	UseDayOfWeekRevenue       bool                               `json:"use_day_of_week_revenue" db:"use_day_of_week_revenue"`
	UseSpecificDateRevenue    bool                               `json:"use_specific_date_revenue" db:"use_specific_date_revenue"`
	DoctorCount               *int                               `json:"doctor_count,omitempty" db:"doctor_count"`
	MinDoctorCount            *int                               `json:"min_doctor_count,omitempty" db:"min_doctor_count"`
	DayOfWeek                 *int                               `json:"day_of_week,omitempty" db:"day_of_week"`
	IsDefault                 bool                               `json:"is_default" db:"is_default"`
	IsActive                  bool                               `json:"is_active" db:"is_active"`
	Priority                  int                                `json:"priority" db:"priority"`
	CreatedAt                 time.Time                          `json:"created_at" db:"created_at"`
	UpdatedAt                 time.Time                          `json:"updated_at" db:"updated_at"`
	PositionRequirements      []ScenarioPositionRequirement      `json:"position_requirements,omitempty"`
	SpecificStaffRequirements []ScenarioSpecificStaffRequirement `json:"specific_staff_requirements,omitempty"`
}

// ScenarioPositionRequirement represents position requirements for a scenario
type ScenarioPositionRequirement struct {
	ID             uuid.UUID `json:"id" db:"id"`
	ScenarioID     uuid.UUID `json:"scenario_id" db:"scenario_id"`
	PositionID     uuid.UUID `json:"position_id" db:"position_id"`
	Position       *Position `json:"position,omitempty"`
	PreferredStaff int       `json:"preferred_staff" db:"preferred_staff"`
	MinimumStaff   int       `json:"minimum_staff" db:"minimum_staff"`
	OverrideBase   bool      `json:"override_base" db:"override_base"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// ScenarioSpecificStaffRequirement represents specific staff requirements for a scenario
type ScenarioSpecificStaffRequirement struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ScenarioID uuid.UUID `json:"scenario_id" db:"scenario_id"`
	StaffID    uuid.UUID `json:"staff_id" db:"staff_id"`
	Staff      *Staff    `json:"staff,omitempty"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// StaffRequirementScenarioCreate represents data for creating a new scenario
type StaffRequirementScenarioCreate struct {
	ScenarioName              string                                   `json:"scenario_name" binding:"required,max=100"`
	Description               *string                                  `json:"description,omitempty"`
	DoctorID                  *uuid.UUID                               `json:"doctor_id,omitempty"`
	BranchID                  *uuid.UUID                               `json:"branch_id,omitempty"`
	RevenueLevelTierID        *uuid.UUID                               `json:"revenue_level_tier_id,omitempty"`
	MinRevenue                *float64                                 `json:"min_revenue,omitempty"`
	MaxRevenue                *float64                                 `json:"max_revenue,omitempty"`
	UseDayOfWeekRevenue       bool                                     `json:"use_day_of_week_revenue"`
	UseSpecificDateRevenue    bool                                     `json:"use_specific_date_revenue"`
	DoctorCount               *int                                     `json:"doctor_count,omitempty"`
	MinDoctorCount            *int                                     `json:"min_doctor_count,omitempty"`
	DayOfWeek                 *int                                     `json:"day_of_week,omitempty"`
	IsDefault                 bool                                     `json:"is_default"`
	IsActive                  bool                                     `json:"is_active"`
	Priority                  int                                      `json:"priority"`
	PositionRequirements      []ScenarioPositionRequirementCreate      `json:"position_requirements,omitempty"`
	SpecificStaffRequirements []ScenarioSpecificStaffRequirementCreate `json:"specific_staff_requirements,omitempty"`
}

// ScenarioPositionRequirementCreate represents data for creating a position requirement
type ScenarioPositionRequirementCreate struct {
	PositionID     uuid.UUID `json:"position_id" binding:"required"`
	PreferredStaff int       `json:"preferred_staff" binding:"min=0"`
	MinimumStaff   int       `json:"minimum_staff" binding:"min=0"`
	OverrideBase   bool      `json:"override_base"`
}

// ScenarioSpecificStaffRequirementCreate represents data for creating a specific staff requirement
type ScenarioSpecificStaffRequirementCreate struct {
	StaffID uuid.UUID `json:"staff_id" binding:"required"`
}

// StaffRequirementScenarioUpdate represents data for updating a scenario
type StaffRequirementScenarioUpdate struct {
	ScenarioName           *string    `json:"scenario_name,omitempty"`
	Description            *string    `json:"description,omitempty"`
	DoctorID               *uuid.UUID `json:"doctor_id,omitempty"`
	BranchID               *uuid.UUID `json:"branch_id,omitempty"`
	RevenueLevelTierID     *uuid.UUID `json:"revenue_level_tier_id,omitempty"`
	MinRevenue             *float64   `json:"min_revenue,omitempty"`
	MaxRevenue             *float64   `json:"max_revenue,omitempty"`
	UseDayOfWeekRevenue    *bool      `json:"use_day_of_week_revenue,omitempty"`
	UseSpecificDateRevenue *bool      `json:"use_specific_date_revenue,omitempty"`
	DoctorCount            *int       `json:"doctor_count,omitempty"`
	MinDoctorCount         *int       `json:"min_doctor_count,omitempty"`
	DayOfWeek              *int       `json:"day_of_week,omitempty"`
	IsDefault              *bool      `json:"is_default,omitempty"`
	IsActive               *bool      `json:"is_active,omitempty"`
	Priority               *int       `json:"priority,omitempty"`
}

// CalculatedRequirement represents calculated staff requirements for a position
type CalculatedRequirement struct {
	PositionID          uuid.UUID  `json:"position_id"`
	PositionName        string     `json:"position_name"`
	BasePreferred       int        `json:"base_preferred"`
	BaseMinimum         int        `json:"base_minimum"`
	CalculatedPreferred int        `json:"calculated_preferred"`
	CalculatedMinimum   int        `json:"calculated_minimum"`
	MatchedScenarioID   *uuid.UUID `json:"matched_scenario_id,omitempty"`
	MatchedScenarioName *string    `json:"matched_scenario_name,omitempty"`
	FactorsApplied      []string   `json:"factors_applied"`
}

// ScenarioMatch represents a scenario that matches given conditions
type ScenarioMatch struct {
	ScenarioID   uuid.UUID `json:"scenario_id"`
	ScenarioName string    `json:"scenario_name"`
	Matches      bool      `json:"matches"`
	MatchReason  string    `json:"match_reason"`
	Priority     int       `json:"priority"`
}
