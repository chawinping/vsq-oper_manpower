package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	PasswordHash string `json:"-" db:"password_hash"`
	RoleID    uuid.UUID `json:"role_id" db:"role_id"`
	Role      *Role     `json:"role,omitempty"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Role struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"` // admin, area_manager, district_manager, branch_manager, viewer
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}



