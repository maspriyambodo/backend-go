package models

import (
	"time"
)

// UserRole represents the user_roles table
type UserRole struct {
	UserID    uint64     `json:"user_id" db:"user_id"`
	RoleID    uint       `json:"role_id" db:"role_id"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
	DeletedBy *uint64    `json:"deleted_by" db:"deleted_by"`
}

// CreateUserRoleRequest for creating a new user-role assignment
type CreateUserRoleRequest struct {
	UserID uint64 `json:"user_id" binding:"required"`
	RoleID uint   `json:"role_id" binding:"required"`
}

// UpdateUserRoleRequest for updating an existing user-role assignment
type UpdateUserRoleRequest struct {
	UserID *uint64 `json:"user_id,omitempty"`
	RoleID *uint   `json:"role_id,omitempty"`
}
