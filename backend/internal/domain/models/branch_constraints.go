package models

import (
	"time"

	"github.com/google/uuid"
)

// BranchConstraints represents minimum staff constraints for a branch per day
type BranchConstraints struct {
	ID                uuid.UUID `json:"id" db:"id"`
	BranchID          uuid.UUID `json:"branch_id" db:"branch_id"`
	Branch            *Branch   `json:"branch,omitempty"`
	DayOfWeek         int       `json:"day_of_week" db:"day_of_week"` // 0=Sunday, 1=Monday, ..., 6=Saturday
	MinFrontStaff     int       `json:"min_front_staff" db:"min_front_staff"`         // Minimum combined front/counter staff (includes managers)
	MinManagers       int       `json:"min_managers" db:"min_managers"`               // Minimum combined managers (Branch Manager + Assistant Branch Manager)
	MinDoctorAssistant int      `json:"min_doctor_assistant" db:"min_doctor_assistant"` // Minimum combined doctor assistants
	MinTotalStaff     int       `json:"min_total_staff" db:"min_total_staff"`         // Minimum total staff in branch
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}
