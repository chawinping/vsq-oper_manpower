package models

import (
	"time"

	"github.com/google/uuid"
)

// StaffGroupPosition represents the mapping between a staff group and a position
type StaffGroupPosition struct {
	ID           uuid.UUID   `json:"id" db:"id"`
	StaffGroupID uuid.UUID   `json:"staff_group_id" db:"staff_group_id"`
	StaffGroup   *StaffGroup `json:"staff_group,omitempty"`
	PositionID   uuid.UUID   `json:"position_id" db:"position_id"`
	Position     *Position   `json:"position,omitempty"`
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
}
