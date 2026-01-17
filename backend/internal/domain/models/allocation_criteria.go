package models

import (
	"time"

	"github.com/google/uuid"
)

// CriteriaPillar represents the three pillars of allocation criteria
type CriteriaPillar string

const (
	PillarClinicWide    CriteriaPillar = "clinic_wide"
	PillarDoctorSpecific CriteriaPillar = "doctor_specific"
	PillarBranchSpecific CriteriaPillar = "branch_specific"
)

// CriteriaType represents the type of criteria
type CriteriaType string

const (
	CriteriaTypeBookings          CriteriaType = "bookings"
	CriteriaTypeRevenue           CriteriaType = "revenue"
	CriteriaTypeMinStaffPosition  CriteriaType = "min_staff_position"
	CriteriaTypeMinStaffBranch    CriteriaType = "min_staff_branch"
	CriteriaTypeDoctorCount       CriteriaType = "doctor_count"
	CriteriaTypeDoctorSpecificStaff CriteriaType = "doctor_specific_staff"
)

// AllocationCriteria represents configurable allocation criteria
type AllocationCriteria struct {
	ID          uuid.UUID      `json:"id" db:"id"`
	Pillar      CriteriaPillar `json:"pillar" db:"pillar"`
	Type        CriteriaType   `json:"type" db:"type"`
	Weight      float64        `json:"weight" db:"weight"` // Weight for this criteria (0.0 - 1.0)
	IsActive    bool           `json:"is_active" db:"is_active"`
	Description string         `json:"description" db:"description"`
	Config      string         `json:"config" db:"config"` // JSON config for criteria-specific settings
	CreatedBy   uuid.UUID      `json:"created_by" db:"created_by"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}
