package models

import (
	"time"
)

// Role represents the roles table
type Role struct {
	ID          uint       `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description *string    `json:"description" db:"description"`
	CreatedAt   *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at" db:"deleted_at"`
	DeletedBy   *uint64    `json:"deleted_by" db:"deleted_by"`
}

// CreateRoleRequest for creating a new role
type CreateRoleRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Description *string `json:"description,omitempty"`
}

// UpdateRoleRequest for updating an existing role
type UpdateRoleRequest struct {
	Name        *string `json:"name,omitempty" binding:"min=1,max=100"`
	Description *string `json:"description,omitempty"`
}

// RoleInheritance represents the role_inheritances table
type RoleInheritance struct {
	ID           uint64     `json:"id" db:"id"`
	RoleID       uint       `json:"role_id" db:"role_id"`
	ParentRoleID uint       `json:"parent_role_id" db:"parent_role_id"`
	CreatedAt    *time.Time `json:"created_at" db:"created_at"`
}

// CreateRoleInheritanceRequest for creating a new role inheritance
type CreateRoleInheritanceRequest struct {
	RoleID       uint `json:"role_id" binding:"required"`
	ParentRoleID uint `json:"parent_role_id" binding:"required"`
}

// UpdateRoleInheritanceRequest for updating an existing role inheritance
type UpdateRoleInheritanceRequest struct {
	RoleID       *uint `json:"role_id,omitempty"`
	ParentRoleID *uint `json:"parent_role_id,omitempty"`
}

// VRole represents the v_roles view
type VRole struct {
	RoleID    uint   `json:"role_id" db:"role_id"`
	RoleName  string `json:"role_name" db:"role_name"`
	ChildID   uint   `json:"child_id" db:"child_id"`
	ChildName string `json:"child_name" db:"child_name"`
	Level     uint   `json:"level" db:"level"`
}
