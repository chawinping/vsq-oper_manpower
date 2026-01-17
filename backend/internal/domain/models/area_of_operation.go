package models

import (
	"time"

	"github.com/google/uuid"
)

type AreaOfOperation struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	Name        string      `json:"name" db:"name"`
	Code        string      `json:"code" db:"code"` // Short code for the area
	Description string      `json:"description" db:"description"`
	IsActive    bool        `json:"is_active" db:"is_active"`
	Zones       []*Zone     `json:"zones,omitempty"`       // Zones associated with this area
	Branches    []*Branch   `json:"branches,omitempty"`    // Individual branches added to this area (apart from zones)
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}


