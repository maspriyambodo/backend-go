package repositories

import (
	"database/sql"
	"fmt"

	"adminbe/internal/app/models"
)

// UserRoleRepository interface defines data access methods for user roles
type UserRoleRepository interface {
	GetAll() ([]models.UserRole, error)
	GetByUserAndRole(userID uint64, roleID uint) (*models.UserRole, error)
	Create(req models.UserRole) error
	Delete(userID uint64, roleID uint, deletedBy *uint64) error
}

// userRoleRepository implements UserRoleRepository
type userRoleRepository struct {
	db *sql.DB
}

// NewUserRoleRepository creates a new user role repository
func NewUserRoleRepository(db *sql.DB) UserRoleRepository {
	return &userRoleRepository{db: db}
}

// GetAll retrieves all active user-role assignments
func (r *userRoleRepository) GetAll() ([]models.UserRole, error) {
	rows, err := r.db.Query(`
		SELECT user_id, role_id, deleted_at, deleted_by
		FROM user_roles
		WHERE deleted_at IS NULL`)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}
	defer rows.Close()

	var userRoles []models.UserRole
	for rows.Next() {
		var ur models.UserRole
		if err := rows.Scan(&ur.UserID, &ur.RoleID, &ur.DeletedAt, &ur.DeletedBy); err != nil {
			return nil, fmt.Errorf("failed to scan user role: %w", err)
		}
		userRoles = append(userRoles, ur)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user roles: %w", err)
	}

	return userRoles, nil
}

// GetByUserAndRole retrieves a user-role assignment by user and role IDs
func (r *userRoleRepository) GetByUserAndRole(userID uint64, roleID uint) (*models.UserRole, error) {
	var ur models.UserRole
	row := r.db.QueryRow(`
		SELECT user_id, role_id, deleted_at, deleted_by
		FROM user_roles
		WHERE user_id = ? AND role_id = ? AND deleted_at IS NULL`,
		userID, roleID)

	err := row.Scan(&ur.UserID, &ur.RoleID, &ur.DeletedAt, &ur.DeletedBy)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan user role: %w", err)
	}

	return &ur, nil
}

// Create inserts a new user-role assignment
func (r *userRoleRepository) Create(req models.UserRole) error {
	_, err := r.db.Exec(`
		INSERT INTO user_roles (user_id, role_id, deleted_at, deleted_by)
		VALUES (?, ?, ?, ?)`,
		req.UserID, req.RoleID, req.DeletedAt, req.DeletedBy)
	return err
}

// Delete performs a soft delete
func (r *userRoleRepository) Delete(userID uint64, roleID uint, deletedBy *uint64) error {
	_, err := r.db.Exec(`
		UPDATE user_roles SET deleted_at = NOW(), deleted_by = ?
		WHERE user_id = ? AND role_id = ? AND deleted_at IS NULL`,
		deletedBy, userID, roleID)
	return err
}
