package models

import (
	"time"
)

// RoleMenu represents the role_menu table
type RoleMenu struct {
	RoleID    uint       `json:"role_id" db:"role_id"`
	MenuID    uint       `json:"menu_id" db:"menu_id"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
	DeletedBy *uint64    `json:"deleted_by" db:"deleted_by"`
}

// CreateRoleMenuRequest for creating a new role-menu assignment
type CreateRoleMenuRequest struct {
	RoleID uint `json:"role_id" binding:"required"`
	MenuID uint `json:"menu_id" binding:"required"`
}

// UpdateRoleMenuRequest for updating an existing role-menu assignment
type UpdateRoleMenuRequest struct {
	RoleID *uint `json:"role_id,omitempty"`
	MenuID *uint `json:"menu_id,omitempty"`
}
