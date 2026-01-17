package models

import (
	"time"

	"github.com/google/uuid"
)

// DoctorAssignment represents a doctor assignment to a branch on a specific date
type DoctorAssignment struct {
	ID              uuid.UUID `json:"id" db:"id"`
	DoctorID        uuid.UUID `json:"doctor_id" db:"doctor_id"`
	Doctor          *Doctor   `json:"doctor,omitempty"` // Linked doctor profile
	DoctorName      string    `json:"doctor_name,omitempty" db:"doctor_name"` // Deprecated: use Doctor.Name
	DoctorCode      string    `json:"doctor_code,omitempty" db:"doctor_code"` // Deprecated: use Doctor.Code
	BranchID        uuid.UUID `json:"branch_id" db:"branch_id"`
	Branch          *Branch   `json:"branch,omitempty"`
	Date            time.Time `json:"date" db:"date"`
	ExpectedRevenue float64   `json:"expected_revenue" db:"expected_revenue"` // Doctor's expected revenue for this branch on this date
	CreatedBy       uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// DoctorOnOffDay represents doctor-on/doctor-off day designation for a branch
type DoctorOnOffDay struct {
	ID        uuid.UUID `json:"id" db:"id"`
	BranchID  uuid.UUID `json:"branch_id" db:"branch_id"`
	Branch    *Branch   `json:"branch,omitempty"`
	Date      time.Time `json:"date" db:"date"`
	IsDoctorOn bool     `json:"is_doctor_on" db:"is_doctor_on"` // true = doctor-on day, false = doctor-off day
	CreatedBy uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// DoctorDefaultSchedule represents a default branch assignment for a doctor on a specific day of the week
// DayOfWeek: 0 = Sunday, 1 = Monday, ..., 6 = Saturday
type DoctorDefaultSchedule struct {
	ID        uuid.UUID `json:"id" db:"id"`
	DoctorID  uuid.UUID `json:"doctor_id" db:"doctor_id"`
	Doctor    *Doctor   `json:"doctor,omitempty"`
	DayOfWeek int       `json:"day_of_week" db:"day_of_week"` // 0-6 (Sunday-Saturday)
	BranchID  uuid.UUID `json:"branch_id" db:"branch_id"`
	Branch    *Branch   `json:"branch,omitempty"`
	CreatedBy uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// DoctorWeeklyOffDay represents a default weekly off day for a doctor
// DayOfWeek: 0 = Sunday, 1 = Monday, ..., 6 = Saturday
type DoctorWeeklyOffDay struct {
	ID        uuid.UUID `json:"id" db:"id"`
	DoctorID  uuid.UUID `json:"doctor_id" db:"doctor_id"`
	Doctor    *Doctor   `json:"doctor,omitempty"`
	DayOfWeek int       `json:"day_of_week" db:"day_of_week"` // 0-6 (Sunday-Saturday)
	CreatedBy uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// DoctorScheduleOverride represents an override for a doctor's schedule on a specific date
// Type: "working" (doctor works at specified branch) or "off" (doctor is off)
type DoctorScheduleOverride struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	DoctorID  uuid.UUID `json:"doctor_id" db:"doctor_id"`
	Doctor    *Doctor    `json:"doctor,omitempty"`
	Date      time.Time  `json:"date" db:"date"`
	Type      string     `json:"type" db:"type"` // "working" or "off"
	BranchID  *uuid.UUID `json:"branch_id,omitempty" db:"branch_id"` // Required if type is "working", null if type is "off"
	Branch    *Branch     `json:"branch,omitempty"`
	CreatedBy uuid.UUID   `json:"created_by" db:"created_by"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt time.Time   `json:"updated_at" db:"updated_at"`
}