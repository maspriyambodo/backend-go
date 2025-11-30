package models

import (
	"time"
)

// User represents the users table
type User struct {
	ID           uint64     `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	Status       uint8      `json:"status" db:"status"`
	CreatedAt    *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at" db:"deleted_at"`
	DeletedBy    *uint64    `json:"deleted_by" db:"deleted_by"`
}

// AuditLog represents the audit_logs table
type AuditLog struct {
	ID        uint64      `json:"id" db:"id"`
	UserID    *uint64     `json:"user_id" db:"user_id"`
	EventType string      `json:"event_type" db:"event_type"`
	TableName string      `json:"table_name" db:"table_name"`
	RecordID  uint64      `json:"record_id" db:"record_id"`
	OldValues interface{} `json:"old_values" db:"old_values"`
	NewValues interface{} `json:"new_values" db:"new_values"`
	IPAddress []byte      `json:"ip_address" db:"ip_address"`
	UserAgent *string     `json:"user_agent" db:"user_agent"`
	CreatedAt *time.Time  `json:"created_at" db:"created_at"`
}

// Menu represents the menu table
type Menu struct {
	ID        uint       `json:"id" db:"id"`
	Label     string     `json:"label" db:"label"`
	Url       *string    `json:"url" db:"url"`
	Icon      *string    `json:"icon" db:"icon"`
	ParentID  *uint      `json:"parent_id" db:"parent_id"`
	SortOrder uint16     `json:"sort_order" db:"sort_order"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
	DeletedBy *uint64    `json:"deleted_by" db:"deleted_by"`
}

// CreateUserRequest for creating a new user
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Status   *uint8 `json:"status,omitempty"`
}

// UpdateUserRequest for updating an existing user
type UpdateUserRequest struct {
	Username string `json:"username,omitempty" binding:"min=3,max=100"`
	Email    string `json:"email,omitempty" binding:"email"`
	Password string `json:"password,omitempty" binding:"min=6"`
	Status   *uint8 `json:"status,omitempty"`
}

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

// MenuNavigation represents the menu_navigation view
type MenuNavigation struct {
	ID       uint   `json:"id" db:"id"`
	Label    string `json:"label" db:"label"`
	URL      string `json:"url" db:"url"`
	Icon     string `json:"icon" db:"icon"`
	Children string `json:"children" db:"children"`
}
