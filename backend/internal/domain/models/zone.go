package models

import (
	"time"

	"github.com/google/uuid"
)

// Zone represents a zone in Bangkok that consists of multiple branches
type Zone struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Code        string    `json:"code" db:"code"` // Short code for the zone
	Description string    `json:"description" db:"description"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	Branches    []*Branch `json:"branches,omitempty"` // Branches in this zone
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ZoneBranch represents the relationship between a zone and a branch
type ZoneBranch struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ZoneID    uuid.UUID `json:"zone_id" db:"zone_id"`
	BranchID  uuid.UUID `json:"branch_id" db:"branch_id"`
	Zone      *Zone     `json:"zone,omitempty"`
	Branch    *Branch   `json:"branch,omitempty"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// AreaOfOperationZone represents the relationship between an area of operation and a zone
type AreaOfOperationZone struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	AreaOfOperationID uuid.UUID `json:"area_of_operation_id" db:"area_of_operation_id"`
	ZoneID             uuid.UUID `json:"zone_id" db:"zone_id"`
	AreaOfOperation    *AreaOfOperation `json:"area_of_operation,omitempty"`
	Zone               *Zone     `json:"zone,omitempty"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}

// AreaOfOperationBranch represents individual branches added to an area of operation (apart from zones)
type AreaOfOperationBranch struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	AreaOfOperationID uuid.UUID `json:"area_of_operation_id" db:"area_of_operation_id"`
	BranchID           uuid.UUID `json:"branch_id" db:"branch_id"`
	AreaOfOperation    *AreaOfOperation `json:"area_of_operation,omitempty"`
	Branch             *Branch   `json:"branch,omitempty"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}
