package models

import (
	"time"

	"github.com/google/uuid"
)

// RevenueLevelTier represents a revenue level tier configuration
type RevenueLevelTier struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	LevelNumber int        `json:"level_number" db:"level_number"`
	LevelName   string     `json:"level_name" db:"level_name"`
	MinRevenue  float64    `json:"min_revenue" db:"min_revenue"`
	MaxRevenue  *float64   `json:"max_revenue,omitempty" db:"max_revenue"`
	DisplayOrder int       `json:"display_order" db:"display_order"`
	ColorCode   *string    `json:"color_code,omitempty" db:"color_code"`
	Description *string    `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// RevenueLevelTierCreate represents data for creating a new revenue level tier
type RevenueLevelTierCreate struct {
	LevelNumber int     `json:"level_number" binding:"required,min=1,max=10"`
	LevelName   string  `json:"level_name" binding:"required,max=50"`
	MinRevenue  float64 `json:"min_revenue" binding:"required,min=0"`
	MaxRevenue  *float64 `json:"max_revenue,omitempty"`
	DisplayOrder int     `json:"display_order"`
	ColorCode   *string  `json:"color_code,omitempty"`
	Description *string  `json:"description,omitempty"`
}

// RevenueLevelTierUpdate represents data for updating a revenue level tier
type RevenueLevelTierUpdate struct {
	LevelName   *string  `json:"level_name,omitempty"`
	MinRevenue  *float64 `json:"min_revenue,omitempty"`
	MaxRevenue  *float64 `json:"max_revenue,omitempty"`
	DisplayOrder *int    `json:"display_order,omitempty"`
	ColorCode   *string  `json:"color_code,omitempty"`
	Description *string  `json:"description,omitempty"`
}
