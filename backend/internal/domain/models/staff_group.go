package models

import (
	"time"

	"github.com/google/uuid"
)

// StaffGroup represents a group of positions that work together
type StaffGroup struct {
	ID          uuid.UUID             `json:"id" db:"id"`
	Name        string                `json:"name" db:"name"`
	Description string                `json:"description" db:"description"`
	IsActive    bool                  `json:"is_active" db:"is_active"`
	Positions   []*StaffGroupPosition `json:"positions,omitempty"`
	CreatedAt   time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at" db:"updated_at"`
}
