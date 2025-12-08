package repositories

import (
	"database/sql"
	"fmt"

	"adminbe/internal/app/models"
)

// RoleMenuRepository interface defines data access methods for role menus
type RoleMenuRepository interface {
	GetAll() ([]models.RoleMenu, error)
	GetByRoleAndMenu(roleID, menuID uint) (*models.RoleMenu, error)
	Create(req models.RoleMenu) error
	Delete(roleID, menuID uint, deletedBy *uint64) error
}

// roleMenuRepository implements RoleMenuRepository
type roleMenuRepository struct {
	db *sql.DB
}

// NewRoleMenuRepository creates a new role menu repository
func NewRoleMenuRepository(db *sql.DB) RoleMenuRepository {
	return &roleMenuRepository{db: db}
}

// GetAll retrieves all active role-menu assignments
func (r *roleMenuRepository) GetAll() ([]models.RoleMenu, error) {
	rows, err := r.db.Query(`
		SELECT role_id, menu_id, deleted_at, deleted_by
		FROM role_menu
		WHERE deleted_at IS NULL`)
	if err != nil {
		return nil, fmt.Errorf("failed to query role menus: %w", err)
	}
	defer rows.Close()

	var roleMenus []models.RoleMenu
	for rows.Next() {
		var rm models.RoleMenu
		if err := rows.Scan(&rm.RoleID, &rm.MenuID, &rm.DeletedAt, &rm.DeletedBy); err != nil {
			return nil, fmt.Errorf("failed to scan role menu: %w", err)
		}
		roleMenus = append(roleMenus, rm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating role menus: %w", err)
	}

	return roleMenus, nil
}

// GetByRoleAndMenu retrieves a role-menu assignment by role and menu IDs
func (r *roleMenuRepository) GetByRoleAndMenu(roleID, menuID uint) (*models.RoleMenu, error) {
	var rm models.RoleMenu
	row := r.db.QueryRow(`
		SELECT role_id, menu_id, deleted_at, deleted_by
		FROM role_menu
		WHERE role_id = ? AND menu_id = ? AND deleted_at IS NULL`,
		roleID, menuID)

	err := row.Scan(&rm.RoleID, &rm.MenuID, &rm.DeletedAt, &rm.DeletedBy)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan role menu: %w", err)
	}

	return &rm, nil
}

// Create inserts a new role-menu assignment
func (r *roleMenuRepository) Create(req models.RoleMenu) error {
	_, err := r.db.Exec(`
		INSERT INTO role_menu (role_id, menu_id, deleted_at, deleted_by)
		VALUES (?, ?, ?, ?)`,
		req.RoleID, req.MenuID, req.DeletedAt, req.DeletedBy)
	return err
}

// Delete performs a soft delete
func (r *roleMenuRepository) Delete(roleID, menuID uint, deletedBy *uint64) error {
	_, err := r.db.Exec(`
		UPDATE role_menu SET deleted_at = NOW(), deleted_by = ?
		WHERE role_id = ? AND menu_id = ? AND deleted_at IS NULL`,
		deletedBy, roleID, menuID)
	return err
}
