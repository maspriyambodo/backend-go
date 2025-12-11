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
