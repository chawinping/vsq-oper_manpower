package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SpecificPreferenceType represents the type of preference
type SpecificPreferenceType string

const (
	SpecificPreferenceTypePositionCount SpecificPreferenceType = "position_count" // Requires a position and count
	SpecificPreferenceTypeStaffName     SpecificPreferenceType = "staff_name"     // Requires a specific staff member
)

// SpecificPreference represents a preference rule that can override or mix with other filters
// It allows setting preferences based on combinations of branch, doctor, and day of week
type SpecificPreference struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	BranchID    *uuid.UUID             `json:"branch_id,omitempty" db:"branch_id"`     // NULL = any branch
	Branch      *Branch                `json:"branch,omitempty"`
	DoctorID    *uuid.UUID             `json:"doctor_id,omitempty" db:"doctor_id"`     // NULL = any doctor
	Doctor      *Doctor                `json:"doctor,omitempty"`
	DayOfWeek   *int                   `json:"day_of_week,omitempty" db:"day_of_week"` // NULL = any day, 0-6 (Sunday-Saturday)
	PreferenceType SpecificPreferenceType `json:"preference_type" db:"preference_type"`
	
	// For position_count type: position_id and staff_count are required
	PositionID  *uuid.UUID             `json:"position_id,omitempty" db:"position_id"`
	Position    *Position               `json:"position,omitempty"`
	StaffCount  *int                   `json:"staff_count,omitempty" db:"staff_count"`
	
	// For staff_name type: staff_id is required
	StaffID     *uuid.UUID             `json:"staff_id,omitempty" db:"staff_id"`
	Staff       *Staff                 `json:"staff,omitempty"`
	
	IsActive    bool                   `json:"is_active" db:"is_active"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// Validate validates the SpecificPreference model
func (sp *SpecificPreference) Validate() error {
	if sp.PreferenceType == SpecificPreferenceTypePositionCount {
		if sp.PositionID == nil {
			return fmt.Errorf("position_id is required for position_count preference type")
		}
		if sp.StaffCount == nil || *sp.StaffCount < 1 {
			return fmt.Errorf("staff_count must be at least 1 for position_count preference type")
		}
	} else if sp.PreferenceType == SpecificPreferenceTypeStaffName {
		if sp.StaffID == nil {
			return fmt.Errorf("staff_id is required for staff_name preference type")
		}
	} else {
		return fmt.Errorf("invalid preference_type: %s", sp.PreferenceType)
	}
	return nil
}

// Matches checks if this preference matches the given branch, doctor, and day of week
func (sp *SpecificPreference) Matches(branchID *uuid.UUID, doctorID *uuid.UUID, dayOfWeek *int) bool {
	if !sp.IsActive {
		return false
	}
	
	// Check branch match (NULL = any branch)
	if sp.BranchID != nil {
		if branchID == nil || *sp.BranchID != *branchID {
			return false
		}
	}
	
	// Check doctor match (NULL = any doctor)
	if sp.DoctorID != nil {
		if doctorID == nil || *sp.DoctorID != *doctorID {
			return false
		}
	}
	
	// Check day of week match (NULL = any day)
	if sp.DayOfWeek != nil {
		if dayOfWeek == nil || *sp.DayOfWeek != *dayOfWeek {
			return false
		}
	}
	
	return true
}
