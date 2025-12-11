package models

import (
	"time"
)

// UserMenu represents the user_menu table
type UserMenu struct {
	UserID    uint64     `json:"user_id" db:"user_id"`
	MenuID    uint       `json:"menu_id" db:"menu_id"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
	DeletedBy *uint64    `json:"deleted_by" db:"deleted_by"`
}

// CreateUserMenuRequest for creating a new user-menu assignment
type CreateUserMenuRequest struct {
	UserID uint64 `json:"user_id" binding:"required"`
	MenuID uint   `json:"menu_id" binding:"required"`
}

// UpdateUserMenuRequest for updating an existing user-menu assignment
type UpdateUserMenuRequest struct {
	UserID *uint64 `json:"user_id,omitempty"`
	MenuID *uint   `json:"menu_id,omitempty"`
}
